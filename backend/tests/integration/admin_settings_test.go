package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestAdminSettingsHandler(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE users (
			id UUID PRIMARY KEY,
			platform_user_id TEXT UNIQUE NOT NULL,
			device_id TEXT,
			platform TEXT NOT NULL,
			app_version TEXT NOT NULL,
			email TEXT UNIQUE,
			ltv NUMERIC(10,2) DEFAULT 0,
			ltv_updated_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ
		);
		CREATE TABLE admin_credentials (
			user_id UUID PRIMARY KEY REFERENCES users(id),
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE admin_audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			admin_id UUID NOT NULL,
			action TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_user_id UUID,
			details JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE admin_settings (
			key TEXT PRIMARY KEY,
			value JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	require.NoError(t, err)

	adminID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("old-password-123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	_, err = db.Exec(ctx,
		`INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		adminID, "admin-1", "device-1", "ios", "1.0.0", "admin@example.com",
	)
	require.NoError(t, err)
	_, err = db.Exec(ctx,
		`INSERT INTO admin_credentials (user_id, password_hash) VALUES ($1, $2)`,
		adminID, string(hash),
	)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Set("user_id", adminID.String())
		c.Next()
	})

	handler := handlers.NewAdminHandler(
		nil,
		nil,
		generated.New(db),
		db,
		nil,
		nil,
		service.NewAuditService(db),
		nil,
		nil,
		nil,
		nil,
	)

	admin := router.Group("/v1/admin")
	admin.GET("/settings", handler.GetPlatformSettings)
	admin.PUT("/settings", handler.UpdatePlatformSettings)
	admin.POST("/settings/password", handler.ChangeAdminPassword)

	t.Run("GET returns defaults before any save", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/settings", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.PlatformSettings `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "Paywall SaaS", resp.Data.General.PlatformName)
		assert.Equal(t, "USD", resp.Data.General.DefaultCurrency)
	})

	t.Run("PUT persists settings", func(t *testing.T) {
		payload := handlers.PlatformSettings{
			General: handlers.GeneralSettings{
				PlatformName:    "Acme Paywall",
				SupportEmail:    "ops@acme.test",
				DefaultCurrency: "eur",
				DarkModeDefault: true,
			},
			Integrations: handlers.IntegrationSettings{
				StripeAPIKey:        "sk_live_123",
				StripeWebhookSecret: "whsec_123",
				StripeTestMode:      true,
				MatomoURL:           "https://matomo.acme.test",
				MatomoSiteID:        "9",
			},
			Notifications: handlers.NotificationSettings{
				NewSubscription:       true,
				PaymentFailed:         true,
				SubscriptionCancelled: false,
				RefundIssued:          true,
				WebhookFailed:         true,
				DunningStarted:        false,
			},
			Security: handlers.PlatformSecurityState{
				JWTExpiryHours:    48,
				RequireMFA:        true,
				EnableIPAllowlist: true,
			},
		}
		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/v1/admin/settings", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.PlatformSettings `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "EUR", resp.Data.General.DefaultCurrency)

		var count int
		err = db.QueryRow(ctx, `SELECT COUNT(*) FROM admin_settings WHERE key = $1`, "platform_settings").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("POST password updates stored hash", func(t *testing.T) {
		body := []byte(`{"current_password":"old-password-123","new_password":"new-password-456","confirm_password":"new-password-456"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/settings/password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		cred, err := generated.New(db).GetAdminCredential(ctx, adminID)
		require.NoError(t, err)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte("new-password-456")))
	})
}
