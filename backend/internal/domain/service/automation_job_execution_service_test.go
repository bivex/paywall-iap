package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAutomationJobRunRepository struct {
	claimed            map[string]*AutomationJobRun
	status             map[uuid.UUID]string
	details            map[uuid.UUID]map[string]any
	allowRetryOnFailed bool
}

func (r *stubAutomationJobRunRepository) ClaimAutomationJobRun(_ context.Context, input AutomationJobRunClaimInput) (*AutomationJobRun, bool, error) {
	if r.claimed == nil {
		r.claimed = make(map[string]*AutomationJobRun)
	}
	if r.status == nil {
		r.status = make(map[uuid.UUID]string)
	}
	if existing, ok := r.claimed[input.IdempotencyKey]; ok {
		if r.allowRetryOnFailed && r.status[existing.ID] == AutomationJobRunStatusFailed {
			r.status[existing.ID] = AutomationJobRunStatusRunning
			return existing, true, nil
		}
		return nil, false, nil
	}
	run := &AutomationJobRun{ID: uuid.New(), JobName: input.JobName, IdempotencyKey: input.IdempotencyKey, Status: AutomationJobRunStatusRunning}
	r.claimed[input.IdempotencyKey] = run
	r.status[run.ID] = AutomationJobRunStatusRunning
	return run, true, nil
}

func (r *stubAutomationJobRunRepository) FinishAutomationJobRun(_ context.Context, runID uuid.UUID, status string, details map[string]any) error {
	if r.status == nil {
		r.status = make(map[uuid.UUID]string)
	}
	if r.details == nil {
		r.details = make(map[uuid.UUID]map[string]any)
	}
	r.status[runID] = status
	r.details[runID] = details
	return nil
}

func TestAutomationJobExecutionService(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 3, 8, 12, 7, 0, 0, time.UTC)
	spec := ScheduledAutomationJobSpec{
		JobName: "bandit:maintenance:process_expired",
		Source:  "asynq_scheduler",
		Window:  15 * time.Minute,
	}
	payload := []byte(`{"batch_size":100}`)

	t.Run("executes once per scheduled window", func(t *testing.T) {
		repo := &stubAutomationJobRunRepository{}
		svc := NewAutomationJobExecutionService(repo)
		svc.now = func() time.Time { return now }

		executions := 0
		executed, err := svc.ExecuteScheduled(ctx, spec, payload, func(context.Context) (map[string]any, error) {
			executions++
			return map[string]any{"processed": 12}, nil
		})
		require.NoError(t, err)
		assert.True(t, executed)
		assert.Equal(t, 1, executions)

		executed, err = svc.ExecuteScheduled(ctx, spec, payload, func(context.Context) (map[string]any, error) {
			executions++
			return map[string]any{"processed": 99}, nil
		})
		require.NoError(t, err)
		assert.False(t, executed)
		assert.Equal(t, 1, executions)
	})

	t.Run("failed runs can be retried with the same idempotency key", func(t *testing.T) {
		repo := &stubAutomationJobRunRepository{allowRetryOnFailed: true}
		svc := NewAutomationJobExecutionService(repo)
		svc.now = func() time.Time { return now }

		executions := 0
		err := errors.New("boom")
		executed, runErr := svc.ExecuteScheduled(ctx, spec, payload, func(context.Context) (map[string]any, error) {
			executions++
			return map[string]any{"phase": "first_attempt"}, err
		})
		assert.True(t, executed)
		require.ErrorIs(t, runErr, err)

		executed, runErr = svc.ExecuteScheduled(ctx, spec, payload, func(context.Context) (map[string]any, error) {
			executions++
			return map[string]any{"phase": "retry"}, nil
		})
		require.NoError(t, runErr)
		assert.True(t, executed)
		assert.Equal(t, 2, executions)
	})
}
