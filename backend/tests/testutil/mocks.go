package testutil

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

// MockQueries is a mock sqlc queries implementation
type MockQueries struct{}

// NewMockUserRepo creates a mock user repository that uses raw SQL
func NewMockUserRepo(pool *pgxpool.Pool) repository.UserRepository {
	return &mockUserRepo{pool: pool}
}

// mockUserRepo implements UserRepository with raw SQL queries
type mockUserRepo struct {
	pool *pgxpool.Pool
}

func (r *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email) VALUES ($1, $2, $3, $4, $5, $6)",
		user.ID, user.PlatformUserID, user.DeviceID, user.Platform, user.AppVersion, user.Email,
	)
	return err
}

func (r *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row := r.pool.QueryRow(ctx,
		"SELECT id, platform_user_id, device_id, platform, app_version, email, ltv, ltv_updated_at, created_at, deleted_at FROM users WHERE id = $1 AND deleted_at IS NULL",
		id,
	)

	var user entity.User
	err := row.Scan(
		&user.ID, &user.PlatformUserID, &user.DeviceID, &user.Platform,
		&user.AppVersion, &user.Email, &user.LTV, &user.LtvUpdatedAt,
		&user.CreatedAt, &user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *mockUserRepo) GetByPlatformID(ctx context.Context, platformUserID string) (*entity.User, error) {
	row := r.pool.QueryRow(ctx,
		"SELECT id, platform_user_id, device_id, platform, app_version, email, ltv, ltv_updated_at, created_at, deleted_at FROM users WHERE platform_user_id = $1 AND deleted_at IS NULL",
		platformUserID,
	)

	var user entity.User
	err := row.Scan(
		&user.ID, &user.PlatformUserID, &user.DeviceID, &user.Platform,
		&user.AppVersion, &user.Email, &user.LTV, &user.LtvUpdatedAt,
		&user.CreatedAt, &user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row := r.pool.QueryRow(ctx,
		"SELECT id, platform_user_id, device_id, platform, app_version, email, ltv, ltv_updated_at, created_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS NULL",
		email,
	)

	var user entity.User
	err := row.Scan(
		&user.ID, &user.PlatformUserID, &user.DeviceID, &user.Platform,
		&user.AppVersion, &user.Email, &user.LTV, &user.LtvUpdatedAt,
		&user.CreatedAt, &user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *mockUserRepo) Update(ctx context.Context, user *entity.User) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET ltv = $2, ltv_updated_at = now() WHERE id = $1",
		user.ID, user.LTV,
	)
	return err
}

func (r *mockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "UPDATE users SET deleted_at = now() WHERE id = $1", id)
	return err
}

func (r *mockUserRepo) ExistsByPlatformID(ctx context.Context, platformUserID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE platform_user_id = $1 AND deleted_at IS NULL)",
		platformUserID,
	).Scan(&exists)
	return exists, err
}

// NewMockSubscriptionRepo creates a mock subscription repository
func NewMockSubscriptionRepo(pool *pgxpool.Pool) repository.SubscriptionRepository {
	return &mockSubscriptionRepo{pool: pool}
}

type mockSubscriptionRepo struct {
	pool *pgxpool.Pool
}

func (r *mockSubscriptionRepo) Create(ctx context.Context, sub *entity.Subscription) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO subscriptions (id, user_id, status, source, platform, product_id, plan_type, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		sub.ID, sub.UserID, sub.Status, sub.Source, sub.Platform, sub.ProductID, sub.PlanType, sub.ExpiresAt,
	)
	return err
}

func (r *mockSubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	return nil, domainErrors.ErrSubscriptionNotFound
}

func (r *mockSubscriptionRepo) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	row := r.pool.QueryRow(ctx,
		"SELECT id, user_id, status, source, platform, product_id, plan_type, expires_at, auto_renew, created_at, updated_at, deleted_at FROM subscriptions WHERE user_id = $1 AND status = 'active' AND deleted_at IS NULL",
		userID,
	)

	var sub entity.Subscription
	err := row.Scan(
		&sub.ID, &sub.UserID, &sub.Status, &sub.Source, &sub.Platform,
		&sub.ProductID, &sub.PlanType, &sub.ExpiresAt, &sub.AutoRenew,
		&sub.CreatedAt, &sub.UpdatedAt, &sub.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.ErrSubscriptionNotActive
		}
		return nil, err
	}
	return &sub, nil
}

func (r *mockSubscriptionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error) {
	return []*entity.Subscription{}, nil
}

func (r *mockSubscriptionRepo) Update(ctx context.Context, sub *entity.Subscription) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE subscriptions SET status = $2, updated_at = now() WHERE id = $1",
		sub.ID, sub.Status,
	)
	return err
}

func (r *mockSubscriptionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SubscriptionStatus) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE subscriptions SET status = $2, updated_at = now() WHERE id = $1",
		id, status,
	)
	return err
}

func (r *mockSubscriptionRepo) UpdateExpiry(ctx context.Context, id uuid.UUID, expiresAt interface{}) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE subscriptions SET expires_at = $2, updated_at = now() WHERE id = $1",
		id, expiresAt.(time.Time),
	)
	return err
}

func (r *mockSubscriptionRepo) Cancel(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE subscriptions SET status = 'cancelled', auto_renew = false, updated_at = now() WHERE id = $1",
		id,
	)
	return err
}

func (r *mockSubscriptionRepo) CanAccess(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM subscriptions WHERE user_id = $1 AND status = 'active' AND expires_at > now() AND deleted_at IS NULL",
		userID,
	).Scan(&count)
	return count > 0, err
}
