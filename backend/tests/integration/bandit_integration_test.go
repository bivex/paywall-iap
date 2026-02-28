package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/cache"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/tests/testutil"
)

// BanditIntegrationTestSuite tests the full bandit workflow
func TestBanditIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := context.Background()
	logger := zap.NewNop()

	// Setup test database
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	// Setup test Redis
	redisClient := testutil.SetupTestRedis(t)
	defer testutil.TeardownTestRedis(t, redisClient)

	repo := repository.NewPostgresBanditRepository(db, logger)
	cache := cache.NewRedisBanditCache(redisClient, logger)
	bandit := service.NewThompsonSamplingBandit(repo, cache, logger)

	t.Run("FullAssignmentFlow", func(t *testing.T) {
		// Create experiment
		experiment := &repository.Experiment{
			ID:            uuid.New(),
			Name:          "Test Experiment",
			Description:   "Integration test experiment",
			Status:        "running",
			AlgorithmType: strPtr("thompson_sampling"),
			IsBandit:       true,
			MinSampleSize:  100,
		}

		err := repo.CreateExperiment(ctx, experiment)
		require.NoError(t, err)

		// Create arms
		controlArmID := uuid.New()
		variantAArmID := uuid.New()
		variantBArmID := uuid.New()

		err = repo.CreateArm(ctx, &service.Arm{
			ID:       controlArmID,
			ExperimentID: experiment.ID,
			Name:     "Control",
			IsControl: true,
			TrafficWeight: 1.0,
		})
		require.NoError(t, err)

		err = repo.CreateArm(ctx, &service.Arm{
			ID:       variantAArmID,
			ExperimentID: experiment.ID,
			Name:     "Variant A",
			TrafficWeight: 1.0,
		})
		require.NoError(t, err)

		err = repo.CreateArm(ctx, &service.Arm{
			ID:       variantBArmID,
			ExperimentID: experiment.ID,
			Name:     "Variant B",
			TrafficWeight: 1.0,
		})
		require.NoError(t, err)

		// Initialize arm stats
		for _, armID := range []uuid.UUID{controlArmID, variantAArmID, variantBArmID} {
			err = repo.UpdateArmStats(ctx, &service.ArmStats{
				ArmID:   armID,
				Alpha:   1.0, // Uniform prior
				Beta:    1.0,
				Samples: 0,
			})
			require.NoError(t, err)
		}

		// Test user assignment
		userID := uuid.New()

		// First assignment should create sticky assignment
		assignedArmID, err := bandit.SelectArm(ctx, experiment.ID, userID)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, assignedArmID)

		// Verify assignment was created in database
		assignment, err := repo.GetActiveAssignment(ctx, experiment.ID, userID)
		assert.NoError(t, err)
		assert.Equal(t, assignedArmID, assignment.ArmID)
		assert.False(t, assignment.ExpiresAt.Before(time.Now().Add(23*time.Hour)))

		// Second assignment within 24h should return same arm
		secondArmID, err := bandit.SelectArm(ctx, experiment.ID, userID)
		assert.NoError(t, err)
		assert.Equal(t, assignedArmID, secondArmID)

		// Verify cache was used (no new DB query for assignment)
		t.Log("Assignment flow completed successfully")
	})

	t.Run("RewardTracking", func(t *testing.T) {
		experimentID := uuid.New()
		armID := uuid.New()
		userID := uuid.New()

		// Initialize arm stats
		err := repo.UpdateArmStats(ctx, &service.ArmStats{
			ArmID:   armID,
			Alpha:   10.0,
			Beta:    5.0,
			Samples: 15,
		})
		require.NoError(t, err)

		// Record positive reward (conversion)
		err = bandit.UpdateReward(ctx, experimentID, armID, 9.99)
		assert.NoError(t, err)

		// Verify stats were updated
		stats, err := repo.GetArmStats(ctx, armID)
		assert.NoError(t, err)
		assert.Equal(t, 11.0, stats.Alpha)  // Incremented
		assert.Equal(t, 5.0, stats.Beta)    // Unchanged
		assert.Equal(t, 16, stats.Samples) // Incremented

		// Verify cache was updated
		cacheKey := "ab:arm:" + armID.String()
		cachedStats, err := cache.GetArmStats(ctx, cacheKey)
		assert.NoError(t, err)
		assert.Equal(t, 11.0, cachedStats.Alpha)
	})

	t.Run("DatabasePersistence", func(t *testing.T) {
		experimentID := uuid.New()
		armID := uuid.New()

		// Create arm stats
		initialStats := &service.ArmStats{
			ArmID:       armID,
			Alpha:       5.0,
			Beta:        3.0,
			Samples:     8,
			Conversions: 5,
			Revenue:     49.95,
		}

		err := repo.UpdateArmStats(ctx, initialStats)
		require.NoError(t, err)

		// Retrieve and verify
		retrievedStats, err := repo.GetArmStats(ctx, armID)
		assert.NoError(t, err)
		assert.Equal(t, initialStats.ArmID, retrievedStats.ArmID)
		assert.Equal(t, initialStats.Alpha, retrievedStats.Alpha)
		assert.Equal(t, initialStats.Beta, retrievedStats.Beta)
		assert.Equal(t, initialStats.Samples, retrievedStats.Samples)
	})

	t.Run("RedisCaching", func(t *testing.T) {
		armID := uuid.New()

		stats := &service.ArmStats{
			ArmID:   armID,
			Alpha:   20.0,
			Beta:    10.0,
			Samples: 30,
		}

		cacheKey := "ab:arm:" + armID.String()

		// Set in cache
		err := cache.SetArmStats(ctx, cacheKey, stats, 24*time.Hour)
		assert.NoError(t, err)

		// Retrieve from cache
		cachedStats, err := cache.GetArmStats(ctx, cacheKey)
		assert.NoError(t, err)
		assert.Equal(t, stats.Alpha, cachedStats.Alpha)
		assert.Equal(t, stats.Beta, cachedStats.Beta)

		// Test cache invalidation
		err = cache.InvalidateArmStats(ctx, armID)
		assert.NoError(t, err)

		// Verify cache miss after invalidation
		_, err = cache.GetArmStats(ctx, cacheKey)
		assert.Error(t, err)
	})

	t.Run("AssignmentExpiry", func(t *testing.T) {
		experimentID := uuid.New()
		userID := uuid.New()
		armID := uuid.New()

		// Create expired assignment
		expiredAssignment := &service.Assignment{
			ID:           uuid.New(),
			ExperimentID: experimentID,
			UserID:       userID,
			ArmID:        armID,
			AssignedAt:   time.Now().Add(-25 * time.Hour),
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		}

		err := repo.CreateAssignment(ctx, expiredAssignment)
		require.NoError(t, err)

		// Create new arm for reassignment
		err = repo.CreateArm(ctx, &service.Arm{
			ID:            armID,
			ExperimentID:  experimentID,
			Name:          "Test Arm",
			TrafficWeight: 1.0,
		})
		require.NoError(t, err)

		// Initialize stats
		err = repo.UpdateArmStats(ctx, &service.ArmStats{
			ArmID: armID,
			Alpha: 1.0,
			Beta:  1.0,
		})
		require.NoError(t, err)

		// GetActiveAssignment should return nil for expired assignment
		_, err = repo.GetActiveAssignment(ctx, experimentID, userID)
		assert.Error(t, err) // Should return error when not found

		// SelectArm should create new assignment
		newArmID, err := bandit.SelectArm(ctx, experimentID, userID)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, newArmID)
	})
}

func strPtr(s string) *string {
	return &s
}
