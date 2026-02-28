package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/application/query"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/config"
	"github.com/bivex/paywall-iap/internal/infrastructure/external/iap"
	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/pool"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	app_handler "github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
	http_response "github.com/bivex/paywall-iap/internal/interfaces/http/response"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	if err := logging.Init(&cfg.Sentry); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logging.Sync()

	logging.Logger.Info("Starting IAP API server",
		zap.Int("port", cfg.Server.Port),
		zap.String("environment", cfg.Sentry.Environment),
	)

	// Initialize database connection
	ctx := context.Background()
	dbPool, err := pool.NewPool(ctx, cfg.Database)
	if err != nil {
		logging.Logger.Fatal("Failed to create database pool", zap.Error(err))
	}
	defer pool.Close(dbPool)

	// Test database connection
	if err := pool.Ping(ctx, dbPool); err != nil {
		logging.Logger.Fatal("Failed to ping database", zap.Error(err))
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.URL,
		Password: cfg.Redis.Password,
		PoolSize: cfg.Redis.PoolSize,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logging.Logger.Fatal("Failed to ping Redis", zap.Error(err))
	}

	// Initialize repositories (using mock sqlc queries for now)
	queries := &generated.Queries{} // Mock - needs actual initialization
	userRepo := repository.NewUserRepository(queries)
	subscriptionRepo := repository.NewSubscriptionRepository(queries)
	transactionRepo := repository.NewTransactionRepository(queries)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(
		cfg.JWT.Secret,
		redisClient,
		cfg.JWT.AccessTTL,
	)
	rateLimiter := middleware.NewRateLimiter(redisClient, true) // fail open

	// Initialize IAP verifiers
	appleVerifier := iap.NewAppleVerifier(cfg.IAP.AppleSharedSecret, true) // sandbox mode
	googleVerifier := iap.NewGoogleVerifier(cfg.IAP.GoogleKeyJSON, false)

	// Initialize commands
	registerCmd := command.NewRegisterCommand(userRepo, jwtMiddleware)
	iapAdapter := iap.NewIAPAdapter(appleVerifier, googleVerifier)
	verifyIAPCmd := command.NewVerifyIAPCommand(
		userRepo,
		subscriptionRepo,
		transactionRepo,
		iapAdapter, // iOS verifier
		iapAdapter, // Android verifier
	)
	cancelSubCmd := command.NewCancelSubscriptionCommand(subscriptionRepo)

	// Initialize queries
	getSubQuery := query.NewGetSubscriptionQuery(subscriptionRepo)
	checkAccessQuery := query.NewCheckAccessQuery(subscriptionRepo)

	// Initialize handlers
	authHandler := app_handler.NewAuthHandler(registerCmd, jwtMiddleware)
	subscriptionHandler := app_handler.NewSubscriptionHandler(
		getSubQuery,
		checkAccessQuery,
		cancelSubCmd,
		jwtMiddleware,
	)

	// Setup Gin router
	if cfg.Sentry.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		gin.Recovery(),
		logging.RequestMiddleware(logging.Logger),
	)

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh",
				rateLimiter.Middleware(middleware.ByIP, middleware.DefaultConfig),
				authHandler.RefreshToken,
			)
		}

		// Protected routes (require JWT)
		protected := v1.Group("")
		protected.Use(jwtMiddleware.Authenticate())
		{
			// Subscription routes
			subs := protected.Group("/subscription")
			subs.GET("", subscriptionHandler.GetSubscription)
			subs.GET("/access",
				rateLimiter.Middleware(middleware.ByUserID, middleware.PollingConfig),
				subscriptionHandler.CheckAccess,
			)
			subs.DELETE("", subscriptionHandler.CancelSubscription)

			// IAP verification
			// protected.POST("/verify/iap", verifyIAPHandler)
		}
	}

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	go func() {
		logging.Logger.Info("Server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logging.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logging.Logger.Info("Server exited")
}
