package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

const (
	TypeAggregateDailyMetrics = "analytics:aggregate_daily"
)

// AnalyticsJobHandler handles analytics background jobs
type AnalyticsJobHandler struct {
	analyticsService *service.AnalyticsService
}

// NewAnalyticsJobHandler creates a new analytics job handler
func NewAnalyticsJobHandler(analyticsService *service.AnalyticsService) *AnalyticsJobHandler {
	return &AnalyticsJobHandler{
		analyticsService: analyticsService,
	}
}

// HandleAggregateDailyMetrics computes and stores daily metrics
func (h *AnalyticsJobHandler) HandleAggregateDailyMetrics(ctx context.Context, t *asynq.Task) error {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	revenue, err := h.analyticsService.CalculateRevenueMetrics(ctx, start, end)
	if err != nil {
		return err
	}

	churn, err := h.analyticsService.CalculateChurnMetrics(ctx, start, end)
	if err != nil {
		return err
	}

	// In a real app, we'd store these in an 'analytics_snapshots' table
	// For now, we'll log them as "computed"
	fmt.Printf("[ANALYTICS] Computed daily metrics for %s: Revenue=%.2f, MRR=%.2f, ChurnRate=%.2f%%\n",
		yesterday.Format("2006-01-02"), revenue.DailyRevenue, revenue.MRR, churn.ChurnRate)

	return nil
}

// ScheduleAnalyticsJobs schedules recurring analytics jobs
func ScheduleAnalyticsJobs(scheduler *asynq.Scheduler) error {
	// Run daily aggregation at 1 AM
	_, err := scheduler.Register("0 1 * * *", asynq.NewTask(TypeAggregateDailyMetrics, nil))
	if err != nil {
		return err
	}

	return nil
}
