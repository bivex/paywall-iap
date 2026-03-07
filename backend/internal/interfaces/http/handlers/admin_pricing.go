package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

type PricingTier struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	MonthlyPrice *float64  `json:"monthly_price"`
	AnnualPrice  *float64  `json:"annual_price"`
	Currency     string    `json:"currency"`
	Features     []string  `json:"features"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type pricingTierUpsertRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	MonthlyPrice *float64 `json:"monthly_price"`
	AnnualPrice  *float64 `json:"annual_price"`
	Currency     string   `json:"currency"`
	Features     []string `json:"features"`
	IsActive     bool     `json:"is_active"`
}

type pricingTierScanner interface {
	Scan(dest ...any) error
}

func normalizePricingTierRequest(req pricingTierUpsertRequest) pricingTierUpsertRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))

	features := make([]string, 0, len(req.Features))
	for _, feature := range req.Features {
		trimmed := strings.TrimSpace(feature)
		if trimmed != "" {
			features = append(features, trimmed)
		}
	}
	req.Features = features

	return req
}

func validatePricingTierRequest(req pricingTierUpsertRequest) string {
	if req.Name == "" {
		return "Pricing tier name is required"
	}
	if req.Currency == "" || len(req.Currency) != 3 {
		return "Currency must be a 3-letter ISO code"
	}
	for _, r := range req.Currency {
		if r < 'A' || r > 'Z' {
			return "Currency must be a 3-letter ISO code"
		}
	}
	if req.MonthlyPrice == nil && req.AnnualPrice == nil {
		return "At least one price is required"
	}
	if req.MonthlyPrice != nil && *req.MonthlyPrice <= 0 {
		return "Monthly price must be greater than zero"
	}
	if req.AnnualPrice != nil && *req.AnnualPrice <= 0 {
		return "Annual price must be greater than zero"
	}
	return ""
}

func scanPricingTier(scanner pricingTierScanner) (PricingTier, error) {
	var (
		id          uuid.UUID
		name        string
		description string
		monthly     sql.NullFloat64
		annual      sql.NullFloat64
		currency    string
		featuresRaw []byte
		isActive    bool
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := scanner.Scan(
		&id,
		&name,
		&description,
		&monthly,
		&annual,
		&currency,
		&featuresRaw,
		&isActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return PricingTier{}, err
	}

	features := []string{}
	if len(featuresRaw) > 0 {
		if err := json.Unmarshal(featuresRaw, &features); err != nil {
			return PricingTier{}, err
		}
	}

	tier := PricingTier{
		ID:          id.String(),
		Name:        name,
		Description: description,
		Currency:    strings.ToUpper(currency),
		Features:    features,
		IsActive:    isActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	if monthly.Valid {
		value := monthly.Float64
		tier.MonthlyPrice = &value
	}
	if annual.Valid {
		value := annual.Float64
		tier.AnnualPrice = &value
	}

	return tier, nil
}

func pricingTierConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func pricingTierTargetID(id string) *uuid.UUID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	return &parsed
}

func (h *AdminHandler) logPricingTierAction(c *gin.Context, action string, tier PricingTier) {
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

	details := map[string]interface{}{
		"name":      tier.Name,
		"currency":  tier.Currency,
		"is_active": tier.IsActive,
		"features":  tier.Features,
	}
	if tier.MonthlyPrice != nil {
		details["monthly_price"] = *tier.MonthlyPrice
	}
	if tier.AnnualPrice != nil {
		details["annual_price"] = *tier.AnnualPrice
	}

	_ = h.auditService.LogAction(c.Request.Context(), adminID, action, "pricing_tier", pricingTierTargetID(tier.ID), details)
}

func (h *AdminHandler) ListPricingTiers(c *gin.Context) {
	rows, err := h.dbPool.Query(c.Request.Context(), `
		SELECT id,
		       name,
		       COALESCE(description, ''),
		       monthly_price::double precision,
		       annual_price::double precision,
		       currency,
		       COALESCE(features, '[]'::jsonb),
		       is_active,
		       created_at,
		       updated_at
		FROM pricing_tiers
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`)
	if err != nil {
		response.InternalError(c, "Failed to load pricing tiers")
		return
	}
	defer rows.Close()

	tiers := make([]PricingTier, 0)
	for rows.Next() {
		tier, err := scanPricingTier(rows)
		if err != nil {
			response.InternalError(c, "Failed to load pricing tiers")
			return
		}
		tiers = append(tiers, tier)
	}
	if rows.Err() != nil {
		response.InternalError(c, "Failed to load pricing tiers")
		return
	}

	response.OK(c, tiers)
}

func (h *AdminHandler) CreatePricingTier(c *gin.Context) {
	var req pricingTierUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid pricing tier payload")
		return
	}
	req = normalizePricingTierRequest(req)
	if msg := validatePricingTierRequest(req); msg != "" {
		response.UnprocessableEntity(c, msg)
		return
	}

	featuresJSON, err := json.Marshal(req.Features)
	if err != nil {
		response.InternalError(c, "Failed to encode pricing tier features")
		return
	}

	tier, err := scanPricingTier(h.dbPool.QueryRow(c.Request.Context(), `
		INSERT INTO pricing_tiers (
			name,
			description,
			monthly_price,
			annual_price,
			currency,
			features,
			is_active,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, now())
		RETURNING id,
		          name,
		          COALESCE(description, ''),
		          monthly_price::double precision,
		          annual_price::double precision,
		          currency,
		          COALESCE(features, '[]'::jsonb),
		          is_active,
		          created_at,
		          updated_at`,
		req.Name,
		req.Description,
		req.MonthlyPrice,
		req.AnnualPrice,
		req.Currency,
		featuresJSON,
		req.IsActive,
	))
	if err != nil {
		if pricingTierConflict(err) {
			response.Conflict(c, "Pricing tier with this name already exists")
			return
		}
		response.InternalError(c, "Failed to create pricing tier")
		return
	}

	h.logPricingTierAction(c, "create_pricing_tier", tier)
	response.Created(c, tier)
}

func (h *AdminHandler) UpdatePricingTier(c *gin.Context) {
	tierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid pricing tier ID")
		return
	}

	var req pricingTierUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid pricing tier payload")
		return
	}
	req = normalizePricingTierRequest(req)
	if msg := validatePricingTierRequest(req); msg != "" {
		response.UnprocessableEntity(c, msg)
		return
	}

	featuresJSON, err := json.Marshal(req.Features)
	if err != nil {
		response.InternalError(c, "Failed to encode pricing tier features")
		return
	}

	tier, err := scanPricingTier(h.dbPool.QueryRow(c.Request.Context(), `
		UPDATE pricing_tiers
		SET name = $2,
		    description = $3,
		    monthly_price = $4,
		    annual_price = $5,
		    currency = $6,
		    features = $7::jsonb,
		    is_active = $8,
		    updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id,
		          name,
		          COALESCE(description, ''),
		          monthly_price::double precision,
		          annual_price::double precision,
		          currency,
		          COALESCE(features, '[]'::jsonb),
		          is_active,
		          created_at,
		          updated_at`,
		tierID,
		req.Name,
		req.Description,
		req.MonthlyPrice,
		req.AnnualPrice,
		req.Currency,
		featuresJSON,
		req.IsActive,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "Pricing tier not found")
			return
		}
		if pricingTierConflict(err) {
			response.Conflict(c, "Pricing tier with this name already exists")
			return
		}
		response.InternalError(c, "Failed to update pricing tier")
		return
	}

	h.logPricingTierAction(c, "update_pricing_tier", tier)
	response.OK(c, tier)
}

func (h *AdminHandler) ActivatePricingTier(c *gin.Context) {
	h.setPricingTierActive(c, true)
}

func (h *AdminHandler) DeactivatePricingTier(c *gin.Context) {
	h.setPricingTierActive(c, false)
}

func (h *AdminHandler) setPricingTierActive(c *gin.Context, isActive bool) {
	tierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid pricing tier ID")
		return
	}

	tier, err := scanPricingTier(h.dbPool.QueryRow(c.Request.Context(), `
		UPDATE pricing_tiers
		SET is_active = $2,
		    updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id,
		          name,
		          COALESCE(description, ''),
		          monthly_price::double precision,
		          annual_price::double precision,
		          currency,
		          COALESCE(features, '[]'::jsonb),
		          is_active,
		          created_at,
		          updated_at`,
		tierID,
		isActive,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "Pricing tier not found")
			return
		}
		response.InternalError(c, "Failed to update pricing tier status")
		return
	}

	action := "activate_pricing_tier"
	if !isActive {
		action = "deactivate_pricing_tier"
	}
	h.logPricingTierAction(c, action, tier)
	response.OK(c, tier)
}