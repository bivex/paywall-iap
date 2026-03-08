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
	var automationPolicyJSON []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, status, start_at, end_at, automation_policy
		FROM ab_tests
		WHERE id = $1`, experimentID).Scan(&state.ID, &state.Status, &startAt, &endAt, &automationPolicyJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrExperimentNotFound
		}
		return nil, fmt.Errorf("failed to load experiment mutation state: %w", err)
	}

	state.StartAt = startAt
	state.EndAt = endAt
	state.AutomationPolicy = service.DefaultExperimentAutomationPolicy()
	if len(automationPolicyJSON) > 0 {
		var policy service.ExperimentAutomationPolicy
		if err := json.Unmarshal(automationPolicyJSON, &policy); err != nil {
			return nil, fmt.Errorf("failed to decode experiment automation policy: %w", err)
		}
		state.AutomationPolicy = service.NormalizeExperimentAutomationPolicy(&policy)
	}
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
	return r.UpdateExperimentStatusWithAudit(ctx, experimentID, "", nextStatus, startAt, endAt, nil)
}

func (r *ExperimentAdminRepository) UpdateExperimentAutomationPolicy(ctx context.Context, experimentID uuid.UUID, policy service.ExperimentAutomationPolicy) error {
	automationPolicyJSON, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal experiment automation policy: %w", err)
	}

	commandTag, err := r.pool.Exec(ctx, `
		UPDATE ab_tests
		SET automation_policy = $2,
		    updated_at = now()
		WHERE id = $1`, experimentID, automationPolicyJSON)
	if err != nil {
		return fmt.Errorf("failed to update experiment automation policy: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return service.ErrExperimentNotFound
	}
	return nil
}

func (r *ExperimentAdminRepository) UpdateExperimentStatusWithAudit(ctx context.Context, experimentID uuid.UUID, currentStatus, nextStatus string, startAt, endAt *time.Time, audit *service.ExperimentStatusTransitionAudit) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin experiment status transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
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

	if audit != nil {
		var detailsJSON []byte
		if audit.Details != nil {
			detailsJSON, err = json.Marshal(audit.Details)
			if err != nil {
				return fmt.Errorf("failed to marshal experiment status audit details: %w", err)
			}
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO experiment_lifecycle_audit_log (
				experiment_id, actor_type, actor_id, source, action, from_status, to_status, idempotency_key, details
			)
			VALUES ($1, $2, $3, $4, 'status_transition', $5, $6, $7, $8)
			ON CONFLICT (idempotency_key) DO NOTHING`,
			experimentID,
			audit.ActorType,
			audit.ActorID,
			audit.Source,
			currentStatus,
			nextStatus,
			audit.IdempotencyKey,
			detailsJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert experiment lifecycle audit log: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit experiment status transaction: %w", err)
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

func (r *ExperimentAdminRepository) CountExperimentAssignments(ctx context.Context, experimentID uuid.UUID) (total int, active int, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::int FROM ab_test_assignments WHERE experiment_id = $1),
			(SELECT COUNT(*)::int FROM ab_test_assignments WHERE experiment_id = $1 AND expires_at > NOW())`, experimentID).Scan(&total, &active)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count experiment assignments: %w", err)
	}
	return total, active, nil
}

func (r *ExperimentAdminRepository) EnsureExperimentArmStats(ctx context.Context, experimentID uuid.UUID) (int, error) {
	commandTag, err := r.pool.Exec(ctx, `
		INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward)
		SELECT a.id, 1.0, 1.0, 0, 0, 0, 0
		FROM ab_test_arms a
		LEFT JOIN ab_test_arm_stats s ON s.arm_id = a.id
		WHERE a.experiment_id = $1
		  AND s.arm_id IS NULL`, experimentID)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure experiment arm stats: %w", err)
	}
	return int(commandTag.RowsAffected()), nil
}

func (r *ExperimentAdminRepository) CountExperimentPendingRewards(ctx context.Context, experimentID uuid.UUID) (total int, expired int, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(COUNT(*)::int, 0) AS total,
			COALESCE(COUNT(*) FILTER (WHERE expires_at < NOW() AND converted = FALSE AND processed_at IS NULL)::int, 0) AS expired
		FROM bandit_pending_rewards
		WHERE experiment_id = $1`, experimentID).Scan(&total, &expired)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count experiment pending rewards: %w", err)
	}
	return total, expired, nil
}

func (r *ExperimentAdminRepository) ProcessExpiredPendingRewards(ctx context.Context, experimentID uuid.UUID) (int, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to begin expired pending rewards transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		WITH expired AS (
			UPDATE bandit_pending_rewards
			SET processed_at = NOW()
			WHERE experiment_id = $1
			  AND expires_at < NOW()
			  AND converted = FALSE
			  AND processed_at IS NULL
			RETURNING arm_id
		)
		SELECT arm_id, COUNT(*)::int AS processed_count
		FROM expired
		GROUP BY arm_id`, experimentID)
	if err != nil {
		return 0, fmt.Errorf("failed to query expired pending rewards: %w", err)
	}

	countsByArm := make(map[uuid.UUID]int)
	processed := 0
	for rows.Next() {
		var armID uuid.UUID
		var count int
		if err := rows.Scan(&armID, &count); err != nil {
			rows.Close()
			return 0, fmt.Errorf("failed to scan expired pending reward count: %w", err)
		}
		countsByArm[armID] = count
		processed += count
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("failed to iterate expired pending rewards: %w", err)
	}
	rows.Close()

	for armID, count := range countsByArm {
		_, err := tx.Exec(ctx, `
			UPDATE ab_test_arm_stats
			SET beta = beta + $2,
			    samples = samples + $2,
			    avg_reward = CASE
			        WHEN samples + $2 > 0 THEN revenue / (samples + $2)
			        ELSE 0
			    END,
			    updated_at = NOW()
			WHERE arm_id = $1`, armID, count)
		if err != nil {
			return 0, fmt.Errorf("failed to update expired pending reward stats: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit expired pending rewards transaction: %w", err)
	}
	return processed, nil
}

func (r *ExperimentAdminRepository) UpdateExperimentWinnerConfidence(ctx context.Context, experimentID uuid.UUID, confidence *float64) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE ab_tests
		SET winner_confidence = $2,
		    updated_at = now()
		WHERE id = $1`, experimentID, confidence)
	if err != nil {
		return fmt.Errorf("failed to update experiment winner confidence: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return service.ErrExperimentNotFound
	}
	return nil
}

func (r *ExperimentAdminRepository) ListExperimentRepairCandidateIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT e.id
		FROM ab_tests e
		WHERE e.is_bandit = TRUE
		  AND (
			e.status IN ('running', 'paused')
			OR EXISTS (
				SELECT 1
				FROM bandit_pending_rewards pr
				WHERE pr.experiment_id = e.id
				  AND pr.processed_at IS NULL
			)
			OR EXISTS (
				SELECT 1
				FROM ab_test_arms a
				LEFT JOIN ab_test_arm_stats s ON s.arm_id = a.id
				WHERE a.experiment_id = e.id
				  AND s.arm_id IS NULL
			)
			OR (e.status = 'completed' AND e.winner_confidence IS NULL)
		  )
		ORDER BY e.updated_at DESC, e.id
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query experiment repair candidates: %w", err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan experiment repair candidate: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate experiment repair candidates: %w", err)
	}

	return ids, nil
}
