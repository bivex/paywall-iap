package iap

import (
	"context"
	"fmt"
	"time"

	iap "github.com/awa/go-iap/appstore"
)

// AppleVerifier verifies Apple IAP receipts
type AppleVerifier struct {
	sharedSecret string
	environment  string // "sandbox" or "production"
}

// NewAppleVerifier creates a new Apple verifier
func NewAppleVerifier(sharedSecret string, isProduction bool) *AppleVerifier {
	env := iap.Sandbox
	if isProduction {
		env = iap.Production
	}

	return &AppleVerifier{
		sharedSecret: sharedSecret,
		environment:  env,
	}
}

// VerifyResponse represents the verification response
type VerifyResponse struct {
	Valid         bool
	TransactionID string
	ProductID     string
	ExpiresAt     time.Time
	IsRenewable   bool
	OriginalTxID  string
}

// VerifyReceipt verifies an Apple IAP receipt
func (v *AppleVerifier) VerifyReceipt(ctx context.Context, receiptData string) (*VerifyResponse, error) {
	if v.sharedSecret == "" {
		// For development, return a mock valid response
		return &VerifyResponse{
			Valid:         true,
			TransactionID: "mock-tx-" + receiptData[:10],
			ProductID:     "com.yourapp.premium.monthly",
			ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
			IsRenewable:   true,
			OriginalTxID:  "mock-original-tx",
		}, nil
	}

	// Create IAP client
	client := iap.New(iap.Config{
		PrivateKey: []byte(v.sharedSecret), // For local validation
	})

	// Verify the receipt
	req := iap.IAPRequest{
		ReceiptData: receiptData,
		Password:    v.sharedSecret,
	}

	result, err := client.Verify(ctx, v.environment, req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify receipt: %w", err)
	}

	// Check status
	if result.Status != 0 {
		return &VerifyResponse{
			Valid: false,
		}, nil
	}

	// Parse receipt info
	latestReceipt := result.LatestReceiptInfo
	if latestReceipt == nil {
		return &VerifyResponse{
			Valid: false,
		}, nil
	}

	// Calculate expires at
	expiresAt := time.Now()
	if latestReceipt.ExpiresDateMs > 0 {
		expiresAt = time.Unix(latestReceipt.ExpiresDateMs/1000, 0)
	} else if latestReceipt.ExpiresDate > 0 {
		// Try ExpiresDate field
		expiresAtStr := latestReceipt.ExpiresDate
		expiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)
	}

	return &VerifyResponse{
		Valid:         true,
		TransactionID: latestReceipt.TransactionID,
		ProductID:     latestReceipt.ProductID,
		ExpiresAt:     expiresAt,
		IsRenewable:   latestReceipt.IsInIntroOfferPeriod == "false",
		OriginalTxID:  latestReceipt.OriginalTransactionID,
	}, nil
}
