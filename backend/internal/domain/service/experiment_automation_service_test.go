package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubExperimentAutomationRepository struct {
	states []ExperimentAutomationState
	err    error
}

func (s *stubExperimentAutomationRepository) ListExperimentAutomationStates(context.Context) ([]ExperimentAutomationState, error) {
	return s.states, s.err
}

type stubExperimentStatusTransitioner struct {
	transitions map[uuid.UUID]transitionCall
	err         error
}

type transitionCall struct {
	Status string
	Audit  *ExperimentStatusTransitionAudit
}

func (s *stubExperimentStatusTransitioner) TransitionExperimentStatusWithAudit(_ context.Context, experimentID uuid.UUID, nextStatus string, audit *ExperimentStatusTransitionAudit) error {
	if s.err != nil {
		return s.err
	}
	if s.transitions == nil {
		s.transitions = make(map[uuid.UUID]transitionCall)
	}
	if _, exists := s.transitions[experimentID]; !exists {
		s.transitions[experimentID] = transitionCall{Status: nextStatus, Audit: audit}
	}
	return nil
}

func TestExperimentAutomationReconciler(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC)

	t.Run("starts eligible draft experiments", func(t *testing.T) {
		experimentID := uuid.New()
		startedAt := now.Add(-time.Hour)
		repo := &stubExperimentAutomationRepository{states: []ExperimentAutomationState{{
			ID:               experimentID,
			Status:           "draft",
			StartAt:          &startedAt,
			AutomationPolicy: ExperimentAutomationPolicy{Enabled: true, AutoStart: true},
		}}}
		transitions := &stubExperimentStatusTransitioner{}
		reconciler := NewExperimentAutomationReconciler(repo, transitions)
		reconciler.now = func() time.Time { return now }

		result, err := reconciler.Reconcile(ctx)

		require.NoError(t, err)
		assert.Equal(t, "running", transitions.transitions[experimentID].Status)
		require.NotNil(t, transitions.transitions[experimentID].Audit)
		assert.Equal(t, "experiment_automation_reconciler", transitions.transitions[experimentID].Audit.Source)
		assert.Equal(t, []uuid.UUID{experimentID}, result.Started)
		assert.Empty(t, result.Completed)
	})

	t.Run("completes running experiments when end time has passed", func(t *testing.T) {
		experimentID := uuid.New()
		endedAt := now.Add(-time.Minute)
		repo := &stubExperimentAutomationRepository{states: []ExperimentAutomationState{{
			ID:               experimentID,
			Status:           "running",
			EndAt:            &endedAt,
			AutomationPolicy: ExperimentAutomationPolicy{Enabled: true, AutoComplete: true, CompleteOnEndTime: true},
		}}}
		transitions := &stubExperimentStatusTransitioner{}
		reconciler := NewExperimentAutomationReconciler(repo, transitions)
		reconciler.now = func() time.Time { return now }

		result, err := reconciler.Reconcile(ctx)

		require.NoError(t, err)
		assert.Equal(t, "completed", transitions.transitions[experimentID].Status)
		assert.Equal(t, []uuid.UUID{experimentID}, result.Completed)
	})

	t.Run("completes running experiments by sample size or winner confidence", func(t *testing.T) {
		samplesID := uuid.New()
		confidenceID := uuid.New()
		winnerConfidence := 0.96
		repo := &stubExperimentAutomationRepository{states: []ExperimentAutomationState{
			{
				ID:                  samplesID,
				Status:              "running",
				MinSampleSize:       200,
				TotalSamples:        250,
				ConfidenceThreshold: 0.95,
				AutomationPolicy:    ExperimentAutomationPolicy{Enabled: true, AutoComplete: true, CompleteOnSampleSize: true},
			},
			{
				ID:                  confidenceID,
				Status:              "running",
				MinSampleSize:       500,
				TotalSamples:        100,
				ConfidenceThreshold: 0.95,
				WinnerConfidence:    &winnerConfidence,
				AutomationPolicy:    ExperimentAutomationPolicy{Enabled: true, AutoComplete: true, CompleteOnConfidence: true},
			},
		}}
		transitions := &stubExperimentStatusTransitioner{}
		reconciler := NewExperimentAutomationReconciler(repo, transitions)

		result, err := reconciler.Reconcile(ctx)

		require.NoError(t, err)
		assert.Equal(t, "completed", transitions.transitions[samplesID].Status)
		assert.Equal(t, "completed", transitions.transitions[confidenceID].Status)
		assert.ElementsMatch(t, []uuid.UUID{samplesID, confidenceID}, result.Completed)
	})

	t.Run("skips experiments with manual override or disabled automation", func(t *testing.T) {
		manualOverrideID := uuid.New()
		disabledID := uuid.New()
		startedAt := now.Add(-time.Hour)
		repo := &stubExperimentAutomationRepository{states: []ExperimentAutomationState{
			{ID: manualOverrideID, Status: "draft", StartAt: &startedAt, AutomationPolicy: ExperimentAutomationPolicy{Enabled: true, AutoStart: true, ManualOverride: true}},
			{ID: disabledID, Status: "draft", StartAt: &startedAt, AutomationPolicy: ExperimentAutomationPolicy{Enabled: false, AutoStart: true}},
		}}
		transitions := &stubExperimentStatusTransitioner{}
		reconciler := NewExperimentAutomationReconciler(repo, transitions)
		reconciler.now = func() time.Time { return now }

		result, err := reconciler.Reconcile(ctx)

		require.NoError(t, err)
		assert.Empty(t, transitions.transitions)
		assert.Equal(t, 2, result.Skipped)
	})
}
