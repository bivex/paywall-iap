package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// RedisBanditCache implements bandit state caching using Redis
type RedisBanditCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisBanditCache creates a new Redis-backed bandit cache
func NewRedisBanditCache(client *redis.Client, logger *zap.Logger) *RedisBanditCache {
	return &RedisBanditCache{
		client: client,
		logger: logger,
	}
}

// ArmStatsCache represents the cached format for arm statistics
type ArmStatsCache struct {
	ArmID       uuid.UUID `json:"arm_id"`
	Alpha       float64   `json:"alpha"`
	Beta        float64   `json:"beta"`
	Samples     int       `json:"samples"`
	Conversions int       `json:"conversions"`
	Revenue     float64   `json:"revenue"`
	AvgReward   float64   `json:"avg_reward"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetArmStats retrieves arm statistics from cache
func (c *RedisBanditCache) GetArmStats(ctx context.Context, key string) (*service.ArmStats, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("arm stats not found in cache")
		}
		return nil, fmt.Errorf("failed to get arm stats: %w", err)
	}

	var cached ArmStatsCache
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arm stats: %w", err)
	}

	return &service.ArmStats{
		ArmID:       cached.ArmID,
		Alpha:       cached.Alpha,
		Beta:        cached.Beta,
		Samples:     cached.Samples,
		Conversions: cached.Conversions,
		Revenue:     cached.Revenue,
		AvgReward:   cached.AvgReward,
		UpdatedAt:   cached.UpdatedAt,
	}, nil
}

// SetArmStats stores arm statistics in cache
func (c *RedisBanditCache) SetArmStats(ctx context.Context, key string, stats *service.ArmStats, ttl time.Duration) error {
	cached := ArmStatsCache{
		ArmID:       stats.ArmID,
		Alpha:       stats.Alpha,
		Beta:        stats.Beta,
		Samples:     stats.Samples,
		Conversions: stats.Conversions,
		Revenue:     stats.Revenue,
		AvgReward:   stats.AvgReward,
		UpdatedAt:   stats.UpdatedAt,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("failed to marshal arm stats: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set arm stats: %w", err)
	}

	c.logger.Debug("Cached arm stats",
		zap.String("key", key),
		zap.String("arm_id", stats.ArmID.String()),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// GetAssignment retrieves a user's assigned arm from cache
func (c *RedisBanditCache) GetAssignment(ctx context.Context, key string) (uuid.UUID, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, fmt.Errorf("assignment not found in cache")
		}
		return uuid.Nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	armID, err := uuid.ParseBytes(data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse arm ID: %w", err)
	}

	return armID, nil
}

// SetAssignment stores a user's assigned arm in cache
func (c *RedisBanditCache) SetAssignment(ctx context.Context, key string, armID uuid.UUID, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, armID.Bytes(), ttl).Err(); err != nil {
		return fmt.Errorf("failed to set assignment: %w", err)
	}

	c.logger.Debug("Cached assignment",
		zap.String("key", key),
		zap.String("arm_id", armID.String()),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// InvalidateArmStats removes arm statistics from cache
func (c *RedisBanditCache) InvalidateArmStats(ctx context.Context, armID uuid.UUID) error {
	key := fmt.Sprintf("ab:arm:%s", armID.String())
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate arm stats: %w", err)
	}

	c.logger.Debug("Invalidated arm stats", zap.String("arm_id", armID.String()))
	return nil
}

// InvalidateAssignment removes a user's assignment from cache
func (c *RedisBanditCache) InvalidateAssignment(ctx context.Context, experimentID, userID uuid.UUID) error {
	key := fmt.Sprintf("ab:assign:%s:%s", experimentID.String(), userID.String())
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate assignment: %w", err)
	}

	c.logger.Debug("Invalidated assignment",
		zap.String("experiment_id", experimentID.String()),
		zap.String("user_id", userID.String()),
	)
	return nil
}

// BulkInvalidateAssignments removes all assignments for an experiment
func (c *RedisBanditCache) BulkInvalidateAssignments(ctx context.Context, experimentID uuid.UUID) error {
	pattern := fmt.Sprintf("ab:assign:%s:*", experimentID.String())
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan assignments: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete assignments: %w", err)
		}
		c.logger.Debug("Bulk invalidated assignments",
			zap.String("experiment_id", experimentID.String()),
			zap.Int("count", len(keys)),
		)
	}

	return nil
}

// GetArmStatsBatch retrieves multiple arm statistics in a single pipeline
func (c *RedisBanditCache) GetArmStatsBatch(ctx context.Context, armIDs []uuid.UUID) (map[uuid.UUID]*service.ArmStats, error) {
	if len(armIDs) == 0 {
		return make(map[uuid.UUID]*service.ArmStats), nil
	}

	// Build pipeline for batch get
	pipe := c.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(armIDs))
	keys := make([]string, len(armIDs))

	for i, armID := range armIDs {
		key := fmt.Sprintf("ab:arm:%s", armID.String())
		keys[i] = key
		cmds[i] = pipe.Get(ctx, key)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute batch get: %w", err)
	}

	// Parse results
	results := make(map[uuid.UUID]*service.ArmStats)
	for i, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			if err == redis.Nil {
				continue // Not in cache
			}
			return nil, fmt.Errorf("failed to get arm stats: %w", err)
		}

		var cached ArmStatsCache
		if err := json.Unmarshal(data, &cached); err != nil {
			c.logger.Warn("Failed to unmarshal arm stats", zap.Error(err))
			continue
		}

		results[cached.ArmID] = &service.ArmStats{
			ArmID:       cached.ArmID,
			Alpha:       cached.Alpha,
			Beta:        cached.Beta,
			Samples:     cached.Samples,
			Conversions: cached.Conversions,
			Revenue:     cached.Revenue,
			AvgReward:   cached.AvgReward,
			UpdatedAt:   cached.UpdatedAt,
		}
	}

	return results, nil
}

// SetArmStatsBatch stores multiple arm statistics in a single pipeline
func (c *RedisBanditCache) SetArmStatsBatch(ctx context.Context, statsList []*service.ArmStats, ttl time.Duration) error {
	if len(statsList) == 0 {
		return nil
	}

	// Build pipeline for batch set
	pipe := c.client.Pipeline()

	for _, stats := range statsList {
		key := fmt.Sprintf("ab:arm:%s", stats.ArmID.String())
		cached := ArmStatsCache{
			ArmID:       stats.ArmID,
			Alpha:       stats.Alpha,
			Beta:        stats.Beta,
			Samples:     stats.Samples,
			Conversions: stats.Conversions,
			Revenue:     stats.Revenue,
			AvgReward:   stats.AvgReward,
			UpdatedAt:   stats.UpdatedAt,
		}

		data, err := json.Marshal(cached)
		if err != nil {
			return fmt.Errorf("failed to marshal arm stats: %w", err)
		}

		pipe.Set(ctx, key, data, ttl)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute batch set: %w", err)
	}

	c.logger.Debug("Batch cached arm stats", zap.Int("count", len(statsList)))
	return nil
}

// Ping checks if the Redis cache is accessible
func (c *RedisBanditCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (c *RedisBanditCache) Close() error {
	return c.client.Close()
}
