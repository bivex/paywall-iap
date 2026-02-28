package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// ResolveGracePeriodCommand resolves a grace period
type ResolveGracePeriodCommand struct {
	gracePeriodService *service.GracePeriodService
}

// ResolveGracePeriodRequest is the request DTO
type ResolveGracePeriodRequest struct {
	UserID         string `json:"user_id" validate:"required,uuid"`
	SubscriptionID string `json:"subscription_id" validate:"required,uuid"`
}

// ResolveGracePeriodResponse is the response DTO
type ResolveGracePeriodResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewResolveGracePeriodCommand creates a new command handler
func NewResolveGracePeriodCommand(gracePeriodService *service.GracePeriodService) *ResolveGracePeriodCommand {
	return &ResolveGracePeriodCommand{
		gracePeriodService: gracePeriodService,
	}
}

// Execute resolves a grace period
func (c *ResolveGracePeriodCommand) Execute(ctx context.Context, req *ResolveGracePeriodRequest) (*ResolveGracePeriodResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	subscriptionID, err := uuid.Parse(req.SubscriptionID)
	if err != nil {
		return nil, err
	}

	err = c.gracePeriodService.ResolveGracePeriod(ctx, userID, subscriptionID)
	if err != nil {
		return nil, err
	}

	return &ResolveGracePeriodResponse{
		Success: true,
		Message: "Grace period resolved successfully",
	}, nil
}
