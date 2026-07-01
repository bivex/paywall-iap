package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
)

const (
	TypeProcessDunningAttempt = "dunning:process_attempt"
	TypeCheckPendingDunning   = "dunning:check_pending"
)

// ProcessDunningAttemptPayload is the payload for processing a dunning retry
type ProcessDunningAttemptPayload struct {
	DunningID      uuid.UUID `json:"dunning_id"`
	PaymentSuccess bool      `json:"payment_success"`
}

// DunningJobHandler handles dunning-related background jobs
type DunningJobHandler struct {
	dunningService *service.DunningService
	asynqClient    *asynq.Client
}

// NewDunningJobHandler creates a new dunning job handler
func NewDunningJobHandler(dunningService *service.DunningService, asynqClient *asynq.Client) *DunningJobHandler {
	return &DunningJobHandler{
		dunningService: dunningService,
		asynqClient:    asynqClient,
	}
}

// HandleProcessDunningAttempt processes a single dunning retry attempt.
// For IAP (App Store/Play Store) subscriptions, dunning is mostly handled by the stores.
// payment_success is set by the webhook that triggers this task (e.g. a renewal event).
func (h *DunningJobHandler) HandleProcessDunningAttempt(ctx context.Context, t *asynq.Task) error {
	var p ProcessDunningAttemptPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	if err := h.dunningService.ProcessDunningAttempt(ctx, p.DunningID, p.PaymentSuccess); err != nil {
		return err
	}

	return nil
}

// HandleCheckPendingDunning checks for dunning processes that need a retry
func (h *DunningJobHandler) HandleCheckPendingDunning(ctx context.Context, t *asynq.Task) error {
	pending, err := h.dunningService.GetPendingDunningAttempts(ctx, 100)
	if err != nil {
		return err
	}

	for _, dunning := range pending {
		payload, err := json.Marshal(ProcessDunningAttemptPayload{
			DunningID:      dunning.ID,
			PaymentSuccess: false, // store-side payment outcome unknown at check time; webhook updates this
		})
		if err != nil {
			logging.Logger.Error("Failed to marshal dunning payload",
				zap.String("dunning_id", dunning.ID.String()),
				zap.Error(err),
			)
			continue
		}
		task := asynq.NewTask(TypeProcessDunningAttempt, payload)
		if _, err := h.asynqClient.Enqueue(task); err != nil {
			logging.Logger.Error("Failed to enqueue dunning attempt task",
				zap.String("dunning_id", dunning.ID.String()),
				zap.Error(err),
			)
		}
	}

	logging.Logger.Info("Checked pending dunning", zap.Int("count", len(pending)))
	return nil
}

// ScheduleDunningJobs schedules recurring dunning jobs
func ScheduleDunningJobs(scheduler *asynq.Scheduler) error {
	// Check for pending dunning every hour
	_, err := scheduler.Register("0 * * * *", asynq.NewTask(TypeCheckPendingDunning, nil))
	if err != nil {
		return fmt.Errorf("failed to register dunning check job: %w", err)
	}
	return nil
}
