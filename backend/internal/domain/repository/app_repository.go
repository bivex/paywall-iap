package repository

import (
	"context"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/google/uuid"
)

// AppRepository defines the data access interface for App entities.
type AppRepository interface {
	// GetByID returns an app by its UUID primary key.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.App, error)

	// GetByBundleID looks up an app by its bundle_id (used during JWT issuance).
	GetByBundleID(ctx context.Context, bundleID string) (*entity.App, error)

	// List returns all apps (active and inactive).
	List(ctx context.Context) ([]*entity.App, error)

	// Create inserts a new app and returns it.
	Create(ctx context.Context, name, bundleID, platform string) (*entity.App, error)

	// Update persists changes to an existing app.
	Update(ctx context.Context, app *entity.App) error

	// Delete soft-deletes (deactivates) an app.
	Delete(ctx context.Context, id uuid.UUID) error
}
