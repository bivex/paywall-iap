package pool

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/password9090/paywall-iap/internal/infrastructure/config"
)

// NewPool creates a new PostgreSQL connection pool
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Max connections: sized at 25 per API instance.
	// With 2 API replicas + 1 worker = 75 total.
	// Postgres default max_connections = 100 → leaves 25 for migrations/admin.
	config.MaxConns = int32(cfg.MaxConnections)
	config.MinConns = int32(cfg.MinConnections)
	config.MaxConnLifetime = cfg.MaxLifetime
	config.MaxConnIdleTime = cfg.MaxIdleTime
	config.HealthCheckPeriod = cfg.HealthCheck

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

// Ping verifies the database connection is alive
func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}

// Close closes the connection pool
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}
