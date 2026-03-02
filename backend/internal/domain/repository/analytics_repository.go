package repository

import (
	"context"
	"time"
)

// MonthlyMRR holds MRR for a single calendar month.
type MonthlyMRR struct {
	Month string  // "2025-09"
	MRR   float64
}

// SubscriptionStatusCounts holds per-status subscription counts.
type SubscriptionStatusCounts struct {
	Active    int
	Grace     int
	Cancelled int
	Expired   int
}

// WebhookProviderHealth holds unprocessed event counts per provider.
type WebhookProviderHealth struct {
	Provider    string
	Unprocessed int
	Total       int
}

// AuditLogEntry is a single recent admin action.
type AuditLogEntry struct {
	Time   time.Time
	Action string
	Detail string // JSON details marshalled to string
}

// AnalyticsRepository defines methods for retrieving analytics data
type AnalyticsRepository interface {
	GetRevenueBetween(ctx context.Context, start, end time.Time) (float64, error)
	GetMRR(ctx context.Context) (float64, error)
	GetActiveSubscriptionCountAt(ctx context.Context, timestamp time.Time) (int, error)
	GetChurnedCountBetween(ctx context.Context, start, end time.Time) (int, error)

	// Dashboard extras
	GetMRRTrend(ctx context.Context, months int) ([]MonthlyMRR, error)
	GetSubscriptionStatusCounts(ctx context.Context) (*SubscriptionStatusCounts, error)
	GetChurnRiskCount(ctx context.Context) (int, error)
	GetWebhookHealthByProvider(ctx context.Context) ([]WebhookProviderHealth, error)
	GetRecentAuditLog(ctx context.Context, limit int) ([]AuditLogEntry, error)
}
