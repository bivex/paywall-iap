package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetByPlatformID retrieves a user by platform user ID
	GetByPlatformID(ctx context.Context, platformUserID string) (*entity.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *entity.User) error

	// SoftDelete soft deletes a user
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// ExistsByPlatformID checks if a user exists with the given platform ID
	ExistsByPlatformID(ctx context.Context, platformUserID string) (bool, error)

	// UpdatePurchaseChannel sets the purchase channel for a user
	UpdatePurchaseChannel(ctx context.Context, id uuid.UUID, channel string) error

	// UpdateEmail updates the email address of a user
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) error

	// IncrementSessionCount increments the session count for a user and returns the new count
	IncrementSessionCount(ctx context.Context, id uuid.UUID) (int, error)

	// UpdateHasViewedAds updates the has_viewed_ads flag for a user
	UpdateHasViewedAds(ctx context.Context, id uuid.UUID, hasViewedAds bool) error
}
