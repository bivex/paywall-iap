//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestUserJourney(t *testing.T) {
	ctx := context.Background()

	// Setup E2E environment
	suite := SetupE2ETestSuite(ctx, t)
	defer suite.Teardown(ctx, t)

	platformUserID := "journey_user_" + time.Now().Format("20060102150405")
	var accessToken string

	t.Run("Step 1: Register User", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"platform_user_id": platformUserID,
			"device_id":        "device-journey",
			"platform":         "ios",
			"app_version":      "1.0.0",
			"email":            platformUserID + "@example.com",
		}

		req, err := suite.APIServer.NewRequest("POST", "/v1/auth/register", reqBody)
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var apiResp testutil.AuthResponse
		err = json.Unmarshal(body, &apiResp)
		require.NoError(t, err)

		assert.NotEmpty(t, apiResp.Data.AccessToken)

		// Store access token for subsequent steps
		accessToken = apiResp.Data.AccessToken
	})

	t.Run("Step 2: Check Access before Subscription", func(t *testing.T) {
		require.NotEmpty(t, accessToken, "Access token is required")

		req, err := testutil.NewTestRequest("GET", suite.GetAPIURL()+"/v1/subscription/access", nil, accessToken)
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp testutil.AccessCheckResponse
		err = json.Unmarshal(body, &apiResp)
		require.NoError(t, err)

		// Access should be false initially
		assert.False(t, apiResp.Data.HasAccess)
	})

	// To complete the journey with real sub we'd need a mock Apple/Google server or we can create it in DB directly
	// For E2E let's use the DB directly to simulate successful IAP purchase processing
	t.Run("Step 3: Simulate IAP Webhook/Creation", func(t *testing.T) {
		// Mock a subscription using our factories
		user, err := suite.APIServer.UserRepo.GetByPlatformID(ctx, platformUserID)
		require.NoError(t, err)

		subFactory := testutil.NewSubscriptionFactory()
		sub := subFactory.CreateActive(user.ID, entity.SourceIAP, entity.PlanMonthly)

		err = suite.APIServer.SubscriptionRepo.Create(ctx, sub)
		require.NoError(t, err)
	})

	t.Run("Step 4: Check Access after Subscription", func(t *testing.T) {
		require.NotEmpty(t, accessToken, "Access token is required")

		req, err := testutil.NewTestRequest("GET", suite.GetAPIURL()+"/v1/subscription/access", nil, accessToken)
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp testutil.AccessCheckResponse
		err = json.Unmarshal(body, &apiResp)
		require.NoError(t, err)

		assert.True(t, apiResp.Data.HasAccess)
	})

	t.Run("Step 5: Cancel Subscription", func(t *testing.T) {
		require.NotEmpty(t, accessToken, "Access token is required")

		req, err := testutil.NewTestRequest("DELETE", suite.GetAPIURL()+"/v1/subscription", nil, accessToken)
		require.NoError(t, err)

		resp, _, err := testutil.DoRequest(suite.HTTPClient, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
