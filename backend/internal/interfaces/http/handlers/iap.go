package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/application/dto"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
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

	// Enforce max body size: 64 KB to prevent oversized receipts
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 65536)

	var req dto.VerifyIAPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if err.Error() == "http: request body too large" {
			response.BadRequest(c, "receipt_data exceeds maximum allowed size (64 KB)")
			return
		}
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	resp, err := h.verifyIAPCmd.Execute(c.Request.Context(), userID, &req)
	if err != nil {
		switch {
		case isValidationError(err):
			response.BadRequest(c, err.Error())
		case isDuplicateReceiptError(err):
			response.OK(c, resp)
		default:
			response.UnprocessableEntity(c, err.Error())
		}
		return
	}

	response.OK(c, resp)
}

func isValidationError(err error) bool {
msg := err.Error()
return strings.HasPrefix(msg, "validation failed") ||
strings.HasPrefix(msg, "invalid input") ||
strings.Contains(msg, domainErrors.ErrInvalidReceipt.Error()) ||
strings.Contains(msg, domainErrors.ErrInvalidInput.Error())
}

func isDuplicateReceiptError(err error) bool {
return strings.Contains(err.Error(), "receipt already processed") ||
strings.Contains(err.Error(), domainErrors.ErrDuplicateReceipt.Error())
}
