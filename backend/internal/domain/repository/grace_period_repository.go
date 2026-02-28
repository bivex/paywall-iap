package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// GracePeriodRepository defines the interface for grace period data access
type GracePeriodRepository interface {
	// Create creates a new grace period
	Create(ctx context.Context, gracePeriod *entity.GracePeriod) error

	// GetByID retrieves a grace period by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.GracePeriod, error)

	// GetActiveByUserID retrieves the active grace period for a user
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.GracePeriod, error)

	// GetActiveBySubscriptionID retrieves the active grace period for a subscription
	GetActiveBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entity.GracePeriod, error)

	// Update updates an existing grace period
	Update(ctx context.Context, gracePeriod *entity.GracePeriod) error

	// GetExpiredGracePeriods retrieves all expired grace periods that need processing
	GetExpiredGracePeriods(ctx context.Context, limit int) ([]*entity.GracePeriod, error)

	// GetExpiringSoon retrieves grace periods expiring within the specified duration
	GetExpiringSoon(ctx context.Context, withinHours int) ([]*entity.GracePeriod, error)
}
