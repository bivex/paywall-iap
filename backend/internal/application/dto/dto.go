package dto

// ========== AUTH DTOs ==========

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	PlatformUserID string `json:"platform_user_id" binding:"required"`
	DeviceID       string `json:"device_id" binding:"required"`
	Platform       string `json:"platform" binding:"required,oneof=ios android"`
	AppVersion     string `json:"app_version" binding:"required"`
	Email          string `json:"email" binding:"omitempty,email"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	UserID          string `json:"user_id"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	ExpiresIn       int64  `json:"expires_in"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// ========== IAP VERIFICATION DTOs ==========

// VerifyIAPRequest represents an IAP verification request
type VerifyIAPRequest struct {
	Platform     string `json:"platform" binding:"required,oneof=ios android"`
	ReceiptData  string `json:"receipt_data" binding:"required"`
	ProductID    string `json:"product_id" binding:"required"`
	TransactionID string `json:"transaction_id,omitempty"`
}

// VerifyIAPResponse represents an IAP verification response
type VerifyIAPResponse struct {
	SubscriptionID   string `json:"subscription_id"`
	Status           string `json:"status"`
	ExpiresAt        string `json:"expires_at"`
	AutoRenew        bool   `json:"auto_renew"`
	PlanType         string `json:"plan_type"`
	IsNew            bool   `json:"is_new"`
}

// ========== SUBSCRIPTION DTOs ==========

// SubscriptionResponse represents a subscription response
type SubscriptionResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Source       string `json:"source"`
	Platform     string `json:"platform"`
	ProductID    string `json:"product_id"`
	PlanType     string `json:"plan_type"`
	ExpiresAt    string `json:"expires_at"`
	AutoRenew    bool   `json:"auto_renew"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// AccessCheckResponse represents an access check response
type AccessCheckResponse struct {
	HasAccess   bool   `json:"has_access"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

// CancelSubscriptionRequest represents a cancel subscription request
type CancelSubscriptionRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ========== PRICING DTOs ==========

// PricingTier represents a pricing tier
type PricingTier struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	MonthlyPrice  float64 `json:"monthly_price"`
	AnnualPrice   float64 `json:"annual_price"`
	Currency      string  `json:"currency"`
	Features      []string `json:"features"`
	IsActive      bool    `json:"is_active"`
}

// ========== ANALYTICS DTOs ==========

// RevenueOverview represents revenue overview data
type RevenueOverview struct {
	TotalRevenue     float64 `json:"total_revenue"`
	MonthlyRevenue   float64 `json:"monthly_revenue"`
	ActiveSubscriptions int  `json:"active_subscriptions"`
	ChurnRate        float64 `json:"churn_rate"`
}

// ========== ERROR DTOs ==========

// ErrorDetail represents a detailed error
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Error   string       `json:"error"`
	Details []ErrorDetail `json:"details"`
}
