package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ErrAssignmentNotFound is returned when no active assignment is found for a user
var ErrAssignmentNotFound = errors.New("assignment not found")

// ErrExperimentArmsNotFound is returned when an experiment has no available arms.
var ErrExperimentArmsNotFound = errors.New("experiment arms not found")

// ErrBanditArmNotFound is returned when a reward references a non-existent arm.
var ErrBanditArmNotFound = errors.New("bandit arm not found")

// BanditRepository defines the interface for bandit data persistence
type BanditRepository interface {
	GetArms(ctx context.Context, experimentID uuid.UUID) ([]Arm, error)
	GetArmStats(ctx context.Context, armID uuid.UUID) (*ArmStats, error)
	UpdateArmStats(ctx context.Context, stats *ArmStats) error
	CreateAssignment(ctx context.Context, assignment *Assignment) error
	GetActiveAssignment(ctx context.Context, experimentID, userID uuid.UUID) (*Assignment, error)

	// Advanced bandit methods
	GetExperimentConfig(ctx context.Context, experimentID uuid.UUID) (*ExperimentConfig, error)
	UpdateObjectiveConfig(ctx context.Context, experimentID uuid.UUID, objectiveType ObjectiveType, objectiveWeights map[string]float64) error
	GetUserContext(ctx context.Context, userID uuid.UUID) (*UserContext, error)
	SetUserContext(ctx context.Context, uctx *UserContext) error
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
	Metadata     map[string]interface{}
}

type AssignmentEventType string

const (
	AssignmentEventTypeAssigned AssignmentEventType = "assigned"
)

type ImpressionEventType string

const (
	ImpressionEventTypeImpression ImpressionEventType = "impression"
)

// =====================================================
// Advanced Bandit Plugin Interfaces
// =====================================================

// RewardStrategy defines how rewards are calculated and recorded
type RewardStrategy interface {
	CalculateReward(ctx context.Context, baseReward float64, arm Arm, userContext UserContext) (float64, error)
	GetType() string
}

// SelectionStrategy defines how arms are selected
type SelectionStrategy interface {
	SelectArm(ctx context.Context, arms []Arm, userContext UserContext) (*Arm, error)
	GetName() string
}

// WindowStrategy defines how historical data is windowed
type WindowStrategy interface {
	GetArmStats(ctx context.Context, armID uuid.UUID) (*ArmStats, error)
	RecordEvent(ctx context.Context, armID uuid.UUID, event RewardEvent) error
	GetType() string
}

// UserContext captures user attributes for contextual bandits
type UserContext struct {
	UserID           uuid.UUID
	Country          string
	Device           string
	AppVersion       string
	DaysSinceInstall int
	TotalSpent       float64
	LastPurchaseAt   *time.Time
	CustomFeatures   map[string]interface{}
}

// RewardEvent represents a reward event with metadata
type RewardEvent struct {
	UserID          uuid.UUID
	ArmID           uuid.UUID
	RewardValue     float64
	Currency        string
	Timestamp       time.Time
	ConversionDelay *time.Duration
	Metadata        map[string]interface{}
}

type ConversionEventType string

const (
	ConversionEventTypeDirectReward         ConversionEventType = "direct_reward"
	ConversionEventTypeDelayedConversion    ConversionEventType = "delayed_conversion"
	ConversionEventTypeExpiredPendingReward ConversionEventType = "expired_pending_reward"
)

type ConversionEvent struct {
	ExperimentID          uuid.UUID
	ArmID                 uuid.UUID
	UserID                *uuid.UUID
	PendingRewardID       *uuid.UUID
	TransactionID         *uuid.UUID
	EventType             ConversionEventType
	OriginalRewardValue   float64
	OriginalCurrency      string
	NormalizedRewardValue float64
	NormalizedCurrency    string
	Metadata              map[string]interface{}
	OccurredAt            time.Time
}

type conversionEventAppender interface {
	AppendConversionEvent(ctx context.Context, event *ConversionEvent) error
}

type ImpressionEvent struct {
	ExperimentID uuid.UUID
	ArmID        uuid.UUID
	UserID       uuid.UUID
	EventType    ImpressionEventType
	Metadata     map[string]interface{}
	OccurredAt   time.Time
}

type impressionEventAppender interface {
	AppendImpressionEvent(ctx context.Context, event *ImpressionEvent) error
}

// ObjectiveType defines the optimization objective
type ObjectiveType string

const (
	ObjectiveConversion ObjectiveType = "conversion"
	ObjectiveLTV        ObjectiveType = "ltv"
	ObjectiveRevenue    ObjectiveType = "revenue"
	ObjectiveHybrid     ObjectiveType = "hybrid"
)

// WindowType defines the windowing strategy
type WindowType string

const (
	WindowTypeEvents WindowType = "events"
	WindowTypeTime   WindowType = "time"
	WindowTypeNone   WindowType = "none"
)

// WindowConfig configures sliding window behavior
type WindowConfig struct {
	Type       WindowType
	Size       int // Number of events or seconds
	MinSamples int // Minimum samples before using window
}

// ExperimentConfig defines per-experiment configuration for advanced features
type ExperimentConfig struct {
	ID               uuid.UUID
	ObjectiveType    ObjectiveType
	ObjectiveWeights map[string]float64 // For hybrid: {"conversion": 0.5, "ltv": 0.3, "revenue": 0.2}
	WindowConfig     *WindowConfig
	EnableContextual bool
	EnableDelayed    bool
	EnableCurrency   bool
	ExplorationAlpha float64 // For LinUCB: exploration parameter
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
		return uuid.Nil, fmt.Errorf("%w: %s", ErrExperimentArmsNotFound, experimentID)
	}

	var bestArm *Arm
	maxSample := -1.0
	armScores := make([]map[string]interface{}, 0, len(arms))

	// Sample from Beta distribution for each arm and select the max
	for _, arm := range arms {
		// Get current statistics from cache or DB
		cacheKey := fmt.Sprintf("ab:arm:%s", arm.ID.String())
		statsSource := "cache"
		stats, err := b.cache.GetArmStats(ctx, cacheKey)
		if err != nil {
			// Fallback to database
			statsSource = "database"
			stats, err = b.repo.GetArmStats(ctx, arm.ID)
			if err != nil {
				b.logger.Warn("Failed to get arm stats, using defaults",
					zap.String("arm_id", arm.ID.String()),
					zap.Error(err),
				)
				// Use default Beta(1,1) = uniform prior
				statsSource = "default_prior"
				stats = &ArmStats{
					ArmID: arm.ID,
					Alpha: 1.0,
					Beta:  1.0,
				}
			}
		}

		// Sample from Beta(alpha, beta)
		sample := b.SampleBeta(stats.Alpha, stats.Beta)

		b.logger.Debug("Arm sample",
			zap.String("arm_id", arm.ID.String()),
			zap.String("arm_name", arm.Name),
			zap.Float64("alpha", stats.Alpha),
			zap.Float64("beta", stats.Beta),
			zap.Float64("sample", sample),
		)

		armScores = append(armScores, map[string]interface{}{
			"arm_id":       arm.ID,
			"arm_name":     arm.Name,
			"is_control":   arm.IsControl,
			"stats_source": statsSource,
			"alpha":        stats.Alpha,
			"beta":         stats.Beta,
			"samples":      stats.Samples,
			"conversions":  stats.Conversions,
			"revenue":      stats.Revenue,
			"sample":       sample,
		})

		if sample > maxSample {
			maxSample = sample
			bestArm = &arm
		}
	}

	if bestArm == nil {
		// Fallback: select random arm
		bestArm = &arms[b.rng.Intn(len(arms))]
	}

	assignedAt := time.Now().UTC()
	assignment := &Assignment{
		ID:           uuid.New(),
		ExperimentID: experimentID,
		UserID:       userID,
		ArmID:        bestArm.ID,
		AssignedAt:   assignedAt,
		ExpiresAt:    assignedAt.Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"selection_strategy": "thompson_sampling",
			"arms_considered":    len(arms),
			"selected_arm_name":  bestArm.Name,
			"selected_sample":    maxSample,
			"arm_scores":         armScores,
		},
	}
	if err := b.repo.CreateAssignment(ctx, assignment); err != nil {
		return uuid.Nil, fmt.Errorf("failed to persist assignment: %w", err)
	}

	// Create sticky assignment in cache
	cacheKey := fmt.Sprintf("ab:assign:%s:%s", experimentID.String(), userID.String())
	if err := b.cache.SetAssignment(ctx, cacheKey, bestArm.ID, 24*time.Hour); err != nil {
		b.logger.Warn("Failed to cache assignment", zap.Error(err))
	}

	return bestArm.ID, nil
}

// UpdateReward updates the alpha/beta parameters for the selected arm
// reward > 0 counts as a conversion (alpha increment)
// reward <= 0 counts as a non-conversion (beta increment)
func (b *ThompsonSamplingBandit) UpdateReward(ctx context.Context, experimentID, armID uuid.UUID, reward float64) error {
	return b.UpdateRewardWithEvent(ctx, experimentID, armID, reward, nil)
}

func (b *ThompsonSamplingBandit) TrackImpression(
	ctx context.Context,
	experimentID, armID, userID uuid.UUID,
	event *ImpressionEvent,
) error {
	arms, err := b.repo.GetArms(ctx, experimentID)
	if err != nil {
		return err
	}
	if len(arms) == 0 {
		return ErrExperimentArmsNotFound
	}

	armFound := false
	for _, arm := range arms {
		if arm.ID == armID {
			armFound = true
			break
		}
	}
	if !armFound {
		return ErrBanditArmNotFound
	}

	appender, ok := b.repo.(impressionEventAppender)
	if !ok {
		return fmt.Errorf("impression event logging not supported")
	}

	normalizedEvent := ImpressionEvent{
		ExperimentID: experimentID,
		ArmID:        armID,
		UserID:       userID,
		EventType:    ImpressionEventTypeImpression,
		OccurredAt:   time.Now().UTC(),
	}
	if event != nil {
		normalizedEvent = *event
		if normalizedEvent.ExperimentID == uuid.Nil {
			normalizedEvent.ExperimentID = experimentID
		}
		if normalizedEvent.ArmID == uuid.Nil {
			normalizedEvent.ArmID = armID
		}
		if normalizedEvent.UserID == uuid.Nil {
			normalizedEvent.UserID = userID
		}
		if normalizedEvent.EventType == "" {
			normalizedEvent.EventType = ImpressionEventTypeImpression
		}
		if normalizedEvent.OccurredAt.IsZero() {
			normalizedEvent.OccurredAt = time.Now().UTC()
		}
	}

	if err := appender.AppendImpressionEvent(ctx, &normalizedEvent); err != nil {
		return fmt.Errorf("failed to append impression event: %w", err)
	}

	return nil
}

func (b *ThompsonSamplingBandit) UpdateRewardWithEvent(
	ctx context.Context,
	experimentID, armID uuid.UUID,
	reward float64,
	event *ConversionEvent,
) error {
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

	if event != nil {
		if appender, ok := b.repo.(conversionEventAppender); ok {
			normalizedEvent := *event
			if normalizedEvent.ExperimentID == uuid.Nil {
				normalizedEvent.ExperimentID = experimentID
			}
			if normalizedEvent.ArmID == uuid.Nil {
				normalizedEvent.ArmID = armID
			}
			if normalizedEvent.EventType == "" {
				normalizedEvent.EventType = ConversionEventTypeDirectReward
			}
			if normalizedEvent.OccurredAt.IsZero() {
				normalizedEvent.OccurredAt = time.Now().UTC()
			}
			if normalizedEvent.NormalizedRewardValue == 0 {
				normalizedEvent.NormalizedRewardValue = reward
			}
			if normalizedEvent.OriginalRewardValue == 0 {
				normalizedEvent.OriginalRewardValue = reward
			}
			if normalizedEvent.NormalizedCurrency == "" {
				normalizedEvent.NormalizedCurrency = normalizedEvent.OriginalCurrency
			}

			if err := appender.AppendConversionEvent(ctx, &normalizedEvent); err != nil {
				return fmt.Errorf("failed to append conversion event: %w", err)
			}
		}
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

// SampleBeta generates a random sample from Beta(α, β)
// Uses Marsaglia and Tsang's method for alpha,beta >= 1
// Falls back to simple uniform for small parameters
func (b *ThompsonSamplingBandit) SampleBeta(alpha, beta float64) float64 {
	// Handle edge cases
	if alpha <= 0 || beta <= 0 {
		return b.rng.Float64()
	}

	// For small parameters, use simple approximation
	if alpha < 1 && beta < 1 {
		return b.sampleBetaJohnk(alpha, beta)
	}

	if alpha < 1 {
		// For alpha < 1, beta >= 1
		return b.SampleBeta(alpha+1, beta) * math.Pow(b.rng.Float64(), 1/alpha)
	}

	if beta < 1 {
		// For beta < 1, alpha >= 1
		return b.SampleBeta(alpha, beta+1) * math.Pow(b.rng.Float64(), 1/beta)
	}

	// Try Marsaglia-Tsang method for alpha,beta >= 1
	if sample := b.sampleBetaMarsagliaTsang(alpha, beta); sample >= 0 {
		return sample
	}

	// Fallback: Cheng's method
	return b.sampleBetaCheng(alpha, beta)
}

// sampleBetaJohnk implements Johnk's method for alpha,beta < 1
func (b *ThompsonSamplingBandit) sampleBetaJohnk(alpha, beta float64) float64 {
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

// sampleBetaMarsagliaTsang implements Marsaglia-Tsang method for alpha,beta >= 1
// Returns -1 if sampling fails
func (b *ThompsonSamplingBandit) sampleBetaMarsagliaTsang(alpha, beta float64) float64 {
	const iterations = 3

	for i := 0; i < iterations; i++ {
		u := b.rng.Float64()
		v := b.rng.Float64()

		gamma := b.sampleGamma(alpha, u)
		gamma2 := b.sampleGamma(beta, v)

		if gamma+gamma2 > 0 {
			return gamma / (gamma + gamma2)
		}
	}

	return -1 // Indicate failure
}

// sampleGamma generates a sample from Gamma(shape, 1) using logarithm
func (b *ThompsonSamplingBandit) sampleGamma(shape, u float64) float64 {
	gamma := -math.Log(u)
	if gamma > 0 {
		gamma = math.Pow(gamma, 1/shape)
	}
	return gamma
}

// sampleBetaCheng implements Cheng's method as a fallback
func (b *ThompsonSamplingBandit) sampleBetaCheng(alpha, beta float64) float64 {
	a := alpha - 1
	bParam := beta - 1

	// Initial theta value
	theta := 1.0
	if a <= bParam {
		theta = a / (a + bParam)
	}

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
		lhs := math.Pow(1-x, bParam)
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
			sample := b.SampleBeta(stats.Alpha, stats.Beta)
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
