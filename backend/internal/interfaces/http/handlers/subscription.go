package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/password9090/paywall-iap/internal/application/command"
	"github.com/password9090/paywall-iap/internal/application/middleware"
	"github.com/password9090/paywall-iap/internal/application/query"
	"github.com/password9090/paywall-iap/internal/interfaces/http/response"
)

// SubscriptionHandler handles subscription endpoints
type SubscriptionHandler struct {
	getSubQuery         *query.GetSubscriptionQuery
	checkAccessQuery    *query.CheckAccessQuery
	cancelCmd           *command.CancelSubscriptionCommand
	jwtMiddleware       *middleware.JWTMiddleware
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(
	getSubQuery *query.GetSubscriptionQuery,
	checkAccessQuery *query.CheckAccessQuery,
	cancelCmd *command.CancelSubscriptionCommand,
	jwtMiddleware *middleware.JWTMiddleware,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		getSubQuery:      getSubQuery,
		checkAccessQuery: checkAccessQuery,
		cancelCmd:        cancelCmd,
		jwtMiddleware:    jwtMiddleware,
	}
}

// GetSubscription returns the user's subscription details
// @Summary Get subscription details
// @Tags subscription
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=dto.SubscriptionResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /subscription [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	resp, err := h.getSubQuery.Execute(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "Subscription not found")
		return
	}

	response.OK(c, resp)
}

// CheckAccess checks if user has access to premium content
// @Summary Check access to premium content
// @Tags subscription
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=dto.AccessCheckResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /subscription/access [get]
func (h *SubscriptionHandler) CheckAccess(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	resp, err := h.checkAccessQuery.Execute(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to check access")
		return
	}

	response.OK(c, resp)
}

// CancelSubscription cancels the user's subscription
// @Summary Cancel subscription
// @Tags subscription
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body command.CancelSubscriptionRequest true "Cancel request"
// @Success 204
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /subscription [delete]
func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if err := h.cancelCmd.Execute(c.Request.Context(), userID); err != nil {
		response.NotFound(c, "No active subscription found")
		return
	}

	response.NoContent(c)
}
