package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubExperimentMutationRepository struct {
	state              *ExperimentMutationState
	loadErr            error
	updatedDraftInput  *UpdateExperimentInput
	updatedStatus      string
	updatedStatusStart *time.Time
	updatedStatusEnd   *time.Time
	updatedStatusAudit *ExperimentStatusTransitionAudit
}

func (s *stubExperimentMutationRepository) GetExperimentMutationState(context.Context, uuid.UUID) (*ExperimentMutationState, error) {
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	return s.state, nil
}

func (s *stubExperimentMutationRepository) UpdateExperimentDraft(_ context.Context, _ uuid.UUID, input UpdateExperimentInput) error {
	s.updatedDraftInput = &input
	return nil
}

func (s *stubExperimentMutationRepository) UpdateExperimentStatus(_ context.Context, _ uuid.UUID, nextStatus string, startAt, endAt *time.Time) error {
	s.updatedStatus = nextStatus
	s.updatedStatusStart = startAt
	s.updatedStatusEnd = endAt
	return nil
}

func (s *stubExperimentMutationRepository) UpdateExperimentStatusWithAudit(_ context.Context, _ uuid.UUID, _ string, nextStatus string, startAt, endAt *time.Time, audit *ExperimentStatusTransitionAudit) error {
	s.updatedStatus = nextStatus
	s.updatedStatusStart = startAt
	s.updatedStatusEnd = endAt
	s.updatedStatusAudit = audit
	return nil
}

func TestExperimentAdminService(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()

	t.Run("UpdateDraftExperiment rejects non-draft experiments", func(t *testing.T) {
		repo := &stubExperimentMutationRepository{state: &ExperimentMutationState{ID: experimentID, Status: "running"}}
		svc := NewExperimentAdminService(repo)

		err := svc.UpdateDraftExperiment(ctx, experimentID, UpdateExperimentInput{Name: "Updated"})

		require.ErrorIs(t, err, ErrExperimentNotEditable)
		assert.Nil(t, repo.updatedDraftInput)
	})

	t.Run("UpdateDraftExperiment forwards validated draft update to repository", func(t *testing.T) {
		algo := "ucb"
		repo := &stubExperimentMutationRepository{state: &ExperimentMutationState{ID: experimentID, Status: "draft"}}
		svc := NewExperimentAdminService(repo)

		input := UpdateExperimentInput{Name: "Draft", AlgorithmType: &algo, IsBandit: true, MinSampleSize: 200, ConfidenceThreshold: 0.95}
		err := svc.UpdateDraftExperiment(ctx, experimentID, input)

		require.NoError(t, err)
		require.NotNil(t, repo.updatedDraftInput)
		assert.Equal(t, input.Name, repo.updatedDraftInput.Name)
		assert.Equal(t, *input.AlgorithmType, *repo.updatedDraftInput.AlgorithmType)
	})

	t.Run("TransitionExperimentStatus starts draft experiments at current time", func(t *testing.T) {
		future := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
		now := time.Date(2026, 3, 8, 8, 0, 0, 0, time.UTC)
		repo := &stubExperimentMutationRepository{state: &ExperimentMutationState{ID: experimentID, Status: "draft", StartAt: &future}}
		svc := NewExperimentAdminService(repo)
		svc.now = func() time.Time { return now }

		err := svc.TransitionExperimentStatus(ctx, experimentID, "running")

		require.NoError(t, err)
		assert.Equal(t, "running", repo.updatedStatus)
		require.NotNil(t, repo.updatedStatusStart)
		assert.True(t, repo.updatedStatusStart.Equal(now))
		assert.Nil(t, repo.updatedStatusEnd)
	})

	t.Run("TransitionExperimentStatus completes running experiments at current time", func(t *testing.T) {
		started := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
		now := time.Date(2026, 3, 8, 9, 30, 0, 0, time.UTC)
		repo := &stubExperimentMutationRepository{state: &ExperimentMutationState{ID: experimentID, Status: "running", StartAt: &started}}
		svc := NewExperimentAdminService(repo)
		svc.now = func() time.Time { return now }

		err := svc.TransitionExperimentStatus(ctx, experimentID, "completed")

		require.NoError(t, err)
		assert.Equal(t, "completed", repo.updatedStatus)
		require.NotNil(t, repo.updatedStatusStart)
		assert.True(t, repo.updatedStatusStart.Equal(started))
		require.NotNil(t, repo.updatedStatusEnd)
		assert.True(t, repo.updatedStatusEnd.Equal(now))
	})

	t.Run("TransitionExperimentStatus rejects invalid transitions", func(t *testing.T) {
		repo := &stubExperimentMutationRepository{state: &ExperimentMutationState{ID: experimentID, Status: "completed"}}
		svc := NewExperimentAdminService(repo)

		err := svc.TransitionExperimentStatus(ctx, experimentID, "paused")

		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidStatusTransition)
		assert.Equal(t, "", repo.updatedStatus)
	})

	t.Run("TransitionExperimentStatus returns not found errors", func(t *testing.T) {
		repo := &stubExperimentMutationRepository{loadErr: ErrExperimentNotFound}
		svc := NewExperimentAdminService(repo)

		err := svc.TransitionExperimentStatus(ctx, experimentID, "running")

		require.ErrorIs(t, err, ErrExperimentNotFound)
	})

	t.Run("TransitionExperimentStatusWithAudit forwards audit metadata", func(t *testing.T) {
		repo := &stubExperimentMutationRepository{state: &ExperimentMutationState{ID: experimentID, Status: "draft"}}
		svc := NewExperimentAdminService(repo)
		auditKey := "experiment:transition:test"

		err := svc.TransitionExperimentStatusWithAudit(ctx, experimentID, "running", &ExperimentStatusTransitionAudit{
			ActorType:      "system",
			Source:         "experiment_automation_reconciler",
			IdempotencyKey: &auditKey,
			Details:        map[string]interface{}{"reason": "auto_start"},
		})

		require.NoError(t, err)
		require.NotNil(t, repo.updatedStatusAudit)
		assert.Equal(t, "system", repo.updatedStatusAudit.ActorType)
		assert.Equal(t, "experiment_automation_reconciler", repo.updatedStatusAudit.Source)
		require.NotNil(t, repo.updatedStatusAudit.IdempotencyKey)
		assert.Equal(t, auditKey, *repo.updatedStatusAudit.IdempotencyKey)
	})

	t.Run("NormalizeExperimentAutomationPolicy applies defaults and preserves explicit flags", func(t *testing.T) {
		policy := NormalizeExperimentAutomationPolicy(&ExperimentAutomationPolicy{
			Enabled:              true,
			AutoStart:            true,
			AutoComplete:         true,
			CompleteOnEndTime:    false,
			CompleteOnSampleSize: true,
			ManualOverride:       true,
		})

		assert.True(t, policy.Enabled)
		assert.True(t, policy.AutoStart)
		assert.True(t, policy.AutoComplete)
		assert.False(t, policy.CompleteOnEndTime)
		assert.True(t, policy.CompleteOnSampleSize)
		assert.False(t, policy.CompleteOnConfidence)
		assert.True(t, policy.ManualOverride)
	})
}
