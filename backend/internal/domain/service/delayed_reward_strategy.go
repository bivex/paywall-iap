package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DelayedRewardStrategy handles delayed feedback for conversions
// that happen after the initial arm selection
type DelayedRewardStrategy struct {
	repo         BanditRepository
	cache        BanditCache
	logger       *zap.Logger
	defaultTTL   time.Duration // How long to wait for conversions
	maxTTL       time.Duration // Maximum time to track pending rewards
}

// PendingReward represents a pending conversion reward
type PendingReward struct {
	ID              uuid.UUID
	ExperimentID    uuid.UUID
	ArmID           uuid.UUID
	UserID          uuid.UUID
	AssignedAt      time.Time
	ExpiresAt       time.Time
	Converted       bool
	ConversionValue float64
	ConversionCurrency string
	ConvertedAt     *time.Time
	ProcessedAt     *time.Time
}

// ConversionLink links a pending reward to an actual transaction
type ConversionLink struct {
	PendingID     uuid.UUID
	TransactionID uuid.UUID
	LinkedAt      time.Time
}

// DelayedRewardRepository defines the repository interface for delayed rewards
type DelayedRewardRepository interface {
	CreatePendingReward(ctx context.Context, reward *PendingReward) error
	GetPendingReward(ctx context.Context, id uuid.UUID) (*PendingReward, error)
	GetPendingRewardsByUser(ctx context.Context, userID, experimentID uuid.UUID) ([]*PendingReward, error)
	GetExpiredPendingRewards(ctx context.Context, limit int) ([]*PendingReward, error)
	UpdatePendingReward(ctx context.Context, reward *PendingReward) error
	LinkConversion(ctx context.Context, link *ConversionLink) error
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*ConversionLink, error)
}

// NewDelayedRewardStrategy creates a new delayed reward strategy
func NewDelayedRewardStrategy(
	repo BanditRepository,
	cache BanditCache,
	logger *zap.Logger,
) *DelayedRewardStrategy {
	return &DelayedRewardStrategy{
		repo:       repo,
		cache:      cache,
		logger:     logger,
		defaultTTL: 7 * 24 * time.Hour,  // 7 days default
		maxTTL:     30 * 24 * time.Hour, // 30 days maximum
	}
}

// RecordPendingReward records a pending reward that will be credited upon conversion
func (s *DelayedRewardStrategy) RecordPendingReward(
	ctx context.Context,
	experimentID, armID, userID uuid.UUID,
) (*PendingReward, error) {
	expiresAt := time.Now().Add(s.defaultTTL)

	pending := &PendingReward{
		ID:           uuid.New(),
		ExperimentID: experimentID,
		ArmID:        armID,
		UserID:       userID,
		AssignedAt:   time.Now(),
		ExpiresAt:    expiresAt,
		Converted:    false,
	}

	// Get the delayed reward repository
	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support delayed rewards")
	}

	if err := delayedRepo.CreatePendingReward(ctx, pending); err != nil {
		return nil, fmt.Errorf("failed to create pending reward: %w", err)
	}

	// Cache the pending reward for quick lookup
	cacheKey := s.getPendingCacheKey(pending.ID)
	if err := s.cachePendingReward(ctx, cacheKey, pending); err != nil {
		s.logger.Warn("Failed to cache pending reward", zap.Error(err))
	}

	s.logger.Info("Pending reward recorded",
		zap.String("pending_id", pending.ID.String()),
		zap.String("experiment_id", experimentID.String()),
		zap.String("arm_id", armID.String()),
		zap.String("user_id", userID.String()),
		zap.Time("expires_at", expiresAt),
	)

	return pending, nil
}

// ProcessConversion processes a conversion and applies the reward to the pending arm
func (s *DelayedRewardStrategy) ProcessConversion(
	ctx context.Context,
	transactionID uuid.UUID,
	userID uuid.UUID,
	conversionValue float64,
	currency string,
) error {
	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return fmt.Errorf("repository does not support delayed rewards")
	}

	// Find pending rewards for this user
	// For now, we'll link to the most recent unconverted pending reward
	pendingRewards, err := delayedRepo.GetPendingRewardsByUser(ctx, userID, uuid.Nil)
	if err != nil {
		return fmt.Errorf("failed to get pending rewards: %w", err)
	}

	var matchedPending *PendingReward
	now := time.Now()

	// Find the most recent unexpired pending reward
	for _, pending := range pendingRewards {
		if !pending.Converted && pending.ExpiresAt.After(now) {
			if matchedPending == nil || pending.AssignedAt.After(matchedPending.AssignedAt) {
				matchedPending = pending
			}
		}
	}

	if matchedPending == nil {
		s.logger.Info("No matching pending reward found for conversion",
			zap.String("transaction_id", transactionID.String()),
			zap.String("user_id", userID.String()),
		)
		return nil // Not an error, just no match
	}

	// Mark as converted
	matchedPending.Converted = true
	matchedPending.ConversionValue = conversionValue
	matchedPending.ConversionCurrency = currency
	convertedAt := now
	matchedPending.ConvertedAt = &convertedAt
	processedAt := now
	matchedPending.ProcessedAt = &processedAt

	// Update the pending reward
	if err := delayedRepo.UpdatePendingReward(ctx, matchedPending); err != nil {
		return fmt.Errorf("failed to update pending reward: %w", err)
	}

	// Create the conversion link
	link := &ConversionLink{
		PendingID:     matchedPending.ID,
		TransactionID: transactionID,
		LinkedAt:      now,
	}
	if err := delayedRepo.LinkConversion(ctx, link); err != nil {
		return fmt.Errorf("failed to link conversion: %w", err)
	}

	// Invalidate cache
	cacheKey := s.getPendingCacheKey(matchedPending.ID)
	if err := s.invalidatePendingCache(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate cache", zap.Error(err))
	}

	s.logger.Info("Conversion processed and linked to pending reward",
		zap.String("pending_id", matchedPending.ID.String()),
		zap.String("transaction_id", transactionID.String()),
		zap.String("arm_id", matchedPending.ArmID.String()),
		zap.Float64("value", conversionValue),
		zap.String("currency", currency),
	)

	return nil
}

// ProcessExpiredRewards processes expired pending rewards as non-conversions
func (s *DelayedRewardStrategy) ProcessExpiredRewards(
	ctx context.Context,
	baseBandit *ThompsonSamplingBandit,
	batchSize int,
) (int, error) {
	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return 0, fmt.Errorf("repository does not support delayed rewards")
	}

	expired, err := delayedRepo.GetExpiredPendingRewards(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired rewards: %w", err)
	}

	processed := 0
	for _, pending := range expired {
		if pending.Converted {
			continue // Already processed
		}

		// Record as non-conversion (reward = 0)
		if err := baseBandit.UpdateReward(ctx, pending.ExperimentID, pending.ArmID, 0); err != nil {
			s.logger.Error("Failed to record expired reward",
				zap.String("pending_id", pending.ID.String()),
				zap.Error(err),
			)
			continue
		}

		// Mark as processed
		now := time.Now()
		pending.ProcessedAt = &now
		if err := delayedRepo.UpdatePendingReward(ctx, pending); err != nil {
			s.logger.Error("Failed to update expired pending reward",
				zap.String("pending_id", pending.ID.String()),
				zap.Error(err),
			)
			continue
		}

		// Invalidate cache
		cacheKey := s.getPendingCacheKey(pending.ID)
		if err := s.invalidatePendingCache(ctx, cacheKey); err != nil {
			s.logger.Warn("Failed to invalidate cache", zap.Error(err))
		}

		processed++
	}

	if processed > 0 {
		s.logger.Info("Processed expired pending rewards",
			zap.Int("count", processed),
		)
	}

	return processed, nil
}

// GetPendingReward retrieves a pending reward by ID
func (s *DelayedRewardStrategy) GetPendingReward(ctx context.Context, id uuid.UUID) (*PendingReward, error) {
	// Try cache first
	cacheKey := s.getPendingCacheKey(id)
	if cached, err := s.getCachedPendingReward(ctx, cacheKey); err == nil {
		return cached, nil
	}

	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support delayed rewards")
	}

	pending, err := delayedRepo.GetPendingReward(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reward: %w", err)
	}

	// Cache it
	if err := s.cachePendingReward(ctx, cacheKey, pending); err != nil {
		s.logger.Warn("Failed to cache pending reward", zap.Error(err))
	}

	return pending, nil
}

// GetPendingRewardsByUser retrieves all pending rewards for a user
func (s *DelayedRewardStrategy) GetPendingRewardsByUser(
	ctx context.Context,
	userID, experimentID uuid.UUID,
) ([]*PendingReward, error) {
	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support delayed rewards")
	}

	return delayedRepo.GetPendingRewardsByUser(ctx, userID, experimentID)
}

// GetStats returns statistics about pending rewards
func (s *DelayedRewardStrategy) GetStats(ctx context.Context) (map[string]interface{}, error) {
	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support delayed rewards")
	}

	// Get a batch of expired to count
	expired, err := delayedRepo.GetExpiredPendingRewards(ctx, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	expiredCount := 0
	for _, p := range expired {
		if !p.Converted {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"expired_unprocessed": expiredCount,
		"sample_batch_size":   len(expired),
	}, nil
}

// SetDefaultTTL sets the default time-to-live for pending rewards
func (s *DelayedRewardStrategy) SetDefaultTTL(ttl time.Duration) {
	if ttl > 0 && ttl <= s.maxTTL {
		s.defaultTTL = ttl
		s.logger.Info("Default TTL updated", zap.Duration("ttl", ttl))
	}
}

// GetDefaultTTL returns the default time-to-live for pending rewards
func (s *DelayedRewardStrategy) GetDefaultTTL() time.Duration {
	return s.defaultTTL
}

// Helper methods

func (s *DelayedRewardStrategy) getPendingCacheKey(id uuid.UUID) string {
	return fmt.Sprintf("bandit:pending:%s", id.String())
}

func (s *DelayedRewardStrategy) cachePendingReward(ctx context.Context, key string, pending *PendingReward) error {
	// Use the cache interface - would need to extend BanditCache for this
	// For now, this is a placeholder
	return nil
}

func (s *DelayedRewardStrategy) getCachedPendingReward(ctx context.Context, key string) (*PendingReward, error) {
	// Use the cache interface - would need to extend BanditCache for this
	// For now, this is a placeholder
	return nil, fmt.Errorf("not cached")
}

func (s *DelayedRewardStrategy) invalidatePendingCache(ctx context.Context, key string) error {
	// Use the cache interface - would need to extend BanditCache for this
	// For now, this is a placeholder
	return nil
}

// GetConversionLinks retrieves all pending rewards linked to a transaction
func (s *DelayedRewardStrategy) GetConversionLinks(
	ctx context.Context,
	transactionID uuid.UUID,
) ([]*ConversionLink, error) {
	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support delayed rewards")
	}

	return delayedRepo.GetByTransactionID(ctx, transactionID)
}
