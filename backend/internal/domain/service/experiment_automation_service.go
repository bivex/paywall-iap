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
	TransitionExperimentStatusWithAudit(ctx context.Context, experimentID uuid.UUID, nextStatus string, audit *ExperimentStatusTransitionAudit) error
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
		decision := r.nextDecision(state, now)
		if decision.NextStatus == "" {
			result.Skipped++
			continue
		}

		auditKey := fmt.Sprintf("experiment:%s:%s", state.ID, decision.NextStatus)
		audit := &ExperimentStatusTransitionAudit{
			ActorType:      "system",
			Source:         "experiment_automation_reconciler",
			IdempotencyKey: &auditKey,
			Details: map[string]interface{}{
				"reason": decision.Reason,
			},
		}

		if err := r.transitions.TransitionExperimentStatusWithAudit(ctx, state.ID, decision.NextStatus, audit); err != nil {
			return result, fmt.Errorf("failed to transition experiment %s to %s: %w", state.ID, decision.NextStatus, err)
		}

		switch decision.NextStatus {
		case "running":
			result.Started = append(result.Started, state.ID)
		case "completed":
			result.Completed = append(result.Completed, state.ID)
		}
	}

	return result, nil
}

type experimentAutomationDecision struct {
	NextStatus string
	Reason     string
}

func (r *ExperimentAutomationReconciler) nextDecision(state ExperimentAutomationState, now time.Time) experimentAutomationDecision {
	if !state.AutomationPolicy.Enabled || state.AutomationPolicy.ManualOverride {
		return experimentAutomationDecision{}
	}

	switch state.Status {
	case "draft":
		if state.AutomationPolicy.AutoStart && state.StartAt != nil && !state.StartAt.After(now) {
			return experimentAutomationDecision{NextStatus: "running", Reason: "auto_start"}
		}
	case "running":
		if state.AutomationPolicy.AutoComplete {
			if reason := autoCompleteReason(state, now); reason != "" {
				return experimentAutomationDecision{NextStatus: "completed", Reason: reason}
			}
		}
	}

	return experimentAutomationDecision{}
}

func autoCompleteReason(state ExperimentAutomationState, now time.Time) string {
	if state.AutomationPolicy.CompleteOnEndTime && state.EndAt != nil && !state.EndAt.After(now) {
		return "auto_complete_end_time"
	}
	if state.AutomationPolicy.CompleteOnSampleSize && state.TotalSamples >= state.MinSampleSize {
		return "auto_complete_sample_size"
	}
	if state.AutomationPolicy.CompleteOnConfidence && state.WinnerConfidence != nil && *state.WinnerConfidence >= state.ConfidenceThreshold {
		return "auto_complete_confidence"
	}
	return ""
}
