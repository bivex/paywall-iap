package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/application/query"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
)

// TestServer holds the test HTTP server and dependencies
type TestServer struct {
	Server           *httptest.Server
	Router           *gin.Engine
	Pool             *pgxpool.Pool
	UserRepo         repository.UserRepository
	SubscriptionRepo repository.SubscriptionRepository
	JWTMiddleware    *middleware.JWTMiddleware
}

// NewTestServer creates a new test server with all handlers configured
func NewTestServer(
	ctx context.Context,
	pool *pgxpool.Pool,
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
) *TestServer {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtMiddleware := middleware.NewJWTMiddleware("test-secret-32-characters!!", nil, 15*time.Minute)

	// Initialize queries and commands
	getSubQuery := query.NewGetSubscriptionQuery(subscriptionRepo)
	checkAccessQuery := query.NewCheckAccessQuery(subscriptionRepo)
	cancelCmd := command.NewCancelSubscriptionCommand(subscriptionRepo)
	registerCmd := command.NewRegisterCommand(userRepo, jwtMiddleware)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(registerCmd, jwtMiddleware)
	subscriptionHandler := handlers.NewSubscriptionHandler(getSubQuery, checkAccessQuery, cancelCmd, jwtMiddleware)

	// Setup routes
	v1 := router.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		subscription := v1.Group("/subscription")
		subscription.Use(func(c *gin.Context) {
			// In tests we inject user_id directly via NewAuthenticatedRequest
			// The JWT middleware requires Redis; use a simplified in-test auth
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "UNAUTHORIZED"})
				c.Abort()
				return
			}
			// Extract bearer token and parse
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenStr := authHeader[7:]
				claims, err := jwtMiddleware.ParseToken(tokenStr)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "UNAUTHORIZED"})
					c.Abort()
					return
				}
				c.Set("user_id", claims.UserID)
			}
			c.Next()
		})
		{
			subscription.GET("", subscriptionHandler.GetSubscription)
			subscription.GET("/access", subscriptionHandler.CheckAccess)
			subscription.DELETE("", subscriptionHandler.CancelSubscription)
		}
	}

	server := httptest.NewServer(router)

	return &TestServer{
		Server:           server,
		Router:           router,
		Pool:             pool,
		UserRepo:         userRepo,
		SubscriptionRepo: subscriptionRepo,
		JWTMiddleware:    jwtMiddleware,
	}
}

// Close shuts down the test server
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// BaseURL returns the test server URL
func (ts *TestServer) BaseURL() string {
	return ts.Server.URL
}

// NewRequest creates a new HTTP request without authentication
func (ts *TestServer) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	return NewTestRequest(method, ts.BaseURL()+path, body, "")
}

// NewAuthenticatedRequest creates a new authenticated HTTP request
func (ts *TestServer) NewAuthenticatedRequest(method, path string, body interface{}, userID string) (*http.Request, error) {
	accessToken, _, err := ts.JWTMiddleware.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}
	return NewTestRequest(method, ts.BaseURL()+path, body, accessToken)
}
