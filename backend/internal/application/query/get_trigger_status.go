package query

import (
	"context"
	"fmt"

	"github.com/bivex/paywall-iap/internal/application/dto"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/google/uuid"
)

// GetTriggerStatusQuery returns the paywall trigger status for a user
type GetTriggerStatusQuery struct {
	triggerService *service.PaywallTriggerService
}

func NewGetTriggerStatusQuery(triggerService *service.PaywallTriggerService) *GetTriggerStatusQuery {
	return &GetTriggerStatusQuery{triggerService: triggerService}
}

func (q *GetTriggerStatusQuery) Execute(ctx context.Context, userID string) (*dto.TriggerStatusResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user ID", domainErrors.ErrInvalidInput)
	}

	status, err := q.triggerService.Evaluate(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate trigger: %w", err)
	}

	return &dto.TriggerStatusResponse{
		ShouldShowPaywall:     status.ShouldShowPaywall,
		ShowD2CButton:         status.ShowD2CButton,
		TriggerReason:         status.TriggerReason,
		SessionCount:          status.SessionCount,
		HasActiveSubscription: status.HasActiveSubscription,
		PurchaseChannel:       status.PurchaseChannel,
	}, nil
}
