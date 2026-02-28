//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestSubscriptionEndpoints(t *testing.T) {
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

	// Create test user
	userFactory := testutil.NewUserFactory()
	user := userFactory.Create(entity.PlatformiOS, true)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test subscription
	subFactory := testutil.NewSubscriptionFactory()
	sub := subFactory.CreateActive(user.ID, entity.SourceIAP, entity.PlanMonthly)
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	t.Run("GET /v1/subscription - returns subscription details", func(t *testing.T) {
		req, err := testServer.NewAuthenticatedRequest("GET", "/v1/subscription", nil, user.ID.String())
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp testutil.SubscriptionResponse
		err = json.Unmarshal(body, &apiResp)
		require.NoError(t, err)

		assert.Equal(t, sub.ID.String(), apiResp.Data.ID)
		assert.Equal(t, "active", apiResp.Data.Status)
		assert.Equal(t, "monthly", apiResp.Data.PlanType)
	})

	t.Run("GET /v1/subscription/access - returns has_access=true", func(t *testing.T) {
		req, err := testServer.NewAuthenticatedRequest("GET", "/v1/subscription/access", nil, user.ID.String())
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp testutil.AccessCheckResponse
		err = json.Unmarshal(body, &apiResp)
		require.NoError(t, err)

		assert.True(t, apiResp.Data.HasAccess)
		assert.NotEmpty(t, apiResp.Data.ExpiresAt)
	})

	t.Run("GET /v1/subscription - no subscription returns 404", func(t *testing.T) {
		// Create user without subscription
		user2 := userFactory.Create(entity.PlatformAndroid, true)
		err = userRepo.Create(ctx, user2)
		require.NoError(t, err)

		req, err := testServer.NewAuthenticatedRequest("GET", "/v1/subscription", nil, user2.ID.String())
		require.NoError(t, err)

		resp, _, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("DELETE /v1/subscription - cancels subscription", func(t *testing.T) {
		// Create a fresh user and subscription for this test
		cancelUser := userFactory.Create(entity.PlatformiOS, true)
		err = userRepo.Create(ctx, cancelUser)
		require.NoError(t, err)

		cancelSub := subFactory.CreateActive(cancelUser.ID, entity.SourceIAP, entity.PlanMonthly)
		err = subRepo.Create(ctx, cancelSub)
		require.NoError(t, err)

		req, err := testServer.NewAuthenticatedRequest("DELETE", "/v1/subscription", nil, cancelUser.ID.String())
		require.NoError(t, err)

		resp, _, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("GET /v1/subscription/access - expired subscription returns has_access=false", func(t *testing.T) {
		// Create user with expired subscription
		user3 := userFactory.Create(entity.PlatformiOS, true)
		err = userRepo.Create(ctx, user3)
		require.NoError(t, err)

		expiredSub := subFactory.CreateExpired(user3.ID)
		err = subRepo.Create(ctx, expiredSub)
		require.NoError(t, err)

		req, err := testServer.NewAuthenticatedRequest("GET", "/v1/subscription/access", nil, user3.ID.String())
		require.NoError(t, err)

		resp, body, err := testutil.DoRequest(nil, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp testutil.AccessCheckResponse
		json.Unmarshal(body, &apiResp)

		assert.False(t, apiResp.Data.HasAccess)
	})
}
