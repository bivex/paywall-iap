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

func TestAdminExperimentsHandler(t *testing.T) {
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
		CREATE TABLE ab_tests (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'paused', 'completed')) DEFAULT 'draft',
			start_at TIMESTAMPTZ,
			end_at TIMESTAMPTZ,
			algorithm_type TEXT CHECK (algorithm_type IN ('thompson_sampling', 'ucb', 'epsilon_greedy')),
			is_bandit BOOLEAN NOT NULL DEFAULT false,
			min_sample_size INT DEFAULT 100,
			confidence_threshold NUMERIC(3,2) DEFAULT 0.95,
			winner_confidence NUMERIC(3,2),
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE ab_test_arms (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			description TEXT,
			is_control BOOLEAN NOT NULL DEFAULT false,
			traffic_weight NUMERIC(3,2) NOT NULL DEFAULT 1.0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE ab_test_arm_stats (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			arm_id UUID NOT NULL UNIQUE REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			alpha NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			beta NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			samples INT NOT NULL DEFAULT 0,
			conversions INT NOT NULL DEFAULT 0,
			revenue NUMERIC(15,2) NOT NULL DEFAULT 0.0,
			avg_reward NUMERIC(10,4),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			CHECK (alpha > 0),
			CHECK (beta > 0),
			CHECK (samples >= 0),
			CHECK (conversions >= 0),
			CHECK (conversions <= samples)
		);
		CREATE TABLE ab_test_assignments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours')
		);`)
	require.NoError(t, err)

	adminID := uuid.New()
	_, err = db.Exec(ctx,
		`INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		adminID,
		"admin-user",
		"admin-device",
		"ios",
		"1.0.0",
		"admin@example.com",
		"admin",
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
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	admin := router.Group("/v1/admin")
	admin.GET("/experiments", handler.ListAdminExperiments)
	admin.POST("/experiments", handler.CreateAdminExperiment)

	t.Run("GET returns empty list before creation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Empty(t, resp.Data)
	})

	t.Run("POST creates a db-backed experiment with arms", func(t *testing.T) {
		body := []byte(`{
			"name":"Pricing homepage CTA",
			"description":"Compare control vs annual emphasis",
			"status":"running",
			"algorithm_type":"thompson_sampling",
			"is_bandit":true,
			"min_sample_size":200,
			"confidence_threshold_percent":95,
			"arms":[
				{"name":"Control","description":"Baseline paywall","is_control":true,"traffic_weight":1},
				{"name":"Variant A","description":"Annual plan emphasis","is_control":false,"traffic_weight":1}
			]
		}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "Pricing homepage CTA", resp.Data.Name)
		assert.Equal(t, "running", resp.Data.Status)
		assert.True(t, resp.Data.IsBandit)
		assert.InDelta(t, 95, resp.Data.ConfidenceThresholdPercent, 0.001)
		assert.Len(t, resp.Data.Arms, 2)
		assert.Equal(t, 2, resp.Data.ArmCount)
		assert.Equal(t, 0, resp.Data.TotalSamples)
	})

	t.Run("GET returns created experiment", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "Pricing homepage CTA", resp.Data[0].Name)
		assert.Len(t, resp.Data[0].Arms, 2)
	})

	t.Run("POST rejects invalid arms payload", func(t *testing.T) {
		body := []byte(`{
			"name":"Broken test",
			"status":"draft",
			"is_bandit":false,
			"min_sample_size":100,
			"confidence_threshold_percent":95,
			"arms":[{"name":"Only arm","is_control":true,"traffic_weight":1}]
		}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}
