package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// CreateGracePeriodCommand creates a new grace period
type CreateGracePeriodCommand struct {
	gracePeriodService *service.GracePeriodService
}

// CreateGracePeriodRequest is the request DTO
type CreateGracePeriodRequest struct {
	UserID         string `json:"user_id" validate:"required,uuid"`
	SubscriptionID string `json:"subscription_id" validate:"required,uuid"`
	DurationDays   int    `json:"duration_days" validate:"required,min=1,max=30"`
}

// CreateGracePeriodResponse is the response DTO
type CreateGracePeriodResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	SubscriptionID string `json:"subscription_id"`
	Status         string `json:"status"`
	ExpiresAt      string `json:"expires_at"`
	DaysRemaining  int    `json:"days_remaining"`
}

// NewCreateGracePeriodCommand creates a new command handler
func NewCreateGracePeriodCommand(gracePeriodService *service.GracePeriodService) *CreateGracePeriodCommand {
	return &CreateGracePeriodCommand{
		gracePeriodService: gracePeriodService,
	}
}

// Execute creates a new grace period
func (c *CreateGracePeriodCommand) Execute(ctx context.Context, req *CreateGracePeriodRequest) (*CreateGracePeriodResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	subscriptionID, err := uuid.Parse(req.SubscriptionID)
	if err != nil {
		return nil, err
	}

	gracePeriod, err := c.gracePeriodService.CreateGracePeriod(ctx, userID, subscriptionID, req.DurationDays)
	if err != nil {
		return nil, err
	}

	return &CreateGracePeriodResponse{
		ID:             gracePeriod.ID.String(),
		UserID:         gracePeriod.UserID.String(),
		SubscriptionID: gracePeriod.SubscriptionID.String(),
		Status:         string(gracePeriod.Status),
		ExpiresAt:      gracePeriod.ExpiresAt.Format(time.RFC3339),
		DaysRemaining:  gracePeriod.DaysRemaining(),
	}, nil
}
