package repository

import (
	"context"
	"encoding/json"
	"fmt"
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

// GetMRRTrend returns monthly MRR for the last N months (oldest first).
func (r *AnalyticsRepositoryImpl) GetMRRTrend(ctx context.Context, months int) ([]domainRepo.MonthlyMRR, error) {
	query := `
		WITH months AS (
			SELECT generate_series(
				date_trunc('month', now()) - ($1 - 1) * interval '1 month',
				date_trunc('month', now()),
				interval '1 month'
			) AS month_start
		)
		SELECT
			to_char(m.month_start, 'YYYY-MM') AS month,
			COALESCE(SUM(
				CASE
					WHEN s.plan_type = 'monthly' THEN t.amount
					WHEN s.plan_type = 'annual'  THEN t.amount / 12.0
					ELSE 0
				END
			), 0) AS mrr
		FROM months m
		LEFT JOIN subscriptions s
			ON s.status = 'active'
			AND date_trunc('month', s.created_at) <= m.month_start
			AND (s.expires_at >= m.month_start + interval '1 month' OR s.status != 'expired')
		LEFT JOIN transactions t
			ON t.subscription_id = s.id
			AND t.status = 'success'
			AND date_trunc('month', t.created_at) = m.month_start
		GROUP BY m.month_start
		ORDER BY m.month_start ASC
	`
	rows, err := r.pool.Query(ctx, query, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domainRepo.MonthlyMRR
	for rows.Next() {
		var m domainRepo.MonthlyMRR
		if err := rows.Scan(&m.Month, &m.MRR); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

// GetSubscriptionStatusCounts returns counts broken down by status.
func (r *AnalyticsRepositoryImpl) GetSubscriptionStatusCounts(ctx context.Context) (*domainRepo.SubscriptionStatusCounts, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'active')    AS active,
			COUNT(*) FILTER (WHERE status = 'grace')     AS grace,
			COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled,
			COUNT(*) FILTER (WHERE status = 'expired')   AS expired
		FROM subscriptions
	`
	c := &domainRepo.SubscriptionStatusCounts{}
	err := r.pool.QueryRow(ctx, query).Scan(&c.Active, &c.Grace, &c.Cancelled, &c.Expired)
	return c, err
}

// GetChurnRiskCount returns the number of subscriptions in grace/dunning state.
func (r *AnalyticsRepositoryImpl) GetChurnRiskCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM subscriptions WHERE status = 'grace'`
	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetWebhookHealthByProvider returns per-provider unprocessed event counts.
func (r *AnalyticsRepositoryImpl) GetWebhookHealthByProvider(ctx context.Context) ([]domainRepo.WebhookProviderHealth, error) {
	query := `
		SELECT
			provider,
			COUNT(*) FILTER (WHERE processed_at IS NULL) AS unprocessed,
			COUNT(*)                                      AS total
		FROM webhook_events
		GROUP BY provider
		ORDER BY provider ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domainRepo.WebhookProviderHealth
	for rows.Next() {
		var h domainRepo.WebhookProviderHealth
		if err := rows.Scan(&h.Provider, &h.Unprocessed, &h.Total); err != nil {
			return nil, err
		}
		result = append(result, h)
	}
	return result, rows.Err()
}

// GetRecentAuditLog returns the most recent N admin audit log entries.
func (r *AnalyticsRepositoryImpl) GetRecentAuditLog(ctx context.Context, limit int) ([]domainRepo.AuditLogEntry, error) {
	query := `
		SELECT a.created_at, a.action, COALESCE(a.details::text, '{}')
		FROM admin_audit_log a
		ORDER BY a.created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domainRepo.AuditLogEntry
	for rows.Next() {
		var e domainRepo.AuditLogEntry
		var detailsRaw string
		if err := rows.Scan(&e.Time, &e.Action, &detailsRaw); err != nil {
			return nil, err
		}
		// Flatten JSONB details to a readable string
		var details map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(detailsRaw), &details); jsonErr == nil {
			parts := ""
			for k, v := range details {
				if parts != "" {
					parts += ", "
				}
				parts += fmt.Sprintf("%s: %v", k, v)
			}
			e.Detail = parts
		} else {
			e.Detail = detailsRaw
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

