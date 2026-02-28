package iap

import (
	"context"
	"fmt"
	"strconv"
	"time"

	iap "github.com/awa/go-iap/appstore"
)

// AppleVerifier verifies Apple IAP receipts
type AppleVerifier struct {
	sharedSecret string
	environment  iap.Environment // iap.Sandbox or iap.Production
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
	client := iap.New()

	// Verify the receipt
	req := iap.IAPRequest{
		ReceiptData: receiptData,
		Password:    v.sharedSecret,
	}

	var result iap.IAPResponse
	if err := client.Verify(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to verify receipt: %w", err)
	}

	// Check status
	if result.Status != 0 {
		return &VerifyResponse{
			Valid: false,
		}, nil
	}

	// Parse receipt info
	latestReceipts := result.LatestReceiptInfo
	if len(latestReceipts) == 0 {
		return &VerifyResponse{
			Valid: false,
		}, nil
	}
	first := latestReceipts[0]

	// Calculate expires at
	expiresAt := time.Now()
	if first.ExpiresDateMS != "" {
		if ms, err := strconv.ParseInt(first.ExpiresDateMS, 10, 64); err == nil {
			expiresAt = time.Unix(ms/1000, 0)
		}
	} else if first.ExpiresDate.ExpiresDate != "" {
		expiresAt, _ = time.Parse(time.RFC3339, first.ExpiresDate.ExpiresDate)
	}

	return &VerifyResponse{
		Valid:         true,
		TransactionID: first.TransactionID,
		ProductID:     first.ProductID,
		ExpiresAt:     expiresAt,
		IsRenewable:   first.IsInIntroOfferPeriod == "false",
		OriginalTxID:  string(first.OriginalTransactionID),
	}, nil
}
