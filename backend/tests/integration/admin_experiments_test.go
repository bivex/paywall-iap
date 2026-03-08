package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func TestAdminExperimentsHandler(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE pricing_tiers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			monthly_price NUMERIC(10,2),
			annual_price NUMERIC(10,2),
			currency CHAR(3) NOT NULL DEFAULT 'USD',
			features JSONB,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ
		);
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
				automation_policy JSONB NOT NULL DEFAULT '{"enabled": false, "auto_start": false, "auto_complete": false, "complete_on_end_time": true, "complete_on_sample_size": false, "complete_on_confidence": false, "manual_override": false}'::jsonb,
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
			pricing_tier_id UUID REFERENCES pricing_tiers(id),
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE ab_test_arm_stats (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			arm_id UUID NOT NULL UNIQUE REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			alpha NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			beta NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			samples INT NOT NULL DEFAULT 0,
			conversions INT NOT NULL DEFAULT 0,
			revenue NUMERIC(15,2) NOT NULL DEFAULT 0.0,
			avg_reward NUMERIC(10,4),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			CHECK (alpha > 0),
			CHECK (beta > 0),
			CHECK (samples >= 0),
			CHECK (conversions >= 0),
			CHECK (conversions <= samples)
		);
		CREATE TABLE ab_test_assignments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours')
			);
			CREATE TABLE experiment_lifecycle_audit_log (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
				actor_type TEXT NOT NULL,
				actor_id UUID,
				source TEXT NOT NULL,
				action TEXT NOT NULL,
				from_status TEXT NOT NULL,
				to_status TEXT NOT NULL,
				idempotency_key TEXT,
				details JSONB,
				created_at TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE UNIQUE INDEX idx_experiment_lifecycle_audit_log_idempotency
				ON experiment_lifecycle_audit_log(idempotency_key);
			CREATE TABLE experiment_winner_recommendation_log (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
				source TEXT NOT NULL,
				recommended BOOLEAN NOT NULL DEFAULT FALSE,
				reason TEXT NOT NULL,
				winning_arm_id UUID REFERENCES ab_test_arms(id) ON DELETE SET NULL,
				confidence_percent DOUBLE PRECISION,
				confidence_threshold_percent DOUBLE PRECISION NOT NULL,
				observed_samples INT NOT NULL,
				min_sample_size INT NOT NULL,
				details JSONB,
				occurred_at TIMESTAMPTZ NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE TABLE admin_audit_log (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				admin_id UUID NOT NULL REFERENCES users(id),
				action TEXT NOT NULL,
				target_type TEXT NOT NULL,
				target_user_id UUID,
				details JSONB,
				created_at TIMESTAMPTZ NOT NULL DEFAULT now()
			);`)
	require.NoError(t, err)

	adminID := uuid.New()
	_, err = db.Exec(ctx,
		`INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email, role)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		adminID,
		"admin-user",
		"admin-device",
		"ios",
		"1.0.0",
		"admin@example.com",
		"admin",
	)
	require.NoError(t, err)

	primaryTierID := uuid.New()
	upsellTierID := uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO pricing_tiers (id, name, description, monthly_price, annual_price, currency, features, is_active)
		VALUES
			($1, 'Pro', 'Primary paid tier', 9.99, 99.99, 'USD', '["Unlimited access"]'::jsonb, TRUE),
			($2, 'Plus', 'Upsell paid tier', 19.99, 199.99, 'USD', '["Priority support"]'::jsonb, TRUE)
	`, primaryTierID, upsellTierID)
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
	admin.GET("/experiments", handler.ListAdminExperiments)
	admin.POST("/experiments", handler.CreateAdminExperiment)
	admin.PUT("/experiments/:id", handler.UpdateAdminExperiment)
	admin.PUT("/experiments/:id/arms/pricing-tiers", handler.UpdateAdminExperimentArmPricingTiers)
	admin.POST("/experiments/:id/confirm-winner", handler.ConfirmAdminExperimentWinner)
	admin.POST("/experiments/:id/hold-for-review", handler.HoldAdminExperimentForReview)
	admin.GET("/experiments/:id/lifecycle-audit", handler.GetAdminExperimentLifecycleAuditHistory)
	admin.GET("/experiments/:id/winner-recommendation-audit", handler.GetAdminExperimentWinnerRecommendationAuditHistory)
	admin.POST("/experiments/:id/pause", handler.PauseAdminExperiment)
	admin.POST("/experiments/:id/resume", handler.ResumeAdminExperiment)
	admin.POST("/experiments/:id/complete", handler.CompleteAdminExperiment)

	var runningExperiment handlers.AdminExperiment
	var draftExperiment handlers.AdminExperiment
	missingTierID := uuid.New()

	t.Run("GET returns empty list before creation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Empty(t, resp.Data)
	})

	t.Run("POST creates a db-backed experiment with arms", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{
			"name":"Pricing homepage CTA",
			"description":"Compare control vs annual emphasis",
			"status":"running",
			"algorithm_type":"thompson_sampling",
			"is_bandit":true,
			"min_sample_size":200,
			"confidence_threshold_percent":95,
				"automation_policy":{"enabled":true,"auto_start":true,"auto_complete":true,"complete_on_sample_size":true},
			"arms":[
				{"name":"Control","description":"Baseline paywall","is_control":true,"traffic_weight":1},
					{"name":"Variant A","description":"Annual plan emphasis","is_control":false,"traffic_weight":1,"pricing_tier_id":%q}
			]
			}`, primaryTierID.String()))
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "Pricing homepage CTA", resp.Data.Name)
		assert.Equal(t, "running", resp.Data.Status)
		assert.True(t, resp.Data.IsBandit)
		assert.InDelta(t, 95, resp.Data.ConfidenceThresholdPercent, 0.001)
		assert.True(t, resp.Data.AutomationPolicy.Enabled)
		assert.True(t, resp.Data.AutomationPolicy.AutoStart)
		assert.True(t, resp.Data.AutomationPolicy.AutoComplete)
		assert.True(t, resp.Data.AutomationPolicy.CompleteOnSampleSize)
		assert.False(t, resp.Data.AutomationPolicy.ManualOverride)
		assert.Len(t, resp.Data.Arms, 2)
		require.NotNil(t, resp.Data.Arms[1].PricingTierID)
		assert.Equal(t, primaryTierID, *resp.Data.Arms[1].PricingTierID)
		assert.Equal(t, 2, resp.Data.ArmCount)
		assert.Equal(t, 0, resp.Data.TotalSamples)
		runningExperiment = resp.Data
	})

	t.Run("POST creates a draft experiment", func(t *testing.T) {
		body := []byte(`{
			"name":"Draft onboarding test",
			"description":"Prepare a staged rollout",
			"status":"draft",
			"algorithm_type":"thompson_sampling",
			"is_bandit":true,
			"min_sample_size":150,
			"confidence_threshold_percent":90,
			"arms":[
				{"name":"Control","description":"Current onboarding","is_control":true,"traffic_weight":1},
				{"name":"Variant B","description":"Shorter onboarding","is_control":false,"traffic_weight":1}
			]
		}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "draft", resp.Data.Status)
		assert.False(t, resp.Data.AutomationPolicy.Enabled)
		assert.True(t, resp.Data.AutomationPolicy.CompleteOnEndTime)
		draftExperiment = resp.Data
	})

	t.Run("GET returns created experiment", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 2)
		assert.Equal(t, "Draft onboarding test", resp.Data[0].Name)
		assert.Len(t, resp.Data[0].Arms, 2)

		var foundRunning bool
		for _, experiment := range resp.Data {
			if experiment.ID != runningExperiment.ID {
				continue
			}
			foundRunning = true
			require.NotNil(t, experiment.Arms[1].PricingTierID)
			assert.Equal(t, primaryTierID, *experiment.Arms[1].PricingTierID)
		}
		assert.True(t, foundRunning)
	})

	t.Run("POST rejects nonexistent pricing tier linkage", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{
				"name":"Broken pricing linkage",
				"status":"draft",
				"is_bandit":false,
				"min_sample_size":100,
				"confidence_threshold_percent":95,
				"arms":[
					{"name":"Control","is_control":true,"traffic_weight":1},
					{"name":"Variant","is_control":false,"traffic_weight":1,"pricing_tier_id":%q}
				]
			}`, missingTierID.String()))
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("GET returns latest lifecycle audit reason when present", func(t *testing.T) {
		_, err = db.Exec(ctx, `
			INSERT INTO experiment_lifecycle_audit_log (
				experiment_id, actor_type, source, action, from_status, to_status, idempotency_key, details
			) VALUES ($1, 'system', 'experiment_automation_reconciler', 'status_transition', 'draft', 'running', $2, '{"reason":"auto_start"}'::jsonb)
		`, draftExperiment.ID, "experiment:"+draftExperiment.ID.String()+":running")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		var found bool
		for _, experiment := range resp.Data {
			if experiment.ID != draftExperiment.ID {
				continue
			}
			found = true
			require.NotNil(t, experiment.LatestLifecycleAudit)
			assert.Equal(t, "system", experiment.LatestLifecycleAudit.ActorType)
			assert.Equal(t, "experiment_automation_reconciler", experiment.LatestLifecycleAudit.Source)
			assert.Equal(t, "draft", experiment.LatestLifecycleAudit.FromStatus)
			assert.Equal(t, "running", experiment.LatestLifecycleAudit.ToStatus)
			assert.Equal(t, "auto_start", experiment.LatestLifecycleAudit.Details["reason"])
		}
		assert.True(t, found)
	})

	t.Run("GET returns winner recommendation for eligible bandit experiments", func(t *testing.T) {
		var controlArmID uuid.UUID
		var variantArmID uuid.UUID
		for _, arm := range runningExperiment.Arms {
			switch arm.Name {
			case "Control":
				controlArmID = arm.ID
			case "Variant A":
				variantArmID = arm.ID
			}
		}
		require.NotEqual(t, uuid.Nil, controlArmID)
		require.NotEqual(t, uuid.Nil, variantArmID)

		_, err = db.Exec(ctx, `
				UPDATE ab_tests
				SET min_sample_size = 20,
				    confidence_threshold = 0.95,
				    winner_confidence = 0.97
				WHERE id = $1`, runningExperiment.ID)
		require.NoError(t, err)

		_, err = db.Exec(ctx, `
				INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward)
				VALUES
					($1, 5, 9, 12, 4, 40, 3.3333),
					($2, 28, 4, 29, 27, 290, 10.0)
				ON CONFLICT (arm_id) DO UPDATE
				SET alpha = EXCLUDED.alpha,
				    beta = EXCLUDED.beta,
				    samples = EXCLUDED.samples,
				    conversions = EXCLUDED.conversions,
				    revenue = EXCLUDED.revenue,
				    avg_reward = EXCLUDED.avg_reward,
				    updated_at = now()`, controlArmID, variantArmID)
		require.NoError(t, err)

		_, err = db.Exec(ctx, `DELETE FROM experiment_winner_recommendation_log WHERE experiment_id = $1`, runningExperiment.ID)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		var found bool
		for _, experiment := range resp.Data {
			if experiment.ID != runningExperiment.ID {
				continue
			}
			found = true
			require.NotNil(t, experiment.WinnerRecommendation)
			assert.True(t, experiment.WinnerRecommendation.Recommended)
			assert.Equal(t, "recommend_winner", experiment.WinnerRecommendation.Reason)
			require.NotNil(t, experiment.WinnerRecommendation.WinningArmID)
			assert.Equal(t, variantArmID, *experiment.WinnerRecommendation.WinningArmID)
			require.NotNil(t, experiment.WinnerRecommendation.WinningArmName)
			assert.Equal(t, "Variant A", *experiment.WinnerRecommendation.WinningArmName)
			require.NotNil(t, experiment.WinnerRecommendation.ConfidencePercent)
			assert.InDelta(t, 97.0, *experiment.WinnerRecommendation.ConfidencePercent, 0.001)
		}
		assert.True(t, found)

		var recommendationRows int
		err = db.QueryRow(ctx, `SELECT COUNT(*)::int FROM experiment_winner_recommendation_log WHERE experiment_id = $1`, runningExperiment.ID).Scan(&recommendationRows)
		require.NoError(t, err)
		assert.Equal(t, 1, recommendationRows)
	})

	t.Run("GET winner recommendation audit history returns newest-first entries", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments/"+runningExperiment.ID.String()+"/winner-recommendation-audit", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperimentWinnerRecommendationAudit `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotEmpty(t, resp.Data)
		assert.Equal(t, "admin_experiments_list", resp.Data[0].Source)
		assert.Equal(t, "recommend_winner", resp.Data[0].Reason)
		assert.True(t, resp.Data[0].Recommended)
		require.NotNil(t, resp.Data[0].WinningArmID)
		assert.Equal(t, runningExperiment.Arms[1].ID, *resp.Data[0].WinningArmID)
	})

	t.Run("POST confirm-winner completes recommended experiment and writes audits", func(t *testing.T) {
		confirmExperimentID := uuid.New()
		confirmControlArmID := uuid.New()
		confirmVariantArmID := uuid.New()

		_, err = db.Exec(ctx, `
			INSERT INTO ab_tests (
				id, name, description, status, algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence, automation_policy
			) VALUES (
				$1, 'Confirm winner regression', 'recommended winner path', 'running', 'thompson_sampling', true, 20, 0.95, 0.97,
				'{"enabled":true,"auto_start":true,"auto_complete":true,"complete_on_end_time":true,"complete_on_sample_size":false,"complete_on_confidence":false,"manual_override":false}'::jsonb
			)
		`, confirmExperimentID)
		require.NoError(t, err)
		_, err = db.Exec(ctx, `
			INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight)
			VALUES
				($1, $3, 'Control', 'Baseline', true, 1.0),
				($2, $3, 'Variant Winner', 'Winner candidate', false, 1.0)
		`, confirmControlArmID, confirmVariantArmID, confirmExperimentID)
		require.NoError(t, err)
		_, err = db.Exec(ctx, `
			INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward)
			VALUES
				($1, 5, 9, 12, 4, 40, 3.3333),
				($2, 28, 4, 29, 27, 290, 10.0)
		`, confirmControlArmID, confirmVariantArmID)
		require.NoError(t, err)
		defer func() {
			_, cleanupErr := db.Exec(ctx, `DELETE FROM ab_tests WHERE id = $1`, confirmExperimentID)
			require.NoError(t, cleanupErr)
		}()

		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+confirmExperimentID.String()+"/confirm-winner", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "completed", resp.Data.Status)
		require.NotNil(t, resp.Data.LatestLifecycleAudit)
		assert.Equal(t, "completed", resp.Data.LatestLifecycleAudit.ToStatus)
		assert.Equal(t, "confirm_recommended_winner", resp.Data.LatestLifecycleAudit.Details["reason"])
		assert.Equal(t, "Variant Winner", resp.Data.LatestLifecycleAudit.Details["winning_arm_name"])

		var lifecycleReason string
		var lifecycleWinningArm string
		err = db.QueryRow(ctx, `
			SELECT details->>'reason', details->>'winning_arm_id'
			FROM experiment_lifecycle_audit_log
			WHERE experiment_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`, confirmExperimentID).Scan(&lifecycleReason, &lifecycleWinningArm)
		require.NoError(t, err)
		assert.Equal(t, "confirm_recommended_winner", lifecycleReason)
		assert.Equal(t, confirmVariantArmID.String(), lifecycleWinningArm)

		var adminAction string
		var adminWinningArm string
		err = db.QueryRow(ctx, `
			SELECT action, details->>'winning_arm_id'
			FROM admin_audit_log
			WHERE admin_id = $1 AND details->>'experiment_id' = $2
			ORDER BY created_at DESC
			LIMIT 1
		`, adminID, confirmExperimentID.String()).Scan(&adminAction, &adminWinningArm)
		require.NoError(t, err)
		assert.Equal(t, "confirm_experiment_winner", adminAction)
		assert.Equal(t, confirmVariantArmID.String(), adminWinningArm)
	})

	t.Run("POST hold-for-review pauses experiment and enables manual override with audits", func(t *testing.T) {
		holdExperimentID := uuid.New()
		holdControlArmID := uuid.New()
		holdVariantArmID := uuid.New()

		_, err = db.Exec(ctx, `
			INSERT INTO ab_tests (
				id, name, description, status, algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence, automation_policy
			) VALUES (
				$1, 'Hold for review regression', 'pause and lock for review', 'running', 'thompson_sampling', true, 20, 0.95, 0.97,
				'{"enabled":true,"auto_start":true,"auto_complete":true,"complete_on_end_time":true,"complete_on_sample_size":false,"complete_on_confidence":false,"manual_override":false}'::jsonb
			)
		`, holdExperimentID)
		require.NoError(t, err)
		_, err = db.Exec(ctx, `
			INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight)
			VALUES
				($1, $3, 'Control', 'Baseline', true, 1.0),
				($2, $3, 'Variant Winner', 'Winner candidate', false, 1.0)
		`, holdControlArmID, holdVariantArmID, holdExperimentID)
		require.NoError(t, err)
		_, err = db.Exec(ctx, `
			INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward)
			VALUES
				($1, 5, 9, 12, 4, 40, 3.3333),
				($2, 28, 4, 29, 27, 290, 10.0)
		`, holdControlArmID, holdVariantArmID)
		require.NoError(t, err)
		defer func() {
			_, cleanupErr := db.Exec(ctx, `DELETE FROM ab_tests WHERE id = $1`, holdExperimentID)
			require.NoError(t, cleanupErr)
		}()

		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+holdExperimentID.String()+"/hold-for-review", nil)
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

		var lifecycleReason string
		var lifecycleWinningArm string
		err = db.QueryRow(ctx, `
			SELECT details->>'reason', details->>'winning_arm_id'
			FROM experiment_lifecycle_audit_log
			WHERE experiment_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`, holdExperimentID).Scan(&lifecycleReason, &lifecycleWinningArm)
		require.NoError(t, err)
		assert.Equal(t, "hold_recommended_winner_review", lifecycleReason)
		assert.Equal(t, holdVariantArmID.String(), lifecycleWinningArm)

		var adminAction string
		var adminLockReason string
		err = db.QueryRow(ctx, `
			SELECT action, details->>'lock_reason'
			FROM admin_audit_log
			WHERE admin_id = $1 AND details->>'experiment_id' = $2
			ORDER BY created_at DESC
			LIMIT 1
		`, adminID, holdExperimentID.String()).Scan(&adminAction, &adminLockReason)
		require.NoError(t, err)
		assert.Equal(t, "hold_experiment_for_review", adminAction)
		assert.Equal(t, "Hold recommended winner for review", adminLockReason)
	})

	t.Run("GET lifecycle audit history returns newest-first entries", func(t *testing.T) {
		_, err := db.Exec(ctx, `DELETE FROM experiment_lifecycle_audit_log WHERE experiment_id = $1`, draftExperiment.ID)
		require.NoError(t, err)

		_, err = db.Exec(ctx, `
			INSERT INTO experiment_lifecycle_audit_log (
				experiment_id, actor_type, source, action, from_status, to_status, idempotency_key, details, created_at
			) VALUES
			($1, 'system', 'experiment_automation_reconciler', 'status_transition', 'draft', 'running', $2, '{"reason":"auto_start"}'::jsonb, '2026-01-03T10:00:00Z'),
			($1, 'admin', 'admin_experiments_api', 'status_transition', 'running', 'paused', $3, '{"reason":"manual_paused"}'::jsonb, '2026-01-04T10:00:00Z')
		`, draftExperiment.ID, "experiment:"+draftExperiment.ID.String()+":seed-running", "experiment:"+draftExperiment.ID.String()+":seed-paused")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments/"+draftExperiment.ID.String()+"/lifecycle-audit", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperimentLifecycleAudit `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 2)
		assert.Equal(t, "admin", resp.Data[0].ActorType)
		assert.Equal(t, "manual_paused", resp.Data[0].Details["reason"])
		assert.Equal(t, "system", resp.Data[1].ActorType)
		assert.Equal(t, "auto_start", resp.Data[1].Details["reason"])
	})

	t.Run("PUT rejects invalid draft metadata payload", func(t *testing.T) {
		body := []byte(`{
			"name":"Draft onboarding test",
			"description":"Prepare a staged rollout",
			"algorithm_type":"thompson_sampling",
			"is_bandit":true,
			"min_sample_size":150,
			"confidence_threshold_percent":101
		}`)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+draftExperiment.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("PUT updates a draft experiment metadata", func(t *testing.T) {
		body := []byte(`{
			"name":"Draft onboarding test v2",
			"description":"Prepare a staged rollout with classic allocation",
			"algorithm_type":"epsilon_greedy",
			"is_bandit":false,
			"min_sample_size":220,
			"confidence_threshold_percent":92,
			"start_at":"2026-01-05T10:00:00Z",
				"end_at":"2026-01-12T10:00:00Z",
				"automation_policy":{"enabled":true,"auto_complete":true,"complete_on_confidence":true,"manual_override":true}
		}`)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+draftExperiment.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "Draft onboarding test v2", resp.Data.Name)
		assert.Equal(t, "Prepare a staged rollout with classic allocation", resp.Data.Description)
		assert.False(t, resp.Data.IsBandit)
		assert.Nil(t, resp.Data.AlgorithmType)
		assert.Equal(t, 220, resp.Data.MinSampleSize)
		assert.InDelta(t, 92, resp.Data.ConfidenceThresholdPercent, 0.001)
		assert.True(t, resp.Data.AutomationPolicy.Enabled)
		assert.True(t, resp.Data.AutomationPolicy.AutoComplete)
		assert.True(t, resp.Data.AutomationPolicy.CompleteOnConfidence)
		assert.True(t, resp.Data.AutomationPolicy.ManualOverride)
		require.NotNil(t, resp.Data.StartAt)
		require.NotNil(t, resp.Data.EndAt)
		assert.Equal(t, "draft", resp.Data.Status)
		assert.Len(t, resp.Data.Arms, 2)
		draftExperiment = resp.Data
	})

	t.Run("PUT updates draft arm pricing tier linkage", func(t *testing.T) {
		var controlArmID uuid.UUID
		var variantArmID uuid.UUID
		for _, arm := range draftExperiment.Arms {
			switch arm.Name {
			case "Control":
				controlArmID = arm.ID
			case "Variant B":
				variantArmID = arm.ID
			}
		}
		require.NotEqual(t, uuid.Nil, controlArmID)
		require.NotEqual(t, uuid.Nil, variantArmID)

		body := []byte(fmt.Sprintf(`{
				"arms":[
					{"arm_id":%q,"pricing_tier_id":%q},
					{"arm_id":%q,"pricing_tier_id":%q}
				]
			}`, controlArmID.String(), primaryTierID.String(), variantArmID.String(), upsellTierID.String()))
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+draftExperiment.ID.String()+"/arms/pricing-tiers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		links := make(map[string]uuid.UUID)
		for _, arm := range resp.Data.Arms {
			if arm.PricingTierID != nil {
				links[arm.Name] = *arm.PricingTierID
			}
		}
		assert.Equal(t, primaryTierID, links["Control"])
		assert.Equal(t, upsellTierID, links["Variant B"])
		draftExperiment = resp.Data
	})

	t.Run("PUT rejects nonexistent draft arm pricing tier linkage", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{
				"arms":[{"arm_id":%q,"pricing_tier_id":%q}]
			}`, draftExperiment.Arms[0].ID.String(), missingTierID.String()))
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+draftExperiment.ID.String()+"/arms/pricing-tiers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("PUT rejects updating arm pricing tiers for non-draft experiment", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{
				"arms":[{"arm_id":%q,"pricing_tier_id":%q}]
			}`, runningExperiment.Arms[0].ID.String(), primaryTierID.String()))
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+runningExperiment.ID.String()+"/arms/pricing-tiers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("PUT updates draft experiment arms through the existing update endpoint", func(t *testing.T) {
		var controlArmID uuid.UUID
		var removedVariantArmID uuid.UUID
		for _, arm := range draftExperiment.Arms {
			if arm.IsControl {
				controlArmID = arm.ID
				continue
			}
			removedVariantArmID = arm.ID
		}
		require.NotEqual(t, uuid.Nil, controlArmID)
		require.NotEqual(t, uuid.Nil, removedVariantArmID)

		body := []byte(fmt.Sprintf(`{
			"name":"Draft onboarding test v3",
			"description":"Builder-style arm rewrite",
			"algorithm_type":"epsilon_greedy",
			"is_bandit":false,
			"min_sample_size":240,
			"confidence_threshold_percent":93,
			"start_at":"2026-01-06T10:00:00Z",
			"end_at":"2026-01-13T10:00:00Z",
			"automation_policy":{"enabled":true,"auto_complete":true,"complete_on_confidence":true},
			"arms":[
				{"id":%q,"name":"Control Prime","description":"Updated control copy","is_control":true,"traffic_weight":1.25,"pricing_tier_id":%q},
				{"name":"Variant C","description":"New premium annual bundle","is_control":false,"traffic_weight":1.75,"pricing_tier_id":%q}
			]
		}`, controlArmID.String(), primaryTierID.String(), upsellTierID.String()))
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+draftExperiment.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "Draft onboarding test v3", resp.Data.Name)
		require.Len(t, resp.Data.Arms, 2)

		armIDs := make(map[uuid.UUID]struct{}, len(resp.Data.Arms))
		armByName := make(map[string]handlers.AdminExperimentArm, len(resp.Data.Arms))
		for _, arm := range resp.Data.Arms {
			armIDs[arm.ID] = struct{}{}
			armByName[arm.Name] = arm
		}
		_, removedStillPresent := armIDs[removedVariantArmID]
		assert.False(t, removedStillPresent)
		controlArm, ok := armByName["Control Prime"]
		require.True(t, ok)
		assert.Equal(t, controlArmID, controlArm.ID)
		assert.InDelta(t, 1.25, controlArm.TrafficWeight, 0.001)
		require.NotNil(t, controlArm.PricingTierID)
		assert.Equal(t, primaryTierID, *controlArm.PricingTierID)
		newVariant, ok := armByName["Variant C"]
		require.True(t, ok)
		assert.NotEqual(t, uuid.Nil, newVariant.ID)
		require.NotNil(t, newVariant.PricingTierID)
		assert.Equal(t, upsellTierID, *newVariant.PricingTierID)
		draftExperiment = resp.Data
	})

	t.Run("PUT rejects draft experiment arm updates with fewer than two arms", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{
			"name":"Draft onboarding test v3",
			"description":"Builder-style arm rewrite",
			"algorithm_type":"epsilon_greedy",
			"is_bandit":false,
			"min_sample_size":240,
			"confidence_threshold_percent":93,
			"arms":[
				{"id":%q,"name":"Control Prime","description":"Updated control copy","is_control":true,"traffic_weight":1.25}
			]
		}`, draftExperiment.Arms[0].ID.String()))
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+draftExperiment.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("PUT rejects draft experiment arm updates with unknown arm id", func(t *testing.T) {
		unknownArmID := uuid.New()
		body := []byte(fmt.Sprintf(`{
			"name":"Draft onboarding test v3",
			"description":"Builder-style arm rewrite",
			"algorithm_type":"epsilon_greedy",
			"is_bandit":false,
			"min_sample_size":240,
			"confidence_threshold_percent":93,
			"arms":[
				{"id":%q,"name":"Control Prime","description":"Updated control copy","is_control":true,"traffic_weight":1.25},
				{"id":%q,"name":"Ghost Arm","description":"Should fail","is_control":false,"traffic_weight":1}
			]
		}`, draftExperiment.Arms[0].ID.String(), unknownArmID.String()))
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+draftExperiment.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("PUT rejects updating a non-draft experiment", func(t *testing.T) {
		body := []byte(`{
			"name":"Running experiment edit",
			"description":"Should fail",
			"algorithm_type":"thompson_sampling",
			"is_bandit":true,
			"min_sample_size":250,
			"confidence_threshold_percent":95
		}`)
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/experiments/"+runningExperiment.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("POST rejects invalid arms payload", func(t *testing.T) {
		body := []byte(`{
			"name":"Broken test",
			"status":"draft",
			"is_bandit":false,
			"min_sample_size":100,
			"confidence_threshold_percent":95,
			"arms":[{"name":"Only arm","is_control":true,"traffic_weight":1}]
		}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("POST pauses a running experiment", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+runningExperiment.ID.String()+"/pause", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "paused", resp.Data.Status)
		require.NotNil(t, resp.Data.LatestLifecycleAudit)
		assert.Equal(t, "admin", resp.Data.LatestLifecycleAudit.ActorType)
		assert.Equal(t, "admin_experiments_api", resp.Data.LatestLifecycleAudit.Source)
		assert.Equal(t, "running", resp.Data.LatestLifecycleAudit.FromStatus)
		assert.Equal(t, "paused", resp.Data.LatestLifecycleAudit.ToStatus)
		assert.Equal(t, "manual_paused", resp.Data.LatestLifecycleAudit.Details["reason"])
		runningExperiment = resp.Data
	})

	t.Run("POST resumes a paused experiment", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+runningExperiment.ID.String()+"/resume", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "running", resp.Data.Status)
		runningExperiment = resp.Data
	})

	t.Run("POST starts a draft experiment via resume", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+draftExperiment.ID.String()+"/resume", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "running", resp.Data.Status)
		assert.NotNil(t, resp.Data.StartAt)
		draftExperiment = resp.Data
	})

	t.Run("POST completes a running experiment", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+runningExperiment.ID.String()+"/complete", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "completed", resp.Data.Status)
		assert.NotNil(t, resp.Data.EndAt)
		runningExperiment = resp.Data
	})

	t.Run("POST rejects invalid lifecycle transition", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/experiments/"+runningExperiment.ID.String()+"/pause", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("GET tolerates missing optional experiment audit tables", func(t *testing.T) {
		_, err := db.Exec(ctx, `
			ALTER TABLE ab_tests DROP COLUMN automation_policy;
			ALTER TABLE ab_test_arms DROP COLUMN pricing_tier_id;
			DROP TABLE experiment_winner_recommendation_log;
			DROP TABLE experiment_lifecycle_audit_log;
		`)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/experiments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []handlers.AdminExperiment `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 2)

		var foundRunning bool
		for _, experiment := range resp.Data {
			assert.Nil(t, experiment.LatestLifecycleAudit)
			if experiment.ID != runningExperiment.ID {
				continue
			}
			foundRunning = true
			require.NotNil(t, experiment.WinnerRecommendation)
			assert.Equal(t, "recommend_winner", experiment.WinnerRecommendation.Reason)
		}
		assert.True(t, foundRunning)
	})
}
