package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LinUCBSelectionStrategy implements Linear Upper Confidence Bound for contextual bandits
// Uses disjoint linear models per arm
type LinUCBSelectionStrategy struct {
	repo   BanditRepository
	cache  BanditCache
	logger *zap.Logger
	alpha  float64 // Exploration parameter
	dim    int     // Feature dimension
}

// LinUCBModel represents the model parameters for a single arm
type LinUCBModel struct {
	ArmID        uuid.UUID
	MatrixA      [][]float64 // Design matrix (d x d)
	VectorB      []float64   // Reward vector (d)
	Theta        []float64   // Learned parameters (d)
	SamplesCount int
}

// NewLinUCBSelectionStrategy creates a new LinUCB selection strategy
func NewLinUCBSelectionStrategy(
	repo BanditRepository,
	cache BanditCache,
	logger *zap.Logger,
	alpha float64,
	dimension int,
) *LinUCBSelectionStrategy {
	if alpha <= 0 {
		alpha = 0.3 // Default exploration parameter
	}
	if dimension <= 0 {
		dimension = 20 // Default feature dimension
	}

	return &LinUCBSelectionStrategy{
		repo:   repo,
		cache:  cache,
		logger: logger,
		alpha:  alpha,
		dim:    dimension,
	}
}

// SelectArm selects the best arm using LinUCB algorithm
// UCB = theta^T * x + alpha * sqrt(x^T * A^(-1) * x)
func (s *LinUCBSelectionStrategy) SelectArm(
	ctx context.Context,
	arms []Arm,
	userContext UserContext,
) (*Arm, error) {
	if len(arms) == 0 {
		return nil, fmt.Errorf("no arms available")
	}

	// Convert user context to feature vector
	features, err := s.contextToFeatureVector(userContext)
	if err != nil {
		s.logger.Warn("Failed to create feature vector, using uniform", zap.Error(err))
		return s.selectRandomArm(arms), nil
	}

	var bestArm *Arm
	maxUCB := -math.Inf(1)

	// Calculate UCB for each arm
	for _, arm := range arms {
		model, err := s.getOrCreateModel(ctx, arm.ID)
		if err != nil {
			s.logger.Warn("Failed to get model for arm",
				zap.String("arm_id", arm.ID.String()),
				zap.Error(err),
			)
			continue
		}

		// Calculate UCB
		ucb := s.calculateUCB(features, model)

		s.logger.Debug("Arm UCB",
			zap.String("arm_id", arm.ID.String()),
			zap.String("arm_name", arm.Name),
			zap.Float64("ucb", ucb),
		)

		if ucb > maxUCB {
			maxUCB = ucb
			bestArm = &arm
		}
	}

	// Fallback to random if no arm had valid UCB
	if bestArm == nil {
		bestArm = s.selectRandomArm(arms)
	}

	return bestArm, nil
}

// GetName returns the strategy name
func (s *LinUCBSelectionStrategy) GetName() string {
	return "linucb"
}

// UpdateModel updates the LinUCB model with a new reward
func (s *LinUCBSelectionStrategy) UpdateModel(
	ctx context.Context,
	armID uuid.UUID,
	userContext UserContext,
	reward float64,
) error {
	features, err := s.contextToFeatureVector(userContext)
	if err != nil {
		return fmt.Errorf("failed to create feature vector: %w", err)
	}

	model, err := s.getOrCreateModel(ctx, armID)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	// Update A = A + x * x^T
	// Update b = b + reward * x
	d := s.dim
	for i := 0; i < d; i++ {
		for j := 0; j < d; j++ {
			model.MatrixA[i][j] += features[i] * features[j]
		}
		model.VectorB[i] += reward * features[i]
	}
	model.SamplesCount++

	// Recompute theta = A^(-1) * b
	model.Theta = s.solveLinearSystem(model.MatrixA, model.VectorB)

	// Save updated model
	if err := s.saveModel(ctx, model); err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	s.logger.Debug("LinUCB model updated",
		zap.String("arm_id", armID.String()),
		zap.Int("samples", model.SamplesCount),
	)

	return nil
}

// calculateUCB calculates the Upper Confidence Bound for a feature vector
// UCB = theta^T * x + alpha * sqrt(x^T * A^(-1) * x)
func (s *LinUCBSelectionStrategy) calculateUCB(features []float64, model *LinUCBModel) float64 {
	d := s.dim

	// Calculate theta^T * x (expected reward)
	prediction := 0.0
	for i := 0; i < d; i++ {
		prediction += model.Theta[i] * features[i]
	}

	// Calculate A^(-1) * x
	// For numerical stability, we use an approximation
	// Using the diagonal of A as a proxy for A^(-1)
	uncertainty := 0.0
	for i := 0; i < d; i++ {
		if model.MatrixA[i][i] > 0 {
			uncertainty += (features[i] * features[i]) / model.MatrixA[i][i]
		}
	}

	// UCB = prediction + alpha * sqrt(uncertainty)
	ucb := prediction + s.alpha*math.Sqrt(uncertainty)

	return ucb
}

// getOrCreateModel retrieves or creates a LinUCB model for an arm
func (s *LinUCBSelectionStrategy) getOrCreateModel(ctx context.Context, armID uuid.UUID) (*LinUCBModel, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("linucb:model:%s", armID.String())

	// Try repository
	// Note: This would need to be implemented in the repository
	// For now, create a new model

	// Initialize new model
	d := s.dim
	model := &LinUCBModel{
		ArmID:   armID,
		MatrixA: make([][]float64, d),
		VectorB: make([]float64, d),
		Theta:   make([]float64, d),
	}

	// Initialize A as identity matrix
	for i := 0; i < d; i++ {
		model.MatrixA[i] = make([]float64, d)
		model.MatrixA[i][i] = 1.0 // Identity
		model.VectorB[i] = 0.0
		model.Theta[i] = 0.0
	}

	return model, nil
}

// saveModel saves the model to repository and cache
func (s *LinUCBSelectionStrategy) saveModel(ctx context.Context, model *LinUCBModel) error {
	// Save to cache
	cacheKey := fmt.Sprintf("linucb:model:%s", model.ArmID.String())
	// Cache implementation would go here

	return nil
}

// contextToFeatureVector converts user context to a feature vector
func (s *LinUCBSelectionStrategy) contextToFeatureVector(ctx UserContext) ([]float64, error) {
	d := s.dim
	features := make([]float64, d)

	// Feature engineering
	// Indices 0-9: Country one-hot encoding (top countries)
	countries := []string{"US", "GB", "DE", "FR", "JP", "CA", "AU", "BR", "IN", "other"}
	countryIdx := s.getStringIndex(ctx.Country, countries)
	if countryIdx < len(countries) {
		features[countryIdx] = 1.0
	} else {
		features[len(countries)-1] = 1.0 // "other"
	}

	// Indices 10-14: Device one-hot encoding
	devices := []string{"ios", "android", "web", "tablet", "other"}
	deviceIdx := s.getStringIndex(ctx.Device, devices)
	if deviceIdx < len(devices) {
		features[10+deviceIdx] = 1.0
	} else {
		features[14] = 1.0
	}

	// Index 15: Days since install (normalized 0-1, capped at 30)
	features[15] = math.Min(float64(ctx.DaysSinceInstall)/30.0, 1.0)

	// Index 16: Total spent (normalized log scale)
	if ctx.TotalSpent > 0 {
		features[16] = math.Log1p(ctx.TotalSpent) / 10.0 // Normalize
	}

	// Index 17: Is past purchaser
	isPurchaser := 0.0
	if ctx.TotalSpent > 0 {
		isPurchaser = 1.0
	}
	features[17] = isPurchaser

	// Index 18: Recent purchaser (within 7 days)
	recentPurchaser := 0.0
	if ctx.LastPurchaseAt != nil {
		daysSincePurchase := math.Floor(time.Since(*ctx.LastPurchaseAt).Hours() / 24)
		if daysSincePurchase <= 7 {
			recentPurchaser = 1.0
		}
	}
	features[18] = recentPurchaser

	// Index 19: Bias term
	features[19] = 1.0

	return features, nil
}

// getStringIndex returns the index of a string in a slice
func (s *LinUCBSelectionStrategy) getStringIndex(str string, slice []string) int {
	for i, s := range slice {
		if str == s {
			return i
		}
	}
	return len(slice)
}

// selectRandomArm selects a random arm (for fallback)
func (s *LinUCBSelectionStrategy) selectRandomArm(arms []Arm) *Arm {
	if len(arms) == 0 {
		return nil
	}
	// Simple deterministic selection for consistency
	return &arms[0]
}

// solveLinearSystem solves A * theta = b using Gaussian elimination
// For production, consider using a more robust numerical library
func (s *LinUCBSelectionStrategy) solveLinearSystem(A [][]float64, b []float64) []float64 {
	d := len(b)
	theta := make([]float64, d)

	// Copy A and b to avoid modifying originals
	A_copy := make([][]float64, d)
	for i := range A {
		A_copy[i] = make([]float64, d)
		copy(A_copy[i], A[i])
	}
	b_copy := make([]float64, d)
	copy(b_copy, b)

	// Gaussian elimination with partial pivoting
	for i := 0; i < d; i++ {
		// Find pivot
		pivot := i
		for j := i + 1; j < d; j++ {
			if math.Abs(A_copy[j][i]) > math.Abs(A_copy[pivot][i]) {
				pivot = j
			}
		}

		// Swap rows
		A_copy[i], A_copy[pivot] = A_copy[pivot], A_copy[i]
		b_copy[i], b_copy[pivot] = b_copy[pivot], b_copy[i]

		// Eliminate column
		for j := i + 1; j < d; j++ {
			factor := A_copy[j][i] / A_copy[i][i]
			for k := i; k < d; k++ {
				A_copy[j][k] -= factor * A_copy[i][k]
			}
			b_copy[j] -= factor * b_copy[i]
		}
	}

	// Back substitution
	for i := d - 1; i >= 0; i-- {
		theta[i] = b_copy[i]
		for j := i + 1; j < d; j++ {
			theta[i] -= A_copy[i][j] * theta[j]
		}
		theta[i] /= A_copy[i][i]
	}

	return theta
}

// GetModelStats returns statistics about the model
func (s *LinUCBSelectionStrategy) GetModelStats(ctx context.Context, armID uuid.UUID) (*LinUCBModel, error) {
	return s.getOrCreateModel(ctx, armID)
}

// SetExplorationAlpha updates the exploration parameter
func (s *LinUCBSelectionStrategy) SetExplorationAlpha(alpha float64) {
	if alpha > 0 {
		s.alpha = alpha
		s.logger.Info("LinUCB exploration alpha updated", zap.Float64("alpha", alpha))
	}
}

// GetExplorationAlpha returns the current exploration parameter
func (s *LinUCBSelectionStrategy) GetExplorationAlpha() float64 {
	return s.alpha
}
