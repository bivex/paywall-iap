package integration

// Edge-case and boundary tests for experiment creation/update validation.
// Covers gaps not exercised by TestExperimentMultitenancy.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	httpmiddleware "github.com/bivex/paywall-iap/internal/interfaces/http/middleware"
	"github.com/bivex/paywall-iap/tests/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExperimentEdgeCases(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, experimentsMultitenancySchema)
	require.NoError(t, err)

	appID := uuid.New()
	_, err = db.Exec(ctx,
		`INSERT INTO apps (id, name, display_name, bundle_id, platform) VALUES ($1,$2,$3,$4,$5)`,
		appID, "Edge Case App", "Edge Case App", "com.edge.test", "ios",
	)
	require.NoError(t, err)

	h := handlers.NewAdminHandler(
		nil, nil,
		generated.New(db),
		db,
		nil, nil,
		service.NewAuditService(db),
		nil, nil, nil, nil, nil,
	)

	newRouter := func() *gin.Engine {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(httpmiddleware.RequireAppID())
		g := r.Group("/v1/admin")
		g.POST("/experiments", h.CreateAdminExperiment)
		g.PUT("/experiments/:id", h.UpdateAdminExperiment)
		g.POST("/experiments/:id/resume", h.ResumeAdminExperiment)
		return r
	}

	baseArms := json.RawMessage(`[
		{"name":"Control",   "description":"ctrl",  "is_control":true,  "traffic_weight":1},
		{"name":"Variant A", "description":"var a", "is_control":false, "traffic_weight":1}
	]`)

	// helper: POST /v1/admin/experiments with given name + overrides, return (statusCode, responseBody)
	create := func(t *testing.T, name string, overrides map[string]interface{}) (int, map[string]interface{}) {
		t.Helper()
		body := map[string]interface{}{
			"name":                         name,
			"description":                  "desc",
			"status":                       "draft",
			"algorithm_type":               "thompson_sampling",
			"is_bandit":                    true,
			"min_sample_size":              100,
			"confidence_threshold_percent": 95,
			"arms":                         baseArms,
		}
		for k, v := range overrides {
			body[k] = v
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appID.String())
		w := httptest.NewRecorder()
		newRouter().ServeHTTP(w, req)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		return w.Code, resp
	}

	// unique name helper to avoid 409 collisions between sub-tests
	uname := func(prefix string) string { return fmt.Sprintf("%s-%s", prefix, uuid.New().String()[:8]) }

	// ── Status edge cases ──────────────────────────────────────────────────────

	t.Run("CREATE status=completed → accepted (validator gap, not 500)", func(t *testing.T) {
		code, _ := create(t, uname("completed"), map[string]interface{}{"status": "completed"})
		// Validator currently allows this. Document: must not 500.
		assert.NotEqual(t, http.StatusInternalServerError, code)
	})

	t.Run("CREATE status=invalid → 422", func(t *testing.T) {
		code, _ := create(t, uname("bad-status"), map[string]interface{}{"status": "banana"})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	// ── Arm validation ─────────────────────────────────────────────────────────

	t.Run("CREATE duplicate arm names → allowed (no unique DB constraint)", func(t *testing.T) {
		arms := json.RawMessage(`[
			{"name":"Control","description":"c","is_control":true, "traffic_weight":1},
			{"name":"Control","description":"d","is_control":false,"traffic_weight":1}
		]`)
		code, _ := create(t, uname("dup-arms"), map[string]interface{}{"arms": arms})
		assert.NotEqual(t, http.StatusInternalServerError, code)
	})

	t.Run("CREATE zero traffic_weight → 422", func(t *testing.T) {
		arms := json.RawMessage(`[
			{"name":"Control",  "description":"c","is_control":true, "traffic_weight":0},
			{"name":"Variant A","description":"v","is_control":false,"traffic_weight":1}
		]`)
		code, _ := create(t, uname("zero-weight"), map[string]interface{}{"arms": arms})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE negative traffic_weight → 422", func(t *testing.T) {
		arms := json.RawMessage(`[
			{"name":"Control",  "description":"c","is_control":true, "traffic_weight":-1},
			{"name":"Variant A","description":"v","is_control":false,"traffic_weight":1}
		]`)
		code, _ := create(t, uname("neg-weight"), map[string]interface{}{"arms": arms})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE two control arms → 422", func(t *testing.T) {
		arms := json.RawMessage(`[
			{"name":"Control A","description":"c1","is_control":true,"traffic_weight":1},
			{"name":"Control B","description":"c2","is_control":true,"traffic_weight":1}
		]`)
		code, _ := create(t, uname("two-ctrl"), map[string]interface{}{"arms": arms})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE one arm only → 422", func(t *testing.T) {
		arms := json.RawMessage(`[{"name":"Control","description":"c","is_control":true,"traffic_weight":1}]`)
		code, _ := create(t, uname("one-arm"), map[string]interface{}{"arms": arms})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE no arms → 422", func(t *testing.T) {
		code, _ := create(t, uname("no-arms"), map[string]interface{}{"arms": json.RawMessage(`[]`)})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	// ── Confidence threshold boundaries ───────────────────────────────────────

	t.Run("CREATE confidence_threshold=0 → 422", func(t *testing.T) {
		code, _ := create(t, uname("ct-0"), map[string]interface{}{"confidence_threshold_percent": 0})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE confidence_threshold=100 → 201 (boundary allowed)", func(t *testing.T) {
		code, _ := create(t, uname("ct-100"), map[string]interface{}{"confidence_threshold_percent": 100})
		assert.Equal(t, http.StatusCreated, code)
	})

	t.Run("CREATE confidence_threshold=101 → 422", func(t *testing.T) {
		code, _ := create(t, uname("ct-101"), map[string]interface{}{"confidence_threshold_percent": 101})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	// ── Date edge cases ────────────────────────────────────────────────────────

	t.Run("CREATE end_at before start_at → 422", func(t *testing.T) {
		code, _ := create(t, uname("bad-dates"), map[string]interface{}{
			"start_at": "2030-01-02T00:00:00Z",
			"end_at":   "2030-01-01T00:00:00Z",
		})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE start_at = end_at → allowed (not strictly before)", func(t *testing.T) {
		code, _ := create(t, uname("equal-dates"), map[string]interface{}{
			"start_at": "2030-01-01T00:00:00Z",
			"end_at":   "2030-01-01T00:00:00Z",
		})
		assert.NotEqual(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE end_at without start_at → 201", func(t *testing.T) {
		code, _ := create(t, uname("no-start"), map[string]interface{}{
			"end_at": "2030-12-31T00:00:00Z",
		})
		assert.Equal(t, http.StatusCreated, code)
	})

	// ── min_sample_size boundaries ─────────────────────────────────────────────

	t.Run("CREATE min_sample_size=0 → 422", func(t *testing.T) {
		code, _ := create(t, uname("mss-0"), map[string]interface{}{"min_sample_size": 0})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE min_sample_size=1 → 201", func(t *testing.T) {
		code, _ := create(t, uname("mss-1"), map[string]interface{}{"min_sample_size": 1})
		assert.Equal(t, http.StatusCreated, code)
	})

	t.Run("CREATE min_sample_size > max int32 → 422", func(t *testing.T) {
		code, _ := create(t, uname("mss-overflow"), map[string]interface{}{"min_sample_size": 2147483648})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	// ── Algorithm types ────────────────────────────────────────────────────────

	t.Run("CREATE algorithm_type=ucb → 201", func(t *testing.T) {
		code, _ := create(t, uname("ucb"), map[string]interface{}{"algorithm_type": "ucb"})
		assert.Equal(t, http.StatusCreated, code)
	})

	t.Run("CREATE algorithm_type=epsilon_greedy → 201", func(t *testing.T) {
		code, _ := create(t, uname("eg"), map[string]interface{}{"algorithm_type": "epsilon_greedy"})
		assert.Equal(t, http.StatusCreated, code)
	})

	t.Run("CREATE algorithm_type=unknown → 422", func(t *testing.T) {
		code, _ := create(t, uname("bad-algo"), map[string]interface{}{"algorithm_type": "random_forest"})
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	// ── Name validation ────────────────────────────────────────────────────────

	t.Run("CREATE empty name → 422", func(t *testing.T) {
		code, _ := create(t, "", nil)
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE name with null byte → 422", func(t *testing.T) {
		code, _ := create(t, "bad\x00name", nil)
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE name with control char → 422", func(t *testing.T) {
		code, _ := create(t, "bad\x01name", nil)
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})

	t.Run("CREATE very long name (1001 chars) → not 500", func(t *testing.T) {
		code, _ := create(t, strings.Repeat("x", 1001), nil)
		assert.NotEqual(t, http.StatusInternalServerError, code)
	})

	// ── Duplicate name ─────────────────────────────────────────────────────────

	t.Run("CREATE duplicate name in same app → allowed (no unique DB constraint in test schema)", func(t *testing.T) {
		name := uname("dup-name")
		code1, _ := create(t, name, nil)
		require.Equal(t, http.StatusCreated, code1)
		code2, _ := create(t, name, nil)
		// No UNIQUE(app_id, name) constraint exists — second insert succeeds.
		// This documents a known gap: production may want to enforce uniqueness.
		assert.Equal(t, http.StatusCreated, code2)
	})

	t.Run("CREATE same name in two apps → 201 each", func(t *testing.T) {
		appB := uuid.New()
		_, err := db.Exec(ctx,
			`INSERT INTO apps (id, name, display_name, bundle_id, platform) VALUES ($1,$2,$3,$4,$5)`,
			appB, "Edge App B", "Edge App B", "com.edge.testb", "android",
		)
		require.NoError(t, err)

		name := uname("cross-app")
		makeReq := func(app uuid.UUID) int {
			body := map[string]interface{}{
				"name":                         name,
				"description":                  "desc",
				"status":                       "draft",
				"algorithm_type":               "thompson_sampling",
				"is_bandit":                    false,
				"min_sample_size":              100,
				"confidence_threshold_percent": 95,
				"arms":                         baseArms,
			}
			b, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-App-ID", app.String())
			w := httptest.NewRecorder()
			newRouter().ServeHTTP(w, req)
			return w.Code
		}
		assert.Equal(t, http.StatusCreated, makeReq(appID))
		assert.Equal(t, http.StatusCreated, makeReq(appB))
	})

	// ── UPDATE arms edge cases ─────────────────────────────────────────────────

	t.Run("UPDATE arms=[] → 422", func(t *testing.T) {
		code, resp := create(t, uname("update-arms"), nil)
		require.Equal(t, http.StatusCreated, code)
		expID := resp["data"].(map[string]interface{})["id"].(string)

		body := map[string]interface{}{
			"name":                         "Updated",
			"description":                  "desc",
			"status":                       "draft",
			"algorithm_type":               "thompson_sampling",
			"is_bandit":                    true,
			"min_sample_size":              100,
			"confidence_threshold_percent": 95,
			"arms":                         []interface{}{},
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+expID, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appID.String())
		w := httptest.NewRecorder()
		newRouter().ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("UPDATE arms=null → arms unchanged", func(t *testing.T) {
		code, resp := create(t, uname("update-null-arms"), nil)
		require.Equal(t, http.StatusCreated, code)
		data := resp["data"].(map[string]interface{})
		expID := data["id"].(string)
		origArms := data["arms"].([]interface{})

		body := map[string]interface{}{
			"name":                         "Updated Name",
			"description":                  "desc",
			"status":                       "draft",
			"algorithm_type":               "thompson_sampling",
			"is_bandit":                    true,
			"min_sample_size":              100,
			"confidence_threshold_percent": 95,
			// arms intentionally omitted → null in JSON
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+expID, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-ID", appID.String())
		w := httptest.NewRecorder()
		newRouter().ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		var upd struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &upd))
		assert.Len(t, upd.Data.Arms, len(origArms), "arms unchanged when null")
	})
}
