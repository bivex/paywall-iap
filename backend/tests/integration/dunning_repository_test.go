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
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestDunningRepository(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	// Setup related objects
	user := entity.NewUser("test-platform", "test-device", entity.PlatformiOS, "1.0", "test@example.com")

	// Create user
	_, err = dbContainer.Pool.Exec(ctx, `
		INSERT INTO users (id, platform_user_id, platform, app_version, email) 
		VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.PlatformUserID, "ios", user.AppVersion, user.Email)
	require.NoError(t, err)

	subscription := entity.NewSubscription(user.ID, "iap", "ios", "premium", "monthly", time.Now().Add(30*24*time.Hour))
	_, err = dbContainer.Pool.Exec(ctx, `
		INSERT INTO subscriptions (id, user_id, status, source, platform, product_id, plan_type, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		subscription.ID, subscription.UserID, subscription.Status, subscription.Source, subscription.Platform, subscription.ProductID, subscription.PlanType, subscription.ExpiresAt)
	require.NoError(t, err)

	// Setup repository
	repo := repository.NewDunningRepository(dbContainer.Pool)

	t.Run("Create and GetByID", func(t *testing.T) {
		dunning := entity.NewDunning(subscription.ID, user.ID, time.Now().Add(24*time.Hour))
		err := repo.Create(ctx, dunning)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, dunning.ID)
		require.NoError(t, err)
		assert.Equal(t, dunning.ID, found.ID)
		assert.Equal(t, dunning.Status, found.Status)
	})

	t.Run("GetActiveBySubscriptionID", func(t *testing.T) {
		err := testutil.TruncateAll(ctx, dbContainer.Pool)
		require.NoError(t, err)

		// Re-setup user and sub after truncation
		_, err = dbContainer.Pool.Exec(ctx, `
			INSERT INTO users (id, platform_user_id, platform, app_version, email) 
			VALUES ($1, $2, $3, $4, $5)`,
			user.ID, user.PlatformUserID, "ios", user.AppVersion, user.Email)
		require.NoError(t, err)

		_, err = dbContainer.Pool.Exec(ctx, `
			INSERT INTO subscriptions (id, user_id, status, source, platform, product_id, plan_type, expires_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			subscription.ID, subscription.UserID, subscription.Status, subscription.Source, subscription.Platform, subscription.ProductID, subscription.PlanType, subscription.ExpiresAt)
		require.NoError(t, err)

		dunning := entity.NewDunning(subscription.ID, user.ID, time.Now().Add(24*time.Hour))
		err = repo.Create(ctx, dunning)
		require.NoError(t, err)

		found, err := repo.GetActiveBySubscriptionID(ctx, subscription.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, dunning.ID, found.ID)
	})

	t.Run("Update", func(t *testing.T) {
		dunning := entity.NewDunning(subscription.ID, user.ID, time.Now().Add(24*time.Hour))
		repo.Create(ctx, dunning)

		dunning.MarkRecovered()
		err := repo.Update(ctx, dunning)
		require.NoError(t, err)

		found, _ := repo.GetByID(ctx, dunning.ID)
		assert.Equal(t, entity.DunningStatusRecovered, found.Status)
		assert.NotNil(t, found.RecoveredAt)
	})

	t.Run("GetPendingAttempts", func(t *testing.T) {
		// Dunning ready for attempt (NextAttemptAt in the past)
		dunning := entity.NewDunning(subscription.ID, user.ID, time.Now().Add(-1*time.Hour))
		repo.Create(ctx, dunning)

		// Dunning NOT ready (NextAttemptAt in the future)
		dunningFuture := entity.NewDunning(subscription.ID, user.ID, time.Now().Add(24*time.Hour))
		repo.Create(ctx, dunningFuture)

		pending, err := repo.GetPendingAttempts(ctx, 10)
		require.NoError(t, err)

		found := false
		for _, d := range pending {
			if d.ID == dunning.ID {
				found = true
			}
			if d.ID == dunningFuture.ID {
				t.Errorf("Should not have found dunning with future attempt at")
			}
		}
		assert.True(t, found)
	})
}
