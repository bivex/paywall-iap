package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// WinbackHandler handles winback offer endpoints
type WinbackHandler struct {
	acceptWinbackCmd *command.AcceptWinbackOfferCommand
	jwtMiddleware    *middleware.JWTMiddleware
}

// NewWinbackHandler creates a new winback handler
func NewWinbackHandler(
	acceptWinbackCmd *command.AcceptWinbackOfferCommand,
	jwtMiddleware *middleware.JWTMiddleware,
) *WinbackHandler {
	return &WinbackHandler{
		acceptWinbackCmd: acceptWinbackCmd,
		jwtMiddleware:    jwtMiddleware,
	}
}

// GetActiveOffers returns all active winback offers for the user
// @Summary Get active winback offers
// @Tags winback
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=[]interface{}}
// @Failure 401 {object} response.ErrorResponse
// @Router /winback/offers [get]
func (h *WinbackHandler) GetActiveOffers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// TODO: Implement query to get active offers
	// For now, return empty list
	response.OK(c, []interface{}{})
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
