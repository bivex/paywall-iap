package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestAdminWinbackHandler(t *testing.T) {
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
			role TEXT NOT NULL DEFAULT 'user',
			ltv NUMERIC(10,2) DEFAULT 0,
			ltv_updated_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ
		);
		CREATE TABLE subscriptions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id),
			status TEXT NOT NULL,
			source TEXT NOT NULL,
			platform TEXT NOT NULL,
			product_id TEXT NOT NULL,
			plan_type TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			auto_renew BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ
		);
		CREATE TABLE winback_offers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id),
			campaign_id TEXT NOT NULL,
			discount_type TEXT NOT NULL CHECK (discount_type IN ('percentage', 'fixed')),
			discount_value NUMERIC(10,2) NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('offered', 'accepted', 'expired', 'declined')),
			offered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			expires_at TIMESTAMPTZ NOT NULL,
			accepted_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE UNIQUE INDEX idx_winback_offers_user_campaign
			ON winback_offers(user_id, campaign_id)
			WHERE status IN ('offered', 'accepted');`)
	require.NoError(t, err)

	adminID := uuid.New()
	recentUserA := uuid.New()
	recentUserB := uuid.New()
	oldUser := uuid.New()
	for idx, id := range []uuid.UUID{adminID, recentUserA, recentUserB, oldUser} {
		_, err = db.Exec(ctx,
			`INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			id,
			fmt.Sprintf("user-%d", idx),
			fmt.Sprintf("device-%d", idx),
			"ios",
			"1.0.0",
			fmt.Sprintf("user-%d@example.com", idx),
			"admin",
		)
		require.NoError(t, err)
	}

	for _, tc := range []struct {
		userID    uuid.UUID
		updatedAt time.Time
	}{
		{userID: recentUserA, updatedAt: time.Now().Add(-48 * time.Hour)},
		{userID: recentUserB, updatedAt: time.Now().Add(-12 * time.Hour)},
		{userID: oldUser, updatedAt: time.Now().Add(-45 * 24 * time.Hour)},
	} {
		_, err = db.Exec(ctx,
			`INSERT INTO subscriptions (id, user_id, status, source, platform, product_id, plan_type, expires_at, auto_renew, created_at, updated_at)
			 VALUES ($1, $2, 'cancelled', 'stripe', 'web', 'pro', 'monthly', $3, false, $4, $4)`,
			uuid.New(), tc.userID, time.Now().Add(30*24*time.Hour), tc.updatedAt,
		)
		require.NoError(t, err)
	}

	queries := generated.New(db)
	userRepo := repository.NewUserRepository(queries)
	subscriptionRepo := repository.NewSubscriptionRepository(queries)
	winbackService := service.NewWinbackService(repository.NewWinbackOfferRepository(db), userRepo, subscriptionRepo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Set("user_id", adminID.String())
		c.Next()
	})

	handler := handlers.NewAdminHandler(
		subscriptionRepo,
		userRepo,
		queries,
		db,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		winbackService,
		nil,
	)

	admin := router.Group("/v1/admin")
	admin.GET("/winback-campaigns", handler.ListWinbackCampaigns)
	admin.POST("/winback-campaigns", handler.LaunchWinbackCampaign)

	t.Run("GET returns empty list before launch", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/winback-campaigns", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.WinbackCampaignSummary `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Empty(t, resp.Data)
	})

	t.Run("POST launches campaign for recently churned users", func(t *testing.T) {
		body := []byte(`{"campaign_id":"reactivate_q1","discount_type":"percentage","discount_value":25,"duration_days":14,"days_since_churn":30}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/winback-campaigns", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.WinbackCampaignSummary `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "reactivate_q1", resp.Data.CampaignID)
		assert.Equal(t, "percentage", resp.Data.DiscountType)
		assert.InDelta(t, 25, resp.Data.DiscountValue, 0.001)
		assert.Equal(t, 2, resp.Data.TotalOffers)
		assert.Equal(t, 2, resp.Data.ActiveOffers)
		assert.Equal(t, 0, resp.Data.AcceptedOffers)
	})

	t.Run("POST rejects duplicate campaign IDs", func(t *testing.T) {
		body := []byte(`{"campaign_id":"reactivate_q1","discount_type":"percentage","discount_value":20,"duration_days":10,"days_since_churn":30}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/winback-campaigns", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("GET returns grouped persisted campaign summary", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/winback-campaigns", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.WinbackCampaignSummary `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "reactivate_q1", resp.Data[0].CampaignID)
		assert.Equal(t, 2, resp.Data[0].TotalOffers)
	})
}
