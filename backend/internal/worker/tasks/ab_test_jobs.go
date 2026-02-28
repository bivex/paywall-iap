package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

const (
	TypeCalculateABTestStats = "ab_test:calculate_stats"
)

// CalculateABTestStatsPayload is the payload for calculating A/B test statistics
type CalculateABTestStatsPayload struct {
	FlagID string `json:"flag_id"`
}

// ABTestJobHandler handles A/B test background jobs
type ABTestJobHandler struct {
	abAnalyticsService *service.ABAnalyticsService
}

// NewABTestJobHandler creates a new A/B test job handler
func NewABTestJobHandler(abAnalyticsService *service.ABAnalyticsService) *ABTestJobHandler {
	return &ABTestJobHandler{
		abAnalyticsService: abAnalyticsService,
	}
}

// HandleCalculateABTestStats calculates statistics for an A/B test
func (h *ABTestJobHandler) HandleCalculateABTestStats(ctx context.Context, t *asynq.Task) error {
	var p CalculateABTestStatsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	stats := h.abAnalyticsService.GetVariantStats(p.FlagID)

	// Log stats (in production, store in database or send to analytics platform)
	for variant, metrics := range stats {
		fmt.Printf("Flag: %s, Variant: %s\n", p.FlagID, variant)
		fmt.Printf("  Conversion Rate: %.2f%%\n", metrics["conversion_rate"])
		fmt.Printf("  Avg Revenue: $%.2f\n", metrics["avg_revenue"])
	}

	return nil
}

// ScheduleABTestJobs schedules recurring A/B test jobs
func ScheduleABTestJobs(scheduler *asynq.Scheduler) error {
	// Calculate A/B test stats daily
	payload, _ := json.Marshal(CalculateABTestStatsPayload{FlagID: "paywall_variant_test"})
	_, err := scheduler.Register("0 2 * * *", asynq.NewTask(TypeCalculateABTestStats, payload))
	if err != nil {
		return err
	}

	return nil
}
