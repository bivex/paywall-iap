package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

type ExperimentAdminRepository struct {
	pool *pgxpool.Pool
}

func NewExperimentAdminRepository(pool *pgxpool.Pool) *ExperimentAdminRepository {
	return &ExperimentAdminRepository{pool: pool}
}

func (r *ExperimentAdminRepository) GetExperimentMutationState(ctx context.Context, experimentID uuid.UUID) (*service.ExperimentMutationState, error) {
	var state service.ExperimentMutationState
	var startAt *time.Time
	var endAt *time.Time

	err := r.pool.QueryRow(ctx, `
		SELECT id, status, start_at, end_at
		FROM ab_tests
		WHERE id = $1`, experimentID).Scan(&state.ID, &state.Status, &startAt, &endAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrExperimentNotFound
		}
		return nil, fmt.Errorf("failed to load experiment mutation state: %w", err)
	}

	state.StartAt = startAt
	state.EndAt = endAt
	return &state, nil
}

func (r *ExperimentAdminRepository) UpdateExperimentDraft(ctx context.Context, experimentID uuid.UUID, input service.UpdateExperimentInput) error {
	automationPolicyJSON, err := json.Marshal(input.AutomationPolicy)
	if err != nil {
		return fmt.Errorf("failed to marshal experiment automation policy: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE ab_tests
		SET name = $2,
		    description = $3,
		    algorithm_type = $4,
		    is_bandit = $5,
		    min_sample_size = $6,
		    confidence_threshold = $7,
		    start_at = $8,
		    end_at = $9,
		    automation_policy = $10,
		    updated_at = now()
		WHERE id = $1`,
		experimentID,
		input.Name,
		input.Description,
		input.AlgorithmType,
		input.IsBandit,
		input.MinSampleSize,
		input.ConfidenceThreshold,
		input.StartAt,
		input.EndAt,
		automationPolicyJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update experiment draft: %w", err)
	}
	return nil
}

func (r *ExperimentAdminRepository) UpdateExperimentStatus(ctx context.Context, experimentID uuid.UUID, nextStatus string, startAt, endAt *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ab_tests
		SET status = $2,
		    start_at = $3,
		    end_at = $4,
		    updated_at = now()
		WHERE id = $1`,
		experimentID,
		nextStatus,
		startAt,
		endAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update experiment status: %w", err)
	}
	return nil
}

func (r *ExperimentAdminRepository) ListExperimentAutomationStates(ctx context.Context) ([]service.ExperimentAutomationState, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id,
		       e.status,
		       e.start_at,
		       e.end_at,
		       e.min_sample_size,
		       e.confidence_threshold::double precision,
		       e.winner_confidence::double precision,
		       COALESCE((SELECT SUM(s.samples)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_samples,
		       e.automation_policy
		FROM ab_tests e
		WHERE e.status IN ('draft', 'running')`)
	if err != nil {
		return nil, fmt.Errorf("failed to query experiment automation states: %w", err)
	}
	defer rows.Close()

	states := make([]service.ExperimentAutomationState, 0)
	for rows.Next() {
		var state service.ExperimentAutomationState
		var winnerConfidence sql.NullFloat64
		var automationPolicyJSON []byte

		if err := rows.Scan(
			&state.ID,
			&state.Status,
			&state.StartAt,
			&state.EndAt,
			&state.MinSampleSize,
			&state.ConfidenceThreshold,
			&winnerConfidence,
			&state.TotalSamples,
			&automationPolicyJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan experiment automation state: %w", err)
		}

		state.AutomationPolicy = service.DefaultExperimentAutomationPolicy()
		if len(automationPolicyJSON) > 0 {
			var policy service.ExperimentAutomationPolicy
			if err := json.Unmarshal(automationPolicyJSON, &policy); err != nil {
				return nil, fmt.Errorf("failed to decode experiment automation policy: %w", err)
			}
			state.AutomationPolicy = service.NormalizeExperimentAutomationPolicy(&policy)
		}
		if winnerConfidence.Valid {
			value := winnerConfidence.Float64
			state.WinnerConfidence = &value
		}

		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate experiment automation states: %w", err)
	}

	return states, nil
}
