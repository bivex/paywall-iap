package handlers

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

type AdminExperimentArm struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	IsControl     bool      `json:"is_control"`
	TrafficWeight float64   `json:"traffic_weight"`
	Samples       int       `json:"samples"`
	Conversions   int       `json:"conversions"`
	Revenue       float64   `json:"revenue"`
	AvgReward     float64   `json:"avg_reward"`
}

type AdminExperiment struct {
	ID                         uuid.UUID            `json:"id"`
	Name                       string               `json:"name"`
	Description                string               `json:"description"`
	Status                     string               `json:"status"`
	AlgorithmType              *string              `json:"algorithm_type"`
	IsBandit                   bool                 `json:"is_bandit"`
	MinSampleSize              int                  `json:"min_sample_size"`
	ConfidenceThresholdPercent float64              `json:"confidence_threshold_percent"`
	WinnerConfidencePercent    *float64             `json:"winner_confidence_percent"`
	StartAt                    *time.Time           `json:"start_at"`
	EndAt                      *time.Time           `json:"end_at"`
	CreatedAt                  time.Time            `json:"created_at"`
	UpdatedAt                  time.Time            `json:"updated_at"`
	ArmCount                   int                  `json:"arm_count"`
	TotalAssignments           int                  `json:"total_assignments"`
	ActiveAssignments          int                  `json:"active_assignments"`
	TotalSamples               int                  `json:"total_samples"`
	TotalConversions           int                  `json:"total_conversions"`
	TotalRevenue               float64              `json:"total_revenue"`
	Arms                       []AdminExperimentArm `json:"arms"`
}

type createAdminExperimentArmRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	IsControl     bool    `json:"is_control"`
	TrafficWeight float64 `json:"traffic_weight"`
}

type createAdminExperimentRequest struct {
	Name                       string                            `json:"name"`
	Description                string                            `json:"description"`
	Status                     string                            `json:"status"`
	AlgorithmType              string                            `json:"algorithm_type"`
	IsBandit                   bool                              `json:"is_bandit"`
	MinSampleSize              int                               `json:"min_sample_size"`
	ConfidenceThresholdPercent float64                           `json:"confidence_threshold_percent"`
	StartAt                    *time.Time                        `json:"start_at"`
	EndAt                      *time.Time                        `json:"end_at"`
	Arms                       []createAdminExperimentArmRequest `json:"arms"`
}

const adminExperimentSelectBase = `
		SELECT e.id,
		       e.name,
		       e.description,
		       e.status,
		       e.algorithm_type,
		       e.is_bandit,
		       e.min_sample_size,
		       e.confidence_threshold::double precision,
		       e.winner_confidence::double precision,
		       e.start_at,
		       e.end_at,
		       e.created_at,
		       e.updated_at,
		       (SELECT COUNT(*)::int FROM ab_test_arms a WHERE a.experiment_id = e.id) AS arm_count,`

const adminExperimentSelectStatsOnly = `
		       0::int AS total_assignments,
		       0::int AS active_assignments,
		       COALESCE((SELECT SUM(s.samples)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_samples,
		       COALESCE((SELECT SUM(s.conversions)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_conversions,
		       COALESCE((SELECT SUM(s.revenue)::double precision FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_revenue
		FROM ab_tests e`

const adminExperimentSelectWithAssignments = `
		       (SELECT COUNT(*)::int FROM ab_test_assignments ass WHERE ass.experiment_id = e.id) AS total_assignments,
		       (SELECT COUNT(*)::int FROM ab_test_assignments ass WHERE ass.experiment_id = e.id AND ass.expires_at > NOW()) AS active_assignments,
		       COALESCE((SELECT SUM(s.samples)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_samples,
		       COALESCE((SELECT SUM(s.conversions)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_conversions,
		       COALESCE((SELECT SUM(s.revenue)::double precision FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_revenue
		FROM ab_tests e`

func normalizeCreateAdminExperimentRequest(req createAdminExperimentRequest) createAdminExperimentRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	req.AlgorithmType = strings.ToLower(strings.TrimSpace(req.AlgorithmType))
	for index := range req.Arms {
		req.Arms[index].Name = strings.TrimSpace(req.Arms[index].Name)
		req.Arms[index].Description = strings.TrimSpace(req.Arms[index].Description)
	}
	return req
}

func validateCreateAdminExperimentRequest(req createAdminExperimentRequest) string {
	if req.Name == "" {
		return "Experiment name is required"
	}
	switch req.Status {
	case "draft", "running", "paused", "completed":
	default:
		return "Status must be draft, running, paused, or completed"
	}
	if req.MinSampleSize <= 0 {
		return "Minimum sample size must be greater than zero"
	}
	if req.ConfidenceThresholdPercent <= 0 || req.ConfidenceThresholdPercent > 100 {
		return "Confidence threshold must be between 0 and 100"
	}
	if req.StartAt != nil && req.EndAt != nil && req.EndAt.Before(*req.StartAt) {
		return "End time must be after start time"
	}
	if len(req.Arms) < 2 {
		return "At least two experiment arms are required"
	}
	controlCount := 0
	for _, arm := range req.Arms {
		if arm.Name == "" {
			return "Every experiment arm must have a name"
		}
		if arm.TrafficWeight <= 0 {
			return "Traffic weight must be greater than zero"
		}
		if arm.IsControl {
			controlCount++
		}
	}
	if controlCount != 1 {
		return "Exactly one control arm is required"
	}
	if req.IsBandit {
		switch req.AlgorithmType {
		case "thompson_sampling", "ucb", "epsilon_greedy":
		default:
			return "Algorithm type must be thompson_sampling, ucb, or epsilon_greedy"
		}
	}
	return ""
}

func scanAdminExperiment(scanner interface{ Scan(dest ...any) error }) (AdminExperiment, error) {
	var experiment AdminExperiment
	var description sql.NullString
	var algorithmType sql.NullString
	var winnerConfidence sql.NullFloat64
	var startAt sql.NullTime
	var endAt sql.NullTime
	var confidenceThreshold float64

	err := scanner.Scan(
		&experiment.ID,
		&experiment.Name,
		&description,
		&experiment.Status,
		&algorithmType,
		&experiment.IsBandit,
		&experiment.MinSampleSize,
		&confidenceThreshold,
		&winnerConfidence,
		&startAt,
		&endAt,
		&experiment.CreatedAt,
		&experiment.UpdatedAt,
		&experiment.ArmCount,
		&experiment.TotalAssignments,
		&experiment.ActiveAssignments,
		&experiment.TotalSamples,
		&experiment.TotalConversions,
		&experiment.TotalRevenue,
	)
	if err != nil {
		return AdminExperiment{}, err
	}

	if description.Valid {
		experiment.Description = description.String
	}
	if algorithmType.Valid {
		experiment.AlgorithmType = &algorithmType.String
	}
	experiment.ConfidenceThresholdPercent = confidenceThreshold * 100
	if winnerConfidence.Valid {
		value := winnerConfidence.Float64 * 100
		experiment.WinnerConfidencePercent = &value
	}
	if startAt.Valid {
		value := startAt.Time
		experiment.StartAt = &value
	}
	if endAt.Valid {
		value := endAt.Time
		experiment.EndAt = &value
	}

	return experiment, nil
}

func scanAdminExperimentArm(scanner interface{ Scan(dest ...any) error }) (AdminExperimentArm, error) {
	var arm AdminExperimentArm
	var description sql.NullString
	err := scanner.Scan(
		&arm.ID,
		&arm.Name,
		&description,
		&arm.IsControl,
		&arm.TrafficWeight,
		&arm.Samples,
		&arm.Conversions,
		&arm.Revenue,
		&arm.AvgReward,
	)
	if err != nil {
		return AdminExperimentArm{}, err
	}
	if description.Valid {
		arm.Description = description.String
	}
	return arm, nil
}

func (h *AdminHandler) hasAssignmentTable(c *gin.Context) bool {
	var exists bool
	err := h.dbPool.QueryRow(c.Request.Context(), `
		SELECT to_regclass('public.ab_test_assignments') IS NOT NULL`).Scan(&exists)
	return err == nil && exists
}

func adminExperimentListQuery(withAssignments bool) string {
	if withAssignments {
		return adminExperimentSelectBase + adminExperimentSelectWithAssignments + `
		ORDER BY e.created_at DESC`
	}
	return adminExperimentSelectBase + adminExperimentSelectStatsOnly + `
		ORDER BY e.created_at DESC`
}

func adminExperimentByIDQuery(withAssignments bool) string {
	if withAssignments {
		return adminExperimentSelectBase + adminExperimentSelectWithAssignments + `
		WHERE e.id = $1`
	}
	return adminExperimentSelectBase + adminExperimentSelectStatsOnly + `
		WHERE e.id = $1`
}

func (h *AdminHandler) listExperimentArms(ctx *gin.Context, experimentID uuid.UUID) ([]AdminExperimentArm, error) {
	rows, err := h.dbPool.Query(ctx.Request.Context(), `
		SELECT a.id,
		       a.name,
		       a.description,
		       a.is_control,
		       a.traffic_weight::double precision,
		       COALESCE(s.samples, 0)::int,
		       COALESCE(s.conversions, 0)::int,
		       COALESCE(s.revenue, 0)::double precision,
		       COALESCE(s.avg_reward, 0)::double precision
		FROM ab_test_arms a
		LEFT JOIN ab_test_arm_stats s ON s.arm_id = a.id
		WHERE a.experiment_id = $1
		ORDER BY a.is_control DESC, a.created_at ASC, a.name ASC`, experimentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	arms := make([]AdminExperimentArm, 0)
	for rows.Next() {
		arm, err := scanAdminExperimentArm(rows)
		if err != nil {
			return nil, err
		}
		arms = append(arms, arm)
	}
	return arms, rows.Err()
}

func (h *AdminHandler) getAdminExperimentByID(c *gin.Context, experimentID uuid.UUID) (AdminExperiment, error) {
	experiment, err := scanAdminExperiment(h.dbPool.QueryRow(c.Request.Context(), adminExperimentByIDQuery(h.hasAssignmentTable(c)), experimentID))
	if err != nil {
		return AdminExperiment{}, err
	}
	arms, err := h.listExperimentArms(c, experimentID)
	if err != nil {
		return AdminExperiment{}, err
	}
	experiment.Arms = arms
	return experiment, nil
}

func (h *AdminHandler) ListAdminExperiments(c *gin.Context) {
	rows, err := h.dbPool.Query(c.Request.Context(), adminExperimentListQuery(h.hasAssignmentTable(c)))
	if err != nil {
		response.InternalError(c, "Failed to load experiments")
		return
	}
	defer rows.Close()

	experiments := make([]AdminExperiment, 0)
	for rows.Next() {
		experiment, err := scanAdminExperiment(rows)
		if err != nil {
			response.InternalError(c, "Failed to load experiments")
			return
		}
		arms, err := h.listExperimentArms(c, experiment.ID)
		if err != nil {
			response.InternalError(c, "Failed to load experiments")
			return
		}
		experiment.Arms = arms
		experiments = append(experiments, experiment)
	}
	if rows.Err() != nil {
		response.InternalError(c, "Failed to load experiments")
		return
	}

	response.OK(c, experiments)
}

func (h *AdminHandler) CreateAdminExperiment(c *gin.Context) {
	var req createAdminExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid experiment payload")
		return
	}
	req = normalizeCreateAdminExperimentRequest(req)
	if message := validateCreateAdminExperimentRequest(req); message != "" {
		response.UnprocessableEntity(c, message)
		return
	}

	ctx := c.Request.Context()
	tx, err := h.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		response.InternalError(c, "Failed to start experiment transaction")
		return
	}
	defer tx.Rollback(ctx)

	experimentID := uuid.New()
	var algorithmType interface{}
	if req.IsBandit {
		algorithmType = req.AlgorithmType
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO ab_tests (
			id, name, description, status, start_at, end_at,
			algorithm_type, is_bandit, min_sample_size, confidence_threshold
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		experimentID,
		req.Name,
		req.Description,
		req.Status,
		req.StartAt,
		req.EndAt,
		algorithmType,
		req.IsBandit,
		req.MinSampleSize,
		req.ConfidenceThresholdPercent/100,
	)
	if err != nil {
		response.InternalError(c, "Failed to create experiment")
		return
	}

	for _, arm := range req.Arms {
		_, err = tx.Exec(ctx, `
			INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(),
			experimentID,
			arm.Name,
			arm.Description,
			arm.IsControl,
			arm.TrafficWeight,
		)
		if err != nil {
			response.InternalError(c, "Failed to create experiment arms")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		response.InternalError(c, "Failed to commit experiment")
		return
	}

	experiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load created experiment")
		return
	}

	response.Created(c, experiment)
}
