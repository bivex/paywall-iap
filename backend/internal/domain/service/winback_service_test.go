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

func TestWinbackService(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	winbackRepo := mocks.NewMockWinbackOfferRepository()
	userRepo := mocks.NewMockUserRepository()
	subRepo := mocks.NewMockSubscriptionRepository()

	winbackService := service.NewWinbackService(winbackRepo, userRepo, subRepo)

	t.Run("CreateWinbackOffer success", func(t *testing.T) {
		userID := uuid.New()

		userRepo.On("GetByID", ctx, userID).Return(&entity.User{ID: userID}, nil).Once()
		winbackRepo.On("GetActiveByUserAndCampaign", ctx, userID, "campaign_123").Return(nil, errors.New("not found")).Once()
		winbackRepo.On("Create", ctx, mock.Anything).Return(nil).Once()

		offer, err := winbackService.CreateWinbackOffer(ctx, userID, "campaign_123", entity.DiscountTypePercentage, 25.0, 30)
		require.NoError(t, err)
		assert.Equal(t, entity.OfferStatusOffered, offer.Status)
		assert.Equal(t, 25.0, offer.DiscountValue)
	})

	t.Run("AcceptWinbackOffer success", func(t *testing.T) {
		userID := uuid.New()
		offerID := uuid.New()

		offer := entity.NewWinbackOffer(userID, "campaign_123", entity.DiscountTypePercentage, 25.0, time.Now().Add(30*24*time.Hour))
		offer.ID = offerID

		winbackRepo.On("GetByID", ctx, offerID).Return(offer, nil).Once()
		winbackRepo.On("Update", ctx, offer).Return(nil).Once()

		acceptedOffer, err := winbackService.AcceptWinbackOffer(ctx, userID, offerID)
		require.NoError(t, err)
		assert.Equal(t, entity.OfferStatusAccepted, acceptedOffer.Status)
		assert.NotNil(t, acceptedOffer.AcceptedAt)
	})

	t.Run("AcceptWinbackOffer with expired offer returns error", func(t *testing.T) {
		userID := uuid.New()
		offerID := uuid.New()

		offer := entity.NewWinbackOffer(userID, "campaign_123", entity.DiscountTypePercentage, 25.0, time.Now().Add(-24*time.Hour))
		offer.ID = offerID

		winbackRepo.On("GetByID", ctx, offerID).Return(offer, nil).Once()

		acceptedOffer, err := winbackService.AcceptWinbackOffer(ctx, userID, offerID)
		assert.Error(t, err)
		assert.Nil(t, acceptedOffer)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("Calculate discount for percentage offer", func(t *testing.T) {
		offer := entity.NewWinbackOffer(uuid.New(), "campaign_123", entity.DiscountTypePercentage, 25.0, time.Now().Add(30*24*time.Hour))

		discount := offer.CalculateDiscountAmount(100.0)
		assert.InDelta(t, 25.0, discount, 0.01)

		finalAmount := offer.CalculateFinalAmount(100.0)
		assert.InDelta(t, 75.0, finalAmount, 0.01)
	})

	t.Run("Calculate discount for fixed offer", func(t *testing.T) {
		offer := entity.NewWinbackOffer(uuid.New(), "campaign_123", entity.DiscountTypeFixed, 20.0, time.Now().Add(30*24*time.Hour))

		discount := offer.CalculateDiscountAmount(100.0)
		assert.InDelta(t, 20.0, discount, 0.01)

		finalAmount := offer.CalculateFinalAmount(100.0)
		assert.InDelta(t, 80.0, finalAmount, 0.01)
	})
}
