package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	matomoClient "github.com/bivex/paywall-iap/internal/infrastructure/external/matomo"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// AnalyticsHandler handles HTTP requests for analytics data
type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
	matomoClient     *matomoClient.Client
	redisClient      *redis.Client
	logger           *zap.Logger
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(
	analyticsService *service.AnalyticsService,
	matomoClient *matomoClient.Client,
	redisClient *redis.Client,
	logger *zap.Logger,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		matomoClient:     matomoClient,
		redisClient:      redisClient,
		logger:           logger,
	}
}

// GetRevenueMetrics returns revenue-related metrics
// @Summary Get revenue metrics
// @Tags analytics
// @Produce json
// @Router /api/v1/analytics/revenue [get]
func (h *AnalyticsHandler) GetRevenueMetrics(c *gin.Context) {
	// Default to last 30 days
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	metrics, err := h.analyticsService.CalculateRevenueMetrics(c.Request.Context(), start, end)
	if err != nil {
		response.InternalError(c, "Failed to calculate revenue metrics")
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetChurnMetrics returns churn-related metrics
// @Summary Get churn metrics
// @Tags analytics
// @Produce json
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Router /api/v1/analytics/churn [get]
func (h *AnalyticsHandler) GetChurnMetrics(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	end := time.Now()
	if endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			end = t
		}
	}

	start := end.AddDate(0, 0, -30)
	if startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			start = t
		}
	}

	metrics, err := h.analyticsService.CalculateChurnMetrics(c.Request.Context(), start, end)
	if err != nil {
		response.InternalError(c, "Failed to calculate churn metrics")
		return
	}

	c.JSON(http.StatusOK, metrics)
}
