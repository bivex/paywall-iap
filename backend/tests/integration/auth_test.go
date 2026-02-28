//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestAuthRegistration(t *testing.T) {
	ctx := context.Background()

	// Setup test database container
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := testutil.NewMockUserRepo(dbContainer.Pool)
	subRepo := testutil.NewMockSubscriptionRepo(dbContainer.Pool)

	// Create test server
	testServer := testutil.NewTestServer(ctx, dbContainer.Pool, userRepo, subRepo)
	defer testServer.Close()

	t.Run("POST /v1/auth/register - successful registration", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"platform_user_id": "test-platform-user-" + time.Now().Format("20060102150405"),
			"device_id":        "test-device-id",
			"platform":         "ios",
			"app_version":      "1.0.0",
			"email":            "test_" + time.Now().Format("20060102150405") + "@example.com",
		}

		req, err := testServer.NewRequest("POST", "/v1/auth/register", reqBody)
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var apiResp testutil.AuthResponse
		err = json.Unmarshal(body, &apiResp)
		require.NoError(t, err)
		assert.NotEmpty(t, apiResp.Data.UserID)
		assert.NotEmpty(t, apiResp.Data.AccessToken)
		assert.NotEmpty(t, apiResp.Data.RefreshToken)
		assert.Equal(t, 900, apiResp.Data.ExpiresIn) // 15 minutes
	})

	t.Run("POST /v1/auth/register - duplicate platform_user_id", func(t *testing.T) {
		platformUserID := "duplicate-user-" + time.Now().Format("20060102150405")

		// First registration
		reqBody1 := map[string]interface{}{
			"platform_user_id": platformUserID,
			"device_id":        "device-1",
			"platform":         "ios",
			"app_version":      "1.0.0",
			"email":            "test1_" + time.Now().Format("20060102150405") + "@example.com",
		}

		req1, _ := testServer.NewRequest("POST", "/v1/auth/register", reqBody1)
		resp1, _, _ := testutil.DoRequest(nil, req1)
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		// Duplicate registration
		reqBody2 := map[string]interface{}{
			"platform_user_id": platformUserID,
			"device_id":        "device-2",
			"platform":         "ios",
			"app_version":      "1.0.0",
			"email":            "test2_" + time.Now().Format("20060102150405") + "@example.com",
		}

		req2, _ := testServer.NewRequest("POST", "/v1/auth/register", reqBody2)
		resp2, body2, _ := testutil.DoRequest(nil, req2)

		assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

		var errorResp map[string]interface{}
		json.Unmarshal(body2, &errorResp)

		assert.Contains(t, errorResp, "error")
		assert.Equal(t, "INVALID_REQUEST", errorResp["error"])
	})

	t.Run("POST /v1/auth/register - invalid platform", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"platform_user_id": "test-user",
			"device_id":        "device-id",
			"platform":         "invalid_platform",
			"app_version":      "1.0.0",
			"email":            "test@example.com",
		}

		req, _ := testServer.NewRequest("POST", "/v1/auth/register", reqBody)
		resp, body, _ := testutil.DoRequest(nil, req)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)
		assert.Contains(t, errorResp, "error")
		assert.Equal(t, "INVALID_REQUEST", errorResp["error"])
	})

	t.Run("POST /v1/auth/register - missing required fields", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"device_id": "device-id",
			// Missing platform_user_id, platform, app_version
		}

		req, _ := testServer.NewRequest("POST", "/v1/auth/register", reqBody)
		resp, _, _ := testutil.DoRequest(nil, req)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
