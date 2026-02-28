package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

// WinbackOfferRepositoryImpl implements WinbackOfferRepository
type WinbackOfferRepositoryImpl struct {
	pool *pgxpool.Pool
}

// NewWinbackOfferRepository creates a new winback offer repository
func NewWinbackOfferRepository(pool *pgxpool.Pool) repository.WinbackOfferRepository {
	return &WinbackOfferRepositoryImpl{pool: pool}
}

// Create creates a new winback offer
func (r *WinbackOfferRepositoryImpl) Create(ctx context.Context, offer *entity.WinbackOffer) error {
	query := `
		INSERT INTO winback_offers (id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		offer.ID,
		offer.UserID,
		offer.CampaignID,
		offer.DiscountType,
		offer.DiscountValue,
		offer.Status,
		offer.OfferedAt,
		offer.ExpiresAt,
		offer.CreatedAt,
	)

	return err
}

// GetByID retrieves a winback offer by ID
func (r *WinbackOfferRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.WinbackOffer, error) {
	query := `
		SELECT id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, accepted_at, created_at
		FROM winback_offers
		WHERE id = $1
	`

	offer := &entity.WinbackOffer{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&offer.ID,
		&offer.UserID,
		&offer.CampaignID,
		&offer.DiscountType,
		&offer.DiscountValue,
		&offer.Status,
		&offer.OfferedAt,
		&offer.ExpiresAt,
		&offer.AcceptedAt,
		&offer.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return offer, nil
}

// GetActiveByUserID retrieves all active winback offers for a user
func (r *WinbackOfferRepositoryImpl) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.WinbackOffer, error) {
	query := `
		SELECT id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, accepted_at, created_at
		FROM winback_offers
		WHERE user_id = $1 AND status = 'offered' AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*entity.WinbackOffer
	for rows.Next() {
		offer := &entity.WinbackOffer{}
		err := rows.Scan(
			&offer.ID,
			&offer.UserID,
			&offer.CampaignID,
			&offer.DiscountType,
			&offer.DiscountValue,
			&offer.Status,
			&offer.OfferedAt,
			&offer.ExpiresAt,
			&offer.AcceptedAt,
			&offer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}

	return offers, rows.Err()
}

// GetActiveByUserAndCampaign retrieves active offer for a user and campaign
func (r *WinbackOfferRepositoryImpl) GetActiveByUserAndCampaign(ctx context.Context, userID uuid.UUID, campaignID string) (*entity.WinbackOffer, error) {
	query := `
		SELECT id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, accepted_at, created_at
		FROM winback_offers
		WHERE user_id = $1 AND campaign_id = $2 AND status = 'offered' AND expires_at > NOW()
		LIMIT 1
	`

	offer := &entity.WinbackOffer{}
	err := r.pool.QueryRow(ctx, query, userID, campaignID).Scan(
		&offer.ID,
		&offer.UserID,
		&offer.CampaignID,
		&offer.DiscountType,
		&offer.DiscountValue,
		&offer.Status,
		&offer.OfferedAt,
		&offer.ExpiresAt,
		&offer.AcceptedAt,
		&offer.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return offer, nil
}

// Update updates an existing winback offer
func (r *WinbackOfferRepositoryImpl) Update(ctx context.Context, offer *entity.WinbackOffer) error {
	query := `
		UPDATE winback_offers
		SET status = $2, accepted_at = $3
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		offer.ID,
		offer.Status,
		offer.AcceptedAt,
	)

	return err
}

// GetExpiredOffers retrieves expired offers that need processing
func (r *WinbackOfferRepositoryImpl) GetExpiredOffers(ctx context.Context, limit int) ([]*entity.WinbackOffer, error) {
	query := `
		SELECT id, user_id, campaign_id, discount_type, discount_value, status, offered_at, expires_at, accepted_at, created_at
		FROM winback_offers
		WHERE status = 'offered' AND expires_at < NOW()
		ORDER BY expires_at ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*entity.WinbackOffer
	for rows.Next() {
		offer := &entity.WinbackOffer{}
		err := rows.Scan(
			&offer.ID,
			&offer.UserID,
			&offer.CampaignID,
			&offer.DiscountType,
			&offer.DiscountValue,
			&offer.Status,
			&offer.OfferedAt,
			&offer.ExpiresAt,
			&offer.AcceptedAt,
			&offer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}

	return offers, rows.Err()
}

var _ repository.WinbackOfferRepository = (*WinbackOfferRepositoryImpl)(nil)
