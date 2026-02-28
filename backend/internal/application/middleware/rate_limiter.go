package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitConfig defines rate limiting parameters
type RateLimitConfig struct {
	Rate  int // requests per second
	Burst int // maximum burst size
}

// RateLimiter manages rate limiting using Redis
type RateLimiter struct {
	redis    *redis.Client
	limiter  *redis_rate.Limiter
	logger   *zap.Logger
	failOpen bool // if true, allow requests when Redis is unavailable
	prefix   string
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client, failOpen bool) *RateLimiter {
	limiter := redis_rate.NewLimiter(redisClient)
	return &RateLimiter{
		redis:    redisClient,
		limiter:  limiter,
		logger:   logging.Logger,
		failOpen: failOpen,
		prefix:   "ratelimit:",
	}
}

// Middleware returns a Gin middleware for rate limiting
func (r *RateLimiter) Middleware(keyFunc func(*gin.Context) string, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		// Create rate limiter for this key
		limiterKey := r.prefix + key
		limit := redis_rate.PerSecond(config.Rate)
		res, err := r.limiter.AllowN(context.Background(), limiterKey, limit, config.Burst)
		if err != nil {
			r.logger.Error("rate limiter error", zap.Error(err))
			if r.failOpen {
				// Fail open - allow the request but log it
				c.Next()
				return
			}
			// Fail closed - return service unavailable
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "SERVICE_UNAVAILABLE",
				"message": "Rate limiting unavailable",
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Rate))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", res.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(res.RetryAfter).Unix()))

		if res.Allowed == 0 {
			retryAfter := int(res.RetryAfter.Seconds()) + 1
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "RATE_LIMIT_EXCEEDED",
				"message":     "Rate limit exceeded",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Key functions for different rate limiting strategies

// ByIP limits requests by client IP address
func ByIP(c *gin.Context) string {
	return "ip:" + c.ClientIP()
}

// ByUserID limits requests by authenticated user ID
func ByUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return "user:" + userID.(string)
	}
	// Fall back to IP if not authenticated
	return ByIP(c)
}

// ByEndpoint limits requests by endpoint path
func ByEndpoint(c *gin.Context) string {
	return "endpoint:" + c.Request.URL.Path
}

// ByIPAndEndpoint limits requests by IP and endpoint combination
func ByIPAndEndpoint(c *gin.Context) string {
	return fmt.Sprintf("ip:%s:endpoint:%s", c.ClientIP(), c.Request.URL.Path)
}

// Predefined rate limit configurations for common endpoints
var (
	// Default rate limit: 100 requests per second (was 1)
	DefaultConfig = RateLimitConfig{
		Rate:  100,
		Burst: 200,
	}

	// Strict rate limit: 10 requests per second (was 1/min range)
	StrictConfig = RateLimitConfig{
		Rate:  10,
		Burst: 50,
	}

	// Generous rate limit: 500 requests per second
	GenerousConfig = RateLimitConfig{
		Rate:  500,
		Burst: 1000,
	}

	// Webhook rate limit: 1000 requests per second
	WebhookConfig = RateLimitConfig{
		Rate:  1000,
		Burst: 2000,
	}

	// High-frequency polling: 100 requests per second
	PollingConfig = RateLimitConfig{
		Rate:  100,
		Burst: 500,
	}
)
