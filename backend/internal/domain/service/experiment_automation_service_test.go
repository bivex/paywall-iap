package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestExperimentAutomationNextDecisionSkipsWhenTimedLockIsActive(t *testing.T) {
	now := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)
	lockedUntil := now.Add(2 * time.Hour)
	reconciler := &ExperimentAutomationReconciler{}

	decision := reconciler.nextDecision(ExperimentAutomationState{
		ID:                  uuid.New(),
		Status:              "draft",
		StartAt:             ptrTime(now.Add(-time.Hour)),
		MinSampleSize:       100,
		ConfidenceThreshold: 0.95,
		AutomationPolicy: ExperimentAutomationPolicy{
			Enabled:     true,
			AutoStart:   true,
			LockedUntil: &lockedUntil,
		},
	}, now)

	assert.Empty(t, decision.NextStatus)
	assert.Empty(t, decision.Reason)
}

func TestExperimentAutomationNextDecisionResumesAfterTimedLockExpires(t *testing.T) {
	now := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)
	lockedUntil := now.Add(-time.Minute)
	reconciler := &ExperimentAutomationReconciler{}

	decision := reconciler.nextDecision(ExperimentAutomationState{
		ID:                  uuid.New(),
		Status:              "draft",
		StartAt:             ptrTime(now.Add(-time.Hour)),
		MinSampleSize:       100,
		ConfidenceThreshold: 0.95,
		AutomationPolicy: ExperimentAutomationPolicy{
			Enabled:     true,
			AutoStart:   true,
			LockedUntil: &lockedUntil,
		},
	}, now)

	assert.Equal(t, "running", decision.NextStatus)
	assert.Equal(t, "auto_start", decision.Reason)
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
