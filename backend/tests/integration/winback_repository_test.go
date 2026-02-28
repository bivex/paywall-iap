//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestWinbackOfferRepository(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	// Create repository
	winbackRepo := repository.NewWinbackOfferRepository(dbContainer.Pool)

	// Create test user
	userRepo := repository.NewUserRepository(generated.New(dbContainer.Pool))
	user := entity.NewUser("test-platform", "test-device", entity.PlatformiOS, "1.0", "test@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("Create and GetByID", func(t *testing.T) {
		offer := entity.NewWinbackOffer(user.ID, "campaign_123", entity.DiscountTypePercentage, 25.0, time.Now().Add(30*24*time.Hour))

		err := winbackRepo.Create(ctx, offer)
		require.NoError(t, err)

		retrieved, err := winbackRepo.GetByID(ctx, offer.ID)
		require.NoError(t, err)
		assert.Equal(t, offer.ID, retrieved.ID)
		assert.Equal(t, entity.OfferStatusOffered, retrieved.Status)
	})

	t.Run("GetActiveByUserID returns active offers", func(t *testing.T) {
		offer := entity.NewWinbackOffer(user.ID, "campaign_456", entity.DiscountTypeFixed, 20.0, time.Now().Add(30*24*time.Hour))
		err := winbackRepo.Create(ctx, offer)
		require.NoError(t, err)

		offers, err := winbackRepo.GetActiveByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, offers)
		// We already created one in previous subtest
		assert.GreaterOrEqual(t, len(offers), 2)
	})

	t.Run("GetActiveByUserAndCampaign returns specific offer", func(t *testing.T) {
		offer, err := winbackRepo.GetActiveByUserAndCampaign(ctx, user.ID, "campaign_123")
		require.NoError(t, err)
		assert.Equal(t, "campaign_123", offer.CampaignID)
	})

	t.Run("Update changes offer status", func(t *testing.T) {
		offer := entity.NewWinbackOffer(user.ID, "campaign_789", entity.DiscountTypePercentage, 30.0, time.Now().Add(30*24*time.Hour))
		err := winbackRepo.Create(ctx, offer)
		require.NoError(t, err)

		err = offer.Accept()
		require.NoError(t, err)

		err = winbackRepo.Update(ctx, offer)
		require.NoError(t, err)

		retrieved, _ := winbackRepo.GetByID(ctx, offer.ID)
		assert.Equal(t, entity.OfferStatusAccepted, retrieved.Status)
		assert.NotNil(t, retrieved.AcceptedAt)
	})

	t.Run("GetExpiredOffers returns expired offers", func(t *testing.T) {
		expiredOffer := entity.NewWinbackOffer(user.ID, "campaign_expired", entity.DiscountTypePercentage, 50.0, time.Now().Add(-24*time.Hour))
		err := winbackRepo.Create(ctx, expiredOffer)
		require.NoError(t, err)

		expired, err := winbackRepo.GetExpiredOffers(ctx, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, expired)

		found := false
		for _, o := range expired {
			if o.ID == expiredOffer.ID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}
