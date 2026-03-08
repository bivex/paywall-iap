package service

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type advancedEngineTestRepo struct {
	experimentConfig      *ExperimentConfig
	updatedConfig         *ExperimentConfig
	arms                  []Arm
	armStats              map[uuid.UUID]*ArmStats
	pendingReward         *PendingReward
	userRewards           []*PendingReward
	windowExperiments     []uuid.UUID
	objectiveExperiments  []uuid.UUID
	updatedObjectiveStats []*ArmObjectiveStats
	deletedContexts       int64
	deletedAssignments    int64
}

func (r *advancedEngineTestRepo) GetArms(ctx context.Context, experimentID uuid.UUID) ([]Arm, error) {
	return r.arms, nil
}

func (r *advancedEngineTestRepo) GetArmStats(ctx context.Context, armID uuid.UUID) (*ArmStats, error) {
	if stats, ok := r.armStats[armID]; ok {
		return stats, nil
	}
	return &ArmStats{ArmID: armID, Alpha: 1, Beta: 1}, nil
}

func (r *advancedEngineTestRepo) UpdateArmStats(ctx context.Context, stats *ArmStats) error {
	return nil
}
func (r *advancedEngineTestRepo) CreateAssignment(ctx context.Context, assignment *Assignment) error {
	return nil
}
func (r *advancedEngineTestRepo) GetActiveAssignment(ctx context.Context, experimentID, userID uuid.UUID) (*Assignment, error) {
	return nil, nil
}
func (r *advancedEngineTestRepo) GetExperimentConfig(ctx context.Context, experimentID uuid.UUID) (*ExperimentConfig, error) {
	return r.experimentConfig, nil
}
func (r *advancedEngineTestRepo) UpdateObjectiveConfig(ctx context.Context, experimentID uuid.UUID, objectiveType ObjectiveType, objectiveWeights map[string]float64) error {
	r.updatedConfig = &ExperimentConfig{ID: experimentID, ObjectiveType: objectiveType, ObjectiveWeights: objectiveWeights}
	return nil
}
func (r *advancedEngineTestRepo) GetUserContext(ctx context.Context, userID uuid.UUID) (*UserContext, error) {
	return &UserContext{UserID: userID}, nil
}
func (r *advancedEngineTestRepo) SetUserContext(ctx context.Context, uctx *UserContext) error {
	return nil
}
func (r *advancedEngineTestRepo) GetObjectiveStats(ctx context.Context, armID uuid.UUID, objectiveType ObjectiveType) (*ArmObjectiveStats, error) {
	stats := r.armStats[armID]
	return &ArmObjectiveStats{ArmID: armID, ObjectiveType: objectiveType, AvgLTV: stats.AvgReward}, nil
}
func (r *advancedEngineTestRepo) UpdateObjectiveStats(ctx context.Context, stats *ArmObjectiveStats) error {
	r.updatedObjectiveStats = append(r.updatedObjectiveStats, stats)
	return nil
}
func (r *advancedEngineTestRepo) GetAllObjectiveStats(ctx context.Context, armID uuid.UUID) (map[ObjectiveType]*ArmObjectiveStats, error) {
	stats := r.armStats[armID]
	return map[ObjectiveType]*ArmObjectiveStats{
		ObjectiveConversion: {ArmID: armID, ObjectiveType: ObjectiveConversion, Alpha: stats.Alpha, Beta: stats.Beta, Samples: stats.Samples, Conversions: stats.Conversions, TotalRevenue: stats.Revenue, AvgLTV: stats.AvgReward},
		ObjectiveLTV:        {ArmID: armID, ObjectiveType: ObjectiveLTV, Alpha: stats.Alpha, Beta: stats.Beta, Samples: stats.Samples, Conversions: stats.Conversions, TotalRevenue: stats.Revenue, AvgLTV: stats.AvgReward},
		ObjectiveRevenue:    {ArmID: armID, ObjectiveType: ObjectiveRevenue, Alpha: stats.Alpha, Beta: stats.Beta, Samples: stats.Samples, Conversions: stats.Conversions, TotalRevenue: stats.Revenue, AvgLTV: stats.AvgReward},
		ObjectiveHybrid:     {ArmID: armID, ObjectiveType: ObjectiveHybrid, Alpha: stats.Alpha, Beta: stats.Beta, Samples: stats.Samples, Conversions: stats.Conversions, TotalRevenue: stats.Revenue, AvgLTV: stats.AvgReward},
	}, nil
}
func (r *advancedEngineTestRepo) CreatePendingReward(ctx context.Context, reward *PendingReward) error {
	return nil
}
func (r *advancedEngineTestRepo) GetPendingReward(ctx context.Context, id uuid.UUID) (*PendingReward, error) {
	return r.pendingReward, nil
}
func (r *advancedEngineTestRepo) GetPendingRewardsByUser(ctx context.Context, userID, experimentID uuid.UUID) ([]*PendingReward, error) {
	return r.userRewards, nil
}
func (r *advancedEngineTestRepo) GetExpiredPendingRewards(ctx context.Context, limit int) ([]*PendingReward, error) {
	return nil, nil
}
func (r *advancedEngineTestRepo) UpdatePendingReward(ctx context.Context, reward *PendingReward) error {
	return nil
}
func (r *advancedEngineTestRepo) LinkConversion(ctx context.Context, link *ConversionLink) error {
	return nil
}
func (r *advancedEngineTestRepo) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*ConversionLink, error) {
	return nil, nil
}
func (r *advancedEngineTestRepo) ListWindowMaintenanceExperimentIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	return r.windowExperiments, nil
}
func (r *advancedEngineTestRepo) ListObjectiveSyncExperimentIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	return r.objectiveExperiments, nil
}
func (r *advancedEngineTestRepo) CleanupStaleUserContext(ctx context.Context, olderThan time.Duration) (int64, error) {
	return r.deletedContexts, nil
}
func (r *advancedEngineTestRepo) CleanupExpiredAssignments(ctx context.Context, olderThan time.Duration) (int64, error) {
	return r.deletedAssignments, nil
}

type advancedEngineTestCache struct{}

func (c *advancedEngineTestCache) GetArmStats(ctx context.Context, key string) (*ArmStats, error) {
	return nil, nil
}
func (c *advancedEngineTestCache) SetArmStats(ctx context.Context, key string, stats *ArmStats, ttl time.Duration) error {
	return nil
}
func (c *advancedEngineTestCache) GetAssignment(ctx context.Context, key string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (c *advancedEngineTestCache) SetAssignment(ctx context.Context, key string, armID uuid.UUID, ttl time.Duration) error {
	return nil
}

func TestAdvancedBanditEngine_GetObjectiveScores_LazilyLoadsExperimentConfig(t *testing.T) {
	experimentID := uuid.New()
	armID := uuid.New()
	repo := &advancedEngineTestRepo{
		experimentConfig: &ExperimentConfig{
			ID:            experimentID,
			ObjectiveType: ObjectiveHybrid,
			ObjectiveWeights: map[string]float64{
				"conversion": 0.5,
				"ltv":        0.3,
				"revenue":    0.2,
			},
		},
		arms: []Arm{{ID: armID, ExperimentID: experimentID, Name: "A"}},
		armStats: map[uuid.UUID]*ArmStats{
			armID: {ArmID: armID, Alpha: 10, Beta: 5, Samples: 20, Conversions: 9, Revenue: 120, AvgReward: 6},
		},
	}
	cache := &advancedEngineTestCache{}
	base := NewThompsonSamplingBandit(repo, cache, zap.NewNop())
	engine := NewAdvancedBanditEngine(base, repo, cache, nil, nil, zap.NewNop(), &EngineConfig{EnableHybrid: true})

	scores, err := engine.GetObjectiveScores(context.Background(), experimentID)

	require.NoError(t, err)
	require.Contains(t, scores, armID)
	require.Contains(t, scores[armID], ObjectiveConversion)
	require.Contains(t, scores[armID], ObjectiveLTV)
	require.Contains(t, scores[armID], ObjectiveRevenue)
	require.Contains(t, scores[armID], ObjectiveHybrid)
}

func TestAdvancedBanditEngine_GetObjectiveConfig_ReturnsDefaultWhenUnset(t *testing.T) {
	experimentID := uuid.New()
	repo := &advancedEngineTestRepo{}
	cache := &advancedEngineTestCache{}
	base := NewThompsonSamplingBandit(repo, cache, zap.NewNop())
	engine := NewAdvancedBanditEngine(base, repo, cache, nil, nil, zap.NewNop(), &EngineConfig{EnableHybrid: true})

	config, err := engine.GetObjectiveConfig(context.Background(), experimentID)

	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, experimentID, config.ID)
	require.Equal(t, ObjectiveConversion, config.ObjectiveType)
	require.Nil(t, config.ObjectiveWeights)
}

func TestAdvancedBanditEngine_GetPendingReward_UsesLazyDelayedStrategy(t *testing.T) {
	pendingID := uuid.New()
	userID := uuid.New()
	repo := &advancedEngineTestRepo{
		pendingReward: &PendingReward{ID: pendingID, UserID: userID},
		userRewards:   []*PendingReward{{ID: pendingID, UserID: userID}},
	}
	cache := &advancedEngineTestCache{}
	base := NewThompsonSamplingBandit(repo, cache, zap.NewNop())
	engine := NewAdvancedBanditEngine(base, repo, cache, nil, nil, zap.NewNop(), &EngineConfig{EnableDelayed: true})

	pendingReward, err := engine.GetPendingReward(context.Background(), pendingID)
	rewards, rewardsErr := engine.GetUserPendingRewards(context.Background(), userID)

	require.NoError(t, err)
	require.NoError(t, rewardsErr)
	require.Equal(t, pendingID, pendingReward.ID)
	require.Len(t, rewards, 1)
	require.Equal(t, pendingID, rewards[0].ID)
}

func TestAdvancedBanditEngine_SetObjectiveConfig_NormalizesHybridWeights(t *testing.T) {
	experimentID := uuid.New()
	repo := &advancedEngineTestRepo{}
	cache := &advancedEngineTestCache{}
	base := NewThompsonSamplingBandit(repo, cache, zap.NewNop())
	engine := NewAdvancedBanditEngine(base, repo, cache, nil, nil, zap.NewNop(), &EngineConfig{EnableHybrid: true})

	config, err := engine.SetObjectiveConfig(context.Background(), experimentID, ObjectiveHybrid, map[string]float64{
		"conversion": 5,
		"ltv":        3,
		"revenue":    2,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.updatedConfig)
	require.Equal(t, ObjectiveHybrid, repo.updatedConfig.ObjectiveType)
	require.InDelta(t, 0.5, config.ObjectiveWeights["conversion"], 0.0001)
	require.InDelta(t, 0.3, config.ObjectiveWeights["ltv"], 0.0001)
	require.InDelta(t, 0.2, config.ObjectiveWeights["revenue"], 0.0001)
}

func TestAdvancedBanditEngine_SyncObjectiveStats_UsesConfiguredHybridObjectives(t *testing.T) {
	experimentID := uuid.New()
	firstArmID := uuid.New()
	secondArmID := uuid.New()
	repo := &advancedEngineTestRepo{
		experimentConfig: &ExperimentConfig{
			ID:            experimentID,
			ObjectiveType: ObjectiveHybrid,
			ObjectiveWeights: map[string]float64{
				"conversion": 0.7,
				"ltv":        0.3,
			},
		},
		objectiveExperiments: []uuid.UUID{experimentID},
		arms:                 []Arm{{ID: firstArmID, ExperimentID: experimentID}, {ID: secondArmID, ExperimentID: experimentID}},
		armStats: map[uuid.UUID]*ArmStats{
			firstArmID:  {ArmID: firstArmID, Alpha: 11, Beta: 3, Samples: 12, Conversions: 10, Revenue: 100, AvgReward: 8.33},
			secondArmID: {ArmID: secondArmID, Alpha: 5, Beta: 7, Samples: 10, Conversions: 4, Revenue: 40, AvgReward: 4},
		},
	}
	cache := &advancedEngineTestCache{}
	base := NewThompsonSamplingBandit(repo, cache, zap.NewNop())
	engine := NewAdvancedBanditEngine(base, repo, cache, nil, nil, zap.NewNop(), &EngineConfig{EnableHybrid: true})

	synced, err := engine.SyncObjectiveStats(context.Background(), 10)

	require.NoError(t, err)
	assert.Equal(t, 4, synced)
	if assert.Len(t, repo.updatedObjectiveStats, 4) {
		seen := make([]string, 0, len(repo.updatedObjectiveStats))
		for _, stat := range repo.updatedObjectiveStats {
			seen = append(seen, stat.ArmID.String()+":"+string(stat.ObjectiveType))
		}
		sort.Strings(seen)
		assert.Equal(t, []string{
			firstArmID.String() + ":conversion",
			firstArmID.String() + ":ltv",
			secondArmID.String() + ":conversion",
			secondArmID.String() + ":ltv",
		}, seen)
	}
}

func TestAdvancedBanditEngine_RunMaintenanceDetailed_ReturnsRepositoryBackedSummary(t *testing.T) {
	experimentID := uuid.New()
	armID := uuid.New()
	repo := &advancedEngineTestRepo{
		experimentConfig:     &ExperimentConfig{ID: experimentID, ObjectiveType: ObjectiveRevenue},
		windowExperiments:    []uuid.UUID{experimentID},
		objectiveExperiments: []uuid.UUID{experimentID},
		deletedContexts:      3,
		deletedAssignments:   2,
		arms:                 []Arm{{ID: armID, ExperimentID: experimentID}},
		armStats: map[uuid.UUID]*ArmStats{
			armID: {ArmID: armID, Alpha: 4, Beta: 2, Samples: 5, Conversions: 3, Revenue: 42, AvgReward: 8.4},
		},
	}
	cache := &advancedEngineTestCache{}
	base := NewThompsonSamplingBandit(repo, cache, zap.NewNop())
	engine := NewAdvancedBanditEngine(base, repo, cache, nil, nil, zap.NewNop(), &EngineConfig{})

	summary, err := engine.RunMaintenanceDetailed(context.Background())

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, 1, summary.WindowExperimentsScanned)
	assert.Equal(t, 1, summary.ObjectiveExperimentsScanned)
	assert.Equal(t, 1, summary.ObjectiveStatsSynced)
	assert.EqualValues(t, 3, summary.StaleContextsDeleted)
	assert.EqualValues(t, 2, summary.ExpiredAssignmentsDeleted)
}
