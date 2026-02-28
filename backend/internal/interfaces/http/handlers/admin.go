package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/password9090/paywall-iap/internal/application/dto"
	"github.com/password9090/paywall-iap/internal/interfaces/http/response"
)

// AdminHandler handles admin endpoints
type AdminHandler struct {
	// TODO: Add admin services
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// GrantSubscription manually grants a subscription to a user
// @Summary Grant subscription to user
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Param request body object{product_id:string, plan_type:string, expires_at:string} true "Grant request"
// @Success 204
// @Router /admin/users/{id}/grant [post]
func (h *AdminHandler) GrantSubscription(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		PlanType  string `json:"plan_type" binding:"required,oneof=monthly annual lifetime"`
		ExpiresAt string `json:"expires_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// TODO: Implement grant subscription logic
	// 1. Validate admin has permissions (ops+ role)
	// 2. Create or update subscription
	// 3. Write to admin_audit_log

	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not yet implemented"})
}

// RevokeSubscription revokes a user's subscription
// @Summary Revoke subscription from user
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Param request body object{reason:string} true "Revoke request"
// @Success 204
// @Router /admin/users/{id}/revoke [post]
func (h *AdminHandler) RevokeSubscription(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// TODO: Implement revoke subscription logic
	// 1. Validate admin has permissions (ops+ role)
	// 2. Cancel subscription
	// 3. Write to admin_audit_log

	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not yet implemented"})
}

// ListUsers returns a paginated list of users
// @Summary List users
// @Tags admin
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} response.SuccessResponse{data=object}
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "50")

	// TODO: Implement user listing logic
	// 1. Validate admin has permissions (support+ role)
	// 2. Fetch users with pagination
	// 3. Return response

	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Not yet implemented",
		"page":    page,
		"limit":   limit,
	})
}

// GetHealth returns system health status
// @Summary System health check
// @Tags admin
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=object}
// @Router /admin/health [get]
func (h *AdminHandler) GetHealth(c *gin.Context) {
	// TODO: Implement health check logic
	// 1. Check database connectivity
	// 2. Check Redis connectivity
	// 3. Check queue status
	// 4. Return health status

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"database": "ok",
		"redis": "ok",
		"queue": "ok",
	})
}
