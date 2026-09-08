package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DelayedRewardCache handles caching for pending rewards
type DelayedRewardCache struct {
	cache      BanditCache
	defaultTTL time.Duration
}

// DelayedRewardPendingStore manages pending reward storage, retrieval, and TTL
type DelayedRewardPendingStore struct {
	repo       BanditRepository
	cache      *DelayedRewardCache
	logger     *zap.Logger
	defaultTTL time.Duration // How long to wait for conversions
	maxTTL     time.Duration // Maximum time to track pending rewards
}

// DelayedRewardConversionProcessorComponent processes conversions for pending rewards
type DelayedRewardConversionProcessorComponent struct {
	repo   BanditRepository
	cache  *DelayedRewardCache
	logger *zap.Logger
}

// DelayedRewardExpiryProcessorComponent handles expired pending rewards
type DelayedRewardExpiryProcessorComponent struct {
	repo   BanditRepository
	cache  *DelayedRewardCache
	logger *zap.Logger
}

// DelayedRewardStrategy handles delayed feedback for conversions
// that happen after the initial arm selection
type DelayedRewardStrategy struct {
	*DelayedRewardCache
	*DelayedRewardPendingStore
	*DelayedRewardConversionProcessorComponent
	*DelayedRewardExpiryProcessorComponent
}

// PendingReward represents a pending conversion reward
type PendingReward struct {
	ID                 uuid.UUID
	ExperimentID       uuid.UUID
	ArmID              uuid.UUID
	UserID             uuid.UUID
	AssignedAt         time.Time
	ExpiresAt          time.Time
	Converted          bool
	ConversionValue    float64
	ConversionCurrency string
	ConvertedAt        *time.Time
	ProcessedAt        *time.Time
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

// ProcessPendingConversionParams contains parameters for processing a pending conversion.
type ProcessPendingConversionParams struct {
	TransactionID   uuid.UUID
	UserID          uuid.UUID
	ConversionValue float64
	Currency        string
	ProcessedAt     time.Time
}

type DelayedConversionProcessor interface {
	ProcessPendingConversion(ctx context.Context, params ProcessPendingConversionParams) (*PendingReward, bool, error)
}

type ExpiredPendingRewardProcessor interface {
	ProcessExpiredPendingReward(ctx context.Context, pendingID uuid.UUID, processedAt time.Time) (bool, error)
}

// NewDelayedRewardStrategy creates a new delayed reward strategy
func NewDelayedRewardStrategy(
	repo BanditRepository,
	cache BanditCache,
	logger *zap.Logger,
) *DelayedRewardStrategy {
	defaultTTL := 7 * 24 * time.Hour
	maxTTL := 30 * 24 * time.Hour
	dCache := &DelayedRewardCache{cache: cache, defaultTTL: defaultTTL}
	pendingStore := &DelayedRewardPendingStore{
		repo:       repo,
		cache:      dCache,
		logger:     logger,
		defaultTTL: defaultTTL,
		maxTTL:     maxTTL,
	}
	conversionProcessor := &DelayedRewardConversionProcessorComponent{
		repo:   repo,
		cache:  dCache,
		logger: logger,
	}
	expiryProcessor := &DelayedRewardExpiryProcessorComponent{
		repo:   repo,
		cache:  dCache,
		logger: logger,
	}
	return &DelayedRewardStrategy{
		DelayedRewardCache:                        dCache,
		DelayedRewardPendingStore:                 pendingStore,
		DelayedRewardConversionProcessorComponent: conversionProcessor,
		DelayedRewardExpiryProcessorComponent:     expiryProcessor,
	}
}

// RecordPendingReward records a pending reward that will be credited upon conversion
func (s *DelayedRewardPendingStore) RecordPendingReward(
	ctx context.Context,
	experimentID, armID, userID uuid.UUID,
) (*PendingReward, error) {
	expiresAt := time.Now().Add(s.defaultTTL)

	pending := &PendingReward{
		ID:           uuid.New(),
		ExperimentID: experimentID,
		ArmID:        armID,
		UserID:       userID,
		AssignedAt:   time.Now().UTC(),
		ExpiresAt:    expiresAt,
		Converted:    false,
	}

	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support delayed rewards")
	}

	if err := delayedRepo.CreatePendingReward(ctx, pending); err != nil {
		return nil, fmt.Errorf("failed to create pending reward: %w", err)
	}

	// Cache it
	cacheKey := s.cache.getPendingCacheKey(pending.ID)
	if err := s.cache.cachePendingReward(ctx, cacheKey, pending); err != nil {
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

type applyConversionRewardParams struct {
	delayedRepo     DelayedRewardRepository
	baseBandit      *ThompsonSamplingBandit
	matchedPending  *PendingReward
	transactionID   uuid.UUID
	conversionValue float64
	currency        string
	now             time.Time
}

// DelayedConversionParams encapsulates parameters for processing delayed conversion
type DelayedConversionParams struct {
	TransactionID   uuid.UUID
	UserID          uuid.UUID
	ConversionValue float64
	Currency        string
	BaseBandit      *ThompsonSamplingBandit
}

type processConversionProcessorParams struct {
	processor       DelayedConversionProcessor
	transactionID   uuid.UUID
	userID          uuid.UUID
	conversionValue float64
	currency        string
	now             time.Time
}

type processConversionFallbackParams struct {
	delayedRepo     DelayedRewardRepository
	transactionID   uuid.UUID
	userID          uuid.UUID
	conversionValue float64
	currency        string
	baseBandit      *ThompsonSamplingBandit
	now             time.Time
}

// ProcessConversion processes a conversion and applies the reward to the pending arm
func (s *DelayedRewardConversionProcessorComponent) ProcessConversion(
	ctx context.Context,
	p DelayedConversionParams,
) error {
	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return fmt.Errorf("repository does not support delayed rewards")
	}
	now := time.Now().UTC()

	if processor, ok := s.repo.(DelayedConversionProcessor); ok {
		return s.processConversionViaProcessor(ctx, processConversionProcessorParams{
			processor:       processor,
			transactionID:   p.TransactionID,
			userID:          p.UserID,
			conversionValue: p.ConversionValue,
			currency:        p.Currency,
			now:             now,
		})
	}

	return s.processConversionFallback(ctx, processConversionFallbackParams{
		delayedRepo:     delayedRepo,
		transactionID:   p.TransactionID,
		userID:          p.UserID,
		conversionValue: p.ConversionValue,
		currency:        p.Currency,
		baseBandit:      p.BaseBandit,
		now:             now,
	})
}

func (s *DelayedRewardConversionProcessorComponent) processConversionViaProcessor(
	ctx context.Context,
	p processConversionProcessorParams,
) error {
	matchedPending, processed, err := p.processor.ProcessPendingConversion(ctx, ProcessPendingConversionParams{
		TransactionID:   p.transactionID,
		UserID:          p.userID,
		ConversionValue: p.conversionValue,
		Currency:        p.currency,
		ProcessedAt:     p.now,
	})
	if err != nil {
		return fmt.Errorf("failed to process pending conversion: %w", err)
	}
	if !processed || matchedPending == nil {
		s.logger.Info("No matching pending reward found for conversion",
			zap.String("transaction_id", p.transactionID.String()),
			zap.String("user_id", p.userID.String()),
		)
		return nil
	}

	cacheKey := s.cache.getPendingCacheKey(matchedPending.ID)
	if err := s.cache.invalidatePendingCache(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate cache", zap.Error(err))
	}

	s.logger.Info("Conversion processed and linked to pending reward",
		zap.String("pending_id", matchedPending.ID.String()),
		zap.String("transaction_id", p.transactionID.String()),
		zap.String("arm_id", matchedPending.ArmID.String()),
		zap.Float64("value", p.conversionValue),
		zap.String("currency", p.currency),
	)
	return nil
}

func (s *DelayedRewardConversionProcessorComponent) processConversionFallback(
	ctx context.Context,
	p processConversionFallbackParams,
) error {
	pendingRewards, err := p.delayedRepo.GetPendingRewardsByUser(ctx, p.userID, uuid.Nil)
	if err != nil {
		return fmt.Errorf("failed to get pending rewards: %w", err)
	}

	matchedPending := findMostRecentPendingReward(pendingRewards, p.now)
	if matchedPending == nil {
		s.logger.Info("No matching pending reward found for conversion",
			zap.String("transaction_id", p.transactionID.String()),
			zap.String("user_id", p.userID.String()),
		)
		return nil
	}

	return s.applyConversionReward(ctx, applyConversionRewardParams{
		delayedRepo:     p.delayedRepo,
		baseBandit:      p.baseBandit,
		matchedPending:  matchedPending,
		transactionID:   p.transactionID,
		conversionValue: p.conversionValue,
		currency:        p.currency,
		now:             p.now,
	})
}

func findMostRecentPendingReward(pendingRewards []*PendingReward, now time.Time) *PendingReward {
	var matchedPending *PendingReward
	for _, pending := range pendingRewards {
		if !pending.Converted && pending.ExpiresAt.After(now) {
			if matchedPending == nil || pending.AssignedAt.After(matchedPending.AssignedAt) {
				matchedPending = pending
			}
		}
	}
	return matchedPending
}

func (s *DelayedRewardConversionProcessorComponent) applyConversionReward(
	ctx context.Context,
	p applyConversionRewardParams,
) error {
	if err := p.baseBandit.UpdateRewardWithEvent(ctx, RewardWithEventParams{
		ExperimentID: p.matchedPending.ExperimentID,
		ArmID:        p.matchedPending.ArmID,
		Reward:       p.conversionValue,
		Event: &ConversionEvent{
			ExperimentID:          p.matchedPending.ExperimentID,
			ArmID:                 p.matchedPending.ArmID,
			UserID:                &p.matchedPending.UserID,
		PendingRewardID:       &p.matchedPending.ID,
		TransactionID:         &p.transactionID,
		EventType:             ConversionEventTypeDelayedConversion,
		OriginalRewardValue:   p.conversionValue,
		OriginalCurrency:      p.currency,
		NormalizedRewardValue: p.conversionValue,
		NormalizedCurrency:    p.currency,
		Metadata: map[string]interface{}{
			"source": "delayed_reward_strategy_fallback",
		},
		OccurredAt: p.now,
	}}); err != nil {
		return fmt.Errorf("failed to apply delayed reward: %w", err)
	}

	p.matchedPending.Converted = true
	p.matchedPending.ConversionValue = p.conversionValue
	p.matchedPending.ConversionCurrency = p.currency
	convertedAt := p.now
	p.matchedPending.ConvertedAt = &convertedAt
	processedAt := p.now
	p.matchedPending.ProcessedAt = &processedAt

	if err := p.delayedRepo.UpdatePendingReward(ctx, p.matchedPending); err != nil {
		return fmt.Errorf("failed to update pending reward: %w", err)
	}

	link := &ConversionLink{
		PendingID:     p.matchedPending.ID,
		TransactionID: p.transactionID,
		LinkedAt:      p.now,
	}
	if err := p.delayedRepo.LinkConversion(ctx, link); err != nil {
		return fmt.Errorf("failed to link conversion: %w", err)
	}

	cacheKey := s.cache.getPendingCacheKey(p.matchedPending.ID)
	if err := s.cache.invalidatePendingCache(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate cache", zap.Error(err))
	}

	s.logger.Info("Conversion processed and linked to pending reward",
		zap.String("pending_id", p.matchedPending.ID.String()),
		zap.String("transaction_id", p.transactionID.String()),
		zap.String("arm_id", p.matchedPending.ArmID.String()),
		zap.Float64("value", p.conversionValue),
		zap.String("currency", p.currency),
	)
	return nil
}

type processExpiredRewardParams struct {
	delayedRepo DelayedRewardRepository
	baseBandit  *ThompsonSamplingBandit
	pending     *PendingReward
	now         time.Time
}

// ProcessExpiredRewards processes expired pending rewards as non-conversions
func (s *DelayedRewardExpiryProcessorComponent) ProcessExpiredRewards(
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
			continue
		}
		if s.processSingleExpiredReward(ctx, delayedRepo, baseBandit, pending) {
			processed++
		}
	}

	if processed > 0 {
		s.logger.Info("Processed expired pending rewards",
			zap.Int("count", processed),
		)
	}

	return processed, nil
}

func (s *DelayedRewardExpiryProcessorComponent) processSingleExpiredReward(
	ctx context.Context,
	delayedRepo DelayedRewardRepository,
	baseBandit *ThompsonSamplingBandit,
	pending *PendingReward,
) bool {
	now := time.Now().UTC()
	if processor, ok := s.repo.(ExpiredPendingRewardProcessor); ok {
		applied, err := processor.ProcessExpiredPendingReward(ctx, pending.ID, now)
		if err != nil {
			s.logger.Error("Failed to record expired reward",
				zap.String("pending_id", pending.ID.String()),
				zap.Error(err),
			)
			return false
		}
		if !applied {
			return false
		}

		cacheKey := s.cache.getPendingCacheKey(pending.ID)
		if err := s.cache.invalidatePendingCache(ctx, cacheKey); err != nil {
			s.logger.Warn("Failed to invalidate cache", zap.Error(err))
		}
		return true
	}

	return s.processExpiredRewardFallback(ctx, processExpiredRewardParams{
		delayedRepo: delayedRepo,
		baseBandit:  baseBandit,
		pending:     pending,
		now:         now,
	})
}

func (s *DelayedRewardExpiryProcessorComponent) processExpiredRewardFallback(
	ctx context.Context,
	p processExpiredRewardParams,
) bool {
	if err := p.baseBandit.UpdateRewardWithEvent(ctx, RewardWithEventParams{
		ExperimentID: p.pending.ExperimentID,
		ArmID:        p.pending.ArmID,
		Reward:       0,
		Event: &ConversionEvent{
			ExperimentID:          p.pending.ExperimentID,
			ArmID:                 p.pending.ArmID,
			UserID:                &p.pending.UserID,
			PendingRewardID:       &p.pending.ID,
			EventType:             ConversionEventTypeExpiredPendingReward,
			OriginalRewardValue:   0,
			NormalizedRewardValue: 0,
			Metadata: map[string]interface{}{
				"source": "delayed_reward_strategy_fallback",
			},
			OccurredAt: p.now,
		},
	}); err != nil {
		s.logger.Error("Failed to record expired reward",
			zap.String("pending_id", p.pending.ID.String()),
			zap.Error(err),
		)
		return false
	}

	p.pending.ProcessedAt = &p.now
	if err := p.delayedRepo.UpdatePendingReward(ctx, p.pending); err != nil {
		s.logger.Error("Failed to update expired pending reward",
			zap.String("pending_id", p.pending.ID.String()),
			zap.Error(err),
		)
		return false
	}

	cacheKey := s.cache.getPendingCacheKey(p.pending.ID)
	if err := s.cache.invalidatePendingCache(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate cache", zap.Error(err))
	}
	return true
}

// GetPendingReward retrieves a pending reward by ID
func (s *DelayedRewardPendingStore) GetPendingReward(ctx context.Context, id uuid.UUID) (*PendingReward, error) {
	// Try cache first
	cacheKey := s.cache.getPendingCacheKey(id)
	if cached, err := s.cache.getCachedPendingReward(ctx, cacheKey); err == nil {
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
	if err := s.cache.cachePendingReward(ctx, cacheKey, pending); err != nil {
		s.logger.Warn("Failed to cache pending reward", zap.Error(err))
	}

	return pending, nil
}

// GetPendingRewardsByUser retrieves all pending rewards for a user
func (s *DelayedRewardPendingStore) GetPendingRewardsByUser(
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
func (s *DelayedRewardPendingStore) GetStats(ctx context.Context) (map[string]interface{}, error) {
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
func (s *DelayedRewardPendingStore) SetDefaultTTL(ttl time.Duration) {
	if ttl > 0 && ttl <= s.maxTTL {
		s.defaultTTL = ttl
		s.cache.defaultTTL = ttl
		s.logger.Info("Default TTL updated", zap.Duration("ttl", ttl))
	}
}

// GetDefaultTTL returns the default time-to-live for pending rewards
func (s *DelayedRewardPendingStore) GetDefaultTTL() time.Duration {
	return s.defaultTTL
}

// Helper methods on DelayedRewardCache

func (c *DelayedRewardCache) getPendingCacheKey(id uuid.UUID) string {
	return fmt.Sprintf("bandit:pending:%s", id.String())
}

func (c *DelayedRewardCache) cachePendingReward(ctx context.Context, key string, pending *PendingReward) error {
	data, err := json.Marshal(pending)
	if err != nil {
		return fmt.Errorf("failed to marshal pending reward: %w", err)
	}
	return c.cache.SetBytes(ctx, key, data, c.defaultTTL)
}

func (c *DelayedRewardCache) getCachedPendingReward(ctx context.Context, key string) (*PendingReward, error) {
	data, err := c.cache.GetBytes(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("not cached")
	}
	var pending PendingReward
	if err := json.Unmarshal(data, &pending); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pending reward: %w", err)
	}
	return &pending, nil
}

func (c *DelayedRewardCache) invalidatePendingCache(ctx context.Context, key string) error {
	return c.cache.DeleteKey(ctx, key)
}

// GetConversionLinks retrieves all pending rewards linked to a transaction
func (s *DelayedRewardPendingStore) GetConversionLinks(
	ctx context.Context,
	transactionID uuid.UUID,
) ([]*ConversionLink, error) {
	delayedRepo, ok := s.repo.(DelayedRewardRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support delayed rewards")
	}

	return delayedRepo.GetByTransactionID(ctx, transactionID)
}
