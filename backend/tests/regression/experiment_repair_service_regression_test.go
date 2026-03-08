package regression

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	service "github.com/bivex/paywall-iap/internal/domain/service"
)

type repairRepoStub struct{ processExpiredHook func() }

func (s *repairRepoStub) GetExperimentMutationState(context.Context, uuid.UUID) (*service.ExperimentMutationState, error) {
	return &service.ExperimentMutationState{Status: "running"}, nil
}
func (s *repairRepoStub) CountExperimentAssignments(context.Context, uuid.UUID) (int, int, error) {
	return 1, 1, nil
}
func (s *repairRepoStub) EnsureExperimentArmStats(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (s *repairRepoStub) GetExperimentObjectiveConfig(context.Context, uuid.UUID) (*service.ExperimentConfig, error) {
	return &service.ExperimentConfig{ObjectiveType: service.ObjectiveConversion}, nil
}
func (s *repairRepoStub) CountExperimentPendingRewards(context.Context, uuid.UUID) (int, int, error) {
	return 1, 1, nil
}
func (s *repairRepoStub) ProcessExpiredPendingRewards(context.Context, uuid.UUID) (int, error) {
	if s.processExpiredHook != nil {
		s.processExpiredHook()
	}
	return 1, nil
}
func (s *repairRepoStub) UpdateExperimentWinnerConfidence(context.Context, uuid.UUID, *float64) error {
	return nil
}

type banditRepoStub struct {
	armID            uuid.UUID
	stats            *service.ArmStats
	updatedObjective []*service.ArmObjectiveStats
}

func (s *banditRepoStub) GetArms(context.Context, uuid.UUID) ([]service.Arm, error) {
	return []service.Arm{{ID: s.armID}}, nil
}
func (s *banditRepoStub) GetArmStats(context.Context, uuid.UUID) (*service.ArmStats, error) {
	return s.stats, nil
}
func (s *banditRepoStub) UpdateArmStats(context.Context, *service.ArmStats) error     { return nil }
func (s *banditRepoStub) CreateAssignment(context.Context, *service.Assignment) error { return nil }
func (s *banditRepoStub) GetActiveAssignment(context.Context, uuid.UUID, uuid.UUID) (*service.Assignment, error) {
	return nil, service.ErrAssignmentNotFound
}
func (s *banditRepoStub) GetExperimentConfig(context.Context, uuid.UUID) (*service.ExperimentConfig, error) {
	return nil, nil
}
func (s *banditRepoStub) UpdateObjectiveConfig(context.Context, uuid.UUID, service.ObjectiveType, map[string]float64) error {
	return nil
}
func (s *banditRepoStub) GetUserContext(context.Context, uuid.UUID) (*service.UserContext, error) {
	return nil, nil
}
func (s *banditRepoStub) SetUserContext(context.Context, *service.UserContext) error { return nil }
func (s *banditRepoStub) UpdateObjectiveStats(_ context.Context, stats *service.ArmObjectiveStats) error {
	copy := *stats
	s.updatedObjective = append(s.updatedObjective, &copy)
	return nil
}

func TestExperimentRepairServiceSyncsObjectiveStatsAfterExpiredPendingRewards(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()
	armID := uuid.New()
	banditRepo := &banditRepoStub{armID: armID, stats: &service.ArmStats{ArmID: armID, Alpha: 1, Beta: 1, Samples: 0}}
	repo := &repairRepoStub{processExpiredHook: func() {
		banditRepo.stats = &service.ArmStats{ArmID: armID, Alpha: 1, Beta: 2, Samples: 1}
	}}

	repairer := service.NewExperimentRepairService(repo, banditRepo)
	summary, err := repairer.RepairExperiment(ctx, experimentID)

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, 1, summary.PendingRewardsProcessed)
	if assert.Len(t, banditRepo.updatedObjective, 1) {
		assert.Equal(t, 1, banditRepo.updatedObjective[0].Samples)
		assert.Equal(t, 2.0, banditRepo.updatedObjective[0].Beta)
	}
}
