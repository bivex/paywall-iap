package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/cache"
	"github.com/bivex/paywall-iap/internal/infrastructure/external/matomo"
	"github.com/bivex/paywall-iap/tests/testutil"
)

// MockMatomoClient mocks the Matomo HTTP client
type MockMatomoClient struct {
	mock.Mock
}

func (m *MockMatomoClient) TrackEvent(ctx context.Context, req matomo.TrackEventRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockMatomoClient) TrackEcommerce(ctx context.Context, req matomo.TrackEcommerceRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockMatomoClient) GetCohorts(ctx context.Context, req matomo.CohortRequest) (*matomo.CohortResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*matomo.CohortResponse), args.Error(1)
}

func (m *MockMatomoClient) GetFunnels(ctx context.Context, req matomo.FunnelRequest) (*matomo.FunnelResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*matomo.FunnelResponse), args.Error(1)
}

func (m *MockMatomoClient) GetRealtimeVisitors(ctx context.Context, minutes int, limit int) ([]matomo.RealtimeVisitor, error) {
	args := m.Called(ctx, minutes, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]matomo.RealtimeVisitor), args.Error(1)
}

func (m *MockMatomoClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMatomoClient) Close() {
	// No-op for mock
}

// TestAnalyticsAPI tests the analytics HTTP endpoints
func TestAnalyticsAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := context.Background()
	logger := zap.NewNop()

	// Setup test Redis
	redisClient := testutil.SetupTestRedis(t)
	defer testutil.TeardownTestRedis(t, redisClient)

	analyticsCache := cache.NewAnalyticsCache(redisClient, logger)

	t.Run("CohortRetrieval", func(t *testing.T) {
		// Create mock cohort data
		cohortData := &matomo.CohortResponse{
			Cohorts: []matomo.CohortData{
				{
					Period: "2026-03-01",
					Retention: map[string]int{
						"day0":  100,
						"day1":  85,
						"day7":  60,
						"day30": 40,
					},
					SampleSize: 1000,
					Metrics: map[string]interface{}{
						"revenue": map[string]interface{}{
							"day1":  999.0,
							"day7":  5994.0,
							"day30": 19980.0,
						},
					},
				},
			},
			Meta: matomo.CohortMeta{
				TotalUsers:       1000,
				AverageRetention: 0.71,
			},
		}

		mockMatomo := new(MockMatomoClient)
		mockMatomo.On("GetCohorts", ctx, mock.Anything).Return(cohortData, nil)

		// Store in cache
		cacheKey := "cohort_day_2026-03-01"
		err := analyticsCache.SetCohortData(ctx, cacheKey, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), &cache.CohortData{
			MetricName: cacheKey,
			Date:       time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			CohortSize: 1000,
			Retention:  cohortData.Cohorts[0].Retention,
			Revenue:    cohortData.Cohorts[0].Metrics["revenue"].(map[string]interface{}),
		})
		assert.NoError(t, err)

		// Retrieve from cache
		cachedData, err := analyticsCache.GetCohortData(ctx, cacheKey, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
		assert.NoError(t, err)
		assert.Equal(t, 1000, cachedData.CohortSize)
		assert.Equal(t, 85, cachedData.Retention["day1"])
	})

	t.Run("LTVCalculationAndCaching", func(t *testing.T) {
		// Setup test database
		db := testutil.SetupTestDB(t)
		defer testutil.TeardownTestDB(t, db)

		mockMatomo := new(MockMatomoClient)

		// Create mock cohort worker
		mockCohortWorker := new(MockCohortWorker)
		mockCohortWorker.On("CalculateLTVFromCohorts", ctx, mock.Anything, mock.Anything).Return(map[string]float64{
			"ltv30":  9.99,
			"ltv90":  29.97,
			"ltv365": 119.88,
		}, nil)

		ltvService := service.NewLTVService(
			mockMatomo,
			mockCohortWorker,
			nil, // subscriptionRepo - will use defaults
			logger,
		)

		userID := uuid.New()

		// Calculate LTV
		estimates, err := ltvService.CalculateLTV(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, estimates)

		// Cache the result
		ltvData := &cache.LTVData{
			UserID:       userID.String(),
			LTV30:        estimates.LTV30,
			LTV90:        estimates.LTV90,
			LTV365:       estimates.LTV365,
			LTVLifetime:  estimates.LTVLifetime,
			Confidence:   estimates.Confidence,
			CalculatedAt: estimates.CalculatedAt,
			Factors:      estimates.Factors,
		}

		err = analyticsCache.SetLTV(ctx, userID.String(), ltvData)
		assert.NoError(t, err)

		// Retrieve from cache
		cachedLTV, err := analyticsCache.GetLTV(ctx, userID.String())
		assert.NoError(t, err)
		assert.Equal(t, estimates.LTV30, cachedLTV.LTV30)
		assert.Equal(t, estimates.LTV90, cachedLTV.LTV90)

		// Invalidate cache
		err = analyticsCache.InvalidateLTV(ctx, userID.String())
		assert.NoError(t, err)

		// Verify cache miss after invalidation
		_, err = analyticsCache.GetLTV(ctx, userID.String())
		assert.Error(t, err) // Cache miss
	})

	t.Run("RealtimeMetricsCaching", func(t *testing.T) {
		metricName := "active_users"

		// Set realtime metric
		metric := &cache.RealtimeMetric{
			Name:      metricName,
			Value:     1234.0,
			Timestamp: time.Now(),
			Tags: map[string]string{
				"platform": "ios",
			},
		}

		err := analyticsCache.SetRealtimeMetric(ctx, metric)
		assert.NoError(t, err)

		// Retrieve from cache
		cachedMetric, err := analyticsCache.GetRealtimeMetric(ctx, metricName)
		assert.NoError(t, err)
		assert.Equal(t, metric.Value, cachedMetric.Value)
		assert.Equal(t, metric.Tags["platform"], cachedMetric.Tags["platform"])

		// Test atomic increment
		err = analyticsCache.IncrementRealtimeMetric(ctx, metricName, 10.0)
		assert.NoError(t, err)

		// Get updated value
		updatedMetric, err := analyticsCache.GetRealtimeMetric(ctx, metricName)
		assert.NoError(t, err)
		assert.InDelta(t, 1244.0, updatedMetric.Value, 1.0)
	})

	t.Run("FunnelDataCaching", func(t *testing.T) {
		// Create mock funnel data
		funnelData := &matomo.FunnelResponse{
			FunnelID:      "funnel_purchase",
			FunnelName:    "Purchase Funnel",
			TotalEntries:  1000,
			TotalExits:    700,
			ConversionRate: 0.30,
			Steps: []matomo.FunnelStep{
				{
					StepID:      "step1",
					StepName:    "View Product",
					Visitors:    1000,
					Dropoff:     0,
					DropoffRate: 0.0,
				},
				{
					StepID:      "step2",
					StepName:    "Add to Cart",
					Visitors:    600,
					Dropoff:     400,
					DropoffRate: 0.4,
				},
				{
					StepID:      "step3",
					StepName:    "Purchase",
					Visitors:    300,
					Dropoff:     300,
					DropoffRate: 0.5,
				},
			},
		}

		dateFrom := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		dateTo := time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC)

		// Cache funnel data
		err := analyticsCache.SetFunnelData(ctx, "funnel_purchase", dateFrom, dateTo, &cache.FunnelData{
			FunnelID:        funnelData.FunnelID,
			FunnelName:      funnelData.FunnelName,
			Steps:           convertFunnelSteps(funnelData.Steps),
			TotalEntries:    funnelData.TotalEntries,
			TotalExits:      funnelData.TotalExits,
			ConversionRate:  funnelData.ConversionRate,
		})
		assert.NoError(t, err)

		// Retrieve from cache
		cachedFunnel, err := analyticsCache.GetFunnelData(ctx, "funnel_purchase", dateFrom, dateTo)
		assert.NoError(t, err)
		assert.Equal(t, "funnel_purchase", cachedFunnel.FunnelID)
		assert.Equal(t, 1000, cachedFunnel.TotalEntries)
		assert.Len(t, cachedFunnel.Steps, 3)
	})
}

// MockCohortWorker mocks the CohortWorker
type MockCohortWorker struct {
	mock.Mock
}

func (m *MockCohortWorker) CalculateLTVFromCohorts(ctx context.Context, userID uuid.UUID) (map[string]float64, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

// TestAnalyticsHTTPEndpoints tests the HTTP endpoints
func TestAnalyticsHTTPEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Setup test infrastructure
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	redisClient := testutil.SetupTestRedis(t)
	defer testutil.TeardownTestRedis(t, redisClient)

	logger := zap.NewNop()
	ctx := context.Background()

	mockMatomo := new(MockMatomoClient)
	mockCohortWorker := new(MockCohortWorker)

	ltvService := service.NewLTVService(mockMatomo, mockCohortWorker, nil, logger)
	analyticsCache := cache.NewAnalyticsCache(redisClient, logger)

	// Setup Gin router
	router := setupAnalyticsRouter(ltvService, analyticsCache)

	t.Run("GET /api/v1/analytics/ltv", func(t *testing.T) {
		userID := uuid.New()

		// Mock LTV calculation
		mockCohortWorker.On("CalculateLTVFromCohorts", ctx, userID).Return(map[string]float64{
			"ltv30":  9.99,
			"ltv90":  29.97,
			"ltv365": 119.88,
		}, nil)

		// Create request
		req := httptest.NewRequest("GET", "/api/v1/analytics/ltv?user_id="+userID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify response structure
		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, userID.String(), data["user_id"])
		assert.InDelta(t, 9.99, data["ltv30"], 0.01)
	})

	t.Run("GET /api/v1/analytics/ltv with cache hit", func(t *testing.T) {
		userID := uuid.New()

		// Pre-populate cache
		ltvData := &cache.LTVData{
			UserID:       userID.String(),
			LTV30:        19.99,
			LTV90:        59.97,
			LTV365:       239.88,
			LTVLifetime:  19.99,
			Confidence:   0.8,
			CalculatedAt: time.Now(),
			Factors:      map[string]float64{},
		}

		err := analyticsCache.SetLTV(ctx, userID.String(), ltvData)
		require.NoError(t, err)

		// Create request
		req := httptest.NewRequest("GET", "/api/v1/analytics/ltv?user_id="+userID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok)
		// Should return cached value
		assert.InDelta(t, 19.99, data["ltv30"], 0.01)
	})

	t.Run("POST /api/v1/analytics/ltv", func(t *testing.T) {
		userID := uuid.New()

		// Create request body
		reqBody := map[string]interface{}{
			"user_id": userID.String(),
			"amount":  19.99,
		}

		bodyJSON, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/analytics/ltv", bodyJSON)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify LTV update was called
		message, ok := response["message"].(string)
		assert.True(t, ok)
		assert.Contains(t, message, "LTV updated")
	})

	t.Run("GET /api/v1/analytics/realtime", func(t *testing.T) {
		// Set up some realtime metrics
		metrics := []cache.RealtimeMetric{
			{Name: "active_users", Value: 5432.0, Timestamp: time.Now()},
			{Name: "page_views", Value: 12345.0, Timestamp: time.Now()},
		}

		for _, metric := range metrics {
			err := analyticsCache.SetRealtimeMetric(ctx, &metric)
			require.NoError(t, err)
		}

		req := httptest.NewRequest("GET", "/api/v1/analytics/realtime?metrics=active_users", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(data), 1)
	})

	t.Run("GET /api/v1/analytics/cache/stats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/analytics/cache/stats", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, data, "keys")
		assert.Contains(t, data, "memory")
	})
}

// Helper functions
func setupAnalyticsRouter(ltvService *service.LTVService, analyticsCache *cache.AnalyticsCache) *http.ServeMux {
	router := http.NewServeMux()

	// Register handlers (simplified for testing)
	router.HandleFunc("/api/v1/analytics/ltv", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			userIDStr := r.URL.Query().Get("user_id")
			userID, _ := uuid.Parse(userIDStr)

			// Try cache first
			cachedLTV, err := analyticsCache.GetLTV(r.Context(), userIDStr)
			if err == nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": cachedLTV,
				})
				return
			}

			// Calculate LTV
			estimates, _ := ltvService.CalculateLTV(r.Context(), userID)

			// Cache result
			ltvData := &cache.LTVData{
				UserID:      userIDStr,
				LTV30:       estimates.LTV30,
				LTV90:       estimates.LTV90,
				LTV365:      estimates.LTV365,
				LTVLifetime: estimates.LTVLifetime,
				Confidence:  estimates.Confidence,
				CalculatedAt: estimates.CalculatedAt,
				Factors:     estimates.Factors,
			}
			analyticsCache.SetLTV(r.Context(), userIDStr, ltvData)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": estimates,
			})
		} else if r.Method == "POST" {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			userID, _ := uuid.Parse(req["user_id"].(string))
			amount := req["amount"].(float64)

			ltvService.UpdateUserLTV(r.Context(), userID, amount)
			analyticsCache.InvalidateLTV(r.Context(), userID.String())

			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "LTV updated successfully",
				"user_id":  req["user_id"],
				"amount":   amount,
			})
		}
	})

	router.HandleFunc("/api/v1/analytics/realtime", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		metricsParam := r.URL.Query().Get("metrics")
		if metricsParam == "" {
			metricsParam = "active_users"
		}

		metrics, _ := analyticsCache.GetRealtimeMetrics(r.Context(), []string{metricsParam})

		result := make([]*cache.RealtimeMetric, 0, len(metrics))
		for _, metric := range metrics {
			result = append(result, metric)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": result,
		})
	})

	router.HandleFunc("/api/v1/analytics/cache/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		stats, _ := analyticsCache.GetCacheStats(r.Context())

		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": stats,
		})
	})

	return router
}

func convertFunnelSteps(steps []matomo.FunnelStep) []cache.FunnelStep {
	result := make([]cache.FunnelStep, len(steps))
	for i, step := range steps {
		result[i] = cache.FunnelStep{
			StepID:      step.StepID,
			StepName:    step.StepName,
			Visitors:    step.Visitors,
			Dropoff:     step.Dropoff,
			DropoffRate: step.DropoffRate,
		}
	}
	return result
}
