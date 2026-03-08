package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ExperimentRepairCandidateRepository interface {
	ListExperimentRepairCandidateIDs(ctx context.Context, limit int) ([]uuid.UUID, error)
}

type ExperimentRepairExecutor interface {
	RepairExperiment(ctx context.Context, experimentID uuid.UUID) (*ExperimentRepairSummary, error)
}

type ExperimentRepairRunResult struct {
	Scanned                 int               `json:"scanned"`
	Repaired                []uuid.UUID       `json:"repaired"`
	Failures                map[string]string `json:"failures,omitempty"`
	MissingArmStatsInserted int               `json:"missing_arm_stats_inserted"`
	ObjectiveStatsSynced    int               `json:"objective_stats_synced"`
	ExpiredPendingRewards   int               `json:"expired_pending_rewards"`
	PendingRewardsProcessed int               `json:"pending_rewards_processed"`
}

type ExperimentRepairReconciler struct {
	candidates   ExperimentRepairCandidateRepository
	repairer     ExperimentRepairExecutor
	defaultLimit int
}

func NewExperimentRepairReconciler(candidates ExperimentRepairCandidateRepository, repairer ExperimentRepairExecutor) *ExperimentRepairReconciler {
	return &ExperimentRepairReconciler{candidates: candidates, repairer: repairer, defaultLimit: 50}
}

func (r *ExperimentRepairReconciler) Reconcile(ctx context.Context, limit int) (ExperimentRepairRunResult, error) {
	if limit <= 0 {
		limit = r.defaultLimit
	}

	candidateIDs, err := r.candidates.ListExperimentRepairCandidateIDs(ctx, limit)
	if err != nil {
		return ExperimentRepairRunResult{}, fmt.Errorf("failed to list experiment repair candidates: %w", err)
	}

	result := ExperimentRepairRunResult{Scanned: len(candidateIDs)}
	for _, experimentID := range candidateIDs {
		summary, repairErr := r.repairer.RepairExperiment(ctx, experimentID)
		if repairErr != nil {
			if result.Failures == nil {
				result.Failures = make(map[string]string)
			}
			result.Failures[experimentID.String()] = repairErr.Error()
			continue
		}

		result.Repaired = append(result.Repaired, experimentID)
		if summary == nil {
			continue
		}
		result.MissingArmStatsInserted += summary.MissingArmStatsInserted
		result.ObjectiveStatsSynced += summary.ObjectiveStatsSynced
		result.ExpiredPendingRewards += summary.ExpiredPendingRewards
		result.PendingRewardsProcessed += summary.PendingRewardsProcessed
	}

	if len(result.Failures) > 0 {
		return result, fmt.Errorf("failed to repair %d experiment(s)", len(result.Failures))
	}

	return result, nil
}
