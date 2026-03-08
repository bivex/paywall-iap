package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	openapi "github.com/bivex/paywall-iap/docs/openapi"
	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/application/query"
	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/cache"
	"github.com/bivex/paywall-iap/internal/infrastructure/config"
	iapext "github.com/bivex/paywall-iap/internal/infrastructure/external/iap"
	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/pool"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	app_handler "github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
)

func main() {
	dumpRoutes := flag.Bool("dump-routes", false, "Print registered HTTP routes and exit")
	flag.Parse()
	if *dumpRoutes {
		cfg := dumpRoutesConfig()
		mustInitLogger(&cfg.Sentry)
		defer logging.Sync()

		router := setupRouter(cfg, dumpRoutesDependencies(), nil)
		printRoutes(os.Stdout, router)
		return
	}

	cfg := mustLoadConfig()
	mustInitLogger(&cfg.Sentry)
	defer logging.Sync()

	logging.Logger.Info("Starting IAP API server",
		zap.Int("port", cfg.Server.Port),
		zap.String("environment", cfg.Sentry.Environment),
	)

	ctx := context.Background()
	dbPool := mustInitDB(ctx, cfg.Database)
	defer pool.Close(dbPool)

	opts := mustInitRedis(ctx, cfg.Redis)
	redisClient := redis.NewClient(opts)
	defer redisClient.Close()

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: opts.Addr, Password: opts.Password})
	defer asynqClient.Close()

	deps := initDependencies(cfg, dbPool, redisClient, asynqClient)
	router := setupRouter(cfg, deps, redisClient)
	if *dumpRoutes {
		printRoutes(os.Stdout, router)
		return
	}

	startServer(cfg, router)
}

func dumpRoutesConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:    "dump-routes-secret-dump-routes-secret",
			AccessTTL: 15 * time.Minute,
		},
		Sentry: config.SentryConfig{
			Environment: "dump-routes",
		},
	}
}

func dumpRoutesDependencies() *dependencies {
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	return &dependencies{
		jwtMiddleware: middleware.NewJWTMiddleware("dump-routes-secret-dump-routes-secret", nil, 15*time.Minute),
		rateLimiter:   middleware.NewRateLimiter(redisClient, true),

		authHandler:           (*app_handler.AuthHandler)(nil),
		iapHandler:            (*app_handler.IAPHandler)(nil),
		subscriptionHandler:   (*app_handler.SubscriptionHandler)(nil),
		adminHandler:          (*app_handler.AdminHandler)(nil),
		webhookHandler:        (*app_handler.WebhookHandler)(nil),
		banditHandler:         (*app_handler.BanditHandler)(nil),
		banditAdvancedHandler: (*app_handler.BanditAdvancedHandler)(nil),
	}
}

func printRoutes(w io.Writer, router *gin.Engine) {
	routes := append([]gin.RouteInfo(nil), router.Routes()...)
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "METHOD\tPATH\tHANDLER")
	for _, route := range routes {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", route.Method, route.Path, route.Handler)
	}
	_, _ = fmt.Fprintf(tw, "\nTOTAL\t%d\t\n", len(routes))
	_ = tw.Flush()
}

// mustLoadConfig loads and returns configuration, exiting on failure
func mustLoadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}

// mustInitLogger initializes the logger, exiting on failure
func mustInitLogger(sentryCfg *config.SentryConfig) {
	if err := logging.Init(sentryCfg); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

// mustInitDB creates and tests database connection
func mustInitDB(ctx context.Context, dbCfg config.DatabaseConfig) *pgxpool.Pool {
	dbPool, err := pool.NewPool(ctx, dbCfg)
	if err != nil {
		logging.Logger.Fatal("Failed to create database pool", zap.Error(err))
	}

	if err := pool.Ping(ctx, dbPool); err != nil {
		pool.Close(dbPool)
		logging.Logger.Fatal("Failed to ping database", zap.Error(err))
	}

	return dbPool
}

// mustInitRedis parses Redis URL and tests connection
func mustInitRedis(ctx context.Context, redisCfg config.RedisConfig) *redis.Options {
	opts, err := redis.ParseURL(redisCfg.URL)
	if err != nil {
		logging.Logger.Fatal("Failed to parse Redis URL", zap.Error(err))
	}
	opts.PoolSize = redisCfg.PoolSize

	// Test connection
	redisClient := redis.NewClient(opts)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient.Close()
		logging.Logger.Fatal("Failed to ping Redis", zap.Error(err))
	}
	redisClient.Close()

	return opts
}

// dependencies holds all initialized dependencies
type dependencies struct {
	queries          *generated.Queries
	userRepo         domainRepo.UserRepository
	subscriptionRepo domainRepo.SubscriptionRepository
	transactionRepo  domainRepo.TransactionRepository
	analyticsRepo    domainRepo.AnalyticsRepository
	banditRepo       service.BanditRepository
	adminCredRepo    domainRepo.AdminCredentialRepository

	analyticsService *service.AnalyticsService
	auditService     *service.AuditService
	banditService    *service.ThompsonSamplingBandit
	advancedBandit   *service.AdvancedBanditEngine
	currencyService  *service.CurrencyRateService

	jwtMiddleware *middleware.JWTMiddleware
	rateLimiter   *middleware.RateLimiter

	registerCmd   *command.RegisterCommand
	cancelSubCmd  *command.CancelSubscriptionCommand
	verifyIAPCmd  *command.VerifyIAPCommand
	adminLoginCmd *command.AdminLoginCommand

	getSubQuery      *query.GetSubscriptionQuery
	checkAccessQuery *query.CheckAccessQuery

	authHandler           *app_handler.AuthHandler
	iapHandler            *app_handler.IAPHandler
	subscriptionHandler   *app_handler.SubscriptionHandler
	adminHandler          *app_handler.AdminHandler
	webhookHandler        *app_handler.WebhookHandler
	banditHandler         *app_handler.BanditHandler
	banditAdvancedHandler *app_handler.BanditAdvancedHandler
}

// initDependencies initializes all repositories, services, middleware, and handlers
func initDependencies(cfg *config.Config, dbPool *pgxpool.Pool, redisClient *redis.Client, asynqClient *asynq.Client) *dependencies {
	// Initialize repositories
	queries := generated.New(dbPool)
	userRepo := repository.NewUserRepository(queries)
	subscriptionRepo := repository.NewSubscriptionRepository(queries)
	transactionRepo := repository.NewTransactionRepository(queries)
	analyticsRepo := repository.NewAnalyticsRepository(dbPool)
	banditRepo := repository.NewPostgresBanditRepository(dbPool, logging.Logger)
	adminCredRepo := repository.NewAdminCredentialRepository(queries)

	// Initialize services
	analyticsService := service.NewAnalyticsService(analyticsRepo, subscriptionRepo)
	auditService := service.NewAuditService(dbPool)
	winbackRepo := repository.NewWinbackOfferRepository(dbPool)
	winbackService := service.NewWinbackService(winbackRepo, userRepo, subscriptionRepo)

	// Bandit components
	banditCache := cache.NewRedisBanditCache(redisClient, logging.Logger)
	banditService := service.NewThompsonSamplingBandit(banditRepo, banditCache, logging.Logger)
	currencyService := service.NewCurrencyRateService(redisClient, logging.Logger)

	advancedBanditEngine := service.NewAdvancedBanditEngine(
		banditService,
		banditRepo,
		banditCache,
		redisClient,
		currencyService,
		logging.Logger,
		&service.EngineConfig{
			ExperimentConfig: nil,
			EnableCurrency:   true,
			EnableContextual: true,
			EnableDelayed:    true,
			EnableWindow:     true,
			EnableHybrid:     true,
		},
	)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWT.Secret, redisClient, cfg.JWT.AccessTTL)
	rateLimiter := middleware.NewRateLimiter(redisClient, true)

	// Initialize IAP verifiers
	appleVerifier := iapext.NewAppleVerifier(cfg.IAP.AppleSharedSecret, cfg.IAP.IsProduction, cfg.IAP.AppleMockURL)
	googleVerifier := iapext.NewGoogleVerifier(cfg.IAP.GoogleKeyJSON, cfg.IAP.IsProduction, cfg.IAP.GoogleIAPBaseURL)
	iapAdapter := iapext.NewIAPAdapter(appleVerifier, googleVerifier)

	// Initialize commands
	registerCmd := command.NewRegisterCommand(userRepo, jwtMiddleware)
	cancelSubCmd := command.NewCancelSubscriptionCommand(subscriptionRepo)
	verifyIAPCmd := command.NewVerifyIAPCommand(
		userRepo,
		subscriptionRepo,
		transactionRepo,
		iapext.NewAppleVerifierAdapter(iapAdapter),
		iapext.NewAndroidVerifierAdapter(iapAdapter),
	)
	adminLoginCmd := command.NewAdminLoginCommand(userRepo, adminCredRepo, jwtMiddleware)

	// Initialize queries
	getSubQuery := query.NewGetSubscriptionQuery(subscriptionRepo)
	checkAccessQuery := query.NewCheckAccessQuery(subscriptionRepo)

	// Initialize handlers
	authHandler := app_handler.NewAuthHandler(registerCmd, adminLoginCmd, jwtMiddleware)
	iapHandler := app_handler.NewIAPHandler(verifyIAPCmd, jwtMiddleware, rateLimiter)
	subscriptionHandler := app_handler.NewSubscriptionHandler(getSubQuery, checkAccessQuery, cancelSubCmd, jwtMiddleware)
	adminHandler := app_handler.NewAdminHandler(
		subscriptionRepo,
		userRepo,
		queries,
		dbPool,
		redisClient,
		analyticsService,
		auditService,
		service.NewRevenueOpsService(dbPool),
		service.NewAnalyticsReportService(dbPool),
		service.NewUserProfileService(dbPool),
		winbackService,
		asynqClient,
	)
	webhookHandler := app_handler.NewWebhookHandler(
		cfg.IAP.StripeWebhookSecret,
		cfg.IAP.AppleWebhookSecret,
		cfg.IAP.GoogleWebhookSecret,
		queries,
		asynqClient,
	)
	banditHandler := app_handler.NewBanditHandler(banditService)
	banditAdvancedHandler := app_handler.NewBanditAdvancedHandler(advancedBanditEngine, currencyService, logging.Logger)

	return &dependencies{
		queries:               queries,
		userRepo:              userRepo,
		subscriptionRepo:      subscriptionRepo,
		transactionRepo:       transactionRepo,
		analyticsRepo:         analyticsRepo,
		banditRepo:            banditRepo,
		adminCredRepo:         adminCredRepo,
		analyticsService:      analyticsService,
		auditService:          auditService,
		banditService:         banditService,
		advancedBandit:        advancedBanditEngine,
		currencyService:       currencyService,
		jwtMiddleware:         jwtMiddleware,
		rateLimiter:           rateLimiter,
		registerCmd:           registerCmd,
		cancelSubCmd:          cancelSubCmd,
		verifyIAPCmd:          verifyIAPCmd,
		adminLoginCmd:         adminLoginCmd,
		getSubQuery:           getSubQuery,
		checkAccessQuery:      checkAccessQuery,
		authHandler:           authHandler,
		iapHandler:            iapHandler,
		subscriptionHandler:   subscriptionHandler,
		adminHandler:          adminHandler,
		webhookHandler:        webhookHandler,
		banditHandler:         banditHandler,
		banditAdvancedHandler: banditAdvancedHandler,
	}
}

// setupRouter configures and returns the Gin router with all routes
func setupRouter(cfg *config.Config, d *dependencies, redisClient *redis.Client) *gin.Engine {
	if cfg.Sentry.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(gin.Recovery(), logging.RequestMiddleware(logging.Logger))
	router.GET("/openapi.yaml", openapi.ServeYAML)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Webhooks (no auth)
	webhooks := router.Group("/webhook")
	{
		webhooks.POST("/stripe", d.webhookHandler.StripeWebhook)
		webhooks.POST("/apple", d.webhookHandler.AppleWebhook)
		webhooks.POST("/google", d.webhookHandler.GoogleWebhook)
	}

	// API v1 routes
	v1 := router.Group("/v1")
	{
		setupAuthRoutes(v1, d)
		setupAdminAuthRoutes(v1, d)
		setupBanditRoutes(v1, d)
		setupProtectedRoutes(v1, d)
		setupAdminRoutes(v1, d, cfg)
	}

	return router
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(v1 *gin.RouterGroup, d *dependencies) {
	auth := v1.Group("/auth")
	{
		auth.POST("/register", d.authHandler.Register)
		auth.POST("/refresh",
			d.rateLimiter.Middleware(middleware.ByIP, middleware.DefaultConfig),
			d.authHandler.RefreshToken,
		)
	}
}

// setupAdminAuthRoutes configures admin authentication routes
func setupAdminAuthRoutes(v1 *gin.RouterGroup, d *dependencies) {
	adminAuth := v1.Group("/admin/auth")
	{
		adminAuth.POST("/login", d.authHandler.AdminLogin)
		adminAuth.POST("/logout", d.authHandler.AdminLogout)
	}
}

// setupBanditRoutes configures multi-armed bandit routes
func setupBanditRoutes(v1 *gin.RouterGroup, d *dependencies) {
	bandit := v1.Group("/bandit")
	{
		bandit.POST("/assign", d.banditHandler.Assign)
		bandit.POST("/impression", d.banditHandler.Impression)
		bandit.POST("/reward", d.banditHandler.Reward)
		bandit.GET("/statistics", d.banditHandler.Statistics)
		bandit.GET("/health", d.banditHandler.Health)

		// Advanced bandit routes
		bandit.GET("/currency/rates", gin.WrapF(d.banditAdvancedHandler.GetCurrencyRates))
		bandit.POST("/currency/update", gin.WrapF(d.banditAdvancedHandler.UpdateCurrencyRates))
		bandit.POST("/currency/convert", gin.WrapF(d.banditAdvancedHandler.ConvertCurrency))
		bandit.GET("/experiments/:id/objectives", gin.WrapF(d.banditAdvancedHandler.GetObjectiveScores))
		bandit.GET("/experiments/:id/objectives/config", gin.WrapF(d.banditAdvancedHandler.GetObjectiveConfig))
		bandit.PUT("/experiments/:id/objectives/config", gin.WrapF(d.banditAdvancedHandler.SetObjectiveConfig))
		bandit.GET("/experiments/:id/window/info", gin.WrapF(d.banditAdvancedHandler.GetWindowInfo))
		bandit.POST("/experiments/:id/window/trim", gin.WrapF(d.banditAdvancedHandler.TrimWindow))
		bandit.GET("/experiments/:id/window/events", gin.WrapF(d.banditAdvancedHandler.ExportWindowEvents))
		bandit.POST("/conversions", gin.WrapF(d.banditAdvancedHandler.ProcessConversion))
		bandit.GET("/pending/:id", gin.WrapF(d.banditAdvancedHandler.GetPendingReward))
		bandit.GET("/users/:id/pending", gin.WrapF(d.banditAdvancedHandler.GetUserPendingRewards))
		bandit.GET("/experiments/:id/metrics", gin.WrapF(d.banditAdvancedHandler.GetMetrics))
		bandit.POST("/maintenance", gin.WrapF(d.banditAdvancedHandler.RunMaintenance))
	}
}

// setupProtectedRoutes configures JWT-protected routes
func setupProtectedRoutes(v1 *gin.RouterGroup, d *dependencies) {
	protected := v1.Group("")
	protected.Use(d.jwtMiddleware.Authenticate())
	{
		protected.POST("/verify/iap", d.iapHandler.VerifyReceipt)

		subs := protected.Group("/subscription")
		{
			subs.GET("", d.subscriptionHandler.GetSubscription)
			subs.GET("/access",
				d.rateLimiter.Middleware(middleware.ByUserID, middleware.PollingConfig),
				d.subscriptionHandler.CheckAccess,
			)
			subs.DELETE("", d.subscriptionHandler.CancelSubscription)
		}
	}
}

// setupAdminRoutes configures admin routes
func setupAdminRoutes(v1 *gin.RouterGroup, d *dependencies, cfg *config.Config) {
	admin := v1.Group("/admin")
	admin.Use(d.jwtMiddleware.Authenticate())
	admin.Use(middleware.AdminMiddleware(d.userRepo, cfg.JWT.Secret))
	{
		admin.POST("/users/:id/grant", d.adminHandler.GrantSubscription)
		admin.POST("/users/:id/revoke", d.adminHandler.RevokeSubscription)
		admin.POST("/users/:id/force-cancel", d.adminHandler.ForceCancel)
		admin.POST("/users/:id/force-renew", d.adminHandler.ForceRenew)
		admin.POST("/users/:id/grant-grace", d.adminHandler.GrantGracePeriod)
		admin.GET("/users", d.adminHandler.ListUsers)
		admin.GET("/users/search", d.adminHandler.SearchUsers)
		admin.GET("/users/:id/profile", d.adminHandler.GetUserProfile)
		admin.GET("/dashboard/metrics", d.adminHandler.GetDashboardMetrics)
		admin.GET("/audit-log", d.adminHandler.GetAuditLog)
		admin.GET("/subscriptions", d.adminHandler.ListSubscriptions)
		admin.GET("/subscriptions/:id", d.adminHandler.GetSubscriptionDetail)
		admin.GET("/transactions", d.adminHandler.ListTransactions)
		admin.GET("/transactions/:id", d.adminHandler.GetTransactionDetail)
		admin.GET("/webhooks", d.adminHandler.ListWebhooks)
		admin.GET("/analytics/report", d.adminHandler.GetAnalyticsReport)
		admin.GET("/revenue-ops", d.adminHandler.GetRevenueOps)
		admin.GET("/experiments", d.adminHandler.ListAdminExperiments)
		admin.POST("/experiments", d.adminHandler.CreateAdminExperiment)
		admin.PUT("/experiments/:id", d.adminHandler.UpdateAdminExperiment)
		admin.GET("/experiments/:id/lifecycle-audit", d.adminHandler.GetAdminExperimentLifecycleAuditHistory)
		admin.POST("/experiments/:id/pause", d.adminHandler.PauseAdminExperiment)
		admin.POST("/experiments/:id/resume", d.adminHandler.ResumeAdminExperiment)
		admin.POST("/experiments/:id/complete", d.adminHandler.CompleteAdminExperiment)
		admin.POST("/experiments/:id/lock", d.adminHandler.LockAdminExperiment)
		admin.POST("/experiments/:id/unlock", d.adminHandler.UnlockAdminExperiment)
		admin.POST("/experiments/:id/repair", d.adminHandler.RepairAdminExperiment)
		admin.GET("/pricing-tiers", d.adminHandler.ListPricingTiers)
		admin.POST("/pricing-tiers", d.adminHandler.CreatePricingTier)
		admin.PUT("/pricing-tiers/:id", d.adminHandler.UpdatePricingTier)
		admin.POST("/pricing-tiers/:id/activate", d.adminHandler.ActivatePricingTier)
		admin.POST("/pricing-tiers/:id/deactivate", d.adminHandler.DeactivatePricingTier)
		admin.GET("/winback-campaigns", d.adminHandler.ListWinbackCampaigns)
		admin.POST("/winback-campaigns", d.adminHandler.LaunchWinbackCampaign)
		admin.POST("/winback-campaigns/:campaignId/deactivate", d.adminHandler.DeactivateWinbackCampaign)
		admin.GET("/settings", d.adminHandler.GetPlatformSettings)
		admin.PUT("/settings", d.adminHandler.UpdatePlatformSettings)
		admin.POST("/settings/password", d.adminHandler.ChangeAdminPassword)
		admin.POST("/webhooks/:id/replay", d.adminHandler.ReplayWebhook)
		admin.GET("/health", d.adminHandler.GetHealth)
	}
}

// startServer starts the HTTP server with graceful shutdown
func startServer(cfg *config.Config, router *gin.Engine) {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in background
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
