package iap

import (
	"context"
	"errors"
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

// VerifyReceipt is not implemented on IAPAdapter directly — use VerifyAppleReceipt
// or VerifyGoogleReceipt (or the platform-specific adapter wrappers).
func (a *IAPAdapter) VerifyReceipt(_ context.Context, _ string) (*IAPVerificationResult, error) {
	return nil, errors.New("IAPAdapter.VerifyReceipt: use platform-specific adapter (AppleVerifierAdapter or AndroidVerifierAdapter)")
}

// suppress "time imported and not used" — time is still used by Apple/Google helpers below
var _ = time.Time{}

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

// AppleVerifierAdapter wraps IAPAdapter to implement command.IAPVerifier for iOS
type AppleVerifierAdapter struct{ adapter *IAPAdapter }

// NewAppleVerifierAdapter creates an adapter that satisfies command.IAPVerifier for iOS
func NewAppleVerifierAdapter(a *IAPAdapter) *AppleVerifierAdapter {
	return &AppleVerifierAdapter{adapter: a}
}

// VerifyReceipt implements command.IAPVerifier
func (v *AppleVerifierAdapter) VerifyReceipt(ctx context.Context, receiptData string) (*command.IAPVerificationResult, error) {
	return v.adapter.VerifyAppleReceipt(ctx, receiptData)
}

// AndroidVerifierAdapter wraps IAPAdapter to implement command.IAPVerifier for Android
type AndroidVerifierAdapter struct{ adapter *IAPAdapter }

// NewAndroidVerifierAdapter creates an adapter that satisfies command.IAPVerifier for Android
func NewAndroidVerifierAdapter(a *IAPAdapter) *AndroidVerifierAdapter {
	return &AndroidVerifierAdapter{adapter: a}
}

// VerifyReceipt implements command.IAPVerifier
func (v *AndroidVerifierAdapter) VerifyReceipt(ctx context.Context, receiptData string) (*command.IAPVerificationResult, error) {
	return v.adapter.VerifyGoogleReceipt(ctx, receiptData)
}
