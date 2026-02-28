package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// DunningRepositoryImpl implements DunningRepository using pgxpool
type DunningRepositoryImpl struct {
	pool *pgxpool.Pool
}

// NewDunningRepository creates a new dunning repository
func NewDunningRepository(pool *pgxpool.Pool) *DunningRepositoryImpl {
	return &DunningRepositoryImpl{
		pool: pool,
	}
}

// Create creates a new dunning process
func (r *DunningRepositoryImpl) Create(ctx context.Context, dunning *entity.Dunning) error {
	query := `
		INSERT INTO dunning (
			id, subscription_id, user_id, status, attempt_count, max_attempts, next_attempt_at, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		dunning.ID, dunning.SubscriptionID, dunning.UserID, dunning.Status,
		dunning.AttemptCount, dunning.MaxAttempts, dunning.NextAttemptAt,
		dunning.CreatedAt, dunning.UpdatedAt,
	)
	return err
}

// GetByID retrieves a dunning record by ID
func (r *DunningRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Dunning, error) {
	query := `
		SELECT 
			id, subscription_id, user_id, status, attempt_count, max_attempts, next_attempt_at, 
			last_attempt_at, recovered_at, failed_at, created_at, updated_at
		FROM dunning
		WHERE id = $1
	`
	d := &entity.Dunning{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.SubscriptionID, &d.UserID, &d.Status, &d.AttemptCount, &d.MaxAttempts, &d.NextAttemptAt,
		&d.LastAttemptAt, &d.RecoveredAt, &d.FailedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("dunning not found")
	}
	return d, err
}

// GetActiveBySubscriptionID retrieves the currently active dunning for a subscription
func (r *DunningRepositoryImpl) GetActiveBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entity.Dunning, error) {
	query := `
		SELECT 
			id, subscription_id, user_id, status, attempt_count, max_attempts, next_attempt_at, 
			last_attempt_at, recovered_at, failed_at, created_at, updated_at
		FROM dunning
		WHERE subscription_id = $1 AND status IN ('pending', 'in_progress')
		LIMIT 1
	`
	d := &entity.Dunning{}
	err := r.pool.QueryRow(ctx, query, subscriptionID).Scan(
		&d.ID, &d.SubscriptionID, &d.UserID, &d.Status, &d.AttemptCount, &d.MaxAttempts, &d.NextAttemptAt,
		&d.LastAttemptAt, &d.RecoveredAt, &d.FailedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Return nil, nil when no active dunning found
	}
	return d, err
}

// Update updates an existing dunning record
func (r *DunningRepositoryImpl) Update(ctx context.Context, dunning *entity.Dunning) error {
	query := `
		UPDATE dunning
		SET 
			status = $2, attempt_count = $3, next_attempt_at = $4, last_attempt_at = $5, 
			recovered_at = $6, failed_at = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		dunning.ID, dunning.Status, dunning.AttemptCount, dunning.NextAttemptAt,
		dunning.LastAttemptAt, dunning.RecoveredAt, dunning.FailedAt, time.Now(),
	)
	return err
}

// GetPendingAttempts retrieves dunning records with NextAttemptAt <= NOW()
func (r *DunningRepositoryImpl) GetPendingAttempts(ctx context.Context, limit int) ([]*entity.Dunning, error) {
	query := `
		SELECT 
			id, subscription_id, user_id, status, attempt_count, max_attempts, next_attempt_at, 
			last_attempt_at, recovered_at, failed_at, created_at, updated_at
		FROM dunning
		WHERE status IN ('pending', 'in_progress') AND next_attempt_at <= $1
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*entity.Dunning
	for rows.Next() {
		d := &entity.Dunning{}
		err := rows.Scan(
			&d.ID, &d.SubscriptionID, &d.UserID, &d.Status, &d.AttemptCount, &d.MaxAttempts, &d.NextAttemptAt,
			&d.LastAttemptAt, &d.RecoveredAt, &d.FailedAt, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, d)
	}

	return results, nil
}
