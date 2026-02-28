package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/password9090/paywall-iap/internal/domain/entity"
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
}
