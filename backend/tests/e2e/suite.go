//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcwait "github.com/testcontainers/testcontainers-go/wait"

	"github.com/bivex/paywall-iap/tests/testutil"
)

// E2ETestSuite holds the E2E test environment
type E2ETestSuite struct {
	DBContainer    *testutil.TestDBContainer
	RedisContainer testcontainers.Container
	APIServer      *testutil.TestServer
	BaseURL        string
	HTTPClient     *http.Client
}

// SetupE2ETestSuite starts all required containers and services
func SetupE2ETestSuite(ctx context.Context, t *testing.T) *E2ETestSuite {
	t.Helper()

	suite := &E2ETestSuite{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Start PostgreSQL
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	if err != nil {
		t.Fatalf("failed to start db container: %v", err)
	}
	suite.DBContainer = dbContainer

	// Start Redis (optional for E2E, skip if unavailable)
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   tcwait.ForLog("Ready to accept connections").WithOccurrence(1),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	if err != nil {
		t.Logf("Warning: could not start Redis container: %v (continuing without Redis)", err)
	} else {
		suite.RedisContainer = redisContainer
	}

	// Run migrations
	if err := testutil.RunMigrations(ctx, dbContainer.Pool); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Initialize repositories using mock repos (backed by real DB)
	userRepo := testutil.NewMockUserRepo(dbContainer.Pool)
	subRepo := testutil.NewMockSubscriptionRepo(dbContainer.Pool)

	// Start API server
	suite.APIServer = testutil.NewTestServer(ctx, dbContainer.Pool, userRepo, subRepo)
	suite.BaseURL = suite.APIServer.BaseURL()

	return suite
}

// Teardown cleans up all containers and services
func (suite *E2ETestSuite) Teardown(ctx context.Context, t *testing.T) {
	t.Helper()

	if suite.APIServer != nil {
		suite.APIServer.Close()
	}

	if suite.RedisContainer != nil {
		suite.RedisContainer.Terminate(ctx)
	}

	if suite.DBContainer != nil {
		suite.DBContainer.Teardown(ctx, t)
	}
}

// GetAPIURL returns the API base URL
func (suite *E2ETestSuite) GetAPIURL() string {
	return suite.BaseURL
}
