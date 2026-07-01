package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// WinbackHandler handles winback offer endpoints
type WinbackHandler struct {
	acceptWinbackCmd *command.AcceptWinbackOfferCommand
	winbackService   *service.WinbackService
	jwtMiddleware    *middleware.JWTMiddleware
}

// NewWinbackHandler creates a new winback handler
func NewWinbackHandler(
	acceptWinbackCmd *command.AcceptWinbackOfferCommand,
	winbackService *service.WinbackService,
	jwtMiddleware *middleware.JWTMiddleware,
) *WinbackHandler {
	return &WinbackHandler{
		acceptWinbackCmd: acceptWinbackCmd,
		winbackService:   winbackService,
		jwtMiddleware:    jwtMiddleware,
	}
}

// WinbackOfferResponse is the DTO for a single offer
type WinbackOfferResponse struct {
	ID            string  `json:"id"`
	CampaignID    string  `json:"campaign_id"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
	ExpiresAt     string  `json:"expires_at"`
}

// GetActiveOffers returns all active winback offers for the user
func (h *WinbackHandler) GetActiveOffers(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	offers, err := h.winbackService.GetActiveWinbackOffers(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to get winback offers")
		return
	}

	result := make([]WinbackOfferResponse, 0, len(offers))
	for _, o := range offers {
		result = append(result, WinbackOfferResponse{
			ID:            o.ID.String(),
			CampaignID:    o.CampaignID,
			DiscountType:  string(o.DiscountType),
			DiscountValue: o.DiscountValue,
			ExpiresAt:     o.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	response.OK(c, result)
}

// AcceptOffer accepts a winback offer
// @Summary Accept winback offer
// @Tags winback
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body command.AcceptWinbackOfferRequest true "Accept offer request"
// @Success 200 {object} response.SuccessResponse{data=command.AcceptWinbackOfferResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /winback/offers/accept [post]
func (h *WinbackHandler) AcceptOffer(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req command.AcceptWinbackOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// Override user ID from JWT for security
	req.UserID = userID

	resp, err := h.acceptWinbackCmd.Execute(c.Request.Context(), &req)
	if err != nil {
		response.UnprocessableEntity(c, err.Error())
		return
	}

	response.OK(c, resp)
}
