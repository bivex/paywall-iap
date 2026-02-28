package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// BanditHandler handles multi-armed bandit endpoints
type BanditHandler struct {
	banditService BanditService
}

// BanditService defines the interface for bandit operations
type BanditService interface {
	SelectArm(ctx context.Context, experimentID, userID uuid.UUID) (uuid.UUID, error)
	UpdateReward(ctx context.Context, experimentID, armID uuid.UUID, reward float64) error
	GetArmStatistics(ctx context.Context, experimentID uuid.UUID) (map[uuid.UUID]*service.ArmStats, error)
	CalculateWinProbability(ctx context.Context, experimentID uuid.UUID, simulations int) (map[uuid.UUID]float64, error)
}

// NewBanditHandler creates a new bandit handler
func NewBanditHandler(banditService BanditService) *BanditHandler {
	return &BanditHandler{
		banditService: banditService,
	}
}

// AssignRequest represents the request to assign a user to a variant
type AssignRequest struct {
	ExperimentID string `json:"experiment_id" binding:"required,uuid"`
	UserID       string `json:"user_id" binding:"required,uuid"`
}

// AssignResponse represents the response with the assigned variant
type AssignResponse struct {
	ExperimentID string `json:"experiment_id"`
	UserID       string `json:"user_id"`
	ArmID        string `json:"arm_id"`
	IsNew        bool   `json:"is_new"` // true if this is a new assignment (not from cache)
}

// Assign assigns a user to an experiment arm using Thompson Sampling
// @Summary Assign user to experiment arm
// @Tags bandit
// @Accept json
// @Produce json
// @Param request body AssignRequest true "Assignment request"
// @Success 200 {object} response.SuccessResponse{data=AssignResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/bandit/assign [post]
func (h *BanditHandler) Assign(c *gin.Context) {
	var req AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	experimentID, err := uuid.Parse(req.ExperimentID)
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	// Get arm assignment using Thompson Sampling
	armID, err := h.banditService.SelectArm(c.Request.Context(), experimentID, userID)
	if err != nil {
		response.InternalError(c, "Failed to assign arm: "+err.Error())
		return
	}

	// Check if this was a new assignment (from DB vs cache)
	// For simplicity, we assume cached assignments return the same arm
	// In production, you'd want to track this more explicitly

	resp := AssignResponse{
		ExperimentID: req.ExperimentID,
		UserID:       req.UserID,
		ArmID:        armID.String(),
		IsNew:        false, // TODO: Track if assignment was from cache
	}

	response.OK(c, resp)
}

// RewardRequest represents the request to record a reward/conversion
type RewardRequest struct {
	ExperimentID string  `json:"experiment_id" binding:"required,uuid"`
	ArmID        string  `json:"arm_id" binding:"required,uuid"`
	UserID       string  `json:"user_id" binding:"required,uuid"`
	Reward       float64 `json:"reward" binding:"required"`
	Currency     string  `json:"currency,omitempty"`
}

// RewardResponse represents the response after recording a reward
type RewardResponse struct {
	ExperimentID string `json:"experiment_id"`
	ArmID        string `json:"arm_id"`
	Reward       float64 `json:"reward"`
	Updated      bool   `json:"updated"`
}

// Reward records a reward (conversion or revenue) for an arm
// @Summary Record reward for arm
// @Tags bandit
// @Accept json
// @Produce json
// @Param request body RewardRequest true "Reward request"
// @Success 200 {object} response.SuccessResponse{data=RewardResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/bandit/reward [post]
func (h *BanditHandler) Reward(c *gin.Context) {
	var req RewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	experimentID, err := uuid.Parse(req.ExperimentID)
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}

	armID, err := uuid.Parse(req.ArmID)
	if err != nil {
		response.BadRequest(c, "Invalid arm ID")
		return
	}

	// For revenue rewards, you'd typically convert to USD here
	// using a currency conversion service
	// reward = convertToUSD(req.Reward, req.Currency)

	reward := req.Reward

	// Update the bandit with the reward
	err = h.banditService.UpdateReward(c.Request.Context(), experimentID, armID, reward)
	if err != nil {
		response.InternalError(c, "Failed to record reward: "+err.Error())
		return
	}

	resp := RewardResponse{
		ExperimentID: req.ExperimentID,
		ArmID:        req.ArmID,
		Reward:       reward,
		Updated:      true,
	}

	response.OK(c, resp)
}

// StatisticsRequest represents the request to get experiment statistics
type StatisticsRequest struct {
	ExperimentID string `form:"experiment_id" binding:"required,uuid"`
}

// StatisticsResponse represents the statistics for all arms
type StatisticsResponse struct {
	ExperimentID string                `json:"experiment_id"`
	Arms         []ArmStatistics       `json:"arms"`
	WinProbs     map[string]float64    `json:"win_probabilities,omitempty"`
}

// ArmStatistics represents statistics for a single arm
type ArmStatistics struct {
	ArmID       string  `json:"arm_id"`
	Alpha       float64 `json:"alpha"`
	Beta        float64 `json:"beta"`
	Samples     int     `json:"samples"`
	Conversions int     `json:"conversions"`
	Revenue     float64 `json:"revenue"`
	AvgReward   float64 `json:"avg_reward"`
	ConversionRate float64 `json:"conversion_rate"`
}

// Statistics returns statistics for all arms in an experiment
// @Summary Get experiment statistics
// @Tags bandit
// @Produce json
// @Param experiment_id query string true "Experiment ID"
// @Param win_probs query bool false "Include win probabilities"
// @Success 200 {object} response.SuccessResponse{data=StatisticsResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/bandit/statistics [get]
func (h *BanditHandler) Statistics(c *gin.Context) {
	var req StatisticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	experimentID, err := uuid.Parse(req.ExperimentID)
	if err != nil {
		response.BadRequest(c, "Invalid experiment ID")
		return
	}

	// Get arm statistics
	armStats, err := h.banditService.GetArmStatistics(c.Request.Context(), experimentID)
	if err != nil {
		response.InternalError(c, "Failed to get statistics: "+err.Error())
		return
	}

	// Convert to response format
	arms := make([]ArmStatistics, 0, len(armStats))
	for _, stats := range armStats {
		conversionRate := 0.0
		if stats.Samples > 0 {
			conversionRate = float64(stats.Conversions) / float64(stats.Samples)
		}

		arms = append(arms, ArmStatistics{
			ArmID:         stats.ArmID.String(),
			Alpha:         stats.Alpha,
			Beta:          stats.Beta,
			Samples:       stats.Samples,
			Conversions:   stats.Conversions,
			Revenue:       stats.Revenue,
			AvgReward:     stats.AvgReward,
			ConversionRate: conversionRate,
		})
	}

	resp := StatisticsResponse{
		ExperimentID: req.ExperimentID,
		Arms:         arms,
	}

	// Optionally include win probabilities
	if c.Query("win_probs") == "true" {
		winProbs, err := h.banditService.CalculateWinProbability(c.Request.Context(), experimentID, 1000)
		if err == nil {
			probs := make(map[string]float64)
			for armID, prob := range winProbs {
				probs[armID.String()] = prob
			}
			resp.WinProbs = probs
		}
	}

	response.OK(c, resp)
}

// Health returns the health status of the bandit service
// @Summary Bandit service health check
// @Tags bandit
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/bandit/health [get]
func (h *BanditHandler) Health(c *gin.Context) {
	response.OK(c, gin.H{
		"status": "healthy",
		"service": "bandit",
	})
}
