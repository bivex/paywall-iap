package integration

// Tests for pricing tier multi-tenancy (app_id isolation).
// Each app sees only its own tiers; X-App-ID header is required.

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

	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	httpmiddleware "github.com/bivex/paywall-iap/internal/interfaces/http/middleware"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestPricingTierMultitenancy(t *testing.T) {
	ctx := context.Background()
	pool, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = pool.Exec(ctx, `
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
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			platform_user_id TEXT UNIQUE NOT NULL,
			device_id        TEXT,
			platform         TEXT NOT NULL,
			app_version      TEXT NOT NULL,
			email            TEXT UNIQUE,
			role             TEXT NOT NULL DEFAULT 'user',
			ltv              NUMERIC(10,2) DEFAULT 0,
			ltv_updated_at   TIMESTAMPTZ,
			created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at       TIMESTAMPTZ
		);

		CREATE TABLE pricing_tiers (
			id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			app_id         UUID NOT NULL REFERENCES apps(id),
			name           TEXT NOT NULL,
			description    TEXT,
			monthly_price  NUMERIC(10,2),
			annual_price   NUMERIC(10,2),
			lifetime_price NUMERIC(10,2),
			currency       CHAR(3) NOT NULL DEFAULT 'USD',
			features       JSONB,
			is_active      BOOLEAN NOT NULL DEFAULT true,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at     TIMESTAMPTZ,
			UNIQUE (app_id, name)
		);

		CREATE TABLE admin_audit_log (
			id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			admin_id       UUID NOT NULL,
			action         TEXT NOT NULL,
			target_type    TEXT NOT NULL,
			target_user_id UUID,
			details        JSONB,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	require.NoError(t, err)

	appA := uuid.New()
	appB := uuid.New()
	adminID := uuid.New()

	_, err = pool.Exec(ctx, `
		INSERT INTO apps (id, name, display_name, platform, bundle_id) VALUES
			($1, 'com.test.app_a', 'App A', 'ios',     'com.test.app_a'),
			($2, 'com.test.app_b', 'App B', 'android', 'com.test.app_b')
	`, appA, appB)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role)
		 VALUES ($1, 'admin-mt', 'dev-mt', 'ios', '1.0.0', 'admin-mt@test.com', 'admin')`,
		adminID,
	)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)

	// router helper: injects admin_id + app_id from header into gin context
	newRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("admin_id", adminID)
			c.Set("user_id", adminID.String())
			c.Next()
		})
		r.Use(httpmiddleware.RequireAppID())
		h := handlers.NewAdminHandler(nil, nil, nil, pool, nil, nil, nil, nil, nil, nil, nil, nil)
		g := r.Group("/v1/admin")
		g.GET("/pricing-tiers", h.ListPricingTiers)
		g.POST("/pricing-tiers", h.CreatePricingTier)
		g.PUT("/pricing-tiers/:id", h.UpdatePricingTier)
		g.POST("/pricing-tiers/:id/activate", h.ActivatePricingTier)
		g.POST("/pricing-tiers/:id/deactivate", h.DeactivatePricingTier)
		return r
	}

	var tierAID string

	t.Run("no X-App-ID header → 400", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/pricing-tiers", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET app_a returns empty before creation", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/pricing-tiers", nil)
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Empty(t, resp.Data)
	})

	t.Run("POST creates tier for app_a", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{"name":"Pro","description":"App A tier","monthly_price":9.99,"currency":"usd","features":["All features"],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		var resp struct {
			Data handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "Pro", resp.Data.Name)
		tierAID = resp.Data.ID
	})

	t.Run("POST creates same-name tier for app_b without conflict", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{"name":"Pro","description":"App B tier","monthly_price":7.99,"currency":"usd","features":["Basic"],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appB.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		// same name is allowed across different apps
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GET app_a sees only its own tier", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/pricing-tiers", nil)
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "Pro", resp.Data[0].Name)
		assert.InDelta(t, 9.99, *resp.Data[0].MonthlyPrice, 0.001)
	})

	t.Run("GET app_b sees only its own tier", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/pricing-tiers", nil)
		req.Header.Set("X-App-ID", appB.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 1)
		assert.InDelta(t, 7.99, *resp.Data[0].MonthlyPrice, 0.001)
	})

	t.Run("POST duplicate name within same app → 409", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{"name":"Pro","description":"Duplicate","monthly_price":5.99,"currency":"usd","features":[],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("PUT updates tier for app_a", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{"name":"Pro","description":"Updated","monthly_price":11.99,"currency":"eur","features":["All features","New"],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/pricing-tiers/"+tierAID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "EUR", resp.Data.Currency)
		assert.InDelta(t, 11.99, *resp.Data.MonthlyPrice, 0.001)
	})

	t.Run("POST deactivate app_a tier", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers/"+tierAID+"/deactivate", nil)
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp.Data.IsActive)
	})

	t.Run("POST activate app_a tier", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers/"+tierAID+"/activate", nil)
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.PricingTier `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Data.IsActive)
	})

	t.Run("POST missing required fields → 422", func(t *testing.T) {
		r := newRouter()
		// missing currency
		body := []byte(`{"name":"Incomplete","description":"No currency","monthly_price":1.99,"features":[],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("POST no prices at all → 422", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{"name":"NoPrices","description":"Missing all prices","currency":"usd","features":[],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/pricing-tiers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("PUT invalid tier UUID → 400", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{"name":"X","description":"X","monthly_price":1.0,"currency":"usd","features":[],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/pricing-tiers/not-a-uuid", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT unknown tier UUID → 404", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{"name":"X","description":"X","monthly_price":1.0,"currency":"usd","features":[],"is_active":true}`)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/pricing-tiers/"+uuid.New().String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
