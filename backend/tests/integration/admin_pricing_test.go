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

	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestAdminPricingHandler(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE apps (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name         TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			platform     TEXT NOT NULL CHECK (platform IN ('ios','android','both')),
			bundle_id    TEXT NOT NULL UNIQUE,
			is_active    BOOLEAN NOT NULL DEFAULT true,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
		);
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
		CREATE TABLE pricing_tiers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			app_id UUID NOT NULL REFERENCES apps(id),
			name TEXT NOT NULL,
			description TEXT,
			monthly_price NUMERIC(10,2),
			annual_price NUMERIC(10,2),
			lifetime_price NUMERIC(10,2),
			currency CHAR(3) NOT NULL DEFAULT 'USD',
			features JSONB,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ,
			UNIQUE (app_id, name)
		);
		CREATE TABLE admin_audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			admin_id UUID NOT NULL,
			action TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_user_id UUID,
			details JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`)
	require.NoError(t, err)

	appID := uuid.New()
	_, err = db.Exec(ctx,
		`INSERT INTO apps (id, name, display_name, bundle_id, platform) VALUES ($1, $2, $3, $4, $5)`,
		appID, "Pricing Test App", "Pricing Test App", "com.pricing.test", "ios",
	)
	require.NoError(t, err)

	adminID := uuid.New()
	_, err = db.Exec(ctx,
		`INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		adminID, "admin-pricing", "device-1", "ios", "1.0.0", "pricing-admin@example.com",
	)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Set("user_id", adminID.String())
		c.Set("app_id", appID)
		c.Next()
	})

	handler := handlers.NewAdminHandler(handlers.AdminHandlerDeps{
		AdminInfraDeps: handlers.AdminInfraDeps{
			Queries: generated.New(db),
			DBPool: db,
		},
	})

	admin := router.Group("/v1/admin")
	admin.GET("/pricing-tiers", handler.ListPricingTiers)
	admin.POST("/pricing-tiers", handler.CreatePricingTier)
	admin.PUT("/pricing-tiers/:id", handler.UpdatePricingTier)
	admin.POST("/pricing-tiers/:id/activate", handler.ActivatePricingTier)
	admin.POST("/pricing-tiers/:id/deactivate", handler.DeactivatePricingTier)

	var created handlers.PricingTier

	t.Run("GET returns empty list before any create", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/pricing-tiers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Empty(t, resp.Data)
	})

	t.Run("POST creates pricing tier", func(t *testing.T) {
		body := []byte(`{"name":"Pro","description":"Primary paid tier","monthly_price":9.99,"annual_price":99.99,"currency":"usd","features":["Unlimited access","Priority support"],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp struct {
			Data handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		created = resp.Data
		assert.Equal(t, "Pro", created.Name)
		assert.Equal(t, "USD", created.Currency)
		require.NotNil(t, created.MonthlyPrice)
		assert.Equal(t, 2, len(created.Features))
	})

	t.Run("PUT updates pricing tier", func(t *testing.T) {
		body := []byte(`{"name":"Pro Plus","description":"Updated tier","monthly_price":12.5,"annual_price":120,"currency":"eur","features":["Unlimited access","Team seats"],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/pricing-tiers/"+created.ID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		created = resp.Data
		assert.Equal(t, "Pro Plus", created.Name)
		assert.Equal(t, "EUR", created.Currency)
		require.NotNil(t, created.MonthlyPrice)
		assert.InDelta(t, 12.5, *created.MonthlyPrice, 0.001)
	})

	t.Run("POST deactivate toggles status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers/"+created.ID+"/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		created = resp.Data
		assert.False(t, created.IsActive)
	})

	t.Run("POST activate toggles status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers/"+created.ID+"/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		created = resp.Data
		assert.True(t, created.IsActive)
	})

	t.Run("GET returns persisted tier", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/pricing-tiers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 1)
		assert.Equal(t, created.ID, resp.Data[0].ID)
	})
}
