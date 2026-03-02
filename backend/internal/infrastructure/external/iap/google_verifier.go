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
	isProduction       bool
	// baseURL overrides the Google Play API endpoint (used for mock/testing)
	baseURL string
}

// NewGoogleVerifier creates a new Google verifier.
// Pass baseURL (e.g. "http://localhost:8080") to redirect calls to a mock server.
func NewGoogleVerifier(serviceAccountJSON string, isProduction bool, baseURL string) *GoogleVerifier {
	return &GoogleVerifier{
		serviceAccountJSON: serviceAccountJSON,
		isProduction:       isProduction,
		baseURL:            baseURL,
	}
}

// VerifyReceipt verifies a Google Play IAP receipt
func (v *GoogleVerifier) VerifyReceipt(ctx context.Context, receiptData string) (*VerifyResponse, error) {
	// For development/mvp, return a mock response
	// In production, this would call the Google Play Developer API
	if v.serviceAccountJSON == "" && v.baseURL == "" {
		// No credentials and no mock server — return hardcoded dev response
		return &VerifyResponse{
			Valid:         true,
			TransactionID: "android-mock-tx-" + receiptData[:10],
			ProductID:     "com.yourapp.premium.monthly",
			ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
			IsRenewable:   true,
			OriginalTxID:  "android-mock-original-tx",
		}, nil
	}

	// Build client options
	var opts []option.ClientOption
	if v.baseURL != "" {
		// Mock / test mode: skip auth and point to local server
		opts = append(opts, option.WithEndpoint(v.baseURL), option.WithoutAuthentication())
	} else {
		// Production / sandbox: use real service-account credentials
		conf, err := google.CredentialsFromJSON(
			ctx,
			[]byte(v.serviceAccountJSON),
			androidpublisher.AndroidpublisherScope,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to parse service account credentials: %w", err)
		}
		opts = append(opts, option.WithTokenSource(conf.TokenSource))
	}

	// Build the Android Publisher client
	service, err := androidpublisher.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Android Publisher service: %w", err)
	}

	// receiptData is a JSON string:
	// {"packageName":"...","productId":"...","purchaseToken":"...","type":"subscription|product"}
	// type defaults to "subscription" when omitted.
	var receipt struct {
		PackageName   string `json:"packageName"`
		ProductID     string `json:"productId"`
		PurchaseToken string `json:"purchaseToken"`
		Type          string `json:"type"` // "subscription" (default) or "product"
	}
	if err := json.Unmarshal([]byte(receiptData), &receipt); err != nil {
		return nil, fmt.Errorf("failed to parse receipt data: %w", err)
	}

	if receipt.Type == "product" {
		prod, err := service.Purchases.Products.Get(
			receipt.PackageName,
			receipt.ProductID,
			receipt.PurchaseToken,
		).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to verify Google Play product: %w", err)
		}
		// purchaseState: 0=purchased, 1=canceled, 2=pending
		isValid := prod.PurchaseState == 0
		return &VerifyResponse{
			Valid:         isValid,
			TransactionID: receipt.PurchaseToken,
			ProductID:     receipt.ProductID,
			ExpiresAt:     time.Time{}, // one-time products have no expiry
			IsRenewable:   false,
			OriginalTxID:  receipt.PurchaseToken,
		}, nil
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
	// Valid: payment received AND not expired AND not canceled (auto-renewing or still within paid period)
	paymentReceived := sub.PaymentState != nil && *sub.PaymentState == 1
	notExpired := expiresAt.After(time.Now())
	notCanceled := sub.CancelReason == 0 && sub.UserCancellationTimeMillis == 0
	isValid := paymentReceived && notExpired && notCanceled

	return &VerifyResponse{
		Valid:         isValid,
		TransactionID: receipt.PurchaseToken,
		ProductID:     receipt.ProductID,
		ExpiresAt:     expiresAt,
		IsRenewable:   sub.AutoRenewing,
		OriginalTxID:  receipt.PurchaseToken,
	}, nil
}
