package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

// GetAuditLog returns a paginated list of admin audit log entries with optional filters.
// Query params: page (1-based), limit (default 20), action, search, from, to (RFC3339)
func (h *AdminHandler) GetAuditLog(c *gin.Context) {
ctx := c.Request.Context()

page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
if page < 1 {
page = 1
}
limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
if limit < 1 || limit > 100 {
limit = 20
}
offset := (page - 1) * limit

action := c.Query("action")
search := c.Query("search")

var from, to time.Time
if v := c.Query("from"); v != "" {
from, _ = time.Parse(time.RFC3339, v)
}
if v := c.Query("to"); v != "" {
to, _ = time.Parse(time.RFC3339, v)
}

pageResult, err := h.analyticsService.GetAuditLogPaginated(ctx, offset, limit, action, search, from, to)
if err != nil {
response.InternalError(c, "Failed to get audit log")
return
}

totalPages := int((pageResult.TotalCount + int64(limit) - 1) / int64(limit))

c.JSON(http.StatusOK, gin.H{
"rows":        pageResult.Rows,
"total":       pageResult.TotalCount,
"page":        page,
"limit":       limit,
"total_pages": totalPages,
})
}

// SearchUsers returns a filtered, paginated list of users.
// Query: page, limit, search (email/platform_user_id), platform (ios/android/web), role
func (h *AdminHandler) SearchUsers(c *gin.Context) {
ctx := c.Request.Context()

page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
if page < 1 {
page = 1
}
limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
if limit < 1 || limit > 200 {
limit = 20
}
offset := (page - 1) * limit

search := c.Query("search")
platform := c.Query("platform")
role := c.Query("role")

args := []interface{}{}
where := []string{}
idx := 1

if search != "" {
args = append(args, "%"+search+"%")
where = append(where, fmt.Sprintf("(u.email ILIKE $%d OR u.platform_user_id ILIKE $%d)", idx, idx))
idx++
}
if platform != "" {
args = append(args, platform)
where = append(where, fmt.Sprintf("u.platform = $%d", idx))
idx++
}
if role != "" {
args = append(args, role)
where = append(where, fmt.Sprintf("u.role = $%d", idx))
idx++
}

whereSQL := ""
if len(where) > 0 {
whereSQL = "WHERE " + strings.Join(where, " AND ")
}

var total int64
countQ := fmt.Sprintf(`SELECT COUNT(*) FROM users u %s`, whereSQL)
if err := h.dbPool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
response.InternalError(c, "Failed to count users")
return
}

args = append(args, limit, offset)
dataQ := fmt.Sprintf(`
SELECT
u.id, u.platform_user_id, u.platform, u.email, u.role,
u.ltv, u.app_version, u.created_at,
COALESCE(s.status, 'none') AS sub_status,
COALESCE(s.expires_at::text, '') AS sub_expires_at
FROM users u
LEFT JOIN LATERAL (
SELECT status, expires_at FROM subscriptions
WHERE user_id = u.id
ORDER BY created_at DESC
LIMIT 1
) s ON true
%s
ORDER BY u.created_at DESC
LIMIT $%d OFFSET $%d
`, whereSQL, idx, idx+1)

rows, err := h.dbPool.Query(ctx, dataQ, args...)
if err != nil {
response.InternalError(c, "Failed to list users")
return
}
defer rows.Close()

type UserRow struct {
ID             string  `json:"id"`
PlatformUserID string  `json:"platform_user_id"`
Platform       string  `json:"platform"`
Email          string  `json:"email"`
Role           string  `json:"role"`
LTV            float64 `json:"ltv"`
AppVersion     string  `json:"app_version"`
CreatedAt      string  `json:"created_at"`
SubStatus      string  `json:"sub_status"`
SubExpiresAt   string  `json:"sub_expires_at"`
}

var uid uuid.UUID
result := make([]UserRow, 0, limit)
for rows.Next() {
var r UserRow
var createdAt time.Time
if err := rows.Scan(&uid, &r.PlatformUserID, &r.Platform, &r.Email, &r.Role,
&r.LTV, &r.AppVersion, &createdAt, &r.SubStatus, &r.SubExpiresAt); err != nil {
response.InternalError(c, "Failed to scan user")
return
}
r.ID = uid.String()
r.CreatedAt = createdAt.Format(time.RFC3339)
result = append(result, r)
}

totalPages := int((total + int64(limit) - 1) / int64(limit))
c.JSON(http.StatusOK, gin.H{
"users":       result,
"total":       total,
"page":        page,
"limit":       limit,
"total_pages": totalPages,
})
}
