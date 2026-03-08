package tasks

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

const TypeReconcileExperimentAutomation = "experiment:automation:reconcile"

func RegisterExperimentAutomationTasks(mux *asynq.ServeMux, reconciler *service.ExperimentAutomationReconciler, logger *zap.Logger) {
	mux.HandleFunc(TypeReconcileExperimentAutomation, func(ctx context.Context, _ *asynq.Task) error {
		result, err := reconciler.Reconcile(ctx)
		if err != nil {
			logger.Error("Failed to reconcile experiment automation", zap.Error(err))
			return err
		}

		logger.Info("Experiment automation reconciliation completed",
			zap.Int("started", len(result.Started)),
			zap.Int("completed", len(result.Completed)),
			zap.Int("skipped", result.Skipped),
		)
		return nil
	})
}

func RegisterExperimentAutomationScheduledTasks(scheduler *asynq.Scheduler) error {
	_, err := scheduler.Register("*/5 * * * *", asynq.NewTask(TypeReconcileExperimentAutomation, nil))
	return err
}
