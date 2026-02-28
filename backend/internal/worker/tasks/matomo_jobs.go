package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
)

// Task types for Matomo event processing
const (
	TypeMatomoProcessBatch = "matomo:process_batch"
	TypeMatomoSendEvent    = "matomo:send_event"
	TypeMatomoSendEcommerce = "matomo:send_ecommerce"
)

// MatomoWorker handles Matomo event processing jobs
type MatomoWorker struct {
	forwarder *service.MatomoForwarder
	logger    *zap.Logger
}

// NewMatomoWorker creates a new Matomo worker
func NewMatomoWorker(forwarder *service.MatomoForwarder) *MatomoWorker {
	return &MatomoWorker{
		forwarder: forwarder,
		logger:    logging.Logger,
	}
}

// ProcessBatchPayload represents the payload for batch processing
type ProcessBatchPayload struct {
	BatchSize int `json:"batch_size,omitempty"`
}

// NewProcessBatchTask creates a new task for processing a batch of Matomo events
func NewProcessBatchTask(batchSize int) (*asynq.Task, error) {
	payload, err := json.Marshal(ProcessBatchPayload{
		BatchSize: batchSize,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeMatomoProcessBatch, payload), nil
}

// HandleProcessBatch handles the batch processing job
func (w *MatomoWorker) HandleProcessBatch(ctx context.Context, t *asynq.Task) error {
	var payload ProcessBatchPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("Failed to unmarshal payload", zap.Error(err))
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing Matomo event batch",
		zap.String("task_id", t.ResultWriter().TaskID()),
	)

	processed, succeeded, failed, err := w.forwarder.ProcessBatch(ctx)
	if err != nil {
		w.logger.Error("Batch processing failed",
			zap.Error(err),
			zap.Int("processed", processed),
			zap.Int("failed", failed),
		)
		return err
	}

	w.logger.Info("Batch processing completed",
		zap.Int("processed", processed),
		zap.Int("succeeded", succeeded),
		zap.Int("failed", failed),
	)

	// If there are still pending events, schedule another job
	queueSize, _ := w.forwarder.GetQueueSize(ctx)
	if queueSize > 0 {
		w.logger.Info("Scheduling next batch", zap.Int("remaining", queueSize))

		// Schedule next batch to run immediately
		nextTask, err := NewProcessBatchTask(payload.BatchSize)
		if err != nil {
			w.logger.Error("Failed to create next batch task", zap.Error(err))
			return nil
		}

		// Enqueue for immediate processing
		info, err := asynq.NewClient(asynq.RedisClientOpt{
			Addr: "",
		}).Enqueue(nextTask)
		if err != nil {
			w.logger.Error("Failed to enqueue next batch", zap.Error(err))
		} else {
			w.logger.Debug("Enqueued next batch", zap.String("task_id", info.ID))
		}
	}

	return nil
}

// SendEventPayload represents the payload for sending a single event
type SendEventPayload struct {
	UserID      string                 `json:"user_id"`
	Category    string                 `json:"category"`
	Action      string                 `json:"action"`
	Name        string                 `json:"name,omitempty"`
	Value       float64                `json:"value,omitempty"`
	CustomVars  map[string]string      `json:"custom_vars,omitempty"`
	EventTime   time.Time              `json:"event_time,omitempty"`
}

// NewSendEventTask creates a new task for sending a single event
func NewSendEventTask(payload SendEventPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeMatomoSendEvent, data), nil
}

// HandleSendEvent handles the send event job
func (w *MatomoWorker) HandleSendEvent(ctx context.Context, t *asynq.Task) error {
	var payload SendEventPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("Failed to unmarshal payload", zap.Error(err))
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// This would need a UUID - for now, we'll track without user ID
	// In production, convert string UUID to uuid.UUID
	if err := w.forwarder.TrackEvent(ctx, nil, payload.Category, payload.Action, payload.Name, payload.Value, payload.CustomVars); err != nil {
		w.logger.Error("Failed to track event",
			zap.String("category", payload.Category),
			zap.String("action", payload.Action),
			zap.Error(err),
		)
		return err
	}

	w.logger.Debug("Event tracked",
		zap.String("category", payload.Category),
		zap.String("action", payload.Action),
	)

	return nil
}

// SendEcommercePayload represents the payload for sending an ecommerce event
type SendEcommercePayload struct {
	UserID     string                       `json:"user_id"`
	OrderID    string                       `json:"order_id"`
	Revenue    float64                      `json:"revenue"`
	Items      []EcommerceItemPayload       `json:"items"`
	CustomVars map[string]string            `json:"custom_vars,omitempty"`
	EventTime  time.Time                    `json:"event_time,omitempty"`
}

// EcommerceItemPayload represents an item in an ecommerce event
type EcommerceItemPayload struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Category string  `json:"category,omitempty"`
}

// NewSendEcommerceTask creates a new task for sending an ecommerce event
func NewSendEcommerceTask(payload SendEcommercePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeMatomoSendEcommerce, data), nil
}

// HandleSendEcommerce handles the send ecommerce job
func (w *MatomoWorker) HandleSendEcommerce(ctx context.Context, t *asynq.Task) error {
	var payload SendEcommercePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("Failed to unmarshal payload", zap.Error(err))
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Convert items
	items := make([]service.EcommerceItemPayload, len(payload.Items))
	for i, item := range payload.Items {
		items[i] = service.EcommerceItemPayload{
			SKU:      item.SKU,
			Name:     item.Name,
			Price:    item.Price,
			Quantity: item.Quantity,
			Category: item.Category,
		}
	}

	// Track purchase (will be queued for async delivery)
	// Note: Converting string UUID to uuid.UUID would happen here
	if err := w.forwarder.TrackPurchase(ctx, nil, payload.OrderID, payload.Revenue, items, payload.CustomVars); err != nil {
		w.logger.Error("Failed to track purchase",
			zap.String("order_id", payload.OrderID),
			zap.Float64("revenue", payload.Revenue),
			zap.Error(err),
		)
		return err
	}

	w.logger.Debug("Ecommerce event tracked",
		zap.String("order_id", payload.OrderID),
		zap.Float64("revenue", payload.Revenue),
	)

	return nil
}

// RegisterMatomoHandlers registers all Matomo task handlers with the Asynq mux
func RegisterMatomoHandlers(mux *asynq.ServeMux, worker *MatomoWorker) {
	mux.HandleFunc(TypeMatomoProcessBatch, worker.HandleProcessBatch)
	mux.HandleFunc(TypeMatomoSendEvent, worker.HandleSendEvent)
	mux.HandleFunc(TypeMatomoSendEcommerce, worker.HandleSendEcommerce)
}

// SchedulePeriodicBatchProcessing schedules periodic batch processing jobs
func SchedulePeriodicBatchProcessing(scheduler *asynq.Scheduler, interval time.Duration, batchSize int) error {
	task, err := NewProcessBatchTask(batchSize)
	if err != nil {
		return err
	}

	// Schedule periodic task - every 5 minutes
	cron := fmt.Sprintf("@every %s", interval.String())

	entryID, err := scheduler.Register(
		cron,
		task,
	)

	if err != nil {
		return fmt.Errorf("failed to register periodic task: %w", err)
	}

	logging.Logger.Info("Scheduled periodic Matomo batch processing",
		zap.String("entry_id", entryID),
		zap.Duration("interval", interval),
	)

	return nil
}
