package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	IAP      IAPConfig
	Sentry   SentryConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret          string
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
	Issuer          string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL            string
	Password       string
	PoolSize       int
	MinIdleConns   int
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	PoolTimeout    time.Duration
}

// IAPConfig holds IAP configuration
type IAPConfig struct {
	AppleSharedSecret string
	GoogleKeyJSON     string
}

// SentryConfig holds Sentry configuration
type SentryConfig struct {
	DSN         string
	Environment string
	Release     string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		// .env file is optional for production (env vars are used)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server_port", 8080)
	viper.SetDefault("server_read_timeout", 10*time.Second)
	viper.SetDefault("server_write_timeout", 10*time.Second)
	viper.SetDefault("server_shutdown_timeout", 30*time.Second)

	// JWT defaults
	viper.SetDefault("jwt_access_ttl", 15*time.Minute)
	viper.SetDefault("jwt_refresh_ttl", 720*time.Hour)
	viper.SetDefault("jwt_issuer", "iap-system")

	// Redis defaults
	viper.SetDefault("redis_pool_size", 10)
	viper.SetDefault("redis_min_idle_conns", 3)
	viper.SetDefault("redis_dial_timeout", 5*time.Second)
	viper.SetDefault("redis_read_timeout", 3*time.Second)
	viper.SetDefault("redis_write_timeout", 3*time.Second)
	viper.SetDefault("redis_pool_timeout", 4*time.Second)
}

func validate(cfg *Config) error {
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	return nil
}
