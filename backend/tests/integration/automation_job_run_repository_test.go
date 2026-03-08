package integration

import (
	"context"
	"testing"
	"time"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutomationJobRunRepository(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE TABLE automation_job_run_log (
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
		);

		CREATE UNIQUE INDEX idx_automation_job_run_log_idempotency
			ON automation_job_run_log(idempotency_key);
	`)
	require.NoError(t, err)

	repo := repository.NewAutomationJobRunRepository(db)
	claim := service.AutomationJobRunClaimInput{
		JobName:         "bandit:maintenance:process_expired",
		Source:          "asynq_scheduler",
		IdempotencyKey:  "bandit:maintenance:process_expired:2026-03-08T12:00:00Z",
		Payload:         []byte(`{"batch_size":100}`),
		WindowStartedAt: time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC),
		WindowDuration:  15 * time.Minute,
	}

	t.Run("duplicate completed run is skipped", func(t *testing.T) {
		run, claimed, err := repo.ClaimAutomationJobRun(ctx, claim)
		require.NoError(t, err)
		require.True(t, claimed)
		require.NoError(t, repo.FinishAutomationJobRun(ctx, run.ID, service.AutomationJobRunStatusCompleted, map[string]any{"processed": 12}))

		run, claimed, err = repo.ClaimAutomationJobRun(ctx, claim)
		require.NoError(t, err)
		assert.False(t, claimed)
		assert.Nil(t, run)
	})

	t.Run("failed run can be claimed again for retry", func(t *testing.T) {
		retryClaim := claim
		retryClaim.IdempotencyKey = "bandit:maintenance:process_expired:2026-03-08T12:15:00Z"
		retryClaim.WindowStartedAt = time.Date(2026, 3, 8, 12, 15, 0, 0, time.UTC)

		run, claimed, err := repo.ClaimAutomationJobRun(ctx, retryClaim)
		require.NoError(t, err)
		require.True(t, claimed)
		firstRunID := run.ID
		require.NoError(t, repo.FinishAutomationJobRun(ctx, run.ID, service.AutomationJobRunStatusFailed, map[string]any{"error": "boom"}))

		run, claimed, err = repo.ClaimAutomationJobRun(ctx, retryClaim)
		require.NoError(t, err)
		require.True(t, claimed)
		assert.Equal(t, firstRunID, run.ID)
	})
}
