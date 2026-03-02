package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// AdminHandler handles admin endpoints
type AdminHandler struct {
	subscriptionRepo domainRepo.SubscriptionRepository
	userRepo         domainRepo.UserRepository
	queries          *generated.Queries
	dbPool           *pgxpool.Pool
	redisClient      *redis.Client
	analyticsService *service.AnalyticsService
	auditService     *service.AuditService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	subscriptionRepo domainRepo.SubscriptionRepository,
	userRepo domainRepo.UserRepository,
	queries *generated.Queries,
	dbPool *pgxpool.Pool,
	redisClient *redis.Client,
	analyticsService *service.AnalyticsService,
	auditService *service.AuditService,
) *AdminHandler {
	return &AdminHandler{
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		queries:          queries,
		dbPool:           dbPool,
		redisClient:      redisClient,
		analyticsService: analyticsService,
		auditService:     auditService,
	}
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
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		PlanType  string `json:"plan_type" binding:"required,oneof=monthly annual lifetime"`
		ExpiresAt string `json:"expires_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		response.BadRequest(c, "Invalid expires_at: expected RFC3339 format")
		return
	}

	sub := entity.NewSubscription(
		userID,
		entity.SourceStripe, // admin-granted via Stripe source
		"web",
		req.ProductID,
		entity.PlanType(req.PlanType),
		expiresAt,
	)

	if err := h.subscriptionRepo.Create(c.Request.Context(), sub); err != nil {
		response.InternalError(c, "Failed to grant subscription")
		return
	}

	// Audit log
	adminID, _ := c.Get("admin_id")
	if aid, ok := adminID.(uuid.UUID); ok {
		_ = h.auditService.LogAction(c.Request.Context(), aid, "grant_subscription", "user", &userID, map[string]interface{}{
			"product_id": req.ProductID,
			"plan_type":  req.PlanType,
			"expires_at": req.ExpiresAt,
		})
	}

	c.Status(http.StatusNoContent)

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
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	sub, err := h.subscriptionRepo.GetActiveByUserID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "No active subscription found for user")
		return
	}

	if err := h.subscriptionRepo.Cancel(c.Request.Context(), sub.ID); err != nil {
		response.InternalError(c, "Failed to revoke subscription")
		return
	}

	// Audit log
	adminID, _ := c.Get("admin_id")
	if aid, ok := adminID.(uuid.UUID); ok {
		_ = h.auditService.LogAction(c.Request.Context(), aid, "revoke_subscription", "user", &userID, map[string]interface{}{
			"reason":          req.Reason,
			"subscription_id": sub.ID,
		})
	}

	c.Status(http.StatusNoContent)

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
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	pageNum, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}
	limitNum, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limitNum < 1 || limitNum > 200 {
		limitNum = 50
	}

	offset := (pageNum - 1) * limitNum
	users, err := h.queries.ListUsers(c.Request.Context(), generated.ListUsersParams{
		Limit:  int32(limitNum),
		Offset: int32(offset),
	})
	if err != nil {
		response.InternalError(c, "Failed to list users")
		return
	}

	total, err := h.queries.CountUsers(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to count users")
		return
	}

	response.OK(c, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  pageNum,
			"limit": limitNum,
			"total": total,
		},
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
	ctx := c.Request.Context()

	dbStatus := "ok"
	if err := h.dbPool.Ping(ctx); err != nil {
		dbStatus = "error: " + err.Error()
	}

	redisStatus := "ok"
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		redisStatus = "error: " + err.Error()
	}

	statusCode := http.StatusOK
	if dbStatus != "ok" || redisStatus != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":   "ok",
		"database": dbStatus,
		"redis":    redisStatus,
	})
}

// GetDashboardMetrics returns comprehensive aggregate metrics for the admin dashboard.
// @Summary Admin dashboard metrics
// @Tags admin
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=object}
// @Router /admin/dashboard/metrics [get]
func (h *AdminHandler) GetDashboardMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now()
	monthAgo := now.AddDate(0, -1, 0)

	// Active user count
	activeUsers, err := h.queries.CountUsers(ctx)
	if err != nil {
		response.InternalError(c, "Failed to count users")
		return
	}

	// Active subscription count
	activeSubs, err := h.queries.GetActiveSubscriptionCount(ctx)
	if err != nil {
		response.InternalError(c, "Failed to count subscriptions")
		return
	}

	// Revenue metrics (MRR / ARR)
	revenue, err := h.analyticsService.CalculateRevenueMetrics(ctx, monthAgo, now)
	if err != nil {
		response.InternalError(c, "Failed to calculate revenue")
		return
	}

	// Churn risk (grace-period subscriptions)
	churnRisk, err := h.analyticsService.GetChurnRiskCount(ctx)
	if err != nil {
		response.InternalError(c, "Failed to calculate churn risk")
		return
	}

	// MRR trend — last 6 months
	mrrTrend, err := h.analyticsService.GetMRRTrend(ctx, 6)
	if err != nil {
		response.InternalError(c, "Failed to calculate MRR trend")
		return
	}

	// Subscription status breakdown
	statusCounts, err := h.analyticsService.GetSubscriptionStatusCounts(ctx)
	if err != nil {
		response.InternalError(c, "Failed to get subscription status counts")
		return
	}

	// Recent audit log (last 5 entries)
	auditLog, err := h.analyticsService.GetRecentAuditLog(ctx, 5)
	if err != nil {
		response.InternalError(c, "Failed to get audit log")
		return
	}

	// Webhook health per provider
	webhookHealth, err := h.analyticsService.GetWebhookHealthByProvider(ctx)
	if err != nil {
		response.InternalError(c, "Failed to get webhook health")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active_users":  activeUsers,
		"active_subs":   activeSubs,
		"mrr":           revenue.MRR,
		"arr":           revenue.ARR,
		"churn_risk":    churnRisk,
		"mrr_trend":     mrrTrend,
		"status_counts": statusCounts,
		"audit_log":     auditLog,
		"webhook_health": webhookHealth,
		"last_updated":  now,
	})
}
