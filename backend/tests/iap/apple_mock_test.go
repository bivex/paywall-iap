//go:build integration

package iap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iapext "github.com/bivex/paywall-iap/internal/infrastructure/external/iap"
)

const (
	appleTestProduct        = "com.mothsalt.game1.premium.monthly"
	appleTestProductYearly  = "com.mothsalt.game1.premium.yearly"
)

// appleMockURL returns the Apple IAP Mock server URL.
// Override with APPLE_MOCK_URL env var (default: http://localhost:9090).
func appleMockURL() string {
	if u := os.Getenv("APPLE_MOCK_URL"); u != "" {
		return u
	}
	return "http://localhost:9090"
}

func newMockAppleVerifier() *iapext.AppleVerifier {
	return iapext.NewAppleVerifier("", false, appleMockURL())
}

// createAppleSub calls POST /subs on the mock to create a subscription
// and returns the receiptToken.
func createAppleSub(t *testing.T, productID string) string {
	t.Helper()
	body := fmt.Sprintf(`{"productId":%q}`, productID)
	resp, err := http.Post(appleMockURL()+"/subs", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "createAppleSub: unexpected status")

	var sub struct {
		ReceiptToken string `json:"receiptToken"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sub))
	require.NotEmpty(t, sub.ReceiptToken, "receiptToken must not be empty")
	return sub.ReceiptToken
}

// cancelAppleSub calls POST /subs/:token/cancel on the mock.
func cancelAppleSub(t *testing.T, token string) {
	t.Helper()
	resp, err := http.Post(appleMockURL()+"/subs/"+token+"/cancel", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "cancelAppleSub: unexpected status")
}

// renewAppleSub calls POST /subs/:token/renew on the mock.
func renewAppleSub(t *testing.T, token string) {
	t.Helper()
	resp, err := http.Post(appleMockURL()+"/subs/"+token+"/renew", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "renewAppleSub: unexpected status")
}

// refundAppleSub calls POST /subs/:token/refund/:txID on the mock.
func refundAppleSub(t *testing.T, token string) {
	t.Helper()
	// First verify to get a transaction ID
	resp, err := newMockAppleVerifier().VerifyReceipt(context.Background(), token)
	require.NoError(t, err)
	require.True(t, resp.Valid)

	url := fmt.Sprintf("%s/subs/%s/refund/%s", appleMockURL(), token, resp.TransactionID)
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	client := &http.Client{}
	r, err := client.Do(req)
	require.NoError(t, err)
	defer r.Body.Close()
	require.True(t, r.StatusCode == http.StatusOK || r.StatusCode == http.StatusNoContent,
		"refundAppleSub: unexpected status %d", r.StatusCode)
}

// ── Tests ──────────────────────────────────────────────────────────────────

func TestAppleMock_ValidActiveSubscription(t *testing.T) {
	token := createAppleSub(t, appleTestProduct)

	resp, err := newMockAppleVerifier().VerifyReceipt(context.Background(), token)
	require.NoError(t, err)
	assert.True(t, resp.Valid, "expected valid=true for fresh active subscription")
	assert.NotEmpty(t, resp.TransactionID)
	assert.NotEmpty(t, resp.ProductID)
	assert.True(t, resp.ExpiresAt.After(time.Now()), "expires_at must be in the future")
}

func TestAppleMock_YearlySubscription(t *testing.T) {
	token := createAppleSub(t, appleTestProductYearly)

	resp, err := newMockAppleVerifier().VerifyReceipt(context.Background(), token)
	require.NoError(t, err)
	assert.True(t, resp.Valid, "expected valid=true for yearly subscription")
	assert.True(t, resp.ExpiresAt.After(time.Now().Add(30*24*time.Hour)),
		"yearly sub should expire more than 30 days from now")
}

func TestAppleMock_CancelledSubscription(t *testing.T) {
	token := createAppleSub(t, appleTestProduct)
	cancelAppleSub(t, token)

	resp, err := newMockAppleVerifier().VerifyReceipt(context.Background(), token)
	require.NoError(t, err)
	assert.False(t, resp.Valid, "expected valid=false for cancelled subscription")
}

func TestAppleMock_RefundedSubscription(t *testing.T) {
	token := createAppleSub(t, appleTestProduct)
	refundAppleSub(t, token)

	resp, err := newMockAppleVerifier().VerifyReceipt(context.Background(), token)
	require.NoError(t, err)
	assert.False(t, resp.Valid, "expected valid=false for refunded transaction")
}

func TestAppleMock_RenewedSubscription(t *testing.T) {
	token := createAppleSub(t, appleTestProduct)

	// Verify once to get initial state
	resp1, err := newMockAppleVerifier().VerifyReceipt(context.Background(), token)
	require.NoError(t, err)
	require.True(t, resp1.Valid)
	firstTxID := resp1.TransactionID

	// Renew
	renewAppleSub(t, token)

	// Verify again — should still be valid, new transaction
	resp2, err := newMockAppleVerifier().VerifyReceipt(context.Background(), token)
	require.NoError(t, err)
	assert.True(t, resp2.Valid, "expected valid=true after renewal")
	assert.NotEqual(t, firstTxID, resp2.TransactionID, "expected new transaction ID after renewal")
}

func TestAppleMock_UnknownReceiptToken_InvalidStatus(t *testing.T) {
	// A receipt token that was never created should return status != 0 (invalid)
	resp, err := newMockAppleVerifier().VerifyReceipt(context.Background(), "totally_unknown_token_xyz")
	require.NoError(t, err)
	assert.False(t, resp.Valid, "unknown token should return valid=false")
}

func TestAppleMock_EmptyReceiptData_Error(t *testing.T) {
	// Empty receipt data — verifier should return an error or invalid
	resp, err := newMockAppleVerifier().VerifyReceipt(context.Background(), "")
	if err != nil {
		// Acceptable: network/parse error
		return
	}
	assert.False(t, resp.Valid, "empty receipt should not be valid")
}

func TestAppleMock_MultipleProductsIsolated(t *testing.T) {
	// Two subs for different products should verify independently
	token1 := createAppleSub(t, appleTestProduct)
	token2 := createAppleSub(t, appleTestProductYearly)

	// Cancel only first
	cancelAppleSub(t, token1)

	resp1, err := newMockAppleVerifier().VerifyReceipt(context.Background(), token1)
	require.NoError(t, err)
	assert.False(t, resp1.Valid, "cancelled sub should be invalid")

	resp2, err := newMockAppleVerifier().VerifyReceipt(context.Background(), token2)
	require.NoError(t, err)
	assert.True(t, resp2.Valid, "uncancelled sub should remain valid")
}
