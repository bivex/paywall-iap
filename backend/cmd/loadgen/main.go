// cmd/loadgen/main.go — load generator for paywall-iap subscription flow.
//
// Creates N users via /api/v1/auth/register, then verifies IAP purchases
// against real mock servers (google-billing-mock, apple-of-my-iap).
// Runs concurrently with configurable workers.
//
// Usage:
//
//	go run ./cmd/loadgen [flags]
//
// Flags:
//
//	--api        API base URL (default: http://localhost:8081)
//	--app-id     App UUID (required, X-App-ID header)
//	--users      Total users to generate (default: 100)
//	--workers    Concurrent goroutines (default: 10)
//	--platform   ios|android|mixed (default: mixed)
//	--dry-run    Print requests without sending
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// ---- scenario mix ----

type scenario struct {
	name     string
	platform string
	token    string // receipt_data / purchase token
	product  string
}

var googleScenarios = []scenario{
	{name: "active_monthly", platform: "android", token: "valid_active_user_123", product: "com.mothsalt.game1.premium.monthly"},
	{name: "active_annual", platform: "android", token: "valid_active_annual_456", product: "com.mothsalt.game1.premium.yearly"},
	{name: "expired", platform: "android", token: "expired_user_789", product: "com.mothsalt.game1.premium.monthly"},
	{name: "canceled_active", platform: "android", token: "canceled_active_user_001", product: "com.mothsalt.game1.premium.monthly"},
	{name: "canceled_user", platform: "android", token: "canceled_user_002", product: "com.mothsalt.game1.premium.monthly"},
	{name: "pending", platform: "android", token: "pending_user_003", product: "com.mothsalt.game1.premium.monthly"},
}

// Apple mock product IDs (must match plans registered in apple-of-my-iap)
var appleProducts = []string{
	"com.mothsalt.game1.premium.monthly",
	"com.mothsalt.game1.premium.yearly",
}

// createAppleReceipt calls the Apple mock server to create a subscription and returns the receipt token.
func createAppleReceipt(mockURL, productID string, dryRun bool) (string, error) {
	if dryRun {
		return "dry_run_receipt_" + productID, nil
	}
	status, resp, err := doJSON("POST", mockURL+"/subs",
		nil, map[string]string{"productId": productID}, false)
	if err != nil {
		return "", fmt.Errorf("apple mock: %w", err)
	}
	if status >= 400 {
		return "", fmt.Errorf("apple mock: status %d", status)
	}
	token, _ := resp["receiptToken"].(string)
	return token, nil
}

// ---- stats ----

type stats struct {
	registered  atomic.Int64
	regFailed   atomic.Int64
	verified    atomic.Int64
	verFailed   atomic.Int64
	verActive   atomic.Int64
	verExpired  atomic.Int64
	verCanceled atomic.Int64
	verPending  atomic.Int64
	duration    time.Duration
}

func (s *stats) print() {
	fmt.Printf("\n========== LOAD GEN RESULTS ==========\n")
	fmt.Printf("Duration:       %v\n", s.duration.Round(time.Millisecond))
	fmt.Printf("Registered:     %d ok / %d failed\n", s.registered.Load(), s.regFailed.Load())
	fmt.Printf("IAP verified:   %d ok / %d failed\n", s.verified.Load(), s.verFailed.Load())
	fmt.Printf("  active:       %d\n", s.verActive.Load())
	fmt.Printf("  expired:      %d\n", s.verExpired.Load())
	fmt.Printf("  canceled:     %d\n", s.verCanceled.Load())
	fmt.Printf("  pending:      %d\n", s.verPending.Load())
	rps := float64(s.registered.Load()+s.verified.Load()) / s.duration.Seconds()
	fmt.Printf("Throughput:     %.1f req/s\n", rps)
	fmt.Printf("=======================================\n")
}

// ---- http helpers ----

func doJSON(method, url string, headers map[string]string, body interface{}, dryRun bool) (int, map[string]interface{}, error) {
	if dryRun {
		b, _ := json.MarshalIndent(body, "", "  ")
		fmt.Printf("[DRY] %s %s\n%s\n", method, url, b)
		return 200, nil, nil
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return resp.StatusCode, result, nil
}

// ---- worker ----

func runUser(i int, apiBase, appID, appleMockURL string, dryRun bool, platformMode string, st *stats) {
	platform := platformMode
	var sc scenario

	if platform == "mixed" {
		if rand.Intn(2) == 0 {
			platform = "android"
		} else {
			platform = "ios"
		}
	}

	if platform == "android" {
		sc = googleScenarios[rand.Intn(len(googleScenarios))]
		purchaseToken := fmt.Sprintf("%s_%d_%d", sc.token, i, time.Now().UnixNano())
		receiptJSON, _ := json.Marshal(map[string]string{
			"packageName":   "com.bivex.game",
			"productId":     sc.product,
			"purchaseToken": purchaseToken,
			"type":          "subscription",
		})
		sc.token = string(receiptJSON)
	} else {
		// iOS: create a real receipt via Apple mock POST /subs
		product := appleProducts[rand.Intn(len(appleProducts))]
		sc = scenario{name: "ios_sub", platform: "ios", product: product}
		receiptToken, err := createAppleReceipt(appleMockURL, product, dryRun)
		if err != nil || receiptToken == "" {
			st.regFailed.Add(1)
			log.Printf("[user %d] apple mock create receipt: %v", i, err)
			return
		}
		sc.token = receiptToken
	}

	// 1. Register user
	userID := fmt.Sprintf("loadgen_user_%d_%d", i, rand.Int63())
	deviceID := fmt.Sprintf("device_%d_%d", i, rand.Int63())

	regBody := map[string]interface{}{
		"platform_user_id": userID,
		"device_id":        deviceID,
		"platform":         sc.platform,
		"app_version":      "1.0.0",
		"app_id":           appID,
	}

	status, regResp, err := doJSON("POST", apiBase+"/v1/auth/register",
		map[string]string{"X-App-ID": appID}, regBody, dryRun)
	if err != nil || status >= 400 {
		st.regFailed.Add(1)
		if err != nil {
			log.Printf("[user %d] register error: %v", i, err)
		} else {
			log.Printf("[user %d] register failed %d: %v", i, status, regResp)
		}
		return
	}
	st.registered.Add(1)

	if dryRun {
		return
	}

	// Extract token — response is wrapped in {"data": {...}}
	token, _ := regResp["access_token"].(string)
	if token == "" {
		if data, ok := regResp["data"].(map[string]interface{}); ok {
			token, _ = data["access_token"].(string)
		}
	}
	if token == "" {
		st.regFailed.Add(1)
		log.Printf("[user %d] no access_token in register response", i)
		return
	}

	// 2. Verify IAP
	iapBody := map[string]interface{}{
		"platform":       sc.platform,
		"receipt_data":   sc.token,
		"product_id":     sc.product,
		"transaction_id": fmt.Sprintf("txn_%d_%d", i, rand.Int63()),
	}

	iapStatus, iapResp, err := doJSON("POST", apiBase+"/v1/verify/iap",
		map[string]string{
			"Authorization": "Bearer " + token,
			"X-App-ID":      appID,
		}, iapBody, dryRun)

	if err != nil || iapStatus >= 500 {
		st.verFailed.Add(1)
		if err != nil {
			log.Printf("[user %d] iap verify error: %v", i, err)
		} else {
			log.Printf("[user %d] iap verify failed %d: %v", i, iapStatus, iapResp)
		}
		return
	}

	st.verified.Add(1)

	subStatus, _ := iapResp["status"].(string)
	if subStatus == "" {
		if data, ok := iapResp["data"].(map[string]interface{}); ok {
			subStatus, _ = data["status"].(string)
		}
	}
	switch subStatus {
	case "active":
		st.verActive.Add(1)
	case "expired":
		st.verExpired.Add(1)
	case "canceled", "cancelled":
		st.verCanceled.Add(1)
	case "pending":
		st.verPending.Add(1)
	}

	if os.Getenv("LOADGEN_VERBOSE") != "" {
		fmt.Printf("[user %d] %s %s → %s (http %d)\n", i, sc.platform, sc.name, subStatus, iapStatus)
	}
}

func main() {
	apiBase      := flag.String("api", "http://localhost:8081", "API base URL")
	appID        := flag.String("app-id", "", "App UUID (X-App-ID header, required)")
	appleMock    := flag.String("apple-mock", "http://localhost:9090", "Apple IAP mock server URL")
	users        := flag.Int("users", 100, "Total users to generate")
	workers      := flag.Int("workers", 10, "Concurrent goroutines")
	platformMode := flag.String("platform", "mixed", "ios|android|mixed")
	dryRun       := flag.Bool("dry-run", false, "Print requests without sending")
	flag.Parse()

	if *appID == "" {
		*appID = os.Getenv("APP_ID")
	}
	if *appID == "" {
		log.Fatal("--app-id or APP_ID env var is required")
	}

	rand.Seed(time.Now().UnixNano())
	http.DefaultClient.Timeout = 10 * time.Second

	fmt.Printf("loadgen: %d users, %d workers, platform=%s, api=%s\n",
		*users, *workers, *platformMode, *apiBase)
	if *dryRun {
		fmt.Println("DRY RUN — no requests sent")
	}

	st := &stats{}
	sem := make(chan struct{}, *workers)
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < *users; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			runUser(idx, *apiBase, *appID, *appleMock, *dryRun, *platformMode, st)
		}(i)
	}

	wg.Wait()
	st.duration = time.Since(start)
	st.print()
}
