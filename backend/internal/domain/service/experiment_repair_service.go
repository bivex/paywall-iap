package service

import (
	"context"
	"errors"
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
	PendingRewardsTotal     int                                `json:"pending_rewards_total"`
	ExpiredPendingRewards   int                                `json:"expired_pending_rewards"`
	PendingRewardsProcessed int                                `json:"pending_rewards_processed"`
	WinnerConfidencePercent *float64                           `json:"winner_confidence_percent"`
}

type ExperimentRepairRepository interface {
	GetExperimentMutationState(ctx context.Context, experimentID uuid.UUID) (*ExperimentMutationState, error)
	CountExperimentAssignments(ctx context.Context, experimentID uuid.UUID) (total int, active int, err error)
	EnsureExperimentArmStats(ctx context.Context, experimentID uuid.UUID) (int, error)
	CountExperimentPendingRewards(ctx context.Context, experimentID uuid.UUID) (total int, expired int, err error)
	ProcessExpiredPendingRewards(ctx context.Context, experimentID uuid.UUID) (int, error)
	UpdateExperimentWinnerConfidence(ctx context.Context, experimentID uuid.UUID, confidence *float64) error
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
	banditRepo  BanditRepository
	simulations int
}

func NewExperimentRepairService(repo ExperimentRepairRepository, banditRepo BanditRepository) *ExperimentRepairService {
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
		PendingRewardsTotal:     pendingTotal,
		ExpiredPendingRewards:   expiredPending,
		PendingRewardsProcessed: processed,
		WinnerConfidencePercent: winnerConfidencePercent,
	}, nil
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
