package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// SlidingWindowStrategy implements sliding window arm statistics
// Uses Redis Sorted Sets for O(log N) operations
type SlidingWindowStrategy struct {
	repo         BanditRepository
	redisClient  *redis.Client
	logger       *zap.Logger
	experimentID uuid.UUID
	config       *WindowConfig
}

// WindowStats represents aggregated statistics within a window
type WindowStats struct {
	ArmID       uuid.UUID
	Samples     int
	Conversions int
	Revenue     float64
	Alpha       float64
	Beta        float64
	AvgReward   float64
	WindowStart time.Time
	WindowEnd   time.Time
}

// NewSlidingWindowStrategy creates a new sliding window strategy
func NewSlidingWindowStrategy(
	repo BanditRepository,
	redisClient *redis.Client,
	logger *zap.Logger,
	experimentID uuid.UUID,
	config *WindowConfig,
) *SlidingWindowStrategy {
	if config == nil {
		config = &WindowConfig{
			Type:      WindowTypeEvents,
			Size:      1000,
			MinSamples: 100,
		}
	}

	return &SlidingWindowStrategy{
		repo:         repo,
		redisClient:  redisClient,
		logger:       logger,
		experimentID: experimentID,
		config:       config,
	}
}

// GetArmStats retrieves arm statistics for the current window
func (s *SlidingWindowStrategy) GetArmStats(ctx context.Context, armID uuid.UUID) (*ArmStats, error) {
	windowKey := s.getWindowKey(armID)
	statsKey := s.getStatsKey(armID)

	// Try to get cached stats first
	cachedStats, err := s.redisClient.Get(ctx, statsKey).Result()
	if err == nil {
		// Parse cached stats - for production, use proper serialization
		return s.parseCachedStats(cachedStats, armID)
	}

	if err != redis.Nil {
		s.logger.Warn("Redis error fetching stats", zap.Error(err))
	}

	// Calculate from window events
	stats, err := s.calculateWindowStats(ctx, armID)
	if err != nil {
		// Fallback to full history from repository
		return s.repo.GetArmStats(ctx, armID)
	}

	// Cache the stats
	if err := s.cacheStats(ctx, armID, stats); err != nil {
		s.logger.Warn("Failed to cache stats", zap.Error(err))
	}

	return stats, nil
}

// RecordEvent records a reward event in the sliding window
func (s *SlidingWindowStrategy) RecordEvent(ctx context.Context, armID uuid.UUID, event RewardEvent) error {
	windowKey := s.getWindowKey(armID)

	// Add event to sorted set (score = timestamp)
	score := float64(event.Timestamp.UnixMilli())
	member := fmt.Sprintf("%s:%f:%s", event.UserID.String(), event.RewardValue, event.Currency)

	pipe := s.redisClient.Pipeline()

	// Add to window
	pipe.ZAdd(ctx, windowKey, redis.Z{
		Score:  score,
		Member: member,
	})

	// Clean up old events based on window type
	switch s.config.Type {
	case WindowTypeEvents:
		// Keep only the most recent N events
		pipe.ZRemRangeByRank(ctx, windowKey, 0, -int64(s.config.Size)-1)
	case WindowTypeTime:
		// Remove events older than window size (seconds)
		cutoff := time.Now().Add(-time.Duration(s.config.Size) * time.Second)
		pipe.ZRemRangeByScore(ctx, windowKey, "0", fmt.Sprintf("%d", cutoff.UnixMilli()))
	}

	// Invalidate cached stats
	statsKey := s.getStatsKey(armID)
	pipe.Del(ctx, statsKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to record event: %w", err)
	}

	s.logger.Debug("Event recorded in sliding window",
		zap.String("arm_id", armID.String()),
		zap.String("event_type", "reward"),
		zap.Float64("reward", event.RewardValue),
	)

	return nil
}

// GetType returns the strategy type
func (s *SlidingWindowStrategy) GetType() string {
	return "sliding_window"
}

// calculateWindowStats calculates statistics from the current window
func (s *SlidingWindowStrategy) calculateWindowStats(ctx context.Context, armID uuid.UUID) (*ArmStats, error) {
	windowKey := s.getWindowKey(armID)

	// Get all events in the window
	events, err := s.redisClient.ZRevRangeWithScores(ctx, windowKey, 0, -1).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get window events: %w", err)
	}

	// Parse events and calculate stats
	samples := 0
	conversions := 0
	revenue := 0.0

	for _, z := range events {
		event, err := s.parseEventMember(z.Member.(string))
		if err != nil {
			s.logger.Warn("Failed to parse event", zap.Error(err))
			continue
		}

		samples++
		if event.RewardValue > 0 {
			conversions++
			revenue += event.RewardValue
		}
	}

	// Calculate Beta distribution parameters
	alpha := 1.0 + float64(conversions)
	beta := 1.0 + float64(samples-conversions)
	avgReward := 0.0
	if samples > 0 {
		avgReward = revenue / float64(samples)
	}

	return &ArmStats{
		ArmID:       armID,
		Alpha:       alpha,
		Beta:        beta,
		Samples:     samples,
		Conversions: conversions,
		Revenue:     revenue,
		AvgReward:   avgReward,
		UpdatedAt:   time.Now(),
	}, nil
}

// getWindowKey returns the Redis key for the window sorted set
func (s *SlidingWindowStrategy) getWindowKey(armID uuid.UUID) string {
	return fmt.Sprintf("bandit:window:%s:%s", s.experimentID.String(), armID.String())
}

// getStatsKey returns the Redis key for cached stats
func (s *SlidingWindowStrategy) getStatsKey(armID uuid.UUID) string {
	return fmt.Sprintf("bandit:window:stats:%s:%s", s.experimentID.String(), armID.String())
}

// parseEventMember parses an event member string
func (s *SlidingWindowStrategy) parseEventMember(member string) (RewardEvent, error) {
	// Format: userID:rewardValue:currency
	var userID uuid.UUID
	var rewardValue float64
	var currency string

	_, err := fmt.Sscanf(member, "%s:%f:%s", &userID, &rewardValue, &currency)
	if err != nil {
		return RewardEvent{}, fmt.Errorf("failed to parse event: %w", err)
	}

	return RewardEvent{
		UserID:      userID,
		RewardValue: rewardValue,
		Currency:    currency,
		Timestamp:   time.Now(),
	}, nil
}

// cacheStats caches the calculated stats
func (s *SlidingWindowStrategy) cacheStats(ctx context.Context, armID uuid.UUID, stats *ArmStats) error {
	statsKey := s.getStatsKey(armID)

	// Serialize stats - for production, use JSON or msgpack
	serialized := fmt.Sprintf("%.2f,%.2f,%d,%d,%.2f",
		stats.Alpha, stats.Beta, stats.Samples, stats.Conversions, stats.Revenue)

	return s.redisClient.Set(ctx, statsKey, serialized, 5*time.Minute).Err()
}

// parseCachedStats parses cached stats from Redis
func (s *SlidingWindowStrategy) parseCachedStats(serialized string, armID uuid.UUID) (*ArmStats, error) {
	var alpha, beta, revenue float64
	var samples, conversions int

	_, err := fmt.Sscanf(serialized, "%f,%f,%d,%d,%f",
		&alpha, &beta, &samples, &conversions, &revenue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cached stats: %w", err)
	}

	avgReward := 0.0
	if samples > 0 {
		avgReward = revenue / float64(samples)
	}

	return &ArmStats{
		ArmID:       armID,
		Alpha:       alpha,
		Beta:        beta,
		Samples:     samples,
		Conversions: conversions,
		Revenue:     revenue,
		AvgReward:   avgReward,
		UpdatedAt:   time.Now(),
	}, nil
}

// GetWindowInfo returns information about the current window
func (s *SlidingWindowStrategy) GetWindowInfo(ctx context.Context, armID uuid.UUID) (*WindowStats, error) {
	windowKey := s.getWindowKey(armID)

	// Get window size
	size, err := s.redisClient.ZCard(ctx, windowKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get window size: %w", err)
	}

	// Get time range
	oldest, err := s.redisClient.ZRange(ctx, windowKey, 0, 0).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get oldest event: %w", err)
	}

	newest, err := s.redisClient.ZRevRange(ctx, windowKey, 0, 0).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get newest event: %w", err)
	}

	stats, err := s.calculateWindowStats(ctx, armID)
	if err != nil {
		return nil, err
	}

	windowStats := &WindowStats{
		ArmID:       armID,
		Samples:     stats.Samples,
		Conversions: stats.Conversions,
		Revenue:     stats.Revenue,
		Alpha:       stats.Alpha,
		Beta:        stats.Beta,
		AvgReward:   stats.AvgReward,
	}

	if len(oldest) > 0 {
		// Parse timestamp from oldest event
		if oldestEvent, err := s.redisClient.ZScore(ctx, windowKey, oldest[0]).Result(); err == nil {
			windowStats.WindowStart = time.UnixMilli(int64(oldestEvent))
		}
	}

	if len(newest) > 0 {
		// Parse timestamp from newest event
		if newestEvent, err := s.redisClient.ZScore(ctx, windowKey, newest[0]).Result(); err == nil {
			windowStats.WindowEnd = time.UnixMilli(int64(newestEvent))
		}
	}

	return windowStats, nil
}

// TrimWindow trims the window to the configured size
func (s *SlidingWindowStrategy) TrimWindow(ctx context.Context, armID uuid.UUID) error {
	windowKey := s.getWindowKey(armID)

	switch s.config.Type {
	case WindowTypeEvents:
		// Keep only the most recent N events
		return s.redisClient.ZRemRangeByRank(ctx, windowKey, 0, -int64(s.config.Size)-1).Err()
	case WindowTypeTime:
		// Remove events older than window size (seconds)
		cutoff := time.Now().Add(-time.Duration(s.config.Size) * time.Second)
		return s.redisClient.ZRemRangeByScore(ctx, windowKey, "0", fmt.Sprintf("%d", cutoff.UnixMilli())).Err()
	default:
		return nil
	}
}

// ClearWindow clears all events for an arm
func (s *SlidingWindowStrategy) ClearWindow(ctx context.Context, armID uuid.UUID) error {
	windowKey := s.getWindowKey(armID)
	statsKey := s.getStatsKey(armID)

	pipe := s.redisClient.Pipeline()
	pipe.Del(ctx, windowKey)
	pipe.Del(ctx, statsKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to clear window: %w", err)
	}

	s.logger.Info("Window cleared", zap.String("arm_id", armID.String()))
	return nil
}

// HasEnoughSamples checks if the window has enough samples for reliable statistics
func (s *SlidingWindowStrategy) HasEnoughSamples(ctx context.Context, armID uuid.UUID) bool {
	stats, err := s.GetArmStats(ctx, armID)
	if err != nil {
		return false
	}

	// Check both minimum samples and minimum conversions
	minSamples := s.config.MinSamples
	if minSamples == 0 {
		minSamples = 100
	}

	return stats.Samples >= minSamples
}

// GetUtilization returns the window utilization (current size / max size)
func (s *SlidingWindowStrategy) GetUtilization(ctx context.Context, armID uuid.UUID) (float64, error) {
	if s.config.Type != WindowTypeEvents {
		return 1.0, nil // Not applicable for time-based windows
	}

	windowKey := s.getWindowKey(armID)
	size, err := s.redisClient.ZCard(ctx, windowKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get window size: %w", err)
	}

	if int(s.config.Size) == 0 {
		return 0, nil
	}

	utilization := float64(size) / float64(s.config.Size)
	return math.Min(utilization, 1.0), nil
}

// ExportEvents exports all events in the window for analysis
func (s *SlidingWindowStrategy) ExportEvents(ctx context.Context, armID uuid.UUID, limit int64) ([]RewardEvent, error) {
	windowKey := s.getWindowKey(armID)

	// Get events from newest to oldest
	events, err := s.redisClient.ZRevRangeWithScores(ctx, windowKey, 0, limit-1).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to export events: %w", err)
	}

	result := make([]RewardEvent, 0, len(events))
	for _, z := range events {
		event, err := s.parseEventMember(z.Member.(string))
		if err != nil {
			s.logger.Warn("Failed to parse event during export", zap.Error(err))
			continue
		}

		// Set the actual timestamp from the score
		event.Timestamp = time.UnixMilli(int64(z.Score))
		result = append(result, event)
	}

	return result, nil
}

// UpdateConfig updates the window configuration
func (s *SlidingWindowStrategy) UpdateConfig(config *WindowConfig) {
	if config != nil {
		s.config = config
		s.logger.Info("Window config updated",
			zap.String("type", string(config.Type)),
			zap.Int("size", config.Size),
			zap.Int("min_samples", config.MinSamples),
		)
	}
}

// GetConfig returns the current window configuration
func (s *SlidingWindowStrategy) GetConfig() *WindowConfig {
	return s.config
}
