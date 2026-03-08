package tasks

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

const TypeReconcileExperimentRepair = "experiment:repair:reconcile"

const defaultExperimentRepairLimit = 50

type ReconcileExperimentRepairPayload struct {
	Limit int `json:"limit"`
}

type experimentRepairRunner interface {
	Reconcile(ctx context.Context, limit int) (service.ExperimentRepairRunResult, error)
}

type experimentRepairScheduledExecutor interface {
	ExecuteScheduled(ctx context.Context, spec service.ScheduledAutomationJobSpec, payload []byte, run func(context.Context) (map[string]any, error)) (bool, error)
}

func RegisterExperimentRepairTasks(mux *asynq.ServeMux, reconciler experimentRepairRunner, executor experimentRepairScheduledExecutor, logger *zap.Logger) {
	mux.HandleFunc(TypeReconcileExperimentRepair, newExperimentRepairTaskHandler(reconciler, executor, logger))
}

func RegisterExperimentRepairScheduledTasks(scheduler *asynq.Scheduler) error {
	_, err := scheduler.Register("*/30 * * * *", asynq.NewTask(TypeReconcileExperimentRepair, mustMarshalJSON(ReconcileExperimentRepairPayload{Limit: defaultExperimentRepairLimit})))
	return err
}

func newExperimentRepairTaskHandler(reconciler experimentRepairRunner, executor experimentRepairScheduledExecutor, logger *zap.Logger) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		payload := ReconcileExperimentRepairPayload{Limit: defaultExperimentRepairLimit}
		if len(task.Payload()) > 0 {
			if err := json.Unmarshal(task.Payload(), &payload); err != nil {
				logger.Warn("Failed to parse experiment repair payload", zap.Error(err))
				payload.Limit = defaultExperimentRepairLimit
			}
		}
		if payload.Limit <= 0 {
			payload.Limit = defaultExperimentRepairLimit
		}

		executed, err := executor.ExecuteScheduled(ctx, service.ScheduledAutomationJobSpec{
			JobName: TypeReconcileExperimentRepair,
			Source:  "asynq_scheduler",
			Window:  30 * time.Minute,
		}, task.Payload(), func(ctx context.Context) (map[string]any, error) {
			result, reconcileErr := reconciler.Reconcile(ctx, payload.Limit)
			details := map[string]any{
				"limit":                      payload.Limit,
				"scanned":                    result.Scanned,
				"repaired":                   len(result.Repaired),
				"failed":                     len(result.Failures),
				"missing_arm_stats_inserted": result.MissingArmStatsInserted,
				"expired_pending_rewards":    result.ExpiredPendingRewards,
				"pending_rewards_processed":  result.PendingRewardsProcessed,
			}
			if len(result.Repaired) > 0 {
				details["repaired_experiment_ids"] = experimentIDsToStrings(result.Repaired)
			}
			if len(result.Failures) > 0 {
				details["failures"] = result.Failures
			}

			if reconcileErr != nil {
				return details, reconcileErr
			}

			logger.Info("Experiment repair reconciliation completed",
				zap.Int("scanned", result.Scanned),
				zap.Int("repaired", len(result.Repaired)),
				zap.Int("failed", len(result.Failures)),
			)
			return details, nil
		})
		if err != nil {
			logger.Error("Failed to reconcile experiment repair", zap.Error(err))
			return err
		}
		if !executed {
			logger.Info("Skipping duplicate experiment repair reconciliation within scheduled window")
		}
		return nil
	}
}

func experimentIDsToStrings(ids []uuid.UUID) []string {
	if len(ids) == 0 {
		return nil
	}
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, id.String())
	}
	sort.Strings(values)
	return values
}
