package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	matomoClient "github.com/bivex/paywall-iap/internal/infrastructure/external/matomo"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// GetFunnelsRequest represents the request to get funnel data
type GetFunnelsRequest struct {
	FunnelID  string `form:"funnel_id" binding:"required"`
	DateFrom  string `form:"date_from" binding:"required"`
	DateTo    string `form:"date_to" binding:"required"`
	Segment   string `form:"segment"`
	UseCache  bool   `form:"use_cache"`
}

// GetFunnels retrieves funnel analysis data from Matomo
// @Summary Get funnel analysis
// @Tags analytics
// @Produce json
// @Param funnel_id query string true "Funnel ID"
// @Param date_from query string true "Start date (YYYY-MM-DD)"
// @Param date_to query string true "End date (YYYY-MM-DD)"
// @Param segment query string false "Segment definition"
// @Param use_cache query bool false "Use cached data"
// @Success 200 {object} response.SuccessResponse{data=matomoClient.FunnelResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/analytics/funnels [get]
func (h *AnalyticsHandler) GetFunnels(c *gin.Context) {
	var req GetFunnelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// Parse dates
	dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
	if err != nil {
		response.BadRequest(c, "Invalid date_from format")
		return
	}

	dateTo, err := time.Parse("2006-01-02", req.DateTo)
	if err != nil {
		response.BadRequest(c, "Invalid date_to format")
		return
	}

	// Check cache first if enabled
	cacheKey := ""
	if req.UseCache {
		cacheKey = fmt.Sprintf("analytics:funnel:%s:%s:%s", req.FunnelID, req.DateFrom, req.DateTo)
		if cached, err := h.getFromCache(c, cacheKey); err == nil && cached != nil {
			response.OK(c, cached)
			return
		}
	}

	// Build Matomo request
	matomoReq := matomoClient.FunnelRequest{
		FunnelID: req.FunnelID,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Segment:  req.Segment,
	}

	// Fetch from Matomo
	funnelData, err := h.matomoClient.GetFunnels(c.Request.Context(), matomoReq)
	if err != nil {
		h.logger.Error("Failed to fetch funnel data from Matomo",
			zap.String("funnel_id", req.FunnelID),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to fetch funnel data")
		return
	}

	// Cache the result
	if req.UseCache && cacheKey != "" {
		h.setToCache(c, cacheKey, funnelData, 30*time.Minute)
	}

	response.OK(c, funnelData)
}

// GetCohortsRequest represents the request to get cohort data
type GetCohortsRequest struct {
	Segment      string `form:"segment"`
	DateFrom     string `form:"date_from" binding:"required"`
	DateTo       string `form:"date_to" binding:"required"`
	CohortPeriod string `form:"cohort_period" binding:"required"`
	UseCache     bool   `form:"use_cache"`
}

// GetCohorts retrieves cohort analysis data from Matomo
// @Summary Get cohort analysis
// @Tags analytics
// @Produce json
// @Param date_from query string true "Start date (YYYY-MM-DD)"
// @Param date_to query string true "End date (YYYY-MM-DD)"
// @Param cohort_period query string true "Cohort period (day/week/month)"
// @Param segment query string false "Segment definition"
// @Param use_cache query bool false "Use cached data"
// @Success 200 {object} response.SuccessResponse{data=matomoClient.CohortResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/analytics/cohorts [get]
func (h *AnalyticsHandler) GetCohorts(c *gin.Context) {
	var req GetCohortsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// Parse dates
	dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
	if err != nil {
		response.BadRequest(c, "Invalid date_from format")
		return
	}

	dateTo, err := time.Parse("2006-01-02", req.DateTo)
	if err != nil {
		response.BadRequest(c, "Invalid date_to format")
		return
	}

	// Check cache first if enabled
	cacheKey := ""
	if req.UseCache {
		cacheKey = fmt.Sprintf("analytics:cohort:%s:%s:%s", req.CohortPeriod, req.DateFrom, req.DateTo)
		if cached, err := h.getFromCache(c, cacheKey); err == nil && cached != nil {
			response.OK(c, cached)
			return
		}
	}

	// Build Matomo request
	matomoReq := matomoClient.CohortRequest{
		Segment:      req.Segment,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		CohortPeriod: req.CohortPeriod,
	}

	// Fetch from Matomo
	cohortData, err := h.matomoClient.GetCohorts(c.Request.Context(), matomoReq)
	if err != nil {
		h.logger.Error("Failed to fetch cohort data from Matomo",
			zap.String("cohort_period", req.CohortPeriod),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to fetch cohort data")
		return
	}

	// Cache the result
	if req.UseCache && cacheKey != "" {
		h.setToCache(c, cacheKey, cohortData, 1*time.Hour) // Cache cohorts longer
	}

	response.OK(c, cohortData)
}

// GetRealtimeRequest represents the request to get realtime data
type GetRealtimeRequest struct {
	Minutes int `form:"minutes" binding:"min=1,max=60"`
	Limit   int `form:"limit" binding:"min=1,max=1000"`
}

// GetRealtime retrieves realtime visitor data from Matomo
// @Summary Get realtime visitors
// @Tags analytics
// @Produce json
// @Param minutes query int false "Last N minutes (default: 30)"
// @Param limit query int false "Max results (default: 100)"
// @Success 200 {object} response.SuccessResponse{data=[]matomoClient.RealtimeVisitor}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/analytics/realtime [get]
func (h *AnalyticsHandler) GetRealtime(c *gin.Context) {
	var req GetRealtimeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Minutes = 30
		req.Limit = 100
	}
	if req.Minutes == 0 {
		req.Minutes = 30
	}
	if req.Limit == 0 {
		req.Limit = 100
	}

	// Fetch from Matomo
	visitors, err := h.matomoClient.GetRealtimeVisitors(c.Request.Context(), req.Minutes, req.Limit)
	if err != nil {
		h.logger.Error("Failed to fetch realtime data from Matomo", zap.Error(err))
		response.InternalError(c, "Failed to fetch realtime data")
		return
	}

	response.OK(c, visitors)
}

// GetLTVRequest represents the request to get LTV data
type GetLTVRequest struct {
	UserID string `form:"user_id" binding:"required,uuid"`
}

// GetLTV retrieves LTV estimates for a user
// @Summary Get user LTV estimates
// @Tags analytics
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {object} response.SuccessResponse{data=LTVResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/analytics/ltv [get]
func (h *AnalyticsHandler) GetLTV(c *gin.Context) {
	var req GetLTVRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	// Check cache first
	cacheKey := fmt.Sprintf("analytics:ltv:%s", userID)
	if cached, err := h.getFromCache(c, cacheKey); err == nil && cached != nil {
		response.OK(c, cached)
		return
	}

	// Calculate LTV estimates
	// TODO: Implement actual LTV calculation using cohort data
	// For now, return mock data
	ltvResponse := LTVResponse{
		UserID: userID.String(),
		Estimates: map[string]float64{
			"ltv30":  9.99,
			"ltv90":  29.97,
			"ltv365": 119.88,
		},
		CalculatedAt: time.Now(),
	}

	// Cache for 1 hour
	h.setToCache(c, cacheKey, ltvResponse, 1*time.Hour)

	response.OK(c, ltvResponse)
}

// LTVResponse represents LTV estimates for a user
type LTVResponse struct {
	UserID      string            `json:"user_id"`
	Estimates   map[string]float64 `json:"estimates"` // ltv30, ltv90, ltv365
	CalculatedAt time.Time        `json:"calculated_at"`
}

// Cache helper functions

func (h *AnalyticsHandler) getFromCache(c *gin.Context, key string) (interface{}, error) {
	data, err := h.redisClient.Get(c.Request.Context(), key).Bytes()
	if err != nil {
		return nil, err
	}

	// Deserialize from JSON
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (h *AnalyticsHandler) setToCache(c *gin.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return h.redisClient.Set(c.Request.Context(), key, data, ttl).Err()
}

// Health check for analytics service
// @Summary Analytics health check
// @Tags analytics
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/analytics/health [get]
func (h *AnalyticsHandler) Health(c *gin.Context) {
	response.OK(c, gin.H{
		"status": "healthy",
		"service": "analytics",
	})
}
