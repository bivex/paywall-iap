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
	ErrGracePeriodNotFound      = errors.New("grace period not found")
	ErrGracePeriodAlreadyExists = errors.New("active grace period already exists for this subscription")
	ErrGracePeriodNotActive     = errors.New("grace period is not active")
)

// GracePeriodService handles grace period business logic
type GracePeriodService struct {
	gracePeriodRepo  repository.GracePeriodRepository
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
}

// NewGracePeriodService creates a new grace period service
func NewGracePeriodService(
	gracePeriodRepo repository.GracePeriodRepository,
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
) *GracePeriodService {
	return &GracePeriodService{
		gracePeriodRepo:  gracePeriodRepo,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
	}
}

// CreateGracePeriod creates a new grace period for a subscription
func (s *GracePeriodService) CreateGracePeriod(ctx context.Context, userID, subscriptionID uuid.UUID, durationDays int) (*entity.GracePeriod, error) {
	// Check if active grace period already exists
	existing, err := s.gracePeriodRepo.GetActiveBySubscriptionID(ctx, subscriptionID)
	if err == nil && existing != nil && existing.IsActive() {
		return nil, ErrGracePeriodAlreadyExists
	}

	// Verify subscription exists and belongs to user
	sub, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	if sub.UserID != userID {
		return nil, errors.New("subscription does not belong to user")
	}

	// Create grace period
	expiresAt := time.Now().Add(time.Duration(durationDays) * 24 * time.Hour)
	gracePeriod := entity.NewGracePeriod(userID, subscriptionID, expiresAt)

	// Update subscription status to grace
	err = s.subscriptionRepo.UpdateStatus(ctx, subscriptionID, entity.StatusGrace)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription status: %w", err)
	}

	// Save grace period
	err = s.gracePeriodRepo.Create(ctx, gracePeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to create grace period: %w", err)
	}

	return gracePeriod, nil
}

// ResolveGracePeriod resolves a grace period when payment succeeds
func (s *GracePeriodService) ResolveGracePeriod(ctx context.Context, userID, subscriptionID uuid.UUID) error {
	// Get active grace period
	gracePeriod, err := s.gracePeriodRepo.GetActiveBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return ErrGracePeriodNotFound
	}

	if !gracePeriod.IsActive() {
		return ErrGracePeriodNotActive
	}

	// Resolve grace period
	err = gracePeriod.Resolve()
	if err != nil {
		return err
	}

	// Update repository
	err = s.gracePeriodRepo.Update(ctx, gracePeriod)
	if err != nil {
		return fmt.Errorf("failed to update grace period: %w", err)
	}

	// Update subscription status back to active
	err = s.subscriptionRepo.UpdateStatus(ctx, subscriptionID, entity.StatusActive)
	if err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	return nil
}

// ExpireGracePeriod expires a grace period and cancels the subscription
func (s *GracePeriodService) ExpireGracePeriod(ctx context.Context, gracePeriodID uuid.UUID) error {
	// Get grace period
	gracePeriod, err := s.gracePeriodRepo.GetByID(ctx, gracePeriodID)
	if err != nil {
		return ErrGracePeriodNotFound
	}

	if gracePeriod.Status != entity.GraceStatusActive {
		return ErrGracePeriodNotActive
	}

	// Expire grace period
	err = gracePeriod.Expire()
	if err != nil {
		return err
	}

	// Update repository
	err = s.gracePeriodRepo.Update(ctx, gracePeriod)
	if err != nil {
		return fmt.Errorf("failed to update grace period: %w", err)
	}

	// Cancel subscription
	err = s.subscriptionRepo.Cancel(ctx, gracePeriod.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}

// ProcessExpiredGracePeriods processes all expired grace periods
func (s *GracePeriodService) ProcessExpiredGracePeriods(ctx context.Context, limit int) (int, error) {
	expiredPeriods, err := s.gracePeriodRepo.GetExpiredGracePeriods(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired grace periods: %w", err)
	}

	processed := 0
	for _, gp := range expiredPeriods {
		err := s.ExpireGracePeriod(ctx, gp.ID)
		if err != nil {
			// Log error but continue processing others
			continue
		}
		processed++
	}

	return processed, nil
}

// GetGracePeriodStatus returns the current grace period status for a user
func (s *GracePeriodService) GetGracePeriodStatus(ctx context.Context, userID uuid.UUID) (*entity.GracePeriod, error) {
	gracePeriod, err := s.gracePeriodRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, ErrGracePeriodNotFound
	}

	return gracePeriod, nil
}

// NotifyExpiringSoon returns grace periods expiring within the specified hours
func (s *GracePeriodService) NotifyExpiringSoon(ctx context.Context, withinHours int) ([]*entity.GracePeriod, error) {
	return s.gracePeriodRepo.GetExpiringSoon(ctx, withinHours)
}
