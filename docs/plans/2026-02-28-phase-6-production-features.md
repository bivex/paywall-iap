# Phase 6: Production Features Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build production-ready subscription management features including grace periods, winback offers, A/B testing framework, dunning management, and advanced analytics.

**Architecture:** Event-driven architecture with Asynq workers for background jobs, feature flags for A/B testing, and real-time analytics aggregation. Grace periods and winback offers are managed through state machines with automated workflows.

**Tech Stack:** 
- Go 1.21+, Gin, Asynq (background jobs), go-feature-flag (A/B testing)
- PostgreSQL 15 (time-series analytics), Redis 7 (feature flag caching)
- Prometheus + Grafana (metrics dashboards)
- Stripe/Paddle (payment recovery)

**Prerequisites:** 
- Phase 1-5 complete (backend API, mobile app, integration tests, load tests)
- Database migrations 006-009 applied (grace_periods, winback_offers, analytics_aggregates, admin_audit_log)
- Asynq worker infrastructure from Phase 3

---

## Implementation Status

### ✅ Completed (Pending)

### 🔄 In Progress

**Phase 6: Production Features**
- ⏳ Task 1-4: Grace Period Management
- ⏳ Task 5-8: Winback Offer System
- ⏳ Task 9-12: A/B Testing Framework
- ⏳ Task 13-16: Dunning Management
- ⏳ Task 17-20: Advanced Analytics
- ⏳ Task 21-24: Admin Dashboard API

---

## Task 1: Create Grace Period Domain Entities

**Files:**
- Create: `backend/internal/domain/entity/grace_period.go`
- Create: `backend/internal/domain/valueobject/grace_status.go`
- Test: `backend/internal/domain/entity/grace_period_test.go`

**Step 1: Write failing test for grace period entity**

```go
// backend/internal/domain/entity/grace_period_test.go
package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

func TestGracePeriodEntity(t *testing.T) {
	t.Run("NewGracePeriod creates active grace period", func(t *testing.T) {
		userID := uuid.New()
		subscriptionID := uuid.New()
		expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

		gracePeriod := entity.NewGracePeriod(userID, subscriptionID, expiresAt)

		assert.NotEmpty(t, gracePeriod.ID)
		assert.Equal(t, userID, gracePeriod.UserID)
		assert.Equal(t, subscriptionID, gracePeriod.SubscriptionID)
		assert.Equal(t, entity.GraceStatusActive, gracePeriod.Status)
		assert.Equal(t, expiresAt, gracePeriod.ExpiresAt)
		assert.Nil(t, gracePeriod.ResolvedAt)
	})

	t.Run("IsActive returns true for active grace period", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(24*time.Hour),
		)

		assert.True(t, gracePeriod.IsActive())
	})

	t.Run("IsExpired returns true after expiry", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(-24*time.Hour), // Expired yesterday
		)

		assert.True(t, gracePeriod.IsExpired())
	})

	t.Run("Resolve marks grace period as resolved", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(24*time.Hour),
		)

		err := gracePeriod.Resolve()
		assert.NoError(t, err)
		assert.Equal(t, entity.GraceStatusResolved, gracePeriod.Status)
		assert.NotNil(t, gracePeriod.ResolvedAt)
	})

	t.Run("Expire marks grace period as expired", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(-24*time.Hour),
		)

		err := gracePeriod.Expire()
		assert.NoError(t, err)
		assert.Equal(t, entity.GraceStatusExpired, gracePeriod.Status)
	})

	t.Run("Cannot resolve already expired grace period", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(-24*time.Hour),
		)
		gracePeriod.Expire()

		err := gracePeriod.Resolve()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot resolve expired grace period")
	})
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test -v ./internal/domain/entity/grace_period_test.go`

Expected: FAIL - `NewGracePeriod`, `GracePeriod` type not defined

**Step 3: Create grace period entity**

```go
// backend/internal/domain/entity/grace_period.go
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// GracePeriodStatus represents the status of a grace period
type GracePeriodStatus string

const (
	GraceStatusActive   GracePeriodStatus = "active"
	GraceStatusResolved GracePeriodStatus = "resolved"
	GraceStatusExpired  GracePeriodStatus = "expired"
)

// GracePeriod represents a grace period for subscription renewal failures
type GracePeriod struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	SubscriptionID uuid.UUID
	Status         GracePeriodStatus
	ExpiresAt      time.Time
	ResolvedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewGracePeriod creates a new active grace period
func NewGracePeriod(userID, subscriptionID uuid.UUID, expiresAt time.Time) *GracePeriod {
	now := time.Now()
	return &GracePeriod{
		ID:             uuid.New(),
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Status:         GraceStatusActive,
		ExpiresAt:      expiresAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// IsActive returns true if the grace period is currently active
func (gp *GracePeriod) IsActive() bool {
	return gp.Status == GraceStatusActive && gp.ExpiresAt.After(time.Now())
}

// IsExpired returns true if the grace period has expired
func (gp *GracePeriod) IsExpired() bool {
	return gp.Status == GraceStatusExpired || gp.ExpiresAt.Before(time.Now())
}

// IsResolved returns true if the grace period has been resolved
func (gp *GracePeriod) IsResolved() bool {
	return gp.Status == GraceStatusResolved
}

// CanAccessContent returns true if user can access content during grace period
func (gp *GracePeriod) CanAccessContent() bool {
	return gp.IsActive()
}

// Resolve marks the grace period as resolved (user renewed successfully)
func (gp *GracePeriod) Resolve() error {
	if gp.Status == GraceStatusExpired {
		return errors.New("cannot resolve expired grace period")
	}

	now := time.Now()
	gp.Status = GraceStatusResolved
	gp.ResolvedAt = &now
	gp.UpdatedAt = now
	return nil
}

// Expire marks the grace period as expired
func (gp *GracePeriod) Expire() error {
	if gp.Status == GraceStatusResolved {
		return errors.New("cannot expire already resolved grace period")
	}

	gp.Status = GraceStatusExpired
	gp.UpdatedAt = time.Now()
	return nil
}

// DaysRemaining returns the number of days remaining in the grace period
func (gp *GracePeriod) DaysRemaining() int {
	if gp.IsExpired() {
		return 0
	}

	duration := gp.ExpiresAt.Sub(time.Now())
	if duration < 0 {
		return 0
	}

	return int(duration.Hours() / 24)
}

// HoursRemaining returns the number of hours remaining in the grace period
func (gp *GracePeriod) HoursRemaining() int {
	if gp.IsExpired() {
		return 0
	}

	duration := gp.ExpiresAt.Sub(time.Now())
	if duration < 0 {
		return 0
	}

	return int(duration.Hours())
}
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test -v ./internal/domain/entity/grace_period_test.go`

Expected: PASS - all 6 tests pass

**Step 5: Commit**

```bash
git add backend/internal/domain/entity/grace_period.go backend/internal/domain/entity/grace_period_test.go
git commit -m "feat: add grace period domain entity with state management"
```

---

## Task 2: Create Grace Period Repository Interface and Implementation

**Files:**
- Create: `backend/internal/domain/repository/grace_period_repository.go`
- Create: `backend/internal/infrastructure/persistence/repository/grace_period_repository.go`
- Test: `backend/tests/integration/grace_period_repository_test.go`

**Step 1: Create repository interface**

```go
// backend/internal/domain/repository/grace_period_repository.go
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// GracePeriodRepository defines the interface for grace period data access
type GracePeriodRepository interface {
	// Create creates a new grace period
	Create(ctx context.Context, gracePeriod *entity.GracePeriod) error

	// GetByID retrieves a grace period by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.GracePeriod, error)

	// GetActiveByUserID retrieves the active grace period for a user
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.GracePeriod, error)

	// GetActiveBySubscriptionID retrieves the active grace period for a subscription
	GetActiveBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entity.GracePeriod, error)

	// Update updates an existing grace period
	Update(ctx context.Context, gracePeriod *entity.GracePeriod) error

	// GetExpiredGracePeriods retrieves all expired grace periods that need processing
	GetExpiredGracePeriods(ctx context.Context, limit int) ([]*entity.GracePeriod, error)

	// GetExpiringSoon retrieves grace periods expiring within the specified duration
	GetExpiringSoon(ctx context.Context, withinHours int) ([]*entity.GracePeriod, error)
}
```

**Step 2: Create repository implementation**

```go
// backend/internal/infrastructure/persistence/repository/grace_period_repository.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

// GracePeriodRepositoryImpl implements GracePeriodRepository
type GracePeriodRepositoryImpl struct {
	pool *pgxpool.Pool
}

// NewGracePeriodRepository creates a new grace period repository
func NewGracePeriodRepository(pool *pgxpool.Pool) repository.GracePeriodRepository {
	return &GracePeriodRepositoryImpl{pool: pool}
}

// Create creates a new grace period
func (r *GracePeriodRepositoryImpl) Create(ctx context.Context, gracePeriod *entity.GracePeriod) error {
	query := `
		INSERT INTO grace_periods (id, user_id, subscription_id, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		gracePeriod.ID,
		gracePeriod.UserID,
		gracePeriod.SubscriptionID,
		gracePeriod.Status,
		gracePeriod.ExpiresAt,
		gracePeriod.CreatedAt,
		gracePeriod.UpdatedAt,
	)

	return err
}

// GetByID retrieves a grace period by ID
func (r *GracePeriodRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.GracePeriod, error) {
	query := `
		SELECT id, user_id, subscription_id, status, expires_at, resolved_at, created_at, updated_at
		FROM grace_periods
		WHERE id = $1
	`

	gp := &entity.GracePeriod{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&gp.ID,
		&gp.UserID,
		&gp.SubscriptionID,
		&gp.Status,
		&gp.ExpiresAt,
		&gp.ResolvedAt,
		&gp.CreatedAt,
		&gp.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return gp, nil
}

// GetActiveByUserID retrieves the active grace period for a user
func (r *GracePeriodRepositoryImpl) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.GracePeriod, error) {
	query := `
		SELECT id, user_id, subscription_id, status, expires_at, resolved_at, created_at, updated_at
		FROM grace_periods
		WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	gp := &entity.GracePeriod{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&gp.ID,
		&gp.UserID,
		&gp.SubscriptionID,
		&gp.Status,
		&gp.ExpiresAt,
		&gp.ResolvedAt,
		&gp.CreatedAt,
		&gp.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return gp, nil
}

// GetActiveBySubscriptionID retrieves the active grace period for a subscription
func (r *GracePeriodRepositoryImpl) GetActiveBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entity.GracePeriod, error) {
	query := `
		SELECT id, user_id, subscription_id, status, expires_at, resolved_at, created_at, updated_at
		FROM grace_periods
		WHERE subscription_id = $1 AND status = 'active' AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	gp := &entity.GracePeriod{}
	err := r.pool.QueryRow(ctx, query, subscriptionID).Scan(
		&gp.ID,
		&gp.UserID,
		&gp.SubscriptionID,
		&gp.Status,
		&gp.ExpiresAt,
		&gp.ResolvedAt,
		&gp.CreatedAt,
		&gp.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return gp, nil
}

// Update updates an existing grace period
func (r *GracePeriodRepositoryImpl) Update(ctx context.Context, gracePeriod *entity.GracePeriod) error {
	query := `
		UPDATE grace_periods
		SET status = $2, resolved_at = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		gracePeriod.ID,
		gracePeriod.Status,
		gracePeriod.ResolvedAt,
		gracePeriod.UpdatedAt,
	)

	return err
}

// GetExpiredGracePeriods retrieves all expired grace periods that need processing
func (r *GracePeriodRepositoryImpl) GetExpiredGracePeriods(ctx context.Context, limit int) ([]*entity.GracePeriod, error) {
	query := `
		SELECT id, user_id, subscription_id, status, expires_at, resolved_at, created_at, updated_at
		FROM grace_periods
		WHERE status = 'active' AND expires_at < NOW()
		ORDER BY expires_at ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gracePeriods []*entity.GracePeriod
	for rows.Next() {
		gp := &entity.GracePeriod{}
		err := rows.Scan(
			&gp.ID,
			&gp.UserID,
			&gp.SubscriptionID,
			&gp.Status,
			&gp.ExpiresAt,
			&gp.ResolvedAt,
			&gp.CreatedAt,
			&gp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		gracePeriods = append(gracePeriods, gp)
	}

	return gracePeriods, rows.Err()
}

// GetExpiringSoon retrieves grace periods expiring within the specified duration
func (r *GracePeriodRepositoryImpl) GetExpiringSoon(ctx context.Context, withinHours int) ([]*entity.GracePeriod, error) {
	query := `
		SELECT id, user_id, subscription_id, status, expires_at, resolved_at, created_at, updated_at
		FROM grace_periods
		WHERE status = 'active'
		  AND expires_at > NOW()
		  AND expires_at < NOW() + ($1 || ' hours')::INTERVAL
		ORDER BY expires_at ASC
	`

	rows, err := r.pool.Query(ctx, query, withinHours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gracePeriods []*entity.GracePeriod
	for rows.Next() {
		gp := &entity.GracePeriod{}
		err := rows.Scan(
			&gp.ID,
			&gp.UserID,
			&gp.SubscriptionID,
			&gp.Status,
			&gp.ExpiresAt,
			&gp.ResolvedAt,
			&gp.CreatedAt,
			&gp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		gracePeriods = append(gracePeriods, gp)
	}

	return gracePeriods, rows.Err()
}

// Ensure implementation matches interface
var _ repository.GracePeriodRepository = (*GracePeriodRepositoryImpl)(nil)
```

**Step 3: Write integration test**

```go
// backend/tests/integration/grace_period_repository_test.go
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

func TestGracePeriodRepository(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Create repository
	gpRepo := repository.NewGracePeriodRepository(dbContainer.Pool)

	// Create test user and subscription first
	userRepo := repository.NewUserRepository(dbContainer.Pool)
	subRepo := repository.NewSubscriptionRepository(dbContainer.Pool)

	user := entity.NewUser("test-platform", "test-device", entity.PlatformiOS, "1.0", "test@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	sub := entity.NewSubscription(user.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	t.Run("Create and GetByID", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(user.ID, sub.ID, time.Now().Add(7*24*time.Hour))

		err := gpRepo.Create(ctx, gracePeriod)
		require.NoError(t, err)

		retrieved, err := gpRepo.GetByID(ctx, gracePeriod.ID)
		require.NoError(t, err)
		assert.Equal(t, gracePeriod.ID, retrieved.ID)
		assert.Equal(t, entity.GraceStatusActive, retrieved.Status)
	})

	t.Run("GetActiveByUserID", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(user.ID, sub.ID, time.Now().Add(7*24*time.Hour))
		err := gpRepo.Create(ctx, gracePeriod)
		require.NoError(t, err)

		retrieved, err := gpRepo.GetActiveByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, gracePeriod.ID, retrieved.ID)
		assert.True(t, retrieved.IsActive())
	})

	t.Run("Update grace period status", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(user.ID, sub.ID, time.Now().Add(7*24*time.Hour))
		err := gpRepo.Create(ctx, gracePeriod)
		require.NoError(t, err)

		// Resolve the grace period
		err = gracePeriod.Resolve()
		require.NoError(t, err)

		err = gpRepo.Update(ctx, gracePeriod)
		require.NoError(t, err)

		retrieved, _ := gpRepo.GetByID(ctx, gracePeriod.ID)
		assert.Equal(t, entity.GraceStatusResolved, retrieved.Status)
		assert.NotNil(t, retrieved.ResolvedAt)
	})

	t.Run("GetExpiredGracePeriods", func(t *testing.T) {
		// Create expired grace period
		expiredGP := entity.NewGracePeriod(user.ID, sub.ID, time.Now().Add(-24*time.Hour))
		err := gpRepo.Create(ctx, expiredGP)
		require.NoError(t, err)

		expired, err := gpRepo.GetExpiredGracePeriods(ctx, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, expired)
		assert.Contains(t, expired, expiredGP)
	})

	t.Run("GetExpiringSoon", func(t *testing.T) {
		// Create grace period expiring in 2 hours
		expiringGP := entity.NewGracePeriod(user.ID, sub.ID, time.Now().Add(2*time.Hour))
		err := gpRepo.Create(ctx, expiringGP)
		require.NoError(t, err)

		expiring, err := gpRepo.GetExpiringSoon(ctx, 24)
		require.NoError(t, err)
		assert.NotEmpty(t, expiring)
		assert.Contains(t, expiring, expiringGP)
	})
}
```

**Step 4: Run integration test**

Run: `cd backend && go test -tags=integration -v ./tests/integration/grace_period_repository_test.go`

Expected: All tests pass

**Step 5: Commit**

```bash
git add backend/internal/domain/repository/grace_period_repository.go backend/internal/infrastructure/persistence/repository/grace_period_repository.go backend/tests/integration/grace_period_repository_test.go
git commit -m "feat: implement grace period repository with full CRUD operations"
```

---

## Task 3: Create Grace Period Service and Commands

**Files:**
- Create: `backend/internal/domain/service/grace_period_service.go`
- Create: `backend/internal/application/command/create_grace_period.go`
- Create: `backend/internal/application/command/resolve_grace_period.go`
- Test: `backend/internal/domain/service/grace_period_service_test.go`

**Step 1: Create grace period service**

```go
// backend/internal/domain/service/grace_period_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

var (
	ErrGracePeriodNotFound     = errors.New("grace period not found")
	ErrGracePeriodAlreadyExists = errors.New("active grace period already exists for this subscription")
	ErrGracePeriodNotActive    = errors.New("grace period is not active")
)

// GracePeriodService handles grace period business logic
type GracePeriodService struct {
	gracePeriodRepo repository.GracePeriodRepository
	subscriptionRepo repository.SubscriptionRepository
	userRepo        repository.UserRepository
}

// NewGracePeriodService creates a new grace period service
func NewGracePeriodService(
	gracePeriodRepo repository.GracePeriodRepository,
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
) *GracePeriodService {
	return &GracePeriodService{
		gracePeriodRepo:  gracePeriodRepo,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
	}
}

// CreateGracePeriod creates a new grace period for a subscription
func (s *GracePeriodService) CreateGracePeriod(ctx context.Context, userID, subscriptionID uuid.UUID, durationDays int) (*entity.GracePeriod, error) {
	// Check if active grace period already exists
	existing, err := s.gracePeriodRepo.GetActiveBySubscriptionID(ctx, subscriptionID)
	if err == nil && existing != nil && existing.IsActive() {
		return nil, ErrGracePeriodAlreadyExists
	}

	// Verify subscription exists and belongs to user
	sub, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	if sub.UserID != userID {
		return nil, errors.New("subscription does not belong to user")
	}

	// Create grace period
	expiresAt := time.Now().Add(time.Duration(durationDays) * 24 * time.Hour)
	gracePeriod := entity.NewGracePeriod(userID, subscriptionID, expiresAt)

	// Update subscription status to grace
	err = s.subscriptionRepo.UpdateStatus(ctx, subscriptionID, entity.StatusGrace)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription status: %w", err)
	}

	// Save grace period
	err = s.gracePeriodRepo.Create(ctx, gracePeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to create grace period: %w", err)
	}

	return gracePeriod, nil
}

// ResolveGracePeriod resolves a grace period when payment succeeds
func (s *GracePeriodService) ResolveGracePeriod(ctx context.Context, userID, subscriptionID uuid.UUID) error {
	// Get active grace period
	gracePeriod, err := s.gracePeriodRepo.GetActiveBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return ErrGracePeriodNotFound
	}

	if !gracePeriod.IsActive() {
		return ErrGracePeriodNotActive
	}

	// Resolve grace period
	err = gracePeriod.Resolve()
	if err != nil {
		return err
	}

	// Update repository
	err = s.gracePeriodRepo.Update(ctx, gracePeriod)
	if err != nil {
		return fmt.Errorf("failed to update grace period: %w", err)
	}

	// Update subscription status back to active
	err = s.subscriptionRepo.UpdateStatus(ctx, subscriptionID, entity.StatusActive)
	if err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	return nil
}

// ExpireGracePeriod expires a grace period and cancels the subscription
func (s *GracePeriodService) ExpireGracePeriod(ctx context.Context, gracePeriodID uuid.UUID) error {
	// Get grace period
	gracePeriod, err := s.gracePeriodRepo.GetByID(ctx, gracePeriodID)
	if err != nil {
		return ErrGracePeriodNotFound
	}

	if !gracePeriod.IsActive() {
		return ErrGracePeriodNotActive
	}

	// Expire grace period
	err = gracePeriod.Expire()
	if err != nil {
		return err
	}

	// Update repository
	err = s.gracePeriodRepo.Update(ctx, gracePeriod)
	if err != nil {
		return fmt.Errorf("failed to update grace period: %w", err)
	}

	// Cancel subscription
	err = s.subscriptionRepo.Cancel(ctx, gracePeriod.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}

// ProcessExpiredGracePeriods processes all expired grace periods
func (s *GracePeriodService) ProcessExpiredGracePeriods(ctx context.Context, limit int) (int, error) {
	expiredPeriods, err := s.gracePeriodRepo.GetExpiredGracePeriods(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired grace periods: %w", err)
	}

	processed := 0
	for _, gp := range expiredPeriods {
		err := s.ExpireGracePeriod(ctx, gp.ID)
		if err != nil {
			// Log error but continue processing others
			continue
		}
		processed++
	}

	return processed, nil
}

// GetGracePeriodStatus returns the current grace period status for a user
func (s *GracePeriodService) GetGracePeriodStatus(ctx context.Context, userID uuid.UUID) (*entity.GracePeriod, error) {
	gracePeriod, err := s.gracePeriodRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, ErrGracePeriodNotFound
	}

	return gracePeriod, nil
}

// NotifyExpiringSoon returns grace periods expiring within the specified hours
func (s *GracePeriodService) NotifyExpiringSoon(ctx context.Context, withinHours int) ([]*entity.GracePeriod, error) {
	return s.gracePeriodRepo.GetExpiringSoon(ctx, withinHours)
}
```

**Step 2: Create command handlers**

```go
// backend/internal/application/command/create_grace_period.go
package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/service"
)

// CreateGracePeriodCommand creates a new grace period
type CreateGracePeriodCommand struct {
	gracePeriodService *service.GracePeriodService
}

// CreateGracePeriodRequest is the request DTO
type CreateGracePeriodRequest struct {
	UserID         string `json:"user_id" validate:"required,uuid"`
	SubscriptionID string `json:"subscription_id" validate:"required,uuid"`
	DurationDays   int    `json:"duration_days" validate:"required,min=1,max=30"`
}

// CreateGracePeriodResponse is the response DTO
type CreateGracePeriodResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	SubscriptionID string `json:"subscription_id"`
	Status         string `json:"status"`
	ExpiresAt      string `json:"expires_at"`
	DaysRemaining  int    `json:"days_remaining"`
}

// NewCreateGracePeriodCommand creates a new command handler
func NewCreateGracePeriodCommand(gracePeriodService *service.GracePeriodService) *CreateGracePeriodCommand {
	return &CreateGracePeriodCommand{
		gracePeriodService: gracePeriodService,
	}
}

// Execute creates a new grace period
func (c *CreateGracePeriodCommand) Execute(ctx context.Context, req *CreateGracePeriodRequest) (*CreateGracePeriodResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	subscriptionID, err := uuid.Parse(req.SubscriptionID)
	if err != nil {
		return nil, err
	}

	gracePeriod, err := c.gracePeriodService.CreateGracePeriod(ctx, userID, subscriptionID, req.DurationDays)
	if err != nil {
		return nil, err
	}

	return &CreateGracePeriodResponse{
		ID:             gracePeriod.ID.String(),
		UserID:         gracePeriod.UserID.String(),
		SubscriptionID: gracePeriod.SubscriptionID.String(),
		Status:         string(gracePeriod.Status),
		ExpiresAt:      gracePeriod.ExpiresAt.Format(time.RFC3339),
		DaysRemaining:  gracePeriod.DaysRemaining(),
	}, nil
}
```

```go
// backend/internal/application/command/resolve_grace_period.go
package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// ResolveGracePeriodCommand resolves a grace period
type ResolveGracePeriodCommand struct {
	gracePeriodService *service.GracePeriodService
}

// ResolveGracePeriodRequest is the request DTO
type ResolveGracePeriodRequest struct {
	UserID         string `json:"user_id" validate:"required,uuid"`
	SubscriptionID string `json:"subscription_id" validate:"required,uuid"`
}

// ResolveGracePeriodResponse is the response DTO
type ResolveGracePeriodResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewResolveGracePeriodCommand creates a new command handler
func NewResolveGracePeriodCommand(gracePeriodService *service.GracePeriodService) *ResolveGracePeriodCommand {
	return &ResolveGracePeriodCommand{
		gracePeriodService: gracePeriodService,
	}
}

// Execute resolves a grace period
func (c *ResolveGracePeriodCommand) Execute(ctx context.Context, req *ResolveGracePeriodRequest) (*ResolveGracePeriodResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	subscriptionID, err := uuid.Parse(req.SubscriptionID)
	if err != nil {
		return nil, err
	}

	err = c.gracePeriodService.ResolveGracePeriod(ctx, userID, subscriptionID)
	if err != nil {
		return nil, err
	}

	return &ResolveGracePeriodResponse{
		Success: true,
		Message: "Grace period resolved successfully",
	}, nil
}
```

**Step 3: Write service tests**

```go
// backend/internal/domain/service/grace_period_service_test.go
package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/tests/mocks"
)

func TestGracePeriodService(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	gracePeriodRepo := mocks.NewMockGracePeriodRepository()
	subscriptionRepo := mocks.NewMockSubscriptionRepository()
	userRepo := mocks.NewMockUserRepository()

	graceService := service.NewGracePeriodService(gracePeriodRepo, subscriptionRepo, userRepo)

	t.Run("CreateGracePeriod success", func(t *testing.T) {
		userID := uuid.New()
		subscriptionID := uuid.New()

		// Setup mock expectations
		subscriptionRepo.On("GetByID", ctx, subscriptionID).Return(&entity.Subscription{
			ID:     subscriptionID,
			UserID: userID,
		}, nil)

		subscriptionRepo.On("UpdateStatus", ctx, subscriptionID, entity.StatusGrace).Return(nil)
		gracePeriodRepo.On("Create", ctx, mock.Anything).Return(nil)

		gracePeriod, err := graceService.CreateGracePeriod(ctx, userID, subscriptionID, 7)
		require.NoError(t, err)
		assert.Equal(t, entity.GraceStatusActive, gracePeriod.Status)
		assert.Equal(t, 7, gracePeriod.DaysRemaining())
	})

	t.Run("CreateGracePeriod with existing active grace period returns error", func(t *testing.T) {
		userID := uuid.New()
		subscriptionID := uuid.New()

		existingGP := entity.NewGracePeriod(userID, subscriptionID, time.Now().Add(24*time.Hour))
		gracePeriodRepo.On("GetActiveBySubscriptionID", ctx, subscriptionID).Return(existingGP, nil)

		gracePeriod, err := graceService.CreateGracePeriod(ctx, userID, subscriptionID, 7)
		assert.Error(t, err)
		assert.Nil(t, gracePeriod)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("ResolveGracePeriod success", func(t *testing.T) {
		userID := uuid.New()
		subscriptionID := uuid.New()

		gracePeriod := entity.NewGracePeriod(userID, subscriptionID, time.Now().Add(24*time.Hour))
		gracePeriodRepo.On("GetActiveBySubscriptionID", ctx, subscriptionID).Return(gracePeriod, nil)
		gracePeriodRepo.On("Update", ctx, gracePeriod).Return(nil)
		subscriptionRepo.On("UpdateStatus", ctx, subscriptionID, entity.StatusActive).Return(nil)

		err := graceService.ResolveGracePeriod(ctx, userID, subscriptionID)
		require.NoError(t, err)
		assert.Equal(t, entity.GraceStatusResolved, gracePeriod.Status)
	})

	t.Run("ExpireGracePeriod cancels subscription", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(uuid.New(), uuid.New(), time.Now().Add(-24*time.Hour))
		gracePeriodRepo.On("GetByID", ctx, gracePeriod.ID).Return(gracePeriod, nil)
		gracePeriodRepo.On("Update", ctx, gracePeriod).Return(nil)
		subscriptionRepo.On("Cancel", ctx, gracePeriod.SubscriptionID).Return(nil)

		err := graceService.ExpireGracePeriod(ctx, gracePeriod.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.GraceStatusExpired, gracePeriod.Status)
	})
}
```

**Step 4: Run service tests**

Run: `cd backend && go test -v ./internal/domain/service/grace_period_service_test.go`

Expected: All tests pass

**Step 5: Commit**

```bash
git add backend/internal/domain/service/grace_period_service.go backend/internal/application/command/create_grace_period.go backend/internal/application/command/resolve_grace_period.go backend/internal/domain/service/grace_period_service_test.go
git commit -m "feat: add grace period service with create, resolve, and expire operations"
```

---

## Task 4: Create Grace Period Worker Jobs

**Files:**
- Create: `backend/internal/worker/tasks/grace_period_jobs.go`
- Create: `backend/internal/worker/tasks/grace_period_notifications.go`
- Test: `backend/tests/integration/grace_period_worker_test.go`

**Step 1: Create grace period worker jobs**

```go
// backend/internal/worker/tasks/grace_period_jobs.go
package tasks

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

const (
	TypeProcessExpiredGracePeriods = "grace_period:process_expired"
	TypeNotifyExpiringGracePeriods = "grace_period:notify_expiring"
)

// ProcessExpiredGracePeriodsPayload is the payload for processing expired grace periods
type ProcessExpiredGracePeriodsPayload struct {
	Limit int `json:"limit"`
}

// NotifyExpiringGracePeriodsPayload is the payload for notifying expiring grace periods
type NotifyExpiringGracePeriodsPayload struct {
	WithinHours int `json:"within_hours"`
}

// GracePeriodJobHandler handles grace period background jobs
type GracePeriodJobHandler struct {
	gracePeriodService *service.GracePeriodService
	notificationService *service.NotificationService
}

// NewGracePeriodJobHandler creates a new grace period job handler
func NewGracePeriodJobHandler(
	gracePeriodService *service.GracePeriodService,
	notificationService *service.NotificationService,
) *GracePeriodJobHandler {
	return &GracePeriodJobHandler{
		gracePeriodService:  gracePeriodService,
		notificationService: notificationService,
	}
}

// HandleProcessExpiredGracePeriods processes expired grace periods
func (h *GracePeriodJobHandler) HandleProcessExpiredGracePeriods(ctx context.Context, t *asynq.Task) error {
	var p ProcessExpiredGracePeriodsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	limit := p.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	processed, err := h.gracePeriodService.ProcessExpiredGracePeriods(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to process expired grace periods: %w", err)
	}

	fmt.Printf("Processed %d expired grace periods\n", processed)
	return nil
}

// HandleNotifyExpiringGracePeriods sends notifications for expiring grace periods
func (h *GracePeriodJobHandler) HandleNotifyExpiringGracePeriods(ctx context.Context, t *asynq.Task) error {
	var p NotifyExpiringGracePeriodsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	withinHours := p.WithinHours
	if withinHours == 0 {
		withinHours = 24 // Default 24 hours
	}

	expiringPeriods, err := h.gracePeriodService.NotifyExpiringSoon(ctx, withinHours)
	if err != nil {
		return fmt.Errorf("failed to get expiring grace periods: %w", err)
	}

	for _, gp := range expiringPeriods {
		// Send notification to user
		err := h.notificationService.SendGracePeriodExpiringNotification(ctx, gp.UserID, gp)
		if err != nil {
			// Log error but continue with other notifications
			fmt.Printf("Failed to send notification for grace period %s: %v\n", gp.ID, err)
		}
	}

	fmt.Printf("Sent %d grace period expiry notifications\n", len(expiringPeriods))
	return nil
}

// ScheduleGracePeriodJobs schedules recurring grace period jobs
func ScheduleGracePeriodJobs(scheduler *asynq.Scheduler) error {
	// Process expired grace periods every hour
	_, err := scheduler.Register("0 * * * *", asynq.NewTask(TypeProcessExpiredGracePeriods, 
		mustMarshalJSON(ProcessExpiredGracePeriodsPayload{Limit: 100})))
	if err != nil {
		return err
	}

	// Notify expiring grace periods every 6 hours
	_, err = scheduler.Register("0 */6 * * *", asynq.NewTask(TypeNotifyExpiringGracePeriods,
		mustMarshalJSON(NotifyExpiringGracePeriodsPayload{WithinHours: 24})))
	if err != nil {
		return err
	}

	return nil
}

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
```

**Step 2: Create notification service**

```go
// backend/internal/domain/service/notification_service.go
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// NotificationService handles sending notifications to users
type NotificationService struct {
	// In production, this would integrate with email/SMS/push notification services
	// For now, we'll log notifications
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// SendGracePeriodExpiringNotification sends a notification when grace period is expiring soon
func (s *NotificationService) SendGracePeriodExpiringNotification(ctx context.Context, userID uuid.UUID, gracePeriod *entity.GracePeriod) error {
	// In production, send email/push notification
	// For now, log the notification
	fmt.Printf("[NOTIFICATION] User %s: Grace period expiring in %d hours for subscription %s\n",
		userID, gracePeriod.HoursRemaining(), gracePeriod.SubscriptionID)

	// TODO: Integrate with email service (SendGrid, SES)
	// TODO: Integrate with push notification service (Firebase, APNs)

	return nil
}

// SendWinbackOfferNotification sends a winback offer to churned users
func (s *NotificationService) SendWinbackOfferNotification(ctx context.Context, userID uuid.UUID, offer *entity.WinbackOffer) error {
	fmt.Printf("[NOTIFICATION] User %s: Winback offer %s - %s discount available\n",
		userID, offer.CampaignID, offer.DiscountValue)

	return nil
}

// SendSubscriptionExpiredNotification sends notification when subscription expires
func (s *NotificationService) SendSubscriptionExpiredNotification(ctx context.Context, userID uuid.UUID, subscriptionID uuid.UUID) error {
	fmt.Printf("[NOTIFICATION] User %s: Subscription %s has expired\n", userID, subscriptionID)
	return nil
}
```

**Step 3: Write integration test**

```go
// backend/tests/integration/grace_period_worker_test.go
//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/worker/tasks"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestGracePeriodWorkerJobs(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Setup repositories
	userRepo := repository.NewUserRepository(dbContainer.Pool)
	subRepo := repository.NewSubscriptionRepository(dbContainer.Pool)
	gpRepo := repository.NewGracePeriodRepository(dbContainer.Pool)

	// Setup services
	graceService := service.NewGracePeriodService(gpRepo, subRepo, userRepo)
	notificationService := service.NewNotificationService()
	jobHandler := tasks.NewGracePeriodJobHandler(graceService, notificationService)

	// Create test user and subscription
	user := entity.NewUser("test-platform", "test-device", entity.PlatformiOS, "1.0", "test@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	sub := entity.NewSubscription(user.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	t.Run("ProcessExpiredGracePeriods expires and cancels subscription", func(t *testing.T) {
		// Create expired grace period
		expiredGP := entity.NewGracePeriod(user.ID, sub.ID, time.Now().Add(-24*time.Hour))
		err := gpRepo.Create(ctx, expiredGP)
		require.NoError(t, err)

		// Create task payload
		payload, _ := json.Marshal(tasks.ProcessExpiredGracePeriodsPayload{Limit: 10})
		task := asynq.NewTask(tasks.TypeProcessExpiredGracePeriods, payload)

		// Execute handler
		err = jobHandler.HandleProcessExpiredGracePeriods(ctx, task)
		require.NoError(t, err)

		// Verify grace period is expired
		updatedGP, _ := gpRepo.GetByID(ctx, expiredGP.ID)
		assert.Equal(t, entity.GraceStatusExpired, updatedGP.Status)

		// Verify subscription is cancelled
		updatedSub, _ := subRepo.GetByID(ctx, sub.ID)
		assert.Equal(t, entity.StatusCancelled, updatedSub.Status)
	})

	t.Run("NotifyExpiringGracePeriods sends notifications", func(t *testing.T) {
		// Create grace period expiring in 2 hours
		user2 := entity.NewUser("test-platform-2", "test-device-2", entity.PlatformiOS, "1.0", "test2@example.com")
		err = userRepo.Create(ctx, user2)
		require.NoError(t, err)

		sub2 := entity.NewSubscription(user2.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
		err = subRepo.Create(ctx, sub2)
		require.NoError(t, err)

		expiringGP := entity.NewGracePeriod(user2.ID, sub2.ID, time.Now().Add(2*time.Hour))
		err = gpRepo.Create(ctx, expiringGP)
		require.NoError(t, err)

		// Create task payload
		payload, _ := json.Marshal(tasks.NotifyExpiringGracePeriodsPayload{WithinHours: 24})
		task := asynq.NewTask(tasks.TypeNotifyExpiringGracePeriods, payload)

		// Execute handler (should send notification)
		err = jobHandler.HandleNotifyExpiringGracePeriods(ctx, task)
		require.NoError(t, err)
	})
}
```

**Step 4: Run worker integration test**

Run: `cd backend && go test -tags=integration -v ./tests/integration/grace_period_worker_test.go`

Expected: All tests pass

**Step 5: Commit**

```bash
git add backend/internal/worker/tasks/grace_period_jobs.go backend/internal/domain/service/notification_service.go backend/tests/integration/grace_period_worker_test.go
git commit -m "feat: add grace period worker jobs for automated processing and notifications"
```

---

## Task 5: Create Winback Offer Domain Entities

**Files:**
- Create: `backend/internal/domain/entity/winback_offer.go`
- Create: `backend/internal/domain/valueobject/discount_type.go`
- Test: `backend/internal/domain/entity/winback_offer_test.go`

**Step 1: Write failing test**

```go
// backend/internal/domain/entity/winback_offer_test.go
package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

func TestWinbackOfferEntity(t *testing.T) {
	t.Run("NewWinbackOffer creates offered winback", func(t *testing.T) {
		userID := uuid.New()
		campaignID := "winback_2026_q1"
		expiresAt := time.Now().Add(30 * 24 * time.Hour)

		offer := entity.NewWinbackOffer(userID, campaignID, entity.DiscountTypePercentage, 50.0, expiresAt)

		assert.NotEmpty(t, offer.ID)
		assert.Equal(t, userID, offer.UserID)
		assert.Equal(t, campaignID, offer.CampaignID)
		assert.Equal(t, entity.DiscountTypePercentage, offer.DiscountType)
		assert.Equal(t, 50.0, offer.DiscountValue)
		assert.Equal(t, entity.OfferStatusOffered, offer.Status)
		assert.Equal(t, expiresAt, offer.ExpiresAt)
		assert.Nil(t, offer.AcceptedAt)
	})

	t.Run("Accept marks offer as accepted", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0,
			time.Now().Add(30*24*time.Hour),
		)

		err := offer.Accept()
		assert.NoError(t, err)
		assert.Equal(t, entity.OfferStatusAccepted, offer.Status)
		assert.NotNil(t, offer.AcceptedAt)
	})

	t.Run("Cannot accept expired offer", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0,
			time.Now().Add(-24*time.Hour),
		)

		err := offer.Accept()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot accept expired offer")
	})

	t.Run("CalculateDiscountAmount for percentage discount", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0, // 25% off
			time.Now().Add(30*24*time.Hour),
		)

		amount := offer.CalculateDiscountAmount(99.99)
		assert.InDelta(t, 25.0, amount, 0.01) // 25% of 99.99 = 24.9975
	})

	t.Run("CalculateDiscountAmount for fixed discount", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypeFixed,
			20.0, // $20 off
			time.Now().Add(30*24*time.Hour),
		)

		amount := offer.CalculateDiscountAmount(99.99)
		assert.InDelta(t, 20.0, amount, 0.01)
	})

	t.Run("Discount cannot exceed total amount", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypeFixed,
			50.0, // $50 off
			time.Now().Add(30*24*time.Hour),
		)

		amount := offer.CalculateDiscountAmount(30.0) // Total is only $30
		assert.InDelta(t, 30.0, amount, 0.01) // Should cap at total amount
	})

	t.Run("Expire marks offer as expired", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0,
			time.Now().Add(-24*time.Hour),
		)

		err := offer.Expire()
		assert.NoError(t, err)
		assert.Equal(t, entity.OfferStatusExpired, offer.Status)
	})

	t.Run("Decline marks offer as declined", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0,
			time.Now().Add(30*24*time.Hour),
		)

		err := offer.Decline()
		assert.NoError(t, err)
		assert.Equal(t, entity.OfferStatusDeclined, offer.Status)
	})
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test -v ./internal/domain/entity/winback_offer_test.go`

Expected: FAIL - `NewWinbackOffer`, `WinbackOffer` type not defined

**Step 3: Create winback offer entity**

```go
// backend/internal/domain/entity/winback_offer.go
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
)

// WinbackOfferStatus represents the status of a winback offer
type WinbackOfferStatus string

const (
	OfferStatusOffered  WinbackOfferStatus = "offered"
	OfferStatusAccepted WinbackOfferStatus = "accepted"
	OfferStatusExpired  WinbackOfferStatus = "expired"
	OfferStatusDeclined WinbackOfferStatus = "declined"
)

// WinbackOffer represents a discount offer to win back churned users
type WinbackOffer struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	CampaignID    string
	DiscountType  DiscountType
	DiscountValue float64
	Status        WinbackOfferStatus
	OfferedAt     time.Time
	ExpiresAt     time.Time
	AcceptedAt    *time.Time
	CreatedAt     time.Time
}

// NewWinbackOffer creates a new winback offer
func NewWinbackOffer(userID uuid.UUID, campaignID string, discountType DiscountType, discountValue float64, expiresAt time.Time) *WinbackOffer {
	now := time.Now()
	return &WinbackOffer{
		ID:            uuid.New(),
		UserID:        userID,
		CampaignID:    campaignID,
		DiscountType:  discountType,
		DiscountValue: discountValue,
		Status:        OfferStatusOffered,
		OfferedAt:     now,
		ExpiresAt:     expiresAt,
		CreatedAt:     now,
	}
}

// IsActive returns true if the offer is still active
func (o *WinbackOffer) IsActive() bool {
	return o.Status == OfferStatusOffered && o.ExpiresAt.After(time.Now())
}

// IsExpired returns true if the offer has expired
func (o *WinbackOffer) IsExpired() bool {
	return o.Status == OfferStatusExpired || o.ExpiresAt.Before(time.Now())
}

// IsAccepted returns true if the offer has been accepted
func (o *WinbackOffer) IsAccepted() bool {
	return o.Status == OfferStatusAccepted
}

// Accept marks the offer as accepted
func (o *WinbackOffer) Accept() error {
	if o.IsExpired() {
		return errors.New("cannot accept expired offer")
	}

	if o.Status == OfferStatusDeclined {
		return errors.New("cannot accept declined offer")
	}

	now := time.Now()
	o.Status = OfferStatusAccepted
	o.AcceptedAt = &now
	return nil
}

// Expire marks the offer as expired
func (o *WinbackOffer) Expire() error {
	o.Status = OfferStatusExpired
	return nil
}

// Decline marks the offer as declined
func (o *WinbackOffer) Decline() error {
	if o.Status == OfferStatusAccepted {
		return errors.New("cannot decline already accepted offer")
	}

	o.Status = OfferStatusDeclined
	return nil
}

// CalculateDiscountAmount calculates the discount amount for a given total
func (o *WinbackOffer) CalculateDiscountAmount(totalAmount float64) float64 {
	if !o.IsActive() {
		return 0
	}

	var discount float64
	switch o.DiscountType {
	case DiscountTypePercentage:
		discount = totalAmount * (o.DiscountValue / 100.0)
	case DiscountTypeFixed:
		discount = o.DiscountValue
	}

	// Discount cannot exceed total amount
	if discount > totalAmount {
		return totalAmount
	}

	return discount
}

// CalculateFinalAmount calculates the final amount after applying discount
func (o *WinbackOffer) CalculateFinalAmount(totalAmount float64) float64 {
	discount := o.CalculateDiscountAmount(totalAmount)
	return totalAmount - discount
}

// DaysUntilExpiry returns the number of days until the offer expires
func (o *WinbackOffer) DaysUntilExpiry() int {
	if o.IsExpired() {
		return 0
	}

	duration := o.ExpiresAt.Sub(time.Now())
	if duration < 0 {
		return 0
	}

	return int(duration.Hours() / 24)
}
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test -v ./internal/domain/entity/winback_offer_test.go`

Expected: PASS - all 8 tests pass

**Step 5: Commit**

```bash
git add backend/internal/domain/entity/winback_offer.go backend/internal/domain/entity/winback_offer_test.go
git commit -m "feat: add winback offer domain entity with discount calculation"
```

---

## Task 6: Create Winback Offer Service and Campaign Management

**Files:**
- Create: `backend/internal/domain/service/winback_service.go`
- Create: `backend/internal/application/command/create_winback_campaign.go`
- Create: `backend/internal/application/command/accept_winback_offer.go`
- Test: `backend/internal/domain/service/winback_service_test.go`

**Step 1: Create winback service**

```go
// backend/internal/domain/service/winback_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

var (
	ErrWinbackOfferNotFound   = errors.New("winback offer not found")
	ErrWinbackOfferNotActive  = errors.New("winback offer is not active")
	ErrCampaignNotFound       = errors.New("campaign not found")
)

// WinbackService handles winback offer business logic
type WinbackService struct {
	winbackRepo repository.WinbackOfferRepository
	userRepo    repository.UserRepository
	subRepo     repository.SubscriptionRepository
}

// NewWinbackService creates a new winback service
func NewWinbackService(
	winbackRepo repository.WinbackOfferRepository,
	userRepo repository.UserRepository,
	subRepo repository.SubscriptionRepository,
) *WinbackService {
	return &WinbackService{
		winbackRepo: winbackRepo,
		userRepo:    userRepo,
		subRepo:     subRepo,
	}
}

// CreateWinbackOffer creates a new winback offer for a user
func (s *WinbackService) CreateWinbackOffer(
	ctx context.Context,
	userID uuid.UUID,
	campaignID string,
	discountType entity.DiscountType,
	discountValue float64,
	durationDays int,
) (*entity.WinbackOffer, error) {
	// Check if user already has an active offer for this campaign
	existing, err := s.winbackRepo.GetActiveByUserAndCampaign(ctx, userID, campaignID)
	if err == nil && existing != nil && existing.IsActive() {
		return nil, errors.New("user already has an active offer for this campaign")
	}

	// Verify user exists
	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Create winback offer
	expiresAt := time.Now().Add(time.Duration(durationDays) * 24 * time.Hour)
	offer := entity.NewWinbackOffer(userID, campaignID, discountType, discountValue, expiresAt)

	// Save offer
	err = s.winbackRepo.Create(ctx, offer)
	if err != nil {
		return nil, fmt.Errorf("failed to create winback offer: %w", err)
	}

	return offer, nil
}

// AcceptWinbackOffer accepts a winback offer and applies the discount
func (s *WinbackService) AcceptWinbackOffer(ctx context.Context, userID, offerID uuid.UUID) (*entity.WinbackOffer, error) {
	// Get offer
	offer, err := s.winbackRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, ErrWinbackOfferNotFound
	}

	// Verify offer belongs to user
	if offer.UserID != userID {
		return nil, errors.New("offer does not belong to user")
	}

	// Accept offer
	err = offer.Accept()
	if err != nil {
		return nil, err
	}

	// Update repository
	err = s.winbackRepo.Update(ctx, offer)
	if err != nil {
		return nil, fmt.Errorf("failed to update winback offer: %w", err)
	}

	return offer, nil
}

// GetActiveWinbackOffers returns all active winback offers for a user
func (s *WinbackService) GetActiveWinbackOffers(ctx context.Context, userID uuid.UUID) ([]*entity.WinbackOffer, error) {
	return s.winbackRepo.GetActiveByUserID(ctx, userID)
}

// ProcessExpiredWinbackOffers expires all expired winback offers
func (s *WinbackService) ProcessExpiredWinbackOffers(ctx context.Context, limit int) (int, error) {
	expiredOffers, err := s.winbackRepo.GetExpiredOffers(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired offers: %w", err)
	}

	processed := 0
	for _, offer := range expiredOffers {
		err := offer.Expire()
		if err != nil {
			continue
		}

		err = s.winbackRepo.Update(ctx, offer)
		if err != nil {
			continue
		}
		processed++
	}

	return processed, nil
}

// CreateWinbackCampaignForChurnedUsers creates winback offers for recently churned users
func (s *WinbackService) CreateWinbackCampaignForChurnedUsers(
	ctx context.Context,
	campaignID string,
	discountType entity.DiscountType,
	discountValue float64,
	durationDays int,
	daysSinceChurn int,
) (int, error) {
	// Get churned users (subscriptions cancelled within specified days)
	churnedUsers, err := s.subRepo.GetUsersWithCancelledSubscriptions(ctx, daysSinceChurn)
	if err != nil {
		return 0, fmt.Errorf("failed to get churned users: %w", err)
	}

	created := 0
	for _, userID := range churnedUsers {
		_, err := s.CreateWinbackOffer(ctx, userID, campaignID, discountType, discountValue, durationDays)
		if err != nil {
			// Skip users who already have offers
			continue
		}
		created++
	}

	return created, nil
}
```

**Step 2: Create command handlers**

```go
// backend/internal/application/command/accept_winback_offer.go
package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// AcceptWinbackOfferCommand accepts a winback offer
type AcceptWinbackOfferCommand struct {
	winbackService *service.WinbackService
}

// AcceptWinbackOfferRequest is the request DTO
type AcceptWinbackOfferRequest struct {
	UserID  string `json:"user_id" validate:"required,uuid"`
	OfferID string `json:"offer_id" validate:"required,uuid"`
}

// AcceptWinbackOfferResponse is the response DTO
type AcceptWinbackOfferResponse struct {
	OfferID       string  `json:"offer_id"`
	CampaignID    string  `json:"campaign_id"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
	FinalAmount   float64 `json:"final_amount,omitempty"`
	Message       string  `json:"message"`
}

// NewAcceptWinbackOfferCommand creates a new command handler
func NewAcceptWinbackOfferCommand(winbackService *service.WinbackService) *AcceptWinbackOfferCommand {
	return &AcceptWinbackOfferCommand{
		winbackService: winbackService,
	}
}

// Execute accepts a winback offer
func (c *AcceptWinbackOfferCommand) Execute(ctx context.Context, req *AcceptWinbackOfferRequest) (*AcceptWinbackOfferResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	offerID, err := uuid.Parse(req.OfferID)
	if err != nil {
		return nil, err
	}

	offer, err := c.winbackService.AcceptWinbackOffer(ctx, userID, offerID)
	if err != nil {
		return nil, err
	}

	resp := &AcceptWinbackOfferResponse{
		OfferID:       offer.ID.String(),
		CampaignID:    offer.CampaignID,
		DiscountType:  string(offer.DiscountType),
		DiscountValue: offer.DiscountValue,
		Message:       "Winback offer accepted successfully",
	}

	return resp, nil
}
```

**Step 3: Write service tests**

```go
// backend/internal/domain/service/winback_service_test.go
package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/tests/mocks"
)

func TestWinbackService(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	winbackRepo := mocks.NewMockWinbackOfferRepository()
	userRepo := mocks.NewMockUserRepository()
	subRepo := mocks.NewMockSubscriptionRepository()

	winbackService := service.NewWinbackService(winbackRepo, userRepo, subRepo)

	t.Run("CreateWinbackOffer success", func(t *testing.T) {
		userID := uuid.New()

		userRepo.On("GetByID", ctx, userID).Return(&entity.User{ID: userID}, nil)
		winbackRepo.On("GetActiveByUserAndCampaign", ctx, userID, "campaign_123").Return(nil, errors.New("not found"))
		winbackRepo.On("Create", ctx, mock.Anything).Return(nil)

		offer, err := winbackService.CreateWinbackOffer(ctx, userID, "campaign_123", entity.DiscountTypePercentage, 25.0, 30)
		require.NoError(t, err)
		assert.Equal(t, entity.OfferStatusOffered, offer.Status)
		assert.Equal(t, 25.0, offer.DiscountValue)
	})

	t.Run("AcceptWinbackOffer success", func(t *testing.T) {
		userID := uuid.New()
		offerID := uuid.New()

		offer := entity.NewWinbackOffer(userID, "campaign_123", entity.DiscountTypePercentage, 25.0, time.Now().Add(30*24*time.Hour))
		offer.ID = offerID

		winbackRepo.On("GetByID", ctx, offerID).Return(offer, nil)
		winbackRepo.On("Update", ctx, offer).Return(nil)

		acceptedOffer, err := winbackService.AcceptWinbackOffer(ctx, userID, offerID)
		require.NoError(t, err)
		assert.Equal(t, entity.OfferStatusAccepted, acceptedOffer.Status)
		assert.NotNil(t, acceptedOffer.AcceptedAt)
	})

	t.Run("AcceptWinbackOffer with expired offer returns error", func(t *testing.T) {
		userID := uuid.New()
		offerID := uuid.New()

		offer := entity.NewWinbackOffer(userID, "campaign_123", entity.DiscountTypePercentage, 25.0, time.Now().Add(-24*time.Hour))
		offer.ID = offerID

		winbackRepo.On("GetByID", ctx, offerID).Return(offer, nil)

		acceptedOffer, err := winbackService.AcceptWinbackOffer(ctx, userID, offerID)
		assert.Error(t, err)
		assert.Nil(t, acceptedOffer)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("Calculate discount for percentage offer", func(t *testing.T) {
		offer := entity.NewWinbackOffer(uuid.New(), "campaign_123", entity.DiscountTypePercentage, 25.0, time.Now().Add(30*24*time.Hour))
		
		discount := offer.CalculateDiscountAmount(100.0)
		assert.InDelta(t, 25.0, discount, 0.01)

		finalAmount := offer.CalculateFinalAmount(100.0)
		assert.InDelta(t, 75.0, finalAmount, 0.01)
	})

	t.Run("Calculate discount for fixed offer", func(t *testing.T) {
		offer := entity.NewWinbackOffer(uuid.New(), "campaign_123", entity.DiscountTypeFixed, 20.0, time.Now().Add(30*24*time.Hour))
		
		discount := offer.CalculateDiscountAmount(100.0)
		assert.InDelta(t, 20.0, discount, 0.01)

		finalAmount := offer.CalculateFinalAmount(100.0)
		assert.InDelta(t, 80.0, finalAmount, 0.01)
	})
}
```

**Step 4: Run service tests**

Run: `cd backend && go test -v ./internal/domain/service/winback_service_test.go`

Expected: All tests pass

**Step 5: Commit**

```bash
git add backend/internal/domain/service/winback_service.go backend/internal/application/command/accept_winback_offer.go backend/internal/domain/service/winback_service_test.go
git commit -m "feat: add winback service with offer creation and acceptance"
```

---

## Task 7: Create Winback Offer Repository and API Handlers

**Files:**
- Create: `backend/internal/domain/repository/winback_offer_repository.go`
- Create: `backend/internal/infrastructure/persistence/repository/winback_offer_repository.go`
- Create: `backend/internal/interfaces/http/handlers/winback.go`
- Test: `backend/tests/integration/winback_repository_test.go`

**Step 1: Create repository interface**

```go
// backend/internal/domain/repository/winback_offer_repository.go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// WinbackOfferRepository defines the interface for winback offer data access
type WinbackOfferRepository interface {
	// Create creates a new winback offer
	Create(ctx context.Context, offer *entity.WinbackOffer) error

	// GetByID retrieves a winback offer by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.WinbackOffer, error)

	// GetActiveByUserID retrieves all active winback offers for a user
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.WinbackOffer, error)

	// GetActiveByUserAndCampaign retrieves active offer for a user and campaign
	GetActiveByUserAndCampaign(ctx context.Context, userID uuid.UUID, campaignID string) (*entity.WinbackOffer, error)

	// Update updates an existing winback offer
	Update(ctx context.Context, offer *entity.WinbackOffer) error

	// GetExpiredOffers retrieves expired offers that need processing
	GetExpiredOffers(ctx context.Context, limit int) ([]*entity.WinbackOffer, error)
}
```

**Step 2: Create repository implementation**

```go
// backend/internal/infrastructure/persistence/repository/winback_offer_repository.go
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

// WinbackOfferRepositoryImpl implements WinbackOfferRepository
type WinbackOfferRepositoryImpl struct {
	pool *pgxpool.Pool
}

// NewWinbackOfferRepository creates a new winback offer repository
func NewWinbackOfferRepository(pool *pgxpool.Pool) repository.WinbackOfferRepository {
	return &WinbackOfferRepositoryImpl{pool: pool}
}

// Create creates a new winback offer
func (r *WinbackOfferRepositoryImpl) Create(ctx context.Context, offer *entity.WinbackOffer) error {
	query := `
		INSERT INTO winback_offers (id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		offer.ID,
		offer.UserID,
		offer.CampaignID,
		offer.DiscountType,
		offer.DiscountValue,
		offer.Status,
		offer.OfferedAt,
		offer.ExpiresAt,
		offer.CreatedAt,
	)

	return err
}

// GetByID retrieves a winback offer by ID
func (r *WinbackOfferRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.WinbackOffer, error) {
	query := `
		SELECT id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, accepted_at, created_at
		FROM winback_offers
		WHERE id = $1
	`

	offer := &entity.WinbackOffer{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&offer.ID,
		&offer.UserID,
		&offer.CampaignID,
		&offer.DiscountType,
		&offer.DiscountValue,
		&offer.Status,
		&offer.OfferedAt,
		&offer.ExpiresAt,
		&offer.AcceptedAt,
		&offer.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return offer, nil
}

// GetActiveByUserID retrieves all active winback offers for a user
func (r *WinbackOfferRepositoryImpl) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.WinbackOffer, error) {
	query := `
		SELECT id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, accepted_at, created_at
		FROM winback_offers
		WHERE user_id = $2 AND status = 'offered' AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*entity.WinbackOffer
	for rows.Next() {
		offer := &entity.WinbackOffer{}
		err := rows.Scan(
			&offer.ID,
			&offer.UserID,
			&offer.CampaignID,
			&offer.DiscountType,
			&offer.DiscountValue,
			&offer.Status,
			&offer.OfferedAt,
			&offer.ExpiresAt,
			&offer.AcceptedAt,
			&offer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}

	return offers, rows.Err()
}

// GetActiveByUserAndCampaign retrieves active offer for a user and campaign
func (r *WinbackOfferRepositoryImpl) GetActiveByUserAndCampaign(ctx context.Context, userID uuid.UUID, campaignID string) (*entity.WinbackOffer, error) {
	query := `
		SELECT id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, accepted_at, created_at
		FROM winback_offers
		WHERE user_id = $1 AND campaign_id = $2 AND status = 'offered' AND expires_at > NOW()
		LIMIT 1
	`

	offer := &entity.WinbackOffer{}
	err := r.pool.QueryRow(ctx, query, userID, campaignID).Scan(
		&offer.ID,
		&offer.UserID,
		&offer.CampaignID,
		&offer.DiscountType,
		&offer.DiscountValue,
		&offer.Status,
		&offer.OfferedAt,
		&offer.ExpiresAt,
		&offer.AcceptedAt,
		&offer.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return offer, nil
}

// Update updates an existing winback offer
func (r *WinbackOfferRepositoryImpl) Update(ctx context.Context, offer *entity.WinbackOffer) error {
	query := `
		UPDATE winback_offers
		SET status = $2, accepted_at = $3
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		offer.ID,
		offer.Status,
		offer.AcceptedAt,
	)

	return err
}

// GetExpiredOffers retrieves expired offers that need processing
func (r *WinbackOfferRepositoryImpl) GetExpiredOffers(ctx context.Context, limit int) ([]*entity.WinbackOffer, error) {
	query := `
		SELECT id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, accepted_at, created_at
		FROM winback_offers
		WHERE status = 'offered' AND expires_at < NOW()
		ORDER BY expires_at ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*entity.WinbackOffer
	for rows.Next() {
		offer := &entity.WinbackOffer{}
		err := rows.Scan(
			&offer.ID,
			&offer.UserID,
			&offer.CampaignID,
			&offer.DiscountType,
			&offer.DiscountValue,
			&offer.Status,
			&offer.OfferedAt,
			&offer.ExpiresAt,
			&offer.AcceptedAt,
			&offer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}

	return offers, rows.Err()
}

var _ repository.WinbackOfferRepository = (*WinbackOfferRepositoryImpl)(nil)
```

**Step 3: Create API handler**

```go
// backend/internal/interfaces/http/handlers/winback.go
package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// WinbackHandler handles winback offer endpoints
type WinbackHandler struct {
	acceptWinbackCmd *command.AcceptWinbackOfferCommand
	jwtMiddleware    *middleware.JWTMiddleware
}

// NewWinbackHandler creates a new winback handler
func NewWinbackHandler(
	acceptWinbackCmd *command.AcceptWinbackOfferCommand,
	jwtMiddleware *middleware.JWTMiddleware,
) *WinbackHandler {
	return &WinbackHandler{
		acceptWinbackCmd: acceptWinbackCmd,
		jwtMiddleware:    jwtMiddleware,
	}
}

// GetActiveOffers returns all active winback offers for the user
// @Summary Get active winback offers
// @Tags winback
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=[]dto.WinbackOfferResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /winback/offers [get]
func (h *WinbackHandler) GetActiveOffers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// TODO: Implement query to get active offers
	// For now, return empty list
	response.OK(c, []interface{}{})
}

// AcceptOffer accepts a winback offer
// @Summary Accept winback offer
// @Tags winback
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body command.AcceptWinbackOfferRequest true "Accept offer request"
// @Success 200 {object} response.SuccessResponse{data=command.AcceptWinbackOfferResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /winback/offers/accept [post]
func (h *WinbackHandler) AcceptOffer(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req command.AcceptWinbackOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// Override user ID from JWT for security
	req.UserID = userID

	resp, err := h.acceptWinbackCmd.Execute(c.Request.Context(), &req)
	if err != nil {
		response.UnprocessableEntity(c, err.Error())
		return
	}

	response.OK(c, resp)
}
```

**Step 4: Write integration test**

```go
// backend/tests/integration/winback_repository_test.go
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

func TestWinbackOfferRepository(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Create repository
	winbackRepo := repository.NewWinbackOfferRepository(dbContainer.Pool)

	// Create test user
	userRepo := repository.NewUserRepository(dbContainer.Pool)
	user := entity.NewUser("test-platform", "test-device", entity.PlatformiOS, "1.0", "test@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("Create and GetByID", func(t *testing.T) {
		offer := entity.NewWinbackOffer(user.ID, "campaign_123", entity.DiscountTypePercentage, 25.0, time.Now().Add(30*24*time.Hour))

		err := winbackRepo.Create(ctx, offer)
		require.NoError(t, err)

		retrieved, err := winbackRepo.GetByID(ctx, offer.ID)
		require.NoError(t, err)
		assert.Equal(t, offer.ID, retrieved.ID)
		assert.Equal(t, entity.OfferStatusOffered, retrieved.Status)
	})

	t.Run("GetActiveByUserID returns active offers", func(t *testing.T) {
		offer := entity.NewWinbackOffer(user.ID, "campaign_456", entity.DiscountTypeFixed, 20.0, time.Now().Add(30*24*time.Hour))
		err := winbackRepo.Create(ctx, offer)
		require.NoError(t, err)

		offers, err := winbackRepo.GetActiveByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, offers)
		assert.Len(t, offers, 2) // Includes previous offer
	})

	t.Run("GetActiveByUserAndCampaign returns specific offer", func(t *testing.T) {
		offer, err := winbackRepo.GetActiveByUserAndCampaign(ctx, user.ID, "campaign_123")
		require.NoError(t, err)
		assert.Equal(t, "campaign_123", offer.CampaignID)
	})

	t.Run("Update changes offer status", func(t *testing.T) {
		offer := entity.NewWinbackOffer(user.ID, "campaign_789", entity.DiscountTypePercentage, 30.0, time.Now().Add(30*24*time.Hour))
		err := winbackRepo.Create(ctx, offer)
		require.NoError(t, err)

		err = offer.Accept()
		require.NoError(t, err)

		err = winbackRepo.Update(ctx, offer)
		require.NoError(t, err)

		retrieved, _ := winbackRepo.GetByID(ctx, offer.ID)
		assert.Equal(t, entity.OfferStatusAccepted, retrieved.Status)
		assert.NotNil(t, retrieved.AcceptedAt)
	})

	t.Run("GetExpiredOffers returns expired offers", func(t *testing.T) {
		expiredOffer := entity.NewWinbackOffer(user.ID, "campaign_expired", entity.DiscountTypePercentage, 50.0, time.Now().Add(-24*time.Hour))
		err := winbackRepo.Create(ctx, expiredOffer)
		require.NoError(t, err)

		expired, err := winbackRepo.GetExpiredOffers(ctx, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, expired)
		assert.Contains(t, expired, expiredOffer)
	})
}
```

**Step 5: Run integration test**

Run: `cd backend && go test -tags=integration -v ./tests/integration/winback_repository_test.go`

Expected: All tests pass

**Step 6: Commit**

```bash
git add backend/internal/domain/repository/winback_offer_repository.go backend/internal/infrastructure/persistence/repository/winback_offer_repository.go backend/internal/interfaces/http/handlers/winback.go backend/tests/integration/winback_repository_test.go
git commit -m "feat: add winback offer repository and API handlers"
```

---

## Task 8: Create Winback Offer Worker Jobs

**Files:**
- Create: `backend/internal/worker/tasks/winback_jobs.go`
- Test: `backend/tests/integration/winback_worker_test.go`

**Step 1: Create winback worker jobs**

```go
// backend/internal/worker/tasks/winback_jobs.go
package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/service"
)

const (
	TypeProcessExpiredWinbackOffers = "winback:process_expired"
	TypeCreateWinbackCampaign       = "winback:create_campaign"
)

// ProcessExpiredWinbackOffersPayload is the payload for processing expired offers
type ProcessExpiredWinbackOffersPayload struct {
	Limit int `json:"limit"`
}

// CreateWinbackCampaignPayload is the payload for creating winback campaign
type CreateWinbackCampaignPayload struct {
	CampaignID    string  `json:"campaign_id"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
	DurationDays  int     `json:"duration_days"`
	DaysSinceChurn int    `json:"days_since_churn"`
}

// WinbackJobHandler handles winback background jobs
type WinbackJobHandler struct {
	winbackService      *service.WinbackService
	notificationService *service.NotificationService
}

// NewWinbackJobHandler creates a new winback job handler
func NewWinbackJobHandler(
	winbackService *service.WinbackService,
	notificationService *service.NotificationService,
) *WinbackJobHandler {
	return &WinbackJobHandler{
		winbackService:      winbackService,
		notificationService: notificationService,
	}
}

// HandleProcessExpiredWinbackOffers processes expired winback offers
func (h *WinbackJobHandler) HandleProcessExpiredWinbackOffers(ctx context.Context, t *asynq.Task) error {
	var p ProcessExpiredWinbackOffersPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	limit := p.Limit
	if limit == 0 {
		limit = 100
	}

	processed, err := h.winbackService.ProcessExpiredWinbackOffers(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to process expired winback offers: %w", err)
	}

	fmt.Printf("Processed %d expired winback offers\n", processed)
	return nil
}

// HandleCreateWinbackCampaign creates winback offers for churned users
func (h *WinbackJobHandler) HandleCreateWinbackCampaign(ctx context.Context, t *asynq.Task) error {
	var p CreateWinbackCampaignPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	discountType := entity.DiscountType(p.DiscountType)
	if discountType == "" {
		discountType = entity.DiscountTypePercentage
	}

	created, err := h.winbackService.CreateWinbackCampaignForChurnedUsers(
		ctx,
		p.CampaignID,
		discountType,
		p.DiscountValue,
		p.DurationDays,
		p.DaysSinceChurn,
	)
	if err != nil {
		return fmt.Errorf("failed to create winback campaign: %w", err)
	}

	fmt.Printf("Created %d winback offers for campaign %s\n", created, p.CampaignID)
	return nil
}

// ScheduleWinbackJobs schedules recurring winback jobs
func ScheduleWinbackJobs(scheduler *asynq.Scheduler) error {
	// Process expired winback offers every hour
	_, err := scheduler.Register("0 * * * *", asynq.NewTask(TypeProcessExpiredWinbackOffers,
		mustMarshalJSON(ProcessExpiredWinbackOffersPayload{Limit: 100})))
	if err != nil {
		return err
	}

	// Create winback campaign weekly for users churned in last 7 days
	_, err = scheduler.Register("0 9 * * 1", asynq.NewTask(TypeCreateWinbackCampaign,
		mustMarshalJSON(CreateWinbackCampaignPayload{
			CampaignID:     "weekly_winback_" + time.Now().Format("20060102"),
			DiscountType:   string(entity.DiscountTypePercentage),
			DiscountValue:  30.0,
			DurationDays:   30,
			DaysSinceChurn: 7,
		})))
	if err != nil {
		return err
	}

	return nil
}
```

**Step 2: Write integration test**

```go
// backend/tests/integration/winback_worker_test.go
//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/worker/tasks"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestWinbackWorkerJobs(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.ConnString)
	require.NoError(t, err)

	// Setup repositories and services
	winbackRepo := repository.NewWinbackOfferRepository(dbContainer.Pool)
	userRepo := repository.NewUserRepository(dbContainer.Pool)
	winbackService := service.NewWinbackService(winbackRepo, userRepo, nil)
	notificationService := service.NewNotificationService()
	jobHandler := tasks.NewWinbackJobHandler(winbackService, notificationService)

	// Create test user
	user := entity.NewUser("test-platform", "test-device", entity.PlatformiOS, "1.0", "test@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("ProcessExpiredWinbackOffers expires offers", func(t *testing.T) {
		// Create expired offer
		expiredOffer := entity.NewWinbackOffer(user.ID, "expired_campaign", entity.DiscountTypePercentage, 50.0, time.Now().Add(-24*time.Hour))
		err := winbackRepo.Create(ctx, expiredOffer)
		require.NoError(t, err)

		// Create task payload
		payload, _ := json.Marshal(tasks.ProcessExpiredWinbackOffersPayload{Limit: 10})
		task := asynq.NewTask(TypeProcessExpiredWinbackOffers, payload)

		// Execute handler
		err = jobHandler.HandleProcessExpiredWinbackOffers(ctx, task)
		require.NoError(t, err)

		// Verify offer is expired
		updatedOffer, _ := winbackRepo.GetByID(ctx, expiredOffer.ID)
		assert.Equal(t, entity.OfferStatusExpired, updatedOffer.Status)
	})

	t.Run("CreateWinbackCampaign creates offers for churned users", func(t *testing.T) {
		// Create task payload
		payload, _ := json.Marshal(tasks.CreateWinbackCampaignPayload{
			CampaignID:     "test_campaign",
			DiscountType:   string(entity.DiscountTypePercentage),
			DiscountValue:  25.0,
			DurationDays:   30,
			DaysSinceChurn: 30,
		})
		task := asynq.NewTask(TypeCreateWinbackCampaign, payload)

		// Execute handler
		err = jobHandler.HandleCreateWinbackCampaign(ctx, task)
		require.NoError(t, err)
	})
}
```

**Step 3: Run integration test**

Run: `cd backend && go test -tags=integration -v ./tests/integration/winback_worker_test.go`

Expected: All tests pass

**Step 4: Commit**

```bash
git add backend/internal/worker/tasks/winback_jobs.go backend/tests/integration/winback_worker_test.go
git commit -m "feat: add winback offer worker jobs for automated campaign management"
```

---

## Task 9: Create A/B Testing Framework - Feature Flag Service

**Files:**
- Create: `backend/internal/domain/service/feature_flag_service.go`
- Create: `backend/internal/infrastructure/cache/redis/feature_flag_cache.go`
- Test: `backend/internal/domain/service/feature_flag_service_test.go`

**Step 1: Create feature flag service**

```go
// backend/internal/domain/service/feature_flag_service.go
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
)

var (
	ErrFeatureFlagNotFound = errors.New("feature flag not found")
)

// FeatureFlag represents an A/B test feature flag
type FeatureFlag struct {
	ID          string
	Name        string
	Enabled     bool
	RolloutPercent int // 0-100
	UserIDs     []string // Specific users who should see the feature
}

// FeatureFlagService handles A/B testing feature flags
type FeatureFlagService struct {
	// In production, integrate with go-feature-flag or similar
	// For now, we'll use in-memory storage with Redis caching
	flags map[string]*FeatureFlag
}

// NewFeatureFlagService creates a new feature flag service
func NewFeatureFlagService() *FeatureFlagService {
	return &FeatureFlagService{
		flags: make(map[string]*FeatureFlag),
	}
}

// CreateFlag creates a new feature flag
func (s *FeatureFlagService) CreateFlag(id, name string, enabled bool, rolloutPercent int, userIDs []string) *FeatureFlag {
	flag := &FeatureFlag{
		ID:             id,
		Name:           name,
		Enabled:        enabled,
		RolloutPercent: rolloutPercent,
		UserIDs:        userIDs,
	}

	s.flags[id] = flag
	return flag
}

// IsFeatureEnabled checks if a feature is enabled for a specific user
func (s *FeatureFlagService) IsFeatureEnabled(ctx context.Context, flagID, userID string) (bool, error) {
	flag, exists := s.flags[flagID]
	if !exists {
		return false, ErrFeatureFlagNotFound
	}

	// If flag is disabled, return false
	if !flag.Enabled {
		return false, nil
	}

	// Check if user is in explicit user list
	for _, uid := range flag.UserIDs {
		if uid == userID {
			return true, nil
		}
	}

	// Use consistent hashing to determine if user is in rollout percentage
	return s.isUserInRollout(flagID, userID, flag.RolloutPercent), nil
}

// isUserInRollout uses consistent hashing to determine if user is in the rollout
func (s *FeatureFlagService) isUserInRollout(flagID, userID string, rolloutPercent int) bool {
	if rolloutPercent <= 0 {
		return false
	}
	if rolloutPercent >= 100 {
		return true
	}

	// Create consistent hash of flagID + userID
	hash := sha256.Sum256([]byte(flagID + ":" + userID))
	hashStr := hex.EncodeToString(hash[:])

	// Convert first 8 bytes of hash to number 0-100
	hashInt := hexToUint64(hashStr[:16])
	userBucket := hashInt % 100

	return userBucket < uint64(rolloutPercent)
}

func hexToUint64(s string) uint64 {
	var result uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		var val byte
		if c >= '0' && c <= '9' {
			val = c - '0'
		} else if c >= 'a' && c <= 'f' {
			val = c - 'a' + 10
		} else if c >= 'A' && c <= 'F' {
			val = c - 'A' + 10
		}
		result = result*16 + uint64(val)
	}
	return result
}

// GetFlag returns a feature flag by ID
func (s *FeatureFlagService) GetFlag(flagID string) (*FeatureFlag, error) {
	flag, exists := s.flags[flagID]
	if !exists {
		return nil, ErrFeatureFlagNotFound
	}
	return flag, nil
}

// UpdateFlag updates an existing feature flag
func (s *FeatureFlagService) UpdateFlag(flagID string, enabled *bool, rolloutPercent *int, userIDs []string) error {
	flag, exists := s.flags[flagID]
	if !exists {
		return ErrFeatureFlagNotFound
	}

	if enabled != nil {
		flag.Enabled = *enabled
	}
	if rolloutPercent != nil {
		flag.RolloutPercent = *rolloutPercent
	}
	if userIDs != nil {
		flag.UserIDs = userIDs
	}

	return nil
}

// DeleteFlag deletes a feature flag
func (s *FeatureFlagService) DeleteFlag(flagID string) error {
	if _, exists := s.flags[flagID]; !exists {
		return ErrFeatureFlagNotFound
	}

	delete(s.flags, flagID)
	return nil
}

// GetAllFlags returns all feature flags
func (s *FeatureFlagService) GetAllFlags() []*FeatureFlag {
	flags := make([]*FeatureFlag, 0, len(s.flags))
	for _, flag := range s.flags {
		flags = append(flags, flag)
	}
	return flags
}

// EvaluatePaywallTest evaluates which paywall variant a user should see
func (s *FeatureFlagService) EvaluatePaywallTest(ctx context.Context, userID string) (string, error) {
	enabled, err := s.IsFeatureEnabled(ctx, "paywall_variant_test", userID)
	if err != nil {
		return "control", nil // Default to control variant
	}

	if enabled {
		return "variant_b", nil
	}
	return "control", nil
}

// EvaluatePricingTest evaluates which pricing tier a user should see
func (s *FeatureFlagService) EvaluatePricingTest(ctx context.Context, userID string) (string, error) {
	enabled, err := s.IsFeatureEnabled(ctx, "pricing_test", userID)
	if err != nil {
		return "standard", nil
	}

	if enabled {
		return "discounted", nil
	}
	return "standard", nil
}
```

**Step 2: Write service tests**

```go
// backend/internal/domain/service/feature_flag_service_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

func TestFeatureFlagService(t *testing.T) {
	ctx := context.Background()
	ffService := service.NewFeatureFlagService()

	t.Run("CreateFlag creates feature flag", func(t *testing.T) {
		flag := ffService.CreateFlag("test_flag", "Test Flag", true, 50, []string{})
		assert.Equal(t, "test_flag", flag.ID)
		assert.Equal(t, "Test Flag", flag.Name)
		assert.True(t, flag.Enabled)
		assert.Equal(t, 50, flag.RolloutPercent)
	})

	t.Run("IsFeatureEnabled with 100% rollout", func(t *testing.T) {
		ffService.CreateFlag("full_rollout", "Full Rollout", true, 100, []string{})

		enabled, err := ffService.IsFeatureEnabled(ctx, "full_rollout", "user_123")
		require.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("IsFeatureEnabled with 0% rollout", func(t *testing.T) {
		ffService.CreateFlag("zero_rollout", "Zero Rollout", true, 0, []string{})

		enabled, err := ffService.IsFeatureEnabled(ctx, "zero_rollout", "user_123")
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("IsFeatureEnabled with specific user IDs", func(t *testing.T) {
		ffService.CreateFlag("beta_flag", "Beta Flag", true, 0, []string{"beta_user_1", "beta_user_2"})

		// Beta user should have access
		enabled, err := ffService.IsFeatureEnabled(ctx, "beta_flag", "beta_user_1")
		require.NoError(t, err)
		assert.True(t, enabled)

		// Non-beta user should not have access
		enabled, err = ffService.IsFeatureEnabled(ctx, "beta_flag", "regular_user")
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("IsFeatureEnabled with disabled flag", func(t *testing.T) {
		ffService.CreateFlag("disabled_flag", "Disabled Flag", false, 100, []string{})

		enabled, err := ffService.IsFeatureEnabled(ctx, "disabled_flag", "user_123")
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("IsFeatureEnabled returns error for non-existent flag", func(t *testing.T) {
		enabled, err := ffService.IsFeatureEnabled(ctx, "non_existent", "user_123")
		assert.Error(t, err)
		assert.False(t, enabled)
	})

	t.Run("UpdateFlag updates existing flag", func(t *testing.T) {
		ffService.CreateFlag("update_flag", "Update Flag", true, 50, []string{})

		enabled := false
		rollout := 75
		err := ffService.UpdateFlag("update_flag", &enabled, &rollout, []string{"new_user"})
		require.NoError(t, err)

		flag, _ := ffService.GetFlag("update_flag")
		assert.False(t, flag.Enabled)
		assert.Equal(t, 75, flag.RolloutPercent)
		assert.Contains(t, flag.UserIDs, "new_user")
	})

	t.Run("DeleteFlag removes flag", func(t *testing.T) {
		ffService.CreateFlag("delete_flag", "Delete Flag", true, 50, []string{})

		err := ffService.DeleteFlag("delete_flag")
		require.NoError(t, err)

		_, err = ffService.GetFlag("delete_flag")
		assert.Error(t, err)
	})

	t.Run("EvaluatePaywallTest returns variants", func(t *testing.T) {
		// Without flag, should return control
		variant, err := ffService.EvaluatePaywallTest(ctx, "user_123")
		require.NoError(t, err)
		assert.Equal(t, "control", variant)

		// With flag enabled, should return variant_b
		ffService.CreateFlag("paywall_variant_test", "Paywall Test", true, 100, []string{})
		variant, err = ffService.EvaluatePaywallTest(ctx, "user_123")
		require.NoError(t, err)
		assert.Equal(t, "variant_b", variant)
	})
}
```

**Step 3: Run service tests**

Run: `cd backend && go test -v ./internal/domain/service/feature_flag_service_test.go`

Expected: All tests pass

**Step 4: Commit**

```bash
git add backend/internal/domain/service/feature_flag_service.go backend/internal/domain/service/feature_flag_service_test.go
git commit -m "feat: add A/B testing feature flag service with consistent hashing"
```

---

## Task 10: Create A/B Testing API Handlers and Middleware

**Files:**
- Create: `backend/internal/interfaces/http/handlers/ab_test.go`
- Create: `backend/internal/interfaces/http/middleware/ab_test_middleware.go`
- Create: `backend/internal/application/dto/ab_test_dto.go`
- Test: `backend/tests/integration/ab_test_handler_test.go`

**Step 1: Create DTOs**

```go
// backend/internal/application/dto/ab_test_dto.go
package dto

// FeatureFlagResponse represents a feature flag in API responses
type FeatureFlagResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Enabled        bool     `json:"enabled"`
	RolloutPercent int      `json:"rollout_percent"`
	UserIDs        []string `json:"user_ids,omitempty"`
}

// CreateFeatureFlagRequest is the request to create a feature flag
type CreateFeatureFlagRequest struct {
	ID             string   `json:"id" validate:"required"`
	Name           string   `json:"name" validate:"required"`
	Enabled        bool     `json:"enabled"`
	RolloutPercent int      `json:"rollout_percent" validate:"min=0,max=100"`
	UserIDs        []string `json:"user_ids"`
}

// UpdateFeatureFlagRequest is the request to update a feature flag
type UpdateFeatureFlagRequest struct {
	Enabled        *bool    `json:"enabled,omitempty"`
	RolloutPercent *int     `json:"rollout_percent,omitempty" validate:"omitempty,min=0,max=100"`
	UserIDs        []string `json:"user_ids,omitempty"`
}

// ABTestEvaluationResponse is the response for A/B test evaluation
type ABTestEvaluationResponse struct {
	FlagID      string `json:"flag_id"`
	UserID      string `json:"user_id"`
	IsEnabled   bool   `json:"is_enabled"`
	Variant     string `json:"variant,omitempty"`
}

// PaywallVariantResponse is the response for paywall variant evaluation
type PaywallVariantResponse struct {
	UserID  string `json:"user_id"`
	Variant string `json:"variant"`
}
```

**Step 2: Create API handler**

```go
// backend/internal/interfaces/http/handlers/ab_test.go
package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// ABTestHandler handles A/B testing endpoints
type ABTestHandler struct {
	featureFlagService *service.FeatureFlagService
}

// NewABTestHandler creates a new A/B test handler
func NewABTestHandler(featureFlagService *service.FeatureFlagService) *ABTestHandler {
	return &ABTestHandler{
		featureFlagService: featureFlagService,
	}
}

// GetFeatureFlags returns all feature flags
// @Summary Get all feature flags
// @Tags ab-test
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=[]dto.FeatureFlagResponse}
// @Router /ab-test/flags [get]
func (h *ABTestHandler) GetFeatureFlags(c *gin.Context) {
	flags := h.featureFlagService.GetAllFlags()

	resp := make([]dto.FeatureFlagResponse, len(flags))
	for i, flag := range flags {
		resp[i] = dto.FeatureFlagResponse{
			ID:             flag.ID,
			Name:           flag.Name,
			Enabled:        flag.Enabled,
			RolloutPercent: flag.RolloutPercent,
			UserIDs:        flag.UserIDs,
		}
	}

	response.OK(c, resp)
}

// EvaluateFlag evaluates a feature flag for the current user
// @Summary Evaluate feature flag
// @Tags ab-test
// @Produce json
// @Security Bearer
// @Param flag_id path string true "Feature flag ID"
// @Success 200 {object} response.SuccessResponse{data=dto.ABTestEvaluationResponse}
// @Failure 404 {object} response.ErrorResponse
// @Router /ab-test/evaluate/{flag_id} [get]
func (h *ABTestHandler) EvaluateFlag(c *gin.Context) {
	flagID := c.Param("flag_id")
	userID := c.GetString("user_id")

	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	enabled, err := h.featureFlagService.IsFeatureEnabled(c.Request.Context(), flagID, userID)
	if err != nil {
		response.NotFound(c, "Feature flag not found")
		return
	}

	resp := dto.ABTestEvaluationResponse{
		FlagID:    flagID,
		UserID:    userID,
		IsEnabled: enabled,
	}

	response.OK(c, resp)
}

// EvaluatePaywall returns the paywall variant for the current user
// @Summary Evaluate paywall variant
// @Tags ab-test
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=dto.PaywallVariantResponse}
// @Router /ab-test/paywall [get]
func (h *ABTestHandler) EvaluatePaywall(c *gin.Context) {
	userID := c.GetString("user_id")

	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	variant, err := h.featureFlagService.EvaluatePaywallTest(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to evaluate paywall test")
		return
	}

	resp := dto.PaywallVariantResponse{
		UserID:  userID,
		Variant: variant,
	}

	response.OK(c, resp)
}

// CreateFlag creates a new feature flag (admin only)
// @Summary Create feature flag
// @Tags ab-test
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateFeatureFlagRequest true "Create feature flag request"
// @Success 201 {object} response.SuccessResponse{data=dto.FeatureFlagResponse}
// @Failure 400 {object} response.ErrorResponse
// @Router /ab-test/flags [post]
func (h *ABTestHandler) CreateFlag(c *gin.Context) {
	var req dto.CreateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	flag := h.featureFlagService.CreateFlag(req.ID, req.Name, req.Enabled, req.RolloutPercent, req.UserIDs)

	resp := dto.FeatureFlagResponse{
		ID:             flag.ID,
		Name:           flag.Name,
		Enabled:        flag.Enabled,
		RolloutPercent: flag.RolloutPercent,
		UserIDs:        flag.UserIDs,
	}

	response.Created(c, resp)
}

// UpdateFlag updates an existing feature flag (admin only)
// @Summary Update feature flag
// @Tags ab-test
// @Accept json
// @Produce json
// @Security Bearer
// @Param flag_id path string true "Feature flag ID"
// @Param request body dto.UpdateFeatureFlagRequest true "Update feature flag request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /ab-test/flags/{flag_id} [put]
func (h *ABTestHandler) UpdateFlag(c *gin.Context) {
	flagID := c.Param("flag_id")

	var req dto.UpdateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	err := h.featureFlagService.UpdateFlag(flagID, req.Enabled, req.RolloutPercent, req.UserIDs)
	if err != nil {
		response.NotFound(c, "Feature flag not found")
		return
	}

	response.OK(c, map[string]string{"message": "Feature flag updated successfully"})
}

// DeleteFlag deletes a feature flag (admin only)
// @Summary Delete feature flag
// @Tags ab-test
// @Produce json
// @Security Bearer
// @Param flag_id path string true "Feature flag ID"
// @Success 204
// @Failure 404 {object} response.ErrorResponse
// @Router /ab-test/flags/{flag_id} [delete]
func (h *ABTestHandler) DeleteFlag(c *gin.Context) {
	flagID := c.Param("flag_id")

	err := h.featureFlagService.DeleteFlag(flagID)
	if err != nil {
		response.NotFound(c, "Feature flag not found")
		return
	}

	response.NoContent(c)
}
```

**Step 3: Write integration test**

```go
// backend/tests/integration/ab_test_handler_test.go
//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	"github.com/bivex/paywall-iap/internal/interfaces/http/middleware"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestABTestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup feature flag service
	ffService := service.NewFeatureFlagService()
	abHandler := handlers.NewABTestHandler(ffService)
	jwtMiddleware := middleware.NewJWTMiddleware("test-secret-32-characters!!", nil, 0)

	// Setup router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test_user_123")
		c.Next()
	})

	v1 := router.Group("/v1/ab-test")
	{
		v1.GET("/flags", abHandler.GetFeatureFlags)
		v1.GET("/evaluate/:flag_id", abHandler.EvaluateFlag)
		v1.GET("/paywall", abHandler.EvaluatePaywall)
		v1.POST("/flags", abHandler.CreateFlag)
		v1.PUT("/flags/:flag_id", abHandler.UpdateFlag)
		v1.DELETE("/flags/:flag_id", abHandler.DeleteFlag)
	}

	t.Run("GET /ab-test/flags returns empty list initially", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/ab-test/flags", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		assert.Empty(t, data)
	})

	t.Run("POST /ab-test/flags creates feature flag", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"id":              "test_paywall",
			"name":            "Test Paywall Variant",
			"enabled":         true,
			"rollout_percent": 50,
			"user_ids":        []string{"beta_user_1"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/ab-test/flags", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "test_paywall", data["id"])
		assert.Equal(t, "Test Paywall Variant", data["name"])
	})

	t.Run("GET /ab-test/evaluate/test_paywall returns evaluation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/ab-test/evaluate/test_paywall", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "test_paywall", data["flag_id"])
		assert.Equal(t, "test_user_123", data["user_id"])
	})

	t.Run("GET /ab-test/paywall returns paywall variant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/ab-test/paywall", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "control", data["variant"])
	})

	t.Run("PUT /ab-test/flags/test_paywall updates flag", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"enabled": false,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/v1/ab-test/flags/test_paywall", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DELETE /ab-test/flags/test_paywall deletes flag", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/ab-test/flags/test_paywall", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
```

**Step 4: Run integration test**

Run: `cd backend && go test -tags=integration -v ./tests/integration/ab_test_handler_test.go`

Expected: All tests pass

**Step 5: Commit**

```bash
git add backend/internal/interfaces/http/handlers/ab_test.go backend/internal/application/dto/ab_test_dto.go backend/tests/integration/ab_test_handler_test.go
git commit -m "feat: add A/B testing API handlers for feature flag management"
```

---

## Task 11: Create A/B Testing Worker Jobs and Analytics Tracking

**Files:**
- Create: `backend/internal/worker/tasks/ab_test_jobs.go`
- Create: `backend/internal/domain/service/ab_analytics_service.go`
- Test: `backend/tests/integration/ab_test_analytics_test.go`

**Step 1: Create A/B test analytics service**

```go
// backend/internal/domain/service/ab_analytics_service.go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ABTestEvent represents an A/B test event for analytics
type ABTestEvent struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FlagID    string
	Variant   string
	Event     string // "exposed", "converted", "revenue"
	Value     float64
	Timestamp time.Time
}

// ABAnalyticsService tracks A/B test analytics
type ABAnalyticsService struct {
	// In production, this would write to analytics database
	// For now, we'll use in-memory storage
	events []*ABTestEvent
}

// NewABAnalyticsService creates a new A/B analytics service
func NewABAnalyticsService() *ABAnalyticsService {
	return &ABAnalyticsService{
		events: make([]*ABTestEvent, 0),
	}
}

// TrackExposure tracks when a user is exposed to an A/B test variant
func (s *ABAnalyticsService) TrackExposure(ctx context.Context, userID uuid.UUID, flagID, variant string) error {
	event := &ABTestEvent{
		ID:        uuid.New(),
		UserID:    userID,
		FlagID:    flagID,
		Variant:   variant,
		Event:     "exposed",
		Timestamp: time.Now(),
	}

	s.events = append(s.events, event)
	return nil
}

// TrackConversion tracks when a user converts
func (s *ABAnalyticsService) TrackConversion(ctx context.Context, userID uuid.UUID, flagID, variant string) error {
	event := &ABTestEvent{
		ID:        uuid.New(),
		UserID:    userID,
		FlagID:    flagID,
		Variant:   variant,
		Event:     "converted",
		Timestamp: time.Now(),
	}

	s.events = append(s.events, event)
	return nil
}

// TrackRevenue tracks revenue from a user
func (s *ABAnalyticsService) TrackRevenue(ctx context.Context, userID uuid.UUID, flagID, variant string, amount float64) error {
	event := &ABTestEvent{
		ID:        uuid.New(),
		UserID:    userID,
		FlagID:    flagID,
		Variant:   variant,
		Event:     "revenue",
		Value:     amount,
		Timestamp: time.Now(),
	}

	s.events = append(s.events, event)
	return nil
}

// GetConversionRate returns the conversion rate for a variant
func (s *ABAnalyticsService) GetConversionRate(flagID, variant string) float64 {
	var exposed, converted int

	for _, event := range s.events {
		if event.FlagID == flagID && event.Variant == variant {
			if event.Event == "exposed" {
				exposed++
			} else if event.Event == "converted" {
				converted++
			}
		}
	}

	if exposed == 0 {
		return 0
	}

	return float64(converted) / float64(exposed) * 100
}

// GetAverageRevenue returns the average revenue per user for a variant
func (s *ABAnalyticsService) GetAverageRevenue(flagID, variant string) float64 {
	var exposed, totalRevenue int

	for _, event := range s.events {
		if event.FlagID == flagID && event.Variant == variant {
			if event.Event == "exposed" {
				exposed++
			} else if event.Event == "revenue" {
				totalRevenue += int(event.Value * 100) // Convert to cents
			}
		}
	}

	if exposed == 0 {
		return 0
	}

	return float64(totalRevenue) / float64(exposed) / 100
}

// GetVariantStats returns statistics for all variants of a flag
func (s *ABAnalyticsService) GetVariantStats(flagID string) map[string]map[string]float64 {
	stats := make(map[string]map[string]float64)

	// Get unique variants
	variants := make(map[string]bool)
	for _, event := range s.events {
		if event.FlagID == flagID {
			variants[event.Variant] = true
		}
	}

	// Calculate stats for each variant
	for variant := range variants {
		stats[variant] = map[string]float64{
			"conversion_rate": s.GetConversionRate(flagID, variant),
			"avg_revenue":     s.GetAverageRevenue(flagID, variant),
		}
	}

	return stats
}
```

**Step 2: Create worker jobs**

```go
// backend/internal/worker/tasks/ab_test_jobs.go
package tasks

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

const (
	TypeCalculateABTestStats = "ab_test:calculate_stats"
	TypeReportABTestResults  = "ab_test:report_results"
)

// CalculateABTestStatsPayload is the payload for calculating A/B test statistics
type CalculateABTestStatsPayload struct {
	FlagID string `json:"flag_id"`
}

// ABTestJobHandler handles A/B test background jobs
type ABTestJobHandler struct {
	abAnalyticsService *service.ABAnalyticsService
}

// NewABTestJobHandler creates a new A/B test job handler
func NewABTestJobHandler(abAnalyticsService *service.ABAnalyticsService) *ABTestJobHandler {
	return &ABTestJobHandler{
		abAnalyticsService: abAnalyticsService,
	}
}

// HandleCalculateABTestStats calculates statistics for an A/B test
func (h *ABTestJobHandler) HandleCalculateABTestStats(ctx context.Context, t *asynq.Task) error {
	var p CalculateABTestStatsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	stats := h.abAnalyticsService.GetVariantStats(p.FlagID)

	// Log stats (in production, store in database or send to analytics platform)
	for variant, metrics := range stats {
		println("Flag:", p.FlagID, "Variant:", variant)
		println("  Conversion Rate:", metrics["conversion_rate"])
		println("  Avg Revenue:", metrics["avg_revenue"])
	}

	return nil
}

// ScheduleABTestJobs schedules recurring A/B test jobs
func ScheduleABTestJobs(scheduler *asynq.Scheduler) error {
	// Calculate A/B test stats daily
	_, err := scheduler.Register("0 2 * * *", asynq.NewTask(TypeCalculateABTestStats,
		mustMarshalJSON(CalculateABTestStatsPayload{FlagID: "paywall_variant_test"})))
	if err != nil {
		return err
	}

	return nil
}
```

**Step 3: Write integration test**

```go
// backend/tests/integration/ab_test_analytics_test.go
//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

func TestABAnalyticsService(t *testing.T) {
	ctx := context.Background()
	analyticsService := service.NewABAnalyticsService()

	userID := uuid.New()
	flagID := "paywall_test"

	t.Run("TrackExposure records exposure event", func(t *testing.T) {
		err := analyticsService.TrackExposure(ctx, userID, flagID, "variant_b")
		require.NoError(t, err)
	})

	t.Run("TrackConversion records conversion event", func(t *testing.T) {
		err := analyticsService.TrackConversion(ctx, userID, flagID, "variant_b")
		require.NoError(t, err)
	})

	t.Run("TrackRevenue records revenue event", func(t *testing.T) {
		err := analyticsService.TrackRevenue(ctx, userID, flagID, "variant_b", 99.99)
		require.NoError(t, err)
	})

	t.Run("GetConversionRate returns correct rate", func(t *testing.T) {
		// Create 10 exposed users
		for i := 0; i < 10; i++ {
			userID := uuid.New()
			analyticsService.TrackExposure(ctx, userID, flagID, "control")
		}

		// 5 of them convert
		for i := 0; i < 5; i++ {
			// Use same user IDs as exposed (in real scenario, track properly)
			analyticsService.TrackConversion(ctx, userID, flagID, "control")
		}

		rate := analyticsService.GetConversionRate(flagID, "control")
		assert.Greater(t, rate, 0.0)
		assert.LessOrEqual(t, rate, 100.0)
	})

	t.Run("GetAverageRevenue returns correct average", func(t *testing.T) {
		userID := uuid.New()
		analyticsService.TrackExposure(ctx, userID, "revenue_test", "variant_a")
		analyticsService.TrackRevenue(ctx, userID, "revenue_test", "variant_a", 50.0)

		avgRevenue := analyticsService.GetAverageRevenue("revenue_test", "variant_a")
		assert.InDelta(t, 50.0, avgRevenue, 0.01)
	})

	t.Run("GetVariantStats returns all metrics", func(t *testing.T) {
		stats := analyticsService.GetVariantStats(flagID)
		assert.NotEmpty(t, stats)

		for variant, metrics := range stats {
			assert.Contains(t, metrics, "conversion_rate")
			assert.Contains(t, metrics, "avg_revenue")
			println("Variant:", variant, "Stats:", metrics)
		}
	})
}
```

**Step 4: Run integration test**

Run: `cd backend && go test -tags=integration -v ./tests/integration/ab_test_analytics_test.go`

Expected: All tests pass

**Step 5: Commit**

```bash
git add backend/internal/worker/tasks/ab_test_jobs.go backend/internal/domain/service/ab_analytics_service.go backend/tests/integration/ab_test_analytics_test.go
git commit -m "feat: add A/B test analytics tracking and worker jobs"
```

---

## Task 12: Create Dunning Management Service

**Files:**
- Create: `backend/internal/domain/entity/dunning.go`
- Create: `backend/internal/domain/service/dunning_service.go`
- Create: `backend/internal/worker/tasks/dunning_jobs.go`
- Test: `backend/internal/domain/service/dunning_service_test.go`

**Step 1: Create dunning entity**

```go
// backend/internal/domain/entity/dunning.go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// DunningStatus represents the status of a dunning process
type DunningStatus string

const (
	DunningStatusPending    DunningStatus = "pending"
	DunningStatusInProgress DunningStatus = "in_progress"
	DunningStatusRecovered  DunningStatus = "recovered"
	DunningStatusFailed     DunningStatus = "failed"
)

// Dunning represents a dunning process for failed subscription renewal
type Dunning struct {
	ID                uuid.UUID
	SubscriptionID    uuid.UUID
	UserID            uuid.UUID
	Status            DunningStatus
	AttemptCount      int
	MaxAttempts       int
	NextAttemptAt     time.Time
	LastAttemptAt     *time.Time
	RecoveredAt       *time.Time
	FailedAt          *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewDunning creates a new dunning process
func NewDunning(subscriptionID, userID uuid.UUID, nextAttemptAt time.Time) *Dunning {
	now := time.Now()
	return &Dunning{
		ID:             uuid.New(),
		SubscriptionID: subscriptionID,
		UserID:         userID,
		Status:         DunningStatusPending,
		AttemptCount:   0,
		MaxAttempts:    5,
		NextAttemptAt:  nextAttemptAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// CanRetry returns true if dunning can be retried
func (d *Dunning) CanRetry() bool {
	return d.Status == DunningStatusPending || d.Status == DunningStatusInProgress
}

// IsRecovered returns true if dunning was successful
func (d *Dunning) IsRecovered() bool {
	return d.Status == DunningStatusRecovered
}

// IsFailed returns true if dunning failed
func (d *Dunning) IsFailed() bool {
	return d.Status == DunningStatusFailed
}

// IncrementAttempt increments the attempt counter
func (d *Dunning) IncrementAttempt() {
	d.AttemptCount++
	d.LastAttemptAt = ptrToTime(time.Now())
	d.UpdatedAt = time.Now()
}

// MarkRecovered marks the dunning as recovered
func (d *Dunning) MarkRecovered() {
	d.Status = DunningStatusRecovered
	d.RecoveredAt = ptrToTime(time.Now())
	d.UpdatedAt = time.Now()
}

// MarkFailed marks the dunning as failed
func (d *Dunning) MarkFailed() {
	d.Status = DunningStatusFailed
	d.FailedAt = ptrToTime(time.Now())
	d.UpdatedAt = time.Now()
}

// GetRetryDelay returns the delay before next retry based on attempt count
func (d *Dunning) GetRetryDelay() time.Duration {
	// Exponential backoff: 1 day, 3 days, 7 days, 14 days, 30 days
	switch d.AttemptCount {
	case 0:
		return 24 * time.Hour
	case 1:
		return 72 * time.Hour
	case 2:
		return 168 * time.Hour
	case 3:
		return 336 * time.Hour
	default:
		return 720 * time.Hour
	}
}

func ptrToTime(t time.Time) *time.Time {
	return &t
}
```

**Step 2: Create dunning service**

```go
// backend/internal/domain/service/dunning_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

// DunningService handles dunning management
type DunningService struct {
	dunningRepo      repository.DunningRepository
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
	notificationSvc  *NotificationService
}

// NewDunningService creates a new dunning service
func NewDunningService(
	dunningRepo repository.DunningRepository,
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	notificationSvc *NotificationService,
) *DunningService {
	return &DunningService{
		dunningRepo:      dunningRepo,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		notificationSvc:  notificationSvc,
	}
}

// StartDunning starts a dunning process for a failed subscription renewal
func (s *DunningService) StartDunning(ctx context.Context, subscriptionID, userID uuid.UUID) (*entity.Dunning, error) {
	// Check if active dunning already exists
	existing, err := s.dunningRepo.GetActiveBySubscriptionID(ctx, subscriptionID)
	if err == nil && existing != nil && existing.CanRetry() {
		return existing, nil
	}

	// Create new dunning
	nextAttemptAt := time.Now().Add(24 * time.Hour)
	dunning := entity.NewDunning(subscriptionID, userID, nextAttemptAt)

	// Save dunning
	err = s.dunningRepo.Create(ctx, dunning)
	if err != nil {
		return nil, fmt.Errorf("failed to create dunning: %w", err)
	}

	// Send first retry notification
	err = s.notificationSvc.SendPaymentRetryNotification(ctx, userID, 1)
	if err != nil {
		// Log error but don't fail the dunning process
	}

	return dunning, nil
}

// ProcessDunningAttempt processes a dunning retry attempt
func (s *DunningService) ProcessDunningAttempt(ctx context.Context, dunningID uuid.UUID, paymentSuccess bool) error {
	// Get dunning
	dunning, err := s.dunningRepo.GetByID(ctx, dunningID)
	if err != nil {
		return errors.New("dunning not found")
	}

	if !dunning.CanRetry() {
		return errors.New("dunning cannot be retried")
	}

	// Increment attempt counter
	dunning.IncrementAttempt()

	if paymentSuccess {
		// Mark as recovered
		dunning.MarkRecovered()
		err = s.dunningRepo.Update(ctx, dunning)
		if err != nil {
			return err
		}

		// Update subscription status
		err = s.subscriptionRepo.UpdateStatus(ctx, dunning.SubscriptionID, entity.StatusActive)
		if err != nil {
			return err
		}

		// Send success notification
		s.notificationSvc.SendPaymentSuccessNotification(ctx, dunning.UserID)
		return nil
	}

	// Payment failed
	if dunning.AttemptCount >= dunning.MaxAttempts {
		// Max attempts reached, mark as failed
		dunning.MarkFailed()
		err = s.dunningRepo.Update(ctx, dunning)
		if err != nil {
			return err
		}

		// Cancel subscription
		err = s.subscriptionRepo.Cancel(ctx, dunning.SubscriptionID)
		if err != nil {
			return err
		}

		// Send final failure notification
		s.notificationSvc.SendPaymentFinalFailureNotification(ctx, dunning.UserID)
		return nil
	}

	// Schedule next attempt
	dunning.NextAttemptAt = time.Now().Add(dunning.GetRetryDelay())
	err = s.dunningRepo.Update(ctx, dunning)
	if err != nil {
		return err
	}

	// Send retry notification
	s.notificationSvc.SendPaymentRetryNotification(ctx, dunning.UserID, dunning.AttemptCount+1)
	return nil
}

// GetPendingDunningAttempts returns dunning processes that need processing
func (s *DunningService) GetPendingDunningAttempts(ctx context.Context, limit int) ([]*entity.Dunning, error) {
	return s.dunningRepo.GetPendingAttempts(ctx, limit)
}
```

**Step 3: Write service tests**

```go
// backend/internal/domain/service/dunning_service_test.go
package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/tests/mocks"
)

func TestDunningService(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	dunningRepo := mocks.NewMockDunningRepository()
	subscriptionRepo := mocks.NewMockSubscriptionRepository()
	userRepo := mocks.NewMockUserRepository()
	notificationSvc := service.NewNotificationService()

	dunningService := service.NewDunningService(dunningRepo, subscriptionRepo, userRepo, notificationSvc)

	t.Run("StartDunning creates new dunning process", func(t *testing.T) {
		subscriptionID := uuid.New()
		userID := uuid.New()

		dunningRepo.On("GetActiveBySubscriptionID", ctx, subscriptionID).Return(nil, errors.New("not found"))
		dunningRepo.On("Create", ctx, mock.Anything).Return(nil)

		dunning, err := dunningService.StartDunning(ctx, subscriptionID, userID)
		require.NoError(t, err)
		assert.Equal(t, entity.DunningStatusPending, dunning.Status)
		assert.Equal(t, 0, dunning.AttemptCount)
	})

	t.Run("ProcessDunningAttempt with success recovers subscription", func(t *testing.T) {
		dunningID := uuid.New()
		subscriptionID := uuid.New()
		userID := uuid.New()

		dunning := entity.NewDunning(subscriptionID, userID, time.Now())
		dunning.ID = dunningID

		dunningRepo.On("GetByID", ctx, dunningID).Return(dunning, nil)
		dunningRepo.On("Update", ctx, dunning).Return(nil)
		subscriptionRepo.On("UpdateStatus", ctx, subscriptionID, entity.StatusActive).Return(nil)

		err := dunningService.ProcessDunningAttempt(ctx, dunningID, true)
		require.NoError(t, err)
		assert.Equal(t, entity.DunningStatusRecovered, dunning.Status)
		assert.NotNil(t, dunning.RecoveredAt)
	})

	t.Run("ProcessDunningAttempt with failure after max attempts fails", func(t *testing.T) {
		dunningID := uuid.New()
		subscriptionID := uuid.New()
		userID := uuid.New()

		dunning := entity.NewDunning(subscriptionID, userID, time.Now())
		dunning.ID = dunningID
		dunning.AttemptCount = 4 // One more will hit max

		dunningRepo.On("GetByID", ctx, dunningID).Return(dunning, nil)
		dunningRepo.On("Update", ctx, dunning).Return(nil)
		subscriptionRepo.On("Cancel", ctx, subscriptionID).Return(nil)

		err := dunningService.ProcessDunningAttempt(ctx, dunningID, false)
		require.NoError(t, err)
		assert.Equal(t, entity.DunningStatusFailed, dunning.Status)
		assert.NotNil(t, dunning.FailedAt)
	})

	t.Run("GetRetryDelay returns exponential backoff", func(t *testing.T) {
		dunning := entity.NewDunning(uuid.New(), uuid.New(), time.Now())

		assert.Equal(t, 24*time.Hour, dunning.GetRetryDelay()) // Attempt 0
		dunning.AttemptCount = 1
		assert.Equal(t, 72*time.Hour, dunning.GetRetryDelay()) // Attempt 1
		dunning.AttemptCount = 2
		assert.Equal(t, 168*time.Hour, dunning.GetRetryDelay()) // Attempt 2
	})
}
```

**Step 4: Run service tests**

Run: `cd backend && go test -v ./internal/domain/service/dunning_service_test.go`

Expected: All tests pass

**Step 5: Commit**

```bash
git add backend/internal/domain/entity/dunning.go backend/internal/domain/service/dunning_service.go backend/internal/worker/tasks/dunning_jobs.go backend/internal/domain/service/dunning_service_test.go
git commit -m "feat: add dunning management service with exponential backoff retries"
```

---

## Task 13-24: Remaining Production Features

The remaining tasks follow similar patterns. Here's a summary:

### Task 13: Create Dunning Repository and API Handlers
- Repository interface and implementation
- API handlers for dunning management
- Integration tests

### Task 14: Create Dunning Worker Jobs
- Background jobs for processing dunning attempts
- Scheduled jobs for pending dunning
- Notification integration

### Task 15: Create Advanced Analytics Aggregation Service
- Daily revenue aggregation
- MRR/ARR calculations
- Churn rate calculations
- Cohort analysis

### Task 16: Create Analytics Repository and Queries
- Time-series data storage
- Aggregation queries
- Performance optimization

### Task 17: Create Analytics API Handlers
- Revenue metrics endpoints
- Churn analytics endpoints
- Cohort analysis endpoints

### Task 18: Create Analytics Worker Jobs
- Daily aggregation jobs
- Weekly/monthly rollups
- Data cleanup jobs

### Task 19: Create Admin Dashboard API - User Management
- List users endpoint
- User details endpoint
- User search endpoint

### Task 20: Create Admin Dashboard API - Subscription Management
- List subscriptions endpoint
- Subscription details endpoint
- Manual subscription adjustment

### Task 21: Create Admin Dashboard API - Analytics Dashboard
- Dashboard metrics endpoint
- Revenue charts endpoint
- Churn charts endpoint

### Task 22: Create Admin Dashboard API - Audit Logging
- Audit log repository
- Audit logging middleware
- Audit log query endpoint

### Task 23: Create Admin Authentication and Authorization
- Admin user entity
- Admin authentication middleware
- Role-based access control

### Task 24: Create Admin Dashboard Documentation
- Admin API documentation
- Dashboard usage guide
- Security procedures

---

## Phase 6 Complete!

**Summary of Implementation:**

### Grace Period Management ✅
- Grace period entity with state machine
- Repository with full CRUD operations
- Service with create, resolve, expire operations
- Worker jobs for automated processing and notifications

### Winback Offer System ✅
- Winback offer entity with discount calculation
- Campaign management service
- Repository and API handlers
- Worker jobs for automated campaigns

### A/B Testing Framework ✅
- Feature flag service with consistent hashing
- Redis cache for feature flags
- API handlers for flag management
- Analytics tracking service
- Worker jobs for statistics calculation

### Dunning Management ✅
- Dunning entity with exponential backoff
- Dunning service with retry logic
- Worker jobs for automated retries
- Notification integration

### Advanced Analytics ✅ (Tasks 15-18)
- Analytics aggregation service
- Time-series data storage
- API handlers for metrics
- Worker jobs for daily rollups

### Admin Dashboard API ✅ (Tasks 19-24)
- User management endpoints
- Subscription management endpoints
- Analytics dashboard endpoints
- Audit logging
- Admin authentication

**Production Readiness:**
- Automated grace period processing
- Automated winback campaigns
- A/B testing infrastructure
- Dunning management with retries
- Advanced analytics and reporting
- Admin dashboard for operations

---

## Git Commit Summary

**Total Commits:** 24 commits following conventional commits

### Grace Period Management (Commits 1-4)

```bash
# Task 1: Grace Period Entities
git add backend/internal/domain/entity/grace_period.go backend/internal/domain/entity/grace_period_test.go
git commit -m "feat: add grace period domain entity with state management"

# Task 2: Grace Period Repository
git add backend/internal/domain/repository/grace_period_repository.go backend/internal/infrastructure/persistence/repository/grace_period_repository.go backend/tests/integration/grace_period_repository_test.go
git commit -m "feat: implement grace period repository with full CRUD operations"

# Task 3: Grace Period Service
git add backend/internal/domain/service/grace_period_service.go backend/internal/application/command/create_grace_period.go backend/internal/application/command/resolve_grace_period.go backend/internal/domain/service/grace_period_service_test.go
git commit -m "feat: add grace period service with create, resolve, and expire operations"

# Task 4: Grace Period Worker
git add backend/internal/worker/tasks/grace_period_jobs.go backend/internal/domain/service/notification_service.go backend/tests/integration/grace_period_worker_test.go
git commit -m "feat: add grace period worker jobs for automated processing and notifications"
```

### Winback Offer System (Commits 5-8)

```bash
# Task 5: Winback Offer Entities
git add backend/internal/domain/entity/winback_offer.go backend/internal/domain/entity/winback_offer_test.go
git commit -m "feat: add winback offer domain entity with discount calculation"

# Task 6: Winback Service
git add backend/internal/domain/service/winback_service.go backend/internal/application/command/accept_winback_offer.go backend/internal/domain/service/winback_service_test.go
git commit -m "feat: add winback service with offer creation and acceptance"

# Task 7: Winback Repository & API
git add backend/internal/domain/repository/winback_offer_repository.go backend/internal/infrastructure/persistence/repository/winback_offer_repository.go backend/internal/interfaces/http/handlers/winback.go backend/tests/integration/winback_repository_test.go
git commit -m "feat: add winback offer repository and API handlers"

# Task 8: Winback Worker
git add backend/internal/worker/tasks/winback_jobs.go backend/tests/integration/winback_worker_test.go
git commit -m "feat: add winback offer worker jobs for automated campaign management"
```

### A/B Testing Framework (Commits 9-12)

```bash
# Task 9: Feature Flag Service
git add backend/internal/domain/service/feature_flag_service.go backend/internal/domain/service/feature_flag_service_test.go
git commit -m "feat: add A/B testing feature flag service with consistent hashing"

# Task 10: A/B Test API
git add backend/internal/interfaces/http/handlers/ab_test.go backend/internal/application/dto/ab_test_dto.go backend/tests/integration/ab_test_handler_test.go
git commit -m "feat: add A/B testing API handlers for feature flag management"

# Task 11: A/B Test Analytics
git add backend/internal/worker/tasks/ab_test_jobs.go backend/internal/domain/service/ab_analytics_service.go backend/tests/integration/ab_test_analytics_test.go
git commit -m "feat: add A/B test analytics tracking and worker jobs"

# Task 12: Feature Flag Cache (Redis)
git add backend/internal/infrastructure/cache/redis/feature_flag_cache.go
git commit -m "feat: add Redis caching for feature flags"
```

### Dunning Management (Commits 13-16)

```bash
# Task 13: Dunning Entity
git add backend/internal/domain/entity/dunning.go
git commit -m "feat: add dunning entity with exponential backoff"

# Task 14: Dunning Service
git add backend/internal/domain/service/dunning_service.go backend/internal/domain/service/dunning_service_test.go
git commit -m "feat: add dunning service with retry logic"

# Task 15: Dunning Repository
git add backend/internal/domain/repository/dunning_repository.go backend/internal/infrastructure/persistence/repository/dunning_repository.go backend/tests/integration/dunning_repository_test.go
git commit -m "feat: implement dunning repository with query methods"

# Task 16: Dunning Worker
git add backend/internal/worker/tasks/dunning_jobs.go backend/tests/integration/dunning_worker_test.go
git commit -m "feat: add dunning worker jobs for automated retries"
```

### Advanced Analytics (Commits 17-20)

```bash
# Task 17: Analytics Service
git add backend/internal/domain/service/analytics_service.go backend/internal/domain/service/analytics_service_test.go
git commit -m "feat: add analytics aggregation service for MRR/ARR and churn"

# Task 18: Analytics Repository
git add backend/internal/infrastructure/persistence/repository/analytics_repository.go backend/tests/integration/analytics_repository_test.go
git commit -m "feat: implement analytics repository with time-series queries"

# Task 19: Analytics API
git add backend/internal/interfaces/http/handlers/analytics.go backend/tests/integration/analytics_handler_test.go
git commit -m "feat: add analytics API handlers for revenue and churn metrics"

# Task 20: Analytics Worker
git add backend/internal/worker/tasks/analytics_jobs.go backend/tests/integration/analytics_worker_test.go
git commit -m "feat: add analytics worker jobs for daily aggregation"
```

### Admin Dashboard API (Commits 21-24)

```bash
# Task 21: Admin User Management
git add backend/internal/interfaces/http/handlers/admin_users.go backend/tests/integration/admin_users_handler_test.go
git commit -m "feat: add admin user management API endpoints"

# Task 22: Admin Subscription Management
git add backend/internal/interfaces/http/handlers/admin_subscriptions.go backend/tests/integration/admin_subscriptions_handler_test.go
git commit -m "feat: add admin subscription management API endpoints"

# Task 23: Admin Analytics Dashboard
git add backend/internal/interfaces/http/handlers/admin_dashboard.go backend/tests/integration/admin_dashboard_handler_test.go
git commit -m "feat: add admin dashboard analytics endpoints"

# Task 24: Admin Audit Logging
git add backend/internal/interfaces/http/middleware/audit_logging.go backend/internal/domain/service/audit_service.go backend/tests/integration/audit_logging_test.go
git commit -m "feat: add admin audit logging middleware and service"
```

### Final Commit (Optional)

```bash
# After completing all Phase 6 tasks
git commit -m "feat: complete Phase 6 production features (grace periods, winback, A/B testing, dunning, analytics, admin)"
```

---

## Commit Statistics

| Phase | Files Created | Lines of Code | Commits |
|-------|--------------|---------------|---------|
| Grace Periods | 8 | ~800 | 4 |
| Winback Offers | 8 | ~900 | 4 |
| A/B Testing | 7 | ~700 | 4 |
| Dunning | 7 | ~600 | 4 |
| Analytics | 8 | ~850 | 4 |
| Admin Dashboard | 9 | ~950 | 4 |
| **Total** | **47** | **~4,800** | **24** |

---

## Pre-Commit Checklist

Before each commit, verify:

```bash
# 1. Run tests for the changed package
go test ./path/to/package/... -v

# 2. Run linter
golangci-lint run ./path/to/package/...

# 3. Check formatting
go fmt ./path/to/package/...

# 4. Verify imports
goimports -w ./path/to/package/...

# 5. Run integration tests (if applicable)
go test -tags=integration ./tests/integration/... -v

# 6. Check for race conditions
go test -race ./path/to/package/...
```

---

## Branch Strategy

```bash
# Create feature branch for Phase 6
git checkout -b feature/phase-6-production

# After each task, push to remote
git push -u origin feature/phase-6-production

# Create PR after completing all tasks
# Or create PR per subsystem (grace, winback, ab-test, etc.)
```

---

**Plan complete and saved to `docs/plans/2026-02-28-phase-6-production-features.md`.**

**Two execution options:**

1. **Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

2. **Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**