package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ExperimentRepairAssignmentSnapshot struct {
	Total  int `json:"total"`
	Active int `json:"active"`
}

type ExperimentRepairSummary struct {
	AssignmentSnapshot      ExperimentRepairAssignmentSnapshot `json:"assignment_snapshot"`
	MissingArmStatsInserted int                                `json:"missing_arm_stats_inserted"`
	ObjectiveStatsSynced    int                                `json:"objective_stats_synced"`
	PendingRewardsTotal     int                                `json:"pending_rewards_total"`
	ExpiredPendingRewards   int                                `json:"expired_pending_rewards"`
	PendingRewardsProcessed int                                `json:"pending_rewards_processed"`
	WinnerConfidencePercent *float64                           `json:"winner_confidence_percent"`
}

type ExperimentRepairRepository interface {
	GetExperimentMutationState(ctx context.Context, experimentID uuid.UUID) (*ExperimentMutationState, error)
	CountExperimentAssignments(ctx context.Context, experimentID uuid.UUID) (total int, active int, err error)
	EnsureExperimentArmStats(ctx context.Context, experimentID uuid.UUID) (int, error)
	GetExperimentObjectiveConfig(ctx context.Context, experimentID uuid.UUID) (*ExperimentConfig, error)
	CountExperimentPendingRewards(ctx context.Context, experimentID uuid.UUID) (total int, expired int, err error)
	ProcessExpiredPendingRewards(ctx context.Context, experimentID uuid.UUID) (int, error)
	UpdateExperimentWinnerConfidence(ctx context.Context, experimentID uuid.UUID, confidence *float64) error
}

type experimentRepairBanditRepository interface {
	BanditRepository
	UpdateObjectiveStats(ctx context.Context, stats *ArmObjectiveStats) error
}

type noopBanditCache struct{}

func (noopBanditCache) GetArmStats(context.Context, string) (*ArmStats, error) {
	return nil, errors.New("cache miss")
}
func (noopBanditCache) SetArmStats(context.Context, string, *ArmStats, time.Duration) error {
	return nil
}
func (noopBanditCache) GetAssignment(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, errors.New("cache miss")
}
func (noopBanditCache) SetAssignment(context.Context, string, uuid.UUID, time.Duration) error {
	return nil
}

type ExperimentRepairService struct {
	repo        ExperimentRepairRepository
	banditRepo  experimentRepairBanditRepository
	simulations int
}

func NewExperimentRepairService(repo ExperimentRepairRepository, banditRepo experimentRepairBanditRepository) *ExperimentRepairService {
	return &ExperimentRepairService{repo: repo, banditRepo: banditRepo, simulations: 2000}
}

func (s *ExperimentRepairService) RepairExperiment(ctx context.Context, experimentID uuid.UUID) (*ExperimentRepairSummary, error) {
	if _, err := s.repo.GetExperimentMutationState(ctx, experimentID); err != nil {
		return nil, err
	}

	totalAssignments, activeAssignments, err := s.repo.CountExperimentAssignments(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	pendingTotal, expiredPending, err := s.repo.CountExperimentPendingRewards(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	inserted, err := s.repo.EnsureExperimentArmStats(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	processed, err := s.repo.ProcessExpiredPendingRewards(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	objectiveStatsSynced, err := s.syncObjectiveStats(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	winnerConfidence, err := s.recalculateWinnerConfidence(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateExperimentWinnerConfidence(ctx, experimentID, winnerConfidence); err != nil {
		return nil, err
	}

	var winnerConfidencePercent *float64
	if winnerConfidence != nil {
		value := *winnerConfidence * 100
		winnerConfidencePercent = &value
	}

	return &ExperimentRepairSummary{
		AssignmentSnapshot:      ExperimentRepairAssignmentSnapshot{Total: totalAssignments, Active: activeAssignments},
		MissingArmStatsInserted: inserted,
		ObjectiveStatsSynced:    objectiveStatsSynced,
		PendingRewardsTotal:     pendingTotal,
		ExpiredPendingRewards:   expiredPending,
		PendingRewardsProcessed: processed,
		WinnerConfidencePercent: winnerConfidencePercent,
	}, nil
}

func (s *ExperimentRepairService) syncObjectiveStats(ctx context.Context, experimentID uuid.UUID) (int, error) {
	config, err := s.repo.GetExperimentObjectiveConfig(ctx, experimentID)
	if err != nil {
		return 0, err
	}
	objectiveTypes := maintenanceObjectiveTypes(config)
	if len(objectiveTypes) == 0 {
		return 0, nil
	}

	arms, err := s.banditRepo.GetArms(ctx, experimentID)
	if err != nil {
		return 0, fmt.Errorf("failed to load arms for objective repair: %w", err)
	}

	synced := 0
	for _, arm := range arms {
		stats, err := s.banditRepo.GetArmStats(ctx, arm.ID)
		if err != nil {
			return synced, fmt.Errorf("failed to load arm stats for objective repair: %w", err)
		}

		for _, objectiveType := range objectiveTypes {
			if err := s.banditRepo.UpdateObjectiveStats(ctx, &ArmObjectiveStats{
				ArmID:         arm.ID,
				ObjectiveType: objectiveType,
				Alpha:         stats.Alpha,
				Beta:          stats.Beta,
				Samples:       stats.Samples,
				Conversions:   stats.Conversions,
				TotalRevenue:  stats.Revenue,
				AvgLTV:        stats.AvgReward,
			}); err != nil {
				return synced, fmt.Errorf("failed to sync objective stats during repair: %w", err)
			}
			synced++
		}
	}

	return synced, nil
}

func (s *ExperimentRepairService) recalculateWinnerConfidence(ctx context.Context, experimentID uuid.UUID) (*float64, error) {
	bandit := NewThompsonSamplingBandit(s.banditRepo, noopBanditCache{}, zap.NewNop())
	winProbabilities, err := bandit.CalculateWinProbability(ctx, experimentID, s.simulations)
	if err != nil {
		return nil, err
	}
	if len(winProbabilities) == 0 {
		return nil, nil
	}

	var best float64
	found := false
	for _, probability := range winProbabilities {
		if !found || probability > best {
			best = probability
			found = true
		}
	}
	if !found {
		return nil, nil
	}
	return &best, nil
}
