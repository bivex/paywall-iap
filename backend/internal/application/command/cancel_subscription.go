package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/password9090/paywall-iap/internal/domain/repository"
	domainErrors "github.com/password9090/paywall-iap/internal/domain/errors"
)

// CancelSubscriptionCommand handles subscription cancellation
type CancelSubscriptionCommand struct {
	subscriptionRepo repository.SubscriptionRepository
}

// NewCancelSubscriptionCommand creates a new cancel subscription command
func NewCancelSubscriptionCommand(subscriptionRepo repository.SubscriptionRepository) *CancelSubscriptionCommand {
	return &CancelSubscriptionCommand{
		subscriptionRepo: subscriptionRepo,
	}
}

// Execute executes the cancel subscription command
func (c *CancelSubscriptionCommand) Execute(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("%w: invalid user ID", domainErrors.ErrInvalidInput)
	}

	// Get active subscription
	sub, err := c.subscriptionRepo.GetActiveByUserID(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("no active subscription found: %w", domainErrors.ErrSubscriptionNotActive)
	}

	// Cancel subscription
	if err := c.subscriptionRepo.Cancel(ctx, sub.ID); err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}
