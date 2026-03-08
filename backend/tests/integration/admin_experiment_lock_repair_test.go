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

	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestAdminExperimentLockAndRepairHandlers(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	for _, statement := range []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE pricing_tiers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			monthly_price NUMERIC(10,2),
			annual_price NUMERIC(10,2),
			currency CHAR(3) NOT NULL DEFAULT 'USD',
			features JSONB,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE users (
			id UUID PRIMARY KEY,
			platform_user_id TEXT UNIQUE NOT NULL,
			device_id TEXT,
			platform TEXT NOT NULL,
			app_version TEXT NOT NULL,
			email TEXT UNIQUE,
			role TEXT NOT NULL DEFAULT 'user',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE ab_tests (
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
			winner_confidence NUMERIC(5,4),
			objective_type TEXT,
			objective_weights JSONB,
			automation_policy JSONB NOT NULL DEFAULT '{"enabled": false, "auto_start": false, "auto_complete": false, "complete_on_end_time": true, "complete_on_sample_size": false, "complete_on_confidence": false, "manual_override": false, "locked_until": null, "locked_by": null, "lock_reason": null}'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE ab_test_arms (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			description TEXT,
			is_control BOOLEAN NOT NULL DEFAULT false,
			traffic_weight NUMERIC(3,2) NOT NULL DEFAULT 1.0,
			pricing_tier_id UUID REFERENCES pricing_tiers(id),
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE ab_test_arm_stats (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			arm_id UUID NOT NULL UNIQUE REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			alpha NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			beta NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			samples INT NOT NULL DEFAULT 0,
			conversions INT NOT NULL DEFAULT 0,
			revenue NUMERIC(15,2) NOT NULL DEFAULT 0.0,
			avg_reward NUMERIC(10,4),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE ab_test_assignments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours')
		)`,
		`CREATE TABLE experiment_lifecycle_audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			actor_type TEXT NOT NULL,
			actor_id UUID,
			source TEXT NOT NULL,
			action TEXT NOT NULL,
			from_status TEXT NOT NULL,
			to_status TEXT NOT NULL,
			idempotency_key TEXT,
			details JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE UNIQUE INDEX idx_experiment_lifecycle_audit_log_idempotency ON experiment_lifecycle_audit_log(idempotency_key)`,
		`CREATE TABLE bandit_pending_rewards (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			assigned_at TIMESTAMPTZ NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			converted BOOLEAN NOT NULL DEFAULT FALSE,
			conversion_value NUMERIC(15,2) NOT NULL DEFAULT 0,
			conversion_currency TEXT,
			converted_at TIMESTAMPTZ,
			processed_at TIMESTAMPTZ
		)`,
		`CREATE TABLE bandit_arm_objective_stats (
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			objective_type TEXT NOT NULL,
			alpha NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			beta NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			samples INT NOT NULL DEFAULT 0,
			conversions INT NOT NULL DEFAULT 0,
			total_revenue NUMERIC(15,2) NOT NULL DEFAULT 0.0,
			avg_ltv NUMERIC(10,4) NOT NULL DEFAULT 0.0,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			PRIMARY KEY (arm_id, objective_type)
		)`,
	} {
		_, err = db.Exec(ctx, statement)
		require.NoError(t, err)
	}

	adminID := uuid.New()
	controlArmID := uuid.New()
	variantArmID := uuid.New()
	experimentID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()
	userC := uuid.New()

	_, err = db.Exec(ctx, `INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role)
		VALUES
		($1, 'admin-user', 'admin-device', 'ios', '1.0.0', 'admin@example.com', 'admin'),
		($2, 'user-a', 'device-a', 'ios', '1.0.0', 'a@example.com', 'user'),
		($3, 'user-b', 'device-b', 'ios', '1.0.0', 'b@example.com', 'user'),
		($4, 'user-c', 'device-c', 'ios', '1.0.0', 'c@example.com', 'user')`, adminID, userA, userB, userC)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, name, description, status, algorithm_type, is_bandit, min_sample_size, confidence_threshold, objective_type, objective_weights, automation_policy)
		VALUES ($1, 'Repairable experiment', 'Needs repair', 'running', 'thompson_sampling', TRUE, 100, 0.95, 'hybrid', '{"conversion":0.7,"ltv":0.3}'::jsonb, '{"enabled": true, "auto_start": true, "auto_complete": true, "complete_on_end_time": true, "complete_on_sample_size": true, "complete_on_confidence": true, "manual_override": false, "locked_until": null, "locked_by": null, "lock_reason": null}'::jsonb);
	`, experimentID)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight)
		VALUES
		($2, $1, 'Control', 'Baseline', TRUE, 1.0),
		($3, $1, 'Variant', 'Variant', FALSE, 1.0)`, experimentID, controlArmID, variantArmID)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward)
		VALUES ($1, 20, 2, 18, 16, 320, 17.7778)`, variantArmID)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO ab_test_assignments (experiment_id, user_id, arm_id, expires_at)
		VALUES
		($1, $2, $3, now() + interval '12 hours'),
		($1, $4, $5, now() - interval '12 hours')`, experimentID, userA, controlArmID, userB, variantArmID)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO bandit_pending_rewards (id, experiment_id, arm_id, user_id, assigned_at, expires_at)
		VALUES
		(gen_random_uuid(), $1, $2, $4, now() - interval '2 days', now() - interval '1 day'),
		(gen_random_uuid(), $1, $3, $5, now() - interval '1 day', now() + interval '1 day')`, experimentID, controlArmID, variantArmID, userA, userC)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Set("user_id", adminID.String())
		c.Next()
	})

	handler := handlers.NewAdminHandler(nil, nil, generated.New(db), db, nil, nil, nil, nil, nil, nil, nil, nil)
	admin := router.Group("/v1/admin")
	admin.POST("/experiments/:id/lock", handler.LockAdminExperiment)
	admin.POST("/experiments/:id/unlock", handler.UnlockAdminExperiment)
	admin.POST("/experiments/:id/repair", handler.RepairAdminExperiment)

	t.Run("lock stores manual override metadata", func(t *testing.T) {
		body := []byte(`{"reason":"Investigating anomaly"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+experimentID.String()+"/lock", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Data.AutomationPolicy.ManualOverride)
		assert.Nil(t, resp.Data.AutomationPolicy.LockedUntil)
		require.NotNil(t, resp.Data.AutomationPolicy.LockedBy)
		assert.Equal(t, adminID, *resp.Data.AutomationPolicy.LockedBy)
		require.NotNil(t, resp.Data.AutomationPolicy.LockReason)
		assert.Equal(t, "Investigating anomaly", *resp.Data.AutomationPolicy.LockReason)
	})

	t.Run("unlock clears lock metadata", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+experimentID.String()+"/unlock", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp.Data.AutomationPolicy.ManualOverride)
		assert.Nil(t, resp.Data.AutomationPolicy.LockedUntil)
		assert.Nil(t, resp.Data.AutomationPolicy.LockedBy)
		assert.Nil(t, resp.Data.AutomationPolicy.LockReason)
	})

	t.Run("lock stores timed lock metadata", func(t *testing.T) {
		lockedUntil := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
		body := []byte(fmt.Sprintf(`{"locked_until":%q,"reason":"Freeze auto-complete"}`, lockedUntil))
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+experimentID.String()+"/lock", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp.Data.AutomationPolicy.ManualOverride)
		require.NotNil(t, resp.Data.AutomationPolicy.LockedUntil)
		require.NotNil(t, resp.Data.AutomationPolicy.LockReason)
		assert.Equal(t, "Freeze auto-complete", *resp.Data.AutomationPolicy.LockReason)
	})

	t.Run("repair restores missing stats row and recomputes derived state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+experimentID.String()+"/repair", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data struct {
				Experiment handlers.AdminExperiment `json:"experiment"`
				Summary    struct {
					AssignmentSnapshot struct {
						Total  int `json:"total"`
						Active int `json:"active"`
					} `json:"assignment_snapshot"`
					MissingArmStatsInserted int      `json:"missing_arm_stats_inserted"`
					ObjectiveStatsSynced    int      `json:"objective_stats_synced"`
					PendingRewardsTotal     int      `json:"pending_rewards_total"`
					ExpiredPendingRewards   int      `json:"expired_pending_rewards"`
					PendingRewardsProcessed int      `json:"pending_rewards_processed"`
					WinnerConfidencePercent *float64 `json:"winner_confidence_percent"`
				} `json:"summary"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, 2, resp.Data.Summary.AssignmentSnapshot.Total)
		assert.Equal(t, 1, resp.Data.Summary.AssignmentSnapshot.Active)
		assert.Equal(t, 1, resp.Data.Summary.MissingArmStatsInserted)
		assert.Equal(t, 4, resp.Data.Summary.ObjectiveStatsSynced)
		assert.Equal(t, 2, resp.Data.Summary.PendingRewardsTotal)
		assert.Equal(t, 1, resp.Data.Summary.ExpiredPendingRewards)
		assert.Equal(t, 1, resp.Data.Summary.PendingRewardsProcessed)
		require.NotNil(t, resp.Data.Summary.WinnerConfidencePercent)
		assert.Greater(t, *resp.Data.Summary.WinnerConfidencePercent, 50.0)
		require.NotNil(t, resp.Data.Experiment.WinnerConfidencePercent)
		assert.Greater(t, *resp.Data.Experiment.WinnerConfidencePercent, 50.0)

		var controlSamples, controlBeta, pendingProcessed int
		err := db.QueryRow(ctx, `SELECT samples, beta::int FROM ab_test_arm_stats WHERE arm_id = $1`, controlArmID).Scan(&controlSamples, &controlBeta)
		require.NoError(t, err)
		assert.Equal(t, 1, controlSamples)
		assert.Equal(t, 2, controlBeta)

		err = db.QueryRow(ctx, `SELECT COUNT(*)::int FROM bandit_pending_rewards WHERE experiment_id = $1 AND processed_at IS NOT NULL`, experimentID).Scan(&pendingProcessed)
		require.NoError(t, err)
		assert.Equal(t, 1, pendingProcessed)

		var objectiveRows int
		err = db.QueryRow(ctx, `SELECT COUNT(*)::int FROM bandit_arm_objective_stats WHERE arm_id IN ($1, $2)`, controlArmID, variantArmID).Scan(&objectiveRows)
		require.NoError(t, err)
		assert.Equal(t, 4, objectiveRows)
	})
}
