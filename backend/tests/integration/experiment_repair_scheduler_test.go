package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	worker_tasks "github.com/bivex/paywall-iap/internal/worker/tasks"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestExperimentRepairScheduledTaskRepairsCandidatesAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	for _, statement := range []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
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
		`CREATE TABLE automation_job_run_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			job_name TEXT NOT NULL,
			source TEXT NOT NULL,
			idempotency_key TEXT NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
			payload JSONB,
			details JSONB,
			window_started_at TIMESTAMPTZ NOT NULL,
			window_duration_seconds INTEGER NOT NULL CHECK (window_duration_seconds > 0),
			started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			finished_at TIMESTAMPTZ
		)`,
		`CREATE UNIQUE INDEX idx_automation_job_run_log_idempotency ON automation_job_run_log(idempotency_key)`,
	} {
		_, err = db.Exec(ctx, statement)
		require.NoError(t, err)
	}

	experimentID := uuid.New()
	controlArmID := uuid.New()
	variantArmID := uuid.New()
	userID := uuid.New()

	_, err = db.Exec(ctx, `INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role) VALUES ($1, 'user-a', 'device-a', 'ios', '1.0.0', 'a@example.com', 'user')`, userID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, name, description, status, algorithm_type, is_bandit, min_sample_size, confidence_threshold, automation_policy) VALUES ($1, 'Repairable experiment', 'Needs repair', 'running', 'thompson_sampling', TRUE, 100, 0.95, '{"enabled": true, "auto_start": true, "auto_complete": true, "complete_on_end_time": true, "complete_on_sample_size": true, "complete_on_confidence": true, "manual_override": false, "locked_until": null, "locked_by": null, "lock_reason": null}'::jsonb)`, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight) VALUES ($1, $3, 'Control', 'Baseline', TRUE, 1.0), ($2, $3, 'Variant', 'Variant', FALSE, 1.0)`, controlArmID, variantArmID, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward) VALUES ($1, 20, 2, 18, 16, 320, 17.7778)`, variantArmID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_assignments (experiment_id, user_id, arm_id, expires_at) VALUES ($1, $2, $3, now() - interval '12 hours')`, experimentID, userID, controlArmID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO bandit_pending_rewards (id, experiment_id, arm_id, user_id, assigned_at, expires_at) VALUES (gen_random_uuid(), $1, $2, $3, now() - interval '2 days', now() - interval '1 day')`, experimentID, controlArmID, userID)
	require.NoError(t, err)

	experimentRepo := repository.NewExperimentAdminRepository(db)
	banditRepo := repository.NewPostgresBanditRepository(db, zap.NewNop())
	repairService := service.NewExperimentRepairService(experimentRepo, banditRepo)
	repairReconciler := service.NewExperimentRepairReconciler(experimentRepo, repairService)
	executor := service.NewAutomationJobExecutionService(repository.NewAutomationJobRunRepository(db))

	mux := asynq.NewServeMux()
	worker_tasks.RegisterExperimentRepairTasks(mux, repairReconciler, executor, zap.NewNop())
	task := asynq.NewTask(worker_tasks.TypeReconcileExperimentRepair, mustMarshalIntegrationJSON(worker_tasks.ReconcileExperimentRepairPayload{Limit: 10}))

	require.NoError(t, mux.ProcessTask(ctx, task))
	require.NoError(t, mux.ProcessTask(ctx, task))

	var statsRows, processedPending, runCount int
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(*)::int FROM ab_test_arm_stats WHERE arm_id = $1`, controlArmID).Scan(&statsRows))
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(*)::int FROM bandit_pending_rewards WHERE experiment_id = $1 AND processed_at IS NOT NULL`, experimentID).Scan(&processedPending))
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(*)::int FROM automation_job_run_log WHERE job_name = $1`, worker_tasks.TypeReconcileExperimentRepair).Scan(&runCount))
	assert.Equal(t, 1, statsRows)
	assert.Equal(t, 1, processedPending)
	assert.Equal(t, 1, runCount)

	var status string
	var detailsJSON []byte
	require.NoError(t, db.QueryRow(ctx, `SELECT status, details FROM automation_job_run_log WHERE job_name = $1`, worker_tasks.TypeReconcileExperimentRepair).Scan(&status, &detailsJSON))
	assert.Equal(t, service.AutomationJobRunStatusCompleted, status)

	var details map[string]any
	require.NoError(t, json.Unmarshal(detailsJSON, &details))
	assert.Equal(t, float64(1), details["scanned"])
	assert.Equal(t, float64(1), details["repaired"])
	assert.Equal(t, float64(1), details["missing_arm_stats_inserted"])
	assert.Equal(t, float64(1), details["pending_rewards_processed"])
	assert.Equal(t, float64(1), details["expired_pending_rewards"])
}

func mustMarshalIntegrationJSON(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return payload
}
