package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubExperimentRepairRepository struct {
	objectiveConfig    *ExperimentConfig
	winnerConfidence   *float64
	mutationStateCalls int
	processExpiredHook func()
}

func (s *stubExperimentRepairRepository) GetExperimentMutationState(context.Context, uuid.UUID) (*ExperimentMutationState, error) {
	s.mutationStateCalls++
	return &ExperimentMutationState{Status: "running"}, nil
}

func (s *stubExperimentRepairRepository) CountExperimentAssignments(context.Context, uuid.UUID) (int, int, error) {
	return 4, 3, nil
}

func (s *stubExperimentRepairRepository) EnsureExperimentArmStats(context.Context, uuid.UUID) (int, error) {
	return 1, nil
}

func (s *stubExperimentRepairRepository) GetExperimentObjectiveConfig(context.Context, uuid.UUID) (*ExperimentConfig, error) {
	return s.objectiveConfig, nil
}

func (s *stubExperimentRepairRepository) CountExperimentPendingRewards(context.Context, uuid.UUID) (int, int, error) {
	return 2, 1, nil
}

func (s *stubExperimentRepairRepository) ProcessExpiredPendingRewards(context.Context, uuid.UUID) (int, error) {
	if s.processExpiredHook != nil {
		s.processExpiredHook()
	}
	return 1, nil
}

func (s *stubExperimentRepairRepository) UpdateExperimentWinnerConfidence(_ context.Context, _ uuid.UUID, confidence *float64) error {
	s.winnerConfidence = confidence
	return nil
}

type stubExperimentRepairBanditRepository struct {
	arms             []Arm
	armStats         map[uuid.UUID]*ArmStats
	updatedObjective []*ArmObjectiveStats
}

func (s *stubExperimentRepairBanditRepository) GetArms(context.Context, uuid.UUID) ([]Arm, error) {
	return s.arms, nil
}

func (s *stubExperimentRepairBanditRepository) GetArmStats(_ context.Context, armID uuid.UUID) (*ArmStats, error) {
	return s.armStats[armID], nil
}

func (s *stubExperimentRepairBanditRepository) UpdateArmStats(context.Context, *ArmStats) error {
	return nil
}
func (s *stubExperimentRepairBanditRepository) CreateAssignment(context.Context, *Assignment) error {
	return nil
}
func (s *stubExperimentRepairBanditRepository) GetActiveAssignment(context.Context, uuid.UUID, uuid.UUID) (*Assignment, error) {
	return nil, ErrAssignmentNotFound
}
func (s *stubExperimentRepairBanditRepository) GetExperimentConfig(context.Context, uuid.UUID) (*ExperimentConfig, error) {
	return nil, nil
}
func (s *stubExperimentRepairBanditRepository) UpdateObjectiveConfig(context.Context, uuid.UUID, ObjectiveType, map[string]float64) error {
	return nil
}
func (s *stubExperimentRepairBanditRepository) GetUserContext(context.Context, uuid.UUID) (*UserContext, error) {
	return nil, nil
}
func (s *stubExperimentRepairBanditRepository) SetUserContext(context.Context, *UserContext) error {
	return nil
}
func (s *stubExperimentRepairBanditRepository) UpdateObjectiveStats(_ context.Context, stats *ArmObjectiveStats) error {
	copy := *stats
	s.updatedObjective = append(s.updatedObjective, &copy)
	return nil
}

func TestExperimentRepairServiceRepairsObjectiveStatsAndWinnerConfidence(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()
	controlArmID := uuid.New()
	variantArmID := uuid.New()
	repo := &stubExperimentRepairRepository{objectiveConfig: &ExperimentConfig{
		ID:            experimentID,
		ObjectiveType: ObjectiveHybrid,
		ObjectiveWeights: map[string]float64{
			"conversion": 0.6,
			"ltv":        0.4,
		},
	}}
	banditRepo := &stubExperimentRepairBanditRepository{
		arms: []Arm{{ID: controlArmID}, {ID: variantArmID}},
		armStats: map[uuid.UUID]*ArmStats{
			controlArmID: {ArmID: controlArmID, Alpha: 2, Beta: 40, Samples: 12, Conversions: 1, Revenue: 10, AvgReward: 0.8},
			variantArmID: {ArmID: variantArmID, Alpha: 80, Beta: 2, Samples: 80, Conversions: 78, Revenue: 640, AvgReward: 8},
		},
	}

	service := NewExperimentRepairService(repo, banditRepo)

	summary, err := service.RepairExperiment(ctx, experimentID)

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, 1, summary.MissingArmStatsInserted)
	assert.Equal(t, 4, summary.ObjectiveStatsSynced)
	assert.Equal(t, 2, summary.PendingRewardsTotal)
	assert.Equal(t, 1, summary.PendingRewardsProcessed)
	require.NotNil(t, summary.WinnerConfidencePercent)
	assert.Greater(t, *summary.WinnerConfidencePercent, 50.0)
	require.NotNil(t, repo.winnerConfidence)
	assert.Greater(t, *repo.winnerConfidence, 0.5)
	if assert.Len(t, banditRepo.updatedObjective, 4) {
		assert.Equal(t, controlArmID, banditRepo.updatedObjective[0].ArmID)
		assert.Equal(t, ObjectiveConversion, banditRepo.updatedObjective[0].ObjectiveType)
	}
}

func TestExperimentRepairServiceSyncsObjectiveStatsAfterExpiredPendingRewards(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()
	controlArmID := uuid.New()
	repo := &stubExperimentRepairRepository{objectiveConfig: &ExperimentConfig{ID: experimentID, ObjectiveType: ObjectiveConversion}}
	banditRepo := &stubExperimentRepairBanditRepository{
		arms: []Arm{{ID: controlArmID}},
		armStats: map[uuid.UUID]*ArmStats{
			controlArmID: {ArmID: controlArmID, Alpha: 1, Beta: 1, Samples: 0, Conversions: 0, Revenue: 0, AvgReward: 0},
		},
	}
	repo.processExpiredHook = func() {
		banditRepo.armStats[controlArmID] = &ArmStats{
			ArmID:       controlArmID,
			Alpha:       1,
			Beta:        2,
			Samples:     1,
			Conversions: 0,
			Revenue:     0,
			AvgReward:   0,
		}
	}

	service := NewExperimentRepairService(repo, banditRepo)

	summary, err := service.RepairExperiment(ctx, experimentID)

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, 1, summary.PendingRewardsProcessed)
	if assert.Len(t, banditRepo.updatedObjective, 1) {
		assert.Equal(t, 1, banditRepo.updatedObjective[0].Samples)
		assert.Equal(t, 2.0, banditRepo.updatedObjective[0].Beta)
	}
}
