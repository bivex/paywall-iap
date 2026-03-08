package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeWinnerProbabilityCalculator struct {
	probabilities map[uuid.UUID]float64
	err           error
	calls         int
}

func (f *fakeWinnerProbabilityCalculator) CalculateWinProbability(ctx context.Context, experimentID uuid.UUID, simulations int) (map[uuid.UUID]float64, error) {
	f.calls++
	return f.probabilities, f.err
}

func TestExperimentWinnerRecommendationService(t *testing.T) {
	ctx := context.Background()
	experimentID := uuid.New()
	controlArmID := uuid.New()
	variantArmID := uuid.New()

	baseInput := ExperimentWinnerRecommendationInput{
		ExperimentID:        experimentID,
		Status:              "running",
		IsBandit:            true,
		MinSampleSize:       100,
		TotalSamples:        120,
		ConfidenceThreshold: 0.95,
		Arms: []ExperimentWinnerRecommendationArm{
			{ID: controlArmID, Name: "Control", IsControl: true},
			{ID: variantArmID, Name: "Variant A", IsControl: false},
		},
	}

	t.Run("returns nil for non-bandit experiments", func(t *testing.T) {
		calculator := &fakeWinnerProbabilityCalculator{}
		svc := NewExperimentWinnerRecommendationServiceWithCalculator(calculator)

		recommendation, err := svc.Recommend(ctx, ExperimentWinnerRecommendationInput{IsBandit: false})

		require.NoError(t, err)
		assert.Nil(t, recommendation)
		assert.Equal(t, 0, calculator.calls)
	})

	t.Run("returns draft reason without calculating probabilities", func(t *testing.T) {
		calculator := &fakeWinnerProbabilityCalculator{}
		svc := NewExperimentWinnerRecommendationServiceWithCalculator(calculator)

		recommendation, err := svc.Recommend(ctx, ExperimentWinnerRecommendationInput{
			ExperimentID: experimentID,
			Status:       "draft",
			IsBandit:     true,
		})

		require.NoError(t, err)
		require.NotNil(t, recommendation)
		assert.False(t, recommendation.Recommended)
		assert.Equal(t, WinnerRecommendationReasonDraftExperiment, recommendation.Reason)
		assert.Equal(t, 0, calculator.calls)
	})

	t.Run("returns current leader but blocks recommendation below min sample size", func(t *testing.T) {
		calculator := &fakeWinnerProbabilityCalculator{probabilities: map[uuid.UUID]float64{controlArmID: 0.03, variantArmID: 0.97}}
		svc := NewExperimentWinnerRecommendationServiceWithCalculator(calculator)
		persistedConfidence := 0.97

		recommendation, err := svc.Recommend(ctx, ExperimentWinnerRecommendationInput{
			ExperimentID:        baseInput.ExperimentID,
			Status:              baseInput.Status,
			IsBandit:            baseInput.IsBandit,
			MinSampleSize:       200,
			TotalSamples:        120,
			ConfidenceThreshold: baseInput.ConfidenceThreshold,
			WinnerConfidence:    &persistedConfidence,
			Arms:                baseInput.Arms,
		})

		require.NoError(t, err)
		require.NotNil(t, recommendation)
		assert.False(t, recommendation.Recommended)
		assert.Equal(t, WinnerRecommendationReasonInsufficientSampleSize, recommendation.Reason)
		require.NotNil(t, recommendation.WinningArmID)
		assert.Equal(t, variantArmID, *recommendation.WinningArmID)
		require.NotNil(t, recommendation.WinningArmName)
		assert.Equal(t, "Variant A", *recommendation.WinningArmName)
		require.NotNil(t, recommendation.ConfidencePercent)
		assert.InDelta(t, 97.0, *recommendation.ConfidencePercent, 0.001)
		assert.Equal(t, 1, calculator.calls)
	})

	t.Run("recommends winner once threshold and sample size are met", func(t *testing.T) {
		calculator := &fakeWinnerProbabilityCalculator{probabilities: map[uuid.UUID]float64{controlArmID: 0.04, variantArmID: 0.96}}
		svc := NewExperimentWinnerRecommendationServiceWithCalculator(calculator)

		recommendation, err := svc.Recommend(ctx, baseInput)

		require.NoError(t, err)
		require.NotNil(t, recommendation)
		assert.True(t, recommendation.Recommended)
		assert.Equal(t, WinnerRecommendationReasonRecommendWinner, recommendation.Reason)
		require.NotNil(t, recommendation.WinningArmID)
		assert.Equal(t, variantArmID, *recommendation.WinningArmID)
		require.NotNil(t, recommendation.ConfidencePercent)
		assert.InDelta(t, 96.0, *recommendation.ConfidencePercent, 0.001)
		assert.Equal(t, 1, calculator.calls)
	})
}
