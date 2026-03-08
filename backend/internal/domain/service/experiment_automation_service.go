package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ExperimentAutomationState struct {
	ID                  uuid.UUID
	Status              string
	StartAt             *time.Time
	EndAt               *time.Time
	MinSampleSize       int
	ConfidenceThreshold float64
	WinnerConfidence    *float64
	TotalSamples        int
	AutomationPolicy    ExperimentAutomationPolicy
}

type ExperimentAutomationRepository interface {
	ListExperimentAutomationStates(ctx context.Context) ([]ExperimentAutomationState, error)
}

type ExperimentStatusTransitioner interface {
	TransitionExperimentStatus(ctx context.Context, experimentID uuid.UUID, nextStatus string) error
}

type ExperimentAutomationRunResult struct {
	Started   []uuid.UUID
	Completed []uuid.UUID
	Skipped   int
}

type ExperimentAutomationReconciler struct {
	repo        ExperimentAutomationRepository
	transitions ExperimentStatusTransitioner
	now         func() time.Time
}

func NewExperimentAutomationReconciler(repo ExperimentAutomationRepository, transitions ExperimentStatusTransitioner) *ExperimentAutomationReconciler {
	return &ExperimentAutomationReconciler{
		repo:        repo,
		transitions: transitions,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func (r *ExperimentAutomationReconciler) Reconcile(ctx context.Context) (ExperimentAutomationRunResult, error) {
	states, err := r.repo.ListExperimentAutomationStates(ctx)
	if err != nil {
		return ExperimentAutomationRunResult{}, fmt.Errorf("failed to list experiment automation states: %w", err)
	}

	now := r.now()
	result := ExperimentAutomationRunResult{}
	for _, state := range states {
		nextStatus := r.nextStatus(state, now)
		if nextStatus == "" {
			result.Skipped++
			continue
		}

		if err := r.transitions.TransitionExperimentStatus(ctx, state.ID, nextStatus); err != nil {
			return result, fmt.Errorf("failed to transition experiment %s to %s: %w", state.ID, nextStatus, err)
		}

		switch nextStatus {
		case "running":
			result.Started = append(result.Started, state.ID)
		case "completed":
			result.Completed = append(result.Completed, state.ID)
		}
	}

	return result, nil
}

func (r *ExperimentAutomationReconciler) nextStatus(state ExperimentAutomationState, now time.Time) string {
	if !state.AutomationPolicy.Enabled || state.AutomationPolicy.ManualOverride {
		return ""
	}

	switch state.Status {
	case "draft":
		if state.AutomationPolicy.AutoStart && state.StartAt != nil && !state.StartAt.After(now) {
			return "running"
		}
	case "running":
		if state.AutomationPolicy.AutoComplete && shouldAutoCompleteExperiment(state, now) {
			return "completed"
		}
	}

	return ""
}

func shouldAutoCompleteExperiment(state ExperimentAutomationState, now time.Time) bool {
	if state.AutomationPolicy.CompleteOnEndTime && state.EndAt != nil && !state.EndAt.After(now) {
		return true
	}
	if state.AutomationPolicy.CompleteOnSampleSize && state.TotalSamples >= state.MinSampleSize {
		return true
	}
	if state.AutomationPolicy.CompleteOnConfidence && state.WinnerConfidence != nil && *state.WinnerConfidence >= state.ConfidenceThreshold {
		return true
	}
	return false
}
