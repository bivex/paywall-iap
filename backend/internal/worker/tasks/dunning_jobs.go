package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

const (
	TypeProcessDunningAttempt = "dunning:process_attempt"
	TypeCheckPendingDunning   = "dunning:check_pending"
)

// ProcessDunningAttemptPayload is the payload for processing a dunning retry
type ProcessDunningAttemptPayload struct {
	DunningID uuid.UUID `json:"dunning_id"`
}

// DunningJobHandler handles dunning-related background jobs
type DunningJobHandler struct {
	dunningService *service.DunningService
}

// NewDunningJobHandler creates a new dunning job handler
func NewDunningJobHandler(dunningService *service.DunningService) *DunningJobHandler {
	return &DunningJobHandler{
		dunningService: dunningService,
	}
}

// HandleProcessDunningAttempt processes a single dunning retry attempt
func (h *DunningJobHandler) HandleProcessDunningAttempt(ctx context.Context, t *asynq.Task) error {
	var p ProcessDunningAttemptPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	// In production, this would call an external payment gateway or IAP service
	// to check if renewal succeeded or try to charge again.
	// For this simulation, we'll assume it's a retry logic.
	// In reality, for IAP (App Store/Play Store), dunning is mostly handled by the stores,
	// but we might want to track it and notify users if the store tells us it's in dunning.

	// For simulation, let's say we have a 20% recovery rate on each retry
	paymentSuccess := false
	// Simulate some logic here...

	err := h.dunningService.ProcessDunningAttempt(ctx, p.DunningID, paymentSuccess)
	if err != nil {
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

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	defer client.Close()

	for _, dunning := range pending {
		payload, _ := json.Marshal(ProcessDunningAttemptPayload{DunningID: dunning.ID})
		task := asynq.NewTask(TypeProcessDunningAttempt, payload)
		_, err := client.Enqueue(task)
		if err != nil {
			fmt.Printf("Failed to enqueue dunning attempt task for %s: %v\n", dunning.ID, err)
		}
	}

	fmt.Printf("Enqueued %d dunning attempt tasks\n", len(pending))
	return nil
}

// ScheduleDunningJobs schedules recurring dunning jobs
func ScheduleDunningJobs(scheduler *asynq.Scheduler) error {
	// Check for pending dunning every hour
	_, err := scheduler.Register("0 * * * *", asynq.NewTask(TypeCheckPendingDunning, nil))
	if err != nil {
		return err
	}

	return nil
}
