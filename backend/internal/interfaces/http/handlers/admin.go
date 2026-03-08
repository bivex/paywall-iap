package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/domain/service"
	persistenceRepo "github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
	"github.com/bivex/paywall-iap/internal/worker/tasks"
)

// AdminHandler handles admin endpoints
type AdminHandler struct {
	subscriptionRepo       domainRepo.SubscriptionRepository
	userRepo               domainRepo.UserRepository
	queries                *generated.Queries
	dbPool                 *pgxpool.Pool
	redisClient            *redis.Client
	analyticsService       *service.AnalyticsService
	auditService           *service.AuditService
	revenueOpsService      *service.RevenueOpsService
	analyticsReportService *service.AnalyticsReportService
	userProfileService     *service.UserProfileService
	winbackService         *service.WinbackService
	experimentAdminService *service.ExperimentAdminService
	asynqClient            *asynq.Client
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
	revenueOpsService *service.RevenueOpsService,
	analyticsReportService *service.AnalyticsReportService,
	userProfileService *service.UserProfileService,
	winbackService *service.WinbackService,
	asynqClient *asynq.Client,
) *AdminHandler {
	var experimentAdminService *service.ExperimentAdminService
	if dbPool != nil {
		experimentAdminService = service.NewExperimentAdminService(persistenceRepo.NewExperimentAdminRepository(dbPool))
	}

	return &AdminHandler{
		subscriptionRepo:       subscriptionRepo,
		userRepo:               userRepo,
		queries:                queries,
		dbPool:                 dbPool,
		redisClient:            redisClient,
		analyticsService:       analyticsService,
		auditService:           auditService,
		revenueOpsService:      revenueOpsService,
		analyticsReportService: analyticsReportService,
		userProfileService:     userProfileService,
		winbackService:         winbackService,
		experimentAdminService: experimentAdminService,
		asynqClient:            asynqClient,
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
		"active_users":   activeUsers,
		"active_subs":    activeSubs,
		"mrr":            revenue.MRR,
		"arr":            revenue.ARR,
		"churn_risk":     churnRisk,
		"mrr_trend":      mrrTrend,
		"status_counts":  statusCounts,
		"audit_log":      auditLog,
		"webhook_health": webhookHealth,
		"last_updated":   now,
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

// ForceCancel hard-cancels a user's active subscription immediately.
// Body: {"reason": "..."}
func (h *AdminHandler) ForceCancel(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Reason == "" {
		req.Reason = "admin_force_cancel"
	}

	sub, err := h.subscriptionRepo.GetActiveByUserID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "No active subscription found")
		return
	}
	if err := h.subscriptionRepo.Cancel(c.Request.Context(), sub.ID); err != nil {
		response.InternalError(c, "Failed to cancel subscription")
		return
	}
	adminID, _ := c.Get("admin_id")
	if aid, ok := adminID.(uuid.UUID); ok {
		_ = h.auditService.LogAction(c.Request.Context(), aid, "revoke_subscription", "user", &userID, map[string]interface{}{
			"reason": req.Reason, "subscription_id": sub.ID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ForceRenew extends the active subscription's expires_at by the given days (default 30).
// Body: {"days": 30, "reason": "..."}
func (h *AdminHandler) ForceRenew(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	var req struct {
		Days   int    `json:"days"`
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Days <= 0 {
		req.Days = 30
	}
	if req.Reason == "" {
		req.Reason = "admin_force_renew"
	}

	sub, err := h.subscriptionRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		// No active sub → try to find latest expired and reactivate
		var subID uuid.UUID
		var expiresAt time.Time
		err2 := h.dbPool.QueryRow(ctx,
			`SELECT id, expires_at FROM subscriptions WHERE user_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`, userID,
		).Scan(&subID, &expiresAt)
		if err2 != nil {
			response.NotFound(c, "No subscription found")
			return
		}
		newExpires := time.Now().UTC().AddDate(0, 0, req.Days)
		_, err3 := h.dbPool.Exec(ctx,
			`UPDATE subscriptions SET status='active', expires_at=$1, updated_at=now() WHERE id=$2`,
			newExpires, subID)
		if err3 != nil {
			response.InternalError(c, "Failed to renew subscription")
			return
		}
		adminID, _ := c.Get("admin_id")
		if aid, ok := adminID.(uuid.UUID); ok {
			_ = h.auditService.LogAction(ctx, aid, "manual_renewal", "subscription", &userID, map[string]interface{}{
				"reason": req.Reason, "days": req.Days, "sub_id": subID,
			})
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "new_expires_at": newExpires.Format(time.RFC3339)})
		return
	}

	// Active sub: extend expires_at
	newExpires := sub.ExpiresAt.AddDate(0, 0, req.Days)
	_, err = h.dbPool.Exec(ctx,
		`UPDATE subscriptions SET expires_at=$1, updated_at=now() WHERE id=$2`,
		newExpires, sub.ID)
	if err != nil {
		response.InternalError(c, "Failed to extend subscription")
		return
	}
	adminID, _ := c.Get("admin_id")
	if aid, ok := adminID.(uuid.UUID); ok {
		_ = h.auditService.LogAction(ctx, aid, "manual_renewal", "subscription", &userID, map[string]interface{}{
			"reason": req.Reason, "days": req.Days, "sub_id": sub.ID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "new_expires_at": newExpires.Format(time.RFC3339)})
}

// GrantGracePeriod grants a grace period of given days to a user's subscription.
// Body: {"days": 7, "reason": "..."}
func (h *AdminHandler) GrantGracePeriod(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	var req struct {
		Days   int    `json:"days"`
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Days <= 0 {
		req.Days = 7
	}
	if req.Reason == "" {
		req.Reason = "admin_grant_grace"
	}

	// Get active or most recent subscription
	var subID uuid.UUID
	err = h.dbPool.QueryRow(ctx,
		`SELECT id FROM subscriptions WHERE user_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`, userID,
	).Scan(&subID)
	if err != nil {
		response.NotFound(c, "No subscription found")
		return
	}

	// Upsert: deactivate any existing active grace period first
	_, _ = h.dbPool.Exec(ctx,
		`UPDATE grace_periods SET status='expired', updated_at=now() WHERE user_id=$1 AND status='active'`, userID)

	gracExpires := time.Now().UTC().AddDate(0, 0, req.Days)
	var graceID uuid.UUID
	err = h.dbPool.QueryRow(ctx, `
INSERT INTO grace_periods (user_id, subscription_id, status, expires_at)
VALUES ($1, $2, 'active', $3)
RETURNING id`,
		userID, subID, gracExpires,
	).Scan(&graceID)
	if err != nil {
		response.InternalError(c, "Failed to grant grace period")
		return
	}

	// Set subscription to grace status
	_, _ = h.dbPool.Exec(ctx,
		`UPDATE subscriptions SET status='grace', updated_at=now() WHERE id=$1`, subID)

	adminID, _ := c.Get("admin_id")
	if aid, ok := adminID.(uuid.UUID); ok {
		_ = h.auditService.LogAction(ctx, aid, "grant_subscription", "user", &userID, map[string]interface{}{
			"reason": req.Reason, "days": req.Days, "grace_id": graceID, "type": "grace_period",
		})
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "grace_expires_at": gracExpires.Format(time.RFC3339)})
}

// GetAnalyticsReport returns a full analytics report with real formulas:
// MRR = active subs × price/month (monthly:9.99, annual:99.99/12)
// Churn rate = churned this month / (active + churned) × 100
// LTV = total successful revenue / distinct users with transactions
// New subs = count of subscriptions created this month
func (h *AdminHandler) GetAnalyticsReport(c *gin.Context) {
	ctx := c.Request.Context()

	report, err := h.analyticsReportService.GetReport(ctx)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(200, report)
}

// GetRevenueOps returns dunning queue, recent webhook events, and matomo staging stats.
// GetRevenueOps returns revenue operations dashboard data
// GET /admin/revenue-ops?wh_page=1&wh_page_size=20&wh_pending=1
func (h *AdminHandler) GetRevenueOps(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination params
	whPage := 1
	if p := c.Query("wh_page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			whPage = v
		}
	}
	whPageSize := 20
	if ps := c.Query("wh_page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			whPageSize = v
		}
	}
	whPendingOnly := c.Query("wh_pending") == "1"

	// Use service to fetch report
	report, err := h.revenueOpsService.GetReport(ctx, whPage, whPageSize, whPendingOnly)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch revenue ops data"})
		return
	}

	c.JSON(200, report)
}

// ListWebhooks returns paginated, filterable webhook events.
// GET /admin/webhooks?page=1&limit=20&provider=stripe&status=pending&search=evt_id
func (h *AdminHandler) ListWebhooks(c *gin.Context) {
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

	provider := c.Query("provider")
	status := c.Query("status") // "pending" | "processed" | "failed"
	search := c.Query("search")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	args := []interface{}{}
	where := []string{}
	idx := 1

	if provider != "" {
		args = append(args, strings.ToLower(provider))
		where = append(where, fmt.Sprintf("LOWER(provider) = $%d", idx))
		idx++
	}
	if status == "pending" {
		where = append(where, "processed_at IS NULL")
	} else if status == "processed" {
		where = append(where, "processed_at IS NOT NULL")
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where = append(where, fmt.Sprintf("(event_id ILIKE $%d OR event_type ILIKE $%d)", idx, idx))
		idx++
	}
	if dateFrom != "" {
		args = append(args, dateFrom)
		where = append(where, fmt.Sprintf("created_at >= $%d::date", idx))
		idx++
	}
	if dateTo != "" {
		args = append(args, dateTo)
		where = append(where, fmt.Sprintf("created_at < ($%d::date + INTERVAL '1 day')", idx))
		idx++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	type Summary struct {
		Total     int64 `json:"total"`
		Pending   int64 `json:"pending"`
		Processed int64 `json:"processed"`
	}
	var summary Summary
	sumQ := fmt.Sprintf(`
		SELECT
		  COUNT(*),
		  COUNT(*) FILTER (WHERE processed_at IS NULL),
		  COUNT(*) FILTER (WHERE processed_at IS NOT NULL)
		FROM webhook_events %s`, whereSQL)
	if err := h.dbPool.QueryRow(ctx, sumQ, args...).Scan(&summary.Total, &summary.Pending, &summary.Processed); err != nil {
		response.InternalError(c, "Failed to get webhook summary")
		return
	}

	dataArgs := append(args, limit, offset)
	dataQ := fmt.Sprintf(`
		SELECT id, provider, event_type, COALESCE(event_id,''), processed_at, created_at
		FROM webhook_events
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereSQL, idx, idx+1)

	rows, err := h.dbPool.Query(ctx, dataQ, dataArgs...)
	if err != nil {
		response.InternalError(c, "Failed to list webhooks")
		return
	}
	defer rows.Close()

	type Row struct {
		ID          string  `json:"id"`
		Provider    string  `json:"provider"`
		EventType   string  `json:"event_type"`
		EventID     string  `json:"event_id"`
		Processed   bool    `json:"processed"`
		ProcessedAt *string `json:"processed_at"`
		CreatedAt   string  `json:"created_at"`
	}

	result := make([]Row, 0, limit)
	for rows.Next() {
		var r Row
		var id uuid.UUID
		var processedAt *time.Time
		var createdAt time.Time
		if scanErr := rows.Scan(&id, &r.Provider, &r.EventType, &r.EventID, &processedAt, &createdAt); scanErr != nil {
			continue
		}
		r.ID = id.String()
		r.CreatedAt = createdAt.Format(time.RFC3339)
		if processedAt != nil {
			s := processedAt.Format(time.RFC3339)
			r.ProcessedAt = &s
			r.Processed = true
		}
		result = append(result, r)
	}

	totalPages := int((summary.Total + int64(limit) - 1) / int64(limit))
	c.JSON(200, gin.H{
		"webhooks":    result,
		"summary":     summary,
		"total":       summary.Total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// ReplayWebhook re-enqueues an unprocessed webhook event for processing.
// POST /v1/admin/webhooks/:id/replay
func (h *AdminHandler) ReplayWebhook(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid webhook ID")
		return
	}

	var provider, eventType, eventID string
	var processedAt *time.Time
	err = h.dbPool.QueryRow(ctx,
		`SELECT provider, event_type, event_id, processed_at FROM webhook_events WHERE id = $1`, id,
	).Scan(&provider, &eventType, &eventID, &processedAt)
	if err != nil {
		response.NotFound(c, "Webhook event not found")
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"provider":   provider,
		"event_type": eventType,
		"event_id":   eventID,
	})
	if _, err := h.asynqClient.Enqueue(asynq.NewTask(tasks.TypeProcessWebhook, payload)); err != nil {
		response.InternalError(c, "Failed to enqueue replay task")
		return
	}

	// Log admin action
	adminID, _ := c.Get("admin_id")
	if aid, ok := adminID.(uuid.UUID); ok {
		_ = h.auditService.LogAction(ctx, aid, "replay_webhook", "webhook_event", &id, map[string]interface{}{
			"provider": provider, "event_type": eventType, "event_id": eventID,
		})
	}

	c.JSON(200, gin.H{"ok": true, "queued": eventID})
}

// GetSubscriptionDetail returns full detail for a single subscription by ID.
// GET /admin/subscriptions/:id
func (h *AdminHandler) GetSubscriptionDetail(c *gin.Context) {
	ctx := c.Request.Context()
	subID := c.Param("id")
	if subID == "" {
		c.JSON(400, gin.H{"error": "missing id"})
		return
	}

	type TxRow struct {
		ID           string  `json:"id"`
		Provider     string  `json:"provider"`
		ProviderTxID string  `json:"provider_tx_id"`
		Amount       float64 `json:"amount"`
		Currency     string  `json:"currency"`
		Status       string  `json:"status"`
		CreatedAt    string  `json:"created_at"`
	}

	type Detail struct {
		ID           string  `json:"id"`
		Status       string  `json:"status"`
		Source       string  `json:"source"`
		Platform     string  `json:"platform"`
		PlanType     string  `json:"plan_type"`
		ExpiresAt    string  `json:"expires_at"`
		CreatedAt    string  `json:"created_at"`
		UpdatedAt    string  `json:"updated_at"`
		UserID       string  `json:"user_id"`
		Email        string  `json:"email"`
		LTV          float64 `json:"ltv"`
		Transactions []TxRow `json:"transactions"`
	}

	var d Detail
	var sID, uID uuid.UUID
	var expiresAt, createdAt, updatedAt time.Time
	err := h.dbPool.QueryRow(ctx, `
		SELECT s.id, s.status, s.source, s.platform, s.plan_type,
		       s.expires_at, s.created_at, s.updated_at,
		       u.id, COALESCE(u.email,''), COALESCE(u.ltv,0)
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = $1 AND s.deleted_at IS NULL
	`, subID).Scan(
		&sID, &d.Status, &d.Source, &d.Platform, &d.PlanType,
		&expiresAt, &createdAt, &updatedAt,
		&uID, &d.Email, &d.LTV,
	)
	if err != nil {
		c.JSON(404, gin.H{"error": "subscription not found"})
		return
	}
	d.ID = sID.String()
	d.UserID = uID.String()
	d.ExpiresAt = expiresAt.Format(time.RFC3339)
	d.CreatedAt = createdAt.Format(time.RFC3339)
	d.UpdatedAt = updatedAt.Format(time.RFC3339)

	// Load transactions (no provider column in transactions table; derive from sub source)
	txRows, _ := h.dbPool.Query(ctx, `
		SELECT id, COALESCE(provider_tx_id,''), amount, currency, status, created_at
		FROM transactions
		WHERE subscription_id = $1
		ORDER BY created_at DESC
	`, sID)
	d.Transactions = make([]TxRow, 0)
	if txRows != nil {
		defer txRows.Close()
		for txRows.Next() {
			var tx TxRow
			var txID uuid.UUID
			var txCreatedAt time.Time
			if scanErr := txRows.Scan(&txID, &tx.ProviderTxID, &tx.Amount, &tx.Currency, &tx.Status, &txCreatedAt); scanErr != nil {
				continue
			}
			tx.ID = txID.String()
			tx.Provider = d.Source // derive from subscription source
			tx.CreatedAt = txCreatedAt.Format(time.RFC3339)
			d.Transactions = append(d.Transactions, tx)
		}
	}

	c.JSON(200, d)
}

// ListSubscriptions returns a paginated, filterable list of all subscriptions.
// GET /admin/subscriptions?page=1&limit=20&status=active&source=iap&platform=ios&plan_type=monthly&search=email&date_from=2024-01-01&date_to=2024-12-31
func (h *AdminHandler) ListSubscriptions(c *gin.Context) {
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

	status := c.Query("status")
	source := c.Query("source")
	platform := c.Query("platform")
	planType := c.Query("plan_type")
	search := c.Query("search")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	args := []interface{}{}
	where := []string{"s.deleted_at IS NULL"}
	idx := 1

	if status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("s.status = $%d", idx))
		idx++
	}
	if source != "" {
		args = append(args, source)
		where = append(where, fmt.Sprintf("s.source = $%d", idx))
		idx++
	}
	if platform != "" {
		args = append(args, platform)
		where = append(where, fmt.Sprintf("s.platform = $%d", idx))
		idx++
	}
	if planType != "" {
		args = append(args, planType)
		where = append(where, fmt.Sprintf("s.plan_type = $%d", idx))
		idx++
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where = append(where, fmt.Sprintf("u.email ILIKE $%d", idx))
		idx++
	}
	if dateFrom != "" {
		args = append(args, dateFrom)
		where = append(where, fmt.Sprintf("s.expires_at >= $%d::date", idx))
		idx++
	}
	if dateTo != "" {
		args = append(args, dateTo)
		where = append(where, fmt.Sprintf("s.expires_at < ($%d::date + INTERVAL '1 day')", idx))
		idx++
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	var total int64
	countQ := fmt.Sprintf(
		`SELECT COUNT(*) FROM subscriptions s JOIN users u ON u.id = s.user_id %s`,
		whereSQL,
	)
	if err := h.dbPool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		response.InternalError(c, "Failed to count subscriptions")
		return
	}

	args = append(args, limit, offset)
	dataQ := fmt.Sprintf(`
SELECT
s.id, s.status, s.source, s.platform, s.plan_type,
s.expires_at, s.created_at,
u.id AS user_id, COALESCE(u.email, '') AS email,
COALESCE(u.ltv, 0) AS ltv
FROM subscriptions s
JOIN users u ON u.id = s.user_id
%s
ORDER BY s.created_at DESC
LIMIT $%d OFFSET $%d
`, whereSQL, idx, idx+1)

	rows, err := h.dbPool.Query(ctx, dataQ, args...)
	if err != nil {
		response.InternalError(c, "Failed to list subscriptions")
		return
	}
	defer rows.Close()

	type SubRow struct {
		ID        string  `json:"id"`
		Status    string  `json:"status"`
		Source    string  `json:"source"`
		Platform  string  `json:"platform"`
		PlanType  string  `json:"plan_type"`
		ExpiresAt string  `json:"expires_at"`
		CreatedAt string  `json:"created_at"`
		UserID    string  `json:"user_id"`
		Email     string  `json:"email"`
		LTV       float64 `json:"ltv"`
	}

	var subID, userID uuid.UUID
	result := make([]SubRow, 0, limit)
	for rows.Next() {
		var r SubRow
		var expiresAt, createdAt time.Time
		if err := rows.Scan(
			&subID, &r.Status, &r.Source, &r.Platform, &r.PlanType,
			&expiresAt, &createdAt,
			&userID, &r.Email, &r.LTV,
		); err != nil {
			response.InternalError(c, "Failed to scan subscription row")
			return
		}
		r.ID = subID.String()
		r.UserID = userID.String()
		r.ExpiresAt = expiresAt.Format(time.RFC3339)
		r.CreatedAt = createdAt.Format(time.RFC3339)
		result = append(result, r)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	c.JSON(http.StatusOK, gin.H{
		"subscriptions": result,
		"total":         total,
		"page":          page,
		"limit":         limit,
		"total_pages":   totalPages,
	})
}

// GetTransactionDetail returns full detail for a single transaction: tx data + user + subscription.
// GET /admin/transactions/:id
func (h *AdminHandler) GetTransactionDetail(c *gin.Context) {
	ctx := c.Request.Context()
	txIDStr := c.Param("id")
	if txIDStr == "" {
		c.JSON(400, gin.H{"error": "missing id"})
		return
	}
	txID, err := uuid.Parse(txIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	type UserSnap struct {
		ID        string  `json:"id"`
		Email     string  `json:"email"`
		LTV       float64 `json:"ltv"`
		CreatedAt string  `json:"created_at"`
	}
	type SubSnap struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Source    string `json:"source"`
		Platform  string `json:"platform"`
		PlanType  string `json:"plan_type"`
		ExpiresAt string `json:"expires_at"`
		CreatedAt string `json:"created_at"`
	}
	type Detail struct {
		ID           string   `json:"id"`
		Amount       float64  `json:"amount"`
		Currency     string   `json:"currency"`
		Status       string   `json:"status"`
		ProviderTxID string   `json:"provider_tx_id"`
		ReceiptHash  string   `json:"receipt_hash"`
		CreatedAt    string   `json:"created_at"`
		User         UserSnap `json:"user"`
		Subscription SubSnap  `json:"subscription"`
	}

	var d Detail
	var uID, subID uuid.UUID
	var tCreatedAt, uCreatedAt, sExpiresAt, sCreatedAt time.Time

	err = h.dbPool.QueryRow(ctx, `
		SELECT
		  t.id, t.amount, t.currency, t.status,
		  COALESCE(t.provider_tx_id,''), COALESCE(t.receipt_hash,''), t.created_at,
		  u.id, COALESCE(u.email,''), COALESCE(u.ltv,0), u.created_at,
		  s.id, s.status, s.source, s.platform, s.plan_type, s.expires_at, s.created_at
		FROM transactions t
		JOIN users u ON u.id = t.user_id
		JOIN subscriptions s ON s.id = t.subscription_id
		WHERE t.id = $1
	`, txID).Scan(
		&txID, &d.Amount, &d.Currency, &d.Status,
		&d.ProviderTxID, &d.ReceiptHash, &tCreatedAt,
		&uID, &d.User.Email, &d.User.LTV, &uCreatedAt,
		&subID, &d.Subscription.Status, &d.Subscription.Source,
		&d.Subscription.Platform, &d.Subscription.PlanType,
		&sExpiresAt, &sCreatedAt,
	)
	if err != nil {
		c.JSON(404, gin.H{"error": "transaction not found"})
		return
	}

	d.ID = txID.String()
	d.CreatedAt = tCreatedAt.Format(time.RFC3339)
	d.User.ID = uID.String()
	d.User.CreatedAt = uCreatedAt.Format(time.RFC3339)
	d.Subscription.ID = subID.String()
	d.Subscription.ExpiresAt = sExpiresAt.Format(time.RFC3339)
	d.Subscription.CreatedAt = sCreatedAt.Format(time.RFC3339)

	c.JSON(200, d)
}

// GET /admin/transactions?page=1&limit=20&status=success&source=iap&platform=ios&search=email&date_from=2024-01-01&date_to=2024-12-31
func (h *AdminHandler) ListTransactions(c *gin.Context) {
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

	status := c.Query("status")
	source := c.Query("source")
	platform := c.Query("platform")
	search := c.Query("search")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	args := []interface{}{}
	where := []string{}
	idx := 1

	if status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("t.status = $%d", idx))
		idx++
	}
	if source != "" {
		args = append(args, source)
		where = append(where, fmt.Sprintf("s.source = $%d", idx))
		idx++
	}
	if platform != "" {
		args = append(args, platform)
		where = append(where, fmt.Sprintf("s.platform = $%d", idx))
		idx++
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where = append(where, fmt.Sprintf("u.email ILIKE $%d", idx))
		idx++
	}
	if dateFrom != "" {
		args = append(args, dateFrom)
		where = append(where, fmt.Sprintf("t.created_at >= $%d::date", idx))
		idx++
	}
	if dateTo != "" {
		args = append(args, dateTo)
		where = append(where, fmt.Sprintf("t.created_at < ($%d::date + INTERVAL '1 day')", idx))
		idx++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	baseQ := fmt.Sprintf(`
FROM transactions t
JOIN users u ON u.id = t.user_id
JOIN subscriptions s ON s.id = t.subscription_id
%s`, whereSQL)

	// totals for reconciliation summary
	type Summary struct {
		TotalCount    int64   `json:"total_count"`
		SuccessCount  int64   `json:"success_count"`
		FailedCount   int64   `json:"failed_count"`
		RefundedCount int64   `json:"refunded_count"`
		TotalRevenue  float64 `json:"total_revenue"`
		TotalRefunded float64 `json:"total_refunded"`
	}
	var summary Summary
	sumQ := fmt.Sprintf(`
SELECT
COUNT(*),
COUNT(*) FILTER (WHERE t.status = 'success'),
COUNT(*) FILTER (WHERE t.status = 'failed'),
COUNT(*) FILTER (WHERE t.status = 'refunded'),
COALESCE(SUM(t.amount) FILTER (WHERE t.status = 'success'), 0),
COALESCE(SUM(t.amount) FILTER (WHERE t.status = 'refunded'), 0)
%s`, baseQ)
	if err := h.dbPool.QueryRow(ctx, sumQ, args...).Scan(
		&summary.TotalCount, &summary.SuccessCount, &summary.FailedCount,
		&summary.RefundedCount, &summary.TotalRevenue, &summary.TotalRefunded,
	); err != nil {
		response.InternalError(c, "Failed to get transaction summary")
		return
	}

	args = append(args, limit, offset)
	dataQ := fmt.Sprintf(`
SELECT
t.id, t.amount, t.currency, t.status,
COALESCE(t.provider_tx_id, '') AS provider_tx_id,
COALESCE(t.receipt_hash, '') AS receipt_hash,
t.created_at,
u.id AS user_id, COALESCE(u.email, '') AS email,
s.source, s.platform, s.plan_type,
s.id AS subscription_id
%s
ORDER BY t.created_at DESC
LIMIT $%d OFFSET $%d`, baseQ, idx, idx+1)

	rows, err := h.dbPool.Query(ctx, dataQ, args...)
	if err != nil {
		response.InternalError(c, "Failed to list transactions")
		return
	}
	defer rows.Close()

	type TxRow struct {
		ID             string  `json:"id"`
		Amount         float64 `json:"amount"`
		Currency       string  `json:"currency"`
		Status         string  `json:"status"`
		ProviderTxID   string  `json:"provider_tx_id"`
		ReceiptHash    string  `json:"receipt_hash"`
		CreatedAt      string  `json:"created_at"`
		UserID         string  `json:"user_id"`
		Email          string  `json:"email"`
		Source         string  `json:"source"`
		Platform       string  `json:"platform"`
		PlanType       string  `json:"plan_type"`
		SubscriptionID string  `json:"subscription_id"`
	}

	result := make([]TxRow, 0, limit)
	for rows.Next() {
		var r TxRow
		var txID, userID, subScanID uuid.UUID
		var createdAt time.Time
		if err := rows.Scan(
			&txID, &r.Amount, &r.Currency, &r.Status,
			&r.ProviderTxID, &r.ReceiptHash, &createdAt,
			&userID, &r.Email,
			&r.Source, &r.Platform, &r.PlanType, &subScanID,
		); err != nil {
			response.InternalError(c, "Failed to scan transaction row")
			return
		}
		r.ID = txID.String()
		r.UserID = userID.String()
		r.SubscriptionID = subScanID.String()
		r.CreatedAt = createdAt.Format(time.RFC3339)
		result = append(result, r)
	}

	totalPages := int((summary.TotalCount + int64(limit) - 1) / int64(limit))
	c.JSON(http.StatusOK, gin.H{
		"transactions": result,
		"summary":      summary,
		"total":        summary.TotalCount,
		"page":         page,
		"limit":        limit,
		"total_pages":  totalPages,
	})
}

// GetUserProfile returns a full 360° user profile: identity, subscriptions, transactions, audit log, dunning.
func (h *AdminHandler) GetUserProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	profile, err := h.userProfileService.GetProfile(ctx, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}
