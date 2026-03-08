package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

type WinbackCampaignSummary struct {
	CampaignID     string    `json:"campaign_id"`
	DiscountType   string    `json:"discount_type"`
	DiscountValue  float64   `json:"discount_value"`
	TotalOffers    int       `json:"total_offers"`
	ActiveOffers   int       `json:"active_offers"`
	AcceptedOffers int       `json:"accepted_offers"`
	ExpiredOffers  int       `json:"expired_offers"`
	DeclinedOffers int       `json:"declined_offers"`
	LaunchedAt     time.Time `json:"launched_at"`
	LatestExpiryAt time.Time `json:"latest_expiry_at"`
}

type launchWinbackCampaignRequest struct {
	CampaignID     string  `json:"campaign_id"`
	DiscountType   string  `json:"discount_type"`
	DiscountValue  float64 `json:"discount_value"`
	DurationDays   int     `json:"duration_days"`
	DaysSinceChurn int     `json:"days_since_churn"`
}

func normalizeLaunchWinbackCampaignRequest(req launchWinbackCampaignRequest) launchWinbackCampaignRequest {
	req.CampaignID = strings.TrimSpace(req.CampaignID)
	req.DiscountType = strings.ToLower(strings.TrimSpace(req.DiscountType))
	return req
}

func validateLaunchWinbackCampaignRequest(req launchWinbackCampaignRequest) string {
	if req.CampaignID == "" {
		return "Campaign ID is required"
	}
	if req.DiscountType != string(entity.DiscountTypePercentage) && req.DiscountType != string(entity.DiscountTypeFixed) {
		return "Discount type must be percentage or fixed"
	}
	if req.DiscountValue <= 0 {
		return "Discount value must be greater than zero"
	}
	if req.DiscountType == string(entity.DiscountTypePercentage) && req.DiscountValue > 100 {
		return "Percentage discount cannot exceed 100"
	}
	if req.DurationDays <= 0 {
		return "Duration days must be greater than zero"
	}
	if req.DaysSinceChurn <= 0 {
		return "Days since churn must be greater than zero"
	}
	return ""
}

func scanWinbackCampaignSummary(scanner interface{ Scan(dest ...any) error }) (WinbackCampaignSummary, error) {
	var summary WinbackCampaignSummary
	err := scanner.Scan(
		&summary.CampaignID,
		&summary.DiscountType,
		&summary.DiscountValue,
		&summary.TotalOffers,
		&summary.ActiveOffers,
		&summary.AcceptedOffers,
		&summary.ExpiredOffers,
		&summary.DeclinedOffers,
		&summary.LaunchedAt,
		&summary.LatestExpiryAt,
	)
	return summary, err
}

func (h *AdminHandler) logWinbackCampaignAction(c *gin.Context, action string, summary WinbackCampaignSummary, details map[string]interface{}) {
	if h.auditService == nil {
		return
	}
	adminIDValue, ok := c.Get("admin_id")
	if !ok {
		return
	}
	adminID, ok := adminIDValue.(uuid.UUID)
	if !ok {
		return
	}

	if details == nil {
		details = map[string]interface{}{}
	}
	details["campaign_id"] = summary.CampaignID
	details["discount_type"] = summary.DiscountType
	details["discount_value"] = summary.DiscountValue
	details["total_offers"] = summary.TotalOffers
	details["active_offers"] = summary.ActiveOffers

	_ = h.auditService.LogAction(c.Request.Context(), adminID, action, "winback_campaign", nil, details)
}

func (h *AdminHandler) getWinbackCampaignSummary(ctx *gin.Context, campaignID string) (WinbackCampaignSummary, error) {
	return scanWinbackCampaignSummary(h.dbPool.QueryRow(ctx.Request.Context(), `
		SELECT campaign_id,
		       MIN(discount_type)::text AS discount_type,
		       MIN(discount_value)::double precision AS discount_value,
		       COUNT(*)::int AS total_offers,
		       COUNT(*) FILTER (WHERE status = 'offered' AND expires_at > NOW())::int AS active_offers,
		       COUNT(*) FILTER (WHERE status = 'accepted')::int AS accepted_offers,
		       COUNT(*) FILTER (WHERE status = 'expired' OR (status = 'offered' AND expires_at <= NOW()))::int AS expired_offers,
		       COUNT(*) FILTER (WHERE status = 'declined')::int AS declined_offers,
		       MIN(offered_at) AS launched_at,
		       MAX(expires_at) AS latest_expiry_at
		FROM winback_offers
		WHERE campaign_id = $1
		GROUP BY campaign_id`, campaignID))
}

func (h *AdminHandler) ListWinbackCampaigns(c *gin.Context) {
	rows, err := h.dbPool.Query(c.Request.Context(), `
		SELECT campaign_id,
		       MIN(discount_type)::text AS discount_type,
		       MIN(discount_value)::double precision AS discount_value,
		       COUNT(*)::int AS total_offers,
		       COUNT(*) FILTER (WHERE status = 'offered' AND expires_at > NOW())::int AS active_offers,
		       COUNT(*) FILTER (WHERE status = 'accepted')::int AS accepted_offers,
		       COUNT(*) FILTER (WHERE status = 'expired' OR (status = 'offered' AND expires_at <= NOW()))::int AS expired_offers,
		       COUNT(*) FILTER (WHERE status = 'declined')::int AS declined_offers,
		       MIN(offered_at) AS launched_at,
		       MAX(expires_at) AS latest_expiry_at
		FROM winback_offers
		GROUP BY campaign_id
		ORDER BY MIN(offered_at) DESC`)
	if err != nil {
		response.InternalError(c, "Failed to load winback campaigns")
		return
	}
	defer rows.Close()

	campaigns := make([]WinbackCampaignSummary, 0)
	for rows.Next() {
		summary, err := scanWinbackCampaignSummary(rows)
		if err != nil {
			response.InternalError(c, "Failed to load winback campaigns")
			return
		}
		campaigns = append(campaigns, summary)
	}
	if rows.Err() != nil {
		response.InternalError(c, "Failed to load winback campaigns")
		return
	}

	response.OK(c, campaigns)
}

func (h *AdminHandler) LaunchWinbackCampaign(c *gin.Context) {
	if h.winbackService == nil {
		response.ServiceUnavailable(c, "Winback service is not configured")
		return
	}

	var req launchWinbackCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid winback campaign payload")
		return
	}
	req = normalizeLaunchWinbackCampaignRequest(req)
	if msg := validateLaunchWinbackCampaignRequest(req); msg != "" {
		response.UnprocessableEntity(c, msg)
		return
	}

	var existingCampaignID string
	err := h.dbPool.QueryRow(c.Request.Context(), `
		SELECT campaign_id
		FROM winback_offers
		WHERE campaign_id = $1
		LIMIT 1`, req.CampaignID).Scan(&existingCampaignID)
	if err == nil {
		response.Conflict(c, "Winback campaign with this ID already exists")
		return
	}
	if err != nil && err != pgx.ErrNoRows {
		response.InternalError(c, "Failed to validate winback campaign ID")
		return
	}

	createdOffers, err := h.winbackService.CreateWinbackCampaignForChurnedUsers(
		c.Request.Context(),
		req.CampaignID,
		entity.DiscountType(req.DiscountType),
		req.DiscountValue,
		req.DurationDays,
		req.DaysSinceChurn,
	)
	if err != nil {
		response.InternalError(c, "Failed to launch winback campaign")
		return
	}
	if createdOffers == 0 {
		response.UnprocessableEntity(c, "No recently cancelled users matched the selected window")
		return
	}

	summary, err := h.getWinbackCampaignSummary(c, req.CampaignID)
	if err != nil {
		response.InternalError(c, "Failed to load launched winback campaign")
		return
	}

	h.logWinbackCampaignAction(c, "launch_winback_campaign", summary, map[string]interface{}{
		"days_since_churn": req.DaysSinceChurn,
		"duration_days":    req.DurationDays,
		"created_offers":   createdOffers,
	})
	response.OK(c, summary)
}

func (h *AdminHandler) DeactivateWinbackCampaign(c *gin.Context) {
	if h.winbackService == nil {
		response.ServiceUnavailable(c, "Winback service is not configured")
		return
	}

	campaignID := strings.TrimSpace(c.Param("campaignId"))
	if campaignID == "" {
		response.BadRequest(c, "Campaign ID is required")
		return
	}

	currentSummary, err := h.getWinbackCampaignSummary(c, campaignID)
	if err == pgx.ErrNoRows {
		response.NotFound(c, "Winback campaign not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to load winback campaign")
		return
	}
	if currentSummary.ActiveOffers == 0 {
		response.UnprocessableEntity(c, "Winback campaign has no active offers to deactivate")
		return
	}

	deactivatedOffers, err := h.winbackService.DeactivateCampaign(c.Request.Context(), campaignID)
	if err != nil {
		response.InternalError(c, "Failed to deactivate winback campaign")
		return
	}

	updatedSummary, err := h.getWinbackCampaignSummary(c, campaignID)
	if err != nil {
		response.InternalError(c, "Failed to load updated winback campaign")
		return
	}

	h.logWinbackCampaignAction(c, "deactivate_winback_campaign", updatedSummary, map[string]interface{}{
		"deactivated_offers":     deactivatedOffers,
		"previous_active_offers": currentSummary.ActiveOffers,
	})
	response.OK(c, updatedSummary)
}
