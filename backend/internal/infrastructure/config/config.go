package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	IAP      IAPConfig      `mapstructure:"iap"`
	Sentry   SentryConfig   `mapstructure:"sentry"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
	Issuer     string        `mapstructure:"issuer"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL          string        `mapstructure:"url"`
	Password     string        `mapstructure:"password"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
}

// IAPConfig holds IAP configuration
type IAPConfig struct {
	AppleSharedSecret   string `mapstructure:"apple_shared_secret"`
	GoogleKeyJSON       string `mapstructure:"google_key_json"`
	StripeWebhookSecret string `mapstructure:"stripe_webhook_secret"`
	AppleWebhookSecret  string `mapstructure:"apple_webhook_secret"`
	GoogleWebhookSecret string `mapstructure:"google_webhook_secret"`
}

// SentryConfig holds Sentry configuration
type SentryConfig struct {
	DSN         string `mapstructure:"dsn"`
	Environment string `mapstructure:"environment"`
	Release     string `mapstructure:"release"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Explicitly bind environment variables
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("database.url", "DATABASE_URL")
	_ = viper.BindEnv("database.max_connections", "DATABASE_MAX_CONNECTIONS")
	_ = viper.BindEnv("redis.url", "REDIS_URL")
	_ = viper.BindEnv("jwt.secret", "JWT_SECRET")

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

	// Explicit webhook secret bindings (viper mapstructure doesn't auto-map these)
	cfg.IAP.StripeWebhookSecret = viper.GetString("STRIPE_WEBHOOK_SECRET")
	cfg.IAP.AppleWebhookSecret = viper.GetString("APPLE_WEBHOOK_SECRET")
	cfg.IAP.GoogleWebhookSecret = viper.GetString("GOOGLE_WEBHOOK_SECRET")

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 10*time.Second)
	viper.SetDefault("server.write_timeout", 10*time.Second)
	viper.SetDefault("server.shutdown_timeout", 30*time.Second)

	// Database defaults
	viper.SetDefault("database.max_connections", 25)
	viper.SetDefault("database.min_connections", 5)
	viper.SetDefault("database.max_lifetime", 1*time.Hour)
	viper.SetDefault("database.max_idle_time", 30*time.Minute)
	viper.SetDefault("database.health_check", 30*time.Second)

	// JWT defaults
	viper.SetDefault("jwt.access_ttl", 15*time.Minute)
	viper.SetDefault("jwt.refresh_ttl", 720*time.Hour)
	viper.SetDefault("jwt.issuer", "iap-system")

	// Redis defaults
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 3)
	viper.SetDefault("redis.dial_timeout", 5*time.Second)
	viper.SetDefault("redis.read_timeout", 3*time.Second)
	viper.SetDefault("redis.write_timeout", 3*time.Second)
	viper.SetDefault("redis.pool_timeout", 4*time.Second)
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
