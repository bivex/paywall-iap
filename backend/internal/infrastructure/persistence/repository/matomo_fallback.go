package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// PostgresMatomoEventRepository implements event persistence using PostgreSQL
type PostgresMatomoEventRepository struct {
	pool   *pgxpool.Pool
	logger Logger
}

// NewPostgresMatomoEventRepository creates a new PostgreSQL-backed event repository
func NewPostgresMatomoEventRepository(pool *pgxpool.Pool, logger Logger) *PostgresMatomoEventRepository {
	return &PostgresMatomoEventRepository{
		pool:   pool,
		logger: logger,
	}
}

// EnqueueEvent inserts a new event into the staging queue
func (r *PostgresMatomoEventRepository) EnqueueEvent(ctx context.Context, event *service.MatomoStagedEvent) error {
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO matomo_staged_events (id, event_type, user_id, payload, retry_count, max_retries, next_retry_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	var userIDPtr *uuid.UUID
	if event.UserID != nil {
		userIDPtr = event.UserID
	}

	_, err = r.pool.Exec(ctx, query,
		event.ID,
		event.EventType,
		userIDPtr,
		payloadJSON,
		event.RetryCount,
		event.MaxRetries,
		event.NextRetryAt,
		event.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to enqueue event: %w", err)
	}

	r.logger.Debug("Enqueued Matomo event",
		"event_id", event.ID.String(),
		"type", event.EventType,
	)

	return nil
}

// GetPendingEvents retrieves events ready for processing
func (r *PostgresMatomoEventRepository) GetPendingEvents(ctx context.Context, limit int) ([]*service.MatomoStagedEvent, error) {
	query := `
		SELECT id, event_type, user_id, payload, retry_count, max_retries, next_retry_at, status, created_at, sent_at, failed_at, error_message
		FROM matomo_staged_events
		WHERE status = 'pending'
			AND next_retry_at <= NOW()
		ORDER BY next_retry_at ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	var events []*service.MatomoStagedEvent
	for rows.Next() {
		var event service.MatomoStagedEvent
		var payloadJSON []byte
		var sentAt, failedAt *time.Time
		var errorMessage *string

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.UserID,
			&payloadJSON,
			&event.RetryCount,
			&event.MaxRetries,
			&event.NextRetryAt,
			&event.Status,
			&event.CreatedAt,
			&sentAt,
			&failedAt,
			&errorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		event.SentAt = sentAt
		event.FailedAt = failedAt
		event.ErrorMessage = errorMessage

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// UpdateEventStatus updates the status of an event after processing attempt
func (r *PostgresMatomoEventRepository) UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status string, processErr error) error {
	if status == "sent" {
		// Mark as successfully sent
		query := `
			UPDATE matomo_staged_events
			SET status = 'sent',
				sent_at = NOW(),
				next_retry_at = NULL
			WHERE id = $1
		`
		_, err := r.pool.Exec(ctx, query, eventID)
		if err != nil {
			return fmt.Errorf("failed to mark event as sent: %w", err)
		}
		r.logger.Debug("Marked event as sent", "event_id", eventID.String())
		return nil
	}

	if processErr != nil {
		// Check if we should retry or mark as permanently failed
		var retryCount, maxRetries int
		checkQuery := `
			SELECT retry_count, max_retries
			FROM matomo_staged_events
			WHERE id = $1
		`
		err := r.pool.QueryRow(ctx, checkQuery, eventID).Scan(&retryCount, &maxRetries)
		if err != nil {
			return fmt.Errorf("failed to get retry info: %w", err)
		}

		if retryCount >= maxRetries {
			// Permanent failure
			errorMsg := processErr.Error()
			query := `
				UPDATE matomo_staged_events
				SET status = 'failed',
					failed_at = NOW(),
					error_message = $2
				WHERE id = $1
			`
			_, err = r.pool.Exec(ctx, query, eventID, errorMsg)
			if err != nil {
				return fmt.Errorf("failed to mark event as failed: %w", err)
			}
			r.logger.Warn("Event marked as permanently failed",
				"event_id", eventID.String(),
				"error", errorMsg,
			)
			return nil
		}

		// Retry with exponential backoff
		query := `
			UPDATE matomo_staged_events
			SET retry_count = retry_count + 1,
				status = 'pending',
				next_retry_at = calculate_next_retry(retry_count + 1, max_retries),
				error_message = $2
			WHERE id = $1
		`
		errorMsg := processErr.Error()
		_, err = r.pool.Exec(ctx, query, eventID, errorMsg)
		if err != nil {
			return fmt.Errorf("failed to schedule retry: %w", err)
		}
		r.logger.Debug("Event scheduled for retry",
			"event_id", eventID.String(),
			"retry_count", retryCount+1,
		)
		return nil
	}

	// Generic status update
	query := `
		UPDATE matomo_staged_events
		SET status = $2
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, eventID, status)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	return nil
}

// GetFailedEvents retrieves permanently failed events
func (r *PostgresMatomoEventRepository) GetFailedEvents(ctx context.Context, limit int) ([]*service.MatomoStagedEvent, error) {
	query := `
		SELECT id, event_type, user_id, payload, retry_count, max_retries, next_retry_at, status, created_at, sent_at, failed_at, error_message
		FROM matomo_staged_events
		WHERE status = 'failed'
		ORDER BY failed_at DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed events: %w", err)
	}
	defer rows.Close()

	var events []*service.MatomoStagedEvent
	for rows.Next() {
		var event service.MatomoStagedEvent
		var payloadJSON []byte
		var sentAt, failedAt *time.Time
		var errorMessage *string

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.UserID,
			&payloadJSON,
			&event.RetryCount,
			&event.MaxRetries,
			&event.NextRetryAt,
			&event.Status,
			&event.CreatedAt,
			&sentAt,
			&failedAt,
			&errorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		event.SentAt = sentAt
		event.FailedAt = failedAt
		event.ErrorMessage = errorMessage

		events = append(events, &event)
	}

	return events, nil
}

// DeleteEvent removes an event from the queue
func (r *PostgresMatomoEventRepository) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	query := `DELETE FROM matomo_staged_events WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// RetryFailedEvent resets a failed event for retry
func (r *PostgresMatomoEventRepository) RetryFailedEvent(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE matomo_staged_events
		SET status = 'pending',
			retry_count = 0,
			next_retry_at = NOW(),
			error_message = NULL,
			failed_at = NULL
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to retry failed event: %w", err)
	}
	r.logger.Info("Retrying failed event", "event_id", eventID.String())
	return nil
}

// CleanupOldSentEvents removes successfully sent events older than the specified duration
func (r *PostgresMatomoEventRepository) CleanupOldSentEvents(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM matomo_staged_events
		WHERE status = 'sent'
			AND sent_at < NOW() - $1::interval
	`

	result, err := r.pool.Exec(ctx, query, olderThan.String())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old events: %w", err)
	}

	count := result.RowsAffected()
	r.logger.Debug("Cleaned up old sent events", "count", count)
	return count, nil
}

// GetEventStats returns statistics about events in the queue
func (r *PostgresMatomoEventRepository) GetEventStats(ctx context.Context) (*EventStats, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending') AS pending,
			COUNT(*) FILTER (WHERE status = 'processing') AS processing,
			COUNT(*) FILTER (WHERE status = 'sent') AS sent,
			COUNT(*) FILTER (WHERE status = 'failed') AS failed,
			COUNT(*) AS total
		FROM matomo_staged_events
	`

	var stats EventStats
	err := r.pool.QueryRow(ctx, query).Scan(
		&stats.Pending,
		&stats.Processing,
		&stats.Sent,
		&stats.Failed,
		&stats.Total,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}

	return &stats, nil
}

// EventStats represents statistics about the event queue
type EventStats struct {
	Pending   int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Sent      int64 `json:"sent"`
	Failed    int64 `json:"failed"`
	Total     int64 `json:"total"`
}

// GetEventByID retrieves a specific event by ID
func (r *PostgresMatomoEventRepository) GetEventByID(ctx context.Context, eventID uuid.UUID) (*service.MatomoStagedEvent, error) {
	query := `
		SELECT id, event_type, user_id, payload, retry_count, max_retries, next_retry_at, status, created_at, sent_at, failed_at, error_message
		FROM matomo_staged_events
		WHERE id = $1
	`

	var event service.MatomoStagedEvent
	var payloadJSON []byte
	var sentAt, failedAt *time.Time
	var errorMessage *string

	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.ID,
		&event.EventType,
		&event.UserID,
		&payloadJSON,
		&event.RetryCount,
		&event.MaxRetries,
		&event.NextRetryAt,
		&event.Status,
		&event.CreatedAt,
		&sentAt,
		&failedAt,
		&errorMessage,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	event.SentAt = sentAt
	event.FailedAt = failedAt
	event.ErrorMessage = errorMessage

	return &event, nil
}
