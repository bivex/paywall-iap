package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

func TestNewUser(t *testing.T) {
	user := entity.NewUser(entity.NewUserParams{
		PlatformUserID: "apple-123",
		DeviceID:       "device-456",
		Platform:       entity.PlatformiOS,
		AppVersion:     "1.0.0",
		Email:          "test@example.com",
		AppID:          uuid.Nil,
	})

	assert.NotNil(t, user.ID)
	assert.Equal(t, "apple-123", user.PlatformUserID)
	assert.Equal(t, "device-456", user.DeviceID)
	assert.Equal(t, entity.PlatformiOS, user.Platform)
	assert.Equal(t, "1.0.0", user.AppVersion)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, 0.0, user.LTV)
	assert.False(t, user.IsDeleted())
	assert.True(t, user.HasEmail())
}

func TestUser_HasEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "with email",
			email:    "test@example.com",
			expected: true,
		},
		{
			name:     "without email",
			email:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := entity.NewUser(entity.NewUserParams{
				PlatformUserID: "id",
				DeviceID:       "device",
				Platform:       entity.PlatformiOS,
				AppVersion:     "1.0",
				Email:          tt.email,
				AppID:          uuid.Nil,
			})
			assert.Equal(t, tt.expected, user.HasEmail())
		})
	}
}

func TestNewSubscription(t *testing.T) {
	userID := uuid.New()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	sub := entity.NewSubscription(entity.NewSubscriptionParams{
		UserID:    userID,
		Source:    entity.SourceIAP,
		Platform:  "ios",
		ProductID: "com.app.premium",
		PlanType:  entity.PlanMonthly,
		ExpiresAt: expiresAt,
	})

	assert.NotNil(t, sub.ID)
	assert.Equal(t, userID, sub.UserID)
	assert.Equal(t, entity.StatusActive, sub.Status)
	assert.Equal(t, entity.SourceIAP, sub.Source)
	assert.Equal(t, "ios", sub.Platform)
	assert.Equal(t, "com.app.premium", sub.ProductID)
	assert.Equal(t, entity.PlanMonthly, sub.PlanType)
	assert.True(t, sub.AutoRenew)
}

func TestSubscription_IsActive(t *testing.T) {
	tests := []struct {
		name       string
		expiresAt  time.Time
		status     entity.SubscriptionStatus
		deletedAt  *time.Time
		expected   bool
	}{
		{
			name:      "active subscription",
			expiresAt: time.Now().Add(1 * time.Hour),
			status:    entity.StatusActive,
			expected:  true,
		},
		{
			name:      "expired subscription",
			expiresAt: time.Now().Add(-1 * time.Hour),
			status:    entity.StatusActive,
			expected:  false,
		},
		{
			name:      "cancelled subscription",
			expiresAt: time.Now().Add(1 * time.Hour),
			status:    entity.StatusCancelled,
			expected:  false,
		},
		{
			name:      "deleted subscription",
			expiresAt: time.Now().Add(1 * time.Hour),
			status:    entity.StatusActive,
			deletedAt: func() *time.Time { t := time.Now(); return &t }(),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &entity.Subscription{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				Status:    tt.status,
				ExpiresAt: tt.expiresAt,
				DeletedAt: tt.deletedAt,
			}
			assert.Equal(t, tt.expected, sub.IsActive())
		})
	}
}

func TestSubscription_CanAccessContent(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		status    entity.SubscriptionStatus
		deletedAt *time.Time
		expected  bool
	}{
		{
			name:      "active and not expired",
			expiresAt: time.Now().Add(1 * time.Hour),
			status:    entity.StatusActive,
			expected:  true,
		},
		{
			name:      "expired",
			expiresAt: time.Now().Add(-1 * time.Hour),
			status:    entity.StatusActive,
			expected:  false,
		},
		{
			name:      "deleted",
			expiresAt: time.Now().Add(1 * time.Hour),
			status:    entity.StatusActive,
			deletedAt: func() *time.Time { t := time.Now(); return &t }(),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &entity.Subscription{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				Status:    tt.status,
				ExpiresAt: tt.expiresAt,
				DeletedAt: tt.deletedAt,
			}
			assert.Equal(t, tt.expected, sub.CanAccessContent())
		})
	}
}

func TestNewTransaction(t *testing.T) {
	appID := uuid.New()
	userID := uuid.New()
	subID := uuid.New()

	txn := entity.NewTransaction(entity.NewTransactionParams{
		AppID:          appID,
		UserID:         userID,
		SubscriptionID: subID,
		Amount:         9.99,
		Currency:       "USD",
	})

	assert.NotNil(t, txn.ID)
	assert.Equal(t, appID, txn.AppID)
	assert.Equal(t, userID, txn.UserID)
	assert.Equal(t, subID, txn.SubscriptionID)
	assert.Equal(t, 9.99, txn.Amount)
	assert.Equal(t, "USD", txn.Currency)
	assert.Equal(t, entity.TransactionStatusSuccess, txn.Status)
	assert.True(t, txn.IsSuccessful())
	assert.False(t, txn.IsFailed())
}
