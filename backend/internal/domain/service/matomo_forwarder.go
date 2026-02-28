package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	matomoClient "github.com/bivex/paywall-iap/internal/infrastructure/external/matomo"
)

// EcommerceItemPayload is an alias for matomo client's EcommerceItem
type EcommerceItemPayload = matomoClient.EcommerceItem


// MatomoForwarder handles event forwarding to Matomo with queuing and retries
type MatomoForwarder struct {
	matomoClient *matomoClient.Client
	repo        MatomoEventRepository
	logger      *zap.Logger
	batchSize   int
}

// MatomoEventRepository defines the interface for event persistence
type MatomoEventRepository interface {
	EnqueueEvent(ctx context.Context, event *MatomoStagedEvent) error
	GetPendingEvents(ctx context.Context, limit int) ([]*MatomoStagedEvent, error)
	UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status string, err error) error
	GetFailedEvents(ctx context.Context, limit int) ([]*MatomoStagedEvent, error)
	DeleteEvent(ctx context.Context, eventID uuid.UUID) error
}

// MatomoStagedEvent represents an event in the staging queue
type MatomoStagedEvent struct {
	ID            uuid.UUID
	EventType     string
	UserID        *uuid.UUID
	Payload       map[string]interface{}
	RetryCount    int
	MaxRetries    int
	NextRetryAt   time.Time
	Status        string
	CreatedAt     time.Time
	SentAt        *time.Time
	FailedAt      *time.Time
	ErrorMessage  *string
}

// NewMatomoForwarder creates a new Matomo forwarder service
func NewMatomoForwarder(
	matomoClient *matomoClient.Client,
	repo MatomoEventRepository,
	logger *zap.Logger,
) *MatomoForwarder {
	return &MatomoForwarder{
		matomoClient: matomoClient,
		repo:        repo,
		logger:      logger,
		batchSize:   100, // Process 100 events at a time
	}
}

// TrackEvent enqueues a standard event for delivery
func (f *MatomoForwarder) TrackEvent(ctx context.Context, userID *uuid.UUID, category, action, name string, value float64, customVars map[string]string) error {
	event := &MatomoStagedEvent{
		ID:        uuid.New(),
		EventType: "event",
		UserID:    userID,
		Payload: map[string]interface{}{
			"category":        category,
			"action":          action,
			"name":            name,
			"value":           value,
			"custom_variables": customVars,
			"event_time":      time.Now(),
		},
		Status:      "pending",
		MaxRetries:  3,
		NextRetryAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	if err := f.repo.EnqueueEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to enqueue event: %w", err)
	}

	f.logger.Debug("Enqueued Matomo event",
		zap.String("event_id", event.ID.String()),
		zap.String("type", event.EventType),
		zap.String("category", category),
		zap.String("action", action),
	)

	return nil
}

// TrackPurchase enqueues an ecommerce event for delivery
func (f *MatomoForwarder) TrackPurchase(ctx context.Context, userID *uuid.UUID, orderID string, revenue float64, items []matomoClient.EcommerceItem, customVars map[string]string) error {
	event := &MatomoStagedEvent{
		ID:        uuid.New(),
		EventType: "ecommerce",
		UserID:    userID,
		Payload: map[string]interface{}{
			"user_id":        userID.String(),
			"revenue":        revenue,
			"order_id":       orderID,
			"items":          items,
			"custom_variables": customVars,
			"event_time":     time.Now(),
		},
		Status:      "pending",
		MaxRetries:  3,
		NextRetryAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	if err := f.repo.EnqueueEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to enqueue ecommerce event: %w", err)
	}

	f.logger.Debug("Enqueued Matomo ecommerce event",
		zap.String("event_id", event.ID.String()),
		zap.String("order_id", orderID),
		zap.Float64("revenue", revenue),
		zap.Int("items", len(items)),
	)

	return nil
}

// ProcessBatch processes a batch of pending events
func (f *MatomoForwarder) ProcessBatch(ctx context.Context) (processed, succeeded, failed int, err error) {
	// Get pending events
	events, err := f.repo.GetPendingEvents(ctx, f.batchSize)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get pending events: %w", err)
	}

	if len(events) == 0 {
		return 0, 0, 0, nil
	}

	f.logger.Debug("Processing Matomo event batch", zap.Int("count", len(events)))

	processed = len(events)

	for _, event := range events {
		if err := f.processEvent(ctx, event); err != nil {
			f.logger.Error("Failed to process event",
				zap.String("event_id", event.ID.String()),
				zap.Error(err),
			)
			failed++
		} else {
			succeeded++
		}
	}

	f.logger.Debug("Processed Matomo event batch",
		zap.Int("processed", processed),
		zap.Int("succeeded", succeeded),
		zap.Int("failed", failed),
	)

	return processed, succeeded, failed, nil
}

// processEvent processes a single event
func (f *MatomoForwarder) processEvent(ctx context.Context, event *MatomoStagedEvent) error {
	var sendErr error

	switch event.EventType {
	case "event":
		sendErr = f.sendEvent(ctx, event)
	case "ecommerce":
		sendErr = f.sendEcommerce(ctx, event)
	default:
		sendErr = fmt.Errorf("unknown event type: %s", event.EventType)
	}

	// Update event status based on send result
	if err := f.repo.UpdateEventStatus(ctx, event.ID, "sent", sendErr); err != nil {
		f.logger.Error("Failed to update event status",
			zap.String("event_id", event.ID.String()),
			zap.Error(err),
		)
		return err
	}

	return sendErr
}

// sendEvent sends a standard event to Matomo
func (f *MatomoForwarder) sendEvent(ctx context.Context, event *MatomoStagedEvent) error {
	category, _ := event.Payload["category"].(string)
	action, _ := event.Payload["action"].(string)
	name, _ := event.Payload["name"].(string)
	value, _ := event.Payload["value"].(float64)

	// Extract custom variables
	customVars := make(map[string]string)
	if cv, ok := event.Payload["custom_variables"].(map[string]interface{}); ok {
		for k, v := range cv {
			if vs, ok := v.(string); ok {
				customVars[k] = vs
			}
		}
	}

	userID := ""
	if event.UserID != nil {
		userID = event.UserID.String()
	}

	req := matomoClient.TrackEventRequest{
		Category:        category,
		Action:          action,
		Name:            name,
		Value:           value,
		UserID:          userID,
		CustomVariables: customVars,
	}

	if err := f.matomoClient.TrackEvent(ctx, req); err != nil {
		return fmt.Errorf("failed to send event to Matomo: %w", err)
	}

	f.logger.Debug("Sent event to Matomo",
		zap.String("event_id", event.ID.String()),
		zap.String("category", category),
		zap.String("action", action),
	)

	return nil
}

// sendEcommerce sends an ecommerce event to Matomo
func (f *MatomoForwarder) sendEcommerce(ctx context.Context, event *MatomoStagedEvent) error {
	revenue, _ := event.Payload["revenue"].(float64)
	orderID, _ := event.Payload["order_id"].(string)

	// Extract items
	var items []matomoClient.EcommerceItem
	if itemsData, ok := event.Payload["items"].([]interface{}); ok {
		for _, itemData := range itemsData {
			if itemMap, ok := itemData.(map[string]interface{}); ok {
				item := matomoClient.EcommerceItem{
					SKU:      toString(itemMap["sku"]),
					Name:     toString(itemMap["name"]),
					Price:    toFloat64(itemMap["price"]),
					Quantity: toInt(itemMap["quantity"]),
					Category: toString(itemMap["category"]),
				}
				items = append(items, item)
			}
		}
	}

	// Extract custom variables
	customVars := make(map[string]string)
	if cv, ok := event.Payload["custom_variables"].(map[string]interface{}); ok {
		for k, v := range cv {
			if vs, ok := v.(string); ok {
				customVars[k] = vs
			}
		}
	}

	userID := ""
	if event.UserID != nil {
		userID = event.UserID.String()
	}

	req := matomoClient.TrackEcommerceRequest{
		UserID:     userID,
		Revenue:    revenue,
		OrderID:    orderID,
		Items:      items,
		CustomVars: customVars,
	}

	if err := f.matomoClient.TrackEcommerce(ctx, req); err != nil {
		return fmt.Errorf("failed to send ecommerce event to Matomo: %w", err)
	}

	f.logger.Debug("Sent ecommerce event to Matomo",
		zap.String("event_id", event.ID.String()),
		zap.String("order_id", orderID),
		zap.Float64("revenue", revenue),
	)

	return nil
}

// HandleError handles errors that occurred during event processing
func (f *MatomoForwarder) HandleError(ctx context.Context, eventID uuid.UUID, err error) error {
	if err == nil {
		// Success - mark as sent
		return f.repo.UpdateEventStatus(ctx, eventID, "sent", nil)
	}

	// Failure - will be retried or marked as permanently failed
	return f.repo.UpdateEventStatus(ctx, eventID, "pending", err)
}

// GetQueueSize returns the number of pending events
func (f *MatomoForwarder) GetQueueSize(ctx context.Context) (int, error) {
	events, err := f.repo.GetPendingEvents(ctx, 1000)
	if err != nil {
		return 0, err
	}
	// Note: This is an approximation since we limit to 1000
	// In production, add a COUNT query to the repository
	return len(events), nil
}

// GetFailedEventsCount returns the number of permanently failed events
func (f *MatomoForwarder) GetFailedEventsCount(ctx context.Context) (int, error) {
	events, err := f.repo.GetFailedEvents(ctx, 1000)
	if err != nil {
		return 0, err
	}
	return len(events), nil
}

// Helper functions for type assertions
func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toFloat64(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func toInt(v interface{}) int {
	if i, ok := v.(float64); ok {
		return int(i)
	}
	if i, ok := v.(int); ok {
		return i
	}
	return 0
}
