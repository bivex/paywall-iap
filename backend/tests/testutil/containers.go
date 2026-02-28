package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcwait "github.com/testcontainers/testcontainers-go/wait"
)

// TestDBContainer holds the PostgreSQL test container
type TestDBContainer struct {
	Container  testcontainers.Container
	ConnString string
	Pool       *pgxpool.Pool
}

// SetupTestDBContainer starts a PostgreSQL test container
func SetupTestDBContainer(ctx context.Context, t *testing.T) (*TestDBContainer, error) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "iap_test",
		},
		WaitingFor: tcwait.ForAll(
			tcwait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			tcwait.ForListeningPort("5432/tcp"),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	connString := fmt.Sprintf("postgres://test:test@%s:%s/iap_test?sslmode=disable", host, mappedPort.Port())

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &TestDBContainer{
		Container:  container,
		ConnString: connString,
		Pool:       pool,
	}, nil
}

// Teardown cleans up the test container
func (tc *TestDBContainer) Teardown(ctx context.Context, t *testing.T) {
	t.Helper()
	if tc.Pool != nil {
		tc.Pool.Close()
	}
	if tc.Container != nil {
		if err := tc.Container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}
}
