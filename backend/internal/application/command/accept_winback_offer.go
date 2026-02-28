package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// AcceptWinbackOfferCommand accepts a winback offer
type AcceptWinbackOfferCommand struct {
	winbackService *service.WinbackService
}

// AcceptWinbackOfferRequest is the request DTO
type AcceptWinbackOfferRequest struct {
	UserID  string `json:"user_id" validate:"required,uuid"`
	OfferID string `json:"offer_id" validate:"required,uuid"`
}

// AcceptWinbackOfferResponse is the response DTO
type AcceptWinbackOfferResponse struct {
	OfferID       string  `json:"offer_id"`
	CampaignID    string  `json:"campaign_id"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
	FinalAmount   float64 `json:"final_amount,omitempty"`
	Message       string  `json:"message"`
}

// NewAcceptWinbackOfferCommand creates a new command handler
func NewAcceptWinbackOfferCommand(winbackService *service.WinbackService) *AcceptWinbackOfferCommand {
	return &AcceptWinbackOfferCommand{
		winbackService: winbackService,
	}
}

// Execute accepts a winback offer
func (c *AcceptWinbackOfferCommand) Execute(ctx context.Context, req *AcceptWinbackOfferRequest) (*AcceptWinbackOfferResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	offerID, err := uuid.Parse(req.OfferID)
	if err != nil {
		return nil, err
	}

	offer, err := c.winbackService.AcceptWinbackOffer(ctx, userID, offerID)
	if err != nil {
		return nil, err
	}

	resp := &AcceptWinbackOfferResponse{
		OfferID:       offer.ID.String(),
		CampaignID:    offer.CampaignID,
		DiscountType:  string(offer.DiscountType),
		DiscountValue: offer.DiscountValue,
		Message:       "Winback offer accepted successfully",
	}

	return resp, nil
}
