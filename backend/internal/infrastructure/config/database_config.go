package config

import (
	"time"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL            string
	MaxConnections int
	MinConnections int
	MaxLifetime    time.Duration
	MaxIdleTime    time.Duration
	HealthCheck    time.Duration
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		MaxConnections: 25,
		MinConnections: 5,
		MaxLifetime:    1 * time.Hour,
		MaxIdleTime:    30 * time.Minute,
		HealthCheck:    30 * time.Second,
	}
}
