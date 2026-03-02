package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

type AnalyticsRepositoryMock struct {
	mock.Mock
}

func (m *AnalyticsRepositoryMock) GetRevenueBetween(ctx context.Context, start, end time.Time) (float64, error) {
	args := m.Called(ctx, start, end)
	return args.Get(0).(float64), args.Error(1)
}

func (m *AnalyticsRepositoryMock) GetMRR(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}

func (m *AnalyticsRepositoryMock) GetActiveSubscriptionCountAt(ctx context.Context, timestamp time.Time) (int, error) {
	args := m.Called(ctx, timestamp)
	return args.Int(0), args.Error(1)
}

func (m *AnalyticsRepositoryMock) GetChurnedCountBetween(ctx context.Context, start, end time.Time) (int, error) {
	args := m.Called(ctx, start, end)
	return args.Int(0), args.Error(1)
}

func (m *AnalyticsRepositoryMock) GetSubscriptionStatusCounts(ctx context.Context) (*repository.SubscriptionStatusCounts, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SubscriptionStatusCounts), args.Error(1)
}

func (m *AnalyticsRepositoryMock) GetChurnRiskCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *AnalyticsRepositoryMock) GetWebhookHealthByProvider(ctx context.Context) ([]repository.WebhookProviderHealth, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.WebhookProviderHealth), args.Error(1)
}

func (m *AnalyticsRepositoryMock) GetRecentAuditLog(ctx context.Context, limit int) ([]repository.AuditLogEntry, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.AuditLogEntry), args.Error(1)
}

func (m *AnalyticsRepositoryMock) GetAuditLogPaginated(ctx context.Context, offset, limit int, action, search string, from, to time.Time) (*repository.AuditLogPage, error) {
	args := m.Called(ctx, offset, limit, action, search, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AuditLogPage), args.Error(1)
}

func (m *AnalyticsRepositoryMock) GetMRRTrend(ctx context.Context, months int) ([]repository.MonthlyMRR, error) {
	args := m.Called(ctx, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.MonthlyMRR), args.Error(1)
}
