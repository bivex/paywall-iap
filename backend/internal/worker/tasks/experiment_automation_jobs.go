package tasks

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

const TypeReconcileExperimentAutomation = "experiment:automation:reconcile"

func RegisterExperimentAutomationTasks(mux *asynq.ServeMux, reconciler *service.ExperimentAutomationReconciler, executor *service.AutomationJobExecutionService, logger *zap.Logger) {
	mux.HandleFunc(TypeReconcileExperimentAutomation, func(ctx context.Context, task *asynq.Task) error {
		executed, err := executor.ExecuteScheduled(ctx, service.ScheduledAutomationJobSpec{
			JobName: TypeReconcileExperimentAutomation,
			Source:  "asynq_scheduler",
			Window:  5 * time.Minute,
		}, task.Payload(), func(ctx context.Context) (map[string]any, error) {
			result, err := reconciler.Reconcile(ctx)
			if err != nil {
				return nil, err
			}

			logger.Info("Experiment automation reconciliation completed",
				zap.Int("started", len(result.Started)),
				zap.Int("completed", len(result.Completed)),
				zap.Int("skipped", result.Skipped),
			)

			return map[string]any{
				"started":   len(result.Started),
				"completed": len(result.Completed),
				"skipped":   result.Skipped,
			}, nil
		})
		if err != nil {
			logger.Error("Failed to reconcile experiment automation", zap.Error(err))
			return err
		}
		if !executed {
			logger.Info("Skipping duplicate experiment automation reconciliation within scheduled window")
		}
		return nil
	})
}

func RegisterExperimentAutomationScheduledTasks(scheduler *asynq.Scheduler) error {
	_, err := scheduler.Register("*/5 * * * *", asynq.NewTask(TypeReconcileExperimentAutomation, nil))
	return err
}
