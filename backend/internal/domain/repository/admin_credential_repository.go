package repository

import (
	"context"

	"github.com/google/uuid"
)

// AdminCredentialRepository manages password-based auth for admin users.
type AdminCredentialRepository interface {
	// SetPassword stores or replaces the bcrypt password hash for an admin user.
	SetPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error

	// GetPasswordHash returns the stored bcrypt hash for the given user.
	GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error)
}
