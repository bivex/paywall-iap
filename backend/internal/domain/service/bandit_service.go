package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BanditRepository defines the interface for bandit data persistence
type BanditRepository interface {
	GetArms(ctx context.Context, experimentID uuid.UUID) ([]Arm, error)
	GetArmStats(ctx context.Context, armID uuid.UUID) (*ArmStats, error)
	UpdateArmStats(ctx context.Context, stats *ArmStats) error
	CreateAssignment(ctx context.Context, assignment *Assignment) error
	GetActiveAssignment(ctx context.Context, experimentID, userID uuid.UUID) (*Assignment, error)
}

// BanditCache defines the interface for caching bandit state
type BanditCache interface {
	GetArmStats(ctx context.Context, key string) (*ArmStats, error)
	SetArmStats(ctx context.Context, key string, stats *ArmStats, ttl time.Duration) error
	GetAssignment(ctx context.Context, key string) (uuid.UUID, error)
	SetAssignment(ctx context.Context, key string, armID uuid.UUID, ttl time.Duration) error
}

// Arm represents an experiment arm (variant)
type Arm struct {
	ID            uuid.UUID
	ExperimentID  uuid.UUID
	Name          string
	Description   string
	IsControl     bool
	TrafficWeight float64
}

// ArmStats represents the statistics for an arm
type ArmStats struct {
	ArmID       uuid.UUID
	Alpha       float64
	Beta        float64
	Samples     int
	Conversions int
	Revenue     float64
	AvgReward   float64
	UpdatedAt   time.Time
}

// Assignment represents a user's assignment to an arm
type Assignment struct {
	ID           uuid.UUID
	ExperimentID uuid.UUID
	UserID       uuid.UUID
	ArmID        uuid.UUID
	AssignedAt   time.Time
	ExpiresAt    time.Time
}

// ThompsonSamplingBandit implements the Thompson Sampling algorithm
type ThompsonSamplingBandit struct {
	repo   BanditRepository
	cache  BanditCache
	logger *zap.Logger
	rng    *rand.Rand
}

// NewThompsonSamplingBandit creates a new Thompson Sampling bandit service
func NewThompsonSamplingBandit(
	repo BanditRepository,
	cache BanditCache,
	logger *zap.Logger,
) *ThompsonSamplingBandit {
	source := rand.NewSource(time.Now().UnixNano())
	return &ThompsonSamplingBandit{
		repo:   repo,
		cache:  cache,
		logger: logger,
		rng:    rand.New(source),
	}
}

// SelectArm selects the best arm using Thompson Sampling
// Returns the arm ID that maximizes the sampled Beta distribution
func (b *ThompsonSamplingBandit) SelectArm(ctx context.Context, experimentID, userID uuid.UUID) (uuid.UUID, error) {
	// First, check if user has an active assignment (sticky assignment)
	if assignment, err := b.repo.GetActiveAssignment(ctx, experimentID, userID); err == nil && assignment != nil {
		b.logger.Debug("Using existing assignment",
			zap.String("experiment_id", experimentID.String()),
			zap.String("user_id", userID.String()),
			zap.String("arm_id", assignment.ArmID.String()),
		)
		return assignment.ArmID, nil
	}

	// Get all arms for this experiment
	arms, err := b.repo.GetArms(ctx, experimentID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get arms: %w", err)
	}

	if len(arms) == 0 {
		return uuid.Nil, fmt.Errorf("no arms found for experiment %s", experimentID)
	}

	var bestArm *Arm
	maxSample := -1.0

	// Sample from Beta distribution for each arm and select the max
	for _, arm := range arms {
		// Get current statistics from cache or DB
		cacheKey := fmt.Sprintf("ab:arm:%s", arm.ID.String())
		stats, err := b.cache.GetArmStats(ctx, cacheKey)
		if err != nil {
			// Fallback to database
			stats, err = b.repo.GetArmStats(ctx, arm.ID)
			if err != nil {
				b.logger.Warn("Failed to get arm stats, using defaults",
					zap.String("arm_id", arm.ID.String()),
					zap.Error(err),
				)
				// Use default Beta(1,1) = uniform prior
				stats = &ArmStats{
					ArmID: arm.ID,
					Alpha: 1.0,
					Beta:  1.0,
				}
			}
		}

		// Sample from Beta(alpha, beta)
		sample := b.sampleBeta(stats.Alpha, stats.Beta)

		b.logger.Debug("Arm sample",
			zap.String("arm_id", arm.ID.String()),
			zap.String("arm_name", arm.Name),
			zap.Float64("alpha", stats.Alpha),
			zap.Float64("beta", stats.Beta),
			zap.Float64("sample", sample),
		)

		if sample > maxSample {
			maxSample = sample
			bestArm = &arm
		}
	}

	if bestArm == nil {
		// Fallback: select random arm
		bestArm = &arms[b.rng.Intn(len(arms))]
	}

	// Create sticky assignment in cache
	cacheKey = fmt.Sprintf("ab:assign:%s:%s", experimentID.String(), userID.String())
	if err := b.cache.SetAssignment(ctx, cacheKey, bestArm.ID, 24*time.Hour); err != nil {
		b.logger.Warn("Failed to cache assignment", zap.Error(err))
	}

	return bestArm.ID, nil
}

// UpdateReward updates the alpha/beta parameters for the selected arm
// reward > 0 counts as a conversion (alpha increment)
// reward <= 0 counts as a non-conversion (beta increment)
func (b *ThompsonSamplingBandit) UpdateReward(ctx context.Context, experimentID, armID uuid.UUID, reward float64) error {
	// Get current stats
	stats, err := b.repo.GetArmStats(ctx, armID)
	if err != nil {
		return fmt.Errorf("failed to get arm stats: %w", err)
	}

	// Update alpha/beta based on reward
	// In Thompson Sampling for conversion rate:
	// - Success (conversion): increment alpha
	// - Failure (no conversion): increment beta
	if reward > 0 {
		stats.Alpha += 1.0
		stats.Conversions++
		stats.Revenue += reward
	} else {
		stats.Beta += 1.0
	}
	stats.Samples++

	// Calculate average reward
	if stats.Samples > 0 {
		stats.AvgReward = stats.Revenue / float64(stats.Samples)
	}

	// Save to database
	if err := b.repo.UpdateArmStats(ctx, stats); err != nil {
		return fmt.Errorf("failed to update arm stats: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("ab:arm:%s", armID.String())
	if err := b.cache.SetArmStats(ctx, cacheKey, stats, 24*time.Hour); err != nil {
		b.logger.Warn("Failed to update cache", zap.Error(err))
	}

	b.logger.Debug("Reward updated",
		zap.String("arm_id", armID.String()),
		zap.Float64("reward", reward),
		zap.Float64("alpha", stats.Alpha),
		zap.Float64("beta", stats.Beta),
		zap.Int("samples", stats.Samples),
	)

	return nil
}

// sampleBeta generates a random sample from Beta(α, β)
// Uses Marsaglia and Tsang's method for alpha,beta >= 1
// Falls back to simple uniform for small parameters
func (b *ThompsonSamplingBandit) sampleBeta(alpha, beta float64) float64 {
	// Handle edge cases
	if alpha <= 0 || beta <= 0 {
		return b.rng.Float64()
	}

	// For small parameters, use simple approximation
	if alpha < 1 && beta < 1 {
		// Johnk's method for alpha,beta < 1
		for {
			u1 := b.rng.Float64()
			u2 := b.rng.Float64()
			if u1 == 0 || u2 == 0 {
				continue
			}
			x := math.Pow(u1, 1/alpha)
			y := math.Pow(u2, 1/beta)
			if x+y <= 1 {
				return x / (x + y)
			}
		}
	}

	if alpha < 1 {
		// For alpha < 1, beta >= 1
		return b.sampleBeta(alpha+1, beta) * math.Pow(b.rng.Float64(), 1/alpha)
	}

	if beta < 1 {
		// For beta < 1, alpha >= 1
		return b.sampleBeta(alpha, beta+1) * math.Pow(b.rng.Float64(), 1/beta)
	}

	// Marsaglia-Tsang method for alpha,beta >= 1
	const (
		// Number of iterations for refinement
		iterations = 3
	)

	a := alpha - 1
	b_param := beta - 1

	// Gamma distribution parameters using Marsaglia-Tsang
 theta := 1.0
	if a <= b_param {
		theta = a / (a + b_param)
	}

	for i := 0; i < iterations; i++ {
		u := b.rng.Float64()
		v := b.rng.Float64()

		// Generate from Gamma(alpha,1)
		gamma := -math.Log(u)
		if gamma > 0 {
			gamma = math.Pow(gamma, 1/alpha)
		}

		// Generate from Gamma(beta,1)
		gamma2 := -math.Log(v)
		if gamma2 > 0 {
			gamma2 = math.Pow(gamma2, 1/beta)
		}

		if gamma+gamma2 > 0 {
			return gamma / (gamma + gamma2)
		}

		// Retry if failed
	}

	// Fallback: Cheng's method
	x := theta
	for {
		u := b.rng.Float64()
		v := b.rng.Float64()

		if u == 0 || v == 0 {
			continue
		}

		w := math.Pow(v, 1/beta)
		x = math.Pow(w/(1+w), 1/alpha)

		if x <= 0 || x >= 1 {
			continue
		}

		// Acceptance-rejection
		lhs := math.Pow(1-x, b_param)
		rhs := math.Pow(x, a-1)

		if u <= lhs*rhs {
			break
		}
	}

	return x
}

// GetArmStatistics returns the current statistics for all arms in an experiment
func (b *ThompsonSamplingBandit) GetArmStatistics(ctx context.Context, experimentID uuid.UUID) (map[uuid.UUID]*ArmStats, error) {
	arms, err := b.repo.GetArms(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	stats := make(map[uuid.UUID]*ArmStats)
	for _, arm := range arms {
		armStats, err := b.repo.GetArmStats(ctx, arm.ID)
		if err != nil {
			b.logger.Warn("Failed to get arm stats", zap.String("arm_id", arm.ID.String()))
			continue
		}
		stats[arm.ID] = armStats
	}

	return stats, nil
}

// CalculateWinProbability calculates the probability that each arm is the best
// using Monte Carlo simulation of Beta distributions
func (b *ThompsonSamplingBandit) CalculateWinProbability(ctx context.Context, experimentID uuid.UUID, simulations int) (map[uuid.UUID]float64, error) {
	arms, err := b.repo.GetArms(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	// Get stats for all arms
	armStats := make([]*ArmStats, 0, len(arms))
	for _, arm := range arms {
		stats, err := b.repo.GetArmStats(ctx, arm.ID)
		if err != nil {
			return nil, err
		}
		stats.ArmID = arm.ID
		armStats = append(armStats, stats)
	}

	// Monte Carlo simulation
	winCounts := make(map[uuid.UUID]int)
	for _, stats := range armStats {
		winCounts[stats.ArmID] = 0
	}

	for i := 0; i < simulations; i++ {
		var bestArmID uuid.UUID
		maxSample := -1.0

		for _, stats := range armStats {
			sample := b.sampleBeta(stats.Alpha, stats.Beta)
			if sample > maxSample {
				maxSample = sample
				bestArmID = stats.ArmID
			}
		}

		if bestArmID != uuid.Nil {
			winCounts[bestArmID]++
		}
	}

	// Convert to probabilities
	winProbs := make(map[uuid.UUID]float64)
	for armID, count := range winCounts {
		winProbs[armID] = float64(count) / float64(simulations)
	}

	return winProbs, nil
}
