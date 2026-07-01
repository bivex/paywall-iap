package iap

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/application/command"
)

// DynamicAppleVerifier resolves Apple credentials per app_id at verify time.
type DynamicAppleVerifier struct {
	resolver *CredentialResolver
	mockURL  string // dev override
}

func NewDynamicAppleVerifier(resolver *CredentialResolver, mockURL string) *DynamicAppleVerifier {
	return &DynamicAppleVerifier{resolver: resolver, mockURL: mockURL}
}

func (v *DynamicAppleVerifier) VerifyReceipt(ctx context.Context, appID uuid.UUID, receiptData string) (*command.IAPVerificationResult, error) {
	creds, err := v.resolver.Resolve(ctx, appID, "apple")
	if err != nil {
		return nil, fmt.Errorf("apple credentials not configured for app %s: %w", appID, err)
	}

	isProduction := creds.AppleEnvironment != "sandbox"
	verifier := NewAppleVerifier(creds.AppleSharedSecret, isProduction, v.mockURL)
	result, err := verifier.VerifyReceipt(ctx, receiptData)
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

// DynamicGoogleVerifier resolves Google credentials per app_id at verify time.
type DynamicGoogleVerifier struct {
	resolver *CredentialResolver
	mockURL  string // dev override
}

func NewDynamicGoogleVerifier(resolver *CredentialResolver, mockURL string) *DynamicGoogleVerifier {
	return &DynamicGoogleVerifier{resolver: resolver, mockURL: mockURL}
}

func (v *DynamicGoogleVerifier) VerifyReceipt(ctx context.Context, appID uuid.UUID, receiptData string) (*command.IAPVerificationResult, error) {
	creds, err := v.resolver.Resolve(ctx, appID, "google")
	if err != nil {
		return nil, fmt.Errorf("google credentials not configured for app %s: %w", appID, err)
	}

	verifier := NewGoogleVerifier(creds.GoogleServiceAccount, true, v.mockURL)
	result, err := verifier.VerifyReceipt(ctx, receiptData)
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
