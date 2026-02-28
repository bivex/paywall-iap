package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// WinbackOfferRepository defines the interface for winback offer data access
type WinbackOfferRepository interface {
	// Create creates a new winback offer
	Create(ctx context.Context, offer *entity.WinbackOffer) error

	// GetByID retrieves a winback offer by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.WinbackOffer, error)

	// GetActiveByUserID retrieves all active winback offers for a user
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.WinbackOffer, error)

	// GetActiveByUserAndCampaign retrieves active offer for a user and campaign
	GetActiveByUserAndCampaign(ctx context.Context, userID uuid.UUID, campaignID string) (*entity.WinbackOffer, error)

	// Update updates an existing winback offer
	Update(ctx context.Context, offer *entity.WinbackOffer) error

	// GetExpiredOffers retrieves expired offers that need processing
	GetExpiredOffers(ctx context.Context, limit int) ([]*entity.WinbackOffer, error)
}
