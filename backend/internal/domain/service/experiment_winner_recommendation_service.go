package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	Source              string
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

type WinnerRecommendationEvent struct {
	ExperimentID               uuid.UUID
	Source                     string
	Recommended                bool
	Reason                     string
	WinningArmID               *uuid.UUID
	ConfidencePercent          *float64
	ConfidenceThresholdPercent float64
	ObservedSamples            int
	MinSampleSize              int
	Details                    map[string]interface{}
	OccurredAt                 time.Time
}

type winnerRecommendationEventAppender interface {
	AppendWinnerRecommendationEvent(ctx context.Context, event *WinnerRecommendationEvent) error
}

type WinnerProbabilityCalculator interface {
	CalculateWinProbability(ctx context.Context, experimentID uuid.UUID, simulations int) (map[uuid.UUID]float64, error)
}

type ExperimentWinnerRecommendationService struct {
	calculator  WinnerProbabilityCalculator
	simulations int
	appender    winnerRecommendationEventAppender
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
	service := &ExperimentWinnerRecommendationService{calculator: bandit, simulations: 2000}
	if appender, ok := banditRepo.(winnerRecommendationEventAppender); ok {
		service.appender = appender
	}
	return service
}

func NewExperimentWinnerRecommendationServiceWithCalculator(calculator WinnerProbabilityCalculator) *ExperimentWinnerRecommendationService {
	return &ExperimentWinnerRecommendationService{calculator: calculator, simulations: 2000}
}

func NewExperimentWinnerRecommendationServiceWithCalculatorAndAppender(calculator WinnerProbabilityCalculator, appender winnerRecommendationEventAppender) *ExperimentWinnerRecommendationService {
	return &ExperimentWinnerRecommendationService{calculator: calculator, simulations: 2000, appender: appender}
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
		return s.finalizeRecommendation(ctx, input, recommendation)
	case "running", "paused", "completed":
	default:
		recommendation.Reason = WinnerRecommendationReasonStatusNotEligible
		return s.finalizeRecommendation(ctx, input, recommendation)
	}

	if len(input.Arms) < 2 {
		recommendation.Reason = WinnerRecommendationReasonInsufficientArms
		return s.finalizeRecommendation(ctx, input, recommendation)
	}
	if input.TotalSamples <= 0 {
		recommendation.Reason = WinnerRecommendationReasonInsufficientData
		return s.finalizeRecommendation(ctx, input, recommendation)
	}

	winProbabilities, err := s.calculator.CalculateWinProbability(ctx, input.ExperimentID, s.simulations)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate win probabilities: %w", err)
	}

	bestArmID, bestProbability, found := highestProbability(winProbabilities)
	if !found {
		recommendation.Reason = WinnerRecommendationReasonInsufficientData
		return s.finalizeRecommendation(ctx, input, recommendation)
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
		return s.finalizeRecommendation(ctx, input, recommendation)
	}
	if confidence < input.ConfidenceThreshold {
		recommendation.Reason = WinnerRecommendationReasonConfidenceBelowThreshold
		return s.finalizeRecommendation(ctx, input, recommendation)
	}

	recommendation.Recommended = true
	recommendation.Reason = WinnerRecommendationReasonRecommendWinner
	return s.finalizeRecommendation(ctx, input, recommendation)
}

func (s *ExperimentWinnerRecommendationService) finalizeRecommendation(ctx context.Context, input ExperimentWinnerRecommendationInput, recommendation *WinnerRecommendation) (*WinnerRecommendation, error) {
	if recommendation == nil || s.appender == nil {
		return recommendation, nil
	}
	source := strings.TrimSpace(input.Source)
	if source == "" {
		source = "winner_recommendation_service"
	}
	details := map[string]interface{}{
		"status":                    input.Status,
		"arms_count":                len(input.Arms),
		"is_bandit":                 input.IsBandit,
		"used_persisted_confidence": input.WinnerConfidence != nil,
	}
	if recommendation.WinningArmName != nil {
		details["winning_arm_name"] = *recommendation.WinningArmName
	}
	if err := s.appender.AppendWinnerRecommendationEvent(ctx, &WinnerRecommendationEvent{
		ExperimentID:               input.ExperimentID,
		Source:                     source,
		Recommended:                recommendation.Recommended,
		Reason:                     recommendation.Reason,
		WinningArmID:               recommendation.WinningArmID,
		ConfidencePercent:          recommendation.ConfidencePercent,
		ConfidenceThresholdPercent: recommendation.ConfidenceThresholdPercent,
		ObservedSamples:            recommendation.ObservedSamples,
		MinSampleSize:              recommendation.MinSampleSize,
		Details:                    details,
		OccurredAt:                 time.Now().UTC(),
	}); err != nil {
		return nil, fmt.Errorf("failed to append winner recommendation event: %w", err)
	}
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
