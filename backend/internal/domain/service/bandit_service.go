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

// BanditCoreRepository defines arm stats and assignment operations
type BanditCoreRepository interface {
	GetArms(ctx context.Context, experimentID uuid.UUID) ([]Arm, error)
	GetArmStats(ctx context.Context, armID uuid.UUID) (*ArmStats, error)
	UpdateArmStats(ctx context.Context, stats *ArmStats) error
	CreateAssignment(ctx context.Context, assignment *Assignment) error
	GetActiveAssignment(ctx context.Context, experimentID, userID uuid.UUID) (*Assignment, error)
}

// BanditContextRepository defines contextual experiment config and user context persistence
type BanditContextRepository interface {
	GetExperimentConfig(ctx context.Context, experimentID uuid.UUID) (*ExperimentConfig, error)
	UpdateObjectiveConfig(ctx context.Context, experimentID uuid.UUID, objectiveType ObjectiveType, objectiveWeights map[string]float64) error
	GetUserContext(ctx context.Context, userID uuid.UUID) (*UserContext, error)
	SetUserContext(ctx context.Context, uctx *UserContext) error
}

// BanditRepository defines the composite interface for bandit data persistence
type BanditRepository interface {
	BanditCoreRepository
	BanditContextRepository
}

// BanditCache defines the interface for caching bandit state
type BanditCache interface {
	GetArmStats(ctx context.Context, key string) (*ArmStats, error)
	SetArmStats(ctx context.Context, key string, stats *ArmStats, ttl time.Duration) error
	GetAssignment(ctx context.Context, key string) (uuid.UUID, error)
	SetAssignment(ctx context.Context, key string, armID uuid.UUID, ttl time.Duration) error
	// Generic JSON cache for arbitrary structs (pending rewards, etc.)
	SetBytes(ctx context.Context, key string, data []byte, ttl time.Duration) error
	GetBytes(ctx context.Context, key string) ([]byte, error)
	DeleteKey(ctx context.Context, key string) error
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

// BetaDistributionSampler samples from Beta and Gamma distributions
type BetaDistributionSampler struct {
	rng *rand.Rand
}

// BanditArmSelector selects bandit arms using Thompson Sampling
type BanditArmSelector struct {
	repo    BanditRepository
	cache   BanditCache
	logger  *zap.Logger
	sampler *BetaDistributionSampler
}

// BanditRewardTracker tracks impressions, rewards, and conversion events
type BanditRewardTracker struct {
	repo   BanditRepository
	cache  BanditCache
	logger *zap.Logger
}

// BanditStatsCalculator calculates arm statistics and Monte Carlo win probabilities
type BanditStatsCalculator struct {
	repo    BanditRepository
	cache   BanditCache
	logger  *zap.Logger
	sampler *BetaDistributionSampler
}

// ThompsonSamplingBandit implements the Thompson Sampling algorithm
type ThompsonSamplingBandit struct {
	*BetaDistributionSampler
	*BanditArmSelector
	*BanditRewardTracker
	*BanditStatsCalculator
}

// NewThompsonSamplingBandit creates a new Thompson Sampling bandit service
func NewThompsonSamplingBandit(
	repo BanditRepository,
	cache BanditCache,
	logger *zap.Logger,
) *ThompsonSamplingBandit {
	source := rand.NewSource(time.Now().UnixNano())
	sampler := &BetaDistributionSampler{rng: rand.New(source)}
	armSelector := &BanditArmSelector{
		repo:    repo,
		cache:   cache,
		logger:  logger,
		sampler: sampler,
	}
	rewardTracker := &BanditRewardTracker{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
	statsCalculator := &BanditStatsCalculator{
		repo:    repo,
		cache:   cache,
		logger:  logger,
		sampler: sampler,
	}
	return &ThompsonSamplingBandit{
		BetaDistributionSampler: sampler,
		BanditArmSelector:       armSelector,
		BanditRewardTracker:     rewardTracker,
		BanditStatsCalculator:   statsCalculator,
	}
}

// SelectArm selects the best arm using Thompson Sampling
// Returns the arm ID that maximizes the sampled Beta distribution
func (b *BanditArmSelector) SelectArm(ctx context.Context, experimentID, userID uuid.UUID) (uuid.UUID, error) {
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

	bestArm, maxSample, armScores := b.sampleCandidateArms(ctx, arms)
	if err := b.persistAndCacheAssignment(ctx, persistAssignmentParams{
		experimentID: experimentID,
		userID:       userID,
		bestArm:      bestArm,
		maxSample:    maxSample,
		armScores:    armScores,
		armsCount:    len(arms),
	}); err != nil {
		return uuid.Nil, err
	}

	return bestArm.ID, nil
}

func (b *BanditArmSelector) sampleCandidateArms(ctx context.Context, arms []Arm) (*Arm, float64, []map[string]interface{}) {
	var bestArm *Arm
	maxSample := -1.0
	armScores := make([]map[string]interface{}, 0, len(arms))

	for _, arm := range arms {
		stats, statsSource := b.resolveArmStatsForSampling(ctx, arm)
		sample := b.sampler.SampleBeta(stats.Alpha, stats.Beta)

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
		bestArm = &arms[b.sampler.rng.Intn(len(arms))]
	}
	return bestArm, maxSample, armScores
}

func (b *BanditArmSelector) resolveArmStatsForSampling(ctx context.Context, arm Arm) (*ArmStats, string) {
	cacheKey := fmt.Sprintf("ab:arm:%s", arm.ID.String())
	stats, err := b.cache.GetArmStats(ctx, cacheKey)
	if err == nil && stats != nil {
		return stats, "cache"
	}

	stats, err = b.repo.GetArmStats(ctx, arm.ID)
	if err == nil && stats != nil {
		return stats, "database"
	}

	b.logger.Warn("Failed to get arm stats, using defaults",
		zap.String("arm_id", arm.ID.String()),
		zap.Error(err),
	)
	return &ArmStats{
		ArmID: arm.ID,
		Alpha: 1.0,
		Beta:  1.0,
	}, "default_prior"
}

type persistAssignmentParams struct {
	experimentID uuid.UUID
	userID       uuid.UUID
	bestArm      *Arm
	maxSample    float64
	armScores    []map[string]interface{}
	armsCount    int
}

func (b *BanditArmSelector) persistAndCacheAssignment(
	ctx context.Context,
	p persistAssignmentParams,
) error {
	assignedAt := time.Now().UTC()
	assignment := &Assignment{
		ID:           uuid.New(),
		ExperimentID: p.experimentID,
		UserID:       p.userID,
		ArmID:        p.bestArm.ID,
		AssignedAt:   assignedAt,
		ExpiresAt:    assignedAt.Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"selection_strategy": "thompson_sampling",
			"arms_considered":    p.armsCount,
			"selected_arm_name":  p.bestArm.Name,
			"selected_sample":    p.maxSample,
			"arm_scores":         p.armScores,
		},
	}
	if err := b.repo.CreateAssignment(ctx, assignment); err != nil {
		return fmt.Errorf("failed to persist assignment: %w", err)
	}

	cacheKey := fmt.Sprintf("ab:assign:%s:%s", p.experimentID.String(), p.userID.String())
	if err := b.cache.SetAssignment(ctx, cacheKey, p.bestArm.ID, 24*time.Hour); err != nil {
		b.logger.Warn("Failed to cache assignment", zap.Error(err))
	}
	return nil
}

// SelectArmWithMeta returns the assigned arm ID and whether it was a new assignment
func (b *BanditArmSelector) SelectArmWithMeta(ctx context.Context, experimentID, userID uuid.UUID) (uuid.UUID, bool, error) {
	// Check for existing assignment first
	if assignment, err := b.repo.GetActiveAssignment(ctx, experimentID, userID); err == nil && assignment != nil {
		return assignment.ArmID, false, nil
	}
	armID, err := b.SelectArm(ctx, experimentID, userID)
	return armID, err == nil, err
}

// TrackImpressionParams contains parameters for tracking an arm impression.
type TrackImpressionParams struct {
	ExperimentID uuid.UUID
	ArmID        uuid.UUID
	UserID       uuid.UUID
	Event        *ImpressionEvent
}

// RewardWithEventParams contains parameters for updating an arm's reward along with a conversion event.
type RewardWithEventParams struct {
	ExperimentID uuid.UUID
	ArmID        uuid.UUID
	Reward       float64
	Event        *ConversionEvent
}

// UpdateReward updates the alpha/beta parameters for the selected arm
// reward > 0 counts as a conversion (alpha increment)
// reward <= 0 counts as a non-conversion (beta increment)
func (b *BanditRewardTracker) UpdateReward(ctx context.Context, experimentID, armID uuid.UUID, reward float64) error {
	return b.UpdateRewardWithEvent(ctx, RewardWithEventParams{
		ExperimentID: experimentID,
		ArmID:        armID,
		Reward:       reward,
		Event:        nil,
	})
}

func (b *BanditRewardTracker) TrackImpression(
	ctx context.Context,
	params TrackImpressionParams,
) error {
	if err := b.validateArmExists(ctx, params.ExperimentID, params.ArmID); err != nil {
		return err
	}

	appender, ok := b.repo.(impressionEventAppender)
	if !ok {
		return fmt.Errorf("impression event logging not supported")
	}

	normalizedEvent := normalizeImpressionEvent(params.ExperimentID, params.ArmID, params.UserID, params.Event)
	if err := appender.AppendImpressionEvent(ctx, &normalizedEvent); err != nil {
		return fmt.Errorf("failed to append impression event: %w", err)
	}

	return nil
}

func (b *BanditRewardTracker) validateArmExists(ctx context.Context, experimentID, armID uuid.UUID) error {
	arms, err := b.repo.GetArms(ctx, experimentID)
	if err != nil {
		return err
	}
	if len(arms) == 0 {
		return ErrExperimentArmsNotFound
	}

	for _, arm := range arms {
		if arm.ID == armID {
			return nil
		}
	}
	return ErrBanditArmNotFound
}

func normalizeImpressionEvent(experimentID, armID, userID uuid.UUID, event *ImpressionEvent) ImpressionEvent {
	if event == nil {
		return ImpressionEvent{
			ExperimentID: experimentID,
			ArmID:        armID,
			UserID:       userID,
			EventType:    ImpressionEventTypeImpression,
			OccurredAt:   time.Now().UTC(),
		}
	}

	normalized := *event
	if normalized.ExperimentID == uuid.Nil {
		normalized.ExperimentID = experimentID
	}
	if normalized.ArmID == uuid.Nil {
		normalized.ArmID = armID
	}
	if normalized.UserID == uuid.Nil {
		normalized.UserID = userID
	}
	if normalized.EventType == "" {
		normalized.EventType = ImpressionEventTypeImpression
	}
	if normalized.OccurredAt.IsZero() {
		normalized.OccurredAt = time.Now().UTC()
	}
	return normalized
}

func (b *BanditRewardTracker) UpdateRewardWithEvent(
	ctx context.Context,
	params RewardWithEventParams,
) error {
	stats, err := b.repo.GetArmStats(ctx, params.ArmID)
	if err != nil {
		return fmt.Errorf("failed to get arm stats: %w", err)
	}

	applyRewardToArmStats(stats, params.Reward)

	if err := b.repo.UpdateArmStats(ctx, stats); err != nil {
		return fmt.Errorf("failed to update arm stats: %w", err)
	}

	if err := b.appendConversionEventIfSupported(ctx, params); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("ab:arm:%s", params.ArmID.String())
	if err := b.cache.SetArmStats(ctx, cacheKey, stats, 24*time.Hour); err != nil {
		b.logger.Warn("Failed to update cache", zap.Error(err))
	}

	b.logger.Debug("Reward updated",
		zap.String("arm_id", params.ArmID.String()),
		zap.Float64("reward", params.Reward),
		zap.Float64("alpha", stats.Alpha),
		zap.Float64("beta", stats.Beta),
		zap.Int("samples", stats.Samples),
	)

	return nil
}

func applyRewardToArmStats(stats *ArmStats, reward float64) {
	if reward > 0 {
		stats.Alpha += 1.0
		stats.Conversions++
		stats.Revenue += reward
	} else {
		stats.Beta += 1.0
	}
	stats.Samples++

	if stats.Samples > 0 {
		stats.AvgReward = stats.Revenue / float64(stats.Samples)
	}
}

func (b *BanditRewardTracker) appendConversionEventIfSupported(
	ctx context.Context,
	params RewardWithEventParams,
) error {
	if params.Event == nil {
		return nil
	}
	appender, ok := b.repo.(conversionEventAppender)
	if !ok {
		return nil
	}

	normalizedEvent := normalizeConversionEvent(params.ExperimentID, params.ArmID, params.Reward, params.Event)
	if err := appender.AppendConversionEvent(ctx, &normalizedEvent); err != nil {
		return fmt.Errorf("failed to append conversion event: %w", err)
	}
	return nil
}

func normalizeConversionEvent(experimentID, armID uuid.UUID, reward float64, event *ConversionEvent) ConversionEvent {
	normalized := *event
	if normalized.ExperimentID == uuid.Nil {
		normalized.ExperimentID = experimentID
	}
	if normalized.ArmID == uuid.Nil {
		normalized.ArmID = armID
	}
	if normalized.EventType == "" {
		normalized.EventType = ConversionEventTypeDirectReward
	}
	if normalized.OccurredAt.IsZero() {
		normalized.OccurredAt = time.Now().UTC()
	}
	if normalized.NormalizedRewardValue == 0 {
		normalized.NormalizedRewardValue = reward
	}
	if normalized.OriginalRewardValue == 0 {
		normalized.OriginalRewardValue = reward
	}
	if normalized.NormalizedCurrency == "" {
		normalized.NormalizedCurrency = normalized.OriginalCurrency
	}
	return normalized
}

// SampleBeta generates a random sample from Beta(α, β)
// Uses Marsaglia and Tsang's method for alpha,beta >= 1
// Falls back to simple uniform for small parameters
func (s *BetaDistributionSampler) SampleBeta(alpha, beta float64) float64 {
	// Handle edge cases
	if alpha <= 0 || beta <= 0 {
		return s.rng.Float64()
	}

	// For small parameters, use simple approximation
	if alpha < 1 && beta < 1 {
		return s.sampleBetaJohnk(alpha, beta)
	}

	if alpha < 1 {
		// For alpha < 1, beta >= 1
		return s.SampleBeta(alpha+1, beta) * math.Pow(s.rng.Float64(), 1/alpha)
	}

	if beta < 1 {
		// For beta < 1, alpha >= 1
		return s.SampleBeta(alpha, beta+1) * math.Pow(s.rng.Float64(), 1/beta)
	}

	// Try Marsaglia-Tsang method for alpha,beta >= 1
	if sample := s.sampleBetaMarsagliaTsang(alpha, beta); sample >= 0 {
		return sample
	}

	// Fallback: Cheng's method
	return s.sampleBetaCheng(alpha, beta)
}

// sampleBetaJohnk implements Johnk's method for alpha,beta < 1
func (s *BetaDistributionSampler) sampleBetaJohnk(alpha, beta float64) float64 {
	for {
		u1 := s.rng.Float64()
		u2 := s.rng.Float64()
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
func (s *BetaDistributionSampler) sampleBetaMarsagliaTsang(alpha, beta float64) float64 {
	gamma1 := s.sampleGamma(alpha)
	gamma2 := s.sampleGamma(beta)

	if gamma1+gamma2 > 0 {
		return gamma1 / (gamma1 + gamma2)
	}

	return -1 // Indicate failure
}

// sampleGamma generates a sample from Gamma(shape, 1) using the Marsaglia-Tsang method (2000)
func (s *BetaDistributionSampler) sampleGamma(shape float64) float64 {
	if shape < 1 {
		return s.sampleGamma(shape+1) * math.Pow(s.rng.Float64(), 1.0/shape)
	}
	d := shape - 1.0/3.0
	c := 1.0 / math.Sqrt(9.0*d)
	for {
		z := s.rng.NormFloat64()
		v := 1.0 + c*z
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := s.rng.Float64()
		if u < 1.0-0.0331*z*z*z*z {
			return d * v
		}
		if math.Log(u) < 0.5*z*z+d-d*v+d*math.Log(v) {
			return d * v
		}
	}
}

// sampleBetaCheng implements Cheng's method as a fallback
func (s *BetaDistributionSampler) sampleBetaCheng(alpha, beta float64) float64 {
	a := alpha - 1
	bParam := beta - 1

	// Initial theta value
	theta := 1.0
	if a <= bParam {
		theta = a / (a + bParam)
	}

	x := theta
	for {
		u := s.rng.Float64()
		v := s.rng.Float64()

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
func (b *BanditStatsCalculator) GetArmStatistics(ctx context.Context, experimentID uuid.UUID) (map[uuid.UUID]*ArmStats, error) {
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
func (b *BanditStatsCalculator) CalculateWinProbability(ctx context.Context, experimentID uuid.UUID, simulations int) (map[uuid.UUID]float64, error) {
	armStats, err := b.loadArmStatsForSimulation(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	winCounts := b.runMonteCarloWinCounts(armStats, simulations)

	winProbs := make(map[uuid.UUID]float64, len(winCounts))
	for armID, count := range winCounts {
		winProbs[armID] = float64(count) / float64(simulations)
	}

	return winProbs, nil
}

func (b *BanditStatsCalculator) loadArmStatsForSimulation(ctx context.Context, experimentID uuid.UUID) ([]*ArmStats, error) {
	arms, err := b.repo.GetArms(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	armStats := make([]*ArmStats, 0, len(arms))
	for _, arm := range arms {
		stats, err := b.repo.GetArmStats(ctx, arm.ID)
		if err != nil {
			return nil, err
		}
		stats.ArmID = arm.ID
		armStats = append(armStats, stats)
	}
	return armStats, nil
}

func (b *BanditStatsCalculator) runMonteCarloWinCounts(armStats []*ArmStats, simulations int) map[uuid.UUID]int {
	winCounts := make(map[uuid.UUID]int, len(armStats))
	for _, stats := range armStats {
		winCounts[stats.ArmID] = 0
	}

	for i := 0; i < simulations; i++ {
		var bestArmID uuid.UUID
		maxSample := -1.0

		for _, stats := range armStats {
			sample := b.sampler.SampleBeta(stats.Alpha, stats.Beta)
			if sample > maxSample {
				maxSample = sample
				bestArmID = stats.ArmID
			}
		}

		if bestArmID != uuid.Nil {
			winCounts[bestArmID]++
		}
	}
	return winCounts
}
