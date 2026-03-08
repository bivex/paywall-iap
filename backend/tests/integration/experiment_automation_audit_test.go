package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestExperimentStatusAuditIdempotency(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE TABLE ab_tests (
			id UUID PRIMARY KEY,
			status VARCHAR(20) NOT NULL,
			start_at TIMESTAMPTZ,
			end_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
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

		CREATE TABLE experiment_automation_decision_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			source TEXT NOT NULL,
			decision_type TEXT NOT NULL,
			reason TEXT,
			from_status TEXT NOT NULL,
			to_status TEXT NOT NULL,
			idempotency_key TEXT,
			details JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);

		CREATE UNIQUE INDEX idx_experiment_automation_decision_log_idempotency
			ON experiment_automation_decision_log(idempotency_key);
	`)
	require.NoError(t, err)

	repo := repository.NewExperimentAdminRepository(db)
	experimentID := uuid.New()
	_, err = db.Exec(ctx, `INSERT INTO ab_tests (id, status) VALUES ($1, 'draft')`, experimentID)
	require.NoError(t, err)

	startedAt := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)
	auditKey := "experiment:" + experimentID.String() + ":running"
	audit := &service.ExperimentStatusTransitionAudit{
		ActorType:      "system",
		Source:         "experiment_automation_reconciler",
		IdempotencyKey: &auditKey,
		Details:        map[string]interface{}{"reason": "auto_start"},
	}

	err = repo.UpdateExperimentStatusWithAudit(ctx, experimentID, "draft", "running", &startedAt, nil, audit)
	require.NoError(t, err)
	err = repo.UpdateExperimentStatusWithAudit(ctx, experimentID, "draft", "running", &startedAt, nil, audit)
	require.NoError(t, err)

	var status string
	require.NoError(t, db.QueryRow(ctx, `SELECT status FROM ab_tests WHERE id = $1`, experimentID).Scan(&status))
	assert.Equal(t, "running", status)

	var auditCount int
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(*) FROM experiment_lifecycle_audit_log WHERE idempotency_key = $1`, auditKey).Scan(&auditCount))
	assert.Equal(t, 1, auditCount)

	var decisionCount int
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(*) FROM experiment_automation_decision_log WHERE idempotency_key = $1`, auditKey).Scan(&decisionCount))
	assert.Equal(t, 1, decisionCount)

	var reason string
	require.NoError(t, db.QueryRow(ctx, `SELECT reason FROM experiment_automation_decision_log WHERE idempotency_key = $1`, auditKey).Scan(&reason))
	assert.Equal(t, "auto_start", reason)
}
