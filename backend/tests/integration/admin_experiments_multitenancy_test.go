package integration

// Tests for experiment (ab_tests) multi-tenancy (app_id isolation).
// Each app sees only its own experiments; X-App-ID header is required.

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
	httpmiddleware "github.com/bivex/paywall-iap/internal/interfaces/http/middleware"
	service "github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/tests/testutil"
)

const experimentsMultitenancySchema = `
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
	deleted_at     TIMESTAMPTZ
);

CREATE TABLE ab_tests (
	id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	app_id               UUID NOT NULL REFERENCES apps(id),
	name                 TEXT NOT NULL,
	description          TEXT,
	status               TEXT NOT NULL CHECK (status IN ('draft','running','paused','completed')) DEFAULT 'draft',
	start_at             TIMESTAMPTZ,
	end_at               TIMESTAMPTZ,
	algorithm_type       TEXT CHECK (algorithm_type IN ('thompson_sampling','ucb','epsilon_greedy')),
	is_bandit            BOOLEAN NOT NULL DEFAULT false,
	min_sample_size      INT DEFAULT 100,
	confidence_threshold NUMERIC(3,2) DEFAULT 0.95,
	winner_confidence    NUMERIC(3,2),
	created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
	automation_policy    JSONB NOT NULL DEFAULT '{"enabled":false,"auto_start":false,"auto_complete":false,"complete_on_end_time":true,"complete_on_sample_size":false,"complete_on_confidence":false,"manual_override":false}'::jsonb,
	CONSTRAINT ab_tests_automation_policy_is_object CHECK (jsonb_typeof(automation_policy) = 'object')
);

CREATE INDEX idx_ab_tests_active ON ab_tests(status) WHERE status = 'running';

CREATE TABLE ab_test_arms (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	experiment_id   UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
	name            TEXT NOT NULL,
	description     TEXT,
	is_control      BOOLEAN NOT NULL DEFAULT false,
	traffic_weight  NUMERIC(3,2) NOT NULL DEFAULT 1.0,
	pricing_tier_id UUID REFERENCES pricing_tiers(id),
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ab_test_arm_stats (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	arm_id      UUID NOT NULL UNIQUE REFERENCES ab_test_arms(id) ON DELETE CASCADE,
	alpha       NUMERIC(10,2) NOT NULL DEFAULT 1.0,
	beta        NUMERIC(10,2) NOT NULL DEFAULT 1.0,
	samples     INT NOT NULL DEFAULT 0,
	conversions INT NOT NULL DEFAULT 0,
	revenue     NUMERIC(15,2) NOT NULL DEFAULT 0.0,
	avg_reward  NUMERIC(10,4),
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	CHECK (alpha > 0),
	CHECK (beta > 0),
	CHECK (samples >= 0),
	CHECK (conversions >= 0),
	CHECK (conversions <= samples)
);

CREATE TABLE ab_test_assignments (
	id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
	user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	arm_id        UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
	assigned_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
	expires_at    TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours')
);

CREATE TABLE experiment_lifecycle_audit_log (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	experiment_id   UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
	actor_type      TEXT NOT NULL,
	actor_id        UUID,
	source          TEXT NOT NULL,
	action          TEXT NOT NULL,
	from_status     TEXT NOT NULL,
	to_status       TEXT NOT NULL,
	idempotency_key TEXT,
	details         JSONB,
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_experiment_lifecycle_audit_idempotency
	ON experiment_lifecycle_audit_log(idempotency_key);

CREATE TABLE experiment_winner_recommendation_log (
	id                           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	experiment_id                UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
	source                       TEXT NOT NULL,
	recommended                  BOOLEAN NOT NULL DEFAULT FALSE,
	reason                       TEXT NOT NULL,
	winning_arm_id               UUID REFERENCES ab_test_arms(id) ON DELETE SET NULL,
	confidence_percent           DOUBLE PRECISION,
	confidence_threshold_percent DOUBLE PRECISION NOT NULL,
	observed_samples             INT NOT NULL,
	min_sample_size              INT NOT NULL,
	details                      JSONB,
	occurred_at                  TIMESTAMPTZ NOT NULL,
	created_at                   TIMESTAMPTZ NOT NULL DEFAULT now()
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
`

func TestExperimentMultitenancy(t *testing.T) {
	ctx := context.Background()
	pool, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = pool.Exec(ctx, experimentsMultitenancySchema)
	require.NoError(t, err)

	appA := uuid.New()
	appB := uuid.New()
	adminID := uuid.New()

	_, err = pool.Exec(ctx, `
		INSERT INTO apps (id, name, display_name, platform, bundle_id) VALUES
			($1, 'com.mt.exp_a', 'Exp App A', 'ios',     'com.mt.exp_a'),
			($2, 'com.mt.exp_b', 'Exp App B', 'android', 'com.mt.exp_b')
	`, appA, appB)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role)
		 VALUES ($1, 'admin-exp-mt', 'dev-exp-mt', 'ios', '1.0.0', 'admin-exp-mt@test.com', 'admin')`,
		adminID,
	)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)

	newRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("admin_id", adminID)
			c.Set("user_id", adminID.String())
			c.Next()
		})
		r.Use(httpmiddleware.RequireAppID())
		h := handlers.NewAdminHandler(
			nil, nil,
			generated.New(pool),
			pool,
			nil, nil,
			service.NewAuditService(pool),
			nil, nil, nil, nil, nil,
		)
		g := r.Group("/v1/admin")
		g.GET("/experiments", h.ListAdminExperiments)
		g.POST("/experiments", h.CreateAdminExperiment)
		g.PUT("/experiments/:id", h.UpdateAdminExperiment)
		g.POST("/experiments/:id/pause", h.PauseAdminExperiment)
		g.POST("/experiments/:id/resume", h.ResumeAdminExperiment)
		g.POST("/experiments/:id/complete", h.CompleteAdminExperiment)
		g.GET("/experiments/:id/lifecycle-audit", h.GetAdminExperimentLifecycleAuditHistory)
		return r
	}

	var expAID string

	t.Run("no X-App-ID → 400", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET app_a returns empty before creation", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Empty(t, resp.Data)
	})

	t.Run("POST creates experiment for app_a", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{
			"name":        "CTA Test App A",
			"description": "Paywall CTA comparison for App A",
			"status":      "draft",
			"algorithm_type": "thompson_sampling",
			"is_bandit":   true,
			"min_sample_size": 200,
			"confidence_threshold_percent": 95,
			"arms": [
				{"name":"Control",   "description":"Original CTA",  "is_control":true,  "traffic_weight":1},
				{"name":"Variant A", "description":"New CTA copy",   "is_control":false, "traffic_weight":1}
			]
		}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "CTA Test App A", resp.Data.Name)
		assert.True(t, resp.Data.IsBandit)
		assert.Len(t, resp.Data.Arms, 2)
		expAID = resp.Data.ID.String()
	})

	t.Run("POST creates experiment for app_b", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{
			"name":        "Price Test App B",
			"description": "Price point test for App B",
			"status":      "draft",
			"algorithm_type": "ucb",
			"is_bandit":   false,
			"min_sample_size": 300,
			"confidence_threshold_percent": 90,
			"arms": [
				{"name":"Control",   "description":"$4.99",  "is_control":true,  "traffic_weight":1},
				{"name":"Variant B", "description":"$7.99",  "is_control":false, "traffic_weight":1}
			]
		}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appB.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GET app_a sees only its own experiment", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "CTA Test App A", resp.Data[0].Name)
	})

	t.Run("GET app_b sees only its own experiment", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		req.Header.Set("X-App-ID", appB.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "Price Test App B", resp.Data[0].Name)
	})

	t.Run("PUT update name of app_a experiment (while draft)", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{
			"name":        "CTA Test App A — Renamed",
			"description": "Updated description",
			"status":      "draft",
			"algorithm_type": "thompson_sampling",
			"is_bandit":   true,
			"min_sample_size": 200,
			"confidence_threshold_percent": 95,
			"arms": [
				{"name":"Control",   "description":"Original CTA",  "is_control":true,  "traffic_weight":1},
				{"name":"Variant A", "description":"New CTA copy",   "is_control":false, "traffic_weight":1}
			]
		}`)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+expAID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "CTA Test App A — Renamed", resp.Data.Name)
		expAID = resp.Data.ID.String()
	})

	t.Run("POST start (draft→running) app_a experiment", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+expAID+"/resume", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "running", resp.Data.Status)
	})

	t.Run("POST pause app_a experiment", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+expAID+"/pause", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "paused", resp.Data.Status)
	})

	t.Run("POST resume app_a experiment", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+expAID+"/resume", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "running", resp.Data.Status)
	})

	t.Run("GET lifecycle-audit returns pause+resume events", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments/"+expAID+"/lifecycle-audit", nil)
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		// pause + resume = 2 audit entries
		assert.GreaterOrEqual(t, len(resp.Data), 2)
	})

	t.Run("POST complete app_a experiment", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+expAID+"/complete", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "completed", resp.Data.Status)
	})

	t.Run("POST create with missing arms → 422", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{
			"name":        "No Arms",
			"description": "Should fail",
			"status":      "draft",
			"algorithm_type": "ucb",
			"is_bandit":   false,
			"min_sample_size": 100,
			"confidence_threshold_percent": 95,
			"arms": []
		}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("POST create with no control arm → 422", func(t *testing.T) {
		r := newRouter()
		body := []byte(`{
			"name":        "No Control",
			"description": "Should fail — no control arm",
			"status":      "draft",
			"algorithm_type": "ucb",
			"is_bandit":   false,
			"min_sample_size": 100,
			"confidence_threshold_percent": 95,
			"arms": [
				{"name":"A","description":"Arm A","is_control":false,"traffic_weight":1},
				{"name":"B","description":"Arm B","is_control":false,"traffic_weight":1}
			]
		}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("POST pause unknown experiment → 404", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+uuid.New().String()+"/pause", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appA.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
