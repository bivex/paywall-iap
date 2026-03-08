package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestBanditMaintenanceSummaryUsesRepositoryBackedMaintenance(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	for _, statement := range []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE users (
			id UUID PRIMARY KEY,
			platform_user_id TEXT UNIQUE NOT NULL,
			device_id TEXT,
			platform TEXT NOT NULL,
			app_version TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE ab_tests (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'paused', 'completed')) DEFAULT 'draft',
			algorithm_type TEXT,
			is_bandit BOOLEAN NOT NULL DEFAULT false,
			objective_type TEXT CHECK (objective_type IN ('conversion', 'ltv', 'revenue', 'hybrid')),
			objective_weights JSONB,
			window_type TEXT,
			window_size INTEGER,
			window_min_samples INTEGER,
			enable_contextual BOOLEAN NOT NULL DEFAULT false,
			enable_delayed BOOLEAN NOT NULL DEFAULT false,
			enable_currency BOOLEAN NOT NULL DEFAULT false,
			exploration_alpha NUMERIC(10,4),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE ab_test_arms (
			id UUID PRIMARY KEY,
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			description TEXT,
			is_control BOOLEAN NOT NULL DEFAULT false,
			traffic_weight NUMERIC(3,2) NOT NULL DEFAULT 1.0
		)`,
		`CREATE TABLE ab_test_arm_stats (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			arm_id UUID NOT NULL UNIQUE REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			alpha NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			beta NUMERIC(10,2) NOT NULL DEFAULT 1.0,
			samples INT NOT NULL DEFAULT 0,
			conversions INT NOT NULL DEFAULT 0,
			revenue NUMERIC(15,2) NOT NULL DEFAULT 0.0,
			avg_reward NUMERIC(10,4),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE ab_test_assignments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			expires_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE bandit_user_context (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			country TEXT,
			device TEXT,
			app_version TEXT,
			days_since_install INTEGER,
			total_spent NUMERIC(12,2) DEFAULT 0,
			last_purchase_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE bandit_arm_objective_stats (
			id BIGSERIAL PRIMARY KEY,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			objective_type VARCHAR(20) NOT NULL CHECK (objective_type IN ('conversion', 'ltv', 'revenue')),
			alpha DECIMAL(10,2) DEFAULT 1.0,
			beta DECIMAL(10,2) DEFAULT 1.0,
			samples BIGINT DEFAULT 0,
			conversions BIGINT DEFAULT 0,
			total_revenue DECIMAL(18,2) DEFAULT 0,
			avg_ltv DECIMAL(12,2),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(arm_id, objective_type)
		)`,
	} {
		_, err = db.Exec(ctx, statement)
		require.NoError(t, err)
	}

	experimentID := uuid.New()
	armID := uuid.New()
	userID := uuid.New()

	_, err = db.Exec(ctx, `INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email) VALUES ($1, 'maintenance-user', 'device-a', 'ios', '1.0.0', 'maintenance@example.com')`, userID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, name, status, algorithm_type, is_bandit, objective_type, objective_weights, window_type, enable_contextual, enable_delayed, enable_currency, exploration_alpha) VALUES ($1, 'Maintenance candidate', 'running', 'thompson_sampling', TRUE, 'revenue', NULL, NULL, TRUE, FALSE, FALSE, 1.0)`, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight) VALUES ($1, $2, 'Variant A', 'A', TRUE, 1.0)`, armID, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward) VALUES ($1, 9, 3, 10, 8, 120, 12)`, armID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_assignments (experiment_id, user_id, arm_id, expires_at) VALUES ($1, $2, $3, now() - interval '3 days')`, experimentID, userID, armID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO bandit_user_context (user_id, country, device, app_version, days_since_install, total_spent, updated_at) VALUES ($1, 'US', 'ios', '1.0.0', 20, 99, now() - interval '120 days')`, userID)
	require.NoError(t, err)

	repo := repository.NewPostgresBanditRepository(db, zap.NewNop())
	cache := &integrationBanditMaintenanceCache{}
	base := service.NewThompsonSamplingBandit(repo, cache, zap.NewNop())
	engine := service.NewAdvancedBanditEngine(base, repo, cache, nil, nil, zap.NewNop(), &service.EngineConfig{EnableHybrid: true})

	summary, err := engine.RunMaintenanceDetailed(ctx)

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, 1, summary.ObjectiveExperimentsScanned)
	assert.Equal(t, 1, summary.ObjectiveStatsSynced)
	assert.EqualValues(t, 1, summary.StaleContextsDeleted)
	assert.EqualValues(t, 1, summary.ExpiredAssignmentsDeleted)

	var objectiveRows, contextRows, assignmentRows int
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(*)::int FROM bandit_arm_objective_stats WHERE arm_id = $1 AND objective_type = 'revenue'`, armID).Scan(&objectiveRows))
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(*)::int FROM bandit_user_context WHERE user_id = $1`, userID).Scan(&contextRows))
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(*)::int FROM ab_test_assignments WHERE experiment_id = $1`, experimentID).Scan(&assignmentRows))
	assert.Equal(t, 1, objectiveRows)
	assert.Equal(t, 0, contextRows)
	assert.Equal(t, 0, assignmentRows)
}

type integrationBanditMaintenanceCache struct{}

func (c *integrationBanditMaintenanceCache) GetArmStats(context.Context, string) (*service.ArmStats, error) {
	return nil, nil
}

func (c *integrationBanditMaintenanceCache) SetArmStats(context.Context, string, *service.ArmStats, time.Duration) error {
	return nil
}

func (c *integrationBanditMaintenanceCache) GetAssignment(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (c *integrationBanditMaintenanceCache) SetAssignment(context.Context, string, uuid.UUID, time.Duration) error {
	return nil
}
