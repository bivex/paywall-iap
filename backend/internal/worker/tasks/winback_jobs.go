package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/service"
)

const (
	TypeProcessExpiredWinbackOffers = "winback:process_expired"
	TypeCreateWinbackCampaign       = "winback:create_campaign"
)

// ProcessExpiredWinbackOffersPayload is the payload for processing expired offers
type ProcessExpiredWinbackOffersPayload struct {
	Limit int `json:"limit"`
}

// CreateWinbackCampaignPayload is the payload for creating winback campaign
type CreateWinbackCampaignPayload struct {
	CampaignID     string  `json:"campaign_id"`
	DiscountType   string  `json:"discount_type"`
	DiscountValue  float64 `json:"discount_value"`
	DurationDays   int     `json:"duration_days"`
	DaysSinceChurn int     `json:"days_since_churn"`
}

// WinbackJobHandler handles winback background jobs
type WinbackJobHandler struct {
	winbackService      *service.WinbackService
	notificationService *service.NotificationService
}

// NewWinbackJobHandler creates a new winback job handler
func NewWinbackJobHandler(
	winbackService *service.WinbackService,
	notificationService *service.NotificationService,
) *WinbackJobHandler {
	return &WinbackJobHandler{
		winbackService:      winbackService,
		notificationService: notificationService,
	}
}

// HandleProcessExpiredWinbackOffers processes expired winback offers
func (h *WinbackJobHandler) HandleProcessExpiredWinbackOffers(ctx context.Context, t *asynq.Task) error {
	var p ProcessExpiredWinbackOffersPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	limit := p.Limit
	if limit == 0 {
		limit = 100
	}

	processed, err := h.winbackService.ProcessExpiredWinbackOffers(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to process expired winback offers: %w", err)
	}

	if processed > 0 {
		fmt.Printf("Processed %d expired winback offers\n", processed)
	}
	return nil
}

// HandleCreateWinbackCampaign creates winback offers for churned users
func (h *WinbackJobHandler) HandleCreateWinbackCampaign(ctx context.Context, t *asynq.Task) error {
	var p CreateWinbackCampaignPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	discountType := entity.DiscountType(p.DiscountType)
	if discountType == "" {
		discountType = entity.DiscountTypePercentage
	}

	created, err := h.winbackService.CreateWinbackCampaignForChurnedUsers(
		ctx,
		p.CampaignID,
		discountType,
		p.DiscountValue,
		p.DurationDays,
		p.DaysSinceChurn,
	)
	if err != nil {
		return fmt.Errorf("failed to create winback campaign: %w", err)
	}

	if created > 0 {
		fmt.Printf("Created %d winback offers for campaign %s\n", created, p.CampaignID)
	}
	return nil
}

// ScheduleWinbackJobs schedules recurring winback jobs
func ScheduleWinbackJobs(scheduler *asynq.Scheduler) error {
	// Process expired winback offers every hour
	expiredPayload, _ := json.Marshal(ProcessExpiredWinbackOffersPayload{Limit: 100})
	_, err := scheduler.Register("0 * * * *", asynq.NewTask(TypeProcessExpiredWinbackOffers, expiredPayload))
	if err != nil {
		return err
	}

	// Create winback campaign weekly for users churned in last 7 days
	campaignPayload, _ := json.Marshal(CreateWinbackCampaignPayload{
		CampaignID:     "weekly_winback_" + time.Now().Format("20060102"),
		DiscountType:   string(entity.DiscountTypePercentage),
		DiscountValue:  30.0,
		DurationDays:   30,
		DaysSinceChurn: 7,
	})
	_, err = scheduler.Register("0 9 * * 1", asynq.NewTask(TypeCreateWinbackCampaign, campaignPayload))
	if err != nil {
		return err
	}

	return nil
}
