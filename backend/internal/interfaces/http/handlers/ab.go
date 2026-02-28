package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// ABTestHandler handles A/B testing endpoints
type ABTestHandler struct {
	featureFlagService *service.FeatureFlagService
}

// NewABTestHandler creates a new A/B test handler
func NewABTestHandler(featureFlagService *service.FeatureFlagService) *ABTestHandler {
	return &ABTestHandler{
		featureFlagService: featureFlagService,
	}
}

// GetFeatureFlags returns all feature flags
// @Summary Get all feature flags
// @Tags ab-test
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=[]dto.FeatureFlagResponse}
// @Router /ab-test/flags [get]
func (h *ABTestHandler) GetFeatureFlags(c *gin.Context) {
	flags := h.featureFlagService.GetAllFlags()

	resp := make([]dto.FeatureFlagResponse, len(flags))
	for i, flag := range flags {
		resp[i] = dto.FeatureFlagResponse{
			ID:             flag.ID,
			Name:           flag.Name,
			Enabled:        flag.Enabled,
			RolloutPercent: flag.RolloutPercent,
			UserIDs:        flag.UserIDs,
		}
	}

	response.OK(c, resp)
}

// EvaluateFlag evaluates a feature flag for the current user
// @Summary Evaluate feature flag
// @Tags ab-test
// @Produce json
// @Security Bearer
// @Param flag_id path string true "Feature flag ID"
// @Success 200 {object} response.SuccessResponse{data=dto.ABTestEvaluationResponse}
// @Failure 404 {object} response.ErrorResponse
// @Router /ab-test/evaluate/{flag_id} [get]
func (h *ABTestHandler) EvaluateFlag(c *gin.Context) {
	flagID := c.Param("flag_id")
	userID := c.GetString("user_id")

	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	enabled, err := h.featureFlagService.IsFeatureEnabled(c.Request.Context(), flagID, userID)
	if err != nil {
		response.NotFound(c, "Feature flag not found")
		return
	}

	resp := dto.ABTestEvaluationResponse{
		FlagID:    flagID,
		UserID:    userID,
		IsEnabled: enabled,
	}

	response.OK(c, resp)
}

// EvaluatePaywall returns the paywall variant for the current user
// @Summary Evaluate paywall variant
// @Tags ab-test
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse{data=dto.PaywallVariantResponse}
// @Router /ab-test/paywall [get]
func (h *ABTestHandler) EvaluatePaywall(c *gin.Context) {
	userID := c.GetString("user_id")

	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	variant, err := h.featureFlagService.EvaluatePaywallTest(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to evaluate paywall test")
		return
	}

	resp := dto.PaywallVariantResponse{
		UserID:  userID,
		Variant: variant,
	}

	response.OK(c, resp)
}

// CreateFlag creates a new feature flag (admin only)
// @Summary Create feature flag
// @Tags ab-test
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateFeatureFlagRequest true "Create feature flag request"
// @Success 201 {object} response.SuccessResponse{data=dto.FeatureFlagResponse}
// @Failure 400 {object} response.ErrorResponse
// @Router /ab-test/flags [post]
func (h *ABTestHandler) CreateFlag(c *gin.Context) {
	var req dto.CreateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	flag := h.featureFlagService.CreateFlag(req.ID, req.Name, req.Enabled, req.RolloutPercent, req.UserIDs)

	resp := dto.FeatureFlagResponse{
		ID:             flag.ID,
		Name:           flag.Name,
		Enabled:        flag.Enabled,
		RolloutPercent: flag.RolloutPercent,
		UserIDs:        flag.UserIDs,
	}

	response.Created(c, resp)
}

// UpdateFlag updates an existing feature flag (admin only)
// @Summary Update feature flag
// @Tags ab-test
// @Accept json
// @Produce json
// @Security Bearer
// @Param flag_id path string true "Feature flag ID"
// @Param request body dto.UpdateFeatureFlagRequest true "Update feature flag request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /ab-test/flags/{flag_id} [put]
func (h *ABTestHandler) UpdateFlag(c *gin.Context) {
	flagID := c.Param("flag_id")

	var req dto.UpdateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	err := h.featureFlagService.UpdateFlag(flagID, req.Enabled, req.RolloutPercent, req.UserIDs)
	if err != nil {
		response.NotFound(c, "Feature flag not found")
		return
	}

	response.OK(c, map[string]string{"message": "Feature flag updated successfully"})
}

// DeleteFlag deletes a feature flag (admin only)
// @Summary Delete feature flag
// @Tags ab-test
// @Produce json
// @Security Bearer
// @Param flag_id path string true "Feature flag ID"
// @Success 204
// @Failure 404 {object} response.ErrorResponse
// @Router /ab-test/flags/{flag_id} [delete]
func (h *ABTestHandler) DeleteFlag(c *gin.Context) {
	flagID := c.Param("flag_id")

	err := h.featureFlagService.DeleteFlag(flagID)
	if err != nil {
		response.NotFound(c, "Feature flag not found")
		return
	}

	response.NoContent(c)
}
