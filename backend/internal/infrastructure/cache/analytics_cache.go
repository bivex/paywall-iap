package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AnalyticsCache handles caching for analytics data
type AnalyticsCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewAnalyticsCache creates a new analytics cache
func NewAnalyticsCache(client *redis.Client, logger *zap.Logger) *AnalyticsCache {
	return &AnalyticsCache{
		client: client,
		logger: logger,
	}
}

// Cache key constants
const (
	KeyRealtimeMetric    = "analytics:realtime:%s"
	KeyCohortData        = "analytics:cohort:%s:%s"
	KeyFunnelData        = "analytics:funnel:%s:%s"
	KeyLTV               = "analytics:ltv:%s"
	KeySegmentedLTV      = "analytics:ltv:segment:%s"
)

// TTL constants
const (
	TTLRealtime    = 30 * time.Second
	TTLCohort      = 1 * time.Hour
	TTLFunnel      = 30 * time.Minute
	TTLLTV         = 1 * time.Hour
	TTLSegmentedLTV = 2 * time.Hour
)

// RealtimeMetric represents a realtime metric
type RealtimeMetric struct {
	Name      string      `json:"name"`
	Value     float64     `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// SetRealtimeMetric stores a realtime metric with 30s TTL
func (c *AnalyticsCache) SetRealtimeMetric(ctx context.Context, metric *RealtimeMetric) error {
	key := fmt.Sprintf(KeyRealtimeMetric, metric.Name)

	data, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	if err := c.client.Set(ctx, key, data, TTLRealtime).Err(); err != nil {
		return fmt.Errorf("failed to set realtime metric: %w", err)
	}

	c.logger.Debug("Cached realtime metric",
		zap.String("name", metric.Name),
		zap.Float64("value", metric.Value),
	)

	return nil
}

// GetRealtimeMetric retrieves a realtime metric
func (c *AnalyticsCache) GetRealtimeMetric(ctx context.Context, name string) (*RealtimeMetric, error) {
	key := fmt.Sprintf(KeyRealtimeMetric, name)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("metric not found")
		}
		return nil, fmt.Errorf("failed to get realtime metric: %w", err)
	}

	var metric RealtimeMetric
	if err := json.Unmarshal(data, &metric); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metric: %w", err)
	}

	return &metric, nil
}

// IncrementRealtimeMetric atomically increments a realtime metric
func (c *AnalyticsCache) IncrementRealtimeMetric(ctx context.Context, name string, delta float64) error {
	key := fmt.Sprintf(KeyRealtimeMetric, name)

	// Use HINCRBYFLOAT for atomic increment
	if err := c.client.IncrByFloat(ctx, key+":value", delta).Err(); err != nil {
		return fmt.Errorf("failed to increment metric: %w", err)
	}

	// Update timestamp
	c.client.Set(ctx, key+":timestamp", time.Now().Unix(), TTLRealtime)

	// Set TTL if not exists
	c.client.Expire(ctx, key, TTLRealtime)

	return nil
}

// SetRealtimeMetrics stores multiple realtime metrics in a pipeline
func (c *AnalyticsCache) SetRealtimeMetrics(ctx context.Context, metrics []RealtimeMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()

	for _, metric := range metrics {
		key := fmt.Sprintf(KeyRealtimeMetric, metric.Name)
		data, err := json.Marshal(metric)
		if err != nil {
			c.logger.Warn("Failed to marshal metric", zap.String("name", metric.Name), zap.Error(err))
			continue
		}
		pipe.Set(ctx, key, data, TTLRealtime)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to set realtime metrics: %w", err)
	}

	c.logger.Debug("Batch cached realtime metrics", zap.Int("count", len(metrics)))
	return nil
}

// GetRealtimeMetrics retrieves multiple realtime metrics
func (c *AnalyticsCache) GetRealtimeMetrics(ctx context.Context, names []string) (map[string]*RealtimeMetric, error) {
	if len(names) == 0 {
		return make(map[string]*RealtimeMetric), nil
	}

	pipe := c.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(names))
	keys := make([]string, len(names))

	for i, name := range names {
		key := fmt.Sprintf(KeyRealtimeMetric, name)
		keys[i] = key
		cmds[i] = pipe.Get(ctx, key)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get realtime metrics: %w", err)
	}

	result := make(map[string]*RealtimeMetric)
	for i, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			if err == redis.Nil {
				continue // Not in cache
			}
			c.logger.Warn("Failed to get metric", zap.String("name", names[i]), zap.Error(err))
			continue
		}

		var metric RealtimeMetric
		if err := json.Unmarshal(data, &metric); err != nil {
			c.logger.Warn("Failed to unmarshal metric", zap.Error(err))
			continue
		}

		result[metric.Name] = &metric
	}

	return result, nil
}

// CohortData represents cached cohort data
type CohortData struct {
	MetricName string                 `json:"metric_name"`
	Date       time.Time              `json:"date"`
	CohortSize int                    `json:"cohort_size"`
	Retention  map[string]int         `json:"retention"`
	Revenue    map[string]float64     `json:"revenue"`
	CachedAt   time.Time              `json:"cached_at"`
}

// SetCohortData stores cohort data with 1h TTL
func (c *AnalyticsCache) SetCohortData(ctx context.Context, metricName string, date time.Time, data *CohortData) error {
	key := fmt.Sprintf(KeyCohortData, metricName, date.Format("2006-01-02"))

	data.CachedAt = time.Now()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cohort data: %w", err)
	}

	if err := c.client.Set(ctx, key, jsonData, TTLCohort).Err(); err != nil {
		return fmt.Errorf("failed to set cohort data: %w", err)
	}

	c.logger.Debug("Cached cohort data",
		zap.String("metric_name", metricName),
		zap.String("date", date.Format("2006-01-02")),
	)

	return nil
}

// GetCohortData retrieves cached cohort data
func (c *AnalyticsCache) GetCohortData(ctx context.Context, metricName string, date time.Time) (*CohortData, error) {
	key := fmt.Sprintf(KeyCohortData, metricName, date.Format("2006-01-02"))

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("cohort data not found")
		}
		return nil, fmt.Errorf("failed to get cohort data: %w", err)
	}

	var cohortData CohortData
	if err := json.Unmarshal(data, &cohortData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cohort data: %w", err)
	}

	return &cohortData, nil
}

// InvalidateCohort removes cached cohort data for a metric
func (c *AnalyticsCache) InvalidateCohort(ctx context.Context, metricName string) error {
	pattern := fmt.Sprintf(KeyCohortData, metricName, "*")

	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan cohort keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cohort keys: %w", err)
		}
		c.logger.Debug("Invalidated cohort data", zap.String("metric_name", metricName), zap.Int("count", len(keys)))
	}

	return nil
}

// FunnelData represents cached funnel data
type FunnelData struct {
	FunnelID      string                 `json:"funnel_id"`
	DateFrom      time.Time              `json:"date_from"`
	DateTo        time.Time              `json:"date_to"`
	Steps         []FunnelStep           `json:"steps"`
	TotalEntries  int                    `json:"total_entries"`
	TotalExits    int                    `json:"total_exits"`
	ConversionRate float64               `json:"conversion_rate"`
	CachedAt      time.Time              `json:"cached_at"`
}

// FunnelStep represents a step in the funnel
type FunnelStep struct {
	StepID      string  `json:"step_id"`
	StepName    string  `json:"step_name"`
	Visitors    int     `json:"visitors"`
	Dropoff     int     `json:"dropoff"`
	DropoffRate float64 `json:"dropoff_rate"`
}

// SetFunnelData stores funnel data with 30min TTL
func (c *AnalyticsCache) SetFunnelData(ctx context.Context, funnelID string, dateFrom, dateTo time.Time, data *FunnelData) error {
	key := fmt.Sprintf(KeyFunnelData, funnelID, fmt.Sprintf("%s:%s", dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02")))

	data.CachedAt = time.Now()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal funnel data: %w", err)
	}

	if err := c.client.Set(ctx, key, jsonData, TTLFunnel).Err(); err != nil {
		return fmt.Errorf("failed to set funnel data: %w", err)
	}

	c.logger.Debug("Cached funnel data", zap.String("funnel_id", funnelID))
	return nil
}

// GetFunnelData retrieves cached funnel data
func (c *AnalyticsCache) GetFunnelData(ctx context.Context, funnelID string, dateFrom, dateTo time.Time) (*FunnelData, error) {
	key := fmt.Sprintf(KeyFunnelData, funnelID, fmt.Sprintf("%s:%s", dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02")))

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("funnel data not found")
		}
		return nil, fmt.Errorf("failed to get funnel data: %w", err)
	}

	var funnelData FunnelData
	if err := json.Unmarshal(data, &funnelData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal funnel data: %w", err)
	}

	return &funnelData, nil
}

// LTVData represents cached LTV data
type LTVData struct {
	UserID       string            `json:"user_id"`
	LTV30        float64           `json:"ltv30"`
	LTV90        float64           `json:"ltv90"`
	LTV365       float64           `json:"ltv365"`
	LTVLifetime  float64           `json:"ltv_lifetime"`
	Confidence   float64           `json:"confidence"`
	CalculatedAt time.Time         `json:"calculated_at"`
	Factors      map[string]float64 `json:"factors"`
}

// SetLTV stores LTV data with 1h TTL
func (c *AnalyticsCache) SetLTV(ctx context.Context, userID string, data *LTVData) error {
	key := fmt.Sprintf(KeyLTV, userID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal LTV data: %w", err)
	}

	if err := c.client.Set(ctx, key, jsonData, TTLLTV).Err(); err != nil {
		return fmt.Errorf("failed to set LTV: %w", err)
	}

	c.logger.Debug("Cached LTV data", zap.String("user_id", userID))
	return nil
}

// GetLTV retrieves cached LTV data
func (c *AnalyticsCache) GetLTV(ctx context.Context, userID string) (*LTVData, error) {
	key := fmt.Sprintf(KeyLTV, userID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("LTV data not found")
		}
		return nil, fmt.Errorf("failed to get LTV: %w", err)
	}

	var ltvData LTVData
	if err := json.Unmarshal(data, &ltvData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LTV data: %w", err)
	}

	return &ltvData, nil
}

// InvalidateLTV removes cached LTV data for a user
func (c *AnalyticsCache) InvalidateLTV(ctx context.Context, userID string) error {
	key := fmt.Sprintf(KeyLTV, userID)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate LTV: %w", err)
	}

	c.logger.Debug("Invalidated LTV data", zap.String("user_id", userID))
	return nil
}

// GetCacheStats returns statistics about cache usage
func (c *AnalyticsCache) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	info, err := c.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	dbSize, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get db size: %w", err)
	}

	stats := &CacheStats{
		Info:   info,
		Keys:   dbSize,
		Memory: parseMemoryUsage(info),
	}

	return stats, nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	Info   string `json:"info"`
	Keys   int64  `json:"keys"`
	Memory int64  `json:"memory"`
}

// parseMemoryUsage extracts memory usage from Redis INFO
func parseMemoryUsage(info string) int64 {
	// Parse "used_memory:1234567" from INFO output
	// This is a simplified implementation
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if len(line) > 12 && line[:12] == "used_memory:" {
			var mem int64
			fmt.Sscanf(line[12:], "%d", &mem)
			return mem
		}
	}
	return 0
}

// FlushPattern removes all keys matching a pattern
func (c *AnalyticsCache) FlushPattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
		c.logger.Debug("Flushed pattern", zap.String("pattern", pattern), zap.Int("count", len(keys)))
	}

	return nil
}
