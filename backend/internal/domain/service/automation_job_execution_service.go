package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	AutomationJobRunStatusRunning   = "running"
	AutomationJobRunStatusCompleted = "completed"
	AutomationJobRunStatusFailed    = "failed"
)

type AutomationJobRun struct {
	ID             uuid.UUID
	JobName        string
	IdempotencyKey string
	Status         string
}

type AutomationJobRunClaimInput struct {
	JobName         string
	Source          string
	IdempotencyKey  string
	Payload         []byte
	WindowStartedAt time.Time
	WindowDuration  time.Duration
}

type AutomationJobRunRepository interface {
	ClaimAutomationJobRun(ctx context.Context, input AutomationJobRunClaimInput) (*AutomationJobRun, bool, error)
	FinishAutomationJobRun(ctx context.Context, runID uuid.UUID, status string, details map[string]any) error
}

type ScheduledAutomationJobSpec struct {
	JobName string
	Source  string
	Window  time.Duration
}

type AutomationJobExecutionService struct {
	repo AutomationJobRunRepository
	now  func() time.Time
}

func NewAutomationJobExecutionService(repo AutomationJobRunRepository) *AutomationJobExecutionService {
	return &AutomationJobExecutionService{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *AutomationJobExecutionService) ExecuteScheduled(
	ctx context.Context,
	spec ScheduledAutomationJobSpec,
	payload []byte,
	run func(context.Context) (map[string]any, error),
) (bool, error) {
	if err := validateScheduledJobSpec(spec); err != nil {
		return false, err
	}

	windowStartedAt := s.now().UTC().Truncate(spec.Window)
	idempotencyKey := scheduledAutomationJobIdempotencyKey(spec.JobName, windowStartedAt, payload)

	jobRun, claimed, err := s.repo.ClaimAutomationJobRun(ctx, AutomationJobRunClaimInput{
		JobName:         spec.JobName,
		Source:          spec.Source,
		IdempotencyKey:  idempotencyKey,
		Payload:         payload,
		WindowStartedAt: windowStartedAt,
		WindowDuration:  spec.Window,
	})
	if err != nil {
		return false, fmt.Errorf("failed to claim scheduled automation job %s: %w", spec.JobName, err)
	}
	if !claimed {
		return false, nil
	}

	details, runErr := run(ctx)
	if err := s.recordJobRunCompletion(ctx, jobRunCompletionParams{
		jobRunID: jobRun.ID,
		jobName:  spec.JobName,
		details:  details,
		runErr:   runErr,
	}); err != nil {
		return true, err
	}
	return true, runErr
}

func validateScheduledJobSpec(spec ScheduledAutomationJobSpec) error {
	if spec.JobName == "" {
		return fmt.Errorf("scheduled automation job requires a job name")
	}
	if spec.Source == "" {
		return fmt.Errorf("scheduled automation job %s requires a source", spec.JobName)
	}
	if spec.Window <= 0 {
		return fmt.Errorf("scheduled automation job %s requires a positive window", spec.JobName)
	}
	return nil
}

type jobRunCompletionParams struct {
	jobRunID uuid.UUID
	jobName  string
	details  map[string]any
	runErr   error
}

func (s *AutomationJobExecutionService) recordJobRunCompletion(
	ctx context.Context,
	p jobRunCompletionParams,
) error {
	if p.runErr != nil {
		failureDetails := cloneAutomationJobDetails(p.details)
		failureDetails["error"] = p.runErr.Error()
		if err := s.repo.FinishAutomationJobRun(ctx, p.jobRunID, AutomationJobRunStatusFailed, failureDetails); err != nil {
			return fmt.Errorf("scheduled automation job %s failed: %w (failed to persist failure status: %v)", p.jobName, p.runErr, err)
		}
		return nil
	}

	if err := s.repo.FinishAutomationJobRun(ctx, p.jobRunID, AutomationJobRunStatusCompleted, p.details); err != nil {
		return fmt.Errorf("failed to persist completion status for scheduled automation job %s: %w", p.jobName, err)
	}
	return nil
}

func scheduledAutomationJobIdempotencyKey(jobName string, windowStartedAt time.Time, payload []byte) string {
	base := fmt.Sprintf("%s:%s", jobName, windowStartedAt.UTC().Format(time.RFC3339))
	if len(payload) == 0 {
		return base
	}
	hash := sha256.Sum256(payload)
	return fmt.Sprintf("%s:%s", base, hex.EncodeToString(hash[:8]))
}

func cloneAutomationJobDetails(details map[string]any) map[string]any {
	if len(details) == 0 {
		return map[string]any{}
	}
	clone := make(map[string]any, len(details))
	for key, value := range details {
		clone[key] = value
	}
	return clone
}
