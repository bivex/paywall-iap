package testutil

import (
	"time"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// UserFactory creates test user entities
type UserFactory struct{}

func NewUserFactory() *UserFactory {
	return &UserFactory{}
}

func (f *UserFactory) Create(platform entity.Platform, withEmail bool) *entity.User {
	platformUserID := uuid.New().String()
	var email string
	if withEmail {
		email = "test_" + uuid.New().String()[:8] + "@example.com"
	}

	return entity.NewUser(entity.NewUserParams{
		PlatformUserID: platformUserID,
		DeviceID:       "test-device-" + uuid.New().String()[:8],
		Platform:       platform,
		AppVersion:     "1.0.0",
		Email:          email,
		AppID:          uuid.Nil,
	})
}

func (f *UserFactory) CreateWithPlatformUserID(platformUserID string, platform entity.Platform) *entity.User {
	return entity.NewUser(entity.NewUserParams{
		PlatformUserID: platformUserID,
		DeviceID:       "test-device-" + uuid.New().String()[:8],
		Platform:       platform,
		AppVersion:     "1.0.0",
		Email:          "test_" + platformUserID[:8] + "@example.com",
		AppID:          uuid.Nil,
	})
}

// SubscriptionFactory creates test subscription entities
type SubscriptionFactory struct{}

func NewSubscriptionFactory() *SubscriptionFactory {
	return &SubscriptionFactory{}
}

func (f *SubscriptionFactory) CreateActive(userID uuid.UUID, source entity.SubscriptionSource, planType entity.PlanType) *entity.Subscription {
	sub := entity.NewSubscription(entity.NewSubscriptionParams{
		UserID:    userID,
		Source:    source,
		Platform:  "ios",
		ProductID: "com.app.premium",
		PlanType:  planType,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})
	return sub
}

func (f *SubscriptionFactory) CreateExpired(userID uuid.UUID) *entity.Subscription {
	sub := entity.NewSubscription(entity.NewSubscriptionParams{
		UserID:    userID,
		Source:    entity.SourceIAP,
		Platform:  "ios",
		ProductID: "com.app.premium",
		PlanType:  entity.PlanMonthly,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	})
	sub.Status = entity.StatusExpired
	return sub
}

func (f *SubscriptionFactory) CreateCancelled(userID uuid.UUID) *entity.Subscription {
	sub := entity.NewSubscription(entity.NewSubscriptionParams{
		UserID:    userID,
		Source:    entity.SourceIAP,
		Platform:  "ios",
		ProductID: "com.app.premium",
		PlanType:  entity.PlanMonthly,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})
	sub.Status = entity.StatusCancelled
	sub.AutoRenew = false
	return sub
}

// TransactionFactory creates test transaction entities
type TransactionFactory struct{}

func NewTransactionFactory() *TransactionFactory {
	return &TransactionFactory{}
}

func (f *TransactionFactory) CreateSuccessful(userID, subscriptionID uuid.UUID, amount float64) *entity.Transaction {
	tx := entity.NewTransaction(entity.NewTransactionParams{
		AppID:          uuid.Nil,
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Amount:         amount,
		Currency:       "USD",
	})
	tx.Status = entity.TransactionStatusSuccess
	tx.ReceiptHash = "sha256_" + uuid.New().String()
	tx.ProviderTxID = "tx_" + uuid.New().String()
	return tx
}

func (f *TransactionFactory) CreateFailed(userID, subscriptionID uuid.UUID) *entity.Transaction {
	tx := entity.NewTransaction(entity.NewTransactionParams{
		AppID:          uuid.Nil,
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Amount:         9.99,
		Currency:       "USD",
	})
	tx.Status = entity.TransactionStatusFailed
	return tx
}
