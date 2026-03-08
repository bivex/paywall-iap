package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubExperimentRepairCandidateRepository struct {
	ids       []uuid.UUID
	err       error
	lastLimit int
}

func (s *stubExperimentRepairCandidateRepository) ListExperimentRepairCandidateIDs(_ context.Context, limit int) ([]uuid.UUID, error) {
	s.lastLimit = limit
	if s.err != nil {
		return nil, s.err
	}
	return s.ids, nil
}

type stubExperimentRepairExecutor struct {
	summaries map[uuid.UUID]*ExperimentRepairSummary
	errs      map[uuid.UUID]error
	repaired  []uuid.UUID
}

func (s *stubExperimentRepairExecutor) RepairExperiment(_ context.Context, experimentID uuid.UUID) (*ExperimentRepairSummary, error) {
	s.repaired = append(s.repaired, experimentID)
	if err := s.errs[experimentID]; err != nil {
		return nil, err
	}
	return s.summaries[experimentID], nil
}

func TestExperimentRepairReconcilerReconcileAggregatesRepairSummaries(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	candidates := &stubExperimentRepairCandidateRepository{ids: []uuid.UUID{firstID, secondID}}
	repairer := &stubExperimentRepairExecutor{summaries: map[uuid.UUID]*ExperimentRepairSummary{
		firstID:  {MissingArmStatsInserted: 1, ExpiredPendingRewards: 2, PendingRewardsProcessed: 2},
		secondID: {MissingArmStatsInserted: 3, ExpiredPendingRewards: 4, PendingRewardsProcessed: 1},
	}}
	reconciler := NewExperimentRepairReconciler(candidates, repairer)

	result, err := reconciler.Reconcile(context.Background(), 0)

	require.NoError(t, err)
	assert.Equal(t, 50, candidates.lastLimit)
	assert.Equal(t, 2, result.Scanned)
	assert.Equal(t, []uuid.UUID{firstID, secondID}, result.Repaired)
	assert.Empty(t, result.Failures)
	assert.Equal(t, 4, result.MissingArmStatsInserted)
	assert.Equal(t, 6, result.ExpiredPendingRewards)
	assert.Equal(t, 3, result.PendingRewardsProcessed)
	assert.Equal(t, []uuid.UUID{firstID, secondID}, repairer.repaired)
}

func TestExperimentRepairReconcilerReconcileReturnsPartialFailure(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	candidates := &stubExperimentRepairCandidateRepository{ids: []uuid.UUID{firstID, secondID}}
	repairer := &stubExperimentRepairExecutor{
		summaries: map[uuid.UUID]*ExperimentRepairSummary{firstID: {PendingRewardsProcessed: 1}},
		errs:      map[uuid.UUID]error{secondID: errors.New("boom")},
	}
	reconciler := NewExperimentRepairReconciler(candidates, repairer)

	result, err := reconciler.Reconcile(context.Background(), 10)

	require.Error(t, err)
	assert.Equal(t, 10, candidates.lastLimit)
	assert.Equal(t, 2, result.Scanned)
	assert.Equal(t, []uuid.UUID{firstID}, result.Repaired)
	assert.Equal(t, map[string]string{secondID.String(): "boom"}, result.Failures)
	assert.Equal(t, 1, result.PendingRewardsProcessed)
	assert.Contains(t, err.Error(), "failed to repair 1 experiment")
}
