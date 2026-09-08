package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
)

// RestoreSubscriptionCommand handles subscription restoration
type RestoreSubscriptionCommand struct {
	subscriptionRepo repository.SubscriptionRepository
}

// NewRestoreSubscriptionCommand creates a new restore subscription command
func NewRestoreSubscriptionCommand(subscriptionRepo repository.SubscriptionRepository) *RestoreSubscriptionCommand {
	return &RestoreSubscriptionCommand{
		subscriptionRepo: subscriptionRepo,
	}
}

// Execute executes the restore subscription command
func (c *RestoreSubscriptionCommand) Execute(ctx context.Context, userID string) (*dto.SubscriptionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user ID", domainErrors.ErrInvalidInput)
	}

	// 1. Check if user already has an active subscription
	sub, err := c.subscriptionRepo.GetActiveByUserID(ctx, userUUID)
	if err == nil && sub != nil {
		return c.toResponse(sub), nil
	}

	// 2. Check all subscriptions for user
	subs, err := c.subscriptionRepo.GetByUserID(ctx, userUUID)
	if err != nil || len(subs) == 0 {
		return nil, fmt.Errorf("no subscription found to restore: %w", domainErrors.ErrSubscriptionNotFound)
	}

	// Pick the latest subscription
	latestSub := subs[0]
	latestSub.Status = entity.StatusActive
	latestSub.AutoRenew = true

	// If it already expired, extend by 30 days so the user has active access upon restore
	if !latestSub.ExpiresAt.After(time.Now()) {
		latestSub.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
	}

	if err := c.subscriptionRepo.Update(ctx, latestSub); err != nil {
		return nil, fmt.Errorf("failed to reactivate subscription: %w", err)
	}

	return c.toResponse(latestSub), nil
}

func (c *RestoreSubscriptionCommand) toResponse(sub *entity.Subscription) *dto.SubscriptionResponse {
	return &dto.SubscriptionResponse{
		ID:        sub.ID.String(),
		Status:    string(sub.Status),
		Source:    string(sub.Source),
		Platform:  sub.Platform,
		ProductID: sub.ProductID,
		PlanType:  string(sub.PlanType),
		ExpiresAt: sub.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		AutoRenew: sub.AutoRenew,
		CreatedAt: sub.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: sub.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
