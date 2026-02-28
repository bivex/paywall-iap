package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

func TestWinbackOfferEntity(t *testing.T) {
	t.Run("NewWinbackOffer creates offered winback", func(t *testing.T) {
		userID := uuid.New()
		campaignID := "winback_2026_q1"
		expiresAt := time.Now().Add(30 * 24 * time.Hour)

		offer := entity.NewWinbackOffer(userID, campaignID, entity.DiscountTypePercentage, 50.0, expiresAt)

		assert.NotEmpty(t, offer.ID)
		assert.Equal(t, userID, offer.UserID)
		assert.Equal(t, campaignID, offer.CampaignID)
		assert.Equal(t, entity.DiscountTypePercentage, offer.DiscountType)
		assert.Equal(t, 50.0, offer.DiscountValue)
		assert.Equal(t, entity.OfferStatusOffered, offer.Status)
		assert.Equal(t, expiresAt.Unix(), offer.ExpiresAt.Unix())
		assert.Nil(t, offer.AcceptedAt)
	})

	t.Run("Accept marks offer as accepted", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0,
			time.Now().Add(30*24*time.Hour),
		)

		err := offer.Accept()
		assert.NoError(t, err)
		assert.Equal(t, entity.OfferStatusAccepted, offer.Status)
		assert.NotNil(t, offer.AcceptedAt)
	})

	t.Run("Cannot accept expired offer", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0,
			time.Now().Add(-24*time.Hour),
		)

		err := offer.Accept()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot accept expired offer")
	})

	t.Run("CalculateDiscountAmount for percentage discount", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0, // 25% off
			time.Now().Add(30*24*time.Hour),
		)

		amount := offer.CalculateDiscountAmount(99.99)
		assert.InDelta(t, 25.0, amount, 0.01) // 25% of 99.99 = 24.9975
	})

	t.Run("CalculateDiscountAmount for fixed discount", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypeFixed,
			20.0, // $20 off
			time.Now().Add(30*24*time.Hour),
		)

		amount := offer.CalculateDiscountAmount(99.99)
		assert.InDelta(t, 20.0, amount, 0.01)
	})

	t.Run("Discount cannot exceed total amount", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypeFixed,
			50.0, // $50 off
			time.Now().Add(30*24*time.Hour),
		)

		amount := offer.CalculateDiscountAmount(30.0) // Total is only $30
		assert.InDelta(t, 30.0, amount, 0.01)         // Should cap at total amount
	})

	t.Run("Expire marks offer as expired", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0,
			time.Now().Add(-24*time.Hour),
		)

		err := offer.Expire()
		assert.NoError(t, err)
		assert.Equal(t, entity.OfferStatusExpired, offer.Status)
	})

	t.Run("Decline marks offer as declined", func(t *testing.T) {
		offer := entity.NewWinbackOffer(
			uuid.New(),
			"campaign_123",
			entity.DiscountTypePercentage,
			25.0,
			time.Now().Add(30*24*time.Hour),
		)

		err := offer.Decline()
		assert.NoError(t, err)
		assert.Equal(t, entity.OfferStatusDeclined, offer.Status)
	})
}
