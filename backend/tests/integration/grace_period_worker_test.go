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

func TestGracePeriodWorkerJobs(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	// Setup queries
	queries := generated.New(dbContainer.Pool)

	// Setup repositories
	userRepo := repository.NewUserRepository(queries)
	subRepo := repository.NewSubscriptionRepository(queries)
	gpRepo := repository.NewGracePeriodRepository(dbContainer.Pool)

	// Setup services
	graceService := service.NewGracePeriodService(gpRepo, subRepo, userRepo)
	notificationService := service.NewNotificationService()
	jobHandler := tasks.NewGracePeriodJobHandler(graceService, notificationService)

	// Create test user and subscription
	user := entity.NewUser("test-platform", "test-device", entity.PlatformiOS, "1.0", "test@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	sub := entity.NewSubscription(user.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	t.Run("ProcessExpiredGracePeriods expires and cancels subscription", func(t *testing.T) {
		// Create expired grace period
		expiredGP := entity.NewGracePeriod(user.ID, sub.ID, time.Now().Add(-24*time.Hour))
		err := gpRepo.Create(ctx, expiredGP)
		require.NoError(t, err)

		// Create task payload
		payload, _ := json.Marshal(tasks.ProcessExpiredGracePeriodsPayload{Limit: 10})
		task := asynq.NewTask(tasks.TypeProcessExpiredGracePeriods, payload)

		// Execute handler
		err = jobHandler.HandleProcessExpiredGracePeriods(ctx, task)
		require.NoError(t, err)

		// Verify grace period is expired
		updatedGP, _ := gpRepo.GetByID(ctx, expiredGP.ID)
		assert.Equal(t, entity.GraceStatusExpired, updatedGP.Status)

		// Verify subscription is cancelled
		updatedSub, _ := subRepo.GetByID(ctx, sub.ID)
		assert.Equal(t, entity.StatusCancelled, updatedSub.Status)
	})

	t.Run("NotifyExpiringGracePeriods sends notifications", func(t *testing.T) {
		// Create grace period expiring in 2 hours
		user2 := entity.NewUser("test-platform-2", "test-device-2", entity.PlatformiOS, "1.0", "test2@example.com")
		err = userRepo.Create(ctx, user2)
		require.NoError(t, err)

		sub2 := entity.NewSubscription(user2.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
		err = subRepo.Create(ctx, sub2)
		require.NoError(t, err)

		expiringGP := entity.NewGracePeriod(user2.ID, sub2.ID, time.Now().Add(2*time.Hour))
		err = gpRepo.Create(ctx, expiringGP)
		require.NoError(t, err)

		// Create task payload
		payload, _ := json.Marshal(tasks.NotifyExpiringGracePeriodsPayload{WithinHours: 24})
		task := asynq.NewTask(tasks.TypeNotifyExpiringGracePeriods, payload)

		// Execute handler (should send notification)
		err = jobHandler.HandleNotifyExpiringGracePeriods(ctx, task)
		require.NoError(t, err)
	})
}
