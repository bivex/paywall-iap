package service

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// HybridObjectiveConfigManager manages objective configuration, weights, and rewards
type HybridObjectiveConfigManager struct {
	repo       BanditRepository
	cache      BanditCache
	logger     *zap.Logger
	config     *ExperimentConfig
	baseBandit *ThompsonSamplingBandit
}

// HybridObjectiveCalculator computes objective and hybrid scores for arms
type HybridObjectiveCalculator struct {
	repo          BanditRepository
	cache         BanditCache
	logger        *zap.Logger
	configManager *HybridObjectiveConfigManager
	baseBandit    *ThompsonSamplingBandit
}

// HybridObjectiveReporter builds objective scores and reports
type HybridObjectiveReporter struct {
	repo          BanditRepository
	cache         BanditCache
	logger        *zap.Logger
	configManager *HybridObjectiveConfigManager
	calculator    *HybridObjectiveCalculator
	baseBandit    *ThompsonSamplingBandit
}

// HybridObjectiveStrategy implements multi-objective optimization
// Supports combining conversion rate, LTV, and revenue into a single score
type HybridObjectiveStrategy struct {
	*HybridObjectiveConfigManager
	*HybridObjectiveCalculator
	*HybridObjectiveReporter
}

// ObjectiveScore represents the score for a single objective
type ObjectiveScore struct {
	ObjectiveType ObjectiveType
	Score         float64
	Alpha         float64
	Beta          float64
	Samples       int
	Conversions   int
	Revenue       float64
	AvgLTV        float64
}

// ArmObjectiveStats represents per-objective statistics for an arm
type ArmObjectiveStats struct {
	ArmID         uuid.UUID
	ObjectiveType ObjectiveType
	Alpha         float64
	Beta          float64
	Samples       int
	Conversions   int
	TotalRevenue  float64
	AvgLTV        float64
}

// ObjectiveRepository defines the repository interface for objective stats
type ObjectiveRepository interface {
	GetObjectiveStats(ctx context.Context, armID uuid.UUID, objectiveType ObjectiveType) (*ArmObjectiveStats, error)
	UpdateObjectiveStats(ctx context.Context, stats *ArmObjectiveStats) error
	GetAllObjectiveStats(ctx context.Context, armID uuid.UUID) (map[ObjectiveType]*ArmObjectiveStats, error)
}

// HybridObjectiveStrategyConfig specifies parameters for creating HybridObjectiveStrategy
type HybridObjectiveStrategyConfig struct {
	Repo       BanditRepository
	Cache      BanditCache
	Logger     *zap.Logger
	Config     *ExperimentConfig
	BaseBandit *ThompsonSamplingBandit
}

// NewHybridObjectiveStrategy creates a new hybrid objective strategy
func NewHybridObjectiveStrategy(cfg HybridObjectiveStrategyConfig) *HybridObjectiveStrategy {
	repo := cfg.Repo
	cache := cfg.Cache
	logger := cfg.Logger
	config := cfg.Config
	baseBandit := cfg.BaseBandit

	if config == nil {
		config = &ExperimentConfig{
			ObjectiveType: ObjectiveConversion,
		}
	}

	// Set default weights for hybrid if not provided
	if config.ObjectiveType == ObjectiveHybrid && config.ObjectiveWeights == nil {
		config.ObjectiveWeights = map[string]float64{
			"conversion": 0.5,
			"ltv":        0.3,
			"revenue":    0.2,
		}
	}

	configManager := &HybridObjectiveConfigManager{
		repo:       repo,
		cache:      cache,
		logger:     logger,
		config:     config,
		baseBandit: baseBandit,
	}
	calculator := &HybridObjectiveCalculator{
		repo:          repo,
		cache:         cache,
		logger:        logger,
		configManager: configManager,
		baseBandit:    baseBandit,
	}
	reporter := &HybridObjectiveReporter{
		repo:          repo,
		cache:         cache,
		logger:        logger,
		configManager: configManager,
		calculator:    calculator,
		baseBandit:    baseBandit,
	}

	return &HybridObjectiveStrategy{
		HybridObjectiveConfigManager: configManager,
		HybridObjectiveCalculator:    calculator,
		HybridObjectiveReporter:      reporter,
	}
}

// CalculateScore calculates the objective score for an arm
func (s *HybridObjectiveCalculator) CalculateScore(
	ctx context.Context,
	armID uuid.UUID,
) (float64, error) {
	switch s.configManager.config.ObjectiveType {
	case ObjectiveConversion:
		return s.calculateConversionScore(ctx, armID)
	case ObjectiveLTV:
		return s.calculateLVTScore(ctx, armID)
	case ObjectiveRevenue:
		return s.calculateRevenueScore(ctx, armID)
	case ObjectiveHybrid:
		return s.calculateHybridScore(ctx, armID)
	default:
		return s.calculateConversionScore(ctx, armID)
	}
}

// calculateConversionScore uses standard Thompson Sampling
func (s *HybridObjectiveCalculator) calculateConversionScore(ctx context.Context, armID uuid.UUID) (float64, error) {
	stats, err := s.repo.GetArmStats(ctx, armID)
	if err != nil {
		return 0, fmt.Errorf("failed to get arm stats: %w", err)
	}

	// Sample from Beta distribution
	return s.baseBandit.SampleBeta(stats.Alpha, stats.Beta), nil
}

// calculateLVTScore uses Expected Value = P(conversion) × AvgLTV
func (s *HybridObjectiveCalculator) calculateLVTScore(ctx context.Context, armID uuid.UUID) (float64, error) {
	stats, err := s.repo.GetArmStats(ctx, armID)
	if err != nil {
		return 0, fmt.Errorf("failed to get arm stats: %w", err)
	}

	// Get objective-specific stats
	objRepo, ok := s.repo.(ObjectiveRepository)
	if !ok {
		// Fall back to basic stats
		if stats.Samples > 0 {
			// Estimate using average revenue as proxy for LTV
			return stats.AvgReward * (stats.Alpha / (stats.Alpha + stats.Beta)), nil
		}
		return 0, nil
	}

	objStats, err := objRepo.GetObjectiveStats(ctx, armID, ObjectiveLTV)
	if err != nil {
		// Fall back to basic conversion probability
		return s.baseBandit.SampleBeta(stats.Alpha, stats.Beta), nil
	}

	// P(conversion)
	conversionProb := stats.Alpha / (stats.Alpha + stats.Beta)

	// Expected LTV = P(conversion) × AvgLTV
	expectedLTV := conversionProb * objStats.AvgLTV

	return expectedLTV, nil
}

// calculateRevenueScore uses Normalized Revenue = P(conv) × (Revenue / Price)
func (s *HybridObjectiveCalculator) calculateRevenueScore(ctx context.Context, armID uuid.UUID) (float64, error) {
	stats, err := s.repo.GetArmStats(ctx, armID)
	if err != nil {
		return 0, fmt.Errorf("failed to get arm stats: %w", err)
	}

	// P(conversion)
	conversionProb := stats.Alpha / (stats.Alpha + stats.Beta)

	// Average revenue per sample
	avgRevenue := stats.AvgReward

	// Score = conversion probability × average revenue
	return conversionProb * avgRevenue, nil
}

// calculateHybridScore combines multiple objectives with weights
func (s *HybridObjectiveCalculator) calculateHybridScore(ctx context.Context, armID uuid.UUID) (float64, error) {
	scores := make(map[string]float64)
	totalWeight := 0.0

	// Calculate score for each objective
	for objective, weight := range s.configManager.config.ObjectiveWeights {
		if weight <= 0 {
			continue
		}

		score, ok := s.evaluateObjectiveScore(ctx, armID, objective)
		if !ok {
			continue
		}

		scores[objective] = score
		totalWeight += weight
	}

	// Normalize and combine scores
	if totalWeight == 0 {
		// Fall back to conversion
		return s.calculateConversionScore(ctx, armID)
	}

	// Normalize scores to [0,1] range before combining
	normalizedScores := s.normalizeScores(scores)

	// Weighted sum
	hybridScore := 0.0
	for objective, weight := range s.configManager.config.ObjectiveWeights {
		if score, ok := normalizedScores[objective]; ok {
			normalizedWeight := weight / totalWeight
			hybridScore += score * normalizedWeight
		}
	}

	return hybridScore, nil
}

func (s *HybridObjectiveCalculator) evaluateObjectiveScore(ctx context.Context, armID uuid.UUID, objective string) (float64, bool) {
	var score float64
	var err error

	switch ObjectiveType(objective) {
	case ObjectiveConversion:
		score, err = s.calculateConversionScore(ctx, armID)
	case ObjectiveLTV:
		score, err = s.calculateLVTScore(ctx, armID)
	case ObjectiveRevenue:
		score, err = s.calculateRevenueScore(ctx, armID)
	default:
		s.logger.Warn("Unknown objective type", zap.String("objective", objective))
		return 0, false
	}

	if err != nil {
		s.logger.Warn("Failed to calculate objective score",
			zap.String("objective", objective),
			zap.Error(err),
		)
		return 0, false
	}

	return score, true
}

// normalizeScores normalizes scores to [0,1] range using min-max normalization
func (s *HybridObjectiveCalculator) normalizeScores(scores map[string]float64) map[string]float64 {
	if len(scores) == 0 {
		return scores
	}

	// Find min and max
	minScore := math.Inf(1)
	maxScore := math.Inf(-1)

	for _, score := range scores {
		if score < minScore {
			minScore = score
		}
		if score > maxScore {
			maxScore = score
		}
	}

	// If all scores are the same, return as-is
	if minScore == maxScore {
		return scores
	}

	// Normalize
	normalized := make(map[string]float64)
	for objective, score := range scores {
		normalized[objective] = (score - minScore) / (maxScore - minScore)
	}

	return normalized
}

// ObjectiveRewardParams specifies parameters for recording an objective reward
type ObjectiveRewardParams struct {
	ArmID         uuid.UUID
	ObjectiveType ObjectiveType
	Reward        float64
	LTV           float64
}

// RecordObjectiveReward records a reward for a specific objective
func (s *HybridObjectiveConfigManager) RecordObjectiveReward(
	ctx context.Context,
	p ObjectiveRewardParams,
) error {
	armID := p.ArmID
	objectiveType := p.ObjectiveType
	reward := p.Reward
	ltv := p.LTV

	objRepo, ok := s.repo.(ObjectiveRepository)
	if !ok {
		return fmt.Errorf("repository does not support objective stats")
	}

	// Get existing stats
	stats, err := objRepo.GetObjectiveStats(ctx, armID, objectiveType)
	if err != nil {
		// Initialize new stats
		stats = &ArmObjectiveStats{
			ArmID:         armID,
			ObjectiveType: objectiveType,
			Alpha:         1.0,
			Beta:          1.0,
			Samples:       0,
			Conversions:   0,
			TotalRevenue:  0,
			AvgLTV:        0,
		}
	}

	// Update stats
	stats.Samples++
	if reward > 0 {
		stats.Alpha += 1.0
		stats.Conversions++
		stats.TotalRevenue += reward
	} else {
		stats.Beta += 1.0
	}

	// Update average LTV
	if ltv > 0 {
		// Exponential moving average for LTV
		if stats.AvgLTV == 0 {
			stats.AvgLTV = ltv
		} else {
			stats.AvgLTV = 0.9*stats.AvgLTV + 0.1*ltv
		}
	}

	// Save updated stats
	if err := objRepo.UpdateObjectiveStats(ctx, stats); err != nil {
		return fmt.Errorf("failed to update objective stats: %w", err)
	}

	return nil
}

// GetObjectiveScores returns all objective scores for an arm
func (s *HybridObjectiveReporter) GetObjectiveScores(
	ctx context.Context,
	armID uuid.UUID,
) (map[ObjectiveType]*ObjectiveScore, error) {
	scores := make(map[ObjectiveType]*ObjectiveScore)

	stats, err := s.repo.GetArmStats(ctx, armID)
	if err != nil {
		return nil, fmt.Errorf("failed to get arm stats: %w", err)
	}

	conversionScore := s.baseBandit.SampleBeta(stats.Alpha, stats.Beta)
	scores[ObjectiveConversion] = &ObjectiveScore{
		ObjectiveType: ObjectiveConversion,
		Score:         conversionScore,
		Alpha:         stats.Alpha,
		Beta:          stats.Beta,
		Samples:       stats.Samples,
		Conversions:   stats.Conversions,
		Revenue:       stats.Revenue,
	}

	s.populateLTVScore(ctx, armID, stats, scores)
	s.populateRevenueScore(ctx, armID, stats, scores)
	s.populateHybridScore(ctx, armID, stats, scores)

	return scores, nil
}

func (s *HybridObjectiveReporter) shouldIncludeObjective(objective ObjectiveType) bool {
	if s.configManager.config == nil {
		return objective == ObjectiveConversion
	}

	if s.configManager.config.ObjectiveType == objective || s.configManager.config.ObjectiveType == ObjectiveHybrid {
		return true
	}

	if s.configManager.config.ObjectiveType == ObjectiveHybrid && s.configManager.config.ObjectiveWeights != nil {
		_, ok := s.configManager.config.ObjectiveWeights[string(objective)]
		return ok
	}

	return false
}

func (s *HybridObjectiveReporter) resolveObjectiveStats(ctx context.Context, armID uuid.UUID, objective ObjectiveType, stats *ArmStats) *ArmObjectiveStats {
	objRepo, ok := s.repo.(ObjectiveRepository)
	if !ok {
		return s.fallbackObjectiveStats(armID, objective, stats)
	}

	objStats, err := objRepo.GetObjectiveStats(ctx, armID, objective)
	if err != nil || objStats == nil {
		return s.fallbackObjectiveStats(armID, objective, stats)
	}

	return objStats
}

func (s *HybridObjectiveReporter) fallbackObjectiveStats(armID uuid.UUID, objective ObjectiveType, stats *ArmStats) *ArmObjectiveStats {
	return &ArmObjectiveStats{
		ArmID:         armID,
		ObjectiveType: objective,
		Alpha:         stats.Alpha,
		Beta:          stats.Beta,
		Samples:       stats.Samples,
		Conversions:   stats.Conversions,
		TotalRevenue:  stats.Revenue,
		AvgLTV:        stats.AvgReward,
	}
}

func (s *HybridObjectiveReporter) populateLTVScore(ctx context.Context, armID uuid.UUID, stats *ArmStats, scores map[ObjectiveType]*ObjectiveScore) {
	if !s.shouldIncludeObjective(ObjectiveLTV) {
		return
	}
	ltvScore, err := s.calculator.calculateLVTScore(ctx, armID)
	if err != nil {
		return
	}
	objStat := s.resolveObjectiveStats(ctx, armID, ObjectiveLTV, stats)
	scores[ObjectiveLTV] = &ObjectiveScore{
		ObjectiveType: ObjectiveLTV,
		Score:         ltvScore,
		Alpha:         objStat.Alpha,
		Beta:          objStat.Beta,
		Samples:       objStat.Samples,
		Conversions:   objStat.Conversions,
		AvgLTV:        objStat.AvgLTV,
	}
}

func (s *HybridObjectiveReporter) populateRevenueScore(ctx context.Context, armID uuid.UUID, stats *ArmStats, scores map[ObjectiveType]*ObjectiveScore) {
	if !s.shouldIncludeObjective(ObjectiveRevenue) {
		return
	}
	revenueScore, err := s.calculator.calculateRevenueScore(ctx, armID)
	if err != nil {
		return
	}
	objStat := s.resolveObjectiveStats(ctx, armID, ObjectiveRevenue, stats)
	scores[ObjectiveRevenue] = &ObjectiveScore{
		ObjectiveType: ObjectiveRevenue,
		Score:         revenueScore,
		Alpha:         objStat.Alpha,
		Beta:          objStat.Beta,
		Samples:       objStat.Samples,
		Conversions:   objStat.Conversions,
		Revenue:       objStat.TotalRevenue,
	}
}

func (s *HybridObjectiveReporter) populateHybridScore(ctx context.Context, armID uuid.UUID, stats *ArmStats, scores map[ObjectiveType]*ObjectiveScore) {
	if s.configManager.config == nil || s.configManager.config.ObjectiveType != ObjectiveHybrid {
		return
	}
	hybridScore, err := s.calculator.calculateHybridScore(ctx, armID)
	if err != nil {
		return
	}
	scores[ObjectiveHybrid] = &ObjectiveScore{
		ObjectiveType: ObjectiveHybrid,
		Score:         hybridScore,
		Alpha:         stats.Alpha,
		Beta:          stats.Beta,
		Samples:       stats.Samples,
		Conversions:   stats.Conversions,
		Revenue:       stats.Revenue,
	}
}

// UpdateConfig updates the objective configuration
func (s *HybridObjectiveConfigManager) UpdateConfig(config *ExperimentConfig) {
	if config != nil {
		s.config = config
		s.logger.Info("Hybrid objective config updated",
			zap.String("objective_type", string(config.ObjectiveType)),
			zap.Any("weights", config.ObjectiveWeights),
		)
	}
}

// GetConfig returns the current configuration
func (s *HybridObjectiveConfigManager) GetConfig() *ExperimentConfig {
	return s.config
}

// ValidateWeights validates that objective weights sum to a reasonable value
func (s *HybridObjectiveConfigManager) ValidateWeights(weights map[string]float64) error {
	if len(weights) == 0 {
		return fmt.Errorf("no weights provided")
	}

	sum := 0.0
	for _, weight := range weights {
		if weight < 0 {
			return fmt.Errorf("weights must be non-negative")
		}
		sum += weight
	}

	if sum == 0 {
		return fmt.Errorf("weights must sum to a positive value")
	}

	// Warn if weights don't sum to 1, but don't error
	if math.Abs(sum-1.0) > 0.01 {
		s.logger.Warn("Objective weights don't sum to 1.0, will be normalized",
			zap.Float64("sum", sum),
		)
	}

	return nil
}

// NormalizeWeights normalizes weights to sum to 1.0
func (s *HybridObjectiveConfigManager) NormalizeWeights(weights map[string]float64) map[string]float64 {
	sum := 0.0
	for _, weight := range weights {
		sum += weight
	}

	if sum == 0 {
		return weights
	}

	normalized := make(map[string]float64)
	for key, weight := range weights {
		normalized[key] = weight / sum
	}

	return normalized
}
