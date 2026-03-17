package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/application/query"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// PaywallHandler handles paywall trigger, email capture, and session tracking
type PaywallHandler struct {
	getTriggerStatusQuery *query.GetTriggerStatusQuery
	captureEmailCmd       *command.CaptureEmailCommand
	trackSessionCmd       *command.TrackSessionCommand
	jwtMiddleware         *middleware.JWTMiddleware
}

func NewPaywallHandler(
	getTriggerStatusQuery *query.GetTriggerStatusQuery,
	captureEmailCmd *command.CaptureEmailCommand,
	trackSessionCmd *command.TrackSessionCommand,
	jwtMiddleware *middleware.JWTMiddleware,
) *PaywallHandler {
	return &PaywallHandler{
		getTriggerStatusQuery: getTriggerStatusQuery,
		captureEmailCmd:       captureEmailCmd,
		trackSessionCmd:       trackSessionCmd,
		jwtMiddleware:         jwtMiddleware,
	}
}

// GetTriggerStatus returns whether to show paywall and D2C button for the authenticated user
// @Summary Get paywall trigger status
// @Tags paywall
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=dto.TriggerStatusResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /user/trigger-status [get]
func (h *PaywallHandler) GetTriggerStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	resp, err := h.getTriggerStatusQuery.Execute(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to get trigger status")
		return
	}

	response.OK(c, resp)
}

// CaptureEmail saves the user's email for D2C purchase flow
// @Summary Capture user email
// @Tags paywall
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body dto.CaptureEmailRequest true "Email"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /user/email [post]
func (h *PaywallHandler) CaptureEmail(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req dto.CaptureEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid email")
		return
	}

	if err := h.captureEmailCmd.Execute(c.Request.Context(), userID, &req); err != nil {
		response.InternalError(c, "Failed to save email")
		return
	}

	response.NoContent(c)
}

// TrackSession increments the session counter for the authenticated user
// @Summary Track a new session
// @Tags paywall
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=dto.TrackSessionResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /user/session [post]
func (h *PaywallHandler) TrackSession(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	count, err := h.trackSessionCmd.Execute(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to track session")
		return
	}

	response.OK(c, dto.TrackSessionResponse{SessionCount: count})
}
