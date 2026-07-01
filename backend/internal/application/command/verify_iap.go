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

// IAPVerifier is the legacy interface (static credentials, no app_id).
// Kept for backward compatibility with tests and adapter wrappers.
type IAPVerifier interface {
	VerifyReceipt(ctx context.Context, receiptData string) (*IAPVerificationResult, error)
}

// DynamicIAPVerifier is the preferred interface — resolves credentials per app_id at call time.
type DynamicIAPVerifier interface {
	VerifyReceipt(ctx context.Context, appID uuid.UUID, receiptData string) (*IAPVerificationResult, error)
}

// IAPVerificationResult represents the result of IAP verification.
type IAPVerificationResult struct {
	Valid         bool
	TransactionID string
	ProductID     string
	ExpiresAt     time.Time
	IsRenewable   bool
	OriginalTxID  string
}

// staticVerifierAdapter wraps a legacy IAPVerifier as a DynamicIAPVerifier,
// ignoring the appID (used when dynamic credentials are not configured).
type staticVerifierAdapter struct{ v IAPVerifier }

func (a *staticVerifierAdapter) VerifyReceipt(ctx context.Context, _ uuid.UUID, receiptData string) (*IAPVerificationResult, error) {
	return a.v.VerifyReceipt(ctx, receiptData)
}

// VerifyIAPCommand handles IAP receipt verification.
type VerifyIAPCommand struct {
	userRepo         repository.UserRepository
	subscriptionRepo repository.SubscriptionRepository
	transactionRepo  repository.TransactionRepository
	iosVerifier      DynamicIAPVerifier
	androidVerifier  DynamicIAPVerifier
}

// NewVerifyIAPCommand creates a new verify IAP command with dynamic (per-app) verifiers.
func NewVerifyIAPCommand(
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
	transactionRepo repository.TransactionRepository,
	iosVerifier DynamicIAPVerifier,
	androidVerifier DynamicIAPVerifier,
) *VerifyIAPCommand {
	return &VerifyIAPCommand{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		transactionRepo:  transactionRepo,
		iosVerifier:      iosVerifier,
		androidVerifier:  androidVerifier,
	}
}

// NewVerifyIAPCommandLegacy wraps legacy static verifiers for tests / backward compat.
func NewVerifyIAPCommandLegacy(
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
	transactionRepo repository.TransactionRepository,
	iosVerifier IAPVerifier,
	androidVerifier IAPVerifier,
) *VerifyIAPCommand {
	return NewVerifyIAPCommand(
		userRepo, subscriptionRepo, transactionRepo,
		&staticVerifierAdapter{iosVerifier},
		&staticVerifierAdapter{androidVerifier},
	)
}

// Execute executes the verify IAP command.
// appID is the app the user belongs to — used to select per-app store credentials.
func (c *VerifyIAPCommand) Execute(ctx context.Context, userID string, appID uuid.UUID, req *dto.VerifyIAPRequest) (*dto.VerifyIAPResponse, error) {
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
	var verifier DynamicIAPVerifier
	if req.Platform == "ios" {
		verifier = c.iosVerifier
	} else {
		verifier = c.androidVerifier
	}

	// Verify receipt — uses per-app credentials from app_credentials table
	result, err := verifier.VerifyReceipt(ctx, appID, req.ReceiptData)
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
		sub, err := c.subscriptionRepo.GetActiveByUserID(ctx, userUUID)
		if err != nil {
			// Subscription may have been cancelled after the receipt was processed.
			// Still idempotent — return a clear error rather than an internal failure.
			return nil, fmt.Errorf("%w: receipt already processed", domainErrors.ErrReceiptAlreadyProcessed)
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
		existingSub.ExpiresAt = result.ExpiresAt
		if err := c.subscriptionRepo.Update(ctx, existingSub); err != nil {
			return nil, fmt.Errorf("failed to update subscription: %w", err)
		}
		sub = existingSub
	} else {
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
		_ = c.userRepo.UpdatePurchaseChannel(ctx, userUUID, entity.PurchaseChannelIAP)
	}

	// Create transaction record
	txn := entity.NewTransaction(appID, userUUID, sub.ID, 0, "USD")
	txn.ReceiptHash = receiptHash
	txn.ProviderTxID = result.TransactionID
	if err := c.transactionRepo.Create(ctx, txn); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update LTV — best-effort, don't fail the whole request
	_ = c.userRepo.IncrementLTV(ctx, userUUID, priceFromPlanType(planType))

	return c.toSubscriptionResponse(sub, isNew), nil
}

func (c *VerifyIAPCommand) determinePlanType(productID string) entity.PlanType {
	if len(productID) > 0 {
		if containsIgnoreCase(productID, "annual") || containsIgnoreCase(productID, "year") {
			return entity.PlanAnnual
		}
	}
	return entity.PlanMonthly
}

func priceFromPlanType(planType entity.PlanType) float64 {
	switch planType {
	case entity.PlanAnnual:
		return 49.99
	default:
		return 9.99
	}
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
	if len(req.ProductID) < 3 {
		return domainErrors.NewValidationError("product_id", "must be at least 3 characters")
	}
	if len(req.ProductID) > 200 {
		return domainErrors.NewValidationError("product_id", "must not exceed 200 characters")
	}
	if !strings.Contains(req.ProductID, ".") {
		return domainErrors.NewValidationError("product_id", "must be in reverse-domain notation (e.g. com.app.product)")
	}
	if len(req.ReceiptData) > 65536 {
		return domainErrors.NewValidationError("receipt_data", "exceeds maximum allowed size (64 KB)")
	}
	if req.Platform == "android" {
		if err := validateAndroidReceipt(req.ReceiptData, req.ProductID); err != nil {
			return err
		}
	}
	return nil
}

type androidReceiptPayload struct {
	PackageName   string `json:"packageName"`
	ProductID     string `json:"productId"`
	PurchaseToken string `json:"purchaseToken"`
	Type          string `json:"type"`
}

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
	if !strings.EqualFold(payload.ProductID, requestProductID) {
		return domainErrors.NewValidationError("product_id",
			fmt.Sprintf("mismatch: request has %q but receipt_data.productId is %q", requestProductID, payload.ProductID))
	}
	return nil
}
