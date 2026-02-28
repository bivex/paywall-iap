package repository

import (
	"context"
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

// NewPostgresBanditRepository creates a new PostgreSQL-backed bandit repository
func NewPostgresBanditRepository(pool *pgxpool.Pool, logger *zap.Logger) *PostgresBanditRepository {
	return &PostgresBanditRepository{
		pool:   pool,
		logger: logger,
	}
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
		// Return default stats if not found
		return &service.ArmStats{
			ArmID:     armID,
			Alpha:     1.0, // Uniform prior
			Beta:      1.0,
			Samples:   0,
			Conversions: 0,
			Revenue:   0,
			AvgReward: 0,
			UpdatedAt: time.Now(),
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
	query := `
		INSERT INTO ab_test_assignments (id, experiment_id, user_id, arm_id, assigned_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (experiment_id, user_id) WHERE expires_at > NOW()
		DO UPDATE SET
			arm_id = $4,
			assigned_at = $5,
			expires_at = $6
	`

	_, err := r.pool.Exec(ctx, query,
		assignment.ID,
		assignment.ExperimentID,
		assignment.UserID,
		assignment.ArmID,
		assignment.AssignedAt,
		assignment.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

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
		return nil, fmt.Errorf("no active assignment found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	return &assignment, nil
}

// SaveConversion records a conversion event for an arm
func (r *PostgresBanditRepository) SaveConversion(ctx context.Context, experimentID, armID, userID uuid.UUID, amount float64) error {
	// This would typically insert into a conversions table for analytics
	// For now, conversions are tracked via UpdateArmStats
	// TODO: Create ab_test_conversions table if detailed tracking is needed

	r.logger.Debug("Conversion saved",
		zap.String("experiment_id", experimentID.String()),
		zap.String("arm_id", armID.String()),
		zap.String("user_id", userID.String()),
		zap.Float64("amount", amount),
	)

	return nil
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
		SELECT id, name, description, status, start_at, end_at, algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence, created_at, updated_at
		FROM ab_tests
		WHERE id = $1
	`

	var experiment Experiment
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
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("experiment not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
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
	ID                    uuid.UUID
	Name                  string
	Description           string
	Status                string
	StartAt               *time.Time
	EndAt                 *time.Time
	AlgorithmType         *string
	IsBandit              bool
	MinSampleSize         int
	ConfidenceThreshold   float64
	WinnerConfidence      *float64
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
