package command

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/google/uuid"
)

// IAPVerifier interface for IAP verification services
type IAPVerifier interface {
	VerifyReceipt(ctx context.Context, receiptData string) (*IAPVerificationResult, error)
}

// IAPVerificationResult represents the result of IAP verification
type IAPVerificationResult struct {
	Valid         bool
	TransactionID string
	ProductID     string
	ExpiresAt     time.Time
	IsRenewable   bool
	OriginalTxID  string
}

// VerifyIAPCommand handles IAP receipt verification
type VerifyIAPCommand struct {
	userRepo         repository.UserRepository
	subscriptionRepo repository.SubscriptionRepository
	transactionRepo  repository.TransactionRepository
	iosVerifier      IAPVerifier
	androidVerifier  IAPVerifier
}

// NewVerifyIAPCommand creates a new verify IAP command
func NewVerifyIAPCommand(
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
	transactionRepo repository.TransactionRepository,
	iosVerifier IAPVerifier,
	androidVerifier IAPVerifier,
) *VerifyIAPCommand {
	return &VerifyIAPCommand{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		transactionRepo:  transactionRepo,
		iosVerifier:      iosVerifier,
		androidVerifier:  androidVerifier,
	}
}

// Execute executes the verify IAP command
func (c *VerifyIAPCommand) Execute(ctx context.Context, userID string, req *dto.VerifyIAPRequest) (*dto.VerifyIAPResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user ID", domainErrors.ErrInvalidInput)
	}

	// Get user (validates existence)
	if _, err := c.userRepo.GetByID(ctx, userUUID); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Validate request fields
	if err := validateIAPRequest(req); err != nil {
		return nil, err
	}

	// Select verifier based on platform
	var verifier IAPVerifier
	if req.Platform == "ios" {
		verifier = c.iosVerifier
	} else {
		verifier = c.androidVerifier
	}

	// Verify receipt
	result, err := verifier.VerifyReceipt(ctx, req.ReceiptData)
	if err != nil {
		return nil, fmt.Errorf("failed to verify receipt: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("%w: receipt is invalid", domainErrors.ErrReceiptInvalid)
	}

	// Check for duplicate receipt (idempotency)
	receiptHash := hashReceipt(req.ReceiptData)
	isDuplicate, err := c.transactionRepo.CheckDuplicateReceipt(ctx, receiptHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate receipt: %w", err)
	}
	if isDuplicate {
		// Return existing subscription instead of error
		sub, err := c.subscriptionRepo.GetActiveByUserID(ctx, userUUID)
		if err != nil {
			return nil, fmt.Errorf("receipt already processed, failed to get subscription: %w", err)
		}
		return c.toSubscriptionResponse(sub, false), nil
	}

	// Determine plan type from product ID
	planType := c.determinePlanType(req.ProductID)

	// Check for existing active subscription
	var sub *entity.Subscription
	existingSub, err := c.subscriptionRepo.GetActiveByUserID(ctx, userUUID)
	isNew := false

	if err == nil && existingSub != nil {
		// Update existing subscription
		existingSub.ExpiresAt = result.ExpiresAt
		if err := c.subscriptionRepo.Update(ctx, existingSub); err != nil {
			return nil, fmt.Errorf("failed to update subscription: %w", err)
		}
		sub = existingSub
	} else {
		// Create new subscription
		sub = entity.NewSubscription(
			userUUID,
			entity.SourceIAP,
			req.Platform,
			req.ProductID,
			planType,
			result.ExpiresAt,
		)
		if err := c.subscriptionRepo.Create(ctx, sub); err != nil {
			return nil, fmt.Errorf("failed to create subscription: %w", err)
		}
		isNew = true
	}

	// Create transaction record
	txn := entity.NewTransaction(userUUID, sub.ID, 0, "USD") // Amount would be from receipt
	txn.ReceiptHash = receiptHash
	txn.ProviderTxID = result.TransactionID
	if err := c.transactionRepo.Create(ctx, txn); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return c.toSubscriptionResponse(sub, isNew), nil
}

func (c *VerifyIAPCommand) determinePlanType(productID string) entity.PlanType {
	// Simple logic - in production, this would be more sophisticated
	if len(productID) > 0 {
		// Assume product ID contains "monthly" or "annual"
		if containsIgnoreCase(productID, "annual") || containsIgnoreCase(productID, "year") {
			return entity.PlanAnnual
		}
	}
	return entity.PlanMonthly
}

func (c *VerifyIAPCommand) toSubscriptionResponse(sub *entity.Subscription, isNew bool) *dto.VerifyIAPResponse {
	return &dto.VerifyIAPResponse{
		SubscriptionID: sub.ID.String(),
		Status:         string(sub.Status),
		ExpiresAt:      sub.ExpiresAt.Format(time.RFC3339),
		AutoRenew:      sub.AutoRenew,
		PlanType:       string(sub.PlanType),
		IsNew:          isNew,
	}
}

func hashReceipt(receiptData string) string {
	hash := sha256.Sum256([]byte(receiptData))
	return hex.EncodeToString(hash[:])
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// validateIAPRequest validates the IAP request fields before sending to verifier.
func validateIAPRequest(req *dto.VerifyIAPRequest) error {
	// product_id: min 3 chars, max 200, must be reverse-domain format (contains dot)
	if len(req.ProductID) < 3 {
		return domainErrors.NewValidationError("product_id", "must be at least 3 characters")
	}
	if len(req.ProductID) > 200 {
		return domainErrors.NewValidationError("product_id", "must not exceed 200 characters")
	}
	if !strings.Contains(req.ProductID, ".") {
		return domainErrors.NewValidationError("product_id", "must be in reverse-domain notation (e.g. com.app.product)")
	}

	// receipt_data: max 64 KB (belt-and-suspenders, handler already enforces body limit)
	if len(req.ReceiptData) > 65536 {
		return domainErrors.NewValidationError("receipt_data", "exceeds maximum allowed size (64 KB)")
	}

	// Android-specific: receipt_data must be valid JSON with required fields
	if req.Platform == "android" {
		if err := validateAndroidReceipt(req.ReceiptData, req.ProductID); err != nil {
			return err
		}
	}

	return nil
}

// androidReceiptPayload mirrors the expected JSON structure of receipt_data for Android.
type androidReceiptPayload struct {
	PackageName   string `json:"packageName"`
	ProductID     string `json:"productId"`
	PurchaseToken string `json:"purchaseToken"`
	Type          string `json:"type"`
}

// validateAndroidReceipt checks that receipt_data is valid JSON, has required fields,
// matches the product_id in the request, and has a non-empty purchaseToken.
func validateAndroidReceipt(receiptData, requestProductID string) error {
	var payload androidReceiptPayload
	if err := json.Unmarshal([]byte(receiptData), &payload); err != nil {
		return domainErrors.NewValidationError("receipt_data", "must be valid JSON for Android platform")
	}
	if payload.PackageName == "" {
		return domainErrors.NewValidationError("receipt_data", "missing required field: packageName")
	}
	if payload.ProductID == "" {
		return domainErrors.NewValidationError("receipt_data", "missing required field: productId")
	}
	if payload.PurchaseToken == "" {
		return domainErrors.NewValidationError("receipt_data", "missing required field: purchaseToken")
	}
	if payload.Type == "" {
		return domainErrors.NewValidationError("receipt_data", "missing required field: type")
	}
	if payload.Type != "subscription" && payload.Type != "inapp" {
		return domainErrors.NewValidationError("receipt_data", `field "type" must be "subscription" or "inapp"`)
	}
	// Cross-check: productId in receipt must match product_id in request
	if !strings.EqualFold(payload.ProductID, requestProductID) {
		return domainErrors.NewValidationError("product_id",
			fmt.Sprintf("mismatch: request has %q but receipt_data.productId is %q", requestProductID, payload.ProductID))
	}
	return nil
}
