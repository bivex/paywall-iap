package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

const (
	TypeProcessExpiredGracePeriods = "grace_period:process_expired"
	TypeNotifyExpiringGracePeriods = "grace_period:notify_expiring"
)

// ProcessExpiredGracePeriodsPayload is the payload for processing expired grace periods
type ProcessExpiredGracePeriodsPayload struct {
	Limit int `json:"limit"`
}

// NotifyExpiringGracePeriodsPayload is the payload for notifying expiring grace periods
type NotifyExpiringGracePeriodsPayload struct {
	WithinHours int `json:"within_hours"`
}

// GracePeriodJobHandler handles grace period background jobs
type GracePeriodJobHandler struct {
	gracePeriodService  *service.GracePeriodService
	notificationService *service.NotificationService
}

// NewGracePeriodJobHandler creates a new grace period job handler
func NewGracePeriodJobHandler(
	gracePeriodService *service.GracePeriodService,
	notificationService *service.NotificationService,
) *GracePeriodJobHandler {
	return &GracePeriodJobHandler{
		gracePeriodService:  gracePeriodService,
		notificationService: notificationService,
	}
}

// HandleProcessExpiredGracePeriods processes expired grace periods
func (h *GracePeriodJobHandler) HandleProcessExpiredGracePeriods(ctx context.Context, t *asynq.Task) error {
	var p ProcessExpiredGracePeriodsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	limit := p.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	processed, err := h.gracePeriodService.ProcessExpiredGracePeriods(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to process expired grace periods: %w", err)
	}

	fmt.Printf("Processed %d expired grace periods\n", processed)
	return nil
}

// HandleNotifyExpiringGracePeriods sends notifications for expiring grace periods
func (h *GracePeriodJobHandler) HandleNotifyExpiringGracePeriods(ctx context.Context, t *asynq.Task) error {
	var p NotifyExpiringGracePeriodsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	withinHours := p.WithinHours
	if withinHours == 0 {
		withinHours = 24 // Default 24 hours
	}

	expiringPeriods, err := h.gracePeriodService.NotifyExpiringSoon(ctx, withinHours)
	if err != nil {
		return fmt.Errorf("failed to get expiring grace periods: %w", err)
	}

	for _, gp := range expiringPeriods {
		// Send notification to user
		err := h.notificationService.SendGracePeriodExpiringNotification(ctx, gp.UserID, gp)
		if err != nil {
			// Log error but continue with other notifications
			fmt.Printf("Failed to send notification for grace period %s: %v\n", gp.ID, err)
		}
	}

	fmt.Printf("Sent %d grace period expiry notifications\n", len(expiringPeriods))
	return nil
}

// ScheduleGracePeriodJobs schedules recurring grace period jobs
func ScheduleGracePeriodJobs(scheduler *asynq.Scheduler) error {
	// Process expired grace periods every hour
	_, err := scheduler.Register("0 * * * *", asynq.NewTask(TypeProcessExpiredGracePeriods,
		mustMarshalJSON(ProcessExpiredGracePeriodsPayload{Limit: 100})))
	if err != nil {
		return err
	}

	// Notify expiring grace periods every 6 hours
	_, err = scheduler.Register("0 */6 * * *", asynq.NewTask(TypeNotifyExpiringGracePeriods,
		mustMarshalJSON(NotifyExpiringGracePeriodsPayload{WithinHours: 24})))
	if err != nil {
		return err
	}

	return nil
}

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
