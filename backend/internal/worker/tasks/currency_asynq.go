package tasks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

type banditMaintenanceEngine interface {
	RunMaintenanceDetailed(ctx context.Context) (*service.BanditMaintenanceSummary, error)
	TrimConfiguredWindows(ctx context.Context, limit int) (int, error)
	ProcessExpiredPendingRewards(ctx context.Context, batchSize int) (int, error)
}

type scheduledJobExecutor interface {
	ExecuteScheduled(ctx context.Context, spec service.ScheduledAutomationJobSpec, payload []byte, run func(context.Context) (map[string]any, error)) (bool, error)
}

// =====================================================
// Asynq-compatible Currency Tasks
// =====================================================

// RegisterCurrencyTasks registers currency-related task handlers
func RegisterCurrencyTasks(mux *asynq.ServeMux, currencyService *service.CurrencyRateService, executor *service.AutomationJobExecutionService, logger *zap.Logger) {
	mux.HandleFunc("currency:update", func(ctx context.Context, t *asynq.Task) error {
		executed, err := executor.ExecuteScheduled(ctx, service.ScheduledAutomationJobSpec{
			JobName: "currency:update",
			Source:  "asynq_scheduler",
			Window:  30 * time.Minute,
		}, t.Payload(), func(ctx context.Context) (map[string]any, error) {
			logger.Info("Processing currency rate update")
			if err := currencyService.UpdateRates(ctx); err != nil {
				return nil, err
			}
			logger.Info("Currency rates updated successfully")
			return map[string]any{"operation": "update_rates"}, nil
		})
		if err != nil {
			logger.Error("Failed to update currency rates", zap.Error(err))
			return err
		}
		if !executed {
			logger.Info("Skipping duplicate currency rate update within scheduled window")
		}
		return nil
	})
}

// =====================================================
// Scheduled Currency Task Registration
// =====================================================

// RegisterCurrencyScheduledTasks registers scheduled currency tasks
func RegisterCurrencyScheduledTasks(scheduler *asynq.Scheduler) error {
	// Update currency rates every hour
	_, err := scheduler.Register("*/30 * * * *", asynq.NewTask("currency:update", nil))
	if err != nil {
		return err
	}

	return nil
}

// =====================================================
// Asynq-compatible Bandit Maintenance Tasks
// =====================================================

// RegisterBanditMaintenanceTasks registers bandit maintenance task handlers
func RegisterBanditMaintenanceTasks(mux *asynq.ServeMux, advancedEngine banditMaintenanceEngine, executor scheduledJobExecutor, logger *zap.Logger) {
	mux.HandleFunc("bandit:maintenance:full", func(ctx context.Context, t *asynq.Task) error {
		executed, err := executor.ExecuteScheduled(ctx, service.ScheduledAutomationJobSpec{
			JobName: "bandit:maintenance:full",
			Source:  "asynq_scheduler",
			Window:  6 * time.Hour,
		}, t.Payload(), func(ctx context.Context) (map[string]any, error) {
			logger.Info("Processing full bandit maintenance")
			summary, err := advancedEngine.RunMaintenanceDetailed(ctx)
			if err != nil {
				return nil, err
			}
			logger.Info("Bandit maintenance completed")
			return banditMaintenanceSummaryDetails(summary), nil
		})
		if err != nil {
			logger.Error("Failed to run maintenance", zap.Error(err))
			return err
		}
		if !executed {
			logger.Info("Skipping duplicate full bandit maintenance within scheduled window")
		}
		return nil
	})

	mux.HandleFunc("bandit:maintenance:trim_windows", func(ctx context.Context, t *asynq.Task) error {
		// Parse batch size from payload
		var payload struct {
			BatchSize int `json:"batch_size"`
		}
		if len(t.Payload()) > 0 {
			if err := json.Unmarshal(t.Payload(), &payload); err != nil {
				logger.Warn("Failed to parse payload", zap.Error(err))
				payload.BatchSize = 100
			}
		} else {
			payload.BatchSize = 100
		}

		executed, err := executor.ExecuteScheduled(ctx, service.ScheduledAutomationJobSpec{
			JobName: "bandit:maintenance:trim_windows",
			Source:  "asynq_scheduler",
			Window:  time.Hour,
		}, t.Payload(), func(ctx context.Context) (map[string]any, error) {
			logger.Info("Processing window trimming", zap.Int("batch_size", payload.BatchSize))
			trimmed, err := advancedEngine.TrimConfiguredWindows(ctx, payload.BatchSize)
			if err != nil {
				return nil, err
			}
			logger.Info("Window trimming completed")
			return map[string]any{"maintenance": "trim_windows", "batch_size": payload.BatchSize, "windows_trimmed": trimmed}, nil
		})
		if err != nil {
			logger.Error("Failed to trim windows", zap.Error(err))
			return err
		}
		if !executed {
			logger.Info("Skipping duplicate window trimming within scheduled window")
		}
		return nil
	})

	mux.HandleFunc("bandit:maintenance:process_expired", func(ctx context.Context, t *asynq.Task) error {
		// Parse batch size from payload
		var payload struct {
			BatchSize int `json:"batch_size"`
		}
		if len(t.Payload()) > 0 {
			if err := json.Unmarshal(t.Payload(), &payload); err != nil {
				logger.Warn("Failed to parse payload", zap.Error(err))
				payload.BatchSize = 100
			}
		} else {
			payload.BatchSize = 100
		}

		executed, err := executor.ExecuteScheduled(ctx, service.ScheduledAutomationJobSpec{
			JobName: "bandit:maintenance:process_expired",
			Source:  "asynq_scheduler",
			Window:  15 * time.Minute,
		}, t.Payload(), func(ctx context.Context) (map[string]any, error) {
			logger.Info("Processing expired pending rewards", zap.Int("batch_size", payload.BatchSize))
			processed, err := advancedEngine.ProcessExpiredPendingRewards(ctx, payload.BatchSize)
			if err != nil {
				return nil, err
			}
			logger.Info("Expired rewards processed")
			return map[string]any{"maintenance": "process_expired", "batch_size": payload.BatchSize, "processed": processed}, nil
		})
		if err != nil {
			logger.Error("Failed to process expired rewards", zap.Error(err))
			return err
		}
		if !executed {
			logger.Info("Skipping duplicate expired reward processing within scheduled window")
		}
		return nil
	})
}

func banditMaintenanceSummaryDetails(summary *service.BanditMaintenanceSummary) map[string]any {
	if summary == nil {
		return map[string]any{"maintenance": "full"}
	}
	return map[string]any{
		"maintenance":                   "full",
		"expired_pending_rewards":       summary.ExpiredPendingRewards,
		"currency_rates_updated":        summary.CurrencyRatesUpdated,
		"window_experiments_scanned":    summary.WindowExperimentsScanned,
		"windows_trimmed":               summary.WindowsTrimmed,
		"objective_experiments_scanned": summary.ObjectiveExperimentsScanned,
		"objective_stats_synced":        summary.ObjectiveStatsSynced,
		"stale_contexts_deleted":        summary.StaleContextsDeleted,
		"expired_assignments_deleted":   summary.ExpiredAssignmentsDeleted,
	}
}

// RegisterBanditMaintenanceScheduledTasks registers scheduled bandit maintenance tasks
func RegisterBanditMaintenanceScheduledTasks(scheduler *asynq.Scheduler) error {
	// Full maintenance every 6 hours
	_, err := scheduler.Register("0 */6 * * *", asynq.NewTask("bandit:maintenance:full", nil))
	if err != nil {
		return err
	}

	// Trim windows every hour
	_, err = scheduler.Register("0 * * * *", asynq.NewTask("bandit:maintenance:trim_windows", nil))
	if err != nil {
		return err
	}

	// Process expired rewards every 15 minutes
	_, err = scheduler.Register("*/15 * * * *", asynq.NewTask("bandit:maintenance:process_expired", nil))
	if err != nil {
		return err
	}

	return nil
}
