package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/infrastructure/cache"
	"github.com/bivex/paywall-iap/internal/infrastructure/config"
	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/pool"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	worker_tasks "github.com/bivex/paywall-iap/internal/worker/tasks"
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

	logging.Logger.Info("Starting IAP Worker server")

	// Initialize database for worker tasks
	ctx := context.Background()
	dbPool, err := pool.NewPool(ctx, cfg.Database)
	if err != nil {
		logging.Logger.Fatal("Failed to create database pool", zap.Error(err))
	}
	defer pool.Close(dbPool)

	queries := generated.New(dbPool)
	taskHandlers := worker_tasks.NewTaskHandlers(queries)

	// Initialize advanced bandit services for worker
	banditRepo := repository.NewPostgresBanditRepository(dbPool, logging.Logger)
	banditCache := cache.NewRedisBanditCache(redisClient, logging.Logger)
	banditService := service.NewThompsonSamplingBandit(banditRepo, banditCache, logging.Logger)
	currencyService := service.NewCurrencyRateService(redisClient, logging.Logger)
	advancedBanditEngine := service.NewAdvancedBanditEngine(
		banditService,
		banditRepo,
		banditCache,
		currencyService,
		logging.Logger,
		&service.EngineConfig{
			EnableCurrency: true,
		},
	)

	currencyJobs := worker_tasks.NewCurrencyJobs(currencyService, logging.Logger)
	banditMaintenanceJobs := worker_tasks.NewBanditMaintenanceJobs(advancedBanditEngine, logging.Logger)

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

	// Initialize Asynq server
	server := asynq.NewServerFromRedisClient(redisClient, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
			// Exponential backoff: 2^n seconds
			return time.Duration(1<<uint(n)) * time.Second
		},
	})

	// Register task handlers
	mux := asynq.NewServeMux()
	worker_tasks.RegisterHandlers(mux, taskHandlers)

	// Register advanced bandit worker handlers
	worker_tasks.RegisterCurrencyTasks(mux, currencyService, logging.Logger)
	worker_tasks.RegisterBanditMaintenanceTasks(mux, advancedBanditEngine, logging.Logger)

	// Start server in background
	if err := server.Start(mux); err != nil {
		logging.Logger.Fatal("Failed to start worker", zap.Error(err))
	}

	// Register scheduled tasks
	scheduler := asynq.NewSchedulerFromRedisClient(redisClient, nil)
	worker_tasks.RegisterScheduledTasks(scheduler)

	// Register advanced bandit scheduled tasks
	worker_tasks.RegisterCurrencyScheduledTasks(scheduler)
	worker_tasks.RegisterBanditMaintenanceScheduledTasks(scheduler)

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		logging.Logger.Fatal("Failed to start scheduler", zap.Error(err))
	}

	logging.Logger.Info("Worker started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Logger.Info("Shutting down worker...")

	scheduler.Shutdown()
	server.Shutdown()

	logging.Logger.Info("Worker exited")
}
