package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

func TestGracePeriodEntity(t *testing.T) {
	t.Run("NewGracePeriod creates active grace period", func(t *testing.T) {
		userID := uuid.New()
		subscriptionID := uuid.New()
		expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

		gracePeriod := entity.NewGracePeriod(userID, subscriptionID, expiresAt)

		assert.NotEmpty(t, gracePeriod.ID)
		assert.Equal(t, userID, gracePeriod.UserID)
		assert.Equal(t, subscriptionID, gracePeriod.SubscriptionID)
		assert.Equal(t, entity.GraceStatusActive, gracePeriod.Status)
		assert.Equal(t, expiresAt, gracePeriod.ExpiresAt)
		assert.Nil(t, gracePeriod.ResolvedAt)
	})

	t.Run("IsActive returns true for active grace period", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(24*time.Hour),
		)

		assert.True(t, gracePeriod.IsActive())
	})

	t.Run("IsExpired returns true after expiry", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(-24*time.Hour), // Expired yesterday
		)

		assert.True(t, gracePeriod.IsExpired())
	})

	t.Run("Resolve marks grace period as resolved", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(24*time.Hour),
		)

		err := gracePeriod.Resolve()
		assert.NoError(t, err)
		assert.Equal(t, entity.GraceStatusResolved, gracePeriod.Status)
		assert.NotNil(t, gracePeriod.ResolvedAt)
	})

	t.Run("Expire marks grace period as expired", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(-24*time.Hour),
		)

		err := gracePeriod.Expire()
		assert.NoError(t, err)
		assert.Equal(t, entity.GraceStatusExpired, gracePeriod.Status)
	})

	t.Run("Cannot resolve already expired grace period", func(t *testing.T) {
		gracePeriod := entity.NewGracePeriod(
			uuid.New(),
			uuid.New(),
			time.Now().Add(-24*time.Hour),
		)
		gracePeriod.Expire()

		err := gracePeriod.Resolve()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot resolve expired grace period")
	})
}
