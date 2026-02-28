package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB creates a test database container
func SetupTestDB(ctx context.Context) (*pgxpool.Pool, func(), error) {
	// Create PostgreSQL container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, func() { pgContainer.Terminate(ctx) }, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Create connection pool
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, func() { pgContainer.Terminate(ctx) }, fmt.Errorf("failed to parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, func() { pgContainer.Terminate(ctx) }, fmt.Errorf("failed to create pool: %w", err)
	}

	// Cleanup function
	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}

	return pool, cleanup, nil
}

// RunMigrations runs migrations on the test database
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// For now, just create the core tables inline
	// In production, this would run golang-migrate
	schema := `
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		platform_user_id    TEXT UNIQUE NOT NULL,
		device_id           TEXT,
		platform            TEXT NOT NULL CHECK (platform IN ('ios', 'android')),
		app_version         TEXT NOT NULL,
		email               TEXT UNIQUE,
		ltv                 NUMERIC(10,2) DEFAULT 0,
		ltv_updated_at      TIMESTAMPTZ,
		created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
		deleted_at          TIMESTAMPTZ
	);

	-- Subscriptions table
	CREATE TABLE IF NOT EXISTS subscriptions (
		id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id         UUID NOT NULL REFERENCES users(id),
		status          TEXT NOT NULL CHECK (status IN ('active', 'expired', 'cancelled', 'grace')),
		source          TEXT NOT NULL CHECK (source IN ('iap', 'stripe', 'paddle')),
		platform        TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
		product_id      TEXT NOT NULL,
		plan_type       TEXT NOT NULL CHECK (plan_type IN ('monthly', 'annual', 'lifetime')),
		expires_at      TIMESTAMPTZ NOT NULL,
		auto_renew      BOOLEAN NOT NULL DEFAULT true,
		created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
		deleted_at      TIMESTAMPTZ
	);

	-- Transactions table
	CREATE TABLE IF NOT EXISTS transactions (
		id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id             UUID NOT NULL REFERENCES users(id),
		subscription_id     UUID NOT NULL REFERENCES subscriptions(id),
		amount              NUMERIC(10,2) NOT NULL,
		currency            CHAR(3) NOT NULL,
		status              TEXT NOT NULL CHECK (status IN ('success', 'failed', 'refunded')),
		receipt_hash        TEXT,
		provider_tx_id      TEXT,
		created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	`

	_, err := pool.Exec(ctx, schema)
	return err
}

// TruncateAll truncates all test tables
func TruncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	tables := []string{"transactions", "subscriptions", "users"}
	for _, table := range tables {
		if _, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			return err
		}
	}
	return nil
}

// AssertDBCount asserts the expected count of rows in a table
func AssertDBCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string, expected int) {
	var count int
	err := pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows in %s: %v", table, err)
	}
	if count != expected {
		t.Errorf("Expected %d rows in %s, got %d", expected, table, count)
	}
}
