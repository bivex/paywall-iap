package repository

import (
	"context"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/google/uuid"
)

// SubscriptionReader defines read operations on Subscription entities.
type SubscriptionReader interface {
	// GetByID retrieves a subscription by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)

	// GetActiveByUserID retrieves the active subscription for a user
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error)

	// GetByUserID retrieves all subscriptions for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error)

	// CanAccess checks if a user can access premium content
	CanAccess(ctx context.Context, userID uuid.UUID) (bool, error)

	// GetUsersWithCancelledSubscriptions retrieves users whose subscriptions were cancelled recently
	GetUsersWithCancelledSubscriptions(ctx context.Context, daysSinceChurn int) ([]uuid.UUID, error)

	// GetTotalRevenue returns the total revenue for a user across all transactions
	GetTotalRevenue(ctx context.Context, userID uuid.UUID) (float64, error)
}

// SubscriptionWriter defines mutation operations on Subscription entities.
type SubscriptionWriter interface {
	// Create creates a new subscription
	Create(ctx context.Context, subscription *entity.Subscription) error

	// Update updates an existing subscription
	Update(ctx context.Context, subscription *entity.Subscription) error

	// UpdateStatus updates the status of a subscription
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SubscriptionStatus) error

	// UpdateExpiry updates the expiry date of a subscription
	UpdateExpiry(ctx context.Context, id uuid.UUID, expiresAt interface{}) error

	// Cancel cancels a subscription
	Cancel(ctx context.Context, id uuid.UUID) error
}

// SubscriptionRepository aggregates subscription repository capabilities via composition.
type SubscriptionRepository interface {
	SubscriptionReader
	SubscriptionWriter
}
