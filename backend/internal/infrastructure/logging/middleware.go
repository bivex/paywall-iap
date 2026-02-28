package logging

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestMiddleware creates a middleware that logs HTTP requests
func RequestMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Create request logger
		requestLogger := logger.With(zap.String("request_id", requestID))
		c.Set("logger", requestLogger)

		// Log request
		requestLogger.Debug("incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("client_ip", c.ClientIP()),
		)

		// Process request
		c.Next()

		// Log response
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		requestLogger.Info("request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
		)
	}
}

// GetLogger retrieves the logger from the Gin context
func GetLogger(c *gin.Context) *zap.Logger {
	if logger, exists := c.Get("logger"); exists {
		if l, ok := logger.(*zap.Logger); ok {
			return l
		}
	}
	return Logger
}
