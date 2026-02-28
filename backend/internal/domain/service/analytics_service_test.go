package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/tests/mocks"
)

func TestAnalyticsService(t *testing.T) {
	ctx := context.Background()

	repo := new(mocks.AnalyticsRepositoryMock)
	subRepo := new(mocks.MockSubscriptionRepository)
	service := NewAnalyticsService(repo, subRepo)

	t.Run("CalculateRevenueMetrics success", func(t *testing.T) {
		start := time.Now().AddDate(0, 0, -30)
		end := time.Now()

		repo.On("GetRevenueBetween", mock.Anything, mock.Anything, mock.Anything).Return(1500.0, nil).Once()
		repo.On("GetMRR", mock.Anything).Return(200.0, nil).Once()

		metrics, err := service.CalculateRevenueMetrics(ctx, start, end)
		assert.NoError(t, err)
		assert.Equal(t, 1500.0, metrics.DailyRevenue)
		assert.Equal(t, 200.0, metrics.MRR)
		assert.Equal(t, 2400.0, metrics.ARR)
		repo.AssertExpectations(t)
	})

	t.Run("CalculateChurnMetrics success", func(t *testing.T) {
		start := time.Now().AddDate(0, 0, -30)
		end := time.Now()

		repo.On("GetActiveSubscriptionCountAt", mock.Anything, mock.Anything).Return(100, nil).Once()
		repo.On("GetChurnedCountBetween", mock.Anything, mock.Anything, mock.Anything).Return(5, nil).Once()

		metrics, err := service.CalculateChurnMetrics(ctx, start, end)
		assert.NoError(t, err)
		assert.Equal(t, 100, metrics.TotalSubscriptions)
		assert.Equal(t, 5, metrics.ChurnedCount)
		assert.Equal(t, 5.0, metrics.ChurnRate)
		repo.AssertExpectations(t)
	})
}
