package tasks

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// =====================================================
// Asynq-compatible Currency Tasks
// =====================================================

// RegisterCurrencyTasks registers currency-related task handlers
func RegisterCurrencyTasks(mux *asynq.ServeMux, currencyService *service.CurrencyRateService, logger *zap.Logger) {
	mux.HandleFunc("currency:update", func(ctx context.Context, t *asynq.Task) error {
		logger.Info("Processing currency rate update")

		if err := currencyService.UpdateRates(ctx); err != nil {
			logger.Error("Failed to update currency rates", zap.Error(err))
			return err
		}

		logger.Info("Currency rates updated successfully")
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
func RegisterBanditMaintenanceTasks(mux *asynq.ServeMux, advancedEngine *service.AdvancedBanditEngine, logger *zap.Logger) {
	mux.HandleFunc("bandit:maintenance:full", func(ctx context.Context, t *asynq.Task) error {
		logger.Info("Processing full bandit maintenance")

		if err := advancedEngine.RunMaintenance(ctx); err != nil {
			logger.Error("Failed to run maintenance", zap.Error(err))
			return err
		}

		logger.Info("Bandit maintenance completed")
		return nil
	})

	mux.HandleFunc("bandit:maintenance:trim_windows", func(ctx context.Context, t *asynq.Task) error {
		logger.Info("Processing window trimming")

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

		if err := advancedEngine.RunMaintenance(ctx); err != nil {
			logger.Error("Failed to trim windows", zap.Error(err))
			return err
		}

		logger.Info("Window trimming completed")
		return nil
	})

	mux.HandleFunc("bandit:maintenance:process_expired", func(ctx context.Context, t *asynq.Task) error {
		logger.Info("Processing expired pending rewards")

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

		if err := advancedEngine.RunMaintenance(ctx); err != nil {
			logger.Error("Failed to process expired rewards", zap.Error(err))
			return err
		}

		logger.Info("Expired rewards processed")
		return nil
	})
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
