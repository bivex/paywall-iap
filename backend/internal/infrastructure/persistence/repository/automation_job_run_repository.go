package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AutomationJobRunRepository struct {
	pool *pgxpool.Pool
}

func NewAutomationJobRunRepository(pool *pgxpool.Pool) *AutomationJobRunRepository {
	return &AutomationJobRunRepository{pool: pool}
}

func (r *AutomationJobRunRepository) ClaimAutomationJobRun(ctx context.Context, input service.AutomationJobRunClaimInput) (*service.AutomationJobRun, bool, error) {
	var payloadJSON []byte
	if len(input.Payload) > 0 {
		payloadJSON = input.Payload
	}

	var run service.AutomationJobRun
	err := r.pool.QueryRow(ctx, `
		WITH claimed AS (
			INSERT INTO automation_job_run_log (
				job_name,
				source,
				idempotency_key,
				status,
				payload,
				window_started_at,
				window_duration_seconds
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (idempotency_key) DO UPDATE
			SET source = EXCLUDED.source,
			    status = EXCLUDED.status,
			    payload = EXCLUDED.payload,
			    window_started_at = EXCLUDED.window_started_at,
			    window_duration_seconds = EXCLUDED.window_duration_seconds,
			    started_at = now(),
			    finished_at = NULL,
			    details = NULL
			WHERE automation_job_run_log.status = 'failed'
			RETURNING id, job_name, idempotency_key, status
		)
		SELECT id, job_name, idempotency_key, status FROM claimed`,
		input.JobName,
		input.Source,
		input.IdempotencyKey,
		service.AutomationJobRunStatusRunning,
		payloadJSON,
		input.WindowStartedAt,
		int(input.WindowDuration.Seconds()),
	).Scan(&run.ID, &run.JobName, &run.IdempotencyKey, &run.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to claim automation job run: %w", err)
	}

	return &run, true, nil
}

func (r *AutomationJobRunRepository) FinishAutomationJobRun(ctx context.Context, runID uuid.UUID, status string, details map[string]any) error {
	var detailsJSON []byte
	var err error
	if len(details) > 0 {
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			return fmt.Errorf("failed to marshal automation job run details: %w", err)
		}
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE automation_job_run_log
		SET status = $2,
		    details = $3,
		    finished_at = now()
		WHERE id = $1`, runID, status, detailsJSON)
	if err != nil {
		return fmt.Errorf("failed to finish automation job run: %w", err)
	}
	return nil
}
