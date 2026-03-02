package service

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsReportService handles analytics report generation
type AnalyticsReportService struct {
	dbPool *pgxpool.Pool
}

// NewAnalyticsReportService creates a new analytics report service
func NewAnalyticsReportService(dbPool *pgxpool.Pool) *AnalyticsReportService {
	return &AnalyticsReportService{dbPool: dbPool}
}

// TrendPoint represents a monthly trend data point
type TrendPoint struct {
	Month       string  `json:"month"`
	MRR         float64 `json:"mrr"`
	ActiveCount int     `json:"active_count"`
	NewSubs     int     `json:"new_subs"`
}

// PlatformRow represents platform statistics
type PlatformRow struct {
	Platform string  `json:"platform"`
	Count    int     `json:"count"`
	MRR      float64 `json:"mrr"`
}

// PlanRow represents plan type statistics
type PlanRow struct {
	PlanType string  `json:"plan_type"`
	Count    int     `json:"count"`
	MRR      float64 `json:"mrr"`
}

// StatusCounts represents subscription status breakdown
type StatusCounts struct {
	Active    int `json:"active"`
	Grace     int `json:"grace"`
	Cancelled int `json:"cancelled"`
	Expired   int `json:"expired"`
}

// Report contains the complete analytics report
type Report struct {
	MRR           float64        `json:"mrr"`
	ARR           float64        `json:"arr"`
	LTV           float64        `json:"ltv"`
	TotalRevenue  float64        `json:"total_revenue"`
	ChurnRate     float64        `json:"churn_rate"`
	NewSubsMonth  int            `json:"new_subs_month"`
	Trend         []TrendPoint   `json:"trend"`
	ByPlatform    []PlatformRow  `json:"by_platform"`
	ByPlan        []PlanRow      `json:"by_plan"`
	StatusCounts  StatusCounts   `json:"status_counts"`
}

// GetReport fetches the complete analytics report
func (s *AnalyticsReportService) GetReport(ctx context.Context) (*Report, error) {
	mrr, err := s.fetchMRR(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch mrr: %w", err)
	}

	ltv, totalRevenue, err := s.fetchLTV(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch ltv: %w", err)
	}

	newSubsMonth, err := s.fetchNewSubsMonth(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch new subs: %w", err)
	}

	churnRate, err := s.fetchChurnRate(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch churn rate: %w", err)
	}

	trend, err := s.fetchTrend(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch trend: %w", err)
	}

	byPlatform, err := s.fetchByPlatform(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch by platform: %w", err)
	}

	byPlan, err := s.fetchByPlan(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch by plan: %w", err)
	}

	statusCounts, err := s.fetchStatusCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch status counts: %w", err)
	}

	return &Report{
		MRR:          mrr,
		ARR:          math.Round(mrr * 12 * 100 / 100),
		LTV:          ltv,
		TotalRevenue: totalRevenue,
		ChurnRate:    churnRate,
		NewSubsMonth: newSubsMonth,
		Trend:        trend,
		ByPlatform:   byPlatform,
		ByPlan:       byPlan,
		StatusCounts: statusCounts,
	}, nil
}

// fetchMRR retrieves current monthly recurring revenue
func (s *AnalyticsReportService) fetchMRR(ctx context.Context) (float64, error) {
	var mrr float64
	err := s.dbPool.QueryRow(ctx, `
		SELECT COALESCE(ROUND(SUM(
			CASE WHEN plan_type='monthly' THEN 9.99
			     WHEN plan_type='annual'  THEN 99.99/12.0
			     ELSE 0 END
		)::numeric, 2), 0)
		FROM subscriptions
		WHERE status IN ('active','grace') AND deleted_at IS NULL`).Scan(&mrr)
	return mrr, err
}

// fetchLTV retrieves lifetime value and total revenue
func (s *AnalyticsReportService) fetchLTV(ctx context.Context) (ltv, totalRevenue float64, err error) {
	err = s.dbPool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0),
		       COALESCE(ROUND(SUM(amount)/NULLIF(COUNT(DISTINCT user_id),0),2),0)
		FROM transactions WHERE status='success'`).Scan(&totalRevenue, &ltv)
	return
}

// fetchNewSubsMonth retrieves count of new subscriptions this month
func (s *AnalyticsReportService) fetchNewSubsMonth(ctx context.Context) (int, error) {
	var count int
	err := s.dbPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM subscriptions
		WHERE deleted_at IS NULL
		  AND date_trunc('month', created_at) = date_trunc('month', now())`).Scan(&count)
	return count, err
}

// fetchChurnRate calculates churn rate for current month
func (s *AnalyticsReportService) fetchChurnRate(ctx context.Context) (float64, error) {
	var churned, activePlusChurned int
	err := s.dbPool.QueryRow(ctx, `
		SELECT
		  COUNT(*) FILTER (WHERE status IN ('cancelled','expired')
		    AND date_trunc('month', updated_at) = date_trunc('month', now())),
		  COUNT(*) FILTER (WHERE status IN ('active','grace','cancelled','expired'))
		FROM subscriptions WHERE deleted_at IS NULL`).Scan(&churned, &activePlusChurned)
	if err != nil {
		return 0, err
	}

	churnRate := 0.0
	if activePlusChurned > 0 {
		churnRate = math.Round(float64(churned)/float64(activePlusChurned)*1000) / 10
	}
	return churnRate, nil
}

// fetchTrend retrieves MRR trend for last 6 months
func (s *AnalyticsReportService) fetchTrend(ctx context.Context) ([]TrendPoint, error) {
	rows, err := s.dbPool.Query(ctx, `
		WITH months AS (
			SELECT generate_series(
				date_trunc('month', now()) - 5 * interval '1 month',
				date_trunc('month', now()),
				interval '1 month'
			) AS month_start
		),
		monthly_subs AS (
			SELECT m.month_start,
				COUNT(s.id) AS active_count,
				COALESCE(ROUND(SUM(CASE
					WHEN s.plan_type='monthly' THEN 9.99
					WHEN s.plan_type='annual'  THEN 99.99/12.0
					ELSE 0 END)::numeric,2),0) AS mrr
			FROM months m
			LEFT JOIN subscriptions s ON s.deleted_at IS NULL
				AND s.created_at < m.month_start + interval '1 month'
				AND (s.expires_at >= m.month_start OR s.status IN ('active','grace'))
			GROUP BY m.month_start
		),
		monthly_new AS (
			SELECT date_trunc('month', created_at) AS ms, COUNT(*) AS new_subs
			FROM subscriptions WHERE deleted_at IS NULL GROUP BY 1
		)
		SELECT to_char(ms.month_start,'YYYY-MM'), ms.mrr, ms.active_count, COALESCE(mn.new_subs,0)
		FROM monthly_subs ms
		LEFT JOIN monthly_new mn ON mn.ms = ms.month_start
		ORDER BY ms.month_start`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trend []TrendPoint
	for rows.Next() {
		var p TrendPoint
		if err := rows.Scan(&p.Month, &p.MRR, &p.ActiveCount, &p.NewSubs); err != nil {
			return nil, err
		}
		trend = append(trend, p)
	}

	return trend, nil
}

// fetchByPlatform retrieves platform breakdown
func (s *AnalyticsReportService) fetchByPlatform(ctx context.Context) ([]PlatformRow, error) {
	rows, err := s.dbPool.Query(ctx, `
		SELECT platform, COUNT(*),
			ROUND(SUM(CASE WHEN plan_type='monthly' THEN 9.99
			               WHEN plan_type='annual'  THEN 99.99/12.0 ELSE 0 END)::numeric,2)
		FROM subscriptions
		WHERE status IN ('active','grace') AND deleted_at IS NULL
		GROUP BY platform ORDER BY platform`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []PlatformRow
	for rows.Next() {
		var p PlatformRow
		if err := rows.Scan(&p.Platform, &p.Count, &p.MRR); err != nil {
			return nil, err
		}
		stats = append(stats, p)
	}

	return stats, nil
}

// fetchByPlan retrieves plan type breakdown
func (s *AnalyticsReportService) fetchByPlan(ctx context.Context) ([]PlanRow, error) {
	rows, err := s.dbPool.Query(ctx, `
		SELECT plan_type, COUNT(*),
			ROUND(SUM(CASE WHEN plan_type='monthly' THEN 9.99
			               WHEN plan_type='annual'  THEN 99.99/12.0 ELSE 0 END)::numeric,2)
		FROM subscriptions
		WHERE status IN ('active','grace') AND deleted_at IS NULL
		GROUP BY plan_type ORDER BY plan_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []PlanRow
	for rows.Next() {
		var p PlanRow
		if err := rows.Scan(&p.PlanType, &p.Count, &p.MRR); err != nil {
			return nil, err
		}
		stats = append(stats, p)
	}

	return stats, nil
}

// fetchStatusCounts retrieves subscription status counts
func (s *AnalyticsReportService) fetchStatusCounts(ctx context.Context) (StatusCounts, error) {
	var counts StatusCounts
	err := s.dbPool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status='active'),
			COUNT(*) FILTER (WHERE status='grace'),
			COUNT(*) FILTER (WHERE status='cancelled'),
			COUNT(*) FILTER (WHERE status='expired')
		FROM subscriptions WHERE deleted_at IS NULL`).Scan(
		&counts.Active, &counts.Grace, &counts.Cancelled, &counts.Expired)
	return counts, err
}
