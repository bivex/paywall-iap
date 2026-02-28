package iap

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

// GoogleVerifier verifies Google Play IAP receipts
type GoogleVerifier struct {
	serviceAccountJSON string
	isProduction      bool
}

// NewGoogleVerifier creates a new Google verifier
func NewGoogleVerifier(serviceAccountJSON string, isProduction bool) *GoogleVerifier {
	return &GoogleVerifier{
		serviceAccountJSON: serviceAccountJSON,
		isProduction:       isProduction,
	}
}

// VerifyReceipt verifies a Google Play IAP receipt
func (v *GoogleVerifier) VerifyReceipt(ctx context.Context, receiptData string) (*VerifyResponse, error) {
	// For development/mvp, return a mock response
	// In production, this would call the Google Play Developer API
	if v.serviceAccountJSON == "" {
		// Mock response for development
		return &VerifyResponse{
			Valid:         true,
			TransactionID: "android-mock-tx-" + receiptData[:10],
			ProductID:     "com.yourapp.premium.monthly",
			ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
			IsRenewable:   true,
			OriginalTxID:  "android-mock-original-tx",
		}, nil
	}

	// Parse service account JSON and create OAuth2 client
	conf, err := google.CredentialsFromJSON(
		ctx,
		[]byte(v.serviceAccountJSON),
		androidpublisher.AndroidpublisherScope,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account credentials: %w", err)
	}

	// Build the Android Publisher client
	service, err := androidpublisher.NewService(ctx, option.WithTokenSource(conf.TokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create Android Publisher service: %w", err)
	}

	// receiptData is a JSON string: {"packageName":"...","productId":"...","purchaseToken":"..."}
	var receipt struct {
		PackageName   string `json:"packageName"`
		ProductID     string `json:"productId"`
		PurchaseToken string `json:"purchaseToken"`
	}
	if err := json.Unmarshal([]byte(receiptData), &receipt); err != nil {
		return nil, fmt.Errorf("failed to parse receipt data: %w", err)
	}

	sub, err := service.Purchases.Subscriptions.Get(
		receipt.PackageName,
		receipt.ProductID,
		receipt.PurchaseToken,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to verify Google Play subscription: %w", err)
	}

	// sub.ExpiryTimeMillis is milliseconds since epoch
	expiresAt := time.Unix(sub.ExpiryTimeMillis/1000, 0)
	isValid := sub.PaymentState != nil && *sub.PaymentState == 1 // 1 = payment received

	return &VerifyResponse{
		Valid:         isValid,
		TransactionID: receipt.PurchaseToken,
		ProductID:     receipt.ProductID,
		ExpiresAt:     expiresAt,
		IsRenewable:   sub.AutoRenewing,
		OriginalTxID:  receipt.PurchaseToken,
	}, nil
}
