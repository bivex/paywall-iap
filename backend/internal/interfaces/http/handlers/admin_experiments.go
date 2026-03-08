package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

const (
	adminExperimentHoldForReviewReason = "Hold recommended winner for review"
	adminExperimentMaxMinSampleSize    = 2147483647
)

type AdminExperimentArm struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	IsControl     bool       `json:"is_control"`
	TrafficWeight float64    `json:"traffic_weight"`
	PricingTierID *uuid.UUID `json:"pricing_tier_id,omitempty"`
	Samples       int        `json:"samples"`
	Conversions   int        `json:"conversions"`
	Revenue       float64    `json:"revenue"`
	AvgReward     float64    `json:"avg_reward"`
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
	Name          string     `json:"name"`
	Description   *string    `json:"description"`
	IsControl     bool       `json:"is_control"`
	TrafficWeight float64    `json:"traffic_weight"`
	PricingTierID *uuid.UUID `json:"pricing_tier_id,omitempty"`
}

type createAdminExperimentRequest struct {
	Name                       string                              `json:"name"`
	Description                *string                             `json:"description"`
	Status                     string                              `json:"status"`
	AlgorithmType              *string                             `json:"algorithm_type"`
	IsBandit                   *bool                               `json:"is_bandit"`
	MinSampleSize              int                                 `json:"min_sample_size"`
	ConfidenceThresholdPercent float64                             `json:"confidence_threshold_percent"`
	StartAt                    *time.Time                          `json:"start_at"`
	EndAt                      *time.Time                          `json:"end_at"`
	AutomationPolicy           *service.ExperimentAutomationPolicy `json:"automation_policy,omitempty"`
	Arms                       []createAdminExperimentArmRequest   `json:"arms"`
}

type updateAdminExperimentRequest struct {
	Name                       string                              `json:"name"`
	Description                *string                             `json:"description"`
	AlgorithmType              *string                             `json:"algorithm_type"`
	IsBandit                   *bool                               `json:"is_bandit"`
	MinSampleSize              int                                 `json:"min_sample_size"`
	ConfidenceThresholdPercent float64                             `json:"confidence_threshold_percent"`
	StartAt                    *time.Time                          `json:"start_at"`
	EndAt                      *time.Time                          `json:"end_at"`
	AutomationPolicy           *service.ExperimentAutomationPolicy `json:"automation_policy,omitempty"`
	Arms                       []updateAdminExperimentArmRequest   `json:"arms,omitempty"`
}

type updateAdminExperimentArmRequest struct {
	ID            *uuid.UUID `json:"id,omitempty"`
	Name          string     `json:"name"`
	Description   *string    `json:"description"`
	IsControl     bool       `json:"is_control"`
	TrafficWeight float64    `json:"traffic_weight"`
	PricingTierID *uuid.UUID `json:"pricing_tier_id,omitempty"`
}

type updateAdminExperimentArmPricingTierRequest struct {
	ArmID         uuid.UUID  `json:"arm_id"`
	PricingTierID *uuid.UUID `json:"pricing_tier_id,omitempty"`
}

type updateAdminExperimentArmPricingTiersRequest struct {
	Arms []updateAdminExperimentArmPricingTierRequest `json:"arms"`
}

type lockAdminExperimentRequest struct {
	LockedUntil *time.Time `json:"locked_until"`
	Reason      string     `json:"reason"`
}

type updateAdminExperimentAutomationPolicyRequest struct {
	Enabled              *bool `json:"enabled"`
	AutoStart            *bool `json:"auto_start"`
	AutoComplete         *bool `json:"auto_complete"`
	CompleteOnEndTime    *bool `json:"complete_on_end_time"`
	CompleteOnSampleSize *bool `json:"complete_on_sample_size"`
	CompleteOnConfidence *bool `json:"complete_on_confidence"`
}

type repairAdminExperimentResponse struct {
	Experiment AdminExperiment                  `json:"experiment"`
	Summary    *service.ExperimentRepairSummary `json:"summary"`
}

var (
	errAdminPricingTierNotFound   = errors.New("pricing tier not found")
	errAdminExperimentArmNotFound = errors.New("experiment arm not found")
)

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
		       e.end_at,`

const adminExperimentSelectAutomationPolicy = `
		       e.automation_policy,`

const adminExperimentSelectAutomationPolicyMissing = `
		       '{"enabled":false,"auto_start":false,"auto_complete":false,"complete_on_end_time":true,"complete_on_sample_size":false,"complete_on_confidence":false,"manual_override":false}'::jsonb AS automation_policy,`

const adminExperimentSelectLatestLifecycle = `
		       latest_lifecycle.actor_type,
		       latest_lifecycle.source,
		       latest_lifecycle.action,
		       latest_lifecycle.from_status,
		       latest_lifecycle.to_status,
		       latest_lifecycle.idempotency_key,
		       latest_lifecycle.details,
		       latest_lifecycle.created_at,`

const adminExperimentSelectLatestLifecycleMissing = `
		       NULL::text AS actor_type,
		       NULL::text AS source,
		       NULL::text AS action,
		       NULL::text AS from_status,
		       NULL::text AS to_status,
		       NULL::text AS idempotency_key,
		       NULL::jsonb AS details,
		       NULL::timestamptz AS created_at,`

const adminExperimentSelectMeta = `
		       e.created_at,
		       e.updated_at,
		       (SELECT COUNT(*)::int FROM ab_test_arms a WHERE a.experiment_id = e.id) AS arm_count,`

const adminExperimentSelectStatsOnly = `
		       0::int AS total_assignments,
		       0::int AS active_assignments,
		       COALESCE((SELECT SUM(s.samples)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_samples,
		       COALESCE((SELECT SUM(s.conversions)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_conversions,
		       COALESCE((SELECT SUM(s.revenue)::double precision FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_revenue`

const adminExperimentSelectWithAssignments = `
		       (SELECT COUNT(*)::int FROM ab_test_assignments ass WHERE ass.experiment_id = e.id) AS total_assignments,
		       (SELECT COUNT(*)::int FROM ab_test_assignments ass WHERE ass.experiment_id = e.id AND ass.expires_at > NOW()) AS active_assignments,
		       COALESCE((SELECT SUM(s.samples)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_samples,
		       COALESCE((SELECT SUM(s.conversions)::int FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_conversions,
		       COALESCE((SELECT SUM(s.revenue)::double precision FROM ab_test_arm_stats s INNER JOIN ab_test_arms a ON a.id = s.arm_id WHERE a.experiment_id = e.id), 0) AS total_revenue`

const adminExperimentSelectFrom = `
		FROM ab_tests e`

const adminExperimentLatestLifecycleJoin = `
		LEFT JOIN LATERAL (
			SELECT actor_type, source, action, from_status, to_status, idempotency_key, details, created_at
			FROM experiment_lifecycle_audit_log l
			WHERE l.experiment_id = e.id
			ORDER BY l.created_at DESC
			LIMIT 1
		) latest_lifecycle ON true`

func normalizeCreateAdminExperimentRequest(req createAdminExperimentRequest) createAdminExperimentRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = normalizeOptionalTrimmedString(req.Description)
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	req.AlgorithmType = normalizeOptionalLowerTrimmedString(req.AlgorithmType)
	for index := range req.Arms {
		req.Arms[index].Name = strings.TrimSpace(req.Arms[index].Name)
		req.Arms[index].Description = normalizeOptionalTrimmedString(req.Arms[index].Description)
	}
	normalizedPolicy := service.NormalizeExperimentAutomationPolicy(req.AutomationPolicy)
	req.AutomationPolicy = &normalizedPolicy
	return req
}

func normalizeUpdateAdminExperimentRequest(req updateAdminExperimentRequest) updateAdminExperimentRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = normalizeOptionalTrimmedString(req.Description)
	req.AlgorithmType = normalizeOptionalLowerTrimmedString(req.AlgorithmType)
	for index := range req.Arms {
		req.Arms[index].Name = strings.TrimSpace(req.Arms[index].Name)
		req.Arms[index].Description = normalizeOptionalTrimmedString(req.Arms[index].Description)
	}
	normalizedPolicy := service.NormalizeExperimentAutomationPolicy(req.AutomationPolicy)
	req.AutomationPolicy = &normalizedPolicy
	return req
}

func normalizeOptionalTrimmedString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	return &normalized
}

func normalizeOptionalLowerTrimmedString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.ToLower(strings.TrimSpace(*value))
	return &normalized
}

func containsNullByte(value string) bool {
	return strings.ContainsRune(value, '\x00')
}

func containsControlCharacter(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func bindRequiredEmptyJSONObject(c *gin.Context) bool {
	var req map[string]json.RawMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return false
	}
	if req == nil {
		response.BadRequest(c, "Invalid request body")
		return false
	}
	if len(req) > 0 {
		response.BadRequest(c, "Invalid request body")
		return false
	}
	return true
}

func normalizeLockAdminExperimentRequest(req lockAdminExperimentRequest) lockAdminExperimentRequest {
	req.Reason = strings.TrimSpace(req.Reason)
	return req
}

func validateUpdateAdminExperimentAutomationPolicyRequest(req updateAdminExperimentAutomationPolicyRequest) string {
	if req.Enabled == nil || req.AutoStart == nil || req.AutoComplete == nil || req.CompleteOnEndTime == nil || req.CompleteOnSampleSize == nil || req.CompleteOnConfidence == nil {
		return "Every automation policy field is required"
	}
	return ""
}

func experimentAutomationPolicyUpdateInputFromRequest(req updateAdminExperimentAutomationPolicyRequest) service.UpdateExperimentAutomationPolicyInput {
	return service.UpdateExperimentAutomationPolicyInput{
		Enabled:              *req.Enabled,
		AutoStart:            *req.AutoStart,
		AutoComplete:         *req.AutoComplete,
		CompleteOnEndTime:    *req.CompleteOnEndTime,
		CompleteOnSampleSize: *req.CompleteOnSampleSize,
		CompleteOnConfidence: *req.CompleteOnConfidence,
	}
}

func serializableExperimentAutomationPolicy(policy service.ExperimentAutomationPolicy) map[string]interface{} {
	result := map[string]interface{}{
		"enabled":                 policy.Enabled,
		"auto_start":              policy.AutoStart,
		"auto_complete":           policy.AutoComplete,
		"complete_on_end_time":    policy.CompleteOnEndTime,
		"complete_on_sample_size": policy.CompleteOnSampleSize,
		"complete_on_confidence":  policy.CompleteOnConfidence,
		"manual_override":         policy.ManualOverride,
		"locked_until":            nil,
		"locked_by":               nil,
		"lock_reason":             nil,
	}
	if policy.LockedUntil != nil {
		result["locked_until"] = policy.LockedUntil.UTC().Format(time.RFC3339)
	}
	if policy.LockedBy != nil {
		result["locked_by"] = policy.LockedBy.String()
	}
	if policy.LockReason != nil {
		result["lock_reason"] = *policy.LockReason
	}
	return result
}

func experimentAutomationPolicyChangedFields(before, after service.ExperimentAutomationPolicy) []string {
	fields := make([]string, 0, 10)
	if before.Enabled != after.Enabled {
		fields = append(fields, "enabled")
	}
	if before.AutoStart != after.AutoStart {
		fields = append(fields, "auto_start")
	}
	if before.AutoComplete != after.AutoComplete {
		fields = append(fields, "auto_complete")
	}
	if before.CompleteOnEndTime != after.CompleteOnEndTime {
		fields = append(fields, "complete_on_end_time")
	}
	if before.CompleteOnSampleSize != after.CompleteOnSampleSize {
		fields = append(fields, "complete_on_sample_size")
	}
	if before.CompleteOnConfidence != after.CompleteOnConfidence {
		fields = append(fields, "complete_on_confidence")
	}
	if before.ManualOverride != after.ManualOverride {
		fields = append(fields, "manual_override")
	}

	beforeLockedUntil := ""
	afterLockedUntil := ""
	if before.LockedUntil != nil {
		beforeLockedUntil = before.LockedUntil.UTC().Format(time.RFC3339)
	}
	if after.LockedUntil != nil {
		afterLockedUntil = after.LockedUntil.UTC().Format(time.RFC3339)
	}
	if beforeLockedUntil != afterLockedUntil {
		fields = append(fields, "locked_until")
	}

	beforeLockedBy := ""
	afterLockedBy := ""
	if before.LockedBy != nil {
		beforeLockedBy = before.LockedBy.String()
	}
	if after.LockedBy != nil {
		afterLockedBy = after.LockedBy.String()
	}
	if beforeLockedBy != afterLockedBy {
		fields = append(fields, "locked_by")
	}

	beforeLockReason := ""
	afterLockReason := ""
	if before.LockReason != nil {
		beforeLockReason = *before.LockReason
	}
	if after.LockReason != nil {
		afterLockReason = *after.LockReason
	}
	if beforeLockReason != afterLockReason {
		fields = append(fields, "lock_reason")
	}

	return fields
}

func (h *AdminHandler) logExperimentAutomationPolicyAction(c *gin.Context, experimentID uuid.UUID, before, after service.ExperimentAutomationPolicy) {
	if h.auditService == nil {
		return
	}
	adminIDValue, ok := c.Get("admin_id")
	if !ok {
		return
	}
	adminID, ok := adminIDValue.(uuid.UUID)
	if !ok {
		return
	}

	changedFields := experimentAutomationPolicyChangedFields(before, after)
	if len(changedFields) == 0 {
		return
	}

	_ = h.auditService.LogAction(c.Request.Context(), adminID, "update_experiment_automation_policy", "experiment", nil, map[string]interface{}{
		"experiment_id":  experimentID.String(),
		"changed_fields": changedFields,
		"before":         serializableExperimentAutomationPolicy(before),
		"after":          serializableExperimentAutomationPolicy(after),
	})
}

func adminExperimentHasActiveAutomationLock(policy service.ExperimentAutomationPolicy, now time.Time) bool {
	if policy.ManualOverride {
		return true
	}
	return policy.LockedUntil != nil && policy.LockedUntil.After(now.UTC())
}

func adminIDFromContext(c *gin.Context) (*uuid.UUID, bool) {
	adminIDValue, ok := c.Get("admin_id")
	if !ok {
		return nil, false
	}
	adminID, ok := adminIDValue.(uuid.UUID)
	if !ok {
		return nil, false
	}
	return &adminID, true
}

func buildAdminExperimentStatusAudit(c *gin.Context, reason string, details map[string]interface{}) *service.ExperimentStatusTransitionAudit {
	adminID, ok := adminIDFromContext(c)
	if !ok {
		return nil
	}
	mergedDetails := map[string]interface{}{
		"reason": reason,
	}
	for key, value := range details {
		mergedDetails[key] = value
	}
	return &service.ExperimentStatusTransitionAudit{
		ActorType: "admin",
		ActorID:   adminID,
		Source:    "admin_experiments_api",
		Details:   mergedDetails,
	}
}

func (h *AdminHandler) logConfirmExperimentWinnerAction(c *gin.Context, experiment AdminExperiment) {
	if h.auditService == nil || experiment.WinnerRecommendation == nil {
		return
	}
	adminID, ok := adminIDFromContext(c)
	if !ok {
		return
	}
	details := map[string]interface{}{
		"experiment_id":         experiment.ID.String(),
		"status":                experiment.Status,
		"recommendation_reason": experiment.WinnerRecommendation.Reason,
		"recommended":           experiment.WinnerRecommendation.Recommended,
		"observed_samples":      experiment.WinnerRecommendation.ObservedSamples,
		"min_sample_size":       experiment.WinnerRecommendation.MinSampleSize,
		"confidence_threshold":  experiment.WinnerRecommendation.ConfidenceThresholdPercent,
		"recommendation_source": "admin_experiments_api",
	}
	if experiment.WinnerRecommendation.WinningArmID != nil {
		details["winning_arm_id"] = experiment.WinnerRecommendation.WinningArmID.String()
	}
	if experiment.WinnerRecommendation.WinningArmName != nil {
		details["winning_arm_name"] = *experiment.WinnerRecommendation.WinningArmName
	}
	if experiment.WinnerRecommendation.ConfidencePercent != nil {
		details["confidence_percent"] = *experiment.WinnerRecommendation.ConfidencePercent
	}
	_ = h.auditService.LogAction(c.Request.Context(), *adminID, "confirm_experiment_winner", "experiment", nil, details)
}

func adminExperimentHasConfirmableWinnerRecommendation(experiment AdminExperiment) bool {
	return experiment.WinnerRecommendation != nil &&
		experiment.WinnerRecommendation.Recommended &&
		experiment.WinnerRecommendation.Reason == "recommend_winner" &&
		experiment.WinnerRecommendation.WinningArmID != nil
}

func adminExperimentWinnerRecommendationAuditDetails(experiment AdminExperiment) map[string]interface{} {
	if experiment.WinnerRecommendation == nil {
		return map[string]interface{}{}
	}
	details := map[string]interface{}{
		"recommended":           experiment.WinnerRecommendation.Recommended,
		"recommendation_reason": experiment.WinnerRecommendation.Reason,
		"observed_samples":      experiment.WinnerRecommendation.ObservedSamples,
		"min_sample_size":       experiment.WinnerRecommendation.MinSampleSize,
		"confidence_threshold":  experiment.WinnerRecommendation.ConfidenceThresholdPercent,
		"recommendation_source": "admin_experiments_api",
	}
	if experiment.WinnerRecommendation.WinningArmID != nil {
		details["winning_arm_id"] = experiment.WinnerRecommendation.WinningArmID.String()
	}
	if experiment.WinnerRecommendation.WinningArmName != nil {
		details["winning_arm_name"] = *experiment.WinnerRecommendation.WinningArmName
	}
	if experiment.WinnerRecommendation.ConfidencePercent != nil {
		details["confidence_percent"] = *experiment.WinnerRecommendation.ConfidencePercent
	}
	return details
}

func (h *AdminHandler) logHoldExperimentForReviewAction(c *gin.Context, experiment AdminExperiment, statusAfter string) {
	if h.auditService == nil || experiment.WinnerRecommendation == nil {
		return
	}
	adminID, ok := adminIDFromContext(c)
	if !ok {
		return
	}
	details := adminExperimentWinnerRecommendationAuditDetails(experiment)
	details["experiment_id"] = experiment.ID.String()
	details["status_before"] = experiment.Status
	details["status_after"] = statusAfter
	details["lock_reason"] = adminExperimentHoldForReviewReason
	details["manual_override"] = true
	_ = h.auditService.LogAction(c.Request.Context(), *adminID, "hold_experiment_for_review", "experiment", nil, details)
}

func validateCreateAdminExperimentRequest(req createAdminExperimentRequest) string {
	if req.Name == "" {
		return "Experiment name is required"
	}
	if containsNullByte(req.Name) {
		return "Experiment name cannot contain null bytes"
	}
	if containsControlCharacter(req.Name) {
		return "Experiment name cannot contain control characters"
	}
	if req.Description == nil {
		return "Experiment description is required"
	}
	if containsNullByte(*req.Description) {
		return "Experiment description cannot contain null bytes"
	}
	if req.AlgorithmType == nil {
		return "Algorithm type is required"
	}
	if containsNullByte(*req.AlgorithmType) {
		return "Algorithm type cannot contain null bytes"
	}
	switch req.Status {
	case "draft", "running", "paused", "completed":
	default:
		return "Status must be draft, running, paused, or completed"
	}
	if req.MinSampleSize <= 0 {
		return "Minimum sample size must be greater than zero"
	}
	if req.IsBandit == nil {
		return "Bandit flag is required"
	}
	if req.MinSampleSize > adminExperimentMaxMinSampleSize {
		return "Minimum sample size must be less than or equal to 2147483647"
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
		if containsNullByte(arm.Name) {
			return "Experiment arm names cannot contain null bytes"
		}
		if containsControlCharacter(arm.Name) {
			return "Experiment arm names cannot contain control characters"
		}
		if arm.Description == nil {
			return "Every experiment arm must include a description"
		}
		if containsNullByte(*arm.Description) {
			return "Experiment arm descriptions cannot contain null bytes"
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
	switch *req.AlgorithmType {
	case "thompson_sampling", "ucb", "epsilon_greedy":
	default:
		return "Algorithm type must be thompson_sampling, ucb, or epsilon_greedy"
	}
	return ""
}

func validateUpdateAdminExperimentRequest(req updateAdminExperimentRequest) string {
	if req.Name == "" {
		return "Experiment name is required"
	}
	if containsNullByte(req.Name) {
		return "Experiment name cannot contain null bytes"
	}
	if containsControlCharacter(req.Name) {
		return "Experiment name cannot contain control characters"
	}
	if req.Description == nil {
		return "Experiment description is required"
	}
	if containsNullByte(*req.Description) {
		return "Experiment description cannot contain null bytes"
	}
	if req.AlgorithmType == nil {
		return "Algorithm type is required"
	}
	if containsNullByte(*req.AlgorithmType) {
		return "Algorithm type cannot contain null bytes"
	}
	if req.IsBandit == nil {
		return "Bandit flag is required"
	}
	if req.MinSampleSize <= 0 {
		return "Minimum sample size must be greater than zero"
	}
	if req.MinSampleSize > adminExperimentMaxMinSampleSize {
		return "Minimum sample size must be less than or equal to 2147483647"
	}
	if req.ConfidenceThresholdPercent <= 0 || req.ConfidenceThresholdPercent > 100 {
		return "Confidence threshold must be between 0 and 100"
	}
	if req.StartAt != nil && req.EndAt != nil && req.EndAt.Before(*req.StartAt) {
		return "End time must be after start time"
	}
	switch *req.AlgorithmType {
	case "thompson_sampling", "ucb", "epsilon_greedy":
	default:
		return "Algorithm type must be thompson_sampling, ucb, or epsilon_greedy"
	}
	if req.Arms != nil {
		if len(req.Arms) < 2 {
			return "At least two experiment arms are required"
		}
		controlCount := 0
		seenIDs := make(map[uuid.UUID]struct{}, len(req.Arms))
		for _, arm := range req.Arms {
			if arm.Name == "" {
				return "Every experiment arm must have a name"
			}
			if containsNullByte(arm.Name) {
				return "Experiment arm names cannot contain null bytes"
			}
			if containsControlCharacter(arm.Name) {
				return "Experiment arm names cannot contain control characters"
			}
			if arm.Description == nil {
				return "Every experiment arm must include a description"
			}
			if containsNullByte(*arm.Description) {
				return "Experiment arm descriptions cannot contain null bytes"
			}
			if arm.TrafficWeight <= 0 {
				return "Traffic weight must be greater than zero"
			}
			if arm.ID != nil {
				if _, exists := seenIDs[*arm.ID]; exists {
					return "Each persisted experiment arm may only appear once"
				}
				seenIDs[*arm.ID] = struct{}{}
			}
			if arm.IsControl {
				controlCount++
			}
		}
		if controlCount != 1 {
			return "Exactly one control arm is required"
		}
	}
	return ""
}

func experimentArmInputsFromUpdateRequest(arms []updateAdminExperimentArmRequest) []service.ExperimentArmInput {
	if arms == nil {
		return nil
	}
	result := make([]service.ExperimentArmInput, 0, len(arms))
	for _, arm := range arms {
		result = append(result, service.ExperimentArmInput{
			ID:            arm.ID,
			Name:          arm.Name,
			Description:   *arm.Description,
			IsControl:     arm.IsControl,
			TrafficWeight: arm.TrafficWeight,
			PricingTierID: arm.PricingTierID,
		})
	}
	return result
}

func validateUpdateAdminExperimentArmPricingTiersRequest(req updateAdminExperimentArmPricingTiersRequest) string {
	if len(req.Arms) == 0 {
		return "At least one arm pricing tier update is required"
	}
	seen := make(map[uuid.UUID]struct{}, len(req.Arms))
	for _, arm := range req.Arms {
		if arm.ArmID == uuid.Nil {
			return "Every arm linkage must include an arm_id"
		}
		if _, exists := seen[arm.ArmID]; exists {
			return "Each arm may only appear once in a pricing tier update"
		}
		seen[arm.ArmID] = struct{}{}
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
	var pricingTierID uuid.NullUUID
	err := scanner.Scan(
		&arm.ID,
		&arm.Name,
		&description,
		&arm.IsControl,
		&arm.TrafficWeight,
		&pricingTierID,
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
	if pricingTierID.Valid {
		value := pricingTierID.UUID
		arm.PricingTierID = &value
	}
	return arm, nil
}

func pricingTierIDsFromCreateExperimentArms(arms []createAdminExperimentArmRequest) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	ids := make([]uuid.UUID, 0)
	for _, arm := range arms {
		if arm.PricingTierID == nil {
			continue
		}
		if _, exists := seen[*arm.PricingTierID]; exists {
			continue
		}
		seen[*arm.PricingTierID] = struct{}{}
		ids = append(ids, *arm.PricingTierID)
	}
	return ids
}

func pricingTierIDsFromArmPricingTierUpdates(arms []updateAdminExperimentArmPricingTierRequest) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	ids := make([]uuid.UUID, 0)
	for _, arm := range arms {
		if arm.PricingTierID == nil {
			continue
		}
		if _, exists := seen[*arm.PricingTierID]; exists {
			continue
		}
		seen[*arm.PricingTierID] = struct{}{}
		ids = append(ids, *arm.PricingTierID)
	}
	return ids
}

func validatePricingTiersExist(ctx context.Context, tx pgx.Tx, pricingTierIDs []uuid.UUID) error {
	for _, pricingTierID := range pricingTierIDs {
		if err := tx.QueryRow(ctx, `
			SELECT 1
			FROM pricing_tiers
			WHERE id = $1 AND deleted_at IS NULL`, pricingTierID).Scan(new(int)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errAdminPricingTierNotFound
			}
			return err
		}
	}
	return nil
}

func applyExperimentArmPricingTierUpdates(ctx context.Context, tx pgx.Tx, experimentID uuid.UUID, arms []updateAdminExperimentArmPricingTierRequest) error {
	for _, arm := range arms {
		commandTag, err := tx.Exec(ctx, `
			UPDATE ab_test_arms
			SET pricing_tier_id = $3,
			    updated_at = now()
			WHERE id = $1 AND experiment_id = $2`, arm.ArmID, experimentID, arm.PricingTierID)
		if err != nil {
			return err
		}
		if commandTag.RowsAffected() == 0 {
			return errAdminExperimentArmNotFound
		}
	}
	return nil
}

func (h *AdminHandler) hasTable(ctx context.Context, relation string) bool {
	var exists bool
	err := h.dbPool.QueryRow(ctx, `
		SELECT to_regclass($1) IS NOT NULL`, relation).Scan(&exists)
	return err == nil && exists
}

func (h *AdminHandler) hasColumn(ctx context.Context, tableName string, columnName string) bool {
	var exists bool
	err := h.dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		)`, tableName, columnName).Scan(&exists)
	return err == nil && exists
}

func (h *AdminHandler) hasAssignmentTable(c *gin.Context) bool {
	return h.hasTable(c.Request.Context(), "public.ab_test_assignments")
}

func (h *AdminHandler) hasLifecycleAuditTable(c *gin.Context) bool {
	return h.hasTable(c.Request.Context(), "public.experiment_lifecycle_audit_log")
}

func (h *AdminHandler) hasWinnerRecommendationLogTable(c *gin.Context) bool {
	return h.hasTable(c.Request.Context(), "public.experiment_winner_recommendation_log")
}

func (h *AdminHandler) hasExperimentAutomationPolicyColumn(c *gin.Context) bool {
	return h.hasColumn(c.Request.Context(), "ab_tests", "automation_policy")
}

func (h *AdminHandler) hasExperimentArmPricingTierColumn(c *gin.Context) bool {
	return h.hasColumn(c.Request.Context(), "ab_test_arms", "pricing_tier_id")
}

func adminExperimentListQuery(withAssignments bool, withLifecycleAudit bool, withAutomationPolicy bool) string {
	lifecycleColumns := adminExperimentSelectLatestLifecycleMissing
	lifecycleJoin := ""
	automationPolicyColumns := adminExperimentSelectAutomationPolicyMissing
	if withAutomationPolicy {
		automationPolicyColumns = adminExperimentSelectAutomationPolicy
	}
	if withLifecycleAudit {
		lifecycleColumns = adminExperimentSelectLatestLifecycle
		lifecycleJoin = adminExperimentLatestLifecycleJoin
	}
	if withAssignments {
		return adminExperimentSelectBase + automationPolicyColumns + lifecycleColumns + adminExperimentSelectMeta + adminExperimentSelectWithAssignments + adminExperimentSelectFrom + lifecycleJoin + `
		ORDER BY e.created_at DESC`
	}
	return adminExperimentSelectBase + automationPolicyColumns + lifecycleColumns + adminExperimentSelectMeta + adminExperimentSelectStatsOnly + adminExperimentSelectFrom + lifecycleJoin + `
		ORDER BY e.created_at DESC`
}

func adminExperimentByIDQuery(withAssignments bool, withLifecycleAudit bool, withAutomationPolicy bool) string {
	lifecycleColumns := adminExperimentSelectLatestLifecycleMissing
	lifecycleJoin := ""
	automationPolicyColumns := adminExperimentSelectAutomationPolicyMissing
	if withAutomationPolicy {
		automationPolicyColumns = adminExperimentSelectAutomationPolicy
	}
	if withLifecycleAudit {
		lifecycleColumns = adminExperimentSelectLatestLifecycle
		lifecycleJoin = adminExperimentLatestLifecycleJoin
	}
	if withAssignments {
		return adminExperimentSelectBase + automationPolicyColumns + lifecycleColumns + adminExperimentSelectMeta + adminExperimentSelectWithAssignments + adminExperimentSelectFrom + lifecycleJoin + `
		WHERE e.id = $1`
	}
	return adminExperimentSelectBase + automationPolicyColumns + lifecycleColumns + adminExperimentSelectMeta + adminExperimentSelectStatsOnly + adminExperimentSelectFrom + lifecycleJoin + `
		WHERE e.id = $1`
}

func (h *AdminHandler) listExperimentArms(ctx *gin.Context, experimentID uuid.UUID) ([]AdminExperimentArm, error) {
	pricingTierSelect := `NULL::uuid`
	if h.hasExperimentArmPricingTierColumn(ctx) {
		pricingTierSelect = `a.pricing_tier_id`
	}
	rows, err := h.dbPool.Query(ctx.Request.Context(), `
		SELECT a.id,
		       a.name,
		       a.description,
		       a.is_control,
		       a.traffic_weight::double precision,
		       `+pricingTierSelect+`,
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
	withAssignments := h.hasAssignmentTable(c)
	withLifecycleAudit := h.hasLifecycleAuditTable(c)
	withAutomationPolicy := h.hasExperimentAutomationPolicyColumn(c)
	experiment, err := scanAdminExperiment(h.dbPool.QueryRow(c.Request.Context(), adminExperimentByIDQuery(withAssignments, withLifecycleAudit, withAutomationPolicy), experimentID))
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
	if !h.hasWinnerRecommendationLogTable(ctx) {
		return []AdminExperimentWinnerRecommendationAudit{}, nil
	}
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
	if !h.hasLifecycleAuditTable(ctx) {
		return []AdminExperimentLifecycleAudit{}, nil
	}
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
	withAssignments := h.hasAssignmentTable(c)
	withLifecycleAudit := h.hasLifecycleAuditTable(c)
	withAutomationPolicy := h.hasExperimentAutomationPolicyColumn(c)
	rows, err := h.dbPool.Query(c.Request.Context(), adminExperimentListQuery(withAssignments, withLifecycleAudit, withAutomationPolicy))
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
		if message == "End time must be after start time" {
			response.Conflict(c, message)
		} else {
			response.UnprocessableEntity(c, message)
		}
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
	if *req.IsBandit {
		algorithmType = *req.AlgorithmType
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
		*req.Description,
		req.Status,
		req.StartAt,
		req.EndAt,
		algorithmType,
		*req.IsBandit,
		req.MinSampleSize,
		req.ConfidenceThresholdPercent/100,
		automationPolicyJSON,
	)
	if err != nil {
		response.InternalError(c, "Failed to create experiment")
		return
	}
	if err := validatePricingTiersExist(ctx, tx, pricingTierIDsFromCreateExperimentArms(req.Arms)); err != nil {
		switch {
		case errors.Is(err, errAdminPricingTierNotFound):
			response.UnprocessableEntity(c, "Linked pricing tier not found")
		default:
			response.InternalError(c, "Failed to validate pricing tier linkage")
		}
		return
	}

	for _, arm := range req.Arms {
		_, err = tx.Exec(ctx, `
				INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight, pricing_tier_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(),
			experimentID,
			arm.Name,
			*arm.Description,
			arm.IsControl,
			arm.TrafficWeight,
			arm.PricingTierID,
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
		if message == "End time must be after start time" {
			response.Conflict(c, message)
		} else {
			response.UnprocessableEntity(c, message)
		}
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
		response.Conflict(c, "Only draft experiments can be edited")
		return
	}

	var algorithmType *string
	if *req.IsBandit {
		algorithmType = req.AlgorithmType
	}

	if h.experimentAdminService == nil {
		response.InternalError(c, "Experiment service is unavailable")
		return
	}

	err = h.experimentAdminService.UpdateDraftExperiment(c.Request.Context(), experimentID, service.UpdateExperimentInput{
		Name:                req.Name,
		Description:         *req.Description,
		AlgorithmType:       algorithmType,
		IsBandit:            *req.IsBandit,
		MinSampleSize:       req.MinSampleSize,
		ConfidenceThreshold: req.ConfidenceThresholdPercent / 100,
		StartAt:             req.StartAt,
		EndAt:               req.EndAt,
		AutomationPolicy:    service.NormalizeExperimentAutomationPolicy(req.AutomationPolicy),
		Arms:                experimentArmInputsFromUpdateRequest(req.Arms),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		case errors.Is(err, service.ErrExperimentNotEditable):
			response.Conflict(c, "Only draft experiments can be edited")
		case errors.Is(err, service.ErrExperimentArmNotFound):
			response.UnprocessableEntity(c, "All persisted arm IDs must belong to the experiment")
		case errors.Is(err, service.ErrPricingTierNotFound):
			response.UnprocessableEntity(c, "Linked pricing tier not found")
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

func (h *AdminHandler) UpdateAdminExperimentAutomationPolicy(c *gin.Context) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}
	if h.experimentAdminService == nil {
		response.InternalError(c, "Experiment service is unavailable")
		return
	}

	var req updateAdminExperimentAutomationPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid experiment automation policy payload")
		return
	}
	if message := validateUpdateAdminExperimentAutomationPolicyRequest(req); message != "" {
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

	err = h.experimentAdminService.UpdateExperimentAutomationPolicy(
		c.Request.Context(),
		experimentID,
		experimentAutomationPolicyUpdateInputFromRequest(req),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		case errors.Is(err, service.ErrExperimentAutomationPolicyNotEditable):
			response.UnprocessableEntity(c, "Completed experiments cannot update automation policy")
		default:
			response.InternalError(c, "Failed to update experiment automation policy")
		}
		return
	}

	updatedExperiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load updated experiment")
		return
	}

	h.logExperimentAutomationPolicyAction(c, experimentID, experiment.AutomationPolicy, updatedExperiment.AutomationPolicy)
	response.OK(c, updatedExperiment)
}

func (h *AdminHandler) UpdateAdminExperimentArmPricingTiers(c *gin.Context) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}

	var req updateAdminExperimentArmPricingTiersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid experiment arm pricing tier payload")
		return
	}
	if message := validateUpdateAdminExperimentArmPricingTiersRequest(req); message != "" {
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

	ctx := c.Request.Context()
	tx, err := h.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		response.InternalError(c, "Failed to start experiment arm pricing tier transaction")
		return
	}
	defer tx.Rollback(ctx)

	if err := validatePricingTiersExist(ctx, tx, pricingTierIDsFromArmPricingTierUpdates(req.Arms)); err != nil {
		switch {
		case errors.Is(err, errAdminPricingTierNotFound):
			response.UnprocessableEntity(c, "Linked pricing tier not found")
		default:
			response.InternalError(c, "Failed to validate pricing tier linkage")
		}
		return
	}

	if err := applyExperimentArmPricingTierUpdates(ctx, tx, experimentID, req.Arms); err != nil {
		switch {
		case errors.Is(err, errAdminExperimentArmNotFound):
			response.UnprocessableEntity(c, "All arm IDs must belong to the experiment")
		default:
			response.InternalError(c, "Failed to update experiment arm pricing tiers")
		}
		return
	}

	if err := tx.Commit(ctx); err != nil {
		response.InternalError(c, "Failed to commit experiment arm pricing tier update")
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

func (h *AdminHandler) ConfirmAdminExperimentWinner(c *gin.Context) {
	if !bindRequiredEmptyJSONObject(c) {
		return
	}

	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}
	if h.experimentAdminService == nil {
		response.InternalError(c, "Experiment service is unavailable")
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
	if !experiment.IsBandit {
		response.NotFound(c, "Only bandit experiments can confirm winner recommendations")
		return
	}
	if experiment.Status != "running" && experiment.Status != "paused" {
		response.NotFound(c, "Only running or paused experiments can confirm winner recommendations")
		return
	}
	if adminExperimentHasActiveAutomationLock(experiment.AutomationPolicy, time.Now()) {
		response.NotFound(c, "Unlock experiment automation before confirming a winner")
		return
	}
	if !adminExperimentHasConfirmableWinnerRecommendation(experiment) {
		response.NotFound(c, "Experiment does not have a confirmable winner recommendation")
		return
	}

	auditDetails := adminExperimentWinnerRecommendationAuditDetails(experiment)

	err = h.experimentAdminService.TransitionExperimentStatusWithAudit(
		c.Request.Context(),
		experimentID,
		"completed",
		buildAdminExperimentStatusAudit(c, "confirm_recommended_winner", auditDetails),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		case errors.Is(err, service.ErrInvalidStatusTransition):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, "Failed to confirm experiment winner")
		}
		return
	}

	h.logConfirmExperimentWinnerAction(c, experiment)
	updatedExperiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load updated experiment")
		return
	}
	response.OK(c, updatedExperiment)
}

func (h *AdminHandler) HoldAdminExperimentForReview(c *gin.Context) {
	if !bindRequiredEmptyJSONObject(c) {
		return
	}

	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}
	if h.experimentAdminService == nil {
		response.InternalError(c, "Experiment service is unavailable")
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
	if !experiment.IsBandit {
		response.NotFound(c, "Only bandit experiments can be held for winner review")
		return
	}
	if experiment.Status != "running" && experiment.Status != "paused" {
		response.NotFound(c, "Only running or paused experiments can be held for winner review")
		return
	}
	if !adminExperimentHasConfirmableWinnerRecommendation(experiment) {
		response.NotFound(c, "Experiment does not have a confirmable winner recommendation")
		return
	}

	adminID, _ := adminIDFromContext(c)
	auditDetails := adminExperimentWinnerRecommendationAuditDetails(experiment)
	auditDetails["lock_reason"] = adminExperimentHoldForReviewReason
	auditDetails["manual_override"] = true
	err = h.experimentAdminService.HoldExperimentForReview(
		c.Request.Context(),
		experimentID,
		service.ExperimentLockInput{
			LockedBy:   adminID,
			LockReason: adminExperimentHoldForReviewReason,
		},
		buildAdminExperimentStatusAudit(c, "hold_recommended_winner_review", auditDetails),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		case errors.Is(err, service.ErrInvalidStatusTransition):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, "Failed to hold experiment for review")
		}
		return
	}

	updatedExperiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.InternalError(c, "Failed to load updated experiment")
		return
	}
	h.logExperimentAutomationPolicyAction(c, experimentID, experiment.AutomationPolicy, updatedExperiment.AutomationPolicy)
	h.logHoldExperimentForReviewAction(c, experiment, updatedExperiment.Status)
	response.OK(c, updatedExperiment)
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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	req = normalizeLockAdminExperimentRequest(req)

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
	if !bindRequiredEmptyJSONObject(c) {
		return
	}

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
	if !bindRequiredEmptyJSONObject(c) {
		return
	}

	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}
	if h.experimentRepairService == nil {
		response.InternalError(c, "Experiment repair service is unavailable")
		return
	}

	experiment, err := h.getAdminExperimentByID(c, experimentID)
	if err != nil {
		response.NotFound(c, "Experiment not found")
		return
	}
	if experiment.Status == "draft" {
		response.NotFound(c, "Experiment repair is unavailable for draft experiments")
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
	if !bindRequiredEmptyJSONObject(c) {
		return
	}

	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}

	if h.experimentAdminService == nil {
		response.InternalError(c, "Experiment service is unavailable")
		return
	}

	audit := buildAdminExperimentStatusAudit(c, "manual_"+nextStatus, nil)

	err = h.experimentAdminService.TransitionExperimentStatusWithAudit(c.Request.Context(), experimentID, nextStatus, audit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExperimentNotFound):
			response.NotFound(c, "Experiment not found")
		case errors.Is(err, service.ErrInvalidStatusTransition):
			response.NotFound(c, err.Error())
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
