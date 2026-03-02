package service

import (
	"context"
	"time"

	"github.com/bivex/paywall-iap/internal/domain/repository"
)

// AnalyticsService handles advanced analytics calculations
type AnalyticsService struct {
	repo             repository.AnalyticsRepository
	subscriptionRepo repository.SubscriptionRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(repo repository.AnalyticsRepository, subscriptionRepo repository.SubscriptionRepository) *AnalyticsService {
	return &AnalyticsService{
		repo:             repo,
		subscriptionRepo: subscriptionRepo,
	}
}

// ChurnMetrics represents churn-related statistics
type ChurnMetrics struct {
	TotalSubscriptions int
	ChurnedCount       int
	ChurnRate          float64
	Period             string
}

// RevenueMetrics represents revenue-related statistics
type RevenueMetrics struct {
	DailyRevenue   float64
	WeeklyRevenue  float64
	MonthlyRevenue float64
	MRR            float64
	ARR            float64
}

// CalculateRevenueMetrics calculates revenue metrics for a period
func (s *AnalyticsService) CalculateRevenueMetrics(ctx context.Context, start, end time.Time) (*RevenueMetrics, error) {
	// Daily revenue
	daily, err := s.repo.GetRevenueBetween(ctx, time.Now().Truncate(24*time.Hour), time.Now())
	if err != nil {
		return nil, err
	}

	// MRR
	mrr, err := s.repo.GetMRR(ctx)
	if err != nil {
		return nil, err
	}

	return &RevenueMetrics{
		DailyRevenue: daily,
		MRR:          mrr,
		ARR:          mrr * 12.0,
	}, nil
}

// CalculateChurnMetrics calculates churn metrics for a period
func (s *AnalyticsService) CalculateChurnMetrics(ctx context.Context, start, end time.Time) (*ChurnMetrics, error) {
	// Total active at start
	total, err := s.repo.GetActiveSubscriptionCountAt(ctx, start)
	if err != nil {
		return nil, err
	}

	// Churned during period
	churned, err := s.repo.GetChurnedCountBetween(ctx, start, end)
	if err != nil {
		return nil, err
	}

	rate := 0.0
	if total > 0 {
		rate = float64(churned) / float64(total) * 100
	}

	return &ChurnMetrics{
		TotalSubscriptions: total,
		ChurnedCount:       churned,
		ChurnRate:          rate,
		Period:             start.Format("2006-01-02") + " to " + end.Format("2006-01-02"),
	}, nil
}

// GetMRRTrend delegates to the repository.
func (s *AnalyticsService) GetMRRTrend(ctx context.Context, months int) ([]repository.MonthlyMRR, error) {
	return s.repo.GetMRRTrend(ctx, months)
}

// GetSubscriptionStatusCounts delegates to the repository.
func (s *AnalyticsService) GetSubscriptionStatusCounts(ctx context.Context) (*repository.SubscriptionStatusCounts, error) {
	return s.repo.GetSubscriptionStatusCounts(ctx)
}

// GetChurnRiskCount delegates to the repository.
func (s *AnalyticsService) GetChurnRiskCount(ctx context.Context) (int, error) {
	return s.repo.GetChurnRiskCount(ctx)
}

// GetWebhookHealthByProvider delegates to the repository.
func (s *AnalyticsService) GetWebhookHealthByProvider(ctx context.Context) ([]repository.WebhookProviderHealth, error) {
	return s.repo.GetWebhookHealthByProvider(ctx)
}

// GetRecentAuditLog delegates to the repository.
func (s *AnalyticsService) GetRecentAuditLog(ctx context.Context, limit int) ([]repository.AuditLogEntry, error) {
	return s.repo.GetRecentAuditLog(ctx, limit)
}

