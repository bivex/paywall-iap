package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
	// Create creates a new subscription
	Create(ctx context.Context, subscription *entity.Subscription) error

	// GetByID retrieves a subscription by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)

	// GetActiveByUserID retrieves the active subscription for a user
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error)

	// GetByUserID retrieves all subscriptions for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error)

	// Update updates an existing subscription
	Update(ctx context.Context, subscription *entity.Subscription) error

	// UpdateStatus updates the status of a subscription
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SubscriptionStatus) error

	// UpdateExpiry updates the expiry date of a subscription
	UpdateExpiry(ctx context.Context, id uuid.UUID, expiresAt interface{}) error

	// Cancel cancels a subscription
	Cancel(ctx context.Context, id uuid.UUID) error

	// CanAccess checks if a user can access premium content
	CanAccess(ctx context.Context, userID uuid.UUID) (bool, error)
}
