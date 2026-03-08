package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExperimentNotFound      = errors.New("experiment not found")
	ErrExperimentNotEditable   = errors.New("only draft experiments can be edited")
	ErrInvalidStatusTransition = errors.New("invalid experiment status transition")
)

type ExperimentMutationState struct {
	ID      uuid.UUID
	Status  string
	StartAt *time.Time
	EndAt   *time.Time
}

type ExperimentAutomationPolicy struct {
	Enabled              bool `json:"enabled"`
	AutoStart            bool `json:"auto_start"`
	AutoComplete         bool `json:"auto_complete"`
	CompleteOnEndTime    bool `json:"complete_on_end_time"`
	CompleteOnSampleSize bool `json:"complete_on_sample_size"`
	CompleteOnConfidence bool `json:"complete_on_confidence"`
	ManualOverride       bool `json:"manual_override"`
}

func DefaultExperimentAutomationPolicy() ExperimentAutomationPolicy {
	return ExperimentAutomationPolicy{
		Enabled:              false,
		AutoStart:            false,
		AutoComplete:         false,
		CompleteOnEndTime:    true,
		CompleteOnSampleSize: false,
		CompleteOnConfidence: false,
		ManualOverride:       false,
	}
}

func NormalizeExperimentAutomationPolicy(policy *ExperimentAutomationPolicy) ExperimentAutomationPolicy {
	if policy == nil {
		return DefaultExperimentAutomationPolicy()
	}

	normalized := DefaultExperimentAutomationPolicy()
	normalized.Enabled = policy.Enabled
	normalized.AutoStart = policy.AutoStart
	normalized.AutoComplete = policy.AutoComplete
	normalized.CompleteOnEndTime = policy.CompleteOnEndTime
	normalized.CompleteOnSampleSize = policy.CompleteOnSampleSize
	normalized.CompleteOnConfidence = policy.CompleteOnConfidence
	normalized.ManualOverride = policy.ManualOverride

	if normalized.AutoComplete && !normalized.CompleteOnEndTime && !normalized.CompleteOnSampleSize && !normalized.CompleteOnConfidence {
		normalized.CompleteOnEndTime = true
	}

	return normalized
}

type UpdateExperimentInput struct {
	Name                string
	Description         string
	AlgorithmType       *string
	IsBandit            bool
	MinSampleSize       int
	ConfidenceThreshold float64
	StartAt             *time.Time
	EndAt               *time.Time
	AutomationPolicy    ExperimentAutomationPolicy
}

type ExperimentStatusTransitionAudit struct {
	ActorType      string
	ActorID        *uuid.UUID
	Source         string
	IdempotencyKey *string
	Details        map[string]interface{}
}

type ExperimentMutationRepository interface {
	GetExperimentMutationState(ctx context.Context, experimentID uuid.UUID) (*ExperimentMutationState, error)
	UpdateExperimentDraft(ctx context.Context, experimentID uuid.UUID, input UpdateExperimentInput) error
	UpdateExperimentStatus(ctx context.Context, experimentID uuid.UUID, nextStatus string, startAt, endAt *time.Time) error
	UpdateExperimentStatusWithAudit(ctx context.Context, experimentID uuid.UUID, currentStatus, nextStatus string, startAt, endAt *time.Time, audit *ExperimentStatusTransitionAudit) error
}

type InvalidExperimentStatusTransitionError struct {
	CurrentStatus string
	NextStatus    string
}

func (e InvalidExperimentStatusTransitionError) Error() string {
	return "Cannot transition experiment from " + e.CurrentStatus + " to " + e.NextStatus
}

func (e InvalidExperimentStatusTransitionError) Unwrap() error {
	return ErrInvalidStatusTransition
}

type ExperimentAdminService struct {
	repo ExperimentMutationRepository
	now  func() time.Time
}

func NewExperimentAdminService(repo ExperimentMutationRepository) *ExperimentAdminService {
	return &ExperimentAdminService{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *ExperimentAdminService) UpdateDraftExperiment(ctx context.Context, experimentID uuid.UUID, input UpdateExperimentInput) error {
	experiment, err := s.repo.GetExperimentMutationState(ctx, experimentID)
	if err != nil {
		return err
	}
	if experiment.Status != "draft" {
		return ErrExperimentNotEditable
	}
	return s.repo.UpdateExperimentDraft(ctx, experimentID, input)
}

func (s *ExperimentAdminService) TransitionExperimentStatus(ctx context.Context, experimentID uuid.UUID, nextStatus string) error {
	return s.transitionExperimentStatus(ctx, experimentID, nextStatus, nil)
}

func (s *ExperimentAdminService) TransitionExperimentStatusWithAudit(ctx context.Context, experimentID uuid.UUID, nextStatus string, audit *ExperimentStatusTransitionAudit) error {
	return s.transitionExperimentStatus(ctx, experimentID, nextStatus, audit)
}

func (s *ExperimentAdminService) transitionExperimentStatus(ctx context.Context, experimentID uuid.UUID, nextStatus string, audit *ExperimentStatusTransitionAudit) error {
	experiment, err := s.repo.GetExperimentMutationState(ctx, experimentID)
	if err != nil {
		return err
	}
	if err := validateExperimentStatusTransition(experiment.Status, nextStatus); err != nil {
		return err
	}

	now := s.now()
	var startAt *time.Time
	if experiment.StartAt != nil {
		value := *experiment.StartAt
		startAt = &value
	}
	if nextStatus == "running" && (experiment.StartAt == nil || experiment.StartAt.After(now)) {
		value := now
		startAt = &value
	}

	var endAt *time.Time
	if experiment.EndAt != nil {
		value := *experiment.EndAt
		endAt = &value
	}
	if nextStatus == "completed" {
		value := now
		endAt = &value
	}

	return s.repo.UpdateExperimentStatusWithAudit(ctx, experimentID, experiment.Status, nextStatus, startAt, endAt, audit)
}

func validateExperimentStatusTransition(currentStatus string, nextStatus string) error {
	switch currentStatus {
	case "draft":
		if nextStatus == "running" {
			return nil
		}
	case "running":
		if nextStatus == "paused" || nextStatus == "completed" {
			return nil
		}
	case "paused":
		if nextStatus == "running" || nextStatus == "completed" {
			return nil
		}
	}

	return InvalidExperimentStatusTransitionError{
		CurrentStatus: currentStatus,
		NextStatus:    nextStatus,
	}
}
