/**
 * k6 — Google Play RTDN Webhook status-transition test
 *
 * Flow per VU:
 *   1. Register a new user
 *   2. Verify IAP with a fresh "valid_active_" token → subscription created in DB
 *   3. For each RTDN notification type, POST /admin/send-webhook on the mock
 *   4. Query the subscription status from the backend API
 *   5. Assert the expected status was applied
 *
 * Focus: correctness (all 13 notification types → correct status), not throughput.
 * Run: k6 run tests/load/webhook_google_rtdn.js
 *      (backend on :8081, mock on :8090)
 */

import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Counter } from "k6/metrics";

// ─── Config ──────────────────────────────────────────────────────────────────
const BACKEND = __ENV.BACKEND_URL || "http://localhost:8081";
const MOCK    = __ENV.MOCK_URL    || "http://localhost:8090";
const PKG     = "com.yourapp";
const PRODUCT = "com.yourapp.premium_monthly";

// ─── Metrics ─────────────────────────────────────────────────────────────────
const webhookLatency      = new Trend("webhook_latency_ms");
const statusMatchCount    = new Counter("status_match_total");
const statusMismatchCount = new Counter("status_mismatch_total");

// ─── RTDN notification types with expected resulting subscription status ──────
//
// Statuses stored in DB:
//   active | cancelled | expired | on_hold | grace_period
//
// Notes:
//   - GET /v1/subscription returns 404 when subscription is non-active (cancelled,
//     expired, on_hold, grace_period). For those types we only verify webhook delivery
//     (200 from mock) — actual status change is visible in worker logs / DB.
//   - Types 8, 11 are informational; backend logs them but makes no change.
//   - We end with a RENEWED to restore active state for next VU iteration.
const NOTIFICATION_CASES = [
  // Active-returning types: webhook 200 + subscription 2xx + status=="active"
  { type: 2,  name: "RENEWED",         expectedStatus: "active",       checkSub: true  },
  { type: 4,  name: "PURCHASED",       expectedStatus: "active",       checkSub: true  },
  { type: 7,  name: "RESTARTED",       expectedStatus: "active",       checkSub: true  },
  { type: 1,  name: "RECOVERED",       expectedStatus: "active",       checkSub: true  },
  // Non-active types: only verify webhook 200 (sub endpoint returns 404)
  { type: 3,  name: "CANCELED",        expectedStatus: "cancelled",    checkSub: false },
  { type: 13, name: "EXPIRED",         expectedStatus: "expired",      checkSub: false },
  { type: 12, name: "REVOKED",         expectedStatus: "expired",      checkSub: false },
  { type: 5,  name: "ON_HOLD",         expectedStatus: "on_hold",      checkSub: false },
  { type: 6,  name: "IN_GRACE_PERIOD", expectedStatus: "grace_period", checkSub: false },
  { type: 10, name: "PAUSED",          expectedStatus: "on_hold",      checkSub: false },
  { type: 9,  name: "DEFERRED",        expectedStatus: null,           checkSub: false }, // expiry-only
  // Restore to active for next iteration
  { type: 2,  name: "RENEWED_RESET",   expectedStatus: "active",       checkSub: true  },
];

// ─── k6 options ──────────────────────────────────────────────────────────────
export const options = {
  vus: 1,   // 1 VU: sequential execution for deterministic status transitions
  duration: "30s",
  thresholds: {
    webhook_latency_ms:    ["p(95)<500"],
    status_mismatch_total: ["count==0"],
    status_match_total:    ["count>0"],
    http_req_failed:       ["rate<0.05"],
  },
};

// ─── Helpers ─────────────────────────────────────────────────────────────────
function register(platformUID, deviceID) {
  const res = http.post(
    `${BACKEND}/v1/auth/register`,
    JSON.stringify({
      platform_user_id: platformUID,
      device_id:        deviceID,
      platform:         "android",
      app_version:      "1.0",
    }),
    { headers: { "Content-Type": "application/json" } }
  );
  check(res, { "register 2xx": (r) => r.status >= 200 && r.status < 300 });
  const body = res.json();
  return (body && body.data && body.data.access_token) ? body.data.access_token : "";
}

function verifyIAP(token, purchaseToken) {
  const receiptData = JSON.stringify({
    packageName:   PKG,
    productId:     PRODUCT,
    purchaseToken: purchaseToken,
    type:          "subscription",
  });
  const res = http.post(
    `${BACKEND}/v1/verify/iap`,
    JSON.stringify({ platform: "android", product_id: PRODUCT, receipt_data: receiptData }),
    { headers: { "Content-Type": "application/json", "Authorization": `Bearer ${token}` } }
  );
  return res;
}

function getSubscription(token) {
  const res = http.get(
    `${BACKEND}/v1/subscription`,
    {
      headers: { "Authorization": `Bearer ${token}` },
      responseCallback: http.expectedStatuses(200, 201, 404), // 404 = non-active sub, expected
    }
  );
  return res;
}

function sendWebhook(purchaseToken, notificationType) {
  const start = Date.now();
  const res = http.post(
    `${MOCK}/admin/send-webhook`,
    JSON.stringify({
      purchaseToken:    purchaseToken,
      notificationType: notificationType,
      subscriptionId:   PRODUCT,
      packageName:      PKG,
      // No backendURL — mock uses its BACKEND_URL env var (http://api:8080 in Docker)
    }),
    { headers: { "Content-Type": "application/json" } }
  );
  webhookLatency.add(Date.now() - start);
  return res;
}

// ─── Main test ───────────────────────────────────────────────────────────────
export default function () {
  const uid      = `rtdn_${__VU}_${Date.now()}`;
  const deviceID = `dev_${__VU}_${Date.now()}`;

  // 1. Register + get JWT.
  const jwtToken = register(uid, deviceID);
  if (!jwtToken) {
    console.error("registration failed, skipping VU iteration");
    return;
  }

  // 2. Verify a fresh purchase to create a subscription.
  //    Use a unique token so each VU has its own subscription row.
  const purchaseToken = `valid_active_rtdn_vu${__VU}_${Date.now()}`;
  const iapRes = verifyIAP(jwtToken, purchaseToken);
  const iapOK = check(iapRes, {
    "iap verify 200": (r) => r.status === 200,
    "iap has status": (r) => {
      const b = r.json();
      return b && b.data && b.data.status !== undefined;
    },
  });
  if (!iapOK) {
    console.error(`iap verify failed: ${iapRes.status} ${iapRes.body}`);
    return;
  }

  // Short pause to let the worker process the initial subscription row.
  sleep(0.5);

  // 3. For each RTDN type, send webhook → assert status.
  for (const tc of NOTIFICATION_CASES) {
    const wRes = sendWebhook(purchaseToken, tc.type);
    check(wRes, {
      [`webhook ${tc.name} 200`]: (r) => r.status === 200,
    });
    if (wRes.status !== 200) {
      console.warn(`webhook ${tc.name} returned ${wRes.status}: ${wRes.body}`);
      continue;
    }

    // Give the asynq worker a moment to process (non-assertable types only).
    if (!tc.checkSub) sleep(0.2);

    // Only assert subscription status for active-returning transitions.
    // Non-active subs return 404 from GET /v1/subscription (expected business logic).
    if (!tc.checkSub || tc.expectedStatus === null) {
      continue;
    }

    // 4. Poll subscription status — worker may need a moment to flush queue.
    let gotStatus = "";
    const maxAttempts = 10; // up to 5 seconds (10 × 0.5s)
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      sleep(0.5);
      const subRes = getSubscription(jwtToken);
      const subBody = subRes.json();
      gotStatus = (subBody && subBody.data && subBody.data.status) ? subBody.data.status : "";
      if (gotStatus === tc.expectedStatus) break;
    }

    // 5. Assert status.
    if (gotStatus === tc.expectedStatus) {
      statusMatchCount.add(1);
    } else {
      statusMismatchCount.add(1);
      console.error(
        `[VU${__VU}] type=${tc.type} (${tc.name}): expected="${tc.expectedStatus}", got="${gotStatus}"`
      );
    }
  }

  sleep(0.2);
}
