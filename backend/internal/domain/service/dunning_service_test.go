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

func TestDunningService(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	dunningRepo := mocks.NewMockDunningRepository()
	subscriptionRepo := mocks.NewMockSubscriptionRepository()
	userRepo := mocks.NewMockUserRepository()
	notificationSvc := service.NewNotificationService()

	dunningService := service.NewDunningService(dunningRepo, subscriptionRepo, userRepo, notificationSvc)

	t.Run("StartDunning creates new dunning process", func(t *testing.T) {
		subscriptionID := uuid.New()
		userID := uuid.New()

		dunningRepo.On("GetActiveBySubscriptionID", ctx, subscriptionID).Return(nil, errors.New("not found")).Once()
		dunningRepo.On("Create", ctx, mock.Anything).Return(nil).Once()

		dunning, err := dunningService.StartDunning(ctx, subscriptionID, userID)
		require.NoError(t, err)
		assert.Equal(t, entity.DunningStatusPending, dunning.Status)
		assert.Equal(t, 0, dunning.AttemptCount)
	})

	t.Run("ProcessDunningAttempt with success recovers subscription", func(t *testing.T) {
		dunningID := uuid.New()
		subscriptionID := uuid.New()
		userID := uuid.New()

		dunning := entity.NewDunning(subscriptionID, userID, time.Now())
		dunning.ID = dunningID

		dunningRepo.On("GetByID", ctx, dunningID).Return(dunning, nil).Once()
		dunningRepo.On("Update", ctx, dunning).Return(nil).Once()
		subscriptionRepo.On("UpdateStatus", ctx, subscriptionID, entity.StatusActive).Return(nil).Once()

		err := dunningService.ProcessDunningAttempt(ctx, dunningID, true)
		require.NoError(t, err)
		assert.Equal(t, entity.DunningStatusRecovered, dunning.Status)
		assert.NotNil(t, dunning.RecoveredAt)
	})

	t.Run("ProcessDunningAttempt with failure after max attempts fails", func(t *testing.T) {
		dunningID := uuid.New()
		subscriptionID := uuid.New()
		userID := uuid.New()

		dunning := entity.NewDunning(subscriptionID, userID, time.Now())
		dunning.ID = dunningID
		dunning.AttemptCount = 4 // One more will hit max

		dunningRepo.On("GetByID", ctx, dunningID).Return(dunning, nil).Once()
		dunningRepo.On("Update", ctx, dunning).Return(nil).Once()
		subscriptionRepo.On("Cancel", ctx, subscriptionID).Return(nil).Once()

		err := dunningService.ProcessDunningAttempt(ctx, dunningID, false)
		require.NoError(t, err)
		assert.Equal(t, entity.DunningStatusFailed, dunning.Status)
		assert.NotNil(t, dunning.FailedAt)
	})
}
