package logging

import (
	"os"
	"time"

	"github.com/bivex/paywall-iap/internal/infrastructure/config"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// Init initializes the global logger
func Init(cfg *config.SentryConfig) error {
	var err error
	var zapConfig zap.Config

	// Use development config in dev/staging, production in prod
	environment := "production"
	if cfg != nil && cfg.Environment != "" {
		environment = cfg.Environment
	}

	if environment == "development" {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Output to stdout by default
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	Logger, err = zapConfig.Build()
	if err != nil {
		return err
	}

	// Add Sentry if configured
	if cfg != nil && cfg.DSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.DSN,
			Environment:      cfg.Environment,
			Release:          cfg.Release,
			TracesSampleRate: 0.1,
		}); err != nil {
			Logger.Warn("Sentry init failed", zap.Error(err))
		} else {
			Logger.Info("Sentry initialized", zap.String("env", cfg.Environment))
		}
	}

	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	sentry.Flush(2 * time.Second)
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// WithComponent creates a child logger with a component field
func WithComponent(component string) *zap.Logger {
	return Logger.With(zap.String("component", component))
}

// WithRequestID creates a child logger with a request_id field
func WithRequestID(requestID string) *zap.Logger {
	return Logger.With(zap.String("request_id", requestID))
}

// WithUserID creates a child logger with a user_id field
func WithUserID(userID string) *zap.Logger {
	return Logger.With(zap.String("user_id", userID))
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
	os.Exit(1)
}
