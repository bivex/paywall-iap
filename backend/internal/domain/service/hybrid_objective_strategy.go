package service

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// HybridObjectiveStrategy implements multi-objective optimization
// Supports combining conversion rate, LTV, and revenue into a single score
type HybridObjectiveStrategy struct {
	repo           BanditRepository
	cache          BanditCache
	logger         *zap.Logger
	config         *ExperimentConfig
	baseBandit     *ThompsonSamplingBandit
}

// ObjectiveScore represents the score for a single objective
type ObjectiveScore struct {
	ObjectiveType ObjectiveType
	Score        float64
	Alpha        float64
	Beta         float64
	Samples      int
	Conversions  int
	Revenue      float64
	AvgLTV       float64
}

// ArmObjectiveStats represents per-objective statistics for an arm
type ArmObjectiveStats struct {
	ArmID        uuid.UUID
	ObjectiveType ObjectiveType
	Alpha        float64
	Beta         float64
	Samples      int
	Conversions  int
	TotalRevenue float64
	AvgLTV       float64
}

// ObjectiveRepository defines the repository interface for objective stats
type ObjectiveRepository interface {
	GetObjectiveStats(ctx context.Context, armID uuid.UUID, objectiveType ObjectiveType) (*ArmObjectiveStats, error)
	UpdateObjectiveStats(ctx context.Context, stats *ArmObjectiveStats) error
	GetAllObjectiveStats(ctx context.Context, armID uuid.UUID) (map[ObjectiveType]*ArmObjectiveStats, error)
}

// NewHybridObjectiveStrategy creates a new hybrid objective strategy
func NewHybridObjectiveStrategy(
	repo BanditRepository,
	cache BanditCache,
	logger *zap.Logger,
	config *ExperimentConfig,
	baseBandit *ThompsonSamplingBandit,
) *HybridObjectiveStrategy {
	if config == nil {
		config = &ExperimentConfig{
			ObjectiveType: ObjectiveConversion,
		}
	}

	// Set default weights for hybrid if not provided
	if config.ObjectiveType == ObjectiveHybrid && config.ObjectiveWeights == nil {
		config.ObjectiveWeights = map[string]float64{
			"conversion": 0.5,
			"ltv":       0.3,
			"revenue":   0.2,
		}
	}

	return &HybridObjectiveStrategy{
		repo:       repo,
		cache:      cache,
		logger:     logger,
		config:     config,
		baseBandit: baseBandit,
	}
}

// CalculateScore calculates the objective score for an arm
func (s *HybridObjectiveStrategy) CalculateScore(
	ctx context.Context,
	armID uuid.UUID,
) (float64, error) {
	switch s.config.ObjectiveType {
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
func (s *HybridObjectiveStrategy) calculateConversionScore(ctx context.Context, armID uuid.UUID) (float64, error) {
	stats, err := s.repo.GetArmStats(ctx, armID)
	if err != nil {
		return 0, fmt.Errorf("failed to get arm stats: %w", err)
	}

	// Sample from Beta distribution
	return s.baseBandit.SampleBeta(stats.Alpha, stats.Beta), nil
}

// calculateLVTScore uses Expected Value = P(conversion) × AvgLTV
func (s *HybridObjectiveStrategy) calculateLVTScore(ctx context.Context, armID uuid.UUID) (float64, error) {
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
func (s *HybridObjectiveStrategy) calculateRevenueScore(ctx context.Context, armID uuid.UUID) (float64, error) {
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
func (s *HybridObjectiveStrategy) calculateHybridScore(ctx context.Context, armID uuid.UUID) (float64, error) {
	scores := make(map[string]float64)
	totalWeight := 0.0

	// Calculate score for each objective
	for objective, weight := range s.config.ObjectiveWeights {
		if weight <= 0 {
			continue
		}

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
			continue
		}

		if err != nil {
			s.logger.Warn("Failed to calculate objective score",
				zap.String("objective", objective),
				zap.Error(err),
			)
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
	for objective, weight := range s.config.ObjectiveWeights {
		if score, ok := normalizedScores[objective]; ok {
			normalizedWeight := weight / totalWeight
			hybridScore += score * normalizedWeight
		}
	}

	return hybridScore, nil
}

// normalizeScores normalizes scores to [0,1] range using min-max normalization
func (s *HybridObjectiveStrategy) normalizeScores(scores map[string]float64) map[string]float64 {
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

// RecordObjectiveReward records a reward for a specific objective
func (s *HybridObjectiveStrategy) RecordObjectiveReward(
	ctx context.Context,
	armID uuid.UUID,
	objectiveType ObjectiveType,
	reward float64,
	ltv float64,
) error {
	objRepo, ok := s.repo.(ObjectiveRepository)
	if !ok {
		return fmt.Errorf("repository does not support objective stats")
	}

	// Get existing stats
	stats, err := objRepo.GetObjectiveStats(ctx, armID, objectiveType)
	if err != nil {
		// Initialize new stats
		stats = &ArmObjectiveStats{
			ArmID:        armID,
			ObjectiveType: objectiveType,
			Alpha:        1.0,
			Beta:         1.0,
			Samples:      0,
			Conversions:  0,
			TotalRevenue: 0,
			AvgLTV:       0,
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
func (s *HybridObjectiveStrategy) GetObjectiveScores(
	ctx context.Context,
	armID uuid.UUID,
) (map[ObjectiveType]*ObjectiveScore, error) {
	scores := make(map[ObjectiveType]*ObjectiveScore)

	// Get basic stats
	stats, err := s.repo.GetArmStats(ctx, armID)
	if err != nil {
		return nil, fmt.Errorf("failed to get arm stats: %w", err)
	}

	// Conversion score
	conversionScore := s.baseBandit.SampleBeta(stats.Alpha, stats.Beta)
	scores[ObjectiveConversion] = &ObjectiveScore{
		ObjectiveType: ObjectiveConversion,
		Score:        conversionScore,
		Alpha:        stats.Alpha,
		Beta:         stats.Beta,
		Samples:      stats.Samples,
		Conversions:  stats.Conversions,
		Revenue:      stats.Revenue,
	}

	// Get objective-specific stats if available
	objRepo, ok := s.repo.(ObjectiveRepository)
	if ok {
		objStats, err := objRepo.GetAllObjectiveStats(ctx, armID)
		if err == nil {
			for objType, objStat := range objStats {
				conversionProb := objStat.Alpha / (objStat.Alpha + objStat.Beta)

				switch objType {
				case ObjectiveLTV:
					score := &ObjectiveScore{
						ObjectiveType: ObjectiveLTV,
						Score:        conversionProb * objStat.AvgLTV,
						Alpha:        objStat.Alpha,
						Beta:         objStat.Beta,
						Samples:      objStat.Samples,
						Conversions:  objStat.Conversions,
						AvgLTV:       objStat.AvgLTV,
					}
					scores[ObjectiveLTV] = score

				case ObjectiveRevenue:
					score := &ObjectiveScore{
						ObjectiveType: ObjectiveRevenue,
						Score:        conversionProb * (objStat.TotalRevenue / float64(objStat.Samples)),
						Alpha:        objStat.Alpha,
						Beta:         objStat.Beta,
						Samples:      objStat.Samples,
						Conversions:  objStat.Conversions,
						Revenue:      objStat.TotalRevenue,
					}
					scores[ObjectiveRevenue] = score
				}
			}
		}
	}

	return scores, nil
}

// UpdateConfig updates the objective configuration
func (s *HybridObjectiveStrategy) UpdateConfig(config *ExperimentConfig) {
	if config != nil {
		s.config = config
		s.logger.Info("Hybrid objective config updated",
			zap.String("objective_type", string(config.ObjectiveType)),
			zap.Any("weights", config.ObjectiveWeights),
		)
	}
}

// GetConfig returns the current configuration
func (s *HybridObjectiveStrategy) GetConfig() *ExperimentConfig {
	return s.config
}

// ValidateWeights validates that objective weights sum to a reasonable value
func (s *HybridObjectiveStrategy) ValidateWeights(weights map[string]float64) error {
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
func (s *HybridObjectiveStrategy) NormalizeWeights(weights map[string]float64) map[string]float64 {
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
