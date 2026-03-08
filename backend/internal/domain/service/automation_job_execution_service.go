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
	if spec.JobName == "" {
		return false, fmt.Errorf("scheduled automation job requires a job name")
	}
	if spec.Source == "" {
		return false, fmt.Errorf("scheduled automation job %s requires a source", spec.JobName)
	}
	if spec.Window <= 0 {
		return false, fmt.Errorf("scheduled automation job %s requires a positive window", spec.JobName)
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
	if runErr != nil {
		failureDetails := cloneAutomationJobDetails(details)
		failureDetails["error"] = runErr.Error()
		if err := s.repo.FinishAutomationJobRun(ctx, jobRun.ID, AutomationJobRunStatusFailed, failureDetails); err != nil {
			return true, fmt.Errorf("scheduled automation job %s failed: %w (failed to persist failure status: %v)", spec.JobName, runErr, err)
		}
		return true, runErr
	}

	if err := s.repo.FinishAutomationJobRun(ctx, jobRun.ID, AutomationJobRunStatusCompleted, details); err != nil {
		return true, fmt.Errorf("failed to persist completion status for scheduled automation job %s: %w", spec.JobName, err)
	}

	return true, nil
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
