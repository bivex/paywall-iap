package dto

// FeatureFlagResponse represents a feature flag in API responses
type FeatureFlagResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Enabled        bool     `json:"enabled"`
	RolloutPercent int      `json:"rollout_percent"`
	UserIDs        []string `json:"user_ids,omitempty"`
}

// CreateFeatureFlagRequest is the request to create a feature flag
type CreateFeatureFlagRequest struct {
	ID             string   `json:"id" validate:"required"`
	Name           string   `json:"name" validate:"required"`
	Enabled        bool     `json:"enabled"`
	RolloutPercent int      `json:"rollout_percent" validate:"min=0,max=100"`
	UserIDs        []string `json:"user_ids"`
}

// UpdateFeatureFlagRequest is the request to update a feature flag
type UpdateFeatureFlagRequest struct {
	Enabled        *bool    `json:"enabled,omitempty"`
	RolloutPercent *int     `json:"rollout_percent,omitempty" validate:"omitempty,min=0,max=100"`
	UserIDs        []string `json:"user_ids,omitempty"`
}

// ABTestEvaluationResponse is the response for A/B test evaluation
type ABTestEvaluationResponse struct {
	FlagID    string `json:"flag_id"`
	UserID    string `json:"user_id"`
	IsEnabled bool   `json:"is_enabled"`
	Variant   string `json:"variant,omitempty"`
}

// PaywallVariantResponse is the response for paywall variant evaluation
type PaywallVariantResponse struct {
	UserID  string `json:"user_id"`
	Variant string `json:"variant"`
}
