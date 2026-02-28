package repository

import (
	"context"
	"time"
)

// AnalyticsRepository defines methods for retrieving analytics data
type AnalyticsRepository interface {
	GetRevenueBetween(ctx context.Context, start, end time.Time) (float64, error)
	GetMRR(ctx context.Context) (float64, error)
	GetActiveSubscriptionCountAt(ctx context.Context, timestamp time.Time) (int, error)
	GetChurnedCountBetween(ctx context.Context, start, end time.Time) (int, error)
}
