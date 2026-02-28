package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
)

// GetSubscriptionQuery handles getting subscription details
type GetSubscriptionQuery struct {
	subscriptionRepo repository.SubscriptionRepository
}

// NewGetSubscriptionQuery creates a new get subscription query
func NewGetSubscriptionQuery(subscriptionRepo repository.SubscriptionRepository) *GetSubscriptionQuery {
	return &GetSubscriptionQuery{
		subscriptionRepo: subscriptionRepo,
	}
}

// Execute executes the get subscription query
func (q *GetSubscriptionQuery) Execute(ctx context.Context, userID string) (*dto.SubscriptionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user ID", domainErrors.ErrInvalidInput)
	}

	sub, err := q.subscriptionRepo.GetActiveByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return q.toResponse(sub), nil
}

func (q *GetSubscriptionQuery) toResponse(sub *entity.Subscription) *dto.SubscriptionResponse {
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

// CheckAccessQuery handles checking user access
type CheckAccessQuery struct {
	subscriptionRepo repository.SubscriptionRepository
}

// NewCheckAccessQuery creates a new check access query
func NewCheckAccessQuery(subscriptionRepo repository.SubscriptionRepository) *CheckAccessQuery {
	return &CheckAccessQuery{
		subscriptionRepo: subscriptionRepo,
	}
}

// Execute executes the access check query
func (q *CheckAccessQuery) Execute(ctx context.Context, userID string) (*dto.AccessCheckResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user ID", domainErrors.ErrInvalidInput)
	}

	hasAccess, err := q.subscriptionRepo.CanAccess(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	resp := &dto.AccessCheckResponse{
		HasAccess: hasAccess,
	}

	if hasAccess {
		sub, err := q.subscriptionRepo.GetActiveByUserID(ctx, userUUID)
		if err == nil {
			resp.ExpiresAt = sub.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		}
	} else {
		resp.Reason = "no_active_subscription"
	}

	return resp, nil
}
