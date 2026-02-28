package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/password9090/paywall-iap/internal/infrastructure/config"
	"github.com/password9090/paywall-iap/internal/infrastructure/logging"
	worker_tasks "github.com/password9090/paywall-iap/internal/worker/tasks"
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

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.URL,
		Password: cfg.Redis.Password,
		DB:       0,
		PoolSize: cfg.Redis.PoolSize,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logging.Logger.Fatal("Failed to ping Redis", zap.Error(err))
	}

	// Initialize Asynq worker
	worker := asynq.NewWorker(
		asynq.Config{
			Concurrency: 10,
			Queues:      []string{"critical", "default", "low"},
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				// Exponential backoff: 2^n seconds
				return time.Duration(1<<uint(n)) * time.Second
			},
		},
		asynq.WithRedisClientOpt(asynq.RedisClientOpt{RedisClient: redisClient}),
	)

	// Register task handlers
	worker_tasks.RegisterHandlers(worker)

	// Start worker in background
	if err := worker.Start(); err != nil {
		logging.Logger.Fatal("Failed to start worker", zap.Error(err))
	}

	// Register scheduled tasks
	scheduler := asynq.NewScheduler(asynq.RedisClientOpt{RedisClient: redisClient})
	worker_tasks.RegisterScheduledTasks(scheduler)

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

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	scheduler.Shutdown()
	worker.Shutdown()

	logging.Logger.Info("Worker exited")
}
