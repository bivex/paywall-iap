//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
)

func TestABTestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup feature flag service
	ffService := service.NewFeatureFlagService()
	abHandler := handlers.NewABTestHandler(ffService)

	// Setup router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test_user_123")
		c.Next()
	})

	v1 := router.Group("/v1/ab-test")
	{
		v1.GET("/flags", abHandler.GetFeatureFlags)
		v1.GET("/evaluate/:flag_id", abHandler.EvaluateFlag)
		v1.GET("/paywall", abHandler.EvaluatePaywall)
		v1.POST("/flags", abHandler.CreateFlag)
		v1.PUT("/flags/:flag_id", abHandler.UpdateFlag)
		v1.DELETE("/flags/:flag_id", abHandler.DeleteFlag)
	}

	t.Run("GET /ab-test/flags returns empty list initially", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/ab-test/flags", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		assert.Empty(t, data)
	})

	t.Run("POST /ab-test/flags creates feature flag", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"id":              "test_paywall",
			"name":            "Test Paywall Variant",
			"enabled":         true,
			"rollout_percent": 50,
			"user_ids":        []string{"beta_user_1"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/ab-test/flags", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "test_paywall", data["id"])
		assert.Equal(t, "Test Paywall Variant", data["name"])
	})

	t.Run("GET /ab-test/evaluate/test_paywall returns evaluation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/ab-test/evaluate/test_paywall", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "test_paywall", data["flag_id"])
		assert.Equal(t, "test_user_123", data["user_id"])
	})

	t.Run("GET /ab-test/paywall returns paywall variant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/ab-test/paywall", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		// Since we haven't created "paywall_variant_test" flag, it should be "control"
		assert.Equal(t, "control", data["variant"])
	})

	t.Run("PUT /ab-test/flags/test_paywall updates flag", func(t *testing.T) {
		enabled := false
		reqBody := map[string]interface{}{
			"enabled": &enabled,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/v1/ab-test/flags/test_paywall", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DELETE /ab-test/flags/test_paywall deletes flag", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/ab-test/flags/test_paywall", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
