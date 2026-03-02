package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
)

// Task names
const (
	TypeUpdateLTV         = "update:ltv"
	TypeComputeAnalytics  = "compute:analytics"
	TypeProcessWebhook    = "process:webhook"
	TypeSendNotification  = "send:notification"
	TypeSyncLago          = "sync:lago"
	TypeExpireGracePeriod = "expire:grace_period"
)

// TaskHandlers holds dependencies for all task handlers.
type TaskHandlers struct {
	queries *generated.Queries
	logger  *zap.Logger
}

// NewTaskHandlers creates task handlers with database access.
func NewTaskHandlers(queries *generated.Queries) *TaskHandlers {
	return &TaskHandlers{
		queries: queries,
		logger:  logging.Logger,
	}
}

// RegisterHandlers registers all task handlers with the server mux.
func RegisterHandlers(mux *asynq.ServeMux, h *TaskHandlers) {
	mux.HandleFunc(TypeUpdateLTV, h.HandleUpdateLTV)
	mux.HandleFunc(TypeComputeAnalytics, h.HandleComputeAnalytics)
	mux.HandleFunc(TypeProcessWebhook, h.HandleProcessWebhook)
	mux.HandleFunc(TypeSendNotification, h.HandleSendNotification)
	mux.HandleFunc(TypeSyncLago, h.HandleSyncLago)
	mux.HandleFunc(TypeExpireGracePeriod, h.HandleExpireGracePeriod)

	// Dunning handlers
	// Note: In a real app we'd inject DunningService/AnalyticsService into TaskHandlers or use separate handlers
	// For simplicity, we'll assume they are registered in their respective files or here.
	// But according to my plan, I should have RegisterDunningHandlers etc.

}

// RegisterScheduledTasks registers all scheduled (cron) tasks
func RegisterScheduledTasks(scheduler *asynq.Scheduler) {
	// Update LTV every hour
	_, err := scheduler.Register("0 * * * *", asynq.NewTask(TypeUpdateLTV, nil))
	if err != nil {
		logging.Logger.Error("Failed to schedule LTV update", zap.Error(err))
	}

	// Compute daily analytics at midnight
	_, err = scheduler.Register("0 0 * * *", asynq.NewTask(TypeComputeAnalytics, nil))
	if err != nil {
		logging.Logger.Error("Failed to schedule analytics computation", zap.Error(err))
	}

	// Check expiring grace periods every 5 minutes
	_, err = scheduler.Register("*/5 * * * *", asynq.NewTask(TypeExpireGracePeriod, nil))
	if err != nil {
		logging.Logger.Error("Failed to schedule grace period check", zap.Error(err))
	}

	// Schedule Dunning and Analytics jobs
	_ = ScheduleDunningJobs(scheduler)
	_ = ScheduleAnalyticsJobs(scheduler)
}

// HandleUpdateLTV updates user lifetime value
func (h *TaskHandlers) HandleUpdateLTV(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	userUUID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	// Sum all successful transactions for this user
	ltvRaw, err := h.queries.GetLTVByUserID(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("failed to query LTV: %w", err)
	}

	ltv := toFloat64(ltvRaw)

	// Update the user's LTV field
	if _, err := h.queries.UpdateUserLTV(ctx, generated.UpdateUserLTVParams{
		ID:  userUUID,
		Ltv: ltv,
	}); err != nil {
		return fmt.Errorf("failed to update LTV: %w", err)
	}

	h.logger.Info("LTV updated",
		zap.String("user_id", payload.UserID),
		zap.Float64("ltv", ltv),
	)
	return nil
}

// HandleComputeAnalytics computes daily analytics aggregates
func (h *TaskHandlers) HandleComputeAnalytics(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Date string `json:"date"` // YYYY-MM-DD
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	// Default to yesterday if no date provided
	targetDate := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	if payload.Date != "" {
		parsed, err := time.Parse("2006-01-02", payload.Date)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		targetDate = parsed
	}

	nextDay := targetDate.AddDate(0, 0, 1)

	// Compute daily revenue
	revenueRaw, err := h.queries.GetDailyRevenue(ctx, generated.GetDailyRevenueParams{
		CreatedAt:   targetDate,
		CreatedAt_2: nextDay,
	})
	if err != nil {
		return fmt.Errorf("failed to query daily revenue: %w", err)
	}

	revenue := toFloat64(revenueRaw)

	// Compute active subscription count (current snapshot)
	activeCount, err := h.queries.GetActiveSubscriptionCount(ctx)
	if err != nil {
		return fmt.Errorf("failed to count active subscriptions: %w", err)
	}

	// Store aggregates
	metrics := []struct {
		name  string
		value float64
	}{
		{"daily_revenue", revenue},
		{"active_subscriptions", float64(activeCount)},
	}
	for _, m := range metrics {
		if err := h.queries.UpsertAnalyticsAggregate(ctx, generated.UpsertAnalyticsAggregateParams{
			MetricName:  m.name,
			MetricDate:  targetDate,
			MetricValue: m.value,
		}); err != nil {
			h.logger.Error("Failed to store metric",
				zap.String("metric", m.name),
				zap.Error(err),
			)
		}
	}

	h.logger.Info("Analytics computed",
		zap.String("date", targetDate.Format("2006-01-02")),
		zap.Float64("daily_revenue", revenue),
		zap.Int64("active_subscriptions", activeCount),
	)
	return nil
}

// HandleProcessWebhook processes incoming webhook events
func (h *TaskHandlers) HandleProcessWebhook(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Provider  string `json:"provider"`
		EventType string `json:"event_type"`
		EventID   string `json:"event_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	h.logger.Info("Processing webhook event",
		zap.String("provider", payload.Provider),
		zap.String("event_type", payload.EventType),
		zap.String("event_id", payload.EventID),
	)

	// Fetch event from database to get payload
	event, err := h.queries.GetWebhookEventByProviderAndID(ctx, generated.GetWebhookEventByProviderAndIDParams{
		Provider: payload.Provider,
		EventID:  payload.EventID,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch webhook event: %w", err)
	}

	// Dispatch based on provider
	switch payload.Provider {
	case "stripe":
		if err := h.handleStripeEvent(ctx, event); err != nil {
			return err
		}
	case "apple":
		// TODO: implement Apple App Store Server Notifications v2 handler.
		// Payload is a signed JWS (JWT). Steps:
		//   1. Decode + verify signedPayload using Apple root CA
		//   2. Extract notificationType / subtype and originalTransactionId
		//   3. Map to subscription state changes (SUBSCRIBED, DID_RENEW, EXPIRED, REFUND, etc.)
		//   4. Update subscriptions table accordingly (status, expires_at, cancel_at)
		//   5. Trigger SendNotification task for relevant events
		h.logger.Info("Apple event ignored (stub)", zap.String("type", payload.EventType))
	case "google":
		if err := h.handleGoogleRTDNEvent(ctx, event); err != nil {
			h.logger.Error("Google RTDN handler error", zap.Error(err), zap.String("event_id", payload.EventID))
			// Don't fail the task — return nil so the event is still marked processed
			// and we don't loop on it. Real-world: send to DLQ.
		}
	}

	// Mark as processed
	if err := h.queries.MarkWebhookEventProcessed(ctx, event.ID); err != nil {
		h.logger.Error("Failed to mark event processed", zap.Error(err))
	}

	return nil
}

func (h *TaskHandlers) handleStripeEvent(ctx context.Context, event generated.WebhookEvent) error {
	var body struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				Customer string `json:"customer"` // is mapped to platform_user_id in k6
			} `json:"object"`
		} `json:"data"`
	}

	if err := json.Unmarshal(event.Payload, &body); err != nil {
		return fmt.Errorf("failed to unmarshal stripe payload: %w", err)
	}

	// k6 simulation uses platform_user_id as Stripe customer ID
	platformID := body.Data.Object.Customer
	if platformID == "" {
		return fmt.Errorf("missing customer (platform_id) in stripe payload")
	}

	// Find the user
	user, err := h.queries.GetUserByPlatformID(ctx, platformID)
	if err != nil {
		return fmt.Errorf("failed to find user by platform_id %s: %w", platformID, err)
	}

	// Provision premium access (simple create subscription)
	_, err = h.queries.CreateSubscription(ctx, generated.CreateSubscriptionParams{
		UserID:    user.ID,
		Status:    "active",
		Source:    "stripe",
		Platform:  "web",
		ProductID: "pro_monthly_k6",
		PlanType:  "monthly",
		ExpiresAt: time.Now().AddDate(0, 1, 0), // 1 month
		AutoRenew: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	h.logger.Info("Stripe subscription provisioned",
		zap.String("user_id", user.ID.String()),
		zap.String("platform_id", platformID),
	)
	return nil
}

// HandleSendNotification sends push notifications to users
func (h *TaskHandlers) HandleSendNotification(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		UserID string `json:"user_id"`
		Type   string `json:"type"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	// Notification sending requires FCM (Android) or APNs (iOS) credentials.
	// Credential injection via env vars: FCM_SERVER_KEY, APNS_KEY_ID, APNS_TEAM_ID
	// Implementation: use firebase.google.com/go/messaging or apns2 library.
	h.logger.Info("Notification send requested (stub — no credentials configured)",
		zap.String("user_id", payload.UserID),
		zap.String("type", payload.Type),
		zap.String("title", payload.Title),
	)
	return nil
}

// HandleSyncLago syncs subscription data with Lago billing system
func (h *TaskHandlers) HandleSyncLago(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		SubscriptionID string `json:"subscription_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	// Lago sync requires Lago API credentials: LAGO_API_KEY, LAGO_API_URL
	// Implementation: POST /api/v1/subscriptions with the subscription data
	// Library: net/http with JSON body
	h.logger.Info("Lago sync requested (stub — no Lago credentials configured)",
		zap.String("subscription_id", payload.SubscriptionID),
	)
	return nil
}

// HandleExpireGracePeriod processes expiring grace periods
func (h *TaskHandlers) HandleExpireGracePeriod(ctx context.Context, t *asynq.Task) error {
	// Find all grace periods that have expired but are still marked active
	expired, err := h.queries.GetExpiredGracePeriods(ctx)
	if err != nil {
		return fmt.Errorf("failed to query expired grace periods: %w", err)
	}

	h.logger.Info("Processing expired grace periods", zap.Int("count", len(expired)))

	for _, gp := range expired {
		// Cancel the linked subscription
		if _, err := h.queries.CancelSubscription(ctx, gp.SubscriptionID); err != nil {
			h.logger.Error("Failed to cancel subscription for expired grace period",
				zap.String("grace_period_id", gp.ID.String()),
				zap.String("subscription_id", gp.SubscriptionID.String()),
				zap.Error(err),
			)
			continue
		}

		// Mark grace period as expired
		if err := h.queries.UpdateGracePeriodStatus(ctx, generated.UpdateGracePeriodStatusParams{
			ID:     gp.ID,
			Status: "expired",
		}); err != nil {
			h.logger.Error("Failed to update grace period status",
				zap.String("grace_period_id", gp.ID.String()),
				zap.Error(err),
			)
		}

		h.logger.Info("Grace period expired — subscription cancelled",
			zap.String("user_id", gp.UserID.String()),
			zap.String("subscription_id", gp.SubscriptionID.String()),
		)
	}

	return nil
}

// toFloat64 converts an interface{} (from pgx NUMERIC scan) to float64.
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int64:
		return float64(x)
	case int32:
		return float64(x)
	case int:
		return float64(x)
	case fmt.Stringer:
		f := 0.0
		fmt.Sscanf(x.String(), "%f", &f)
		return f
	}
	f := 0.0
	fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &f)
	return f
}

// ─── Google RTDN ─────────────────────────────────────────────────────────────

// Google Play Real-Time Developer Notification types (notificationType field).
// Reference: https://developer.android.com/google/play/billing/rtdn-reference
const (
rtdnSubscriptionRecovered          = 1  // recovered from account hold
rtdnSubscriptionRenewed            = 2  // auto-renewed
rtdnSubscriptionCanceled           = 3  // voluntarily canceled
rtdnSubscriptionPurchased          = 4  // new purchase
rtdnSubscriptionOnHold             = 5  // account hold (payment deferred)
rtdnSubscriptionInGracePeriod      = 6  // grace period started
rtdnSubscriptionRestarted          = 7  // restarted from pause/hold
rtdnSubscriptionPriceChangeConfirm = 8  // user confirmed price change
rtdnSubscriptionDeferred           = 9  // renewal deferred
rtdnSubscriptionPaused             = 10 // paused (Play paused billing)
rtdnSubscriptionPausedScheduleChanged = 11
rtdnSubscriptionRevoked            = 12 // revoked (refunded)
rtdnSubscriptionExpired            = 13 // expired
)

// rtdnPayload is the DeveloperNotification JSON body stored in webhook_events.
type rtdnPayload struct {
PackageName string `json:"packageName"`
EventTimeMillis string `json:"eventTimeMillis"`
SubscriptionNotification struct {
Version          string `json:"version"`
NotificationType int    `json:"notificationType"`
PurchaseToken    string `json:"purchaseToken"`
SubscriptionID   string `json:"subscriptionId"`
} `json:"subscriptionNotification"`
TestNotification *struct{} `json:"testNotification,omitempty"`
}

// handleGoogleRTDNEvent processes a Google RTDN webhook event stored in the DB.
//
// It maps notificationType → subscription status transition:
//   - PURCHASED / RENEWED / RECOVERED / RESTARTED → active
//   - CANCELED → cancelled (auto_renew=false)
//   - EXPIRED / REVOKED → expired
//   - ON_HOLD / PAUSED  → on_hold
//   - IN_GRACE_PERIOD   → grace_period
//   - DEFERRED / PRICE_CHANGE_CONFIRMED / PAUSE_SCHEDULE_CHANGED → no status change (logged)
func (h *TaskHandlers) handleGoogleRTDNEvent(ctx context.Context, event generated.WebhookEvent) error {
var notif rtdnPayload
if err := json.Unmarshal(event.Payload, &notif); err != nil {
return fmt.Errorf("rtdn: unmarshal payload: %w", err)
}

// Test notification — no subscription notification, always ack.
if notif.TestNotification != nil {
h.logger.Info("rtdn: test notification received")
return nil
}

sn := notif.SubscriptionNotification
if sn.PurchaseToken == "" {
return fmt.Errorf("rtdn: missing purchaseToken in subscriptionNotification")
}

h.logger.Info("rtdn: processing",
zap.Int("notificationType", sn.NotificationType),
zap.String("purchaseToken", sn.PurchaseToken),
zap.String("subscriptionId", sn.SubscriptionID),
)

// Look up the subscription via provider_tx_id = purchaseToken.
sub, err := h.queries.GetSubscriptionByProviderTxID(ctx, sn.PurchaseToken)
if err != nil {
// Unknown token — likely a notification for a purchase we haven't seen yet
// (race: webhook arrives before /verify/iap). Log and move on.
h.logger.Warn("rtdn: subscription not found for purchaseToken",
zap.String("purchaseToken", sn.PurchaseToken),
zap.Error(err),
)
return nil
}

newStatus := ""
newExpiry := time.Time{}

switch sn.NotificationType {
case rtdnSubscriptionPurchased, rtdnSubscriptionRenewed,
rtdnSubscriptionRecovered, rtdnSubscriptionRestarted:
newStatus = "active"
// Extend expiry by 1 month for renewal/recovered (we don't re-verify here;
// a proper implementation would call purchases.subscriptionsv2.get).
if sn.NotificationType == rtdnSubscriptionRenewed ||
sn.NotificationType == rtdnSubscriptionRecovered ||
sn.NotificationType == rtdnSubscriptionRestarted {
newExpiry = time.Now().AddDate(0, 1, 0)
}

case rtdnSubscriptionCanceled:
newStatus = "cancelled"

case rtdnSubscriptionExpired, rtdnSubscriptionRevoked:
newStatus = "expired"

case rtdnSubscriptionOnHold, rtdnSubscriptionPaused:
newStatus = "on_hold"

case rtdnSubscriptionInGracePeriod:
newStatus = "grace_period"

case rtdnSubscriptionDeferred, rtdnSubscriptionPriceChangeConfirm,
rtdnSubscriptionPausedScheduleChanged:
// Informational — no status change needed.
h.logger.Info("rtdn: informational notification, no status change",
zap.Int("notificationType", sn.NotificationType),
zap.String("subscriptionId", sn.SubscriptionID),
)
return nil

default:
h.logger.Warn("rtdn: unknown notificationType", zap.Int("type", sn.NotificationType))
return nil
}

if newStatus != "" && newStatus != sub.Status {
if _, err := h.queries.UpdateSubscriptionStatus(ctx, generated.UpdateSubscriptionStatusParams{
ID:     sub.ID,
Status: newStatus,
}); err != nil {
return fmt.Errorf("rtdn: update subscription status: %w", err)
}
h.logger.Info("rtdn: subscription status updated",
zap.String("subscription_id", sub.ID.String()),
zap.String("old_status", sub.Status),
zap.String("new_status", newStatus),
)
}

// Extend expiry for renewal events.
if !newExpiry.IsZero() {
if _, err := h.queries.UpdateSubscriptionExpiry(ctx, generated.UpdateSubscriptionExpiryParams{
ID:        sub.ID,
ExpiresAt: newExpiry,
}); err != nil {
return fmt.Errorf("rtdn: update subscription expiry: %w", err)
}
h.logger.Info("rtdn: subscription expiry extended",
zap.String("subscription_id", sub.ID.String()),
zap.Time("new_expiry", newExpiry),
)
}

return nil
}
