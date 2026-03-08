package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

type fakeExperimentRepairRunner struct {
	lastLimit int
	result    service.ExperimentRepairRunResult
	err       error
}

func (f *fakeExperimentRepairRunner) Reconcile(_ context.Context, limit int) (service.ExperimentRepairRunResult, error) {
	f.lastLimit = limit
	return f.result, f.err
}

type fakeExperimentRepairExecutor struct {
	lastSpec    service.ScheduledAutomationJobSpec
	lastPayload []byte
	lastDetails map[string]any
	lastRunErr  error
	skip        bool
}

func (f *fakeExperimentRepairExecutor) ExecuteScheduled(ctx context.Context, spec service.ScheduledAutomationJobSpec, payload []byte, run func(context.Context) (map[string]any, error)) (bool, error) {
	f.lastSpec = spec
	f.lastPayload = payload
	if f.skip {
		return false, nil
	}
	details, err := run(ctx)
	f.lastDetails = details
	f.lastRunErr = err
	return true, err
}

func TestExperimentRepairTaskHandlerUsesDefaultLimitWhenPayloadIsInvalid(t *testing.T) {
	experimentID := uuid.New()
	runner := &fakeExperimentRepairRunner{result: service.ExperimentRepairRunResult{Scanned: 1, Repaired: []uuid.UUID{experimentID}, ObjectiveStatsSynced: 2}}
	executor := &fakeExperimentRepairExecutor{}
	handler := newExperimentRepairTaskHandler(runner, executor, zap.NewNop())

	err := handler(context.Background(), asynq.NewTask(TypeReconcileExperimentRepair, []byte("{")))

	require.NoError(t, err)
	assert.Equal(t, defaultExperimentRepairLimit, runner.lastLimit)
	assert.Equal(t, TypeReconcileExperimentRepair, executor.lastSpec.JobName)
	assert.Equal(t, 30*time.Minute, executor.lastSpec.Window)
	assert.Equal(t, defaultExperimentRepairLimit, executor.lastDetails["limit"])
	assert.Equal(t, 1, executor.lastDetails["repaired"])
	assert.Equal(t, 2, executor.lastDetails["objective_stats_synced"])
	assert.Equal(t, []string{experimentID.String()}, executor.lastDetails["repaired_experiment_ids"])
}

func TestExperimentRepairTaskHandlerPropagatesRepairFailureWithDetails(t *testing.T) {
	failedID := uuid.New()
	runner := &fakeExperimentRepairRunner{
		result: service.ExperimentRepairRunResult{
			Scanned:  1,
			Failures: map[string]string{failedID.String(): "boom"},
		},
		err: errors.New("failed to repair 1 experiment(s)"),
	}
	executor := &fakeExperimentRepairExecutor{}
	handler := newExperimentRepairTaskHandler(runner, executor, zap.NewNop())

	err := handler(context.Background(), asynq.NewTask(TypeReconcileExperimentRepair, mustMarshalJSON(ReconcileExperimentRepairPayload{Limit: 12})))

	require.Error(t, err)
	assert.Equal(t, 12, runner.lastLimit)
	assert.Equal(t, 1, executor.lastDetails["failed"])
	assert.Equal(t, map[string]string{failedID.String(): "boom"}, executor.lastDetails["failures"])
	assert.EqualError(t, executor.lastRunErr, "failed to repair 1 experiment(s)")
}
