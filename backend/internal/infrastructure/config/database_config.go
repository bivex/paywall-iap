package config

import (
	"time"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL            string        `mapstructure:"url"`
	MaxConnections int           `mapstructure:"max_connections"`
	MinConnections int           `mapstructure:"min_connections"`
	MaxLifetime    time.Duration `mapstructure:"max_lifetime"`
	MaxIdleTime    time.Duration `mapstructure:"max_idle_time"`
	HealthCheck    time.Duration `mapstructure:"health_check"`
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
