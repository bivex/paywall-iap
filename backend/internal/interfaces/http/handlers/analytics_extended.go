package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/cache"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// AnalyticsHandlersExtended handles extended analytics endpoints
type AnalyticsHandlersExtended struct {
	ltvService     *service.LTVService
	analyticsCache *cache.AnalyticsCache
	logger         *zap.Logger
}

// NewAnalyticsHandlersExtended creates a new extended analytics handlers group
func NewAnalyticsHandlersExtended(
	ltvService *service.LTVService,
	analyticsCache *cache.AnalyticsCache,
	logger *zap.Logger,
) *AnalyticsHandlersExtended {
	return &AnalyticsHandlersExtended{
		ltvService:     ltvService,
		analyticsCache: analyticsCache,
		logger:         logger,
	}
}

// GetLTV retrieves LTV estimates for a user
func (h *AnalyticsHandlersExtended) GetLTV(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID format")
		return
	}

	// Check cache first
	cachedLTV, err := h.analyticsCache.GetLTV(c.Request.Context(), userIDStr)
	if err == nil && cachedLTV != nil {
		response.OK(c, cachedLTV)
		return
	}

	// Calculate LTV
	estimates, err := h.ltvService.CalculateLTV(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to calculate LTV",
			zap.String("user_id", userIDStr),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to calculate LTV")
		return
	}

	// Cache the result
	ltvData := &cache.LTVData{
		UserID:       userIDStr,
		LTV30:        estimates.LTV30,
		LTV90:        estimates.LTV90,
		LTV365:       estimates.LTV365,
		LTVLifetime:  estimates.LTVLifetime,
		Confidence:   estimates.Confidence,
		CalculatedAt: estimates.CalculatedAt,
		Factors:      estimates.Factors,
	}
	h.analyticsCache.SetLTV(c.Request.Context(), userIDStr, ltvData)

	response.OK(c, estimates)
}

// UpdateLTV updates LTV after a purchase
func (h *AnalyticsHandlersExtended) UpdateLTV(c *gin.Context) {
	var req UpdateLTVRequestExtended
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.ltvService.UpdateUserLTV(c.Request.Context(), userID, req.Amount); err != nil {
		h.logger.Error("Failed to update LTV",
			zap.String("user_id", req.UserID),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to update LTV")
		return
	}

	h.analyticsCache.InvalidateLTV(c.Request.Context(), req.UserID)

	response.OK(c, gin.H{
		"message": "LTV updated successfully",
		"user_id": req.UserID,
		"amount":  req.Amount,
	})
}

type UpdateLTVRequestExtended struct {
	UserID string  `json:"user_id" binding:"required,uuid"`
	Amount float64 `json:"amount" binding:"required"`
}

// GetCohortLTV retrieves LTV for an entire cohort
func (h *AnalyticsHandlersExtended) GetCohortLTV(c *gin.Context) {
	cohortDateStr := c.Query("cohort_date")
	if cohortDateStr == "" {
		response.BadRequest(c, "cohort_date is required")
		return
	}

	cohortDate, err := time.Parse("2006-01-02", cohortDateStr)
	if err != nil {
		response.BadRequest(c, "Invalid cohort_date format")
		return
	}

	cohortLTV, err := h.ltvService.GetCohortLTV(c.Request.Context(), cohortDate)
	if err != nil {
		h.logger.Error("Failed to get cohort LTV",
			zap.String("cohort_date", cohortDateStr),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to get cohort LTV")
		return
	}

	response.OK(c, cohortLTV)
}

// GetChurnRisk predicts the likelihood of user churn
func (h *AnalyticsHandlersExtended) GetChurnRisk(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	risk, err := h.ltvService.PredictChurnRisk(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to predict churn risk",
			zap.String("user_id", userIDStr),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to predict churn risk")
		return
	}

	resp := ChurnRiskResponseExtended{
		UserID:    userIDStr,
		Risk:      risk,
		RiskLevel: getRiskLevelExtended(risk),
		Timestamp: time.Now(),
	}

	response.OK(c, resp)
}

type ChurnRiskResponseExtended struct {
	UserID    string    `json:"user_id"`
	Risk      float64   `json:"risk"`
	RiskLevel string    `json:"risk_level"`
	Timestamp time.Time `json:"timestamp"`
}

func getRiskLevelExtended(risk float64) string {
	if risk >= 0.7 {
		return "high"
	} else if risk >= 0.4 {
		return "medium"
	}
	return "low"
}
