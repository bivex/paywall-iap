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

// DunningService handles dunning management
type DunningService struct {
	dunningRepo      repository.DunningRepository
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
	notificationSvc  *NotificationService
}

// NewDunningService creates a new dunning service
func NewDunningService(
	dunningRepo repository.DunningRepository,
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	notificationSvc *NotificationService,
) *DunningService {
	return &DunningService{
		dunningRepo:      dunningRepo,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		notificationSvc:  notificationSvc,
	}
}

// StartDunning starts a dunning process for a failed subscription renewal
func (s *DunningService) StartDunning(ctx context.Context, subscriptionID, userID uuid.UUID) (*entity.Dunning, error) {
	// Check if active dunning already exists
	existing, err := s.dunningRepo.GetActiveBySubscriptionID(ctx, subscriptionID)
	if err == nil && existing != nil && existing.CanRetry() {
		return existing, nil
	}

	// Create new dunning
	nextAttemptAt := time.Now().Add(24 * time.Hour)
	dunning := entity.NewDunning(subscriptionID, userID, nextAttemptAt)

	// Save dunning
	err = s.dunningRepo.Create(ctx, dunning)
	if err != nil {
		return nil, fmt.Errorf("failed to create dunning: %w", err)
	}

	// Send first retry notification
	_ = s.notificationSvc.SendPaymentRetryNotification(ctx, userID, 1)

	return dunning, nil
}

// ProcessDunningAttempt processes a dunning retry attempt
func (s *DunningService) ProcessDunningAttempt(ctx context.Context, dunningID uuid.UUID, paymentSuccess bool) error {
	// Get dunning
	dunning, err := s.dunningRepo.GetByID(ctx, dunningID)
	if err != nil {
		return errors.New("dunning not found")
	}

	if !dunning.CanRetry() {
		return errors.New("dunning cannot be retried")
	}

	// Increment attempt counter
	dunning.IncrementAttempt()

	if paymentSuccess {
		// Mark as recovered
		dunning.MarkRecovered()
		err = s.dunningRepo.Update(ctx, dunning)
		if err != nil {
			return err
		}

		// Update subscription status
		err = s.subscriptionRepo.UpdateStatus(ctx, dunning.SubscriptionID, entity.StatusActive)
		if err != nil {
			return err
		}

		// Send success notification
		s.notificationSvc.SendPaymentSuccessNotification(ctx, dunning.UserID)
		return nil
	}

	// Payment failed
	if dunning.AttemptCount >= dunning.MaxAttempts {
		// Max attempts reached, mark as failed
		dunning.MarkFailed()
		err = s.dunningRepo.Update(ctx, dunning)
		if err != nil {
			return err
		}

		// Cancel subscription
		err = s.subscriptionRepo.Cancel(ctx, dunning.SubscriptionID)
		if err != nil {
			return err
		}

		// Send final failure notification
		s.notificationSvc.SendPaymentFinalFailureNotification(ctx, dunning.UserID)
		return nil
	}

	// Schedule next attempt
	dunning.NextAttemptAt = time.Now().Add(dunning.GetRetryDelay())
	err = s.dunningRepo.Update(ctx, dunning)
	if err != nil {
		return err
	}

	// Send retry notification
	s.notificationSvc.SendPaymentRetryNotification(ctx, dunning.UserID, dunning.AttemptCount+1)
	return nil
}

// GetPendingDunningAttempts returns dunning processes that need processing
func (s *DunningService) GetPendingDunningAttempts(ctx context.Context, limit int) ([]*entity.Dunning, error) {
	return s.dunningRepo.GetPendingAttempts(ctx, limit)
}
