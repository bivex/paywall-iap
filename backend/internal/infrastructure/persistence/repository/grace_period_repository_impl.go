package repository

import (
	"context"

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

// GetExpiringSoon retrieves grace periods expiring within the specified number of hours
func (r *GracePeriodRepositoryImpl) GetExpiringSoon(ctx context.Context, withinHours int) ([]*entity.GracePeriod, error) {
	query := `
		SELECT id, user_id, subscription_id, status, expires_at, resolved_at, created_at, updated_at
		FROM grace_periods
		WHERE status = 'active'
		  AND expires_at > NOW()
		  AND expires_at < NOW() + ($1 * INTERVAL '1 hour')
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
