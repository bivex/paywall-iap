package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/application/query"
	"github.com/bivex/paywall-iap/internal/domain/entity"
	app_handler "github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestSubscriptionEndpoints(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	pool, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	// Run migrations
	err = testutil.RunMigrations(ctx, pool)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := testutil.NewMockUserRepo(pool)
	subscriptionRepo := testutil.NewMockSubscriptionRepo(pool)

	// Create test user
	user := entity.NewUser("test-platform-id", "test-device", entity.PlatformiOS, "1.0", "test@example.com")
	err = pool.QueryRow(ctx,
		"INSERT INTO users (id, platform_user_id, device_id, platform, app_version, email) VALUES ($1, $2, $3, $4, $5, $6)",
		user.ID, user.PlatformUserID, user.DeviceID, user.Platform, user.AppVersion, user.Email,
	).Err()
	require.NoError(t, err)

	// Create test subscription
	sub := entity.NewSubscription(user.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
	err = pool.QueryRow(ctx,
		"INSERT INTO subscriptions (id, user_id, status, source, platform, product_id, plan_type, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		sub.ID, sub.UserID, sub.Status, sub.Source, sub.Platform, sub.ProductID, sub.PlanType, sub.ExpiresAt,
	).Err()
	require.NoError(t, err)

	// Initialize handlers
	jwtMiddleware := middleware.NewJWTMiddleware("test-secret-32-characters!!", nil, 15*time.Minute)
	getSubQuery := query.NewGetSubscriptionQuery(subscriptionRepo)
	checkAccessQuery := query.NewCheckAccessQuery(subscriptionRepo)
	subscriptionHandler := app_handler.NewSubscriptionHandler(
		getSubQuery,
		checkAccessQuery,
		nil, // cancelCmd not needed for this test
		jwtMiddleware,
	)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID.String())
		c.Next()
	})

	v1 := router.Group("/v1")
	subs := v1.Group("/subscription")
	{
		subs.GET("", subscriptionHandler.GetSubscription)
		subs.GET("/access", subscriptionHandler.CheckAccess)
	}

	// Test GET /subscription
	t.Run("GET /subscription returns subscription details", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/subscription", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, sub.ID.String(), data["id"])
		assert.Equal(t, "active", data["status"])
	})

	// Test GET /subscription/access
	t.Run("GET /subscription/access returns access=true", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/subscription/access", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.True(t, data["has_access"].(bool))
		assert.NotEmpty(t, data["expires_at"])
	})
}

func TestAuthEndpoints(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	pool, cleanup, err := testutil.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	// Run migrations
	err = testutil.RunMigrations(ctx, pool)
	require.NoError(t, err)

	// Initialize repositories and handlers
	userRepo := testutil.NewMockUserRepo(pool)
	jwtMiddleware := middleware.NewJWTMiddleware("test-secret-32-characters!!", nil, 15*time.Minute)
	registerCmd := command.NewRegisterCommand(userRepo, jwtMiddleware)
	authHandler := app_handler.NewAuthHandler(registerCmd, jwtMiddleware)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
	}

	// Test POST /auth/register
	t.Run("POST /auth/register creates user and returns tokens", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"platform_user_id": "new-user-id",
			"device_id":        "new-device-id",
			"platform":         "ios",
			"app_version":       "1.0",
			"email":            "newuser@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["user_id"])
		assert.NotEmpty(t, data["access_token"])
		assert.NotEmpty(t, data["refresh_token"])
		assert.Equal(t, float64(900), data["expires_in"]) // 15 minutes
	})

	t.Run("POST /auth/register with duplicate user returns error", func(t *testing.T) {
		// First registration
		reqBody := map[string]interface{}{
			"platform_user_id": "duplicate-user",
			"device_id":        "device-1",
			"platform":         "ios",
			"app_version":       "1.0",
			"email":            "duplicate@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/v1/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req)

		assert.Equal(t, http.StatusCreated, w1.Code)

		// Duplicate registration
		req = httptest.NewRequest("POST", "/v1/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req)

		assert.Equal(t, http.StatusBadRequest, w2.Code)
	})
}
