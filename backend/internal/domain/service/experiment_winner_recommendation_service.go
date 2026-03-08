package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	WinnerRecommendationReasonDraftExperiment          = "draft_experiment"
	WinnerRecommendationReasonStatusNotEligible        = "status_not_eligible"
	WinnerRecommendationReasonInsufficientArms         = "insufficient_arms"
	WinnerRecommendationReasonInsufficientData         = "insufficient_data"
	WinnerRecommendationReasonInsufficientSampleSize   = "insufficient_sample_size"
	WinnerRecommendationReasonConfidenceBelowThreshold = "confidence_below_threshold"
	WinnerRecommendationReasonRecommendWinner          = "recommend_winner"
)

type ExperimentWinnerRecommendationArm struct {
	ID        uuid.UUID
	Name      string
	IsControl bool
}

type ExperimentWinnerRecommendationInput struct {
	ExperimentID        uuid.UUID
	Status              string
	IsBandit            bool
	MinSampleSize       int
	TotalSamples        int
	ConfidenceThreshold float64
	WinnerConfidence    *float64
	Arms                []ExperimentWinnerRecommendationArm
}

type WinnerRecommendation struct {
	Recommended                bool       `json:"recommended"`
	Reason                     string     `json:"reason"`
	WinningArmID               *uuid.UUID `json:"winning_arm_id,omitempty"`
	WinningArmName             *string    `json:"winning_arm_name,omitempty"`
	ConfidencePercent          *float64   `json:"confidence_percent,omitempty"`
	ConfidenceThresholdPercent float64    `json:"confidence_threshold_percent"`
	ObservedSamples            int        `json:"observed_samples"`
	MinSampleSize              int        `json:"min_sample_size"`
}

type WinnerProbabilityCalculator interface {
	CalculateWinProbability(ctx context.Context, experimentID uuid.UUID, simulations int) (map[uuid.UUID]float64, error)
}

type ExperimentWinnerRecommendationService struct {
	calculator  WinnerProbabilityCalculator
	simulations int
}

type recommendationNoopBanditCache struct{}

func (recommendationNoopBanditCache) GetArmStats(context.Context, string) (*ArmStats, error) {
	return nil, errors.New("cache miss")
}

func (recommendationNoopBanditCache) SetArmStats(context.Context, string, *ArmStats, time.Duration) error {
	return nil
}

func (recommendationNoopBanditCache) GetAssignment(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, errors.New("cache miss")
}

func (recommendationNoopBanditCache) SetAssignment(context.Context, string, uuid.UUID, time.Duration) error {
	return nil
}

func NewExperimentWinnerRecommendationService(banditRepo BanditRepository) *ExperimentWinnerRecommendationService {
	bandit := NewThompsonSamplingBandit(banditRepo, recommendationNoopBanditCache{}, zap.NewNop())
	return &ExperimentWinnerRecommendationService{calculator: bandit, simulations: 2000}
}

func NewExperimentWinnerRecommendationServiceWithCalculator(calculator WinnerProbabilityCalculator) *ExperimentWinnerRecommendationService {
	return &ExperimentWinnerRecommendationService{calculator: calculator, simulations: 2000}
}

func (s *ExperimentWinnerRecommendationService) Recommend(ctx context.Context, input ExperimentWinnerRecommendationInput) (*WinnerRecommendation, error) {
	if !input.IsBandit {
		return nil, nil
	}

	recommendation := &WinnerRecommendation{
		ConfidenceThresholdPercent: input.ConfidenceThreshold * 100,
		ObservedSamples:            input.TotalSamples,
		MinSampleSize:              input.MinSampleSize,
	}

	switch input.Status {
	case "draft":
		recommendation.Reason = WinnerRecommendationReasonDraftExperiment
		return recommendation, nil
	case "running", "paused", "completed":
	default:
		recommendation.Reason = WinnerRecommendationReasonStatusNotEligible
		return recommendation, nil
	}

	if len(input.Arms) < 2 {
		recommendation.Reason = WinnerRecommendationReasonInsufficientArms
		return recommendation, nil
	}
	if input.TotalSamples <= 0 {
		recommendation.Reason = WinnerRecommendationReasonInsufficientData
		return recommendation, nil
	}

	winProbabilities, err := s.calculator.CalculateWinProbability(ctx, input.ExperimentID, s.simulations)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate win probabilities: %w", err)
	}

	bestArmID, bestProbability, found := highestProbability(winProbabilities)
	if !found {
		recommendation.Reason = WinnerRecommendationReasonInsufficientData
		return recommendation, nil
	}

	recommendation.WinningArmID = &bestArmID
	if bestArm := findRecommendationArm(input.Arms, bestArmID); bestArm != nil {
		name := bestArm.Name
		recommendation.WinningArmName = &name
	}

	confidence := bestProbability
	if input.WinnerConfidence != nil {
		confidence = *input.WinnerConfidence
	}
	confidencePercent := confidence * 100
	recommendation.ConfidencePercent = &confidencePercent

	if input.TotalSamples < input.MinSampleSize {
		recommendation.Reason = WinnerRecommendationReasonInsufficientSampleSize
		return recommendation, nil
	}
	if confidence < input.ConfidenceThreshold {
		recommendation.Reason = WinnerRecommendationReasonConfidenceBelowThreshold
		return recommendation, nil
	}

	recommendation.Recommended = true
	recommendation.Reason = WinnerRecommendationReasonRecommendWinner
	return recommendation, nil
}

func highestProbability(probabilities map[uuid.UUID]float64) (uuid.UUID, float64, bool) {
	var bestArmID uuid.UUID
	var bestProbability float64
	found := false
	for armID, probability := range probabilities {
		if !found || probability > bestProbability {
			bestArmID = armID
			bestProbability = probability
			found = true
		}
	}
	return bestArmID, bestProbability, found
}

func findRecommendationArm(arms []ExperimentWinnerRecommendationArm, armID uuid.UUID) *ExperimentWinnerRecommendationArm {
	for _, arm := range arms {
		if arm.ID == armID {
			copy := arm
			return &copy
		}
	}
	return nil
}
