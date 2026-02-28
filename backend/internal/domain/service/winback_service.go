package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

var (
	ErrWinbackOfferNotFound  = errors.New("winback offer not found")
	ErrWinbackOfferNotActive = errors.New("winback offer is not active")
	ErrCampaignNotFound      = errors.New("campaign not found")
)

// WinbackService handles winback offer business logic
type WinbackService struct {
	winbackRepo repository.WinbackOfferRepository
	userRepo    repository.UserRepository
	subRepo     repository.SubscriptionRepository
}

// NewWinbackService creates a new winback service
func NewWinbackService(
	winbackRepo repository.WinbackOfferRepository,
	userRepo repository.UserRepository,
	subRepo repository.SubscriptionRepository,
) *WinbackService {
	return &WinbackService{
		winbackRepo: winbackRepo,
		userRepo:    userRepo,
		subRepo:     subRepo,
	}
}

// CreateWinbackOffer creates a new winback offer for a user
func (s *WinbackService) CreateWinbackOffer(
	ctx context.Context,
	userID uuid.UUID,
	campaignID string,
	discountType entity.DiscountType,
	discountValue float64,
	durationDays int,
) (*entity.WinbackOffer, error) {
	// Check if user already has an active offer for this campaign
	existing, err := s.winbackRepo.GetActiveByUserAndCampaign(ctx, userID, campaignID)
	if err == nil && existing != nil && existing.IsActive() {
		return nil, errors.New("user already has an active offer for this campaign")
	}

	// Verify user exists
	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Create winback offer
	expiresAt := time.Now().Add(time.Duration(durationDays) * 24 * time.Hour)
	offer := entity.NewWinbackOffer(userID, campaignID, discountType, discountValue, expiresAt)

	// Save offer
	err = s.winbackRepo.Create(ctx, offer)
	if err != nil {
		return nil, fmt.Errorf("failed to create winback offer: %w", err)
	}

	return offer, nil
}

// AcceptWinbackOffer accepts a winback offer and applies the discount
func (s *WinbackService) AcceptWinbackOffer(ctx context.Context, userID, offerID uuid.UUID) (*entity.WinbackOffer, error) {
	// Get offer
	offer, err := s.winbackRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, ErrWinbackOfferNotFound
	}

	// Verify offer belongs to user
	if offer.UserID != userID {
		return nil, errors.New("offer does not belong to user")
	}

	// Accept offer
	err = offer.Accept()
	if err != nil {
		return nil, err
	}

	// Update repository
	err = s.winbackRepo.Update(ctx, offer)
	if err != nil {
		return nil, fmt.Errorf("failed to update winback offer: %w", err)
	}

	return offer, nil
}

// GetActiveWinbackOffers returns all active winback offers for a user
func (s *WinbackService) GetActiveWinbackOffers(ctx context.Context, userID uuid.UUID) ([]*entity.WinbackOffer, error) {
	return s.winbackRepo.GetActiveByUserID(ctx, userID)
}

// ProcessExpiredWinbackOffers expires all expired winback offers
func (s *WinbackService) ProcessExpiredWinbackOffers(ctx context.Context, limit int) (int, error) {
	expiredOffers, err := s.winbackRepo.GetExpiredOffers(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired offers: %w", err)
	}

	processed := 0
	for _, offer := range expiredOffers {
		err := offer.Expire()
		if err != nil {
			continue
		}

		err = s.winbackRepo.Update(ctx, offer)
		if err != nil {
			continue
		}
		processed++
	}

	return processed, nil
}

// CreateWinbackCampaignForChurnedUsers creates winback offers for recently churned users
func (s *WinbackService) CreateWinbackCampaignForChurnedUsers(
	ctx context.Context,
	campaignID string,
	discountType entity.DiscountType,
	discountValue float64,
	durationDays int,
	daysSinceChurn int,
) (int, error) {
	// Get churned users (subscriptions cancelled within specified days)
	churnedUsers, err := s.subRepo.GetUsersWithCancelledSubscriptions(ctx, daysSinceChurn)
	if err != nil {
		return 0, fmt.Errorf("failed to get churned users: %w", err)
	}

	created := 0
	for _, userID := range churnedUsers {
		_, err := s.CreateWinbackOffer(ctx, userID, campaignID, discountType, discountValue, durationDays)
		if err != nil {
			// Skip users who already have offers
			continue
		}
		created++
	}

	return created, nil
}
