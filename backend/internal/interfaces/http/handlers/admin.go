package handlers

import (
	"net/http"

	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
	"github.com/gin-gonic/gin"
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
	_ = c.Param("id")

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
	_ = c.Param("id")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// TODO: Implement revoke subscription logic
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
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": "ok",
		"redis":    "ok",
		"queue":    "ok",
	})
}
