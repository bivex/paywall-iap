package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// PostgresBanditRepository implements bandit data persistence using PostgreSQL
type PostgresBanditRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

type pendingRewardScanner interface {
	Scan(dest ...any) error
}

// NewPostgresBanditRepository creates a new PostgreSQL-backed bandit repository
func NewPostgresBanditRepository(pool *pgxpool.Pool, logger *zap.Logger) *PostgresBanditRepository {
	return &PostgresBanditRepository{
		pool:   pool,
		logger: logger,
	}
}

func scanPendingReward(scanner pendingRewardScanner, reward *service.PendingReward) error {
	var conversionValue sql.NullFloat64
	var conversionCurrency sql.NullString
	var convertedAt sql.NullTime
	var processedAt sql.NullTime

	err := scanner.Scan(
		&reward.ID,
		&reward.ExperimentID,
		&reward.ArmID,
		&reward.UserID,
		&reward.AssignedAt,
		&reward.ExpiresAt,
		&reward.Converted,
		&conversionValue,
		&conversionCurrency,
		&convertedAt,
		&processedAt,
	)
	if err != nil {
		return err
	}

	if conversionValue.Valid {
		reward.ConversionValue = conversionValue.Float64
	}
	if conversionCurrency.Valid {
		reward.ConversionCurrency = conversionCurrency.String
	}
	if convertedAt.Valid {
		t := convertedAt.Time
		reward.ConvertedAt = &t
	}
	if processedAt.Valid {
		t := processedAt.Time
		reward.ProcessedAt = &t
	}

	return nil
}

// GetArms retrieves all arms for an experiment
func (r *PostgresBanditRepository) GetArms(ctx context.Context, experimentID uuid.UUID) ([]service.Arm, error) {
	query := `
		SELECT id, experiment_id, name, description, is_control, traffic_weight
		FROM ab_test_arms
		WHERE experiment_id = $1
		ORDER BY is_control DESC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query arms: %w", err)
	}
	defer rows.Close()

	var arms []service.Arm
	for rows.Next() {
		var arm service.Arm
		if err := rows.Scan(
			&arm.ID,
			&arm.ExperimentID,
			&arm.Name,
			&arm.Description,
			&arm.IsControl,
			&arm.TrafficWeight,
		); err != nil {
			return nil, fmt.Errorf("failed to scan arm: %w", err)
		}
		arms = append(arms, arm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating arms: %w", err)
	}

	return arms, nil
}

// GetArmStats retrieves statistics for a specific arm
func (r *PostgresBanditRepository) GetArmStats(ctx context.Context, armID uuid.UUID) (*service.ArmStats, error) {
	query := `
		SELECT arm_id, alpha, beta, samples, conversions, revenue, avg_reward, updated_at
		FROM ab_test_arm_stats
		WHERE arm_id = $1
	`

	var stats service.ArmStats
	err := r.pool.QueryRow(ctx, query, armID).Scan(
		&stats.ArmID,
		&stats.Alpha,
		&stats.Beta,
		&stats.Samples,
		&stats.Conversions,
		&stats.Revenue,
		&stats.AvgReward,
		&stats.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		var armExists bool
		existsQuery := `SELECT EXISTS(SELECT 1 FROM ab_test_arms WHERE id = $1)`
		if existsErr := r.pool.QueryRow(ctx, existsQuery, armID).Scan(&armExists); existsErr != nil {
			return nil, fmt.Errorf("failed to verify arm existence: %w", existsErr)
		}

		if !armExists {
			return nil, service.ErrBanditArmNotFound
		}

		// Return default stats if not found
		return &service.ArmStats{
			ArmID:       armID,
			Alpha:       1.0, // Uniform prior
			Beta:        1.0,
			Samples:     0,
			Conversions: 0,
			Revenue:     0,
			AvgReward:   0,
			UpdatedAt:   time.Now(),
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get arm stats: %w", err)
	}

	return &stats, nil
}

// UpdateArmStats updates statistics for a specific arm
func (r *PostgresBanditRepository) UpdateArmStats(ctx context.Context, stats *service.ArmStats) error {
	query := `
		INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (arm_id)
		DO UPDATE SET
			alpha = $2,
			beta = $3,
			samples = $4,
			conversions = $5,
			revenue = $6,
			avg_reward = $7,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query,
		stats.ArmID,
		stats.Alpha,
		stats.Beta,
		stats.Samples,
		stats.Conversions,
		stats.Revenue,
		stats.AvgReward,
	)

	if err != nil {
		return fmt.Errorf("failed to update arm stats: %w", err)
	}

	r.logger.Debug("Updated arm stats",
		zap.String("arm_id", stats.ArmID.String()),
		zap.Float64("alpha", stats.Alpha),
		zap.Float64("beta", stats.Beta),
		zap.Int("samples", stats.Samples),
	)

	return nil
}

// CreateAssignment creates a new user assignment
func (r *PostgresBanditRepository) CreateAssignment(ctx context.Context, assignment *service.Assignment) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin assignment transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	assignedAt := normalizeOccurredAt(assignment.AssignedAt)
	var assignmentID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO ab_test_assignments (id, experiment_id, user_id, arm_id, assigned_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (experiment_id, user_id)
		DO UPDATE SET
			arm_id = EXCLUDED.arm_id,
			assigned_at = EXCLUDED.assigned_at,
			expires_at = EXCLUDED.expires_at
		RETURNING id
	`,
		assignment.ID,
		assignment.ExperimentID,
		assignment.UserID,
		assignment.ArmID,
		assignedAt,
		assignment.ExpiresAt,
	).Scan(&assignmentID)
	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO bandit_assignment_events (
			assignment_id,
			experiment_id,
			user_id,
			arm_id,
			event_type,
			occurred_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		assignmentID,
		assignment.ExperimentID,
		assignment.UserID,
		assignment.ArmID,
		service.AssignmentEventTypeAssigned,
		assignedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to append assignment event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit assignment transaction: %w", err)
	}
	assignment.ID = assignmentID

	r.logger.Debug("Created assignment",
		zap.String("experiment_id", assignment.ExperimentID.String()),
		zap.String("user_id", assignment.UserID.String()),
		zap.String("arm_id", assignment.ArmID.String()),
	)

	return nil
}

// GetActiveAssignment retrieves the active (non-expired) assignment for a user in an experiment
func (r *PostgresBanditRepository) GetActiveAssignment(ctx context.Context, experimentID, userID uuid.UUID) (*service.Assignment, error) {
	query := `
		SELECT id, experiment_id, user_id, arm_id, assigned_at, expires_at
		FROM ab_test_assignments
		WHERE experiment_id = $1
			AND user_id = $2
			AND expires_at > NOW()
		ORDER BY assigned_at DESC
		LIMIT 1
	`

	var assignment service.Assignment
	err := r.pool.QueryRow(ctx, query, experimentID, userID).Scan(
		&assignment.ID,
		&assignment.ExperimentID,
		&assignment.UserID,
		&assignment.ArmID,
		&assignment.AssignedAt,
		&assignment.ExpiresAt,
	)

	if err == pgx.ErrNoRows {
		return nil, service.ErrAssignmentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	return &assignment, nil
}

// SaveConversion records a conversion event for an arm
func (r *PostgresBanditRepository) SaveConversion(ctx context.Context, experimentID, armID, userID uuid.UUID, amount float64) error {
	return r.AppendConversionEvent(ctx, &service.ConversionEvent{
		ExperimentID:          experimentID,
		ArmID:                 armID,
		UserID:                &userID,
		EventType:             service.ConversionEventTypeDirectReward,
		OriginalRewardValue:   amount,
		NormalizedRewardValue: amount,
		OccurredAt:            time.Now().UTC(),
	})
}

func (r *PostgresBanditRepository) AppendConversionEvent(ctx context.Context, event *service.ConversionEvent) error {
	var metadataJSON []byte
	var err error
	if event.Metadata != nil {
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal conversion event metadata: %w", err)
		}
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO bandit_conversion_events (
			experiment_id,
			arm_id,
			user_id,
			pending_reward_id,
			transaction_id,
			event_type,
			original_reward_value,
			original_currency,
			normalized_reward_value,
			normalized_currency,
			metadata,
			occurred_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, NULLIF($10, ''), $11, $12)
	`,
		event.ExperimentID,
		event.ArmID,
		nullableUUID(event.UserID),
		nullableUUID(event.PendingRewardID),
		nullableUUID(event.TransactionID),
		event.EventType,
		event.OriginalRewardValue,
		event.OriginalCurrency,
		event.NormalizedRewardValue,
		event.NormalizedCurrency,
		metadataJSON,
		normalizeOccurredAt(event.OccurredAt),
	)
	if err != nil {
		return fmt.Errorf("failed to append conversion event: %w", err)
	}

	return nil
}

func (r *PostgresBanditRepository) AppendImpressionEvent(ctx context.Context, event *service.ImpressionEvent) error {
	var metadataJSON []byte
	var err error
	if event.Metadata != nil {
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal impression event metadata: %w", err)
		}
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO bandit_impression_events (
			experiment_id,
			arm_id,
			user_id,
			event_type,
			metadata,
			occurred_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		event.ExperimentID,
		event.ArmID,
		event.UserID,
		event.EventType,
		metadataJSON,
		normalizeOccurredAt(event.OccurredAt),
	)
	if err != nil {
		return fmt.Errorf("failed to append impression event: %w", err)
	}

	return nil
}

func (r *PostgresBanditRepository) AppendWinnerRecommendationEvent(ctx context.Context, event *service.WinnerRecommendationEvent) error {
	var detailsJSON []byte
	var err error
	if event.Details != nil {
		detailsJSON, err = json.Marshal(event.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal winner recommendation details: %w", err)
		}
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO experiment_winner_recommendation_log (
			experiment_id,
			source,
			recommended,
			reason,
			winning_arm_id,
			confidence_percent,
			confidence_threshold_percent,
			observed_samples,
			min_sample_size,
			details,
			occurred_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		event.ExperimentID,
		event.Source,
		event.Recommended,
		event.Reason,
		event.WinningArmID,
		event.ConfidencePercent,
		event.ConfidenceThresholdPercent,
		event.ObservedSamples,
		event.MinSampleSize,
		detailsJSON,
		normalizeOccurredAt(event.OccurredAt),
	)
	if err != nil {
		return fmt.Errorf("failed to append winner recommendation event: %w", err)
	}

	return nil
}

func (r *PostgresBanditRepository) ProcessPendingConversion(ctx context.Context, transactionID, userID uuid.UUID, conversionValue float64, currency string, processedAt time.Time) (*service.PendingReward, bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, false, fmt.Errorf("failed to begin pending conversion transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	matchedPending := &service.PendingReward{}
	err = scanPendingReward(tx.QueryRow(ctx, `
		SELECT id, experiment_id, arm_id, user_id, assigned_at, expires_at, converted,
		       conversion_value, conversion_currency, converted_at, processed_at
		FROM bandit_pending_rewards
		WHERE user_id = $1
		  AND converted = FALSE
		  AND processed_at IS NULL
		  AND expires_at > $2
		ORDER BY assigned_at DESC
		LIMIT 1
		FOR UPDATE
	`, userID, processedAt), matchedPending)
	if err == pgx.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to load pending reward for conversion: %w", err)
	}

	inserted, err := r.insertConversionEventTx(ctx, tx, &service.ConversionEvent{
		ExperimentID:          matchedPending.ExperimentID,
		ArmID:                 matchedPending.ArmID,
		UserID:                &matchedPending.UserID,
		PendingRewardID:       &matchedPending.ID,
		TransactionID:         &transactionID,
		EventType:             service.ConversionEventTypeDelayedConversion,
		OriginalRewardValue:   conversionValue,
		OriginalCurrency:      currency,
		NormalizedRewardValue: conversionValue,
		NormalizedCurrency:    currency,
		Metadata: map[string]interface{}{
			"source": "postgres_bandit_repository",
		},
		OccurredAt: processedAt,
	})
	if err != nil {
		return nil, false, err
	}
	if !inserted {
		if err := tx.Commit(ctx); err != nil {
			return nil, false, fmt.Errorf("failed to commit duplicate pending conversion transaction: %w", err)
		}
		return matchedPending, false, nil
	}

	if err := r.applyRewardToArmTx(ctx, tx, matchedPending.ArmID, conversionValue); err != nil {
		return nil, false, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE bandit_pending_rewards
		SET converted = TRUE,
		    conversion_value = $2,
		    conversion_currency = NULLIF($3, ''),
		    converted_at = $4,
		    processed_at = $4
		WHERE id = $1
	`, matchedPending.ID, conversionValue, currency, processedAt)
	if err != nil {
		return nil, false, fmt.Errorf("failed to update pending reward conversion state: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO bandit_conversion_links (pending_id, transaction_id, linked_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, matchedPending.ID, transactionID, processedAt)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create conversion link: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("failed to commit pending conversion transaction: %w", err)
	}

	convertedAt := processedAt
	matchedPending.Converted = true
	matchedPending.ConversionValue = conversionValue
	matchedPending.ConversionCurrency = currency
	matchedPending.ConvertedAt = &convertedAt
	matchedPending.ProcessedAt = &convertedAt
	return matchedPending, true, nil
}

func (r *PostgresBanditRepository) ProcessExpiredPendingReward(ctx context.Context, pendingID uuid.UUID, processedAt time.Time) (bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to begin expired pending reward transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	pending := &service.PendingReward{}
	err = scanPendingReward(tx.QueryRow(ctx, `
		SELECT id, experiment_id, arm_id, user_id, assigned_at, expires_at, converted,
		       conversion_value, conversion_currency, converted_at, processed_at
		FROM bandit_pending_rewards
		WHERE id = $1
		  AND converted = FALSE
		  AND processed_at IS NULL
		  AND expires_at <= $2
		FOR UPDATE
	`, pendingID, processedAt), pending)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to load expired pending reward: %w", err)
	}

	inserted, err := r.insertConversionEventTx(ctx, tx, &service.ConversionEvent{
		ExperimentID:          pending.ExperimentID,
		ArmID:                 pending.ArmID,
		UserID:                &pending.UserID,
		PendingRewardID:       &pending.ID,
		EventType:             service.ConversionEventTypeExpiredPendingReward,
		OriginalRewardValue:   0,
		NormalizedRewardValue: 0,
		Metadata: map[string]interface{}{
			"source": "postgres_bandit_repository",
		},
		OccurredAt: processedAt,
	})
	if err != nil {
		return false, err
	}
	if !inserted {
		if err := tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("failed to commit duplicate expired pending reward transaction: %w", err)
		}
		return false, nil
	}

	if err := r.applyRewardToArmTx(ctx, tx, pending.ArmID, 0); err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE bandit_pending_rewards
		SET processed_at = $2
		WHERE id = $1
	`, pending.ID, processedAt)
	if err != nil {
		return false, fmt.Errorf("failed to mark expired pending reward processed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit expired pending reward transaction: %w", err)
	}

	return true, nil
}

// GetAssignmentHistory retrieves historical assignments for a user
func (r *PostgresBanditRepository) GetAssignmentHistory(ctx context.Context, userID uuid.UUID, limit int) ([]service.Assignment, error) {
	query := `
		SELECT id, experiment_id, user_id, arm_id, assigned_at, expires_at
		FROM ab_test_assignments
		WHERE user_id = $1
		ORDER BY assigned_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignment history: %w", err)
	}
	defer rows.Close()

	var assignments []service.Assignment
	for rows.Next() {
		var assignment service.Assignment
		if err := rows.Scan(
			&assignment.ID,
			&assignment.ExperimentID,
			&assignment.UserID,
			&assignment.ArmID,
			&assignment.AssignedAt,
			&assignment.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// CleanupExpiredAssignments removes expired assignments older than the specified duration
func (r *PostgresBanditRepository) CleanupExpiredAssignments(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM ab_test_assignments
		WHERE expires_at < NOW() - $1::interval
	`

	result, err := r.pool.Exec(ctx, query, olderThan.String())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired assignments: %w", err)
	}

	count := result.RowsAffected()
	r.logger.Debug("Cleaned up expired assignments", zap.Int64("count", count))

	return count, nil
}

func (r *PostgresBanditRepository) ListWindowMaintenanceExperimentIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id
		FROM ab_tests
		WHERE is_bandit = TRUE
		  AND status IN ('running', 'paused')
		  AND window_type IS NOT NULL
		  AND window_type <> 'none'
		ORDER BY updated_at DESC, id
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query window maintenance experiments: %w", err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan window maintenance experiment id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate window maintenance experiments: %w", err)
	}
	return ids, nil
}

func (r *PostgresBanditRepository) ListObjectiveSyncExperimentIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id
		FROM ab_tests
		WHERE is_bandit = TRUE
		  AND status IN ('running', 'paused', 'completed')
		  AND objective_type IS NOT NULL
		ORDER BY updated_at DESC, id
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query objective sync experiments: %w", err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan objective sync experiment id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate objective sync experiments: %w", err)
	}
	return ids, nil
}

func (r *PostgresBanditRepository) CleanupStaleUserContext(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM bandit_user_context
		WHERE updated_at < NOW() - $1::interval
	`

	result, err := r.pool.Exec(ctx, query, olderThan.String())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup stale user context: %w", err)
	}

	count := result.RowsAffected()
	r.logger.Debug("Cleaned up stale user context", zap.Int64("count", count))
	return count, nil
}

// GetAllArmStatsForExperiment retrieves statistics for all arms in an experiment
func (r *PostgresBanditRepository) GetAllArmStatsForExperiment(ctx context.Context, experimentID uuid.UUID) (map[uuid.UUID]*service.ArmStats, error) {
	query := `
		SELECT s.arm_id, s.alpha, s.beta, s.samples, s.conversions, s.revenue, s.avg_reward, s.updated_at
		FROM ab_test_arm_stats s
		INNER JOIN ab_test_arms a ON a.id = s.arm_id
		WHERE a.experiment_id = $1
	`

	rows, err := r.pool.Query(ctx, query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query arm stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[uuid.UUID]*service.ArmStats)
	for rows.Next() {
		var s service.ArmStats
		if err := rows.Scan(
			&s.ArmID,
			&s.Alpha,
			&s.Beta,
			&s.Samples,
			&s.Conversions,
			&s.Revenue,
			&s.AvgReward,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan arm stats: %w", err)
		}
		stats[s.ArmID] = &s
	}

	return stats, nil
}

// CreateExperiment creates a new experiment
func (r *PostgresBanditRepository) CreateExperiment(ctx context.Context, experiment *Experiment) error {
	query := `
		INSERT INTO ab_tests (id, name, description, status, start_at, end_at, algorithm_type, is_bandit, min_sample_size, confidence_threshold)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		experiment.ID,
		experiment.Name,
		experiment.Description,
		experiment.Status,
		experiment.StartAt,
		experiment.EndAt,
		experiment.AlgorithmType,
		experiment.IsBandit,
		experiment.MinSampleSize,
		experiment.ConfidenceThreshold,
	)

	if err != nil {
		return fmt.Errorf("failed to create experiment: %w", err)
	}

	return nil
}

// GetExperiment retrieves an experiment by ID
func (r *PostgresBanditRepository) GetExperiment(ctx context.Context, experimentID uuid.UUID) (*Experiment, error) {
	query := `
		SELECT id, name, description, status, start_at, end_at, algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence, created_at, updated_at, automation_policy
		FROM ab_tests
		WHERE id = $1
	`

	var experiment Experiment
	var automationPolicyJSON []byte
	err := r.pool.QueryRow(ctx, query, experimentID).Scan(
		&experiment.ID,
		&experiment.Name,
		&experiment.Description,
		&experiment.Status,
		&experiment.StartAt,
		&experiment.EndAt,
		&experiment.AlgorithmType,
		&experiment.IsBandit,
		&experiment.MinSampleSize,
		&experiment.ConfidenceThreshold,
		&experiment.WinnerConfidence,
		&experiment.CreatedAt,
		&experiment.UpdatedAt,
		&automationPolicyJSON,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("experiment not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	experiment.AutomationPolicy = service.DefaultExperimentAutomationPolicy()
	if len(automationPolicyJSON) > 0 {
		var policy service.ExperimentAutomationPolicy
		if err := json.Unmarshal(automationPolicyJSON, &policy); err != nil {
			return nil, fmt.Errorf("failed to decode experiment automation policy: %w", err)
		}
		experiment.AutomationPolicy = service.NormalizeExperimentAutomationPolicy(&policy)
	}

	return &experiment, nil
}

// CreateArm creates a new arm for an experiment
func (r *PostgresBanditRepository) CreateArm(ctx context.Context, arm *service.Arm) error {
	query := `
		INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		arm.ID,
		arm.ExperimentID,
		arm.Name,
		arm.Description,
		arm.IsControl,
		arm.TrafficWeight,
	)

	if err != nil {
		return fmt.Errorf("failed to create arm: %w", err)
	}

	return nil
}

// Experiment represents an A/B test experiment
type Experiment struct {
	ID                  uuid.UUID
	Name                string
	Description         string
	Status              string
	StartAt             *time.Time
	EndAt               *time.Time
	AlgorithmType       *string
	IsBandit            bool
	MinSampleSize       int
	ConfidenceThreshold float64
	WinnerConfidence    *float64
	CreatedAt           time.Time
	UpdatedAt           time.Time
	AutomationPolicy    service.ExperimentAutomationPolicy
	// Advanced bandit fields
	WindowType       *string
	WindowSize       *int
	WindowMinSamples *int
	ObjectiveType    *string
	ObjectiveWeights *map[string]float64
	EnableContextual *bool
	EnableDelayed    *bool
	EnableCurrency   *bool
	ExplorationAlpha *float64
}

// =====================================================
// Advanced Bandit Repository Methods
// =====================================================

// GetExperimentConfig retrieves the experiment configuration
func (r *PostgresBanditRepository) GetExperimentConfig(ctx context.Context, experimentID uuid.UUID) (*service.ExperimentConfig, error) {
	query := `
		SELECT id, objective_type, objective_weights, window_type, window_size, window_min_samples,
		       enable_contextual, enable_delayed, enable_currency, exploration_alpha
		FROM ab_tests
		WHERE id = $1
	`

	var config service.ExperimentConfig
	var objectiveWeightsJSON []byte
	var windowType, windowSize, windowMinSamples interface{}

	err := r.pool.QueryRow(ctx, query, experimentID).Scan(
		&config.ID,
		&config.ObjectiveType,
		&objectiveWeightsJSON,
		&windowType,
		&windowSize,
		&windowMinSamples,
		&config.EnableContextual,
		&config.EnableDelayed,
		&config.EnableCurrency,
		&config.ExplorationAlpha,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("experiment not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get experiment config: %w", err)
	}

	// Parse JSONB for objective weights
	if len(objectiveWeightsJSON) > 0 {
		if err := json.Unmarshal(objectiveWeightsJSON, &config.ObjectiveWeights); err != nil {
			r.logger.Warn("Failed to parse objective weights", zap.Error(err))
		}
	}

	// Build window config if any values are set
	if windowType != nil || windowSize != nil || windowMinSamples != nil {
		config.WindowConfig = &service.WindowConfig{}

		if wt, ok := windowType.(string); ok {
			config.WindowConfig.Type = service.WindowType(wt)
		}
		if ws, ok := windowSize.(int32); ok {
			config.WindowConfig.Size = int(ws)
		}
		if wms, ok := windowMinSamples.(int32); ok {
			config.WindowConfig.MinSamples = int(wms)
		}
	}

	return &config, nil
}

// UpdateObjectiveConfig persists objective configuration fields for an experiment.
func (r *PostgresBanditRepository) UpdateObjectiveConfig(
	ctx context.Context,
	experimentID uuid.UUID,
	objectiveType service.ObjectiveType,
	objectiveWeights map[string]float64,
) error {
	var objectiveWeightsJSON []byte
	var err error
	if objectiveWeights != nil {
		objectiveWeightsJSON, err = json.Marshal(objectiveWeights)
		if err != nil {
			return fmt.Errorf("failed to marshal objective weights: %w", err)
		}
	}

	query := `
		UPDATE ab_tests
		SET objective_type = $2,
		    objective_weights = $3,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, experimentID, objectiveType, objectiveWeightsJSON)
	if err != nil {
		return fmt.Errorf("failed to update objective config: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("experiment not found")
	}

	return nil
}

// GetUserContext retrieves user context for contextual bandits
func (r *PostgresBanditRepository) GetUserContext(ctx context.Context, userID uuid.UUID) (*service.UserContext, error) {
	query := `
		SELECT user_id, country, device, app_version, days_since_install, total_spent, last_purchase_at, updated_at
		FROM bandit_user_context
		WHERE user_id = $1
	`

	var userCtx service.UserContext
	var updatedAt time.Time
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&userCtx.UserID,
		&userCtx.Country,
		&userCtx.Device,
		&userCtx.AppVersion,
		&userCtx.DaysSinceInstall,
		&userCtx.TotalSpent,
		&userCtx.LastPurchaseAt,
		&updatedAt,
	)

	if err == pgx.ErrNoRows {
		// Return empty context if not found
		return &service.UserContext{UserID: userID}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	return &userCtx, nil
}

// SetUserContext saves or updates user context
func (r *PostgresBanditRepository) SetUserContext(ctx context.Context, userCtx *service.UserContext) error {
	query := `
		INSERT INTO bandit_user_context (user_id, country, device, app_version, days_since_install, total_spent, last_purchase_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id)
		DO UPDATE SET
			country = EXCLUDED.country,
			device = EXCLUDED.device,
			app_version = EXCLUDED.app_version,
			days_since_install = EXCLUDED.days_since_install,
			total_spent = EXCLUDED.total_spent,
			last_purchase_at = EXCLUDED.last_purchase_at,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query,
		userCtx.UserID,
		userCtx.Country,
		userCtx.Device,
		userCtx.AppVersion,
		userCtx.DaysSinceInstall,
		userCtx.TotalSpent,
		userCtx.LastPurchaseAt,
	)

	if err != nil {
		return fmt.Errorf("failed to set user context: %w", err)
	}

	return nil
}

// =====================================================
// Objective Stats Methods (implements service.ObjectiveRepository)
// =====================================================

// GetObjectiveStats retrieves objective-specific statistics for an arm
func (r *PostgresBanditRepository) GetObjectiveStats(ctx context.Context, armID uuid.UUID, objectiveType service.ObjectiveType) (*service.ArmObjectiveStats, error) {
	query := `
		SELECT arm_id, objective_type, alpha, beta, samples, conversions, total_revenue, avg_ltv
		FROM bandit_arm_objective_stats
		WHERE arm_id = $1 AND objective_type = $2
	`

	var stats service.ArmObjectiveStats
	err := r.pool.QueryRow(ctx, query, armID, objectiveType).Scan(
		&stats.ArmID,
		&stats.ObjectiveType,
		&stats.Alpha,
		&stats.Beta,
		&stats.Samples,
		&stats.Conversions,
		&stats.TotalRevenue,
		&stats.AvgLTV,
	)

	if err == pgx.ErrNoRows {
		// Return default stats if not found
		return &service.ArmObjectiveStats{
			ArmID:         armID,
			ObjectiveType: objectiveType,
			Alpha:         1.0,
			Beta:          1.0,
			Samples:       0,
			Conversions:   0,
			TotalRevenue:  0,
			AvgLTV:        0,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get objective stats: %w", err)
	}

	return &stats, nil
}

// UpdateObjectiveStats updates objective-specific statistics
func (r *PostgresBanditRepository) UpdateObjectiveStats(ctx context.Context, stats *service.ArmObjectiveStats) error {
	query := `
		INSERT INTO bandit_arm_objective_stats (arm_id, objective_type, alpha, beta, samples, conversions, total_revenue, avg_ltv)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (arm_id, objective_type)
		DO UPDATE SET
			alpha = $3,
			beta = $4,
			samples = $5,
			conversions = $6,
			total_revenue = $7,
			avg_ltv = $8,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query,
		stats.ArmID,
		stats.ObjectiveType,
		stats.Alpha,
		stats.Beta,
		stats.Samples,
		stats.Conversions,
		stats.TotalRevenue,
		stats.AvgLTV,
	)

	if err != nil {
		return fmt.Errorf("failed to update objective stats: %w", err)
	}

	return nil
}

// GetAllObjectiveStats retrieves all objective statistics for an arm
func (r *PostgresBanditRepository) GetAllObjectiveStats(ctx context.Context, armID uuid.UUID) (map[service.ObjectiveType]*service.ArmObjectiveStats, error) {
	query := `
		SELECT arm_id, objective_type, alpha, beta, samples, conversions, total_revenue, avg_ltv
		FROM bandit_arm_objective_stats
		WHERE arm_id = $1
	`

	rows, err := r.pool.Query(ctx, query, armID)
	if err != nil {
		return nil, fmt.Errorf("failed to query objective stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[service.ObjectiveType]*service.ArmObjectiveStats)
	for rows.Next() {
		var s service.ArmObjectiveStats
		if err := rows.Scan(
			&s.ArmID,
			&s.ObjectiveType,
			&s.Alpha,
			&s.Beta,
			&s.Samples,
			&s.Conversions,
			&s.TotalRevenue,
			&s.AvgLTV,
		); err != nil {
			return nil, fmt.Errorf("failed to scan objective stats: %w", err)
		}
		stats[s.ObjectiveType] = &s
	}

	return stats, nil
}

// =====================================================
// Pending Reward Methods (implements service.DelayedRewardRepository)
// =====================================================

// CreatePendingReward creates a new pending reward
func (r *PostgresBanditRepository) CreatePendingReward(ctx context.Context, reward *service.PendingReward) error {
	query := `
		INSERT INTO bandit_pending_rewards (id, experiment_id, arm_id, user_id, assigned_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		reward.ID,
		reward.ExperimentID,
		reward.ArmID,
		reward.UserID,
		reward.AssignedAt,
		reward.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create pending reward: %w", err)
	}

	return nil
}

// GetPendingReward retrieves a pending reward by ID
func (r *PostgresBanditRepository) GetPendingReward(ctx context.Context, id uuid.UUID) (*service.PendingReward, error) {
	query := `
		SELECT id, experiment_id, arm_id, user_id, assigned_at, expires_at, converted,
		       conversion_value, conversion_currency, converted_at, processed_at
		FROM bandit_pending_rewards
		WHERE id = $1
	`

	var reward service.PendingReward
	err := scanPendingReward(r.pool.QueryRow(ctx, query, id), &reward)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("pending reward not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get pending reward: %w", err)
	}

	return &reward, nil
}

// GetPendingRewardsByUser retrieves pending rewards for a user
func (r *PostgresBanditRepository) GetPendingRewardsByUser(ctx context.Context, userID, experimentID uuid.UUID) ([]*service.PendingReward, error) {
	query := `
		SELECT id, experiment_id, arm_id, user_id, assigned_at, expires_at, converted,
		       conversion_value, conversion_currency, converted_at, processed_at
		FROM bandit_pending_rewards
		WHERE user_id = $1 AND ($2::uuid IS NULL OR experiment_id = $2)
		ORDER BY assigned_at DESC
	`

	var experimentParam interface{}
	if experimentID != uuid.Nil {
		experimentParam = experimentID
	}

	rows, err := r.pool.Query(ctx, query, userID, experimentParam)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending rewards: %w", err)
	}
	defer rows.Close()

	var rewards []*service.PendingReward
	for rows.Next() {
		var reward service.PendingReward
		if err := scanPendingReward(rows, &reward); err != nil {
			return nil, fmt.Errorf("failed to scan pending reward: %w", err)
		}
		rewards = append(rewards, &reward)
	}

	return rewards, nil
}

// GetExpiredPendingRewards retrieves expired pending rewards
func (r *PostgresBanditRepository) GetExpiredPendingRewards(ctx context.Context, limit int) ([]*service.PendingReward, error) {
	query := `
		SELECT id, experiment_id, arm_id, user_id, assigned_at, expires_at, converted,
		       conversion_value, conversion_currency, converted_at, processed_at
		FROM bandit_pending_rewards
		WHERE expires_at < NOW() AND converted = FALSE AND processed_at IS NULL
		ORDER BY expires_at ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired rewards: %w", err)
	}
	defer rows.Close()

	var rewards []*service.PendingReward
	for rows.Next() {
		var reward service.PendingReward
		if err := scanPendingReward(rows, &reward); err != nil {
			return nil, fmt.Errorf("failed to scan pending reward: %w", err)
		}
		rewards = append(rewards, &reward)
	}

	return rewards, nil
}

// UpdatePendingReward updates a pending reward
func (r *PostgresBanditRepository) UpdatePendingReward(ctx context.Context, reward *service.PendingReward) error {
	query := `
		UPDATE bandit_pending_rewards
		SET converted = $2,
		    conversion_value = $3,
		    conversion_currency = $4,
		    converted_at = $5,
		    processed_at = $6
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		reward.ID,
		reward.Converted,
		reward.ConversionValue,
		reward.ConversionCurrency,
		reward.ConvertedAt,
		reward.ProcessedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update pending reward: %w", err)
	}

	return nil
}

// LinkConversion links a pending reward to a transaction
func (r *PostgresBanditRepository) LinkConversion(ctx context.Context, link *service.ConversionLink) error {
	query := `
		INSERT INTO bandit_conversion_links (pending_id, transaction_id)
		VALUES ($1, $2)
	`

	_, err := r.pool.Exec(ctx, query, link.PendingID, link.TransactionID)

	if err != nil {
		return fmt.Errorf("failed to link conversion: %w", err)
	}

	return nil
}

func (r *PostgresBanditRepository) applyRewardToArmTx(ctx context.Context, tx pgx.Tx, armID uuid.UUID, reward float64) error {
	stats, err := r.loadArmStatsTx(ctx, tx, armID)
	if err != nil {
		return err
	}

	if reward > 0 {
		stats.Alpha += 1.0
		stats.Conversions++
		stats.Revenue += reward
	} else {
		stats.Beta += 1.0
	}
	stats.Samples++
	if stats.Samples > 0 {
		stats.AvgReward = stats.Revenue / float64(stats.Samples)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (arm_id)
		DO UPDATE SET
			alpha = EXCLUDED.alpha,
			beta = EXCLUDED.beta,
			samples = EXCLUDED.samples,
			conversions = EXCLUDED.conversions,
			revenue = EXCLUDED.revenue,
			avg_reward = EXCLUDED.avg_reward,
			updated_at = NOW()
	`, stats.ArmID, stats.Alpha, stats.Beta, stats.Samples, stats.Conversions, stats.Revenue, stats.AvgReward)
	if err != nil {
		return fmt.Errorf("failed to persist transactional arm stats: %w", err)
	}

	return nil
}

func (r *PostgresBanditRepository) loadArmStatsTx(ctx context.Context, tx pgx.Tx, armID uuid.UUID) (*service.ArmStats, error) {
	stats := &service.ArmStats{}
	err := tx.QueryRow(ctx, `
		SELECT arm_id, alpha, beta, samples, conversions, revenue, avg_reward, updated_at
		FROM ab_test_arm_stats
		WHERE arm_id = $1
		FOR UPDATE
	`, armID).Scan(
		&stats.ArmID,
		&stats.Alpha,
		&stats.Beta,
		&stats.Samples,
		&stats.Conversions,
		&stats.Revenue,
		&stats.AvgReward,
		&stats.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return &service.ArmStats{ArmID: armID, Alpha: 1.0, Beta: 1.0}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load arm stats for update: %w", err)
	}
	return stats, nil
}

func (r *PostgresBanditRepository) insertConversionEventTx(ctx context.Context, tx pgx.Tx, event *service.ConversionEvent) (bool, error) {
	var metadataJSON []byte
	var err error
	if event.Metadata != nil {
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return false, fmt.Errorf("failed to marshal conversion event metadata: %w", err)
		}
	}

	result, err := tx.Exec(ctx, `
		INSERT INTO bandit_conversion_events (
			experiment_id,
			arm_id,
			user_id,
			pending_reward_id,
			transaction_id,
			event_type,
			original_reward_value,
			original_currency,
			normalized_reward_value,
			normalized_currency,
			metadata,
			occurred_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, NULLIF($10, ''), $11, $12)
		ON CONFLICT DO NOTHING
	`,
		event.ExperimentID,
		event.ArmID,
		nullableUUID(event.UserID),
		nullableUUID(event.PendingRewardID),
		nullableUUID(event.TransactionID),
		event.EventType,
		event.OriginalRewardValue,
		event.OriginalCurrency,
		event.NormalizedRewardValue,
		event.NormalizedCurrency,
		metadataJSON,
		normalizeOccurredAt(event.OccurredAt),
	)
	if err != nil {
		return false, fmt.Errorf("failed to insert conversion event: %w", err)
	}
	return result.RowsAffected() > 0, nil
}

func nullableUUID(id *uuid.UUID) interface{} {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return *id
}

func normalizeOccurredAt(occurredAt time.Time) time.Time {
	if occurredAt.IsZero() {
		return time.Now().UTC()
	}
	return occurredAt.UTC()
}

// GetByTransactionID retrieves conversion links by transaction ID
func (r *PostgresBanditRepository) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*service.ConversionLink, error) {
	query := `
		SELECT pending_id, transaction_id, linked_at
		FROM bandit_conversion_links
		WHERE transaction_id = $1
	`

	rows, err := r.pool.Query(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversion links: %w", err)
	}
	defer rows.Close()

	var links []*service.ConversionLink
	for rows.Next() {
		var link service.ConversionLink
		if err := rows.Scan(&link.PendingID, &link.TransactionID, &link.LinkedAt); err != nil {
			return nil, fmt.Errorf("failed to scan conversion link: %w", err)
		}
		links = append(links, &link)
	}

	return links, nil
}
