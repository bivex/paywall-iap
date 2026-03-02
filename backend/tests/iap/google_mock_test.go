//go:build integration

package iap_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iapext "github.com/bivex/paywall-iap/internal/infrastructure/external/iap"
)

// mockBaseURL returns the Google Billing Mock server URL.
// Override with GOOGLE_IAP_BASE_URL env var (default: http://localhost:8080).
func mockBaseURL() string {
	if u := os.Getenv("GOOGLE_IAP_BASE_URL"); u != "" {
		return u
	}
	return "http://localhost:8090"
}

func receipt(pkg, productID, token string) string {
	return receiptWithType(pkg, productID, token, "subscription")
}

func receiptWithType(pkg, productID, token, rtype string) string {
	b, _ := json.Marshal(map[string]string{
		"packageName":   pkg,
		"productId":     productID,
		"purchaseToken": token,
		"type":          rtype,
	})
	return string(b)
}

func newMockGoogleVerifier() *iapext.GoogleVerifier {
	return iapext.NewGoogleVerifier("", false, mockBaseURL())
}

const (
	testPackage = "com.example.app"
	testProduct = "com.example.sub"
)

func TestGoogleMock_ValidActiveSubscription(t *testing.T) {
	resp, err := newMockGoogleVerifier().VerifyReceipt(context.Background(),
		receipt(testPackage, testProduct, "valid_active_user_123"))
	require.NoError(t, err)
	assert.True(t, resp.Valid, "expected valid=true for valid_active token")
	assert.True(t, resp.IsRenewable)
	assert.NotZero(t, resp.ExpiresAt)
}

func TestGoogleMock_ExpiredSubscription(t *testing.T) {
	resp, err := newMockGoogleVerifier().VerifyReceipt(context.Background(),
		receipt(testPackage, testProduct, "expired_user_456"))
	require.NoError(t, err)
	assert.False(t, resp.Valid, "expected valid=false for expired token")
	assert.False(t, resp.IsRenewable)
}

func TestGoogleMock_CanceledSubscription(t *testing.T) {
	resp, err := newMockGoogleVerifier().VerifyReceipt(context.Background(),
		receipt(testPackage, testProduct, "canceled_user_789"))
	require.NoError(t, err)
	assert.False(t, resp.Valid, "expected valid=false for canceled token")
}

func TestGoogleMock_InvalidToken_Returns410(t *testing.T) {
	_, err := newMockGoogleVerifier().VerifyReceipt(context.Background(),
		receipt(testPackage, testProduct, "invalid_token_abc"))
	require.Error(t, err, "expected error for 410 expired token")
	assert.Contains(t, err.Error(), "verify Google Play subscription")
}

func TestGoogleMock_ValidProduct(t *testing.T) {
	resp, err := newMockGoogleVerifier().VerifyReceipt(context.Background(),
		receiptWithType(testPackage, "com.example.product1", "product_valid_token_001", "product"))
	require.NoError(t, err)
	assert.True(t, resp.Valid)
}

func TestGoogleMock_PendingSubscription(t *testing.T) {
	resp, err := newMockGoogleVerifier().VerifyReceipt(context.Background(),
		receipt(testPackage, testProduct, "pending_payment_user_001"))
	require.NoError(t, err)
	assert.False(t, resp.Valid, "pending payment should not be valid")
}
