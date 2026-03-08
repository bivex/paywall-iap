package regression

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	service "github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	"github.com/bivex/paywall-iap/tests/testutil"
)

type automationPolicyRepoStub struct {
	state         *service.ExperimentMutationState
	updatedPolicy *service.ExperimentAutomationPolicy
}

func (s *automationPolicyRepoStub) GetExperimentMutationState(context.Context, uuid.UUID) (*service.ExperimentMutationState, error) {
	return s.state, nil
}

func (s *automationPolicyRepoStub) UpdateExperimentDraft(context.Context, uuid.UUID, service.UpdateExperimentInput) error {
	return nil
}

func (s *automationPolicyRepoStub) UpdateExperimentStatus(context.Context, uuid.UUID, string, *time.Time, *time.Time) error {
	return nil
}

func (s *automationPolicyRepoStub) UpdateExperimentStatusWithAudit(context.Context, uuid.UUID, string, string, *time.Time, *time.Time, *service.ExperimentStatusTransitionAudit) error {
	return nil
}

func (s *automationPolicyRepoStub) UpdateExperimentAutomationPolicy(_ context.Context, _ uuid.UUID, policy service.ExperimentAutomationPolicy) error {
	value := policy
	s.updatedPolicy = &value
	return nil
}

func TestExperimentAdminServiceUpdateExperimentAutomationPolicyPreservesLocks(t *testing.T) {
	lockedUntil := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	lockedBy := uuid.New()
	lockReason := "Freeze automation"
	repo := &automationPolicyRepoStub{state: &service.ExperimentMutationState{
		ID:     uuid.New(),
		Status: "running",
		AutomationPolicy: service.ExperimentAutomationPolicy{
			Enabled:        false,
			ManualOverride: true,
			LockedUntil:    &lockedUntil,
			LockedBy:       &lockedBy,
			LockReason:     &lockReason,
		},
	}}

	svc := service.NewExperimentAdminService(repo)
	err := svc.UpdateExperimentAutomationPolicy(context.Background(), repo.state.ID, service.UpdateExperimentAutomationPolicyInput{
		Enabled:      true,
		AutoStart:    true,
		AutoComplete: true,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.updatedPolicy)
	assert.True(t, repo.updatedPolicy.Enabled)
	assert.True(t, repo.updatedPolicy.AutoStart)
	assert.True(t, repo.updatedPolicy.AutoComplete)
	assert.True(t, repo.updatedPolicy.CompleteOnEndTime)
	assert.True(t, repo.updatedPolicy.ManualOverride)
	require.NotNil(t, repo.updatedPolicy.LockedUntil)
	assert.True(t, repo.updatedPolicy.LockedUntil.Equal(lockedUntil))
	require.NotNil(t, repo.updatedPolicy.LockedBy)
	assert.Equal(t, lockedBy, *repo.updatedPolicy.LockedBy)
	require.NotNil(t, repo.updatedPolicy.LockReason)
	assert.Equal(t, lockReason, *repo.updatedPolicy.LockReason)
}

func TestExperimentAdminServiceUpdateExperimentAutomationPolicyRejectsCompletedExperiments(t *testing.T) {
	repo := &automationPolicyRepoStub{state: &service.ExperimentMutationState{ID: uuid.New(), Status: "completed"}}
	svc := service.NewExperimentAdminService(repo)

	err := svc.UpdateExperimentAutomationPolicy(context.Background(), repo.state.ID, service.UpdateExperimentAutomationPolicyInput{Enabled: true})

	require.ErrorIs(t, err, service.ErrExperimentAutomationPolicyNotEditable)
	assert.Nil(t, repo.updatedPolicy)
}

func TestAdminExperimentAutomationPolicyEndpointPersistsFlagsAndAuditLog(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE users (
			id UUID PRIMARY KEY,
			platform_user_id TEXT UNIQUE NOT NULL,
			device_id TEXT,
			platform TEXT NOT NULL,
			app_version TEXT NOT NULL,
			email TEXT UNIQUE,
			role TEXT NOT NULL DEFAULT 'user',
			ltv NUMERIC(10,2) DEFAULT 0,
			ltv_updated_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ
		);
		CREATE TABLE admin_audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			admin_id UUID NOT NULL REFERENCES users(id),
			action TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_user_id UUID,
			details JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE ab_tests (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'paused', 'completed')) DEFAULT 'draft',
			start_at TIMESTAMPTZ,
			end_at TIMESTAMPTZ,
			algorithm_type TEXT CHECK (algorithm_type IN ('thompson_sampling', 'ucb', 'epsilon_greedy')),
			is_bandit BOOLEAN NOT NULL DEFAULT false,
			min_sample_size INT DEFAULT 100,
			confidence_threshold NUMERIC(3,2) DEFAULT 0.95,
			winner_confidence NUMERIC(3,2),
			automation_policy JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE ab_test_arms (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			description TEXT,
			is_control BOOLEAN NOT NULL DEFAULT false,
			traffic_weight NUMERIC(3,2) NOT NULL DEFAULT 1.0,
			pricing_tier_id UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE ab_test_arm_stats (
			arm_id UUID PRIMARY KEY REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			samples INT NOT NULL DEFAULT 0,
			conversions INT NOT NULL DEFAULT 0,
			revenue NUMERIC(10,2) NOT NULL DEFAULT 0,
			avg_reward NUMERIC(10,4) NOT NULL DEFAULT 0
		);
	`)
	require.NoError(t, err)

	adminID := uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role)
		VALUES ($1, 'admin-user', 'admin-device', 'ios', '1.0.0', 'admin@example.com', 'admin')
	`, adminID)
	require.NoError(t, err)

	experimentID := uuid.New()
	armID := uuid.New()
	lockedUntil := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	policyJSON, err := json.Marshal(map[string]interface{}{
		"enabled":                 false,
		"auto_start":              false,
		"auto_complete":           false,
		"complete_on_end_time":    true,
		"complete_on_sample_size": false,
		"complete_on_confidence":  false,
		"manual_override":         true,
		"locked_until":            lockedUntil.Format(time.RFC3339),
		"locked_by":               adminID.String(),
		"lock_reason":             "Freeze automation",
	})
	require.NoError(t, err)

	_, err = db.Exec(ctx, `
		INSERT INTO ab_tests (
			id, name, description, status, algorithm_type, is_bandit, min_sample_size, confidence_threshold, automation_policy
		) VALUES ($1, 'Live automation policy', 'Regression test', 'running', 'thompson_sampling', true, 200, 0.95, $2)
	`, experimentID, policyJSON)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `
		INSERT INTO ab_test_arms (id, experiment_id, name, is_control, traffic_weight)
		VALUES ($1, $2, 'Control', true, 1.0)
	`, armID, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id) VALUES ($1)`, armID)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Set("user_id", adminID.String())
		c.Next()
	})

	handler := handlers.NewAdminHandler(
		nil,
		nil,
		generated.New(db),
		db,
		nil,
		nil,
		service.NewAuditService(db),
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	admin := router.Group("/v1/admin")
	admin.PUT("/experiments/:id/automation-policy", handler.UpdateAdminExperimentAutomationPolicy)

	body := []byte(`{"enabled":true,"auto_start":true,"auto_complete":true,"complete_on_end_time":false,"complete_on_sample_size":true,"complete_on_confidence":true}`)
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+experimentID.String()+"/automation-policy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data handlers.AdminExperiment `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Data.AutomationPolicy.Enabled)
	assert.True(t, resp.Data.AutomationPolicy.AutoStart)
	assert.True(t, resp.Data.AutomationPolicy.AutoComplete)
	assert.False(t, resp.Data.AutomationPolicy.CompleteOnEndTime)
	assert.True(t, resp.Data.AutomationPolicy.CompleteOnSampleSize)
	assert.True(t, resp.Data.AutomationPolicy.CompleteOnConfidence)
	assert.True(t, resp.Data.AutomationPolicy.ManualOverride)
	require.NotNil(t, resp.Data.AutomationPolicy.LockedUntil)
	assert.True(t, resp.Data.AutomationPolicy.LockedUntil.Equal(lockedUntil))
	require.NotNil(t, resp.Data.AutomationPolicy.LockedBy)
	assert.Equal(t, adminID, *resp.Data.AutomationPolicy.LockedBy)
	require.NotNil(t, resp.Data.AutomationPolicy.LockReason)
	assert.Equal(t, "Freeze automation", *resp.Data.AutomationPolicy.LockReason)

	var action string
	var details []byte
	require.NoError(t, db.QueryRow(ctx, `SELECT action, details FROM admin_audit_log WHERE admin_id = $1 ORDER BY created_at DESC LIMIT 1`, adminID).Scan(&action, &details))
	assert.Equal(t, "update_experiment_automation_policy", action)

	var auditDetails map[string]interface{}
	require.NoError(t, json.Unmarshal(details, &auditDetails))
	assert.Equal(t, experimentID.String(), auditDetails["experiment_id"])
	changedFields := auditDetails["changed_fields"].([]interface{})
	assert.Contains(t, changedFields, "enabled")
	assert.Contains(t, changedFields, "auto_start")
	assert.Contains(t, changedFields, "auto_complete")
	assert.Contains(t, changedFields, "complete_on_end_time")
	assert.Contains(t, changedFields, "complete_on_sample_size")
	assert.Contains(t, changedFields, "complete_on_confidence")
	assert.NotContains(t, changedFields, "manual_override")
	assert.NotContains(t, changedFields, "locked_until")
}
