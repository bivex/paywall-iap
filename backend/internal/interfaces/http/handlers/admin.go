package handlers

import (
	"encoding/json"
	"fmt"
	"math"
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
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
	"github.com/bivex/paywall-iap/internal/worker/tasks"
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
	asynqClient      *asynq.Client
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
	asynqClient *asynq.Client,
) *AdminHandler {
	return &AdminHandler{
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		queries:          queries,
		dbPool:           dbPool,
		redisClient:      redisClient,
		analyticsService: analyticsService,
		auditService:     auditService,
		asynqClient:      asynqClient,
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

// GetUserProfile returns a full 360° user profile: identity, subscriptions, transactions, audit log, dunning.
func (h *AdminHandler) GetUserProfile(c *gin.Context) {
ctx := c.Request.Context()
userID := c.Param("id")

uid, err := uuid.Parse(userID)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
return
}

// --- Identity ---
type UserInfo struct {
ID             string  `json:"id"`
PlatformUserID string  `json:"platform_user_id"`
DeviceID       *string `json:"device_id"`
Platform       string  `json:"platform"`
AppVersion     string  `json:"app_version"`
Email          string  `json:"email"`
Role           string  `json:"role"`
LTV            float64 `json:"ltv"`
CreatedAt      string  `json:"created_at"`
}
var user UserInfo
var createdAt time.Time
err = h.dbPool.QueryRow(ctx,
`SELECT id, platform_user_id, device_id, platform, app_version, email, role, ltv, created_at
 FROM users WHERE id = $1`, uid,
).Scan(&uid, &user.PlatformUserID, &user.DeviceID, &user.Platform, &user.AppVersion,
&user.Email, &user.Role, &user.LTV, &createdAt)
if err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
return
}
user.ID = uid.String()
user.CreatedAt = createdAt.Format(time.RFC3339)

// --- Subscriptions ---
type SubRow struct {
ID        string `json:"id"`
Status    string `json:"status"`
Source    string `json:"source"`
Platform  string `json:"platform"`
ProductID string `json:"product_id"`
PlanType  string `json:"plan_type"`
ExpiresAt string `json:"expires_at"`
AutoRenew bool   `json:"auto_renew"`
CreatedAt string `json:"created_at"`
}
subRows, err := h.dbPool.Query(ctx,
`SELECT id, status, source, platform, product_id, plan_type, expires_at, auto_renew, created_at
 FROM subscriptions WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 10`, uid)
if err != nil {
response.InternalError(c, "Failed to get subscriptions")
return
}
defer subRows.Close()
subs := make([]SubRow, 0)
for subRows.Next() {
var s SubRow
var sid uuid.UUID
var exp, cat time.Time
if err := subRows.Scan(&sid, &s.Status, &s.Source, &s.Platform, &s.ProductID, &s.PlanType, &exp, &s.AutoRenew, &cat); err != nil {
continue
}
s.ID = sid.String()
s.ExpiresAt = exp.Format(time.RFC3339)
s.CreatedAt = cat.Format(time.RFC3339)
subs = append(subs, s)
}

// --- Transactions ---
type TxRow struct {
ID       string  `json:"id"`
Amount   float64 `json:"amount"`
Currency string  `json:"currency"`
Status   string  `json:"status"`
TxID     *string `json:"provider_tx_id"`
Date     string  `json:"date"`
}
txRows, err := h.dbPool.Query(ctx,
`SELECT id, amount, currency, status, provider_tx_id, created_at
 FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 20`, uid)
if err != nil {
response.InternalError(c, "Failed to get transactions")
return
}
defer txRows.Close()
txs := make([]TxRow, 0)
for txRows.Next() {
var t TxRow
var tid uuid.UUID
var tdate time.Time
var amountStr string
if err := txRows.Scan(&tid, &amountStr, &t.Currency, &t.Status, &t.TxID, &tdate); err != nil {
continue
}
t.ID = tid.String()
fmt.Sscan(amountStr, &t.Amount)
t.Date = tdate.Format(time.RFC3339)
txs = append(txs, t)
}

// --- Audit log for this user ---
type AuditRow struct {
Action     string `json:"action"`
AdminEmail string `json:"admin_email"`
Detail     string `json:"detail"`
Date       string `json:"date"`
}
auditRows, err := h.dbPool.Query(ctx,
`SELECT a.action, COALESCE(u2.email, a.admin_id::text), COALESCE(a.details::text,'{}'), a.created_at
 FROM admin_audit_log a LEFT JOIN users u2 ON u2.id = a.admin_id
 WHERE a.target_user_id = $1 ORDER BY a.created_at DESC LIMIT 10`, uid)
if err != nil {
response.InternalError(c, "Failed to get audit log")
return
}
defer auditRows.Close()
audits := make([]AuditRow, 0)
for auditRows.Next() {
var a AuditRow
var adate time.Time
if err := auditRows.Scan(&a.Action, &a.AdminEmail, &a.Detail, &adate); err != nil {
continue
}
a.Date = adate.Format(time.RFC3339)
audits = append(audits, a)
}

// --- Dunning ---
type DunningRow struct {
Status      string  `json:"status"`
AttemptCount int    `json:"attempt_count"`
MaxAttempts  int    `json:"max_attempts"`
NextAttempt  *string `json:"next_attempt_at"`
CreatedAt    string  `json:"created_at"`
}
dunRows, err := h.dbPool.Query(ctx,
`SELECT status, attempt_count, max_attempts, next_attempt_at, created_at
 FROM dunning WHERE user_id = $1 ORDER BY created_at DESC LIMIT 5`, uid)
if err != nil {
response.InternalError(c, "Failed to get dunning")
return
}
defer dunRows.Close()
dunnings := make([]DunningRow, 0)
for dunRows.Next() {
var d DunningRow
var ddate time.Time
var nextAt *time.Time
if err := dunRows.Scan(&d.Status, &d.AttemptCount, &d.MaxAttempts, &nextAt, &ddate); err != nil {
continue
}
if nextAt != nil {
s := nextAt.Format(time.RFC3339)
d.NextAttempt = &s
}
d.CreatedAt = ddate.Format(time.RFC3339)
dunnings = append(dunnings, d)
}

c.JSON(http.StatusOK, gin.H{
"user":         user,
"subscriptions": subs,
"transactions": txs,
"audit_log":    audits,
"dunning":      dunnings,
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

	// ── Current MRR (active + grace) ─────────────────────────────────
	var mrr float64
	err := h.dbPool.QueryRow(ctx, `
		SELECT COALESCE(ROUND(SUM(
			CASE WHEN plan_type='monthly' THEN 9.99
			     WHEN plan_type='annual'  THEN 99.99/12.0
			     ELSE 0 END
		)::numeric, 2), 0)
		FROM subscriptions
		WHERE status IN ('active','grace') AND deleted_at IS NULL`).Scan(&mrr)
	if err != nil {
		response.InternalError(c, "mrr query failed")
		return
	}

	// ── LTV: total revenue / distinct users ──────────────────────────
	var ltv float64
	var totalRevenue float64
	err = h.dbPool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0),
		       COALESCE(ROUND(SUM(amount)/NULLIF(COUNT(DISTINCT user_id),0),2),0)
		FROM transactions WHERE status='success'`).Scan(&totalRevenue, &ltv)
	if err != nil {
		response.InternalError(c, "ltv query failed")
		return
	}

	// ── New subs this month ──────────────────────────────────────────
	var newSubsMonth int
	err = h.dbPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM subscriptions
		WHERE deleted_at IS NULL
		  AND date_trunc('month', created_at) = date_trunc('month', now())`).Scan(&newSubsMonth)
	if err != nil {
		response.InternalError(c, "new_subs query failed")
		return
	}

	// ── Churn rate this month ────────────────────────────────────────
	var churned int
	var activePlusChurned int
	err = h.dbPool.QueryRow(ctx, `
		SELECT
		  COUNT(*) FILTER (WHERE status IN ('cancelled','expired')
		    AND date_trunc('month', updated_at) = date_trunc('month', now())),
		  COUNT(*) FILTER (WHERE status IN ('active','grace','cancelled','expired'))
		FROM subscriptions WHERE deleted_at IS NULL`).Scan(&churned, &activePlusChurned)
	if err != nil {
		response.InternalError(c, "churn query failed")
		return
	}
	churnRate := 0.0
	if activePlusChurned > 0 {
		churnRate = math.Round(float64(churned)/float64(activePlusChurned)*1000) / 10 // 1 decimal
	}

	// ── MRR trend: last 6 months ─────────────────────────────────────
	type TrendPoint struct {
		Month       string  `json:"month"`
		MRR         float64 `json:"mrr"`
		ActiveCount int     `json:"active_count"`
		NewSubs     int     `json:"new_subs"`
	}
	trendRows, err := h.dbPool.Query(ctx, `
		WITH months AS (
			SELECT generate_series(
				date_trunc('month', now()) - 5 * interval '1 month',
				date_trunc('month', now()),
				interval '1 month'
			) AS month_start
		),
		monthly_subs AS (
			SELECT m.month_start,
				COUNT(s.id) AS active_count,
				COALESCE(ROUND(SUM(CASE
					WHEN s.plan_type='monthly' THEN 9.99
					WHEN s.plan_type='annual'  THEN 99.99/12.0
					ELSE 0 END)::numeric,2),0) AS mrr
			FROM months m
			LEFT JOIN subscriptions s ON s.deleted_at IS NULL
				AND s.created_at < m.month_start + interval '1 month'
				AND (s.expires_at >= m.month_start OR s.status IN ('active','grace'))
			GROUP BY m.month_start
		),
		monthly_new AS (
			SELECT date_trunc('month', created_at) AS ms, COUNT(*) AS new_subs
			FROM subscriptions WHERE deleted_at IS NULL GROUP BY 1
		)
		SELECT to_char(ms.month_start,'YYYY-MM'), ms.mrr, ms.active_count, COALESCE(mn.new_subs,0)
		FROM monthly_subs ms
		LEFT JOIN monthly_new mn ON mn.ms = ms.month_start
		ORDER BY ms.month_start`)
	if err != nil {
		response.InternalError(c, "trend query failed")
		return
	}
	defer trendRows.Close()
	var trend []TrendPoint
	for trendRows.Next() {
		var p TrendPoint
		if err := trendRows.Scan(&p.Month, &p.MRR, &p.ActiveCount, &p.NewSubs); err != nil {
			response.InternalError(c, "trend scan failed")
			return
		}
		trend = append(trend, p)
	}

	// ── By platform (active + grace) ─────────────────────────────────
	type PlatformRow struct {
		Platform string  `json:"platform"`
		Count    int     `json:"count"`
		MRR      float64 `json:"mrr"`
	}
	platformRows, err := h.dbPool.Query(ctx, `
		SELECT platform, COUNT(*),
			ROUND(SUM(CASE WHEN plan_type='monthly' THEN 9.99
			               WHEN plan_type='annual'  THEN 99.99/12.0 ELSE 0 END)::numeric,2)
		FROM subscriptions
		WHERE status IN ('active','grace') AND deleted_at IS NULL
		GROUP BY platform ORDER BY platform`)
	if err != nil {
		response.InternalError(c, "platform query failed")
		return
	}
	defer platformRows.Close()
	var byPlatform []PlatformRow
	for platformRows.Next() {
		var p PlatformRow
		if err := platformRows.Scan(&p.Platform, &p.Count, &p.MRR); err != nil {
			response.InternalError(c, "platform scan failed")
			return
		}
		byPlatform = append(byPlatform, p)
	}

	// ── By plan type (active + grace) ────────────────────────────────
	type PlanRow struct {
		PlanType string  `json:"plan_type"`
		Count    int     `json:"count"`
		MRR      float64 `json:"mrr"`
	}
	planRows, err := h.dbPool.Query(ctx, `
		SELECT plan_type, COUNT(*),
			ROUND(SUM(CASE WHEN plan_type='monthly' THEN 9.99
			               WHEN plan_type='annual'  THEN 99.99/12.0 ELSE 0 END)::numeric,2)
		FROM subscriptions
		WHERE status IN ('active','grace') AND deleted_at IS NULL
		GROUP BY plan_type ORDER BY plan_type`)
	if err != nil {
		response.InternalError(c, "plan query failed")
		return
	}
	defer planRows.Close()
	var byPlan []PlanRow
	for planRows.Next() {
		var p PlanRow
		if err := planRows.Scan(&p.PlanType, &p.Count, &p.MRR); err != nil {
			response.InternalError(c, "plan scan failed")
			return
		}
		byPlan = append(byPlan, p)
	}

	// ── Status counts ─────────────────────────────────────────────────
	var statusActive, statusGrace, statusCancelled, statusExpired int
	err = h.dbPool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status='active'),
			COUNT(*) FILTER (WHERE status='grace'),
			COUNT(*) FILTER (WHERE status='cancelled'),
			COUNT(*) FILTER (WHERE status='expired')
		FROM subscriptions WHERE deleted_at IS NULL`).Scan(
		&statusActive, &statusGrace, &statusCancelled, &statusExpired)
	if err != nil {
		response.InternalError(c, "status query failed")
		return
	}

	c.JSON(200, gin.H{
		"mrr":             mrr,
		"arr":             math.Round(mrr*12*100) / 100,
		"ltv":             ltv,
		"total_revenue":   totalRevenue,
		"churn_rate":      churnRate,
		"new_subs_month":  newSubsMonth,
		"trend":           trend,
		"by_platform":     byPlatform,
		"by_plan":         byPlan,
		"status_counts": gin.H{
			"active":    statusActive,
			"grace":     statusGrace,
			"cancelled": statusCancelled,
			"expired":   statusExpired,
		},
	})
}

// GetRevenueOps returns dunning queue, recent webhook events, and matomo staging stats.
func (h *AdminHandler) GetRevenueOps(c *gin.Context) {
ctx := c.Request.Context()

// Pagination params for webhook events
whPage := 1
whPageSize := 20
if p := c.Query("wh_page"); p != "" {
	if v, err := strconv.Atoi(p); err == nil && v > 0 {
		whPage = v
	}
}
if ps := c.Query("wh_page_size"); ps != "" {
	if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
		whPageSize = v
	}
}
whOffset := (whPage - 1) * whPageSize

// --- Dunning queue (active / in_progress only) ---
type DunningQueueRow struct {
ID           string  `json:"id"`
Email        string  `json:"email"`
UserID       string  `json:"user_id"`
PlanType     string  `json:"plan_type"`
Status       string  `json:"status"`
AttemptCount int     `json:"attempt_count"`
MaxAttempts  int     `json:"max_attempts"`
NextAttempt  *string `json:"next_attempt_at"`
LastAttempt  *string `json:"last_attempt_at"`
CreatedAt    string  `json:"created_at"`
}
dunningRows, err := h.dbPool.Query(ctx, `
SELECT d.id, u.email, d.user_id, s.plan_type,
       d.status, d.attempt_count, d.max_attempts,
       d.next_attempt_at, d.last_attempt_at, d.created_at
FROM dunning d
JOIN users u ON u.id = d.user_id
JOIN subscriptions s ON s.id = d.subscription_id
WHERE d.status IN ('pending','in_progress')
ORDER BY d.next_attempt_at ASC
LIMIT 50
`)
dunning := make([]DunningQueueRow, 0)
if err == nil {
defer dunningRows.Close()
for dunningRows.Next() {
var row DunningQueueRow
var nextAt, lastAt *time.Time
var createdAt time.Time
var id, uid string
if scanErr := dunningRows.Scan(&id, &row.Email, &uid, &row.PlanType,
&row.Status, &row.AttemptCount, &row.MaxAttempts,
&nextAt, &lastAt, &createdAt); scanErr != nil {
continue
}
row.ID = id
row.UserID = uid
row.CreatedAt = createdAt.Format(time.RFC3339)
if nextAt != nil {
s := nextAt.Format(time.RFC3339)
row.NextAttempt = &s
}
if lastAt != nil {
s := lastAt.Format(time.RFC3339)
row.LastAttempt = &s
}
dunning = append(dunning, row)
}
}

// --- Dunning counts by status ---
type DunningStats struct {
Pending    int `json:"pending"`
InProgress int `json:"in_progress"`
Recovered  int `json:"recovered"`
Failed     int `json:"failed"`
}
var dStats DunningStats
statsRows, _ := h.dbPool.Query(ctx, `
SELECT status, COUNT(*) FROM dunning GROUP BY status
`)
if statsRows != nil {
defer statsRows.Close()
for statsRows.Next() {
var st string
var cnt int
if err2 := statsRows.Scan(&st, &cnt); err2 != nil {
continue
}
switch st {
case "pending":
dStats.Pending = cnt
case "in_progress":
dStats.InProgress = cnt
case "recovered":
dStats.Recovered = cnt
case "failed":
dStats.Failed = cnt
}
}
}

// --- Recent webhook events ---
type WebhookRow struct {
ID          string  `json:"id"`
Provider    string  `json:"provider"`
EventType   string  `json:"event_type"`
EventID     string  `json:"event_id"`
Processed   bool    `json:"processed"`
ProcessedAt *string `json:"processed_at"`
CreatedAt   string  `json:"created_at"`
}
whRows, err2 := h.dbPool.Query(ctx, `
SELECT id, provider, event_type, event_id, processed_at, created_at
FROM webhook_events
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`, whPageSize, whOffset)
webhooks := make([]WebhookRow, 0)
if err2 == nil {
defer whRows.Close()
for whRows.Next() {
var row WebhookRow
var processedAt *time.Time
var createdAt time.Time
if scanErr := whRows.Scan(&row.ID, &row.Provider, &row.EventType, &row.EventID, &processedAt, &createdAt); scanErr != nil {
continue
}
row.CreatedAt = createdAt.Format(time.RFC3339)
if processedAt != nil {
s := processedAt.Format(time.RFC3339)
row.ProcessedAt = &s
row.Processed = true
}
webhooks = append(webhooks, row)
}
}

// Webhook summary counts
var whTotal, whUnprocessed int
_ = h.dbPool.QueryRow(ctx, `SELECT COUNT(*) FROM webhook_events`).Scan(&whTotal)
_ = h.dbPool.QueryRow(ctx, `SELECT COUNT(*) FROM webhook_events WHERE processed_at IS NULL`).Scan(&whUnprocessed)

type WebhookProviderStat struct {
Provider  string `json:"provider"`
Total     int    `json:"total"`
Processed int    `json:"processed"`
}
provRows, _ := h.dbPool.Query(ctx, `
SELECT provider,
       COUNT(*) as total,
       COUNT(processed_at) as processed
FROM webhook_events
GROUP BY provider
ORDER BY total DESC
`)
provStats := make([]WebhookProviderStat, 0)
if provRows != nil {
defer provRows.Close()
for provRows.Next() {
var p WebhookProviderStat
if err3 := provRows.Scan(&p.Provider, &p.Total, &p.Processed); err3 == nil {
provStats = append(provStats, p)
}
}
}

// --- Matomo staging stats ---
type MatomoStats struct {
Pending    int `json:"pending"`
Processing int `json:"processing"`
Sent       int `json:"sent"`
Failed     int `json:"failed"`
Total      int `json:"total"`
}
var mStats MatomoStats
mRows, _ := h.dbPool.Query(ctx, `SELECT status, COUNT(*) FROM matomo_staged_events GROUP BY status`)
if mRows != nil {
defer mRows.Close()
for mRows.Next() {
var st string
var cnt int
if scanErr := mRows.Scan(&st, &cnt); scanErr != nil {
continue
}
switch st {
case "pending":
mStats.Pending = cnt
case "processing":
mStats.Processing = cnt
case "sent":
mStats.Sent = cnt
case "failed":
mStats.Failed = cnt
}
}
}
mStats.Total = mStats.Pending + mStats.Processing + mStats.Sent + mStats.Failed

	whTotalPages := (whTotal + whPageSize - 1) / whPageSize
	if whTotalPages < 1 {
		whTotalPages = 1
	}

	c.JSON(200, gin.H{
"dunning": gin.H{
"queue":  dunning,
"stats":  dStats,
},
"webhooks": gin.H{
"events":       webhooks,
"total":        whTotal,
"unprocessed":  whUnprocessed,
"by_provider":  provStats,
"page":         whPage,
"page_size":    whPageSize,
"total_pages":  whTotalPages,
},
"matomo": gin.H{
"stats": mStats,
},
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

// ListTransactions returns a paginated, filterable list of all transactions for reconciliation.
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
TotalCount     int64   `json:"total_count"`
SuccessCount   int64   `json:"success_count"`
FailedCount    int64   `json:"failed_count"`
RefundedCount  int64   `json:"refunded_count"`
TotalRevenue   float64 `json:"total_revenue"`
TotalRefunded  float64 `json:"total_refunded"`
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
s.source, s.platform, s.plan_type
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
ID          string  `json:"id"`
Amount      float64 `json:"amount"`
Currency    string  `json:"currency"`
Status      string  `json:"status"`
ProviderTxID string `json:"provider_tx_id"`
ReceiptHash string  `json:"receipt_hash"`
CreatedAt   string  `json:"created_at"`
UserID      string  `json:"user_id"`
Email       string  `json:"email"`
Source      string  `json:"source"`
Platform    string  `json:"platform"`
PlanType    string  `json:"plan_type"`
}

result := make([]TxRow, 0, limit)
for rows.Next() {
var r TxRow
var txID, userID uuid.UUID
var createdAt time.Time
if err := rows.Scan(
&txID, &r.Amount, &r.Currency, &r.Status,
&r.ProviderTxID, &r.ReceiptHash, &createdAt,
&userID, &r.Email,
&r.Source, &r.Platform, &r.PlanType,
); err != nil {
response.InternalError(c, "Failed to scan transaction row")
return
}
r.ID = txID.String()
r.UserID = userID.String()
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
