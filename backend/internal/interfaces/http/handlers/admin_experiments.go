package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bivex/paywall-iap/internal/domain/service"
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
	ID                         uuid.UUID                          `json:"id"`
	Name                       string                             `json:"name"`
	Description                string                             `json:"description"`
	Status                     string                             `json:"status"`
	AlgorithmType              *string                            `json:"algorithm_type"`
	IsBandit                   bool                               `json:"is_bandit"`
	MinSampleSize              int                                `json:"min_sample_size"`
	ConfidenceThresholdPercent float64                            `json:"confidence_threshold_percent"`
	WinnerConfidencePercent    *float64                           `json:"winner_confidence_percent"`
	WinnerRecommendation       *service.WinnerRecommendation      `json:"winner_recommendation,omitempty"`
	StartAt                    *time.Time                         `json:"start_at"`
	EndAt                      *time.Time                         `json:"end_at"`
	AutomationPolicy           service.ExperimentAutomationPolicy `json:"automation_policy"`
	LatestLifecycleAudit       *AdminExperimentLifecycleAudit     `json:"latest_lifecycle_audit,omitempty"`
	CreatedAt                  time.Time                          `json:"created_at"`
	UpdatedAt                  time.Time                          `json:"updated_at"`
	ArmCount                   int                                `json:"arm_count"`
	TotalAssignments           int                                `json:"total_assignments"`
	ActiveAssignments          int                                `json:"active_assignments"`
	TotalSamples               int                                `json:"total_samples"`
	TotalConversions           int                                `json:"total_conversions"`
	TotalRevenue               float64                            `json:"total_revenue"`
	Arms                       []AdminExperimentArm               `json:"arms"`
}

type AdminExperimentLifecycleAudit struct {
	ActorType      string                 `json:"actor_type"`
	Source         string                 `json:"source"`
	Action         string                 `json:"action"`
	FromStatus     string                 `json:"from_status"`
	ToStatus       string                 `json:"to_status"`
	IdempotencyKey *string                `json:"idempotency_key,omitempty"`
	Details        map[string]interface{} `json:"details,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

type AdminExperimentWinnerRecommendationAudit struct {
	Source                     string                 `json:"source"`
	Recommended                bool                   `json:"recommended"`
	Reason                     string                 `json:"reason"`
	WinningArmID               *uuid.UUID             `json:"winning_arm_id,omitempty"`
	ConfidencePercent          *float64               `json:"confidence_percent,omitempty"`
	ConfidenceThresholdPercent float64                `json:"confidence_threshold_percent"`
	ObservedSamples            int                    `json:"observed_samples"`
	MinSampleSize              int                    `json:"min_sample_size"`
	Details                    map[string]interface{} `json:"details,omitempty"`
	OccurredAt                 time.Time              `json:"occurred_at"`
}

func scanAdminExperimentLifecycleAudit(scanner interface{ Scan(dest ...any) error }) (AdminExperimentLifecycleAudit, error) {
	var audit AdminExperimentLifecycleAudit
	var idempotencyKey sql.NullString
	var detailsJSON []byte

	err := scanner.Scan(
		&audit.ActorType,
		&audit.Source,
		&audit.Action,
		&audit.FromStatus,
		&audit.ToStatus,
		&idempotencyKey,
		&detailsJSON,
		&audit.CreatedAt,
	)
	if err != nil {
		return AdminExperimentLifecycleAudit{}, err
	}
	if idempotencyKey.Valid {
		value := idempotencyKey.String
		audit.IdempotencyKey = &value
	}
	if len(detailsJSON) > 0 {
		if err := json.Unmarshal(detailsJSON, &audit.Details); err != nil {
			return AdminExperimentLifecycleAudit{}, err
		}
	}
	return audit, nil
}

func scanAdminExperimentWinnerRecommendationAudit(scanner interface{ Scan(dest ...any) error }) (AdminExperimentWinnerRecommendationAudit, error) {
	var audit AdminExperimentWinnerRecommendationAudit
	var winningArmID uuid.NullUUID
	var confidencePercent sql.NullFloat64
	var detailsJSON []byte

	err := scanner.Scan(
		&audit.Source,
		&audit.Recommended,
		&audit.Reason,
		&winningArmID,
		&confidencePercent,
		&audit.ConfidenceThresholdPercent,
		&audit.ObservedSamples,
		&audit.MinSampleSize,
		&detailsJSON,
		&audit.OccurredAt,
	)
	if err != nil {
		return AdminExperimentWinnerRecommendationAudit{}, err
	}
	if winningArmID.Valid {
		value := winningArmID.UUID
		audit.WinningArmID = &value
	}
	if confidencePercent.Valid {
		value := confidencePercent.Float64
		audit.ConfidencePercent = &value
	}
	if len(detailsJSON) > 0 {
		if err := json.Unmarshal(detailsJSON, &audit.Details); err != nil {
			return AdminExperimentWinnerRecommendationAudit{}, err
		}
	}
	return audit, nil
}

type createAdminExperimentArmRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	IsControl     bool    `json:"is_control"`
	TrafficWeight float64 `json:"traffic_weight"`
}

type createAdminExperimentRequest struct {
	Name                       string                              `json:"name"`
	Description                string                              `json:"description"`
	Status                     string                              `json:"status"`
	AlgorithmType              string                              `json:"algorithm_type"`
	IsBandit                   bool                                `json:"is_bandit"`
	MinSampleSize              int                                 `json:"min_sample_size"`
	ConfidenceThresholdPercent float64                             `json:"confidence_threshold_percent"`
	StartAt                    *time.Time                          `json:"start_at"`
	EndAt                      *time.Time                          `json:"end_at"`
	AutomationPolicy           *service.ExperimentAutomationPolicy `json:"automation_policy,omitempty"`
	Arms                       []createAdminExperimentArmRequest   `json:"arms"`
}

type updateAdminExperimentRequest struct {
	Name                       string                              `json:"name"`
	Description                string                              `json:"description"`
	AlgorithmType              string                              `json:"algorithm_type"`
	IsBandit                   bool                                `json:"is_bandit"`
	MinSampleSize              int                                 `json:"min_sample_size"`
	ConfidenceThresholdPercent float64                             `json:"confidence_threshold_percent"`
	StartAt                    *time.Time                          `json:"start_at"`
	EndAt                      *time.Time                          `json:"end_at"`
	AutomationPolicy           *service.ExperimentAutomationPolicy `json:"automation_policy,omitempty"`
}

type lockAdminExperimentRequest struct {
	LockedUntil *time.Time `json:"locked_until"`
	Reason      string     `json:"reason"`
}

type repairAdminExperimentResponse struct {
	Experiment AdminExperiment                  `json:"experiment"`
	Summary    *service.ExperimentRepairSummary `json:"summary"`
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
		       e.automation_policy,
		       latest_lifecycle.actor_type,
		       latest_lifecycle.source,
		       latest_lifecycle.action,
		       latest_lifecycle.from_status,
		       latest_lifecycle.to_status,
		       latest_lifecycle.idempotency_key,
		       latest_lifecycle.details,
		       latest_lifecycle.created_at,
		       e.created_at,
		       e.updated_at,
		       (SELECT COUNT(*)::int FROM ab_test_arms a WHERE a.experiment_id = e.id) AS arm_count,`

const adminExperimentSelectStatsOnly = `
		       0::int AS total_assignments,
		       0::int AS active_assignments,
		       COALESCE((SELECT SUM(s.samples)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_samples,
		       COALESCE((SELECT SUM(s.conversions)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_conversions,
		       COALESCE((SELECT SUM(s.revenue)::double precision FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_revenue
		FROM ab_tests e
		LEFT JOIN LATERAL (
			SELECT actor_type, source, action, from_status, to_status, idempotency_key, details, created_at
			FROM experiment_lifecycle_audit_log l
			WHERE l.experiment_id = e.id
			ORDER BY l.created_at DESC
			LIMIT 1
		) latest_lifecycle ON true`

const adminExperimentSelectWithAssignments = `
		       (SELECT COUNT(*)::int FROM ab_test_assignments ass WHERE ass.experiment_id = e.id) AS total_assignments,
		       (SELECT COUNT(*)::int FROM ab_test_assignments ass WHERE ass.experiment_id = e.id AND ass.expires_at > NOW()) AS active_assignments,
		       COALESCE((SELECT SUM(s.samples)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_samples,
		       COALESCE((SELECT SUM(s.conversions)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_conversions,
		       COALESCE((SELECT SUM(s.revenue)::double precision FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_revenue
		FROM ab_tests e
		LEFT JOIN LATERAL (
			SELECT actor_type, source, action, from_status, to_status, idempotency_key, details, created_at
			FROM experiment_lifecycle_audit_log l
			WHERE l.experiment_id = e.id
			ORDER BY l.created_at DESC
			LIMIT 1
		) latest_lifecycle ON true`

func normalizeCreateAdminExperimentRequest(req createAdminExperimentRequest) createAdminExperimentRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	req.AlgorithmType = strings.ToLower(strings.TrimSpace(req.AlgorithmType))
	for index := range req.Arms {
		req.Arms[index].Name = strings.TrimSpace(req.Arms[index].Name)
		req.Arms[index].Description = strings.TrimSpace(req.Arms[index].Description)
	}
	normalizedPolicy := service.NormalizeExperimentAutomationPolicy(req.AutomationPolicy)
	req.AutomationPolicy = &normalizedPolicy
	return req
}

func normalizeUpdateAdminExperimentRequest(req updateAdminExperimentRequest) updateAdminExperimentRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.AlgorithmType = strings.ToLower(strings.TrimSpace(req.AlgorithmType))
	normalizedPolicy := service.NormalizeExperimentAutomationPolicy(req.AutomationPolicy)
	req.AutomationPolicy = &normalizedPolicy
	return req
}

func normalizeLockAdminExperimentRequest(req lockAdminExperimentRequest) lockAdminExperimentRequest {
	req.Reason = strings.TrimSpace(req.Reason)
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

func validateUpdateAdminExperimentRequest(req updateAdminExperimentRequest) string {
	if req.Name == "" {
		return "Experiment name is required"
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
	var automationPolicyJSON []byte
	var latestActorType sql.NullString
	var latestSource sql.NullString
	var latestAction sql.NullString
	var latestFromStatus sql.NullString
	var latestToStatus sql.NullString
	var latestIdempotencyKey sql.NullString
	var latestDetailsJSON []byte
	var latestCreatedAt sql.NullTime
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
		&automationPolicyJSON,
		&latestActorType,
		&latestSource,
		&latestAction,
		&latestFromStatus,
		&latestToStatus,
		&latestIdempotencyKey,
		&latestDetailsJSON,
		&latestCreatedAt,
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
	experiment.AutomationPolicy = service.DefaultExperimentAutomationPolicy()
	if len(automationPolicyJSON) > 0 {
		var policy service.ExperimentAutomationPolicy
		if err := json.Unmarshal(automationPolicyJSON, &policy); err != nil {
			return AdminExperiment{}, err
		}
		experiment.AutomationPolicy = service.NormalizeExperimentAutomationPolicy(&policy)
	}
	if latestCreatedAt.Valid {
		audit := &AdminExperimentLifecycleAudit{
			ActorType:  latestActorType.String,
			Source:     latestSource.String,
			Action:     latestAction.String,
			FromStatus: latestFromStatus.String,
			ToStatus:   latestToStatus.String,
			CreatedAt:  latestCreatedAt.Time,
		}
		if latestIdempotencyKey.Valid {
			value := latestIdempotencyKey.String
			audit.IdempotencyKey = &value
		}
		if len(latestDetailsJSON) > 0 {
			if err := json.Unmarshal(latestDetailsJSON, &audit.Details); err != nil {
				return AdminExperiment{}, err
			}
		}
		experiment.LatestLifecycleAudit = audit
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
	if err := h.enrichWinnerRecommendation(c.Request.Context(), &experiment, "admin_experiments_detail"); err != nil {
		return AdminExperiment{}, err
	}
	return experiment, nil
}

func (h *AdminHandler) enrichWinnerRecommendation(ctx context.Context, experiment *AdminExperiment, source ...string) error {
	if experiment == nil || h.winnerRecommendationService == nil {
		return nil
	}
	recommendationSource := "admin_experiments_api"
	if len(source) > 0 && strings.TrimSpace(source[0]) != "" {
		recommendationSource = strings.TrimSpace(source[0])
	}

	var winnerConfidence *float64
	if experiment.WinnerConfidencePercent != nil {
		value := *experiment.WinnerConfidencePercent / 100
		winnerConfidence = &value
	}

	input := service.ExperimentWinnerRecommendationInput{
		ExperimentID:        experiment.ID,
		Source:              recommendationSource,
		Status:              experiment.Status,
		IsBandit:            experiment.IsBandit,
		MinSampleSize:       experiment.MinSampleSize,
		TotalSamples:        experiment.TotalSamples,
		ConfidenceThreshold: experiment.ConfidenceThresholdPercent / 100,
		WinnerConfidence:    winnerConfidence,
		Arms:                make([]service.ExperimentWinnerRecommendationArm, 0, len(experiment.Arms)),
	}
	for _, arm := range experiment.Arms {
		input.Arms = append(input.Arms, service.ExperimentWinnerRecommendationArm{
			ID:        arm.ID,
			Name:      arm.Name,
			IsControl: arm.IsControl,
		})
	}

	recommendation, err := h.winnerRecommendationService.Recommend(ctx, input)
	if err != nil {
		return err
	}
	experiment.WinnerRecommendation = recommendation
	return nil
}

func (h *AdminHandler) listExperimentWinnerRecommendationAuditHistory(ctx *gin.Context, experimentID uuid.UUID) ([]AdminExperimentWinnerRecommendationAudit, error) {
	rows, err := h.dbPool.Query(ctx.Request.Context(), `
		SELECT source,
		       recommended,
		       reason,
		       winning_arm_id,
		       confidence_percent,
		       confidence_threshold_percent,
		       observed_samples,
		       min_sample_size,
		       details,
		       occurred_at
		FROM experiment_winner_recommendation_log
		WHERE experiment_id = $1
		ORDER BY occurred_at DESC, id DESC`, experimentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]AdminExperimentWinnerRecommendationAudit, 0)
	for rows.Next() {
		audit, err := scanAdminExperimentWinnerRecommendationAudit(rows)
		if err != nil {
			return nil, err
		}
		history = append(history, audit)
	}
	return history, rows.Err()
}

func (h *AdminHandler) listExperimentLifecycleAuditHistory(ctx *gin.Context, experimentID uuid.UUID) ([]AdminExperimentLifecycleAudit, error) {
	rows, err := h.dbPool.Query(ctx.Request.Context(), `
		SELECT actor_type,
		       source,
		       action,
		       from_status,
		       to_status,
		       idempotency_key,
		       details,
		       created_at
		FROM experiment_lifecycle_audit_log
		WHERE experiment_id = $1
		ORDER BY created_at DESC, id DESC`, experimentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]AdminExperimentLifecycleAudit, 0)
	for rows.Next() {
		audit, err := scanAdminExperimentLifecycleAudit(rows)
		if err != nil {
			return nil, err
		}
		history = append(history, audit)
	}
	return history, rows.Err()
}

func (h *AdminHandler) GetAdminExperimentLifecycleAuditHistory(c *gin.Context) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}

	if err := h.dbPool.QueryRow(c.Request.Context(), `SELECT 1 FROM ab_tests WHERE id = $1`, experimentID).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "Experiment not found")
			return
		}
		response.InternalError(c, "Failed to load experiment")
		return
	}

	history, err := h.listExperimentLifecycleAuditHistory(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load experiment lifecycle audit history")
		return
	}

	response.OK(c, history)
}

func (h *AdminHandler) GetAdminExperimentWinnerRecommendationAuditHistory(c *gin.Context) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}

	if err := h.dbPool.QueryRow(c.Request.Context(), `SELECT 1 FROM ab_tests WHERE id = $1`, experimentID).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "Experiment not found")
			return
		}
		response.InternalError(c, "Failed to load experiment")
		return
	}

	history, err := h.listExperimentWinnerRecommendationAuditHistory(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load experiment winner recommendation history")
		return
	}

	response.OK(c, history)
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
			if err := h.enrichWinnerRecommendation(c.Request.Context(), &experiment, "admin_experiments_list"); err != nil {
			response.InternalError(c, "Failed to load experiments")
			return
		}
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
	automationPolicyJSON, err := json.Marshal(req.AutomationPolicy)
	if err != nil {
		response.InternalError(c, "Failed to encode experiment automation policy")
		return
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO ab_tests (
			id, name, description, status, start_at, end_at,
			algorithm_type, is_bandit, min_sample_size, confidence_threshold, automation_policy
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
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
		automationPolicyJSON,
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

func (h *AdminHandler) UpdateAdminExperiment(c *gin.Context) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}

	var req updateAdminExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid experiment payload")
		return
	}
	req = normalizeUpdateAdminExperimentRequest(req)
	if message := validateUpdateAdminExperimentRequest(req); message != "" {
		response.UnprocessableEntity(c, message)
		return
	}

	experiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "Experiment not found")
			return
		}
		response.InternalError(c, "Failed to load experiment")
		return
	}
	if experiment.Status != "draft" {
		response.UnprocessableEntity(c, "Only draft experiments can be edited")
		return
	}

	var algorithmType *string
	if req.IsBandit {
		algorithmType = &req.AlgorithmType
	}

	if h.experimentAdminService == nil {
		response.InternalError(c, "Experiment service is unavailable")
		return
	}

	err = h.experimentAdminService.UpdateDraftExperiment(c.Request.Context(), experimentID, service.UpdateExperimentInput{
		Name:                req.Name,
		Description:         req.Description,
		AlgorithmType:       algorithmType,
		IsBandit:            req.IsBandit,
		MinSampleSize:       req.MinSampleSize,
		ConfidenceThreshold: req.ConfidenceThresholdPercent / 100,
		StartAt:             req.StartAt,
		EndAt:               req.EndAt,
		AutomationPolicy:    service.NormalizeExperimentAutomationPolicy(req.AutomationPolicy),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		case errors.Is(err, service.ErrExperimentNotEditable):
			response.UnprocessableEntity(c, "Only draft experiments can be edited")
		default:
			response.InternalError(c, "Failed to update experiment")
		}
		return
	}

	updatedExperiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load updated experiment")
		return
	}

	response.OK(c, updatedExperiment)
}

func (h *AdminHandler) PauseAdminExperiment(c *gin.Context) {
	h.updateAdminExperimentStatus(c, "paused")
}

func (h *AdminHandler) ResumeAdminExperiment(c *gin.Context) {
	h.updateAdminExperimentStatus(c, "running")
}

func (h *AdminHandler) CompleteAdminExperiment(c *gin.Context) {
	h.updateAdminExperimentStatus(c, "completed")
}

func (h *AdminHandler) LockAdminExperiment(c *gin.Context) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}
	if h.experimentAdminService == nil {
		response.InternalError(c, "Experiment service is unavailable")
		return
	}

	var req lockAdminExperimentRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request body")
			return
		}
	}
	req = normalizeLockAdminExperimentRequest(req)
	if req.LockedUntil != nil && !req.LockedUntil.After(time.Now().UTC()) {
		response.UnprocessableEntity(c, "locked_until must be in the future")
		return
	}

	var adminID *uuid.UUID
	if value, ok := c.Get("admin_id"); ok {
		if id, ok := value.(uuid.UUID); ok {
			adminID = &id
		}
	}

	err = h.experimentAdminService.LockExperimentAutomation(c.Request.Context(), experimentID, service.ExperimentLockInput{
		LockedUntil: req.LockedUntil,
		LockedBy:    adminID,
		LockReason:  req.Reason,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		default:
			response.InternalError(c, "Failed to lock experiment automation")
		}
		return
	}

	updatedExperiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load updated experiment")
		return
	}

	response.OK(c, updatedExperiment)
}

func (h *AdminHandler) UnlockAdminExperiment(c *gin.Context) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}
	if h.experimentAdminService == nil {
		response.InternalError(c, "Experiment service is unavailable")
		return
	}

	err = h.experimentAdminService.UnlockExperimentAutomation(c.Request.Context(), experimentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		default:
			response.InternalError(c, "Failed to unlock experiment automation")
		}
		return
	}

	updatedExperiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load updated experiment")
		return
	}

	response.OK(c, updatedExperiment)
}

func (h *AdminHandler) RepairAdminExperiment(c *gin.Context) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}
	if h.experimentRepairService == nil {
		response.InternalError(c, "Experiment repair service is unavailable")
		return
	}

	summary, err := h.experimentRepairService.RepairExperiment(c.Request.Context(), experimentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		default:
			response.InternalError(c, "Failed to repair experiment derived state")
		}
		return
	}

	updatedExperiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load updated experiment")
		return
	}

	response.OK(c, repairAdminExperimentResponse{Experiment: updatedExperiment, Summary: summary})
}

func (h *AdminHandler) updateAdminExperimentStatus(c *gin.Context, nextStatus string) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}

	if h.experimentAdminService == nil {
		response.InternalError(c, "Experiment service is unavailable")
		return
	}

	var audit *service.ExperimentStatusTransitionAudit
	adminID, _ := c.Get("admin_id")
	if aid, ok := adminID.(uuid.UUID); ok {
		audit = &service.ExperimentStatusTransitionAudit{
			ActorType: "admin",
			ActorID:   &aid,
			Source:    "admin_experiments_api",
			Details: map[string]interface{}{
				"reason": "manual_" + nextStatus,
			},
		}
	}

	err = h.experimentAdminService.TransitionExperimentStatusWithAudit(c.Request.Context(), experimentID, nextStatus, audit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		case errors.Is(err, service.ErrInvalidStatusTransition):
			response.UnprocessableEntity(c, err.Error())
		default:
			response.InternalError(c, "Failed to update experiment status")
		}
		return
	}

	updatedExperiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load updated experiment")
		return
	}

	response.OK(c, updatedExperiment)
}
