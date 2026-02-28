package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/tests/mocks"
)

func TestGracePeriodService(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateGracePeriod success", func(t *testing.T) {
		gracePeriodRepo := mocks.NewMockGracePeriodRepository()
		subscriptionRepo := mocks.NewMockSubscriptionRepository()
		userRepo := mocks.NewMockUserRepository()
		graceService := service.NewGracePeriodService(gracePeriodRepo, subscriptionRepo, userRepo)

		userID := uuid.New()
		subscriptionID := uuid.New()

		// No active grace period exists
		gracePeriodRepo.On("GetActiveBySubscriptionID", ctx, subscriptionID).Return(nil, errors.New("not found"))
		subscriptionRepo.On("GetByID", ctx, subscriptionID).Return(&entity.Subscription{
			ID:     subscriptionID,
			UserID: userID,
		}, nil)
		subscriptionRepo.On("UpdateStatus", ctx, subscriptionID, entity.StatusGrace).Return(nil)
		gracePeriodRepo.On("Create", ctx, mock.Anything).Return(nil)

		gracePeriod, err := graceService.CreateGracePeriod(ctx, userID, subscriptionID, 7)
		require.NoError(t, err)
		assert.Equal(t, entity.GraceStatusActive, gracePeriod.Status)
		assert.Equal(t, 7, gracePeriod.DaysRemaining())
	})

	t.Run("CreateGracePeriod with existing active grace period returns error", func(t *testing.T) {
		gracePeriodRepo := mocks.NewMockGracePeriodRepository()
		subscriptionRepo := mocks.NewMockSubscriptionRepository()
		userRepo := mocks.NewMockUserRepository()
		graceService := service.NewGracePeriodService(gracePeriodRepo, subscriptionRepo, userRepo)

		userID := uuid.New()
		subscriptionID := uuid.New()

		existingGP := entity.NewGracePeriod(userID, subscriptionID, time.Now().Add(24*time.Hour))
		gracePeriodRepo.On("GetActiveBySubscriptionID", ctx, subscriptionID).Return(existingGP, nil)

		gracePeriod, err := graceService.CreateGracePeriod(ctx, userID, subscriptionID, 7)
		assert.Error(t, err)
		assert.Nil(t, gracePeriod)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("ResolveGracePeriod success", func(t *testing.T) {
		gracePeriodRepo := mocks.NewMockGracePeriodRepository()
		subscriptionRepo := mocks.NewMockSubscriptionRepository()
		userRepo := mocks.NewMockUserRepository()
		graceService := service.NewGracePeriodService(gracePeriodRepo, subscriptionRepo, userRepo)

		userID := uuid.New()
		subscriptionID := uuid.New()

		gracePeriod := entity.NewGracePeriod(userID, subscriptionID, time.Now().Add(24*time.Hour))
		gracePeriodRepo.On("GetActiveBySubscriptionID", ctx, subscriptionID).Return(gracePeriod, nil)
		gracePeriodRepo.On("Update", ctx, gracePeriod).Return(nil)
		subscriptionRepo.On("UpdateStatus", ctx, subscriptionID, entity.StatusActive).Return(nil)

		err := graceService.ResolveGracePeriod(ctx, userID, subscriptionID)
		require.NoError(t, err)
		assert.Equal(t, entity.GraceStatusResolved, gracePeriod.Status)
	})

	t.Run("ExpireGracePeriod cancels subscription", func(t *testing.T) {
		gracePeriodRepo := mocks.NewMockGracePeriodRepository()
		subscriptionRepo := mocks.NewMockSubscriptionRepository()
		userRepo := mocks.NewMockUserRepository()
		graceService := service.NewGracePeriodService(gracePeriodRepo, subscriptionRepo, userRepo)

		gracePeriod := entity.NewGracePeriod(uuid.New(), uuid.New(), time.Now().Add(-24*time.Hour))
		gracePeriodRepo.On("GetByID", ctx, gracePeriod.ID).Return(gracePeriod, nil)
		gracePeriodRepo.On("Update", ctx, gracePeriod).Return(nil)
		subscriptionRepo.On("Cancel", ctx, gracePeriod.SubscriptionID).Return(nil)

		err := graceService.ExpireGracePeriod(ctx, gracePeriod.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.GraceStatusExpired, gracePeriod.Status)
	})
}
