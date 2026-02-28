package iap

import (
	"context"
	"fmt"
	"time"

	"github.com/bivex/paywall-iap/internal/application/command"
)

// IAPAdapter implements the IAPVerifier interface for the command layer
type IAPAdapter struct {
	appleVerifier   *AppleVerifier
	googleVerifier *GoogleVerifier
}

// NewIAPAdapter creates a new IAP adapter
func NewIAPAdapter(appleVerifier *AppleVerifier, googleVerifier *GoogleVerifier) *IAPAdapter {
	return &IAPAdapter{
		appleVerifier:  appleVerifier,
		googleVerifier: googleVerifier,
	}
}

// IAPVerificationResult represents the result from the adapter
type IAPVerificationResult struct {
	Valid        bool
	TransactionID string
	ProductID    string
	ExpiresAt    time.Time
	IsRenewable  bool
	OriginalTxID string
}

// VerifyReceipt verifies an IAP receipt (platform-agnostic)
func (a *IAPAdapter) VerifyReceipt(ctx context.Context, receiptData string) (*IAPVerificationResult, error) {
	// This adapter needs to know which platform to use
	// For now, we'll return a mock implementation
	return &IAPVerificationResult{
		Valid:        true,
		TransactionID: "adapter-mock-tx",
		ProductID:    "com.yourapp.premium.monthly",
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
		IsRenewable:  true,
		OriginalTxID: "adapter-mock-original",
	}, nil
}

// VerifyAppleReceipt verifies an Apple receipt
func (a *IAPAdapter) VerifyAppleReceipt(ctx context.Context, receiptData string) (*command.IAPVerificationResult, error) {
	result, err := a.appleVerifier.VerifyReceipt(ctx, receiptData)
	if err != nil {
		return nil, err
	}

	return &command.IAPVerificationResult{
		Valid:         result.Valid,
		TransactionID: result.TransactionID,
		ProductID:     result.ProductID,
		ExpiresAt:     result.ExpiresAt,
		IsRenewable:   result.IsRenewable,
		OriginalTxID:  result.OriginalTxID,
	}, nil
}

// VerifyGoogleReceipt verifies a Google Play receipt
func (a *IAPAdapter) VerifyGoogleReceipt(ctx context.Context, receiptData string) (*command.IAPVerificationResult, error) {
	result, err := a.googleVerifier.VerifyReceipt(ctx, receiptData)
	if err != nil {
		return nil, err
	}

	return &command.IAPVerificationResult{
		Valid:         result.Valid,
		TransactionID: result.TransactionID,
		ProductID:     result.ProductID,
		ExpiresAt:     result.ExpiresAt,
		IsRenewable:   result.IsRenewable,
		OriginalTxID:  result.OriginalTxID,
	}, nil
}
