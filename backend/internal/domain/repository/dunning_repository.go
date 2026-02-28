package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// DunningRepository defines the interface for dunning data access
type DunningRepository interface {
	// Create creates a new dunning process
	Create(ctx context.Context, dunning *entity.Dunning) error

	// GetByID retrieves a dunning record by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Dunning, error)

	// GetActiveBySubscriptionID retrieves the currently active dunning for a subscription
	GetActiveBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entity.Dunning, error)

	// Update updates an existing dunning record
	Update(ctx context.Context, dunning *entity.Dunning) error

	// GetPendingAttempts retrieves dunning records with NextAttemptAt <= NOW()
	GetPendingAttempts(ctx context.Context, limit int) ([]*entity.Dunning, error)
}
