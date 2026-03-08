package unit

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/cache"
)

// MockBanditRepository is a mock for BanditRepository
type MockBanditRepository struct {
	mock.Mock
}

func (m *MockBanditRepository) GetArms(ctx context.Context, experimentID uuid.UUID) ([]service.Arm, error) {
	args := m.Called(ctx, experimentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.Arm), args.Error(1)
}

func (m *MockBanditRepository) GetArmStats(ctx context.Context, armID uuid.UUID) (*service.ArmStats, error) {
	args := m.Called(ctx, armID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ArmStats), args.Error(1)
}

func (m *MockBanditRepository) UpdateArmStats(ctx context.Context, stats *service.ArmStats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func (m *MockBanditRepository) CreateAssignment(ctx context.Context, assignment *service.Assignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *MockBanditRepository) GetActiveAssignment(ctx context.Context, experimentID, userID uuid.UUID) (*service.Assignment, error) {
	args := m.Called(ctx, experimentID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Assignment), args.Error(1)
}

func (m *MockBanditRepository) GetExperimentConfig(ctx context.Context, experimentID uuid.UUID) (*service.ExperimentConfig, error) {
	args := m.Called(ctx, experimentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ExperimentConfig), args.Error(1)
}

func (m *MockBanditRepository) UpdateObjectiveConfig(ctx context.Context, experimentID uuid.UUID, objectiveType service.ObjectiveType, objectiveWeights map[string]float64) error {
	args := m.Called(ctx, experimentID, objectiveType, objectiveWeights)
	return args.Error(0)
}

func (m *MockBanditRepository) GetUserContext(ctx context.Context, userID uuid.UUID) (*service.UserContext, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UserContext), args.Error(1)
}

func (m *MockBanditRepository) SetUserContext(ctx context.Context, uctx *service.UserContext) error {
	args := m.Called(ctx, uctx)
	return args.Error(0)
}

func (m *MockBanditRepository) AppendConversionEvent(ctx context.Context, event *service.ConversionEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockBanditRepository) AppendImpressionEvent(ctx context.Context, event *service.ImpressionEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// MockBanditCache is a mock for BanditCache
type MockBanditCache struct {
	mock.Mock
	data        map[string]*service.ArmStats
	assignments map[string]uuid.UUID
}

func NewMockBanditCache() *MockBanditCache {
	return &MockBanditCache{
		data:        make(map[string]*service.ArmStats),
		assignments: make(map[string]uuid.UUID),
	}
}

func (m *MockBanditCache) GetArmStats(ctx context.Context, key string) (*service.ArmStats, error) {
	if data, ok := m.data[key]; ok {
		return data, nil
	}
	return nil, cache.ErrNotFound
}

func (m *MockBanditCache) SetArmStats(ctx context.Context, key string, stats *service.ArmStats, ttl time.Duration) error {
	m.data[key] = stats
	if len(m.ExpectedCalls) > 0 {
		m.Called(ctx, key, stats, ttl)
	}
	return nil
}

func (m *MockBanditCache) GetAssignment(ctx context.Context, key string) (uuid.UUID, error) {
	if armID, ok := m.assignments[key]; ok {
		return armID, nil
	}
	return uuid.Nil, cache.ErrNotFound
}

func (m *MockBanditCache) SetAssignment(ctx context.Context, key string, armID uuid.UUID, ttl time.Duration) error {
	m.assignments[key] = armID
	if len(m.ExpectedCalls) > 0 {
		m.Called(ctx, key, armID, ttl)
	}
	return nil
}

// TestSelectArm tests the SelectArm method
func TestSelectArm(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()
	userID := uuid.New()
	logger := zap.NewNop()

	repo := new(MockBanditRepository)
	cache := NewMockBanditCache()

	bandit := service.NewThompsonSamplingBandit(repo, cache, logger)

	t.Run("returns existing sticky assignment", func(t *testing.T) {
		existingArmID := uuid.New()
		existingAssignment := &service.Assignment{
			ID:           uuid.New(),
			ExperimentID: experimentID,
			UserID:       userID,
			ArmID:        existingArmID,
			AssignedAt:   time.Now().Add(-1 * time.Hour),
			ExpiresAt:    time.Now().Add(23 * time.Hour),
		}

		repo.On("GetActiveAssignment", ctx, experimentID, userID).Return(existingAssignment, nil)

		armID, err := bandit.SelectArm(ctx, experimentID, userID)

		assert.NoError(t, err)
		assert.Equal(t, existingArmID, armID)
		repo.AssertExpectations(t)
	})

	t.Run("selects arm using Thompson Sampling", func(t *testing.T) {
		arm1ID := uuid.New()
		arm2ID := uuid.New()
		arm3ID := uuid.New()

		arms := []service.Arm{
			{ID: arm1ID, Name: "Arm 1"},
			{ID: arm2ID, Name: "Arm 2"},
			{ID: arm3ID, Name: "Arm 3"},
		}

		repo = new(MockBanditRepository)
		cache = NewMockBanditCache()
		bandit = service.NewThompsonSamplingBandit(repo, cache, logger)

		repo.On("GetActiveAssignment", ctx, experimentID, userID).Return(nil, service.ErrAssignmentNotFound)
		repo.On("GetArms", ctx, experimentID).Return(arms, nil)

		// Mock stats - Arm 1 has best performance
		stats1 := &service.ArmStats{ArmID: arm1ID, Alpha: 10, Beta: 2} // High conversion
		stats2 := &service.ArmStats{ArmID: arm2ID, Alpha: 5, Beta: 5}  // Medium
		stats3 := &service.ArmStats{ArmID: arm3ID, Alpha: 2, Beta: 10} // Low conversion

		repo.On("GetArmStats", ctx, arm1ID).Return(stats1, nil)
		repo.On("GetArmStats", ctx, arm2ID).Return(stats2, nil)
		repo.On("GetArmStats", ctx, arm3ID).Return(stats3, nil)
		repo.On("CreateAssignment", ctx, mock.MatchedBy(func(assignment *service.Assignment) bool {
			return assignment.ExperimentID == experimentID &&
				assignment.UserID == userID &&
				containsUUID([]uuid.UUID{arm1ID, arm2ID, arm3ID}, assignment.ArmID) &&
				assignment.ExpiresAt.After(assignment.AssignedAt)
		})).Return(nil)

		armID, err := bandit.SelectArm(ctx, experimentID, userID)

		assert.NoError(t, err)
		assert.Contains(t, []uuid.UUID{arm1ID, arm2ID, arm3ID}, armID)
		repo.AssertExpectations(t)
	})

	t.Run("caches assignment after selection", func(t *testing.T) {
		arm1ID := uuid.New()
		arms := []service.Arm{{ID: arm1ID, Name: "Arm 1"}}

		repo = new(MockBanditRepository)
		cache = NewMockBanditCache()
		bandit = service.NewThompsonSamplingBandit(repo, cache, logger)

		repo.On("GetActiveAssignment", ctx, experimentID, userID).Return(nil, service.ErrAssignmentNotFound)
		repo.On("GetArms", ctx, experimentID).Return(arms, nil)
		repo.On("GetArmStats", ctx, arm1ID).Return(&service.ArmStats{ArmID: arm1ID, Alpha: 1, Beta: 1}, nil)
		repo.On("CreateAssignment", ctx, mock.MatchedBy(func(assignment *service.Assignment) bool {
			armScores, ok := assignment.Metadata["arm_scores"].([]map[string]interface{})
			return assignment.ExperimentID == experimentID &&
				assignment.UserID == userID &&
				assignment.ArmID == arm1ID &&
				assignment.ExpiresAt.After(assignment.AssignedAt) &&
				assignment.Metadata["selection_strategy"] == "thompson_sampling" &&
				assignment.Metadata["arms_considered"] == len(arms) &&
				assignment.Metadata["selected_arm_name"] == "Arm 1" &&
				assignment.Metadata["selected_sample"] != nil &&
				ok && len(armScores) == 1
		})).Return(nil)

		selectedArmID, err := bandit.SelectArm(ctx, experimentID, userID)

		assert.NoError(t, err)
		assert.Equal(t, arm1ID, selectedArmID)

		// Check that assignment was cached
		assignmentKey := "ab:assign:" + experimentID.String() + ":" + userID.String()
		cachedArmID, _ := cache.GetAssignment(ctx, assignmentKey)
		assert.Equal(t, arm1ID, cachedArmID)
	})

	t.Run("returns typed error when experiment has no arms", func(t *testing.T) {
		repo = new(MockBanditRepository)
		cache = NewMockBanditCache()
		bandit = service.NewThompsonSamplingBandit(repo, cache, logger)

		repo.On("GetActiveAssignment", ctx, experimentID, userID).Return(nil, service.ErrAssignmentNotFound)
		repo.On("GetArms", ctx, experimentID).Return([]service.Arm{}, nil)

		armID, err := bandit.SelectArm(ctx, experimentID, userID)

		assert.ErrorIs(t, err, service.ErrExperimentArmsNotFound)
		assert.Equal(t, uuid.Nil, armID)
		repo.AssertExpectations(t)
	})
}

func TestTrackImpression(t *testing.T) {
	ctx := context.Background()
	repo := new(MockBanditRepository)
	cache := NewMockBanditCache()
	logger := zap.NewNop()
	bandit := service.NewThompsonSamplingBandit(repo, cache, logger)
	experimentID := uuid.New()
	armID := uuid.New()
	userID := uuid.New()

	t.Run("AppendsImmutableImpressionEvent", func(t *testing.T) {
		repo.ExpectedCalls = nil
		repo.Calls = nil
		repo.On("GetArms", ctx, experimentID).Return([]service.Arm{{ID: armID, ExperimentID: experimentID, Name: "Control"}}, nil)
		repo.On("AppendImpressionEvent", ctx, mock.MatchedBy(func(event *service.ImpressionEvent) bool {
			return event.ExperimentID == experimentID &&
				event.ArmID == armID &&
				event.UserID == userID &&
				event.EventType == service.ImpressionEventTypeImpression &&
				event.Metadata["placement"] == "paywall"
		})).Return(nil)

		err := bandit.TrackImpression(ctx, experimentID, armID, userID, &service.ImpressionEvent{Metadata: map[string]interface{}{"placement": "paywall"}})

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("ReturnsNotFoundForUnknownArm", func(t *testing.T) {
		repo.ExpectedCalls = nil
		repo.Calls = nil
		repo.On("GetArms", ctx, experimentID).Return([]service.Arm{{ID: uuid.New(), ExperimentID: experimentID, Name: "Other"}}, nil)

		err := bandit.TrackImpression(ctx, experimentID, armID, userID, nil)

		assert.ErrorIs(t, err, service.ErrBanditArmNotFound)
		repo.AssertExpectations(t)
	})
}

// TestUpdateReward tests the UpdateReward method
func TestUpdateReward(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()
	armID := uuid.New()
	logger := zap.NewNop()

	repo := new(MockBanditRepository)
	cache := NewMockBanditCache()
	bandit := service.NewThompsonSamplingBandit(repo, cache, logger)

	t.Run("increments alpha on positive reward", func(t *testing.T) {
		initialStats := &service.ArmStats{
			ArmID:       armID,
			Alpha:       5.0,
			Beta:        3.0,
			Samples:     8,
			Conversions: 5,
			Revenue:     49.95,
		}

		repo.On("GetArmStats", ctx, armID).Return(initialStats, nil)
		repo.On("UpdateArmStats", ctx, mock.MatchedBy(func(stats *service.ArmStats) bool {
			return stats.Alpha == 6.0 && stats.Beta == 3.0 && stats.Samples == 9
		})).Return(nil)

		err := bandit.UpdateReward(ctx, experimentID, armID, 9.99)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("increments beta on zero/negative reward", func(t *testing.T) {
		initialStats := &service.ArmStats{
			ArmID:   armID,
			Alpha:   5.0,
			Beta:    3.0,
			Samples: 8,
		}

		repo = new(MockBanditRepository)
		bandit = service.NewThompsonSamplingBandit(repo, cache, logger)

		repo.On("GetArmStats", ctx, armID).Return(initialStats, nil)
		repo.On("UpdateArmStats", ctx, mock.MatchedBy(func(stats *service.ArmStats) bool {
			return stats.Alpha == 5.0 && stats.Beta == 4.0 && stats.Samples == 9
		})).Return(nil)

		err := bandit.UpdateReward(ctx, experimentID, armID, 0.0)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("updates cache with new stats", func(t *testing.T) {
		initialStats := &service.ArmStats{
			ArmID:   armID,
			Alpha:   5.0,
			Beta:    3.0,
			Samples: 8,
		}

		repo = new(MockBanditRepository)
		cache = NewMockBanditCache()
		bandit = service.NewThompsonSamplingBandit(repo, cache, logger)

		repo.On("GetArmStats", ctx, armID).Return(initialStats, nil)
		repo.On("UpdateArmStats", ctx, mock.Anything).Return(nil)

		err := bandit.UpdateReward(ctx, experimentID, armID, 9.99)

		assert.NoError(t, err)

		// Check cache was updated
		cacheKey := "ab:arm:" + armID.String()
		cachedStats, err := cache.GetArmStats(ctx, cacheKey)
		assert.NoError(t, err)
		assert.NotNil(t, cachedStats)
	})

	t.Run("appends conversion event when reward metadata is provided", func(t *testing.T) {
		userID := uuid.New()
		initialStats := &service.ArmStats{
			ArmID:       armID,
			Alpha:       2.0,
			Beta:        1.0,
			Samples:     3,
			Conversions: 2,
			Revenue:     8.0,
		}

		repo = new(MockBanditRepository)
		cache = NewMockBanditCache()
		bandit = service.NewThompsonSamplingBandit(repo, cache, logger)

		repo.On("GetArmStats", ctx, armID).Return(initialStats, nil)
		repo.On("UpdateArmStats", ctx, mock.Anything).Return(nil)
		repo.On("AppendConversionEvent", ctx, mock.MatchedBy(func(event *service.ConversionEvent) bool {
			return event.EventType == service.ConversionEventTypeDirectReward &&
				event.UserID != nil && *event.UserID == userID &&
				event.OriginalCurrency == "USD" &&
				event.NormalizedRewardValue == 9.99
		})).Return(nil)

		err := bandit.UpdateRewardWithEvent(ctx, experimentID, armID, 9.99, &service.ConversionEvent{
			UserID:                &userID,
			EventType:             service.ConversionEventTypeDirectReward,
			OriginalRewardValue:   9.99,
			OriginalCurrency:      "USD",
			NormalizedRewardValue: 9.99,
			NormalizedCurrency:    "USD",
		})

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

// TestSampleBeta tests the Beta distribution sampling
func TestSampleBeta(t *testing.T) {
	logger := zap.NewNop()

	t.Run("returns values in valid range", func(t *testing.T) {
		repo := new(MockBanditRepository)
		cache := NewMockBanditCache()
		bandit := service.NewThompsonSamplingBandit(repo, cache, logger)

		// Test various alpha, beta pairs
		testCases := []struct {
			alpha float64
			beta  float64
		}{
			{1.0, 1.0},     // Uniform
			{10.0, 5.0},    // Skewed toward success
			{5.0, 10.0},    // Skewed toward failure
			{100.0, 100.0}, // Peaked around 0.5
		}

		for _, tc := range testCases {
			samples := make([]float64, 1000)
			for i := 0; i < 1000; i++ {
				sample := bandit.SampleBeta(tc.alpha, tc.beta)
				assert.True(t, sample >= 0 && sample <= 1,
					"Sample out of range [0,1] for alpha=%f, beta=%f: %f", tc.alpha, tc.beta, sample)
				samples[i] = sample
			}

			// Calculate mean to verify it's roughly alpha/(alpha+beta)
			mean := 0.0
			for _, s := range samples {
				mean += s
			}
			mean /= float64(len(samples))

			expectedMean := tc.alpha / (tc.alpha + tc.beta)
			assert.InDelta(t, expectedMean, mean, 0.1,
				"Mean mismatch for alpha=%f, beta=%f", tc.alpha, tc.beta)
		}
	})
}

// TestStickyAssignment tests assignment stickiness
func TestStickyAssignment(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()
	userID := uuid.New()
	logger := zap.NewNop()

	repo := new(MockBanditRepository)
	cache := NewMockBanditCache()
	bandit := service.NewThompsonSamplingBandit(repo, cache, logger)

	armID := uuid.New()
	arms := []service.Arm{{ID: armID, Name: "Arm 1"}}

	t.Run("same user gets same arm within 24h", func(t *testing.T) {
		repo.On("GetActiveAssignment", ctx, experimentID, userID).Return(nil, service.ErrAssignmentNotFound)
		repo.On("GetArms", ctx, experimentID).Return(arms, nil)
		repo.On("GetArmStats", ctx, armID).Return(&service.ArmStats{ArmID: armID, Alpha: 1, Beta: 1}, nil)
		repo.On("CreateAssignment", ctx, mock.MatchedBy(func(assignment *service.Assignment) bool {
			return assignment.ExperimentID == experimentID &&
				assignment.UserID == userID &&
				assignment.ArmID == armID &&
				assignment.ExpiresAt.After(assignment.AssignedAt)
		})).Return(nil)

		// First call
		firstArmID, err := bandit.SelectArm(ctx, experimentID, userID)
		assert.NoError(t, err)

		// Mock the active assignment for second call
		existingAssignment := &service.Assignment{
			ID:           uuid.New(),
			ExperimentID: experimentID,
			UserID:       userID,
			ArmID:        firstArmID,
			AssignedAt:   time.Now().Add(-1 * time.Hour),
			ExpiresAt:    time.Now().Add(23 * time.Hour),
		}

		repo = new(MockBanditRepository)
		bandit = service.NewThompsonSamplingBandit(repo, cache, logger)

		repo.On("GetActiveAssignment", ctx, experimentID, userID).Return(existingAssignment, nil)

		// Second call should return same arm
		secondArmID, err := bandit.SelectArm(ctx, experimentID, userID)

		assert.NoError(t, err)
		assert.Equal(t, firstArmID, secondArmID)
	})
}

func containsUUID(ids []uuid.UUID, target uuid.UUID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

// TestConfidenceCalculation tests win probability calculation
func TestConfidenceCalculation(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()
	logger := zap.NewNop()

	repo := new(MockBanditRepository)
	cache := NewMockBanditCache()
	bandit := service.NewThompsonSamplingBandit(repo, cache, logger)

	t.Run("calculates win probabilities", func(t *testing.T) {
		arm1ID := uuid.New()
		arm2ID := uuid.New()

		// Arm 1 has much better performance
		stats1 := &service.ArmStats{ArmID: arm1ID, Alpha: 100, Beta: 10}
		stats2 := &service.ArmStats{ArmID: arm2ID, Alpha: 30, Beta: 70}

		repo.On("GetArms", ctx, experimentID).Return([]service.Arm{
			{ID: arm1ID}, {ID: arm2ID},
		}, nil)

		repo.On("GetArmStats", ctx, arm1ID).Return(stats1, nil)
		repo.On("GetArmStats", ctx, arm2ID).Return(stats2, nil)

		winProbs, err := bandit.CalculateWinProbability(ctx, experimentID, 1000)

		assert.NoError(t, err)
		assert.NotNil(t, winProbs)

		// Arm 1 should have higher win probability
		assert.Greater(t, winProbs[arm1ID], winProbs[arm2ID])

		// Probabilities should sum to approximately 1.0
		totalProb := winProbs[arm1ID] + winProbs[arm2ID]
		assert.InDelta(t, 1.0, totalProb, 0.1)
	})
}

// TestCalculateWinProbability tests the Monte Carlo simulation
func TestCalculateWinProbability(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()
	logger := zap.NewNop()

	repo := new(MockBanditRepository)
	cache := NewMockBanditCache()
	bandit := service.NewThompsonSamplingBandit(repo, cache, logger)

	t.Run("returns valid probabilities with enough simulations", func(t *testing.T) {
		arm1ID := uuid.New()
		arm2ID := uuid.New()

		repo.On("GetArms", ctx, experimentID).Return([]service.Arm{
			{ID: arm1ID}, {ID: arm2ID},
		}, nil)

		repo.On("GetArmStats", ctx, arm1ID).Return(&service.ArmStats{ArmID: arm1ID, Alpha: 50, Beta: 50}, nil)
		repo.On("GetArmStats", ctx, arm2ID).Return(&service.ArmStats{ArmID: arm2ID, Alpha: 50, Beta: 50}, nil)

		winProbs, err := bandit.CalculateWinProbability(ctx, experimentID, 1000)

		assert.NoError(t, err)
		assert.Len(t, winProbs, 2)

		// With identical distributions, probabilities should be roughly equal
		diff := math.Abs(winProbs[arm1ID] - winProbs[arm2ID])
		assert.Less(t, diff, 0.15) // Allow some variance due to Monte Carlo
	})
}

// BenchmarkSelectArm benchmarks the arm selection
func BenchmarkSelectArm(b *testing.B) {
	ctx := context.Background()
	experimentID := uuid.New()
	userID := uuid.New()
	logger := zap.NewNop()

	repo := new(MockBanditRepository)
	cache := NewMockBanditCache()
	bandit := service.NewThompsonSamplingBandit(repo, cache, logger)

	armID := uuid.New()
	arms := []service.Arm{{ID: armID, Name: "Arm 1"}}

	repo.On("GetActiveAssignment", ctx, experimentID, userID).Return(nil, service.ErrAssignmentNotFound)
	repo.On("GetArms", mock.Anything, mock.Anything).Return(arms, nil)
	repo.On("GetArmStats", ctx, mock.Anything).Return(&service.ArmStats{ArmID: armID, Alpha: 1, Beta: 1}, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bandit.SelectArm(ctx, experimentID, userID)
	}
}
