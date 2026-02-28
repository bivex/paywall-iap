package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// IAPHandler handles IAP verification endpoints
type IAPHandler struct {
	verifyIAPCmd     *command.VerifyIAPCommand
	jwtMiddleware   *middleware.JWTMiddleware
	rateLimiter     *middleware.RateLimiter
}

// NewIAPHandler creates a new IAP handler
func NewIAPHandler(
	verifyIAPCmd *command.VerifyIAPCommand,
	jwtMiddleware *middleware.JWTMiddleware,
	rateLimiter *middleware.RateLimiter,
) *IAPHandler {
	return &IAPHandler{
		verifyIAPCmd:   verifyIAPCmd,
		jwtMiddleware: jwtMiddleware,
		rateLimiter:   rateLimiter,
	}
}

// VerifyReceipt handles IAP receipt verification
// @Summary Verify IAP receipt
// @Tags iap
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.VerifyIAPRequest true "IAP verification request"
// @Success 200 {object} response.SuccessResponse{data=dto.VerifyIAPResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /verify/iap [post]
func (h *IAPHandler) VerifyReceipt(c *gin.Context) {
	// Get user ID from JWT context
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req dto.VerifyIAPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	resp, err := h.verifyIAPCmd.Execute(c.Request.Context(), userID, &req)
	if err != nil {
		// Handle specific errors
		if err.Error() == "receipt already processed" {
			// Return the existing subscription (idempotency)
			response.OK(c, resp)
			return
		}
		response.UnprocessableEntity(c, err.Error())
		return
	}

	response.OK(c, resp)
}
