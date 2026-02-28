//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

func TestABAnalyticsService(t *testing.T) {
	ctx := context.Background()
	analyticsService := service.NewABAnalyticsService()

	userID := uuid.New()
	flagID := "paywall_test"

	t.Run("TrackExposure records exposure event", func(t *testing.T) {
		err := analyticsService.TrackExposure(ctx, userID, flagID, "variant_b")
		require.NoError(t, err)
	})

	t.Run("TrackConversion records conversion event", func(t *testing.T) {
		err := analyticsService.TrackConversion(ctx, userID, flagID, "variant_b")
		require.NoError(t, err)
	})

	t.Run("TrackRevenue records revenue event", func(t *testing.T) {
		err := analyticsService.TrackRevenue(ctx, userID, flagID, "variant_b", 99.99)
		require.NoError(t, err)
	})

	t.Run("GetConversionRate returns correct rate", func(t *testing.T) {
		// New test case for conversion rate
		testFlag := "conversion_test"

		// Create 10 exposed users
		for i := 0; i < 10; i++ {
			u := uuid.New()
			analyticsService.TrackExposure(ctx, u, testFlag, "control")
			if i < 5 { // 50% conversion rate
				analyticsService.TrackConversion(ctx, u, testFlag, "control")
			}
		}

		rate := analyticsService.GetConversionRate(testFlag, "control")
		assert.InDelta(t, 50.0, rate, 0.01)
	})

	t.Run("GetAverageRevenue returns correct average", func(t *testing.T) {
		testFlag := "revenue_test"

		// Create 2 exposed users for variant_a
		u1 := uuid.New()
		u2 := uuid.New()
		analyticsService.TrackExposure(ctx, u1, testFlag, "variant_a")
		analyticsService.TrackExposure(ctx, u2, testFlag, "variant_a")

		// One user pays $50
		analyticsService.TrackRevenue(ctx, u1, testFlag, "variant_a", 50.0)

		avg := analyticsService.GetAverageRevenue(testFlag, "variant_a")
		assert.InDelta(t, 25.0, avg, 0.01) // 50 / 2 = 25
	})
}
