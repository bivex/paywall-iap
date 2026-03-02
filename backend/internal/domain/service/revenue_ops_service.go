package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RevenueOpsService handles revenue operations dashboard data
type RevenueOpsService struct {
	dbPool *pgxpool.Pool
}

// NewRevenueOpsService creates a new revenue ops service
func NewRevenueOpsService(dbPool *pgxpool.Pool) *RevenueOpsService {
	return &RevenueOpsService{dbPool: dbPool}
}

// DunningQueueRow represents a row in the dunning queue
type DunningQueueRow struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	UserID      string  `json:"user_id"`
	PlanType    string  `json:"plan_type"`
	Status      string  `json:"status"`
	AttemptCount int    `json:"attempt_count"`
	MaxAttempts int     `json:"max_attempts"`
	NextAttempt *string `json:"next_attempt_at"`
	LastAttempt *string `json:"last_attempt_at"`
	CreatedAt   string  `json:"created_at"`
}

// DunningStats holds dunning statistics
type DunningStats struct {
	Pending    int `json:"pending"`
	InProgress int `json:"in_progress"`
	Recovered  int `json:"recovered"`
	Failed     int `json:"failed"`
}

// DunningData contains dunning queue and stats
type DunningData struct {
	Queue []DunningQueueRow `json:"queue"`
	Stats DunningStats      `json:"stats"`
}

// WebhookRow represents a webhook event row
type WebhookRow struct {
	ID          string  `json:"id"`
	Provider    string  `json:"provider"`
	EventType   string  `json:"event_type"`
	EventID     string  `json:"event_id"`
	Processed   bool    `json:"processed"`
	ProcessedAt *string `json:"processed_at"`
	CreatedAt   string  `json:"created_at"`
}

// WebhookProviderStat holds provider statistics
type WebhookProviderStat struct {
	Provider  string `json:"provider"`
	Total     int    `json:"total"`
	Processed int    `json:"processed"`
}

// WebhookData contains webhook events and statistics
type WebhookData struct {
	Events         []WebhookRow         `json:"events"`
	PendingEvents  []WebhookRow         `json:"pending_events"`
	Total          int                  `json:"total"`
	Unprocessed    int                  `json:"unprocessed"`
	ByProvider     []WebhookProviderStat `json:"by_provider"`
	Page           int                  `json:"page"`
	PageSize       int                  `json:"page_size"`
	TotalPages     int                  `json:"total_pages"`
}

// MatomoStats holds Matomo staging statistics
type MatomoStats struct {
	Pending    int `json:"pending"`
	Processing int `json:"processing"`
	Sent       int `json:"sent"`
	Failed     int `json:"failed"`
	Total      int `json:"total"`
}

// RevenueOpsReport contains all revenue operations data
type RevenueOpsReport struct {
	Dunning  DunningData `json:"dunning"`
	Webhooks WebhookData `json:"webhooks"`
	Matomo   struct {
		Stats MatomoStats `json:"stats"`
	} `json:"matomo"`
}

// GetReport fetches the complete revenue operations report
func (s *RevenueOpsService) GetReport(ctx context.Context, page, pageSize int, pendingOnly bool) (*RevenueOpsReport, error) {
	dunning, err := s.fetchDunningData(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch dunning data: %w", err)
	}

	webhooks, err := s.fetchWebhookData(ctx, page, pageSize, pendingOnly)
	if err != nil {
		return nil, fmt.Errorf("fetch webhook data: %w", err)
	}

	matomoStats := s.fetchMatomoStats(ctx)

	return &RevenueOpsReport{
		Dunning:  *dunning,
		Webhooks: *webhooks,
		Matomo: struct {
			Stats MatomoStats `json:"stats"`
		}{
			Stats: matomoStats,
		},
	}, nil
}

// fetchDunningData retrieves dunning queue and statistics
func (s *RevenueOpsService) fetchDunningData(ctx context.Context) (*DunningData, error) {
	queue, err := s.fetchDunningQueue(ctx)
	if err != nil {
		return nil, err
	}

	stats := s.fetchDunningStats(ctx)

	return &DunningData{
		Queue: queue,
		Stats: stats,
	}, nil
}

// fetchDunningQueue retrieves active dunning queue entries
func (s *RevenueOpsService) fetchDunningQueue(ctx context.Context) ([]DunningQueueRow, error) {
	rows, err := s.dbPool.Query(ctx, `
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

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queue []DunningQueueRow
	for rows.Next() {
		var row DunningQueueRow
		var nextAt, lastAt *time.Time
		var createdAt time.Time
		var id, uid string

		if scanErr := rows.Scan(&id, &row.Email, &uid, &row.PlanType,
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

		queue = append(queue, row)
	}

	return queue, nil
}

// fetchDunningStats retrieves dunning statistics grouped by status
func (s *RevenueOpsService) fetchDunningStats(ctx context.Context) DunningStats {
	var stats DunningStats

	rows, _ := s.dbPool.Query(ctx, `SELECT status, COUNT(*) FROM dunning GROUP BY status`)
	if rows == nil {
		return stats
	}
	defer rows.Close()

	for rows.Next() {
		var st string
		var cnt int
		if err := rows.Scan(&st, &cnt); err != nil {
			continue
		}

		switch st {
		case "pending":
			stats.Pending = cnt
		case "in_progress":
			stats.InProgress = cnt
		case "recovered":
			stats.Recovered = cnt
		case "failed":
			stats.Failed = cnt
		}
	}

	return stats
}

// fetchWebhookData retrieves webhook events and statistics
func (s *RevenueOpsService) fetchWebhookData(ctx context.Context, page, pageSize int, pendingOnly bool) (*WebhookData, error) {
	offset := (page - 1) * pageSize

	events, err := s.fetchWebhookEvents(ctx, pageSize, offset, pendingOnly)
	if err != nil {
		return nil, err
	}

	pendingEvents := s.fetchPendingWebhooks(ctx)
	total, unprocessed := s.fetchWebhookCounts(ctx)
	providerStats := s.fetchWebhookProviderStats(ctx)

	paginationTotal := total
	if pendingOnly {
		paginationTotal = unprocessed
	}

	totalPages := paginationTotal / pageSize
	if paginationTotal%pageSize > 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}

	return &WebhookData{
		Events:        events,
		PendingEvents: pendingEvents,
		Total:         paginationTotal,
		Unprocessed:   unprocessed,
		ByProvider:    providerStats,
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    totalPages,
	}, nil
}

// fetchWebhookEvents retrieves paginated webhook events
func (s *RevenueOpsService) fetchWebhookEvents(ctx context.Context, limit, offset int, pendingOnly bool) ([]WebhookRow, error) {
	whereClause := ""
	if pendingOnly {
		whereClause = "WHERE processed_at IS NULL "
	}

	query := `
		SELECT id, provider, event_type, event_id, processed_at, created_at
		FROM webhook_events
		` + whereClause + `ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.dbPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []WebhookRow
	for rows.Next() {
		var row WebhookRow
		var processedAt *time.Time
		var createdAt time.Time

		if scanErr := rows.Scan(&row.ID, &row.Provider, &row.EventType, &row.EventID, &processedAt, &createdAt); scanErr != nil {
			continue
		}

		row.CreatedAt = createdAt.Format(time.RFC3339)
		if processedAt != nil {
			s := processedAt.Format(time.RFC3339)
			row.ProcessedAt = &s
			row.Processed = true
		}

		events = append(events, row)
	}

	return events, nil
}

// fetchPendingWebhooks retrieves all unprocessed webhook events
func (s *RevenueOpsService) fetchPendingWebhooks(ctx context.Context) []WebhookRow {
	rows, _ := s.dbPool.Query(ctx, `
		SELECT id, provider, event_type, event_id, created_at
		FROM webhook_events
		WHERE processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT 200
	`)
	if rows == nil {
		return nil
	}
	defer rows.Close()

	var events []WebhookRow
	for rows.Next() {
		var row WebhookRow
		var createdAt time.Time

		if scanErr := rows.Scan(&row.ID, &row.Provider, &row.EventType, &row.EventID, &createdAt); scanErr != nil {
			continue
		}

		row.CreatedAt = createdAt.Format(time.RFC3339)
		row.Processed = false

		events = append(events, row)
	}

	return events
}

// fetchWebhookCounts retrieves total and unprocessed webhook counts
func (s *RevenueOpsService) fetchWebhookCounts(ctx context.Context) (total, unprocessed int) {
	_ = s.dbPool.QueryRow(ctx, `SELECT COUNT(*) FROM webhook_events`).Scan(&total)
	_ = s.dbPool.QueryRow(ctx, `SELECT COUNT(*) FROM webhook_events WHERE processed_at IS NULL`).Scan(&unprocessed)
	return
}

// fetchWebhookProviderStats retrieves statistics grouped by provider
func (s *RevenueOpsService) fetchWebhookProviderStats(ctx context.Context) []WebhookProviderStat {
	rows, _ := s.dbPool.Query(ctx, `
		SELECT provider,
		       COUNT(*) as total,
		       COUNT(processed_at) as processed
		FROM webhook_events
		GROUP BY provider
		ORDER BY total DESC
	`)
	if rows == nil {
		return nil
	}
	defer rows.Close()

	var stats []WebhookProviderStat
	for rows.Next() {
		var p WebhookProviderStat
		if err := rows.Scan(&p.Provider, &p.Total, &p.Processed); err == nil {
			stats = append(stats, p)
		}
	}

	return stats
}

// fetchMatomoStats retrieves Matomo staging statistics
func (s *RevenueOpsService) fetchMatomoStats(ctx context.Context) MatomoStats {
	var stats MatomoStats

	rows, _ := s.dbPool.Query(ctx, `SELECT status, COUNT(*) FROM matomo_staged_events GROUP BY status`)
	if rows == nil {
		return stats
	}
	defer rows.Close()

	for rows.Next() {
		var st string
		var cnt int
		if scanErr := rows.Scan(&st, &cnt); scanErr != nil {
			continue
		}

		switch st {
		case "pending":
			stats.Pending = cnt
		case "processing":
			stats.Processing = cnt
		case "sent":
			stats.Sent = cnt
		case "failed":
			stats.Failed = cnt
		}
	}

	stats.Total = stats.Pending + stats.Processing + stats.Sent + stats.Failed
	return stats
}
