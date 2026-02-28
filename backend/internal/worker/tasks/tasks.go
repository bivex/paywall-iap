package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/password9090/paywall-iap/internal/infrastructure/logging"
)

// Task names
const (
	TypeUpdateLTV            = "update:ltv"
	TypeComputeAnalytics     = "compute:analytics"
	TypeProcessWebhook      = "process:webhook"
	TypeSendNotification    = "send:notification"
	TypeSyncLago             = "sync:lago"
	TypeExpireGracePeriod    = "expire:grace_period"
)

// RegisterHandlers registers all task handlers with the worker
func RegisterHandlers(worker *asynq.Worker) {
	// Register task handlers
	worker.Register(TypeUpdateLTV, HandleUpdateLTV)
	worker.Register(TypeComputeAnalytics, HandleComputeAnalytics)
	worker.Register(TypeProcessWebhook, HandleProcessWebhook)
	worker.Register(TypeSendNotification, HandleSendNotification)
	worker.Register(TypeSyncLago, HandleSyncLago)
	worker.Register(TypeExpireGracePeriod, HandleExpireGracePeriod)
}

// RegisterScheduledTasks registers all scheduled (cron) tasks
func RegisterScheduledTasks(scheduler *asynq.Scheduler) {
	// Update LTV every hour
	_, err := scheduler.Register("*/60 * * * *", func() (string, []any) {
		return fmt.Sprintf("%s:%s", time.Now().Format("20060102-15:04"), TypeUpdateLTV), nil
	})
	if err != nil {
		logging.Logger.Error("Failed to schedule LTV update", zap.Error(err))
	}

	// Compute daily analytics at midnight
	_, err = scheduler.Register("0 0 * * *", func() (string, []any) {
		return fmt.Sprintf("%s:%s", time.Now().Format("20060102"), TypeComputeAnalytics), nil
	})
	if err != nil {
		logging.Logger.Error("Failed to schedule analytics computation", zap.Error(err))
	}

	// Check expiring grace periods every 5 minutes
	_, err = scheduler.Register("*/5 * * * *", func() (string, []any) {
		return fmt.Sprintf("%s:%s", time.Now().Format("20060102-15:04"), TypeExpireGracePeriod), nil
	})
	if err != nil {
		logging.Logger.Error("Failed to schedule grace period check", zap.Error(err))
	}
}

// HandleUpdateLTV updates user lifetime value
func HandleUpdateLTV(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		UserID string  `json:"user_id"`
		Amount  float64 `json:"amount"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	logging.Logger.Info("Updating LTV",
		zap.String("task_id", t.ResultWriterID()),
		zap.String("user_id", payload.UserID),
		zap.Float64("amount", payload.Amount),
	)

	// TODO: Implement LTV update logic
	// 1. Fetch all transactions for user
	// 2. Sum up successful transactions
	// 3. Update users.ltv and users.ltv_updated_at

	return nil
}

// HandleComputeAnalytics computes daily analytics aggregates
func HandleComputeAnalytics(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Date string `json:"date"` // Format: YYYY-MM-DD
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	logging.Logger.Info("Computing analytics",
		zap.String("task_id", t.ResultWriterID()),
		zap.String("date", payload.Date),
	)

	// TODO: Implement analytics computation
	// 1. Calculate daily revenue
	// 2. Calculate MRR/ARR
	// 3. Calculate churn rate
	// 4. Store in analytics_aggregates table

	return nil
}

// HandleProcessWebhook processes incoming webhook events
func HandleProcessWebhook(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Provider  string `json:"provider"`
		EventType string `json:"event_type"`
		EventID   string `json:"event_id"`
		Payload   []byte `json:"payload"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	logging.Logger.Info("Processing webhook",
		zap.String("task_id", t.ResultWriterID()),
		zap.String("provider", payload.Provider),
		zap.String("event_type", payload.EventType),
	)

	// TODO: Implement webhook processing logic
	// 1. Check for idempotency using webhook_events table
	// 2. Process based on event type:
	//    - Stripe: payment succeeded, subscription.created, etc.
	//    - Apple: renewal info, cancellations
	//    - Google: subscription notifications
	// 3. Update database accordingly

	return nil
}

// HandleSendNotification sends push notifications to users
func HandleSendNotification(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		UserID  string `json:"user_id"`
		Type    string `json:"type"`
		Title   string `json:"title"`
		Body    string `json:"body"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	logging.Logger.Info("Sending notification",
		zap.String("task_id", t.ResultWriterID()),
		zap.String("user_id", payload.UserID),
		zap.String("type", payload.Type),
	)

	// TODO: Implement notification sending
	// 1. Validate user has notification token
	// 2. Send push notification via Firebase Cloud Messaging or Apple Push Service
	// 3. Log delivery status

	return nil
}

// HandleSyncLago syncs subscription data with Lago billing system
func HandleSyncLago(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		SubscriptionID string `json:"subscription_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	logging.Logger.Info("Syncing with Lago",
		zap.String("task_id", t.ResultWriterID()),
		zap.String("subscription_id", payload.SubscriptionID),
	)

	// TODO: Implement Lago sync logic
	// 1. Fetch subscription details
	// 2. Call Lago API to create/update subscription
	// 3. Handle errors with retry logic

	return nil
}

// HandleExpireGracePeriod processes expiring grace periods
func HandleExpireGracePeriod(ctx context.Context, t *asynq.Task) error {
	logging.Logger.Info("Checking expiring grace periods",
		zap.String("task_id", t.ResultWriterID()),
	)

	// TODO: Implement grace period expiration logic
	// 1. Find grace periods that have expired
	// 2. For expired grace periods:
	//    - If user hasn't repaid, cancel subscription
	//    - Send winback offer if eligible
	//    - Send notification to user

	return nil
}
