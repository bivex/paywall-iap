package tasks

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// BanditMaintenanceJobs contains bandit maintenance background jobs
type BanditMaintenanceJobs struct {
	engine *service.AdvancedBanditEngine
	logger *zap.Logger
}

// NewBanditMaintenanceJobs creates a new bandit maintenance jobs handler
func NewBanditMaintenanceJobs(
	engine *service.AdvancedBanditEngine,
	logger *zap.Logger,
) *BanditMaintenanceJobs {
	return &BanditMaintenanceJobs{
		engine: engine,
		logger: logger,
	}
}

// ProcessExpiredPendingRewardsArgs represents arguments for processing expired rewards
type ProcessExpiredPendingRewardsArgs struct {
	BatchSize int `json:"batch_size"`
}

func (ProcessExpiredPendingRewardsArgs) Kind() string { return "process_expired_pending_rewards" }

// ProcessExpiredPendingRewardsResult represents the result of processing expired rewards
type ProcessExpiredPendingRewardsResult struct {
	Processed int `json:"processed"`
}

// ProcessExpiredPendingRewards processes expired pending rewards as non-conversions
func (j *BanditMaintenanceJobs) ProcessExpiredPendingRewards(
	ctx context.Context,
	args *river.Job[ProcessExpiredPendingRewardsArgs],
) (ProcessExpiredPendingRewardsResult, error) {
	batchSize := 100
	if args.Args.BatchSize > 0 {
		batchSize = args.Args.BatchSize
	}

	j.logger.Info("Processing expired pending rewards", zap.Int("batch_size", batchSize))

	processed, err := j.engine.ProcessExpiredPendingRewards(ctx, batchSize)
	if err != nil {
		j.logger.Error("Failed to process expired rewards", zap.Error(err))
		return ProcessExpiredPendingRewardsResult{}, err
	}

	j.logger.Info("Expired pending rewards processed")
	return ProcessExpiredPendingRewardsResult{
		Processed: processed,
	}, nil
}

// TrimSlidingWindowsArgs represents arguments for trimming sliding windows
type TrimSlidingWindowsArgs struct {
	ExperimentID string `json:"experiment_id,omitempty"`
}

func (TrimSlidingWindowsArgs) Kind() string { return "trim_sliding_windows" }

// TrimSlidingWindowsResult represents the result of trimming windows
type TrimSlidingWindowsResult struct {
	WindowsTrimmed int `json:"windows_trimmed"`
}

// TrimSlidingWindows trims sliding windows to configured size
func (j *BanditMaintenanceJobs) TrimSlidingWindows(
	ctx context.Context,
	args *river.Job[TrimSlidingWindowsArgs],
) (TrimSlidingWindowsResult, error) {
	j.logger.Info("Trimming sliding windows")

	if args.Args.ExperimentID != "" {
		experimentID, err := uuid.Parse(args.Args.ExperimentID)
		if err != nil {
			return TrimSlidingWindowsResult{}, err
		}
		if err := j.engine.TrimWindow(ctx, experimentID); err != nil {
			j.logger.Error("Failed to trim experiment window", zap.Error(err))
			return TrimSlidingWindowsResult{}, err
		}
		return TrimSlidingWindowsResult{WindowsTrimmed: 1}, nil
	}

	trimmed, err := j.engine.TrimConfiguredWindows(ctx, 100)
	if err != nil {
		j.logger.Error("Failed to trim windows", zap.Error(err))
		return TrimSlidingWindowsResult{}, err
	}

	j.logger.Info("Sliding windows trimmed")
	return TrimSlidingWindowsResult{
		WindowsTrimmed: trimmed,
	}, nil
}

// CleanupOldContextDataArgs represents arguments for cleaning up old context data
type CleanupOldContextDataArgs struct {
	DaysToKeep int `json:"days_to_keep"`
}

func (CleanupOldContextDataArgs) Kind() string { return "cleanup_old_context_data" }

// CleanupOldContextDataResult represents the result of cleanup
type CleanupOldContextDataResult struct {
	RecordsDeleted int `json:"records_deleted"`
}

// CleanupOldContextData cleans up old user context data
func (j *BanditMaintenanceJobs) CleanupOldContextData(
	ctx context.Context,
	args *river.Job[CleanupOldContextDataArgs],
) (CleanupOldContextDataResult, error) {
	daysToKeep := 90
	if args.Args.DaysToKeep > 0 {
		daysToKeep = args.Args.DaysToKeep
	}

	j.logger.Info("Cleaning up old context data", zap.Int("days_to_keep", daysToKeep))

	deleted, err := j.engine.CleanupOldContextData(ctx, time.Duration(daysToKeep)*24*time.Hour)
	if err != nil {
		j.logger.Error("Failed to cleanup old context data", zap.Error(err))
		return CleanupOldContextDataResult{}, err
	}

	j.logger.Info("Old context data cleaned up")
	return CleanupOldContextDataResult{
		RecordsDeleted: int(deleted),
	}, nil
}

// CalculateWinProbabilitiesArgs represents arguments for calculating win probabilities
type CalculateWinProbabilitiesArgs struct {
	ExperimentID string `json:"experiment_id"`
	Simulations  int    `json:"simulations"`
}

func (CalculateWinProbabilitiesArgs) Kind() string { return "calculate_win_probabilities" }

// CalculateWinProbabilitiesResult represents the result of calculating win probabilities
type CalculateWinProbabilitiesResult struct {
	WinProbabilities map[string]float64 `json:"win_probabilities"`
}

// CalculateWinProbabilities calculates win probabilities for experiment arms
func (j *BanditMaintenanceJobs) CalculateWinProbabilities(
	ctx context.Context,
	args *river.Job[CalculateWinProbabilitiesArgs],
) (CalculateWinProbabilitiesResult, error) {
	if args.Args.ExperimentID == "" {
		return CalculateWinProbabilitiesResult{}, nil
	}

	j.logger.Info("Calculating win probabilities",
		zap.String("experiment_id", args.Args.ExperimentID),
		zap.Int("simulations", args.Args.Simulations),
	)

	// This would require access to the base bandit
	// For now, return placeholder

	return CalculateWinProbabilitiesResult{
		WinProbabilities: make(map[string]float64),
	}, nil
}

// RunFullMaintenanceArgs represents arguments for full maintenance run
type RunFullMaintenanceArgs struct {
	// No arguments needed
}

func (RunFullMaintenanceArgs) Kind() string { return "run_full_maintenance" }

// RunFullMaintenanceResult represents the result of full maintenance
type RunFullMaintenanceResult struct {
	TasksCompleted []string `json:"tasks_completed"`
}

// RunFullMaintenance runs all maintenance tasks
func (j *BanditMaintenanceJobs) RunFullMaintenance(
	ctx context.Context,
	_ *river.Job[RunFullMaintenanceArgs],
) (RunFullMaintenanceResult, error) {
	j.logger.Info("Running full bandit maintenance")

	summary, err := j.engine.RunMaintenanceDetailed(ctx)
	if err != nil {
		j.logger.Error("Failed to run full maintenance", zap.Error(err))
		return RunFullMaintenanceResult{}, err
	}

	tasks := make([]string, 0, 6)
	if summary.ExpiredPendingRewards >= 0 {
		tasks = append(tasks, "processed_expired_pending_rewards")
	}
	if summary.WindowsTrimmed > 0 || summary.WindowExperimentsScanned > 0 {
		tasks = append(tasks, "trimmed_sliding_windows")
	}
	if summary.CurrencyRatesUpdated {
		tasks = append(tasks, "updated_currency_rates")
	}
	if summary.StaleContextsDeleted > 0 || summary.ExpiredAssignmentsDeleted > 0 {
		tasks = append(tasks, "cleaned_old_context_data", "cleaned_expired_assignments")
	}
	if summary.ObjectiveStatsSynced > 0 || summary.ObjectiveExperimentsScanned > 0 {
		tasks = append(tasks, "synced_objective_stats")
	}

	j.logger.Info("Full bandit maintenance completed")
	return RunFullMaintenanceResult{
		TasksCompleted: tasks,
	}, nil
}

// SyncObjectiveStatsArgs represents arguments for syncing objective stats
type SyncObjectiveStatsArgs struct {
	ExperimentID string `json:"experiment_id"`
}

func (SyncObjectiveStatsArgs) Kind() string { return "sync_objective_stats" }

// SyncObjectiveStatsResult represents the result of syncing stats
type SyncObjectiveStatsResult struct {
	StatsSynced int `json:"stats_synced"`
}

// SyncObjectiveStats syncs objective statistics with base statistics
func (j *BanditMaintenanceJobs) SyncObjectiveStats(
	ctx context.Context,
	args *river.Job[SyncObjectiveStatsArgs],
) (SyncObjectiveStatsResult, error) {
	j.logger.Info("Syncing objective statistics")

	synced, err := j.engine.SyncObjectiveStats(ctx, 100)
	if err != nil {
		j.logger.Error("Failed to sync objective statistics", zap.Error(err))
		return SyncObjectiveStatsResult{}, err
	}

	j.logger.Info("Objective statistics synced")
	return SyncObjectiveStatsResult{
		StatsSynced: synced,
	}, nil
}
