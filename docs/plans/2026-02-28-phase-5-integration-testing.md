# Phase 5: Integration Testing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build comprehensive integration, E2E, and load testing infrastructure for the IAP system with full API coverage, performance benchmarks, and automated test pipelines.

**Architecture:** Three-tier testing strategy:
1. **Integration Tests** - API-level tests with real database, testing handler/query/command interactions
2. **E2E Tests** - Full system tests with mobile app simulator, testing complete user journeys
3. **Load Tests** - Performance and stress tests with k6, establishing baseline metrics and breaking points

**Tech Stack:** 
- Go testing: testify, gomock, testcontainers-go
- E2E: Custom Go client + mobile simulator
- Load Testing: k6 (Grafana), Prometheus for metrics
- CI/CD: GitHub Actions workflows

**Prerequisites:** 
- Phase 1-4 complete (backend API handlers, mobile app, database layer)
- Docker Compose available for test environments
- Existing integration test: `backend/tests/integration/api_test.go`

---

## Implementation Status

### ✅ Completed (Pending)

### 🔄 In Progress

**Phase 5: Integration Testing**
- ⏳ Task 1-3: Test infrastructure setup
- ⏳ Task 4-8: Integration test suites
- ⏳ Task 9-12: E2E test framework
- ⏳ Task 13-16: Load testing infrastructure

---

## Task 1: Set Up Test Infrastructure - Test Containers

**Files:**
- Create: `backend/tests/testutil/containers.go`
- Create: `backend/tests/testutil/migrations.go`
- Modify: `backend/Makefile` (add test targets)

**Step 1: Create testcontainers helper**

```go
// backend/tests/testutil/containers.go
package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDBContainer holds the PostgreSQL test container
type TestDBContainer struct {
	Container testcontainers.Container
	ConnString string
	Pool *pgxpool.Pool
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
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort("5432/tcp"),
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
```

**Step 2: Create migration helper**

```go
// backend/tests/testutil/migrations.go
package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs all up migrations on the test database
func RunMigrations(ctx context.Context, connString string) error {
	// Get absolute path to migrations directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Navigate to migrations directory (adjust path as needed)
	migrationsPath := filepath.Join(wd, "../../migrations")
	
	// Check if we're in backend/tests directory
	if _, err := os.Stat(filepath.Join(migrationsPath, "001_create_users.up.sql")); os.IsNotExist(err) {
		// Try alternative path
		migrationsPath = filepath.Join(wd, "../../../migrations")
	}

	db, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		&postgres.Postgres{},
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := db.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RollbackMigrations rolls back all migrations
func RollbackMigrations(ctx context.Context, connString string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	migrationsPath := filepath.Join(wd, "../../migrations")
	if _, err := os.Stat(filepath.Join(migrationsPath, "001_create_users.up.sql")); os.IsNotExist(err) {
		migrationsPath = filepath.Join(wd, "../../../migrations")
	}

	db, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		&postgres.Postgres{},
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := db.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return nil
}
```

**Step 3: Update Makefile with test targets**

```makefile
# backend/Makefile - Add these targets

# Test targets
test-all: test-unit test-integration test-e2e

test-unit:
	go test -race -count=1 ./internal/domain/... ./internal/application/...

test-integration:
	go test -race -count=1 -tags=integration ./tests/integration/...

test-e2e:
	go test -race -count=1 -tags=e2e ./tests/e2e/...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@awk '/^total:/ { if ($$3+0 < 80) { print "Coverage below 80%: " $$3; exit 1 } }' coverage.out

test-clean:
	go clean -testcache

# Load test targets
load-test:
	k6 run tests/load/basic_load_test.js

load-test-stress:
	k6 run tests/load/stress_test.js

load-test-soak:
	k6 run tests/load/soak_test.js
```

**Step 4: Verify testcontainers setup**

Run: `cd backend && go get github.com/testcontainers/testcontainers-go`

Expected: Package downloaded successfully

**Step 5: Commit**

```bash
git add backend/tests/testutil/containers.go backend/tests/testutil/migrations.go backend/Makefile
git commit -m "test: add testcontainers infrastructure and migration helpers"
```

---

## Task 2: Create Mock Repositories and Factories

**Files:**
- Create: `backend/tests/mocks/user_repository_mock.go`
- Create: `backend/tests/mocks/subscription_repository_mock.go`
- Create: `backend/tests/testutil/factories.go`

**Step 1: Create user repository mock**

```go
// backend/tests/mocks/user_repository_mock.go
package mocks

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
	mu      sync.RWMutex
	users   map[uuid.UUID]*entity.User
	byEmail map[string]*entity.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:   make(map[uuid.UUID]*entity.User),
		byEmail: make(map[string]*entity.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, user)
	if args.Error(0) != nil {
		return args.Error(0)
	}

	m.users[user.ID] = user
	if user.Email != "" {
		m.byEmail[user.Email] = user
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(ctx, id)
	if args.Error(0) != nil {
		return nil, args.Error(0)
	}
	return args.Get(0).(*entity.User), nil
}

func (m *MockUserRepository) GetByPlatformID(ctx context.Context, platformUserID string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(ctx, platformUserID)
	if args.Error(0) != nil {
		return nil, args.Error(0)
	}
	return args.Get(0).(*entity.User), nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(ctx, email)
	if args.Error(0) != nil {
		return nil, args.Error(0)
	}
	return args.Get(0).(*entity.User), nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, user)
	if args.Error(0) != nil {
		return args.Error(0)
	}

	m.users[user.ID] = user
	if user.Email != "" {
		m.byEmail[user.Email] = user
	}
	return nil
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, id)
	if args.Error(0) != nil {
		return args.Error(0)
	}

	if user, ok := m.users[id]; ok {
		now := user.CreatedAt // Use created_at as deleted_at for simplicity
		user.DeletedAt = &now
	}
	return nil
}

func (m *MockUserRepository) ExistsByPlatformID(ctx context.Context, platformUserID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(ctx, platformUserID)
	return args.Bool(0), args.Error(1)
}

// Ensure MockUserRepository implements repository.UserRepository
var _ repository.UserRepository = (*MockUserRepository)(nil)
```

**Step 2: Create subscription repository mock**

```go
// backend/tests/mocks/subscription_repository_mock.go
package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

// MockSubscriptionRepository is a mock implementation of SubscriptionRepository
type MockSubscriptionRepository struct {
	mock.Mock
	mu            sync.RWMutex
	subscriptions map[uuid.UUID]*entity.Subscription
	byUserID      map[uuid.UUID][]*entity.Subscription
}

func NewMockSubscriptionRepository() *MockSubscriptionRepository {
	return &MockSubscriptionRepository{
		subscriptions: make(map[uuid.UUID]*entity.Subscription),
		byUserID:      make(map[uuid.UUID][]*entity.Subscription),
	}
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, subscription *entity.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, subscription)
	if args.Error(0) != nil {
		return args.Error(0)
	}

	m.subscriptions[subscription.ID] = subscription
	m.byUserID[subscription.UserID] = append(m.byUserID[subscription.UserID], subscription)
	return nil
}

func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(ctx, id)
	if args.Error(0) != nil {
		return nil, args.Error(0)
	}
	return args.Get(0).(*entity.Subscription), nil
}

func (m *MockSubscriptionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(ctx, userID)
	if args.Error(0) != nil {
		return nil, args.Error(0)
	}
	return args.Get(0).(*entity.Subscription), nil
}

func (m *MockSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(ctx, userID)
	if args.Error(0) != nil {
		return nil, args.Error(0)
	}
	return args.Get(0).([]*entity.Subscription), nil
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, subscription *entity.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, subscription)
	if args.Error(0) != nil {
		return args.Error(0)
	}

	m.subscriptions[subscription.ID] = subscription
	return nil
}

func (m *MockSubscriptionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SubscriptionStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, id, status)
	if args.Error(0) != nil {
		return args.Error(0)
	}

	if sub, ok := m.subscriptions[id]; ok {
		sub.Status = status
		sub.UpdatedAt = time.Now()
	}
	return nil
}

func (m *MockSubscriptionRepository) UpdateExpiry(ctx context.Context, id uuid.UUID, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, id, expiresAt)
	if args.Error(0) != nil {
		return args.Error(0)
	}

	if sub, ok := m.subscriptions[id]; ok {
		sub.ExpiresAt = expiresAt
		sub.UpdatedAt = time.Now()
	}
	return nil
}

func (m *MockSubscriptionRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, id)
	if args.Error(0) != nil {
		return args.Error(0)
	}

	if sub, ok := m.subscriptions[id]; ok {
		sub.Status = entity.StatusCancelled
		sub.AutoRenew = false
		sub.UpdatedAt = time.Now()
	}
	return nil
}

func (m *MockSubscriptionRepository) CanAccess(ctx context.Context, userID uuid.UUID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// Ensure MockSubscriptionRepository implements repository.SubscriptionRepository
var _ repository.SubscriptionRepository = (*MockSubscriptionRepository)(nil)
```

**Step 3: Create test factories**

```go
// backend/tests/testutil/factories.go
package testutil

import (
	"time"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// UserFactory creates test user entities
type UserFactory struct{}

func NewUserFactory() *UserFactory {
	return &UserFactory{}
}

func (f *UserFactory) Create(platform entity.Platform, withEmail bool) *entity.User {
	platformUserID := uuid.New().String()
	var email string
	if withEmail {
		email = "test_" + uuid.New().String()[:8] + "@example.com"
	}

	return entity.NewUser(
		platformUserID,
		"test-device-"+uuid.New().String()[:8],
		platform,
		"1.0.0",
		email,
	)
}

func (f *UserFactory) CreateWithPlatformUserID(platformUserID string, platform entity.Platform) *entity.User {
	return entity.NewUser(
		platformUserID,
		"test-device-"+uuid.New().String()[:8],
		platform,
		"1.0.0",
		"test_"+platformUserID[:8]+"@example.com",
	)
}

// SubscriptionFactory creates test subscription entities
type SubscriptionFactory struct{}

func NewSubscriptionFactory() *SubscriptionFactory {
	return &SubscriptionFactory{}
}

func (f *SubscriptionFactory) CreateActive(userID uuid.UUID, source entity.SubscriptionSource, planType entity.PlanType) *entity.Subscription {
	sub := entity.NewSubscription(
		userID,
		source,
		string(source),
		"com.app.premium",
		planType,
		time.Now().Add(30*24*time.Hour),
	)
	return sub
}

func (f *SubscriptionFactory) CreateExpired(userID uuid.UUID) *entity.Subscription {
	sub := entity.NewSubscription(
		userID,
		entity.SourceIAP,
		"ios",
		"com.app.premium",
		entity.PlanMonthly,
		time.Now().Add(-24*time.Hour),
	)
	sub.Status = entity.StatusExpired
	return sub
}

func (f *SubscriptionFactory) CreateCancelled(userID uuid.UUID) *entity.Subscription {
	sub := entity.NewSubscription(
		userID,
		entity.SourceIAP,
		"ios",
		"com.app.premium",
		entity.PlanMonthly,
		time.Now().Add(30*24*time.Hour),
	)
	sub.Status = entity.StatusCancelled
	sub.AutoRenew = false
	return sub
}

// TransactionFactory creates test transaction entities
type TransactionFactory struct{}

func NewTransactionFactory() *TransactionFactory {
	return &TransactionFactory{}
}

func (f *TransactionFactory) CreateSuccessful(userID, subscriptionID uuid.UUID, amount float64) *entity.Transaction {
	tx := entity.NewTransaction(userID, subscriptionID, amount, "USD")
	tx.Status = entity.TransactionStatusSuccess
	tx.ReceiptHash = "sha256_" + uuid.New().String()
	tx.ProviderTxID = "tx_" + uuid.New().String()
	return tx
}

func (f *TransactionFactory) CreateFailed(userID, subscriptionID uuid.UUID) *entity.Transaction {
	tx := entity.NewTransaction(userID, subscriptionID, 9.99, "USD")
	tx.Status = entity.TransactionStatusFailed
	return tx
}
```

**Step 4: Verify mocks compile**

Run: `cd backend && go build ./tests/mocks/... ./tests/testutil/...`

Expected: No errors

**Step 5: Commit**

```bash
git add backend/tests/mocks/ backend/tests/testutil/factories.go
git commit -m "test: add mock repositories and test factories"
```

---

## Task 3: Create Test API Client

**Files:**
- Create: `backend/tests/testutil/api_client.go`
- Create: `backend/tests/testutil/test_server.go`

**Step 1: Create test server helper**

```go
// backend/tests/testutil/test_server.go
package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/application/query"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	"github.com/bivex/paywall-iap/internal/interfaces/http/middleware as http_middleware"
)

// TestServer holds the test HTTP server and dependencies
type TestServer struct {
	Server           *httptest.Server
	Router           *gin.Engine
	Pool             *pgxpool.Pool
	UserRepo         repository.UserRepository
	SubscriptionRepo repository.SubscriptionRepository
	JWTMiddleware    *middleware.JWTMiddleware
}

// NewTestServer creates a new test server with all handlers configured
func NewTestServer(ctx context.Context, pool *pgxpool.Pool, userRepo, subRepo repository.UserRepository, subscriptionRepo repository.SubscriptionRepository) *TestServer {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtMiddleware := middleware.NewJWTMiddleware("test-secret-32-characters!!", nil, 15*time.Minute)

	// Initialize queries and commands
	getSubQuery := query.NewGetSubscriptionQuery(subscriptionRepo)
	checkAccessQuery := query.NewCheckAccessQuery(subscriptionRepo)
	cancelCmd := command.NewCancelSubscriptionCommand(subscriptionRepo)
	registerCmd := command.NewRegisterCommand(userRepo, jwtMiddleware)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(registerCmd, jwtMiddleware)
	subscriptionHandler := handlers.NewSubscriptionHandler(getSubQuery, checkAccessQuery, cancelCmd, jwtMiddleware)

	// Setup routes
	v1 := router.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		subscription := v1.Group("/subscription")
		subscription.Use(http_middleware.JWTAuthMiddleware(jwtMiddleware))
		{
			subscription.GET("", subscriptionHandler.GetSubscription)
			subscription.GET("/access", subscriptionHandler.CheckAccess)
			subscription.DELETE("", subscriptionHandler.CancelSubscription)
		}
	}

	server := httptest.NewServer(router)

	return &TestServer{
		Server:           server,
		Router:           router,
		Pool:             pool,
		UserRepo:         userRepo,
		SubscriptionRepo: subscriptionRepo,
		JWTMiddleware:    jwtMiddleware,
	}
}

// Close shuts down the test server
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// BaseURL returns the test server URL
func (ts *TestServer) BaseURL() string {
	return ts.Server.URL
}

// NewRequest creates a new HTTP request with authentication
func (ts *TestServer) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	return NewTestRequest(method, ts.BaseURL()+path, body, "")
}

// NewAuthenticatedRequest creates a new authenticated HTTP request
func (ts *TestServer) NewAuthenticatedRequest(method, path string, body interface{}, userID string) (*http.Request, error) {
	token, err := ts.JWTMiddleware.GenerateToken(userID)
	if err != nil {
		return nil, err
	}
	return NewTestRequest(method, ts.BaseURL()+path, body, token)
}
```

**Step 2: Create API client**

```go
// backend/tests/testutil/api_client.go
package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// TestRequest creates a new HTTP request for testing
func NewTestRequest(method, url string, body interface{}, token string) (*http.Request, error) {
	var req *http.Request
	var err error

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

// DoRequest executes a request and returns the response
func DoRequest(client *http.Client, req *http.Request) (*http.Response, []byte, error) {
	if client == nil {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var body []byte
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, nil, err
	}
	body = buf.Bytes()

	return resp, body, nil
}

// DecodeResponse decodes a JSON response into the provided struct
func DecodeResponse(body []byte, v interface{}) error {
	return json.Unmarshal(body, v)
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Success      bool   `json:"success"`
	Data         AuthData `json:"data"`
}

type AuthData struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// SubscriptionResponse represents subscription response
type SubscriptionResponse struct {
	Success bool       `json:"success"`
	Data    SubscriptionData `json:"data"`
}

type SubscriptionData struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	PlanType  string `json:"plan_type"`
	ExpiresAt string `json:"expires_at"`
	AutoRenew bool   `json:"auto_renew"`
}

// AccessCheckResponse represents access check response
type AccessCheckResponse struct {
	Success  bool   `json:"success"`
	Data     AccessData `json:"data"`
}

type AccessData struct {
	HasAccess bool   `json:"has_access"`
	ExpiresAt string `json:"expires_at,omitempty"`
	PlanType  string `json:"plan_type,omitempty"`
}
```

**Step 3: Verify test server compiles**

Run: `cd backend && go build ./tests/testutil/...`

Expected: No errors (fix any import path issues)

**Step 4: Commit**

```bash
git add backend/tests/testutil/api_client.go backend/tests/testutil/test_server.go
git commit -m "test: add test server and API client helpers"
```

---

## Task 4: Integration Tests - Authentication Endpoints

**Files:**
- Create: `backend/tests/integration/auth_test.go`
- Test: `backend/tests/integration/auth_test.go`

**Step 1: Write registration tests**

```go
// backend/tests/integration/auth_test.go
//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestAuthRegistration(t *testing.T) {
	ctx := context.Background()

	// Setup test database container
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := testutil.NewMockUserRepo(dbContainer.Pool)
	subRepo := testutil.NewMockSubscriptionRepo(dbContainer.Pool)

	// Create test server
	testServer := testutil.NewTestServer(ctx, dbContainer.Pool, userRepo, subRepo)
	defer testServer.Close()

	t.Run("POST /v1/auth/register - successful registration", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"platform_user_id": "test-platform-user-" + time.Now().Format("20060102150405"),
			"device_id":        "test-device-id",
			"platform":         "ios",
			"app_version":      "1.0.0",
			"email":            "test_" + time.Now().Format("20060102150405") + "@example.com",
		}

		req, err := testServer.NewRequest("POST", "/v1/auth/register", reqBody)
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var apiResp testutil.AuthResponse
		err = json.Unmarshal(body, &apiResp)
		require.NoError(t, err)

		assert.True(t, apiResp.Success)
		assert.NotEmpty(t, apiResp.Data.UserID)
		assert.NotEmpty(t, apiResp.Data.AccessToken)
		assert.NotEmpty(t, apiResp.Data.RefreshToken)
		assert.Equal(t, 900, apiResp.Data.ExpiresIn) // 15 minutes
	})

	t.Run("POST /v1/auth/register - duplicate platform_user_id", func(t *testing.T) {
		platformUserID := "duplicate-user-" + time.Now().Format("20060102150405")

		// First registration
		reqBody1 := map[string]interface{}{
			"platform_user_id": platformUserID,
			"device_id":        "device-1",
			"platform":         "ios",
			"app_version":      "1.0.0",
			"email":            "test1@example.com",
		}

		req1, _ := testServer.NewRequest("POST", "/v1/auth/register", reqBody1)
		resp1, _, _ := testutil.DoRequest(nil, req1)
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		// Duplicate registration
		reqBody2 := map[string]interface{}{
			"platform_user_id": platformUserID,
			"device_id":        "device-2",
			"platform":         "ios",
			"app_version":      "1.0.0",
			"email":            "test2@example.com",
		}

		req2, _ := testServer.NewRequest("POST", "/v1/auth/register", reqBody2)
		resp2, body2, _ := testutil.DoRequest(nil, req2)

		assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

		var errorResp map[string]interface{}
		json.Unmarshal(body2, &errorResp)
		assert.False(t, errorResp["success"].(bool))
		assert.Contains(t, errorResp["error"].(string), "already exists")
	})

	t.Run("POST /v1/auth/register - invalid platform", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"platform_user_id": "test-user",
			"device_id":        "device-id",
			"platform":         "invalid_platform",
			"app_version":      "1.0.0",
			"email":            "test@example.com",
		}

		req, _ := testServer.NewRequest("POST", "/v1/auth/register", reqBody)
		resp, body, _ := testutil.DoRequest(nil, req)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)
		assert.False(t, errorResp["success"].(bool))
	})

	t.Run("POST /v1/auth/register - missing required fields", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"device_id": "device-id",
			// Missing platform_user_id, platform, app_version
		}

		req, _ := testServer.NewRequest("POST", "/v1/auth/register", reqBody)
		resp, _, _ := testutil.DoRequest(nil, req)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
```

**Step 2: Run auth integration tests**

Run: `cd backend && go test -tags=integration -v ./tests/integration/auth_test.go`

Expected: All tests pass

**Step 3: Commit**

```bash
git add backend/tests/integration/auth_test.go
git commit -m "test: add authentication integration tests"
```

---

## Task 5: Integration Tests - Subscription Endpoints

**Files:**
- Create: `backend/tests/integration/subscription_test.go`

**Step 1: Write subscription tests**

```go
// backend/tests/integration/subscription_test.go
//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestSubscriptionEndpoints(t *testing.T) {
	ctx := context.Background()

	// Setup test database container
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := testutil.NewMockUserRepo(dbContainer.Pool)
	subRepo := testutil.NewMockSubscriptionRepo(dbContainer.Pool)

	// Create test server
	testServer := testutil.NewTestServer(ctx, dbContainer.Pool, userRepo, subRepo)
	defer testServer.Close()

	// Create test user
	userFactory := testutil.NewUserFactory()
	user := userFactory.Create(entity.PlatformiOS, true)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test subscription
	subFactory := testutil.NewSubscriptionFactory()
	sub := subFactory.CreateActive(user.ID, entity.SourceIAP, entity.PlanMonthly)
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	t.Run("GET /v1/subscription - returns subscription details", func(t *testing.T) {
		req, err := testServer.NewAuthenticatedRequest("GET", "/v1/subscription", nil, user.ID.String())
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp testutil.SubscriptionResponse
		err = json.Unmarshal(body, &apiResp)
		require.NoError(t, err)

		assert.True(t, apiResp.Success)
		assert.Equal(t, sub.ID.String(), apiResp.Data.ID)
		assert.Equal(t, "active", apiResp.Data.Status)
		assert.Equal(t, "monthly", apiResp.Data.PlanType)
	})

	t.Run("GET /v1/subscription/access - returns has_access=true", func(t *testing.T) {
		req, err := testServer.NewAuthenticatedRequest("GET", "/v1/subscription/access", nil, user.ID.String())
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp testutil.AccessCheckResponse
		err = json.Unmarshal(body, &apiResp)
		require.NoError(t, err)

		assert.True(t, apiResp.Success)
		assert.True(t, apiResp.Data.HasAccess)
		assert.NotEmpty(t, apiResp.Data.ExpiresAt)
	})

	t.Run("GET /v1/subscription - no subscription returns 404", func(t *testing.T) {
		// Create user without subscription
		user2 := userFactory.Create(entity.PlatformAndroid, true)
		err = userRepo.Create(ctx, user2)
		require.NoError(t, err)

		req, err := testServer.NewAuthenticatedRequest("GET", "/v1/subscription", nil, user2.ID.String())
		require.NoError(t, err)

		resp, _, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("DELETE /v1/subscription - cancels subscription", func(t *testing.T) {
		req, err := testServer.NewAuthenticatedRequest("DELETE", "/v1/subscription", nil, user.ID.String())
		require.NoError(t, err)

		resp, _, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify subscription is cancelled
		updatedSub, err := subRepo.GetByID(ctx, sub.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusCancelled, updatedSub.Status)
		assert.False(t, updatedSub.AutoRenew)
	})

	t.Run("GET /v1/subscription/access - expired subscription returns has_access=false", func(t *testing.T) {
		// Create user with expired subscription
		user3 := userFactory.Create(entity.PlatformiOS, true)
		err = userRepo.Create(ctx, user3)
		require.NoError(t, err)

		expiredSub := subFactory.CreateExpired(user3.ID)
		err = subRepo.Create(ctx, expiredSub)
		require.NoError(t, err)

		req, err := testServer.NewAuthenticatedRequest("GET", "/v1/subscription/access", nil, user3.ID.String())
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp testutil.AccessCheckResponse
		json.Unmarshal(body, &apiResp)

		assert.True(t, apiResp.Success)
		assert.False(t, apiResp.Data.HasAccess)
	})
}
```

**Step 2: Run subscription integration tests**

Run: `cd backend && go test -tags=integration -v ./tests/integration/subscription_test.go`

Expected: All tests pass

**Step 3: Commit**

```bash
git add backend/tests/integration/subscription_test.go
git commit -m "test: add subscription integration tests"
```

---

## Task 6: Integration Tests - IAP Verification Endpoints

**Files:**
- Create: `backend/tests/integration/iap_test.go`
- Create: `backend/tests/mocks/iap_service_mock.go`

**Step 1: Create IAP service mock**

```go
// backend/tests/mocks/iap_service_mock.go
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// MockIAPService is a mock implementation of IAP verification service
type MockIAPService struct {
	mock.Mock
}

func NewMockIAPService() *MockIAPService {
	return &MockIAPService{}
}

func (m *MockIAPService) VerifyReceipt(ctx context.Context, platform entity.Platform, receiptData string) (*entity.IAPReceipt, error) {
	args := m.Called(ctx, platform, receiptData)
	if args.Error(0) != nil {
		return nil, args.Error(0)
	}
	return args.Get(0).(*entity.IAPReceipt), nil
}

func (m *MockIAPService) RenewSubscription(ctx context.Context, platform entity.Platform, receiptData string) (*entity.IAPReceipt, error) {
	args := m.Called(ctx, platform, receiptData)
	if args.Error(0) != nil {
		return nil, args.Error(0)
	}
	return args.Get(0).(*entity.IAPReceipt), nil
}

func (m *MockIAPService) CancelSubscription(ctx context.Context, platform entity.Platform, originalTransactionID string) error {
	args := m.Called(ctx, platform, originalTransactionID)
	return args.Error(0)
}
```

**Step 2: Write IAP verification tests**

```go
// backend/tests/integration/iap_test.go
//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestIAPVerification(t *testing.T) {
	ctx := context.Background()

	// Setup test database container
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := testutil.NewMockUserRepo(dbContainer.Pool)
	subRepo := testutil.NewMockSubscriptionRepo(dbContainer.Pool)
	iapService := mocks.NewMockIAPService()

	// Create test server with IAP handler
	testServer := setupIAPTestServer(ctx, dbContainer.Pool, userRepo, subRepo, iapService)
	defer testServer.Close()

	// Create test user
	userFactory := testutil.NewUserFactory()
	user := userFactory.Create(entity.PlatformiOS, true)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("POST /v1/verify/iap - valid receipt creates subscription", func(t *testing.T) {
		// Mock IAP service response
		iapService.On("VerifyReceipt", mock.Anything, entity.PlatformiOS, "valid_receipt_data").
			Return(&entity.IAPReceipt{
				OriginalTransactionID: "original_tx_123",
				ProductID:             "com.app.premium",
				ExpiresDate:           time.Now().Add(30 * 24 * time.Hour),
			}, nil).Once()

		reqBody := map[string]interface{}{
			"platform":      "ios",
			"receipt_data":  "valid_receipt_data",
			"product_id":    "com.app.premium",
			"transaction_id": "tx_123",
		}

		req, err := testServer.NewAuthenticatedRequest("POST", "/v1/verify/iap", reqBody, user.ID.String())
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp map[string]interface{}
		json.Unmarshal(body, &apiResp)
		assert.True(t, apiResp["success"].(bool))

		// Verify subscription was created
		sub, err := subRepo.GetActiveByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusActive, sub.Status)
	})

	t.Run("POST /v1/verify/iap - duplicate receipt returns existing subscription", func(t *testing.T) {
		// Mock IAP service for duplicate receipt
		iapService.On("VerifyReceipt", mock.Anything, entity.PlatformiOS, "duplicate_receipt").
			Return(&entity.IAPReceipt{
				OriginalTransactionID: "original_tx_456",
				ProductID:             "com.app.premium",
				ExpiresDate:           time.Now().Add(30 * 24 * time.Hour),
			}, nil).Once()

		// First verification
		reqBody1 := map[string]interface{}{
			"platform":      "ios",
			"receipt_data":  "duplicate_receipt",
			"product_id":    "com.app.premium",
			"transaction_id": "tx_456",
		}

		req1, _ := testServer.NewAuthenticatedRequest("POST", "/v1/verify/iap", reqBody1, user.ID.String())
		resp1, _, _ := testutil.DoRequest(nil, req1)
		assert.Equal(t, http.StatusOK, resp1.StatusCode)

		// Duplicate verification (same receipt)
		req2, _ := testServer.NewAuthenticatedRequest("POST", "/v1/verify/iap", reqBody1, user.ID.String())
		resp2, body2, _ := testutil.DoRequest(nil, req2)

		assert.Equal(t, http.StatusOK, resp2.StatusCode)

		var apiResp map[string]interface{}
		json.Unmarshal(body2, &apiResp)
		assert.True(t, apiResp["success"].(bool))
	})

	t.Run("POST /v1/verify/iap - invalid receipt returns error", func(t *testing.T) {
		// Mock IAP service error
		iapService.On("VerifyReceipt", mock.Anything, entity.PlatformiOS, "invalid_receipt").
			Return(nil, assert.AnError).Once()

		reqBody := map[string]interface{}{
			"platform":      "ios",
			"receipt_data":  "invalid_receipt",
			"product_id":    "com.app.premium",
			"transaction_id": "tx_789",
		}

		req, _ := testServer.NewAuthenticatedRequest("POST", "/v1/verify/iap", reqBody, user.ID.String())
		resp, body, _ := testutil.DoRequest(nil, req)

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)
		assert.False(t, errorResp["success"].(bool))
	})
}

func setupIAPTestServer(ctx context.Context, pool *pgxpool.Pool, userRepo, subRepo repository.UserRepository, iapService *mocks.MockIAPService) *testutil.TestServer {
	// TODO: Implement IAP handler setup
	// This requires VerifyIAPCommand and IAPHandler initialization
	return testutil.NewTestServer(ctx, pool, userRepo, subRepo)
}
```

**Step 3: Run IAP integration tests**

Run: `cd backend && go test -tags=integration -v ./tests/integration/iap_test.go`

Expected: Tests pass (may need to complete IAP handler implementation first)

**Step 4: Commit**

```bash
git add backend/tests/integration/iap_test.go backend/tests/mocks/iap_service_mock.go
git commit -m "test: add IAP verification integration tests"
```

---

## Task 7: Integration Tests - Webhook Handlers

**Files:**
- Create: `backend/tests/integration/webhook_test.go`
- Create: `backend/tests/testdata/apple_notification.json`
- Create: `backend/tests/testdata/google_notification.json`

**Step 1: Create test webhook payloads**

```json
// backend/tests/testdata/apple_notification.json
{
  "signedPayload": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJyZW5ld2FsX2luZm8iOnsib3JpZ2luYWxfdHJhbnNhY3Rpb25faWQiOiIxMDAwMDAwMDAwMDAwMDAiLCJhdXRvX3JlbmV3X3Byb2R1Y3RfaWQiOiJjb20uYXBwLnByZW1pdW0iLCJleHBpcmVzX2RhdGUiOiIxNzA5MjUxMjAwMDAwIn19.signature",
  "version": "2.0",
  "notificationType": "DID_RENEW",
  "environment": "Sandbox"
}
```

```json
// backend/tests/testdata/google_notification.json
{
  "message": {
    "data": "eyJvbmVUaW1lUGF5bG9hZCI6IntcInZlcnNpb25cIjpcIjEuMFwiLFwicGFja2FnZU5hbWVcIjpcImNvbS5hcHAucHJlbWl1bVwiLFwiZXZlbnRUeXBlVGV4dFwiOlwiU1VCU0NSSVBUSU9OX1JFTkVXQURcIixcInB1cmNoYXNlVG9rZW5cIjpcImduYS40MDAwMDAwMC0wMDAwMDAwMC0wMDAwMDAwMFwifSJ9",
    "messageId": "msg_123",
    "publishTime": "2024-02-28T12:00:00Z"
  },
  "subscription": "projects/your-project/topics/google-play-billing"
}
```

**Step 2: Write webhook tests**

```go
// backend/tests/integration/webhook_test.go
//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/interfaces/webhook/handlers"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestAppleWebhookHandler(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := testutil.NewMockUserRepo(dbContainer.Pool)
	subRepo := testutil.NewMockSubscriptionRepo(dbContainer.Pool)

	// Create test user and subscription
	userFactory := testutil.NewUserFactory()
	user := userFactory.Create(entity.PlatformiOS, true)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	subFactory := testutil.NewSubscriptionFactory()
	sub := subFactory.CreateActive(user.ID, entity.SourceIAP, entity.PlanMonthly)
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	// Setup webhook handler
	webhookHandler := handlers.NewWebhookHandler(subRepo, nil, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/webhooks/apple", webhookHandler.HandleAppleNotification)

	t.Run("POST /webhooks/apple - DID_RENEW updates subscription expiry", func(t *testing.T) {
		// Load test payload
		payload, err := os.ReadFile("../../testdata/apple_notification.json")
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/webhooks/apple", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify subscription was updated
		updatedSub, err := subRepo.GetByID(ctx, sub.ID)
		require.NoError(t, err)
		assert.True(t, updatedSub.ExpiresAt.After(sub.ExpiresAt))
	})

	t.Run("POST /webhooks/apple - DID_EXPIRE cancels subscription", func(t *testing.T) {
		// Create expired notification payload
		payload := map[string]interface{}{
			"signedPayload":      "expired_payload",
			"version":            "2.0",
			"notificationType":   "DID_EXPIRE",
			"environment":        "Sandbox",
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/webhooks/apple", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify subscription was cancelled
		updatedSub, err := subRepo.GetByID(ctx, sub.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusExpired, updatedSub.Status)
	})

	t.Run("POST /webhooks/apple - invalid signature returns 401", func(t *testing.T) {
		payload := map[string]interface{}{
			"signedPayload":      "invalid_signature",
			"version":            "2.0",
			"notificationType":   "DID_RENEW",
			"environment":        "Sandbox",
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/webhooks/apple", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGoogleWebhookHandler(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := testutil.NewMockUserRepo(dbContainer.Pool)
	subRepo := testutil.NewMockSubscriptionRepo(dbContainer.Pool)

	// Create test user and subscription
	user := userFactory.Create(entity.PlatformAndroid, true)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	sub := subFactory.CreateActive(user.ID, entity.SourceIAP, entity.PlanMonthly)
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	// Setup webhook handler
	webhookHandler := handlers.NewWebhookHandler(subRepo, nil, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/webhooks/google", webhookHandler.HandleGoogleNotification)

	t.Run("POST /webhooks/google - SUBSCRIPTION_RENEWAL updates expiry", func(t *testing.T) {
		// Create Google notification payload
		oneTimePayload := map[string]interface{}{
			"version":          "1.0",
			"packageName":      "com.app.premium",
			"eventTime":        time.Now().UnixMilli(),
			"subscriptionNotification": map[string]interface{}{
				"subscriptionToken": "sub_token_123",
				"notificationType":  3, // SUBSCRIPTION_RENEWAL
				"purchaseToken":     "purchase_token_123",
			},
		}

		payloadBytes, _ := json.Marshal(oneTimePayload)
		encodedPayload := base64.StdEncoding.EncodeToString(payloadBytes)

		payload := map[string]interface{}{
			"message": map[string]interface{}{
				"data":      encodedPayload,
				"messageId": "msg_123",
			},
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/webhooks/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST /webhooks/google - SUBSCRIPTION_EXPIRED cancels subscription", func(t *testing.T) {
		oneTimePayload := map[string]interface{}{
			"version":          "1.0",
			"packageName":      "com.app.premium",
			"eventTime":        time.Now().UnixMilli(),
			"subscriptionNotification": map[string]interface{}{
				"subscriptionToken": "sub_token_123",
				"notificationType":  4, // SUBSCRIPTION_EXPIRED
				"purchaseToken":     "purchase_token_123",
			},
		}

		payloadBytes, _ := json.Marshal(oneTimePayload)
		encodedPayload := base64.StdEncoding.EncodeToString(payloadBytes)

		payload := map[string]interface{}{
			"message": map[string]interface{}{
				"data":      encodedPayload,
				"messageId": "msg_123",
			},
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/webhooks/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify subscription was cancelled
		updatedSub, err := subRepo.GetByID(ctx, sub.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusExpired, updatedSub.Status)
	})
}
```

**Step 3: Run webhook integration tests**

Run: `cd backend && go test -tags=integration -v ./tests/integration/webhook_test.go`

Expected: Tests pass

**Step 4: Commit**

```bash
git add backend/tests/integration/webhook_test.go backend/tests/testdata/
git commit -m "test: add webhook handler integration tests"
```

---

## Task 8: Integration Tests - Database Repository Tests

**Files:**
- Create: `backend/tests/integration/repository_test.go`

**Step 1: Write repository tests**

```go
// backend/tests/integration/repository_test.go
//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestUserRepositoryIntegration(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Create real repository
	userRepo := repository.NewUserRepository(dbContainer.Pool)

	t.Run("Create and GetUserByID", func(t *testing.T) {
		user := entity.NewUser("platform-user-123", "device-123", entity.PlatformiOS, "1.0.0", "test@example.com")

		// Create
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)

		// Get by ID
		retrieved, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.PlatformUserID, retrieved.PlatformUserID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("GetByPlatformID", func(t *testing.T) {
		platformUserID := "platform-user-" + uuid.New().String()
		user := entity.NewUser(platformUserID, "device-456", entity.PlatformAndroid, "1.0.0", "test2@example.com")

		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := userRepo.GetByPlatformID(ctx, platformUserID)
		require.NoError(t, err)
		assert.Equal(t, platformUserID, retrieved.PlatformUserID)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		email := "test_" + uuid.New().String() + "@example.com"
		user := entity.NewUser("platform-user-789", "device-789", entity.PlatformiOS, "1.0.0", email)

		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := userRepo.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, email, retrieved.Email)
	})

	t.Run("UpdateUserLTV", func(t *testing.T) {
		user := entity.NewUser("platform-user-ltv", "device-ltv", entity.PlatformiOS, "1.0.0", "ltv@example.com")
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Update LTV
		updated, err := userRepo.UpdateLTV(ctx, user.ID, 99.99)
		require.NoError(t, err)
		assert.Equal(t, 99.99, updated.LTV)
		assert.NotNil(t, updated.LTVUpdatedAt)
	})

	t.Run("SoftDeleteUser", func(t *testing.T) {
		user := entity.NewUser("platform-user-delete", "device-delete", entity.PlatformiOS, "1.0.0", "delete@example.com")
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Soft delete
		err = userRepo.SoftDelete(ctx, user.ID)
		require.NoError(t, err)

		// Verify user is not found by normal queries
		_, err = userRepo.GetByID(ctx, user.ID)
		assert.Error(t, err) // Should return not found
	})
}

func TestSubscriptionRepositoryIntegration(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Create repositories
	userRepo := repository.NewUserRepository(dbContainer.Pool)
	subRepo := repository.NewSubscriptionRepository(dbContainer.Pool)

	// Create test user
	user := entity.NewUser("platform-user-sub", "device-sub", entity.PlatformiOS, "1.0.0", "sub@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("Create and GetSubscriptionByID", func(t *testing.T) {
		sub := entity.NewSubscription(user.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))

		err := subRepo.Create(ctx, sub)
		require.NoError(t, err)

		retrieved, err := subRepo.GetByID(ctx, sub.ID)
		require.NoError(t, err)
		assert.Equal(t, sub.ID, retrieved.ID)
		assert.Equal(t, entity.StatusActive, retrieved.Status)
	})

	t.Run("GetActiveByUserID", func(t *testing.T) {
		sub := entity.NewSubscription(user.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanAnnual, time.Now().Add(365*24*time.Hour))
		err := subRepo.Create(ctx, sub)
		require.NoError(t, err)

		retrieved, err := subRepo.GetActiveByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, sub.ID, retrieved.ID)
		assert.Equal(t, entity.PlanAnnual, retrieved.PlanType)
	})

	t.Run("GetAccessCheck", func(t *testing.T) {
		hasAccess, err := subRepo.CanAccess(ctx, user.ID)
		require.NoError(t, err)
		assert.True(t, hasAccess)
	})

	t.Run("UpdateSubscriptionStatus", func(t *testing.T) {
		sub := entity.NewSubscription(user.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
		err := subRepo.Create(ctx, sub)
		require.NoError(t, err)

		err = subRepo.UpdateStatus(ctx, sub.ID, entity.StatusCancelled)
		require.NoError(t, err)

		retrieved, _ := subRepo.GetByID(ctx, sub.ID)
		assert.Equal(t, entity.StatusCancelled, retrieved.Status)
	})

	t.Run("CancelSubscription", func(t *testing.T) {
		sub := entity.NewSubscription(user.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
		err := subRepo.Create(ctx, sub)
		require.NoError(t, err)

		err = subRepo.Cancel(ctx, sub.ID)
		require.NoError(t, err)

		retrieved, _ := subRepo.GetByID(ctx, sub.ID)
		assert.Equal(t, entity.StatusCancelled, retrieved.Status)
		assert.False(t, retrieved.AutoRenew)
	})
}

func TestTransactionRepositoryIntegration(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Create repositories
	userRepo := repository.NewUserRepository(dbContainer.Pool)
	subRepo := repository.NewSubscriptionRepository(dbContainer.Pool)
	txRepo := repository.NewTransactionRepository(dbContainer.Pool)

	// Create test user and subscription
	user := entity.NewUser("platform-user-tx", "device-tx", entity.PlatformiOS, "1.0.0", "tx@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	sub := entity.NewSubscription(user.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	t.Run("Create and GetTransactionByID", func(t *testing.T) {
		tx := entity.NewTransaction(user.ID, sub.ID, 9.99, "USD")
		tx.Status = entity.TransactionStatusSuccess
		tx.ReceiptHash = "sha256_test_hash"
		tx.ProviderTxID = "provider_tx_123"

		err := txRepo.Create(ctx, tx)
		require.NoError(t, err)

		retrieved, err := txRepo.GetByID(ctx, tx.ID)
		require.NoError(t, err)
		assert.Equal(t, tx.ID, retrieved.ID)
		assert.Equal(t, 9.99, retrieved.Amount)
	})

	t.Run("GetTransactionsByUserID", func(t *testing.T) {
		// Create multiple transactions
		for i := 0; i < 5; i++ {
			tx := entity.NewTransaction(user.ID, sub.ID, 9.99, "USD")
			tx.Status = entity.TransactionStatusSuccess
			tx.ReceiptHash = "sha256_hash_" + string(rune(i))
			err := txRepo.Create(ctx, tx)
			require.NoError(t, err)
		}

		txs, err := txRepo.GetByUserID(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, txs, 5)
	})

	t.Run("CheckDuplicateReceipt", func(t *testing.T) {
		receiptHash := "sha256_duplicate_test"

		// First transaction
		tx1 := entity.NewTransaction(user.ID, sub.ID, 9.99, "USD")
		tx1.ReceiptHash = receiptHash
		err := txRepo.Create(ctx, tx1)
		require.NoError(t, err)

		// Check duplicate
		isDuplicate, err := txRepo.CheckDuplicateReceipt(ctx, receiptHash)
		require.NoError(t, err)
		assert.True(t, isDuplicate)

		// Check non-existent receipt
		isDuplicate, err = txRepo.CheckDuplicateReceipt(ctx, "sha256_non_existent")
		require.NoError(t, err)
		assert.False(t, isDuplicate)
	})
}
```

**Step 2: Run repository integration tests**

Run: `cd backend && go test -tags=integration -v ./tests/integration/repository_test.go`

Expected: All tests pass

**Step 3: Commit**

```bash
git add backend/tests/integration/repository_test.go
git commit -m "test: add repository integration tests with real database"
```

---

## Task 9: Set Up E2E Test Framework

**Files:**
- Create: `backend/tests/e2e/e2e_test.go`
- Create: `backend/tests/e2e/suite.go`
- Create: `backend/tests/e2e/scenarios/user_journey_test.go`

**Step 1: Create E2E test suite**

```go
// backend/tests/e2e/suite.go
package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/bivex/paywall-iap/tests/testutil"
)

// E2ETestSuite holds the E2E test environment
type E2ETestSuite struct {
	DBContainer     *testutil.TestDBContainer
	RedisContainer  testcontainers.Container
	APIServer       *testutil.TestServer
	BaseURL         string
	HTTPClient      *http.Client
}

// SetupE2EEnvironment starts all required containers and services
func SetupE2ETestSuite(ctx context.Context, t *testing.T) *E2ETestSuite {
	t.Helper()

	suite := &E2ETestSuite{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Start PostgreSQL
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	suite.DBContainer = dbContainer

	// Start Redis
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections").WithOccurrence(1),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	require.NoError(t, err)
	suite.RedisContainer = redisContainer

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbContainer.Pool)
	subRepo := repository.NewSubscriptionRepository(dbContainer.Pool)

	// Start API server (in real E2E, this would be a separate process)
	// For now, we use the test server
	suite.APIServer = testutil.NewTestServer(ctx, dbContainer.Pool, userRepo, subRepo)
	suite.BaseURL = suite.APIServer.BaseURL()

	return suite
}

// TeardownE2EEnvironment cleans up all containers and services
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
```

**Step 2: Create E2E test entry point**

```go
// backend/tests/e2e/e2e_test.go
//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/bivex/paywall-iap/tests/e2e/scenarios"
)

var testSuite *E2ETestSuite

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup
	// Note: In a real E2E test, this would be done in a separate setup script
	// and the tests would run against a deployed environment

	// Run tests
	code := m.Run()

	// Exit
	os.Exit(code)
}
```

**Step 3: Create user journey scenario**

```go
// backend/tests/e2e/scenarios/user_journey_test.go
//go:build e2e

package scenarios

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/tests/e2e"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestCompleteUserJourney(t *testing.T) {
	ctx := context.Background()

	// Setup E2E environment
	suite := e2e.SetupE2ETestSuite(ctx, t)
	defer suite.Teardown(ctx, t)

	t.Run("Complete User Journey - Registration to Subscription", func(t *testing.T) {
		// Step 1: Register new user
		t.Logf("Step 1: Registering new user")
		registerReq := map[string]interface{}{
			"platform_user_id": "e2e-user-" + time.Now().Format("20060102150405"),
			"device_id":        "e2e-device-" + time.Now().Format("20060102150405"),
			"platform":         "ios",
			"app_version":      "1.0.0",
			"email":            "e2e_" + time.Now().Format("20060102150405") + "@example.com",
		}

		req, _ := testutil.NewTestRequest("POST", suite.GetAPIURL()+"/v1/auth/register", registerReq, "")
		resp, body, err := testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var authResp testutil.AuthResponse
		json.Unmarshal(body, &authResp)
		require.NotEmpty(t, authResp.Data.UserID)
		require.NotEmpty(t, authResp.Data.AccessToken)

		userID := authResp.Data.UserID
		accessToken := authResp.Data.AccessToken
		t.Logf("User registered: %s", userID)

		// Step 2: Check access (should be false)
		t.Logf("Step 2: Checking access before subscription")
		req, _ = testutil.NewTestRequest("GET", suite.GetAPIURL()+"/v1/subscription/access", nil, accessToken)
		resp, body, err = testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accessResp testutil.AccessCheckResponse
		json.Unmarshal(body, &accessResp)
		assert.False(t, accessResp.Data.HasAccess)
		t.Logf("Access check: has_access=%v", accessResp.Data.HasAccess)

		// Step 3: Verify IAP receipt (simulating purchase)
		t.Logf("Step 3: Verifying IAP receipt")
		verifyReq := map[string]interface{}{
			"platform":       "ios",
			"receipt_data":   "e2e_test_receipt_" + time.Now().Format("20060102150405"),
			"product_id":     "com.app.premium",
			"transaction_id": "tx_" + time.Now().Format("20060102150405"),
		}

		req, _ = testutil.NewTestRequest("POST", suite.GetAPIURL()+"/v1/verify/iap", verifyReq, accessToken)
		resp, body, err = testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		t.Logf("IAP verified successfully")

		// Step 4: Check access (should be true now)
		t.Logf("Step 4: Checking access after subscription")
		req, _ = testutil.NewTestRequest("GET", suite.GetAPIURL()+"/v1/subscription/access", nil, accessToken)
		resp, body, err = testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		json.Unmarshal(body, &accessResp)
		assert.True(t, accessResp.Data.HasAccess)
		t.Logf("Access check: has_access=%v", accessResp.Data.HasAccess)

		// Step 5: Get subscription details
		t.Logf("Step 5: Getting subscription details")
		req, _ = testutil.NewTestRequest("GET", suite.GetAPIURL()+"/v1/subscription", nil, accessToken)
		resp, body, err = testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var subResp testutil.SubscriptionResponse
		json.Unmarshal(body, &subResp)
		assert.Equal(t, "active", subResp.Data.Status)
		assert.Equal(t, "monthly", subResp.Data.PlanType)
		t.Logf("Subscription: status=%s, plan=%s", subResp.Data.Status, subResp.Data.PlanType)

		// Step 6: Cancel subscription
		t.Logf("Step 6: Cancelling subscription")
		req, _ = testutil.NewTestRequest("DELETE", suite.GetAPIURL()+"/v1/subscription", nil, accessToken)
		resp, _, err = testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		t.Logf("Subscription cancelled")

		// Step 7: Check access (should be false after cancellation)
		t.Logf("Step 7: Checking access after cancellation")
		req, _ = testutil.NewTestRequest("GET", suite.GetAPIURL()+"/v1/subscription/access", nil, accessToken)
		resp, body, err = testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)

		json.Unmarshal(body, &accessResp)
		assert.False(t, accessResp.Data.HasAccess)
		t.Logf("Access check after cancellation: has_access=%v", accessResp.Data.HasAccess)
	})
}
```

**Step 4: Run E2E tests**

Run: `cd backend && go test -tags=e2e -v ./tests/e2e/...`

Expected: E2E user journey test passes

**Step 5: Commit**

```bash
git add backend/tests/e2e/
git commit -m "test: add E2E test framework and user journey scenario"
```

---

## Task 10: Set Up Load Testing Infrastructure with k6

**Files:**
- Create: `backend/tests/load/basic_load_test.js`
- Create: `backend/tests/load/stress_test.js`
- Create: `backend/tests/load/soak_test.js`
- Create: `backend/tests/load/config.js`

**Step 1: Create k6 configuration**

```javascript
// backend/tests/load/config.js
export const config = {
  // Target API URL (set via environment variable)
  baseURL: __ENV.BASE_URL || 'http://localhost:8080',

  // Test credentials (should be set via environment)
  testUserEmail: __ENV.TEST_USER_EMAIL || 'loadtest@example.com',
  testUserPassword: __ENV.TEST_USER_PASSWORD || 'loadtest123',

  // Common headers
  headers: {
    'Content-Type': 'application/json',
    'User-Agent': 'k6-load-test',
  },
};

// Helper function to get random item from array
export function randomItem(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

// Helper function to generate random string
export function randomString(length) {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}
```

**Step 2: Create basic load test**

```javascript
// backend/tests/load/basic_load_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { config, randomString } from './config.js';

// Custom metrics
const errorRate = new Rate('errors');
const loginTime = new Trend('login_time');
const subscriptionCheckTime = new Trend('subscription_check_time');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 users
    { duration: '1m', target: 10 },    // Stay at 10 users
    { duration: '30s', target: 20 },   // Ramp up to 20 users
    { duration: '1m', target: 20 },    // Stay at 20 users
    { duration: '30s', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% of requests should be below 500ms
    errors: ['rate<0.01'],             // Error rate should be less than 1%
  },
};

export default function () {
  const timestamp = new Date().getTime();

  // Scenario 1: Register new user
  const registerPayload = {
    platform_user_id: `loadtest-${timestamp}-${randomString(8)}`,
    device_id: `device-${randomString(8)}`,
    platform: 'ios',
    app_version: '1.0.0',
    email: `loadtest+${timestamp}@example.com`,
  };

  const registerRes = http.post(
    `${config.baseURL}/v1/auth/register`,
    JSON.stringify(registerPayload),
    { headers: config.headers }
  );

  const registerOk = check(registerRes, {
    'register status is 201': (r) => r.status === 201,
    'register has user_id': (r) => {
      const body = r.json();
      return body.data && body.data.user_id;
    },
  });

  errorRate.add(registerOk ? 0 : 1);

  if (!registerOk) {
    sleep(1);
    return;
  }

  const authData = registerRes.json().data;
  const accessToken = authData.access_token;

  sleep(1);

  // Scenario 2: Check subscription access
  const accessRes = http.get(
    `${config.baseURL}/v1/subscription/access`,
    {
      headers: {
        ...config.headers,
        'Authorization': `Bearer ${accessToken}`,
      },
    }
  );

  const accessOk = check(accessRes, {
    'access check status is 200': (r) => r.status === 200,
    'access check has has_access': (r) => {
      const body = r.json();
      return body.data && typeof body.data.has_access === 'boolean';
    },
  });

  errorRate.add(accessOk ? 0 : 1);
  subscriptionCheckTime.add(accessRes.timings.duration);

  sleep(1);

  // Scenario 3: Get subscription details
  const subRes = http.get(
    `${config.baseURL}/v1/subscription`,
    {
      headers: {
        ...config.headers,
        'Authorization': `Bearer ${accessToken}`,
      },
    }
  );

  check(subRes, {
    'subscription status is 200 or 404': (r) => r.status === 200 || r.status === 404,
  });

  sleep(2);
}
```

**Step 3: Create stress test**

```javascript
// backend/tests/load/stress_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { config, randomString } from './config.js';

const errorRate = new Rate('errors');

// Stress test configuration - push system to breaking point
export const options = {
  stages: [
    { duration: '10s', target: 50 },   // Rapid ramp up to 50 users
    { duration: '30s', target: 50 },   // Stay at 50 users
    { duration: '10s', target: 100 },  // Ramp up to 100 users
    { duration: '30s', target: 100 },  // Stay at 100 users
    { duration: '10s', target: 200 },  // Ramp up to 200 users
    { duration: '1m', target: 200 },   // Stay at 200 users (stress)
    { duration: '10s', target: 0 },    // Rapid ramp down
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],    // 5% error rate threshold
    http_req_duration: ['p(95)<1000'], // 95% under 1s
  },
};

export default function () {
  // Heavy scenario: Register + Verify IAP + Check Access
  const timestamp = new Date().getTime();

  // Register
  const registerPayload = {
    platform_user_id: `stresstest-${timestamp}`,
    device_id: `device-${randomString(8)}`,
    platform: 'ios',
    app_version: '1.0.0',
    email: `stresstest+${timestamp}@example.com`,
  };

  const registerRes = http.post(
    `${config.baseURL}/v1/auth/register`,
    JSON.stringify(registerPayload),
    { headers: config.headers }
  );

  if (registerRes.status !== 201) {
    errorRate.add(1);
    sleep(0.5);
    return;
  }

  const accessToken = registerRes.json().data.access_token;
  errorRate.add(0);

  sleep(0.5);

  // Verify IAP (simulated)
  const verifyPayload = {
    platform: 'ios',
    receipt_data: `receipt-${randomString(16)}`,
    product_id: 'com.app.premium',
    transaction_id: `tx-${randomString(8)}`,
  };

  const verifyRes = http.post(
    `${config.baseURL}/v1/verify/iap`,
    JSON.stringify(verifyPayload),
    {
      headers: {
        ...config.headers,
        'Authorization': `Bearer ${accessToken}`,
      },
    }
  );

  check(verifyRes, {
    'verify status is 200 or 422': (r) => r.status === 200 || r.status === 422,
  });

  sleep(0.5);

  // Check access
  http.get(
    `${config.baseURL}/v1/subscription/access`,
    {
      headers: {
        ...config.headers,
        'Authorization': `Bearer ${accessToken}`,
      },
    }
  );

  sleep(1);
}
```

**Step 4: Create soak test**

```javascript
// backend/tests/load/soak_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { config } from './config.js';

const errorRate = new Rate('errors');
const memoryLeakCheck = new Trend('response_size');

// Soak test - long running test to detect memory leaks, degradation
export const options = {
  stages: [
    { duration: '5m', target: 20 },  // Ramp up to 20 users
    { duration: '30m', target: 20 }, // Stay at 20 users for 30 minutes
    { duration: '5m', target: 0 },   // Ramp down
  ],
  thresholds: {
    errors: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

export default function () {
  // Simulate typical user behavior over time
  const scenarios = [
    'check_access',
    'get_subscription',
    'refresh_token',
  ];

  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];

  // Use existing test user credentials (set via environment)
  const accessToken = __ENV.TEST_ACCESS_TOKEN;

  if (!accessToken) {
    console.log('No access token, skipping');
    sleep(1);
    return;
  }

  let res;

  if (scenario === 'check_access') {
    res = http.get(
      `${config.baseURL}/v1/subscription/access`,
      {
        headers: {
          ...config.headers,
          'Authorization': `Bearer ${accessToken}`,
        },
      }
    );

    const ok = check(res, {
      'access check status is 200': (r) => r.status === 200,
    });
    errorRate.add(ok ? 0 : 1);
    memoryLeakCheck.add(res.body.length);

  } else if (scenario === 'get_subscription') {
    res = http.get(
      `${config.baseURL}/v1/subscription`,
      {
        headers: {
          ...config.headers,
          'Authorization': `Bearer ${accessToken}`,
        },
      }
    );

    check(res, {
      'subscription status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    });
    memoryLeakCheck.add(res.body.length);
  }

  sleep(2);
}
```

**Step 5: Create Makefile targets for load tests**

```makefile
# backend/Makefile - Add these load test targets

# Load test targets
load-test:
	@echo "Running basic load test..."
	k6 run tests/load/basic_load_test.js

load-test-stress:
	@echo "Running stress test..."
	k6 run tests/load/stress_test.js

load-test-soak:
	@echo "Running soak test (30 minutes)..."
	k6 run tests/load/soak_test.js

load-test-all: load-test load-test-stress load-test-soak
	@echo "All load tests completed"

load-test-with-coverage:
	@echo "Running load test with Prometheus metrics..."
	K6_PROMETHEUS_REMOTE_URL=http://localhost:9090/api/v1/write \
	k6 run --out prometheus tests/load/basic_load_test.js
```

**Step 6: Verify k6 installation**

Run: `k6 version`

Expected: k6 version output (e.g., `k6 v0.47.0`)

If not installed:
```bash
# macOS
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Step 7: Commit**

```bash
git add backend/tests/load/ backend/Makefile
git commit -m "test: add k6 load testing infrastructure with basic, stress, and soak tests"
```

---

## Task 11: Create Load Test Reports and Dashboards

**Files:**
- Create: `backend/tests/load/dashboard.json` (Grafana dashboard)
- Create: `docs/testing/load-testing-guide.md`

**Step 1: Create Grafana dashboard**

```json
// backend/tests/load/dashboard.json
{
  "dashboard": {
    "title": "IAP System Load Test Dashboard",
    "tags": ["load-test", "k6"],
    "timezone": "browser",
    "panels": [
      {
        "title": "Requests per Second",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(k6_http_reqs_total[1m])",
            "legendFormat": "RPS"
          }
        ]
      },
      {
        "title": "Response Time (95th percentile)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(k6_http_req_duration_bucket[1m]))",
            "legendFormat": "p95"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(k6_http_req_failed_total[1m])",
            "legendFormat": "Error Rate"
          }
        ]
      },
      {
        "title": "Active VUs",
        "type": "graph",
        "targets": [
          {
            "expr": "k6_vus",
            "legendFormat": "Virtual Users"
          }
        ]
      }
    ]
  }
}
```

**Step 2: Create load testing guide**

```markdown
# Load Testing Guide

## Overview

This guide covers load testing procedures for the IAP system using k6.

## Prerequisites

- k6 installed (`brew install k6`)
- Test environment running (Docker Compose)
- Test user credentials configured

## Quick Start

```bash
# Start test environment
cd infra/docker-compose
docker compose -f docker-compose.test.yml up -d

# Run basic load test
cd backend
make load-test

# Run stress test
make load-test-stress

# Run soak test (30 minutes)
make load-test-soak
```

## Test Scenarios

### Basic Load Test
- **Purpose**: Establish baseline performance metrics
- **Duration**: ~3 minutes
- **Peak Users**: 20 concurrent users
- **Thresholds**:
  - 95% of requests < 500ms
  - Error rate < 1%

### Stress Test
- **Purpose**: Find breaking point and system limits
- **Duration**: ~2 minutes
- **Peak Users**: 200 concurrent users
- **Thresholds**:
  - Error rate < 5%
  - 95% of requests < 1s

### Soak Test
- **Purpose**: Detect memory leaks and performance degradation
- **Duration**: ~40 minutes
- **Users**: 20 concurrent users (sustained)
- **Thresholds**:
  - Error rate < 1%
  - No performance degradation over time

## Interpreting Results

### Key Metrics

1. **HTTP Request Duration**
   - p(95) < 500ms: Good
   - p(95) 500-1000ms: Acceptable
   - p(95) > 1000ms: Needs optimization

2. **Error Rate**
   - < 1%: Excellent
   - 1-5%: Acceptable under stress
   - > 5%: System overloaded

3. **Throughput**
   - Monitor requests/second
   - Should scale linearly with users until saturation

### Common Issues

**High Response Times**
- Check database connection pool size
- Monitor Redis latency
- Review slow query logs

**High Error Rates**
- Check database connection limits
- Monitor memory usage
- Review application logs

**Performance Degradation (Soak Test)**
- Monitor memory usage over time
- Check for connection leaks
- Review garbage collection metrics

## Performance Benchmarks

### Target Metrics (Production)

| Metric | Target | Warning | Critical |
|--------|--------|---------|----------|
| p95 Response Time | < 300ms | 300-500ms | > 500ms |
| Error Rate | < 0.1% | 0.1-1% | > 1% |
| Throughput | > 1000 RPS | 500-1000 RPS | < 500 RPS |
| Concurrent Users | > 1000 | 500-1000 | < 500 |

## CI/CD Integration

Load tests run automatically on:
- Pull requests to `main` (basic load test only)
- Nightly builds (all load tests)
- Before production deployments (stress test)

## Troubleshooting

### Test Fails with "Connection Refused"
```bash
# Check if test environment is running
docker compose -f docker-compose.test.yml ps

# Restart if needed
docker compose -f docker-compose.test.yml down
docker compose -f docker-compose.test.yml up -d
```

### k6 Installation Issues
```bash
# Verify installation
k6 version

# Reinstall if needed
brew reinstall k6
```

## Reporting

After each load test run:
1. Export results to InfluxDB/Prometheus
2. Generate Grafana dashboard snapshot
3. Compare with baseline metrics
4. Document any regressions

## Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 Best Practices](https://k6.io/docs/best-practices/)
- [Grafana Dashboard Guide](https://grafana.com/docs/)
```

**Step 3: Commit**

```bash
git add backend/tests/load/dashboard.json docs/testing/load-testing-guide.md
git commit -m "docs: add load testing guide and Grafana dashboard configuration"
```

---

## Task 12: Create CI/CD Workflows for Automated Testing

**Files:**
- Create: `.github/workflows/integration-tests.yml`
- Create: `.github/workflows/load-tests.yml`
- Modify: `.github/workflows/backend-ci.yml`

**Step 1: Create integration test workflow**

```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests

on:
  push:
    branches: [main, develop]
    paths:
      - 'backend/**'
  pull_request:
    branches: [main, develop]
    paths:
      - 'backend/**'

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_USER: test
          POSTGRES_DB: iap_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Install dependencies
        working-directory: ./backend
        run: go mod download

      - name: Run migrations
        working-directory: ./backend
        run: |
          go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
          migrate -path migrations -database "postgres://test:test@localhost:5432/iap_test?sslmode=disable" up

      - name: Run integration tests
        working-directory: ./backend
        run: go test -tags=integration -race -count=1 ./tests/integration/...

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: integration-test-results
          path: backend/test-results.xml
```

**Step 2: Create E2E test workflow**

```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_USER: test
          POSTGRES_DB: iap_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Install dependencies
        working-directory: ./backend
        run: go mod download

      - name: Run migrations
        working-directory: ./backend
        run: |
          go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
          migrate -path migrations -database "postgres://test:test@localhost:5432/iap_test?sslmode=disable" up

      - name: Run E2E tests
        working-directory: ./backend
        run: go test -tags=e2e -race -count=1 ./tests/e2e/...

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-test-results
          path: backend/e2e-results.xml
```

**Step 3: Create load test workflow**

```yaml
# .github/workflows/load-tests.yml
name: Load Tests

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC
  workflow_dispatch:
    inputs:
      test_type:
        description: 'Type of load test'
        required: true
        default: 'basic'
        type: choice
        options:
          - basic
          - stress
          - soak

jobs:
  load-tests:
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule' || github.event_name == 'workflow_dispatch'

    steps:
      - uses: actions/checkout@v4

      - name: Install k6
        run: |
          sudo gpg -k
          sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6

      - name: Start test environment
        working-directory: ./infra/docker-compose
        run: docker compose -f docker-compose.test.yml up -d

      - name: Wait for services
        run: sleep 30

      - name: Run basic load test
        if: github.event.inputs.test_type == 'basic' || github.event_name == 'schedule'
        working-directory: ./backend
        run: k6 run tests/load/basic_load_test.js
        env:
          BASE_URL: http://localhost:8080

      - name: Run stress test
        if: github.event.inputs.test_type == 'stress'
        working-directory: ./backend
        run: k6 run tests/load/stress_test.js
        env:
          BASE_URL: http://localhost:8080

      - name: Upload load test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: load-test-results
          path: backend/load-results.json

      - name: Cleanup
        if: always()
        working-directory: ./infra/docker-compose
        run: docker compose -f docker-compose.test.yml down
```

**Step 4: Update backend CI workflow**

```yaml
# .github/workflows/backend-ci.yml - Update to include integration tests

name: Backend CI

on:
  push:
    branches: [main, develop]
    paths:
      - 'backend/**'
      - '.github/workflows/backend-ci.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'backend/**'

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          working-directory: backend

  test-unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run unit tests
        working-directory: ./backend
        run: make test-unit

  test-integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_USER: test
          POSTGRES_DB: iap_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run migrations
        working-directory: ./backend
        run: |
          go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
          migrate -path migrations -database "postgres://test:test@localhost:5432/iap_test?sslmode=disable" up
      - name: Run integration tests
        working-directory: ./backend
        run: make test-integration

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./backend/...'
      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: gosec-results.sarif
```

**Step 5: Commit**

```bash
git add .github/workflows/integration-tests.yml .github/workflows/e2e-tests.yml .github/workflows/load-tests.yml .github/workflows/backend-ci.yml
git commit -m "ci: add automated integration, E2E, and load test workflows"
```

---

## Task 13: Create Test Documentation and Runbooks

**Files:**
- Create: `docs/testing/testing-strategy.md`
- Create: `docs/testing/e2e-testing-guide.md`
- Create: `docs/runbooks/test-failure-troubleshooting.md`

**Step 1: Create testing strategy document**

```markdown
# Testing Strategy

## Overview

The IAP system uses a three-tier testing strategy:

1. **Unit Tests** - Test individual components in isolation
2. **Integration Tests** - Test component interactions with real database
3. **E2E Tests** - Test complete user journeys
4. **Load Tests** - Test system performance under load

## Test Pyramid

```
        /\
       /  \      E2E Tests (Few, slow, high confidence)
      /----\
     /      \    Integration Tests (Some, medium speed)
    /--------\
   /          \  Unit Tests (Many, fast, isolated)
  /------------\
```

## Unit Tests

**Location**: `backend/internal/**/*_test.go`

**Purpose**: Test individual functions, methods, and components in isolation.

**Characteristics**:
- Fast execution (< 100ms per test)
- No external dependencies
- Use mocks for dependencies
- High coverage target (> 80%)

**Example**:
```go
func TestUserEntity(t *testing.T) {
    user := entity.NewUser("platform-id", "device-id", entity.PlatformiOS, "1.0", "test@example.com")
    assert.False(t, user.IsDeleted())
    assert.True(t, user.HasEmail())
}
```

**Run**: `make test-unit`

## Integration Tests

**Location**: `backend/tests/integration/*_test.go`

**Purpose**: Test component interactions with real database and external services.

**Characteristics**:
- Medium execution (< 5s per test)
- Real PostgreSQL database (testcontainers)
- Mock external services
- Test database queries and handlers

**Example**:
```go
func TestSubscriptionEndpoints(t *testing.T) {
    // Setup real database
    dbContainer := SetupTestDB()
    defer dbContainer.Teardown()
    
    // Make HTTP requests
    resp := httptest.NewRequest("GET", "/v1/subscription", nil)
    assert.Equal(t, http.StatusOK, resp.Code)
}
```

**Run**: `make test-integration`

## E2E Tests

**Location**: `backend/tests/e2e/*_test.go`

**Purpose**: Test complete user journeys from start to finish.

**Characteristics**:
- Slow execution (< 30s per scenario)
- Full system stack
- Real database, cache, and external services (mocked)
- Test user workflows

**Example**:
```go
func TestCompleteUserJourney(t *testing.T) {
    // 1. Register user
    // 2. Verify IAP receipt
    // 3. Check access
    // 4. Cancel subscription
    // 5. Verify access revoked
}
```

**Run**: `make test-e2e`

## Load Tests

**Location**: `backend/tests/load/*.js`

**Purpose**: Test system performance and identify bottlenecks.

**Characteristics**:
- k6 JavaScript tests
- Simulate concurrent users
- Measure response times, throughput, error rates
- Identify breaking points

**Scenarios**:
- Basic load test (20 users)
- Stress test (200 users)
- Soak test (20 users, 30 minutes)

**Run**: `make load-test`

## Test Data Management

### Test Factories

Use factories to create test data:
```go
userFactory := testutil.NewUserFactory()
user := userFactory.Create(entity.PlatformiOS, true)

subFactory := testutil.NewSubscriptionFactory()
sub := subFactory.CreateActive(user.ID, entity.SourceIAP, entity.PlanMonthly)
```

### Test Fixtures

Store test data in `backend/tests/testdata/`:
- JSON payloads
- Webhook notifications
- Expected responses

## CI/CD Integration

| Test Type | PR | Main Branch | Nightly |
|-----------|-----|-------------|---------|
| Unit      | ✓   | ✓           | ✓       |
| Integration | ✓ | ✓           | ✓       |
| E2E       | ✓   | ✓           | ✓       |
| Load (Basic) |   | ✓           | ✓       |
| Load (Stress) |   |             | ✓       |
| Load (Soak) |   |             | Weekly  |

## Coverage Targets

| Component | Target |
|-----------|--------|
| Domain Layer | > 90% |
| Application Layer | > 85% |
| Infrastructure Layer | > 80% |
| Interfaces Layer | > 75% |
| Overall | > 80% |

## Test Quality Guidelines

1. **Test Behavior, Not Implementation**
   - Test what the code does, not how it does it
   - Avoid testing private methods directly

2. **Use Descriptive Test Names**
   - `TestUserRegistration_Success` not `TestRegister1`
   - Include scenario in name: `TestSubscription_AccessCheck_Expired`

3. **Arrange-Act-Assert Pattern**
   ```go
   // Arrange
   user := createUser()
   
   // Act
   result := service.GetUser(user.ID)
   
   // Assert
   assert.Equal(t, user.ID, result.ID)
   ```

4. **Clean Up Resources**
   - Use `defer` for cleanup
   - Close database connections
   - Terminate containers

5. **Parallel Tests When Possible**
   ```go
   t.Parallel()
   ```

## Troubleshooting

See: [Test Failure Troubleshooting Runbook](docs/runbooks/test-failure-troubleshooting.md)
```

**Step 2: Create E2E testing guide**

```markdown
# E2E Testing Guide

## Overview

End-to-End (E2E) tests verify complete user workflows by testing the entire system stack.

## Prerequisites

- Docker and Docker Compose
- Go 1.21+
- k6 (for load testing)

## Quick Start

```bash
# Run all E2E tests
cd backend
make test-e2e

# Run specific E2E test
go test -tags=e2e -v ./tests/e2e/... -run TestCompleteUserJourney
```

## E2E Test Structure

### Test File Organization

```
backend/tests/e2e/
├── e2e_test.go          # Test entry point
├── suite.go             # Test suite setup/teardown
└── scenarios/
    ├── user_journey_test.go
    ├── subscription_flow_test.go
    └── webhook_flow_test.go
```

### Test Suite Pattern

```go
func TestCompleteUserJourney(t *testing.T) {
    ctx := context.Background()
    
    // Setup
    suite := e2e.SetupE2ETestSuite(ctx, t)
    defer suite.Teardown(ctx, t)
    
    // Test steps
    t.Run("Step 1: Register user", func(t *testing.T) {
        // ...
    })
    
    t.Run("Step 2: Verify IAP", func(t *testing.T) {
        // ...
    })
}
```

## Common Scenarios

### User Registration Flow

1. POST `/v1/auth/register`
2. Verify response contains user_id and tokens
3. Validate user created in database

### Subscription Purchase Flow

1. Register user
2. POST `/v1/verify/iap` with receipt
3. Verify subscription created
4. GET `/v1/subscription/access` returns true

### Subscription Cancellation Flow

1. Create user with active subscription
2. DELETE `/v1/subscription`
3. Verify subscription status = cancelled
4. GET `/v1/subscription/access` returns false

### Webhook Processing Flow

1. Create user with subscription
2. POST webhook notification
3. Verify subscription updated
4. Check access status changed

## Best Practices

### 1. Use Test Factories

```go
userFactory := testutil.NewUserFactory()
user := userFactory.Create(entity.PlatformiOS, true)
```

### 2. Clean Test Data

```go
defer func() {
    // Cleanup test data
    db.Exec("DELETE FROM users WHERE email LIKE 'e2e_%'")
}()
```

### 3. Use Descriptive Logging

```go
t.Logf("Step 1: Registering user %s", email)
t.Logf("Step 2: Verifying IAP receipt")
```

### 4. Handle Timing Issues

```go
// Wait for async processing
time.Sleep(100 * time.Millisecond)

// Or use polling
require.Eventually(t, func() bool {
    sub, _ := subRepo.GetByID(ctx, subID)
    return sub.Status == entity.StatusActive
}, 5*time.Second, 100*time.Millisecond)
```

### 5. Test Error Scenarios

```go
t.Run("Invalid receipt returns error", func(t *testing.T) {
    req := map[string]interface{}{
        "receipt_data": "invalid",
    }
    resp := post("/v1/verify/iap", req)
    assert.Equal(t, 422, resp.StatusCode)
})
```

## Debugging E2E Tests

### Enable Verbose Logging

```bash
go test -tags=e2e -v ./tests/e2e/...
```

### Keep Test Environment Running

```bash
# Don't teardown after test
export KEEP_ENV=true
go test -tags=e2e ./tests/e2e/...
```

### Inspect Database

```bash
# Connect to test database
docker exec -it postgres psql -U test -d iap_test
```

## Common Issues

### Test Fails with "Connection Refused"

**Cause**: Database container not ready

**Fix**:
```bash
docker compose -f docker-compose.test.yml down
docker compose -f docker-compose.test.yml up -d
sleep 10
```

### Test Times Out

**Cause**: Slow test or deadlock

**Fix**:
```bash
# Increase timeout
go test -tags=e2e -timeout 5m ./tests/e2e/...
```

### Flaky Tests

**Cause**: Race conditions or timing issues

**Fix**:
- Use `require.Eventually` instead of `time.Sleep`
- Add proper synchronization
- Check test isolation

## Adding New E2E Tests

1. Create new test file in `tests/e2e/scenarios/`
2. Follow existing test pattern
3. Use test factories for data creation
4. Add descriptive logging
5. Run test locally before committing
6. Update this guide with new scenario

## Performance Guidelines

- E2E tests should complete in < 30 seconds
- Each scenario should test one workflow
- Avoid unnecessary sleeps
- Clean up test data after each test
```

**Step 3: Create troubleshooting runbook**

```markdown
# Test Failure Troubleshooting Runbook

## Quick Diagnosis

### 1. Check Test Output

```bash
go test -tags=integration -v ./tests/integration/... 2>&1 | tail -50
```

Look for:
- Error messages
- Stack traces
- Failed assertions

### 2. Identify Failure Type

| Symptom | Likely Cause | Next Step |
|---------|--------------|-----------|
| "connection refused" | Database not running | Check containers |
| "migration failed" | Schema issue | Check migrations |
| "timeout" | Slow test or deadlock | Increase timeout |
| "assertion failed" | Logic error | Review test code |
| "race detected" | Concurrent access | Fix race condition |

## Database Issues

### Container Won't Start

```bash
# Check container status
docker compose -f docker-compose.test.yml ps

# View logs
docker compose -f docker-compose.test.yml logs postgres

# Restart
docker compose -f docker-compose.test.yml restart postgres
```

### Migration Failures

```bash
# Check migration status
migrate -path backend/migrations -database "postgres://test:test@localhost:5432/iap_test?sslmode=disable" version

# Rollback
migrate -path backend/migrations -database "postgres://test:test@localhost:5432/iap_test?sslmode=disable" down 1

# Re-apply
migrate -path backend/migrations -database "postgres://test:test@localhost:5432/iap_test?sslmode=disable" up
```

### Connection Pool Exhausted

**Symptoms**: "too many connections", "context deadline exceeded"

**Fix**:
```bash
# Check active connections
docker exec -it postgres psql -U test -d iap_test -c "SELECT count(*) FROM pg_stat_activity;"

# Restart database
docker compose -f docker-compose.test.yml restart postgres
```

## Test Code Issues

### Flaky Tests

**Identify**:
```bash
# Run test 10 times
for i in {1..10}; do
  go test -tags=integration -run TestName ./tests/integration/...
done
```

**Fix**:
- Replace `time.Sleep` with `require.Eventually`
- Add proper synchronization
- Ensure test isolation

### Race Conditions

**Detect**:
```bash
go test -tags=integration -race ./tests/integration/...
```

**Fix**:
- Use mutex for shared state
- Avoid global variables
- Use channels for communication

### Slow Tests

**Profile**:
```bash
go test -tags=integration -cpuprofile=cpu.prof -memprofile=mem.prof ./tests/integration/...
go tool pprof cpu.prof
```

**Optimize**:
- Reduce test data setup
- Parallelize independent tests
- Cache expensive setup

## CI/CD Failures

### GitHub Actions Timeout

**Check**:
- Workflow run logs
- Service health checks
- Resource limits

**Fix**:
```yaml
# Increase timeout
services:
  postgres:
    options: >-
      --health-cmd pg_isready
      --health-interval 5s  # Reduce from 10s
```

### Out of Disk Space

**Symptoms**: "no space left on device"

**Fix**:
```bash
# Clean Docker
docker system prune -af

# Remove old images
docker image prune -af
```

## Load Test Issues

### k6 Test Fails

**Check**:
```bash
k6 run tests/load/basic_load_test.js --verbose
```

**Common Issues**:
- Target service not running
- Incorrect BASE_URL
- Insufficient system resources

### High Error Rate

**Investigate**:
1. Check application logs
2. Monitor database connections
3. Review memory usage
4. Check rate limiting

**Fix**:
- Increase connection pool size
- Optimize slow queries
- Add caching

## Getting Help

### Logs

- Application: `docker compose logs api`
- Database: `docker compose logs postgres`
- Redis: `docker compose logs redis`

### Metrics

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001

### Contact

- Backend Team: #backend-channel
- DevOps Team: #devops-channel
```

**Step 4: Commit**

```bash
git add docs/testing/ docs/runbooks/test-failure-troubleshooting.md
git commit -m "docs: add comprehensive testing documentation and troubleshooting runbooks"
```

---

## Task 14: Run Full Test Suite and Establish Baseline

**Files:**
- Create: `backend/tests/TEST_RESULTS.md`

**Step 1: Run all tests**

```bash
# Navigate to backend directory
cd backend

# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Run E2E tests
make test-e2e

# Run basic load test
make load-test
```

**Step 2: Generate coverage report**

```bash
# Generate coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html

# View coverage by package
go tool cover -func=coverage.out | grep total
```

**Step 3: Document baseline results**

```markdown
# Test Results Baseline

**Date**: 2026-02-28
**Commit**: [latest commit hash]

## Unit Tests

```
PASS
ok      github.com/bivex/paywall-iap/internal/domain/entity     0.015s
ok      github.com/bivex/paywall-iap/internal/domain/valueobject 0.012s
ok      github.com/bivex/paywall-iap/internal/application/command 0.025s
ok      github.com/bivex/paywall-iap/internal/application/query 0.018s
```

**Total**: XX tests, XX passed, 0 failed
**Coverage**: XX%

## Integration Tests

```
PASS
ok      github.com/bivex/paywall-iap/tests/integration  5.234s
```

**Total**: XX tests, XX passed, 0 failed
**Duration**: ~5 seconds

## E2E Tests

```
PASS
ok      github.com/bivex/paywall-iap/tests/e2e  12.456s
```

**Total**: X scenarios, X passed, 0 failed
**Duration**: ~12 seconds

## Load Tests

### Basic Load Test

- **Duration**: 3 minutes
- **Peak Users**: 20
- **Requests**: XX,XXX
- **p95 Response Time**: XXXms
- **Error Rate**: X.XX%

### Stress Test

- **Duration**: 2 minutes
- **Peak Users**: 200
- **Requests**: XX,XXX
- **p95 Response Time**: XXXms
- **Error Rate**: X.XX%

## Coverage Summary

| Package | Coverage |
|---------|----------|
| Domain Layer | XX.X% |
| Application Layer | XX.X% |
| Infrastructure Layer | XX.X% |
| Interfaces Layer | XX.X% |
| **Total** | **XX.X%** |

## Performance Benchmarks

| Endpoint | p50 | p95 | p99 |
|----------|-----|-----|-----|
| POST /auth/register | XXms | XXms | XXms |
| GET /subscription | XXms | XXms | XXms |
| GET /subscription/access | XXms | XXms | XXms |
| POST /verify/iap | XXms | XXms | XXms |

## Issues Found

None

## Recommendations

1. [ ] Add more E2E scenarios for webhook flows
2. [ ] Increase load test duration for soak test
3. [ ] Add performance regression thresholds to CI
```

**Step 4: Commit**

```bash
git add backend/tests/TEST_RESULTS.md
git commit -m "test: establish baseline test results and performance benchmarks"
```

---

## Task 15: Create Test Maintenance Guidelines

**Files:**
- Modify: `docs/testing/testing-strategy.md` (add maintenance section)

**Step 1: Add maintenance section to testing strategy**

```markdown
## Test Maintenance

### When to Update Tests

1. **New Feature**: Add tests for new functionality
2. **Bug Fix**: Add regression test for the bug
3. **Refactoring**: Update tests if implementation changes
4. **API Change**: Update integration/E2E tests for API changes

### Test Debt Management

**Monthly Review**:
- Check for flaky tests
- Review test execution times
- Update outdated test data
- Remove duplicate tests

**Quarterly**:
- Review coverage gaps
- Update performance benchmarks
- Refresh load test scenarios
- Archive obsolete tests

### Deprecating Tests

When removing a test:
1. Ensure functionality is covered elsewhere
2. Update documentation
3. Remove test data/fixtures
4. Update CI/CD workflows

### Adding New Test Scenarios

1. Create test file following existing patterns
2. Add to appropriate test suite
3. Update documentation
4. Add to CI/CD workflow
5. Run locally to verify

### Test Data Management

**DO**:
- Use factories for test data creation
- Clean up test data after each test
- Use unique identifiers (timestamps, UUIDs)
- Store fixtures in version control

**DON'T**:
- Hardcode test data values
- Share test data between tests
- Leave test data in database
- Use production data in tests

### Performance Test Maintenance

**Monthly**:
- Review baseline metrics
- Update performance thresholds
- Check for test drift

**Quarterly**:
- Add new load scenarios
- Update user behavior patterns
- Review infrastructure capacity

### Flaky Test Protocol

1. **Identify**: Mark test as flaky in CI
2. **Investigate**: Determine root cause
3. **Fix**: Address underlying issue
4. **Verify**: Run test 10+ times to confirm stability
5. **Document**: Add to flaky test log

### Test Documentation Updates

Update documentation when:
- Adding new test types
- Changing test infrastructure
- Modifying CI/CD workflows
- Updating performance benchmarks

## Appendix: Test Commands Quick Reference

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# E2E tests
make test-e2e

# All tests
make test-all

# With coverage
make test-coverage

# Load tests
make load-test
make load-test-stress
make load-test-soak

# Clean test cache
make test-clean
```
```

**Step 2: Final commit**

```bash
git add docs/testing/testing-strategy.md
git commit -m "docs: add test maintenance guidelines and quick reference"
```

---

## Phase 5 Complete!

**Summary of Implementation:**

### Test Infrastructure ✅
- Testcontainers setup for isolated test databases
- Mock repositories and test factories
- Test API client and server helpers

### Integration Tests ✅
- Authentication endpoint tests
- Subscription endpoint tests
- IAP verification tests
- Webhook handler tests
- Repository integration tests

### E2E Tests ✅
- E2E test framework with suite setup
- Complete user journey scenario
- Test environment management

### Load Tests ✅
- k6 configuration and helpers
- Basic load test (20 users)
- Stress test (200 users)
- Soak test (30 minutes)
- Grafana dashboard

### CI/CD Integration ✅
- Automated integration test workflow
- E2E test workflow
- Scheduled load test workflow
- Updated backend CI pipeline

### Documentation ✅
- Testing strategy document
- E2E testing guide
- Load testing guide
- Troubleshooting runbook
- Test maintenance guidelines

**Test Coverage**: Target > 80%
**Performance Benchmarks**: Established baseline
**CI/CD**: Automated test execution on PR and push

---

**Ready for feedback!**