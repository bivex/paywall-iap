package regression

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	service "github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestAdminConfirmExperimentWinnerCompletesRecommendedBanditAndWritesAudits(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE pricing_tiers (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL UNIQUE, description TEXT, monthly_price NUMERIC(10,2), annual_price NUMERIC(10,2), currency CHAR(3) NOT NULL DEFAULT 'USD', features JSONB, is_active BOOLEAN NOT NULL DEFAULT true, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ);
		CREATE TABLE users (id UUID PRIMARY KEY, platform_user_id TEXT UNIQUE NOT NULL, device_id TEXT, platform TEXT NOT NULL, app_version TEXT NOT NULL, email TEXT UNIQUE, role TEXT NOT NULL DEFAULT 'user', ltv NUMERIC(10,2) DEFAULT 0, ltv_updated_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ);
		CREATE TABLE admin_audit_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), admin_id UUID NOT NULL REFERENCES users(id), action TEXT NOT NULL, target_type TEXT NOT NULL, target_user_id UUID, details JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_tests (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL, description TEXT, status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'paused', 'completed')) DEFAULT 'draft', start_at TIMESTAMPTZ, end_at TIMESTAMPTZ, algorithm_type TEXT CHECK (algorithm_type IN ('thompson_sampling', 'ucb', 'epsilon_greedy')), is_bandit BOOLEAN NOT NULL DEFAULT false, min_sample_size INT DEFAULT 100, confidence_threshold NUMERIC(3,2) DEFAULT 0.95, winner_confidence NUMERIC(3,2), automation_policy JSONB NOT NULL DEFAULT '{"enabled": false, "auto_start": false, "auto_complete": false, "complete_on_end_time": true, "complete_on_sample_size": false, "complete_on_confidence": false, "manual_override": false}'::jsonb, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_arms (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, name TEXT NOT NULL, description TEXT, is_control BOOLEAN NOT NULL DEFAULT false, traffic_weight NUMERIC(3,2) NOT NULL DEFAULT 1.0, pricing_tier_id UUID REFERENCES pricing_tiers(id), created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_arm_stats (arm_id UUID PRIMARY KEY REFERENCES ab_test_arms(id) ON DELETE CASCADE, alpha NUMERIC(10,2) NOT NULL DEFAULT 1.0, beta NUMERIC(10,2) NOT NULL DEFAULT 1.0, samples INT NOT NULL DEFAULT 0, conversions INT NOT NULL DEFAULT 0, revenue NUMERIC(15,2) NOT NULL DEFAULT 0, avg_reward NUMERIC(10,4), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_assignments (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE, assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(), expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours'));
		CREATE TABLE experiment_lifecycle_audit_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, actor_type TEXT NOT NULL, actor_id UUID, source TEXT NOT NULL, action TEXT NOT NULL, from_status TEXT NOT NULL, to_status TEXT NOT NULL, idempotency_key TEXT, details JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE UNIQUE INDEX idx_experiment_lifecycle_audit_log_idempotency ON experiment_lifecycle_audit_log(idempotency_key);
		CREATE TABLE experiment_winner_recommendation_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, source TEXT NOT NULL, recommended BOOLEAN NOT NULL DEFAULT FALSE, reason TEXT NOT NULL, winning_arm_id UUID REFERENCES ab_test_arms(id) ON DELETE SET NULL, confidence_percent DOUBLE PRECISION, confidence_threshold_percent DOUBLE PRECISION NOT NULL, observed_samples INT NOT NULL, min_sample_size INT NOT NULL, details JSONB, occurred_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
	`)
	require.NoError(t, err)

	adminID := uuid.New()
	experimentID := uuid.New()
	controlArmID := uuid.New()
	winnerArmID := uuid.New()

	_, err = db.Exec(ctx, `INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role) VALUES ($1, 'admin-user', 'admin-device', 'ios', '1.0.0', 'admin@example.com', 'admin')`, adminID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, name, description, status, algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence) VALUES ($1, 'Confirm winner regression', 'recommended winner path', 'running', 'thompson_sampling', true, 20, 0.95, 0.97)`, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight) VALUES ($1, $2, 'Control', 'Baseline', true, 1.0), ($3, $2, 'Variant Winner', 'Winner candidate', false, 1.0)`, controlArmID, experimentID, winnerArmID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward) VALUES ($1, 5, 9, 12, 4, 40, 3.3333), ($2, 28, 4, 29, 27, 290, 10.0)`, controlArmID, winnerArmID)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Set("user_id", adminID.String())
		c.Next()
	})
	handler := handlers.NewAdminHandler(nil, nil, generated.New(db), db, nil, nil, service.NewAuditService(db), nil, nil, nil, nil, nil)
	router.POST("/v1/admin/experiments/:id/confirm-winner", handler.ConfirmAdminExperimentWinner)
	router.POST("/v1/admin/experiments/:id/hold-for-review", handler.HoldAdminExperimentForReview)

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+experimentID.String()+"/confirm-winner", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data handlers.AdminExperiment `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "completed", resp.Data.Status)
	require.NotNil(t, resp.Data.LatestLifecycleAudit)
	assert.Equal(t, "confirm_recommended_winner", resp.Data.LatestLifecycleAudit.Details["reason"])
	assert.Equal(t, winnerArmID.String(), resp.Data.LatestLifecycleAudit.Details["winning_arm_id"])

	var lifecycleReason, lifecycleWinningArm string
	err = db.QueryRow(ctx, `SELECT details->>'reason', details->>'winning_arm_id' FROM experiment_lifecycle_audit_log WHERE experiment_id = $1 ORDER BY created_at DESC LIMIT 1`, experimentID).Scan(&lifecycleReason, &lifecycleWinningArm)
	require.NoError(t, err)
	assert.Equal(t, "confirm_recommended_winner", lifecycleReason)
	assert.Equal(t, winnerArmID.String(), lifecycleWinningArm)

	var adminAction, adminWinningArm string
	err = db.QueryRow(ctx, `SELECT action, details->>'winning_arm_id' FROM admin_audit_log WHERE admin_id = $1 ORDER BY created_at DESC LIMIT 1`, adminID).Scan(&adminAction, &adminWinningArm)
	require.NoError(t, err)
	assert.Equal(t, "confirm_experiment_winner", adminAction)
	assert.Equal(t, winnerArmID.String(), adminWinningArm)
}

func TestAdminHoldExperimentForReviewPausesRecommendedBanditAndWritesAudits(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE pricing_tiers (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL UNIQUE, description TEXT, monthly_price NUMERIC(10,2), annual_price NUMERIC(10,2), currency CHAR(3) NOT NULL DEFAULT 'USD', features JSONB, is_active BOOLEAN NOT NULL DEFAULT true, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ);
		CREATE TABLE users (id UUID PRIMARY KEY, platform_user_id TEXT UNIQUE NOT NULL, device_id TEXT, platform TEXT NOT NULL, app_version TEXT NOT NULL, email TEXT UNIQUE, role TEXT NOT NULL DEFAULT 'user', ltv NUMERIC(10,2) DEFAULT 0, ltv_updated_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ);
		CREATE TABLE admin_audit_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), admin_id UUID NOT NULL REFERENCES users(id), action TEXT NOT NULL, target_type TEXT NOT NULL, target_user_id UUID, details JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_tests (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL, description TEXT, status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'paused', 'completed')) DEFAULT 'draft', start_at TIMESTAMPTZ, end_at TIMESTAMPTZ, algorithm_type TEXT CHECK (algorithm_type IN ('thompson_sampling', 'ucb', 'epsilon_greedy')), is_bandit BOOLEAN NOT NULL DEFAULT false, min_sample_size INT DEFAULT 100, confidence_threshold NUMERIC(3,2) DEFAULT 0.95, winner_confidence NUMERIC(3,2), automation_policy JSONB NOT NULL DEFAULT '{"enabled": false, "auto_start": false, "auto_complete": false, "complete_on_end_time": true, "complete_on_sample_size": false, "complete_on_confidence": false, "manual_override": false}'::jsonb, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_arms (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, name TEXT NOT NULL, description TEXT, is_control BOOLEAN NOT NULL DEFAULT false, traffic_weight NUMERIC(3,2) NOT NULL DEFAULT 1.0, pricing_tier_id UUID REFERENCES pricing_tiers(id), created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_arm_stats (arm_id UUID PRIMARY KEY REFERENCES ab_test_arms(id) ON DELETE CASCADE, alpha NUMERIC(10,2) NOT NULL DEFAULT 1.0, beta NUMERIC(10,2) NOT NULL DEFAULT 1.0, samples INT NOT NULL DEFAULT 0, conversions INT NOT NULL DEFAULT 0, revenue NUMERIC(15,2) NOT NULL DEFAULT 0, avg_reward NUMERIC(10,4), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_assignments (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE, assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(), expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours'));
		CREATE TABLE experiment_lifecycle_audit_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, actor_type TEXT NOT NULL, actor_id UUID, source TEXT NOT NULL, action TEXT NOT NULL, from_status TEXT NOT NULL, to_status TEXT NOT NULL, idempotency_key TEXT, details JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE UNIQUE INDEX idx_experiment_lifecycle_audit_log_idempotency ON experiment_lifecycle_audit_log(idempotency_key);
		CREATE TABLE experiment_winner_recommendation_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, source TEXT NOT NULL, recommended BOOLEAN NOT NULL DEFAULT FALSE, reason TEXT NOT NULL, winning_arm_id UUID REFERENCES ab_test_arms(id) ON DELETE SET NULL, confidence_percent DOUBLE PRECISION, confidence_threshold_percent DOUBLE PRECISION NOT NULL, observed_samples INT NOT NULL, min_sample_size INT NOT NULL, details JSONB, occurred_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
	`)
	require.NoError(t, err)

	adminID := uuid.New()
	experimentID := uuid.New()
	controlArmID := uuid.New()
	winnerArmID := uuid.New()

	_, err = db.Exec(ctx, `INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role) VALUES ($1, 'admin-user', 'admin-device', 'ios', '1.0.0', 'admin@example.com', 'admin')`, adminID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, name, description, status, algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence) VALUES ($1, 'Hold for review regression', 'recommended winner path', 'running', 'thompson_sampling', true, 20, 0.95, 0.97)`, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight) VALUES ($1, $2, 'Control', 'Baseline', true, 1.0), ($3, $2, 'Variant Winner', 'Winner candidate', false, 1.0)`, controlArmID, experimentID, winnerArmID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward) VALUES ($1, 5, 9, 12, 4, 40, 3.3333), ($2, 28, 4, 29, 27, 290, 10.0)`, controlArmID, winnerArmID)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Set("user_id", adminID.String())
		c.Next()
	})
	handler := handlers.NewAdminHandler(nil, nil, generated.New(db), db, nil, nil, service.NewAuditService(db), nil, nil, nil, nil, nil)
	router.POST("/v1/admin/experiments/:id/hold-for-review", handler.HoldAdminExperimentForReview)

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+experimentID.String()+"/hold-for-review", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data handlers.AdminExperiment `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "paused", resp.Data.Status)
	assert.True(t, resp.Data.AutomationPolicy.ManualOverride)
	require.NotNil(t, resp.Data.AutomationPolicy.LockReason)
	assert.Equal(t, "Hold recommended winner for review", *resp.Data.AutomationPolicy.LockReason)
	require.NotNil(t, resp.Data.LatestLifecycleAudit)
	assert.Equal(t, "hold_recommended_winner_review", resp.Data.LatestLifecycleAudit.Details["reason"])

	var lifecycleReason, lifecycleWinningArm string
	err = db.QueryRow(ctx, `SELECT details->>'reason', details->>'winning_arm_id' FROM experiment_lifecycle_audit_log WHERE experiment_id = $1 ORDER BY created_at DESC LIMIT 1`, experimentID).Scan(&lifecycleReason, &lifecycleWinningArm)
	require.NoError(t, err)
	assert.Equal(t, "hold_recommended_winner_review", lifecycleReason)
	assert.Equal(t, winnerArmID.String(), lifecycleWinningArm)

	var adminAction, adminLockReason string
	err = db.QueryRow(ctx, `SELECT action, details->>'lock_reason' FROM admin_audit_log WHERE admin_id = $1 ORDER BY created_at DESC LIMIT 1`, adminID).Scan(&adminAction, &adminLockReason)
	require.NoError(t, err)
	assert.Equal(t, "hold_experiment_for_review", adminAction)
	assert.Equal(t, "Hold recommended winner for review", adminLockReason)
}

func TestAdminConfirmExperimentWinnerRejectsLockedExperiment(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE pricing_tiers (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL UNIQUE, description TEXT, monthly_price NUMERIC(10,2), annual_price NUMERIC(10,2), currency CHAR(3) NOT NULL DEFAULT 'USD', features JSONB, is_active BOOLEAN NOT NULL DEFAULT true, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ);
		CREATE TABLE users (id UUID PRIMARY KEY, platform_user_id TEXT UNIQUE NOT NULL, device_id TEXT, platform TEXT NOT NULL, app_version TEXT NOT NULL, email TEXT UNIQUE, role TEXT NOT NULL DEFAULT 'user', ltv NUMERIC(10,2) DEFAULT 0, ltv_updated_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ);
		CREATE TABLE admin_audit_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), admin_id UUID NOT NULL REFERENCES users(id), action TEXT NOT NULL, target_type TEXT NOT NULL, target_user_id UUID, details JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_tests (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL, description TEXT, status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'paused', 'completed')) DEFAULT 'draft', start_at TIMESTAMPTZ, end_at TIMESTAMPTZ, algorithm_type TEXT CHECK (algorithm_type IN ('thompson_sampling', 'ucb', 'epsilon_greedy')), is_bandit BOOLEAN NOT NULL DEFAULT false, min_sample_size INT DEFAULT 100, confidence_threshold NUMERIC(3,2) DEFAULT 0.95, winner_confidence NUMERIC(3,2), automation_policy JSONB NOT NULL DEFAULT '{"enabled": false, "auto_start": false, "auto_complete": false, "complete_on_end_time": true, "complete_on_sample_size": false, "complete_on_confidence": false, "manual_override": false}'::jsonb, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_arms (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, name TEXT NOT NULL, description TEXT, is_control BOOLEAN NOT NULL DEFAULT false, traffic_weight NUMERIC(3,2) NOT NULL DEFAULT 1.0, pricing_tier_id UUID REFERENCES pricing_tiers(id), created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_arm_stats (arm_id UUID PRIMARY KEY REFERENCES ab_test_arms(id) ON DELETE CASCADE, alpha NUMERIC(10,2) NOT NULL DEFAULT 1.0, beta NUMERIC(10,2) NOT NULL DEFAULT 1.0, samples INT NOT NULL DEFAULT 0, conversions INT NOT NULL DEFAULT 0, revenue NUMERIC(15,2) NOT NULL DEFAULT 0, avg_reward NUMERIC(10,4), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_assignments (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE, assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(), expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours'));
		CREATE TABLE experiment_lifecycle_audit_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, actor_type TEXT NOT NULL, actor_id UUID, source TEXT NOT NULL, action TEXT NOT NULL, from_status TEXT NOT NULL, to_status TEXT NOT NULL, idempotency_key TEXT, details JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE UNIQUE INDEX idx_experiment_lifecycle_audit_log_idempotency ON experiment_lifecycle_audit_log(idempotency_key);
		CREATE TABLE experiment_winner_recommendation_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, source TEXT NOT NULL, recommended BOOLEAN NOT NULL DEFAULT FALSE, reason TEXT NOT NULL, winning_arm_id UUID REFERENCES ab_test_arms(id) ON DELETE SET NULL, confidence_percent DOUBLE PRECISION, confidence_threshold_percent DOUBLE PRECISION NOT NULL, observed_samples INT NOT NULL, min_sample_size INT NOT NULL, details JSONB, occurred_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
	`)
	require.NoError(t, err)

	adminID := uuid.New()
	experimentID := uuid.New()
	controlArmID := uuid.New()
	winnerArmID := uuid.New()

	_, err = db.Exec(ctx, `INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role) VALUES ($1, 'admin-user', 'admin-device', 'ios', '1.0.0', 'admin@example.com', 'admin')`, adminID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, name, description, status, algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence, automation_policy) VALUES ($1, 'Locked confirm winner regression', 'manual override blocks confirmation', 'running', 'thompson_sampling', true, 20, 0.95, 0.97, '{"enabled": true, "auto_start": true, "auto_complete": true, "complete_on_end_time": true, "complete_on_sample_size": false, "complete_on_confidence": false, "manual_override": true, "lock_reason": "Operator hold"}'::jsonb)`, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight) VALUES ($1, $2, 'Control', 'Baseline', true, 1.0), ($3, $2, 'Variant Winner', 'Winner candidate', false, 1.0)`, controlArmID, experimentID, winnerArmID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward) VALUES ($1, 5, 9, 12, 4, 40, 3.3333), ($2, 28, 4, 29, 27, 290, 10.0)`, controlArmID, winnerArmID)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Set("user_id", adminID.String())
		c.Next()
	})
	handler := handlers.NewAdminHandler(nil, nil, generated.New(db), db, nil, nil, service.NewAuditService(db), nil, nil, nil, nil, nil)
	router.POST("/v1/admin/experiments/:id/confirm-winner", handler.ConfirmAdminExperimentWinner)

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+experimentID.String()+"/confirm-winner", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var status string
	err = db.QueryRow(ctx, `SELECT status FROM ab_tests WHERE id = $1`, experimentID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "running", status)

	var lifecycleRows int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM experiment_lifecycle_audit_log WHERE experiment_id = $1`, experimentID).Scan(&lifecycleRows)
	require.NoError(t, err)
	assert.Zero(t, lifecycleRows)

	var adminRows int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM admin_audit_log WHERE admin_id = $1`, adminID).Scan(&adminRows)
	require.NoError(t, err)
	assert.Zero(t, adminRows)
}

func TestAdminHoldExperimentForReviewKeepsPausedExperimentPaused(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE pricing_tiers (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL UNIQUE, description TEXT, monthly_price NUMERIC(10,2), annual_price NUMERIC(10,2), currency CHAR(3) NOT NULL DEFAULT 'USD', features JSONB, is_active BOOLEAN NOT NULL DEFAULT true, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ);
		CREATE TABLE users (id UUID PRIMARY KEY, platform_user_id TEXT UNIQUE NOT NULL, device_id TEXT, platform TEXT NOT NULL, app_version TEXT NOT NULL, email TEXT UNIQUE, role TEXT NOT NULL DEFAULT 'user', ltv NUMERIC(10,2) DEFAULT 0, ltv_updated_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ);
		CREATE TABLE admin_audit_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), admin_id UUID NOT NULL REFERENCES users(id), action TEXT NOT NULL, target_type TEXT NOT NULL, target_user_id UUID, details JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_tests (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL, description TEXT, status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'paused', 'completed')) DEFAULT 'draft', start_at TIMESTAMPTZ, end_at TIMESTAMPTZ, algorithm_type TEXT CHECK (algorithm_type IN ('thompson_sampling', 'ucb', 'epsilon_greedy')), is_bandit BOOLEAN NOT NULL DEFAULT false, min_sample_size INT DEFAULT 100, confidence_threshold NUMERIC(3,2) DEFAULT 0.95, winner_confidence NUMERIC(3,2), automation_policy JSONB NOT NULL DEFAULT '{"enabled": false, "auto_start": false, "auto_complete": false, "complete_on_end_time": true, "complete_on_sample_size": false, "complete_on_confidence": false, "manual_override": false}'::jsonb, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_arms (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, name TEXT NOT NULL, description TEXT, is_control BOOLEAN NOT NULL DEFAULT false, traffic_weight NUMERIC(3,2) NOT NULL DEFAULT 1.0, pricing_tier_id UUID REFERENCES pricing_tiers(id), created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_arm_stats (arm_id UUID PRIMARY KEY REFERENCES ab_test_arms(id) ON DELETE CASCADE, alpha NUMERIC(10,2) NOT NULL DEFAULT 1.0, beta NUMERIC(10,2) NOT NULL DEFAULT 1.0, samples INT NOT NULL DEFAULT 0, conversions INT NOT NULL DEFAULT 0, revenue NUMERIC(15,2) NOT NULL DEFAULT 0, avg_reward NUMERIC(10,4), updated_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE ab_test_assignments (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE, assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(), expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours'));
		CREATE TABLE experiment_lifecycle_audit_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, actor_type TEXT NOT NULL, actor_id UUID, source TEXT NOT NULL, action TEXT NOT NULL, from_status TEXT NOT NULL, to_status TEXT NOT NULL, idempotency_key TEXT, details JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE UNIQUE INDEX idx_experiment_lifecycle_audit_log_idempotency ON experiment_lifecycle_audit_log(idempotency_key);
		CREATE TABLE experiment_winner_recommendation_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, source TEXT NOT NULL, recommended BOOLEAN NOT NULL DEFAULT FALSE, reason TEXT NOT NULL, winning_arm_id UUID REFERENCES ab_test_arms(id) ON DELETE SET NULL, confidence_percent DOUBLE PRECISION, confidence_threshold_percent DOUBLE PRECISION NOT NULL, observed_samples INT NOT NULL, min_sample_size INT NOT NULL, details JSONB, occurred_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
	`)
	require.NoError(t, err)

	adminID := uuid.New()
	experimentID := uuid.New()
	controlArmID := uuid.New()
	winnerArmID := uuid.New()

	_, err = db.Exec(ctx, `INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role) VALUES ($1, 'admin-user', 'admin-device', 'ios', '1.0.0', 'admin@example.com', 'admin')`, adminID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, name, description, status, algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence) VALUES ($1, 'Paused hold for review regression', 'lock metadata only path', 'paused', 'thompson_sampling', true, 20, 0.95, 0.97)`, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight) VALUES ($1, $2, 'Control', 'Baseline', true, 1.0), ($3, $2, 'Variant Winner', 'Winner candidate', false, 1.0)`, controlArmID, experimentID, winnerArmID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward) VALUES ($1, 5, 9, 12, 4, 40, 3.3333), ($2, 28, 4, 29, 27, 290, 10.0)`, controlArmID, winnerArmID)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Set("user_id", adminID.String())
		c.Next()
	})
	handler := handlers.NewAdminHandler(nil, nil, generated.New(db), db, nil, nil, service.NewAuditService(db), nil, nil, nil, nil, nil)
	router.POST("/v1/admin/experiments/:id/hold-for-review", handler.HoldAdminExperimentForReview)

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+experimentID.String()+"/hold-for-review", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data handlers.AdminExperiment `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "paused", resp.Data.Status)
	assert.True(t, resp.Data.AutomationPolicy.ManualOverride)
	require.NotNil(t, resp.Data.AutomationPolicy.LockReason)
	assert.Equal(t, "Hold recommended winner for review", *resp.Data.AutomationPolicy.LockReason)
	assert.Nil(t, resp.Data.LatestLifecycleAudit)

	var lifecycleRows int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM experiment_lifecycle_audit_log WHERE experiment_id = $1`, experimentID).Scan(&lifecycleRows)
	require.NoError(t, err)
	assert.Zero(t, lifecycleRows)

	var adminAction, adminLockReason string
	err = db.QueryRow(ctx, `SELECT action, details->>'lock_reason' FROM admin_audit_log WHERE admin_id = $1 ORDER BY created_at DESC LIMIT 1`, adminID).Scan(&adminAction, &adminLockReason)
	require.NoError(t, err)
	assert.Equal(t, "hold_experiment_for_review", adminAction)
	assert.Equal(t, "Hold recommended winner for review", adminLockReason)
}
