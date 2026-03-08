package tasks

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

type fakeBanditMaintenanceEngine struct {
	fullSummary   *service.BanditMaintenanceSummary
	fullErr       error
	trimmed       int
	trimErr       error
	processed     int
	processErr    error
	lastTrimLimit int
	lastBatchSize int
}

func (f *fakeBanditMaintenanceEngine) RunMaintenanceDetailed(context.Context) (*service.BanditMaintenanceSummary, error) {
	return f.fullSummary, f.fullErr
}

func (f *fakeBanditMaintenanceEngine) TrimConfiguredWindows(_ context.Context, limit int) (int, error) {
	f.lastTrimLimit = limit
	return f.trimmed, f.trimErr
}

func (f *fakeBanditMaintenanceEngine) ProcessExpiredPendingRewards(_ context.Context, batchSize int) (int, error) {
	f.lastBatchSize = batchSize
	return f.processed, f.processErr
}

type fakeBanditMaintenanceExecutor struct {
	lastSpec    service.ScheduledAutomationJobSpec
	lastDetails map[string]any
}

func (f *fakeBanditMaintenanceExecutor) ExecuteScheduled(ctx context.Context, spec service.ScheduledAutomationJobSpec, _ []byte, run func(context.Context) (map[string]any, error)) (bool, error) {
	f.lastSpec = spec
	details, err := run(ctx)
	f.lastDetails = details
	return true, err
}

func TestRegisterBanditMaintenanceTasks_FullMaintenanceWritesStructuredSummary(t *testing.T) {
	engine := &fakeBanditMaintenanceEngine{fullSummary: &service.BanditMaintenanceSummary{ExpiredPendingRewards: 2, WindowsTrimmed: 4, ObjectiveStatsSynced: 6}}
	executor := &fakeBanditMaintenanceExecutor{}
	mux := asynq.NewServeMux()
	RegisterBanditMaintenanceTasks(mux, engine, executor, zap.NewNop())

	err := mux.ProcessTask(context.Background(), asynq.NewTask("bandit:maintenance:full", nil))

	require.NoError(t, err)
	assert.Equal(t, "bandit:maintenance:full", executor.lastSpec.JobName)
	assert.Equal(t, 6*time.Hour, executor.lastSpec.Window)
	assert.Equal(t, 2, executor.lastDetails["expired_pending_rewards"])
	assert.Equal(t, 4, executor.lastDetails["windows_trimmed"])
	assert.Equal(t, 6, executor.lastDetails["objective_stats_synced"])
}

func TestRegisterBanditMaintenanceTasks_TargetedJobsUseDedicatedEngineOperations(t *testing.T) {
	engine := &fakeBanditMaintenanceEngine{trimmed: 7, processed: 9}
	executor := &fakeBanditMaintenanceExecutor{}
	mux := asynq.NewServeMux()
	RegisterBanditMaintenanceTasks(mux, engine, executor, zap.NewNop())

	err := mux.ProcessTask(context.Background(), asynq.NewTask("bandit:maintenance:trim_windows", mustMarshalJSON(struct {
		BatchSize int `json:"batch_size"`
	}{BatchSize: 12})))
	require.NoError(t, err)
	assert.Equal(t, 12, engine.lastTrimLimit)
	assert.Equal(t, 7, executor.lastDetails["windows_trimmed"])

	err = mux.ProcessTask(context.Background(), asynq.NewTask("bandit:maintenance:process_expired", mustMarshalJSON(struct {
		BatchSize int `json:"batch_size"`
	}{BatchSize: 15})))
	require.NoError(t, err)
	assert.Equal(t, 15, engine.lastBatchSize)
	assert.Equal(t, 9, executor.lastDetails["processed"])
}
