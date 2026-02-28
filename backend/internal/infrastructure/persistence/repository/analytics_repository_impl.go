package repository

import (
	"context"
	"time"

	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsRepositoryImpl implements AnalyticsRepository using pgxpool
type AnalyticsRepositoryImpl struct {
	pool *pgxpool.Pool
}

// NewAnalyticsRepository creates a new analytics repository implementation
func NewAnalyticsRepository(pool *pgxpool.Pool) domainRepo.AnalyticsRepository {
	return &AnalyticsRepositoryImpl{
		pool: pool,
	}
}

func (r *AnalyticsRepositoryImpl) GetRevenueBetween(ctx context.Context, start, end time.Time) (float64, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE status = 'success' AND created_at >= $1 AND created_at < $2`
	var amount float64
	err := r.pool.QueryRow(ctx, query, start, end).Scan(&amount)
	return amount, err
}

func (r *AnalyticsRepositoryImpl) GetMRR(ctx context.Context) (float64, error) {
	query := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN plan_type = 'monthly' THEN amount
				WHEN plan_type = 'annual' THEN amount / 12.0
				ELSE 0 
			END
		), 0)
		FROM (
			SELECT DISTINCT ON (s.id) s.plan_type, t.amount
			FROM subscriptions s
			JOIN transactions t ON s.id = t.subscription_id
			WHERE s.status = 'active' AND t.status = 'success'
			ORDER BY s.id, t.created_at DESC
		) as active_subs
	`
	var mrr float64
	err := r.pool.QueryRow(ctx, query).Scan(&mrr)
	return mrr, err
}

func (r *AnalyticsRepositoryImpl) GetActiveSubscriptionCountAt(ctx context.Context, timestamp time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM subscriptions WHERE created_at < $1 AND (expires_at >= $1 OR status != 'expired')`
	var count int
	err := r.pool.QueryRow(ctx, query, timestamp).Scan(&count)
	return count, err
}

func (r *AnalyticsRepositoryImpl) GetChurnedCountBetween(ctx context.Context, start, end time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM subscriptions WHERE status = 'expired' AND updated_at >= $1 AND updated_at < $2`
	var count int
	err := r.pool.QueryRow(ctx, query, start, end).Scan(&count)
	return count, err
}
