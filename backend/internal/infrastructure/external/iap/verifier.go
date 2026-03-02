package iap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	iap "github.com/awa/go-iap/appstore"
)

// numericStringPatcher is an http.RoundTripper that patches non-numeric
// original_transaction_id values so go-iap's NumericString can parse them.
type numericStringPatcher struct {
	wrapped http.RoundTripper
}

var nonNumericOrigTxRe = regexp.MustCompile(`"original_transaction_id"\s*:\s*"([^"]+)"`)

// msFieldRe matches numeric (unquoted) _ms timestamp fields.
var msFieldRe = regexp.MustCompile(`"([a-z_]+_ms)"\s*:\s*(\d+)`)

func (p *numericStringPatcher) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := p.wrapped.RoundTrip(req)
	if err != nil || resp == nil {
		return resp, err
	}

	body, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		resp.Body = io.NopCloser(bytes.NewReader(body))
		return resp, nil
	}

	// 1. Patch non-numeric original_transaction_id → stable numeric hash.
	patched := nonNumericOrigTxRe.ReplaceAllFunc(body, func(match []byte) []byte {
		sub := nonNumericOrigTxRe.FindSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		val := string(sub[1])
		if _, err := strconv.ParseInt(val, 10, 64); err == nil {
			return match
		}
		var h uint64 = 14695981039346656037
		for _, c := range []byte(val) {
			h ^= uint64(c)
			h *= 1099511628211
		}
		return []byte(fmt.Sprintf(`"original_transaction_id":%d`, h&0x7FFFFFFFFFFFFFFF))
	})

	// 2. Patch numeric _ms fields → quoted strings (go-iap expects string for all *DateMS).
	patched = msFieldRe.ReplaceAll(patched, []byte(`"$1":"$2"`))

	resp.Body = io.NopCloser(bytes.NewReader(patched))
	resp.ContentLength = int64(len(patched))
	return resp, nil
}

// AppleVerifier verifies Apple IAP receipts
type AppleVerifier struct {
	sharedSecret string
	environment  iap.Environment // iap.Sandbox or iap.Production
	mockURL      string          // if set, overrides both Sandbox and Production URLs (for local testing)
}

// NewAppleVerifier creates a new Apple verifier.
// If mockURL is non-empty (APPLE_MOCK_URL env), all receipt verifications are
// redirected to the local apple-of-my-iap mock server instead of Apple's servers.
func NewAppleVerifier(sharedSecret string, isProduction bool, mockURL string) *AppleVerifier {
	env := iap.Sandbox
	if isProduction {
		env = iap.Production
	}

	return &AppleVerifier{
		sharedSecret: sharedSecret,
		environment:  env,
		mockURL:      mockURL,
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
	// No secret and no mock → dev stub (always valid)
	if v.sharedSecret == "" && v.mockURL == "" {
		return &VerifyResponse{
			Valid:         true,
			TransactionID: "mock-tx-" + receiptData[:min(10, len(receiptData))],
			ProductID:     "com.yourapp.premium.monthly",
			ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
			IsRenewable:   true,
			OriginalTxID:  "mock-original-tx",
		}, nil
	}

	// Build go-iap client and optionally redirect to mock.
	// When using the local mock, also wrap the transport to patch non-numeric
	// original_transaction_id values (mock uses string IDs; go-iap expects numbers).
	var client *iap.Client
	if v.mockURL != "" {
		client = iap.NewWithClient(&http.Client{
			Transport: &numericStringPatcher{wrapped: http.DefaultTransport},
		})
		mockVerifyURL := v.mockURL + "/verifyReceipt"
		client.ProductionURL = mockVerifyURL
		client.SandboxURL = mockVerifyURL
	} else {
		client = iap.New()
	}

	req := iap.IAPRequest{
		ReceiptData: receiptData,
		Password:    v.sharedSecret,
	}

	var result iap.IAPResponse
	if err := client.Verify(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to verify receipt: %w", err)
	}

	// Non-zero status → invalid receipt
	if result.Status != 0 {
		return &VerifyResponse{Valid: false}, nil
	}

	latestReceipts := result.LatestReceiptInfo
	if len(latestReceipts) == 0 {
		return &VerifyResponse{Valid: false}, nil
	}
	first := latestReceipts[0]

	// Refunded/cancelled transactions have cancellation_date set → reject.
	if first.CancellationDate.CancellationDate != "" || first.CancellationDate.CancellationDateMS != "" {
		return &VerifyResponse{Valid: false}, nil
	}

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
