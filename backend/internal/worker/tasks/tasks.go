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

	// Mark the webhook_event as processed once we handle it
	eventUUID, err := uuid.Parse(payload.EventID)
	_ = eventUUID
	_ = err

	h.logger.Info("Processing webhook event",
		zap.String("provider", payload.Provider),
		zap.String("event_type", payload.EventType),
		zap.String("event_id", payload.EventID),
	)

	// Dispatch based on provider and event type
	// Each case should update subscriptions, trigger notifications, etc.
	switch payload.Provider {
	case "stripe":
		h.logger.Info("Stripe event dispatched", zap.String("type", payload.EventType))
		// TODO per event type: invoice.payment_succeeded → renew sub
		//                       customer.subscription.deleted → cancel sub
	case "apple":
		h.logger.Info("Apple event dispatched", zap.String("type", payload.EventType))
		// TODO per event type: DID_RENEW → extend expiry
		//                       EXPIRED → mark expired
	case "google":
		h.logger.Info("Google event dispatched", zap.String("type", payload.EventType))
		// TODO per event type: notificationType 2 → renewed, 3 → cancelled
	}

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
