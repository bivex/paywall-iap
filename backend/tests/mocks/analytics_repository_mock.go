package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
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
