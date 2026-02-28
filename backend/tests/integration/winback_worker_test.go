//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/worker/tasks"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestWinbackWorkerJobs(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	q := generated.New(dbContainer.Pool)

	// Setup repositories and services
	winbackRepo := repository.NewWinbackOfferRepository(dbContainer.Pool)
	userRepo := repository.NewUserRepository(q)
	subRepo := repository.NewSubscriptionRepository(q)
	winbackService := service.NewWinbackService(winbackRepo, userRepo, subRepo)
	notificationService := service.NewNotificationService()
	jobHandler := tasks.NewWinbackJobHandler(winbackService, notificationService)

	// Create test user
	user := entity.NewUser("test-platform", "test-device", entity.PlatformiOS, "1.0", "test@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("ProcessExpiredWinbackOffers expires offers", func(t *testing.T) {
		// Create expired offer
		expiredOffer := entity.NewWinbackOffer(user.ID, "expired_campaign", entity.DiscountTypePercentage, 50.0, time.Now().Add(-24*time.Hour))
		err := winbackRepo.Create(ctx, expiredOffer)
		require.NoError(t, err)

		// Create task payload
		payload, _ := json.Marshal(tasks.ProcessExpiredWinbackOffersPayload{Limit: 10})
		task := asynq.NewTask(tasks.TypeProcessExpiredWinbackOffers, payload)

		// Execute handler
		err = jobHandler.HandleProcessExpiredWinbackOffers(ctx, task)
		require.NoError(t, err)

		// Verify offer is expired
		updatedOffer, _ := winbackRepo.GetByID(ctx, expiredOffer.ID)
		assert.Equal(t, entity.OfferStatusExpired, updatedOffer.Status)
	})

	t.Run("CreateWinbackCampaign creates offers for churned users", func(t *testing.T) {
		// Create a cancelled subscription for the user
		sub := &entity.Subscription{
			UserID:    user.ID,
			Status:    entity.StatusCancelled,
			Source:    entity.SourceIAP,
			Platform:  "ios",
			ProductID: "standard_monthly",
			PlanType:  entity.PlanMonthly,
			ExpiresAt: time.Now().Add(-24 * time.Hour),
			AutoRenew: false,
		}
		err = subRepo.Create(ctx, sub)
		require.NoError(t, err)

		// Create task payload
		payload, _ := json.Marshal(tasks.CreateWinbackCampaignPayload{
			CampaignID:     "test_campaign",
			DiscountType:   string(entity.DiscountTypePercentage),
			DiscountValue:  25.0,
			DurationDays:   30,
			DaysSinceChurn: 30,
		})
		task := asynq.NewTask(tasks.TypeCreateWinbackCampaign, payload)

		// Execute handler
		err = jobHandler.HandleCreateWinbackCampaign(ctx, task)
		require.NoError(t, err)

		// Verify offer was created
		offers, err := winbackRepo.GetActiveByUserID(ctx, user.ID)
		require.NoError(t, err)

		found := false
		for _, o := range offers {
			if o.CampaignID == "test_campaign" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}
