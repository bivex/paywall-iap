package regression

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

func TestPostgresBanditRepositoryCreateAssignmentPersistsSelectionMetadata(t *testing.T) {
	ctx := context.Background()
	db, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE ab_tests (id UUID PRIMARY KEY, name TEXT NOT NULL);
		CREATE TABLE ab_test_arms (id UUID PRIMARY KEY, experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE, name TEXT NOT NULL);
		CREATE TABLE ab_test_assignments (
			id UUID PRIMARY KEY,
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			assigned_at TIMESTAMPTZ NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			CONSTRAINT unique_assignment UNIQUE (experiment_id, user_id)
		);
		CREATE TABLE bandit_assignment_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			assignment_id UUID NOT NULL REFERENCES ab_test_assignments(id) ON DELETE CASCADE,
			experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
			event_type TEXT NOT NULL,
			metadata JSONB,
			occurred_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	require.NoError(t, err)

	experimentID := uuid.New()
	armID := uuid.New()
	userID := uuid.New()
	assignedAt := time.Date(2026, 3, 8, 21, 0, 0, 0, time.UTC)
	require.NoError(t, db.QueryRow(ctx, `INSERT INTO ab_tests (id, name) VALUES ($1, 'Bandit assignment metadata test') RETURNING id`, experimentID).Scan(&experimentID))
	_, err = db.Exec(ctx, `INSERT INTO ab_test_arms (id, experiment_id, name) VALUES ($1, $2, 'Variant A')`, armID, experimentID)
	require.NoError(t, err)

	repo := repository.NewPostgresBanditRepository(db, zap.NewNop())
	err = repo.CreateAssignment(ctx, &service.Assignment{
		ID:           uuid.New(),
		ExperimentID: experimentID,
		UserID:       userID,
		ArmID:        armID,
		AssignedAt:   assignedAt,
		ExpiresAt:    assignedAt.Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"selection_strategy": "thompson_sampling",
			"arms_considered":    2,
			"selected_arm_name":  "Variant A",
			"arm_scores": []map[string]interface{}{{
				"arm_id": armID,
				"sample": 0.91,
			}},
		},
	})
	require.NoError(t, err)

	var eventType string
	var metadata []byte
	require.NoError(t, db.QueryRow(ctx, `SELECT event_type, metadata FROM bandit_assignment_events WHERE experiment_id = $1 AND user_id = $2`, experimentID, userID).Scan(&eventType, &metadata))
	assert.Equal(t, string(service.AssignmentEventTypeAssigned), eventType)
	assert.JSONEq(t, `{"selection_strategy":"thompson_sampling","arms_considered":2,"selected_arm_name":"Variant A","arm_scores":[{"arm_id":"`+armID.String()+`","sample":0.91}]}`, string(metadata))
}
