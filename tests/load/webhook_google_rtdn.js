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
const webhookLatency     = new Trend("webhook_latency_ms");
const statusMatchCount   = new Counter("status_match_total");
const statusMismatchCount = new Counter("status_mismatch_total");

// ─── RTDN notification types with expected resulting subscription status ──────
//
// Statuses stored in DB:
//   active | cancelled | expired | on_hold | grace_period
//
// Notes:
//   - Type 3 CANCELED sets cancelled (auto_renew=false; user still has access
//     until expiry, but we reflect the "will not renew" state as cancelled)
//   - Types 8, 11 are informational; backend logs them but makes no change.
//     We skip them here since there's no observable status assertion to make.
const NOTIFICATION_CASES = [
  { type: 1,  name: "RECOVERED",            expectedStatus: "active"       },
  { type: 2,  name: "RENEWED",              expectedStatus: "active"       },
  { type: 3,  name: "CANCELED",             expectedStatus: "cancelled"    },
  { type: 4,  name: "PURCHASED",            expectedStatus: "active"       },
  { type: 5,  name: "ON_HOLD",              expectedStatus: "on_hold"      },
  { type: 6,  name: "IN_GRACE_PERIOD",      expectedStatus: "grace_period" },
  { type: 7,  name: "RESTARTED",            expectedStatus: "active"       },
  { type: 9,  name: "DEFERRED",             expectedStatus: null           }, // expiry only
  { type: 10, name: "PAUSED",               expectedStatus: "on_hold"      },
  { type: 12, name: "REVOKED",              expectedStatus: "expired"      },
  { type: 13, name: "EXPIRED",              expectedStatus: "expired"      },
];

// ─── k6 options ──────────────────────────────────────────────────────────────
export const options = {
  vus: 5,
  duration: "30s",
  thresholds: {
    // At least 95 % of webhook round-trips complete < 500 ms.
    webhook_latency_ms: ["p(95)<500"],
    // No status mismatches allowed.
    status_mismatch_total: ["count==0"],
    // Sanity: we should have many matches.
    status_match_total: ["count>0"],
    http_req_failed: ["rate<0.05"],
  },
};

// ─── Helpers ─────────────────────────────────────────────────────────────────
function uniqueEmail() {
  return `rtdn_${__VU}_${Date.now()}@test.invalid`;
}

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
    { headers: { "Authorization": `Bearer ${token}` } }
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
      backendURL:       BACKEND,
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
    "iap valid=true":  (r) => r.json("valid") === true,
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
    }

    if (tc.expectedStatus === null) {
      // Informational type — skip status assertion.
      continue;
    }

    // Give the asynq worker a moment to process.
    sleep(0.3);

    // 4. Fetch current subscription status.
    const subRes = getSubscription(jwtToken);
    const gotStatus = subRes.json("status") || subRes.json("subscription.status") || "";
    check(subRes, { [`sub status 2xx after ${tc.name}`]: (r) => r.status >= 200 && r.status < 300 });

    // 5. Assert.
    if (gotStatus === tc.expectedStatus) {
      statusMatchCount.add(1);
    } else {
      statusMismatchCount.add(1);
      console.error(
        `[VU${__VU}] type=${tc.type} (${tc.name}): expected="${tc.expectedStatus}", got="${gotStatus}" | sub=${subRes.body}`
      );
    }
  }

  sleep(0.2);
}
