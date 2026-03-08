package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestPostgresBanditRepository_ProcessPendingConversionPersistsImmutableEvent(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE ab_tests (id UUID PRIMARY KEY, name TEXT NOT NULL);
		CREATE TABLE ab_test_arms (id UUID PRIMARY KEY, experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, name TEXT NOT NULL);
		CREATE TABLE ab_test_arm_stats (
			arm_id UUID PRIMARY KEY REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			alpha DOUBLE PRECISION NOT NULL,
			beta DOUBLE PRECISION NOT NULL,
			samples INTEGER NOT NULL,
			conversions INTEGER NOT NULL,
			revenue DOUBLE PRECISION NOT NULL,
			avg_reward DOUBLE PRECISION NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE bandit_pending_rewards (
			id UUID PRIMARY KEY,
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			assigned_at TIMESTAMPTZ NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			converted BOOLEAN NOT NULL DEFAULT FALSE,
			conversion_value DOUBLE PRECISION,
			conversion_currency TEXT,
			converted_at TIMESTAMPTZ,
			processed_at TIMESTAMPTZ
		);
		CREATE TABLE bandit_conversion_links (
			pending_id UUID NOT NULL REFERENCES bandit_pending_rewards(id) ON DELETE CASCADE,
			transaction_id UUID NOT NULL,
			linked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			PRIMARY KEY (pending_id, transaction_id)
		);
		CREATE TABLE bandit_conversion_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			user_id UUID,
			pending_reward_id UUID REFERENCES bandit_pending_rewards(id) ON DELETE SET NULL,
			transaction_id UUID,
			event_type TEXT NOT NULL,
			original_reward_value DOUBLE PRECISION NOT NULL DEFAULT 0,
			original_currency TEXT,
			normalized_reward_value DOUBLE PRECISION NOT NULL DEFAULT 0,
			normalized_currency TEXT,
			metadata JSONB,
			occurred_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE UNIQUE INDEX idx_bandit_conversion_events_pending_event ON bandit_conversion_events(pending_reward_id, event_type) WHERE pending_reward_id IS NOT NULL;
		CREATE UNIQUE INDEX idx_bandit_conversion_events_transaction_delayed ON bandit_conversion_events(transaction_id) WHERE transaction_id IS NOT NULL AND event_type = 'delayed_conversion';
	`)
	require.NoError(t, err)

	experimentID := uuid.New()
	armID := uuid.New()
	userID := uuid.New()
	pendingID := uuid.New()
	transactionID := uuid.New()
	processedAt := time.Date(2026, 3, 8, 15, 0, 0, 0, time.UTC)

	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, name) VALUES ($1, 'Bandit conversion test')`, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name) VALUES ($1, $2, 'Variant A')`, armID, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward) VALUES ($1, 5, 2, 6, 4, 30, 5)`, armID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO bandit_pending_rewards (id, experiment_id, arm_id, user_id, assigned_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6)`, pendingID, experimentID, armID, userID, processedAt.Add(-time.Hour), processedAt.Add(time.Hour))
	require.NoError(t, err)

	repo := repository.NewPostgresBanditRepository(db, zap.NewNop())
	matched, processed, err := repo.ProcessPendingConversion(ctx, transactionID, userID, 19.99, "USD", processedAt)
	require.NoError(t, err)
	require.True(t, processed)
	require.NotNil(t, matched)
	assert.Equal(t, pendingID, matched.ID)
	assert.True(t, matched.Converted)

	var alpha, beta, revenue float64
	var samples, conversions int
	require.NoError(t, db.QueryRow(ctx, `SELECT alpha, beta, samples, conversions, revenue FROM ab_test_arm_stats WHERE arm_id = $1`, armID).Scan(&alpha, &beta, &samples, &conversions, &revenue))
	assert.Equal(t, 6.0, alpha)
	assert.Equal(t, 2.0, beta)
	assert.Equal(t, 7, samples)
	assert.Equal(t, 5, conversions)
	assert.InDelta(t, 49.99, revenue, 0.0001)

	var converted bool
	var storedValue float64
	var storedCurrency string
	require.NoError(t, db.QueryRow(ctx, `SELECT converted, conversion_value, conversion_currency FROM bandit_pending_rewards WHERE id = $1`, pendingID).Scan(&converted, &storedValue, &storedCurrency))
	assert.True(t, converted)
	assert.InDelta(t, 19.99, storedValue, 0.0001)
	assert.Equal(t, "USD", storedCurrency)

	var linkCount int
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(*) FROM bandit_conversion_links WHERE pending_id = $1 AND transaction_id = $2`, pendingID, transactionID).Scan(&linkCount))
	assert.Equal(t, 1, linkCount)

	var eventType string
	var originalValue, normalizedValue float64
	var originalCurrency, normalizedCurrency string
	require.NoError(t, db.QueryRow(ctx, `SELECT event_type, original_reward_value, normalized_reward_value, COALESCE(original_currency, ''), COALESCE(normalized_currency, '') FROM bandit_conversion_events WHERE pending_reward_id = $1`, pendingID).Scan(&eventType, &originalValue, &normalizedValue, &originalCurrency, &normalizedCurrency))
	assert.Equal(t, "delayed_conversion", eventType)
	assert.InDelta(t, 19.99, originalValue, 0.0001)
	assert.InDelta(t, 19.99, normalizedValue, 0.0001)
	assert.Equal(t, "USD", originalCurrency)
	assert.Equal(t, "USD", normalizedCurrency)
}

func TestPostgresBanditRepository_ProcessExpiredPendingRewardPersistsImmutableEvent(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE ab_tests (id UUID PRIMARY KEY, name TEXT NOT NULL);
		CREATE TABLE ab_test_arms (id UUID PRIMARY KEY, experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, name TEXT NOT NULL);
		CREATE TABLE ab_test_arm_stats (
			arm_id UUID PRIMARY KEY REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			alpha DOUBLE PRECISION NOT NULL,
			beta DOUBLE PRECISION NOT NULL,
			samples INTEGER NOT NULL,
			conversions INTEGER NOT NULL,
			revenue DOUBLE PRECISION NOT NULL,
			avg_reward DOUBLE PRECISION NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE bandit_pending_rewards (
			id UUID PRIMARY KEY,
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			assigned_at TIMESTAMPTZ NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			converted BOOLEAN NOT NULL DEFAULT FALSE,
			conversion_value DOUBLE PRECISION,
			conversion_currency TEXT,
			converted_at TIMESTAMPTZ,
			processed_at TIMESTAMPTZ
		);
		CREATE TABLE bandit_conversion_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			user_id UUID,
			pending_reward_id UUID REFERENCES bandit_pending_rewards(id) ON DELETE SET NULL,
			transaction_id UUID,
			event_type TEXT NOT NULL,
			original_reward_value DOUBLE PRECISION NOT NULL DEFAULT 0,
			original_currency TEXT,
			normalized_reward_value DOUBLE PRECISION NOT NULL DEFAULT 0,
			normalized_currency TEXT,
			metadata JSONB,
			occurred_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE UNIQUE INDEX idx_bandit_conversion_events_pending_event ON bandit_conversion_events(pending_reward_id, event_type) WHERE pending_reward_id IS NOT NULL;
	`)
	require.NoError(t, err)

	experimentID := uuid.New()
	armID := uuid.New()
	userID := uuid.New()
	pendingID := uuid.New()
	processedAt := time.Date(2026, 3, 8, 16, 0, 0, 0, time.UTC)

	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, name) VALUES ($1, 'Bandit expiry test')`, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name) VALUES ($1, $2, 'Control')`, armID, experimentID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward) VALUES ($1, 3, 4, 9, 3, 42, 4.6667)`, armID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO bandit_pending_rewards (id, experiment_id, arm_id, user_id, assigned_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6)`, pendingID, experimentID, armID, userID, processedAt.Add(-2*time.Hour), processedAt.Add(-time.Minute))
	require.NoError(t, err)

	repo := repository.NewPostgresBanditRepository(db, zap.NewNop())
	processed, err := repo.ProcessExpiredPendingReward(ctx, pendingID, processedAt)
	require.NoError(t, err)
	require.True(t, processed)

	var alpha, beta, revenue float64
	var samples, conversions int
	require.NoError(t, db.QueryRow(ctx, `SELECT alpha, beta, samples, conversions, revenue FROM ab_test_arm_stats WHERE arm_id = $1`, armID).Scan(&alpha, &beta, &samples, &conversions, &revenue))
	assert.Equal(t, 3.0, alpha)
	assert.Equal(t, 5.0, beta)
	assert.Equal(t, 10, samples)
	assert.Equal(t, 3, conversions)
	assert.InDelta(t, 42.0, revenue, 0.0001)

	var processedAtDB time.Time
	require.NoError(t, db.QueryRow(ctx, `SELECT processed_at FROM bandit_pending_rewards WHERE id = $1`, pendingID).Scan(&processedAtDB))
	assert.Equal(t, processedAt, processedAtDB.UTC())

	var eventType string
	require.NoError(t, db.QueryRow(ctx, `SELECT event_type FROM bandit_conversion_events WHERE pending_reward_id = $1`, pendingID).Scan(&eventType))
	assert.Equal(t, "expired_pending_reward", eventType)
}
