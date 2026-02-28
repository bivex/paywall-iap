package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/application/query"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/config"
	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/pool"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	app_handler "github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
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
	opts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		logging.Logger.Fatal("Failed to parse Redis URL", zap.Error(err))
	}
	opts.PoolSize = cfg.Redis.PoolSize
	redisClient := redis.NewClient(opts)
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logging.Logger.Fatal("Failed to ping Redis", zap.Error(err))
	}

	// Initialize repositories
	queries := generated.New(dbPool)
	userRepo := repository.NewUserRepository(queries)
	subscriptionRepo := repository.NewSubscriptionRepository(queries)
	analyticsRepo := repository.NewAnalyticsRepository(dbPool)

	// Initialize services
	analyticsService := service.NewAnalyticsService(analyticsRepo, subscriptionRepo)
	auditService := service.NewAuditService(dbPool)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(
		cfg.JWT.Secret,
		redisClient,
		cfg.JWT.AccessTTL,
	)
	rateLimiter := middleware.NewRateLimiter(redisClient, true) // fail open

	// Initialize commands
	registerCmd := command.NewRegisterCommand(userRepo, jwtMiddleware)
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
	adminHandler := app_handler.NewAdminHandler(
		subscriptionRepo,
		userRepo,
		queries,
		dbPool,
		redisClient,
		analyticsService,
		auditService,
	)
	webhookHandler := app_handler.NewWebhookHandler(
		cfg.IAP.StripeWebhookSecret,
		cfg.IAP.AppleWebhookSecret,
		cfg.IAP.GoogleWebhookSecret,
		queries,
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

	// Webhook routes (no auth — verified by signature)
	webhooks := router.Group("/webhook")
	{
		webhooks.POST("/stripe", webhookHandler.StripeWebhook)
		webhooks.POST("/apple", webhookHandler.AppleWebhook)
		webhooks.POST("/google", webhookHandler.GoogleWebhook)
	}

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
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(jwtMiddleware.Authenticate())
		admin.Use(middleware.AdminMiddleware(userRepo, cfg.JWT.Secret))
		{
			admin.POST("/users/:id/grant", adminHandler.GrantSubscription)
			admin.POST("/users/:id/revoke", adminHandler.RevokeSubscription)
			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/dashboard/metrics", adminHandler.GetDashboardMetrics)
			admin.GET("/health", adminHandler.GetHealth)
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
