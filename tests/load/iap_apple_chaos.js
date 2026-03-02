/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-03-02
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

/**
 * Apple IAP Chaos Load Test — 20-second realistic scenario mix
 *
 * Simulates concurrent iOS IAP verifications through the apple-of-my-iap mock
 * server with a production-realistic scenario distribution.
 *
 * How it works:
 *   1. Each VU registers a user → gets JWT
 *   2. Picks a scenario from the weighted pool
 *   3. If scenario requires a real receipt:
 *      a. Creates a subscription on the Apple mock (POST /subs)
 *      b. Optionally mutates state (cancel / expire / refund)
 *      c. Sends the receipt token to the backend (POST /v1/verify/iap platform=apple)
 *   4. For invalid scenarios: sends garbage receipt-data → backend gets 21002 from mock → 422
 *
 * Scenario weights mirror real-world iOS IAP analytics:
 *   50% active           → 200  valid auto-renewing subscription
 *   15% canceled_active  → 200  canceled but still within paid period (MUST accept)
 *   15% invalid_receipt  → 422  garbage/unknown receipt (Apple status 21002)
 *    8% expired          → 422  subscription expired (Apple status 21006)
 *    7% canceled_expired → 422  canceled AND already expired
 *    5% refunded         → 422  transaction refunded
 *
 * Apple mock: tests/apple-of-my-iap  (default port 9090)
 *   Start: cd tests/apple-of-my-iap && go build -o bin/server ./cmd/server && ./bin/server
 *
 * Usage:
 *   k6 run tests/load/iap_apple_chaos.js
 *   k6 run tests/load/iap_apple_chaos.js \
 *       --env BASE_URL=http://localhost:8081 \
 *       --env APPLE_MOCK_URL=http://localhost:9090 \
 *       --env PRODUCT_ID=com.example.unlimited.1mo \
 *       --env VUS=50
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// ──────────────────────────────────────────────
// Config
// ──────────────────────────────────────────────
const BASE_URL        = __ENV.BASE_URL        || 'http://localhost:8081';
const APPLE_MOCK_URL  = __ENV.APPLE_MOCK_URL  || 'http://localhost:9090';
const PRODUCT_ID      = __ENV.PRODUCT_ID      || 'com.example.unlimited.1mo';
const MAX_VUS         = parseInt(__ENV.VUS    || '50', 10);

// ──────────────────────────────────────────────
// Apple IAP status codes (from apple-of-my-iap/iap-api/status.go)
// ──────────────────────────────────────────────
const APPLE_STATUS = {
  VALID:    0,
  EXPIRED:  21006,
  BAD:      21002,
  UNAUTH:   21003,
};

// ──────────────────────────────────────────────
// Scenario distribution (weights sum to 100)
// ──────────────────────────────────────────────
const SCENARIOS = [
  { label: 'active',           expect: 200, needsMock: true,  mutate: null,               weight: 50 },
  { label: 'canceled_active',  expect: 200, needsMock: true,  mutate: 'cancel',            weight: 15 },
  { label: 'invalid_receipt',  expect: 422, needsMock: false, mutate: null,               weight: 15 },
  { label: 'expired',          expect: 422, needsMock: true,  mutate: 'expire',            weight:  8 },
  { label: 'canceled_expired', expect: 422, needsMock: true,  mutate: 'cancel_then_expire',weight:  7 },
  { label: 'refunded',         expect: 422, needsMock: true,  mutate: 'refund',            weight:  5 },
];

const WEIGHTED_POOL = [];
SCENARIOS.forEach(s => {
  for (let i = 0; i < s.weight; i++) WEIGHTED_POOL.push(s);
});

function pickScenario() {
  return WEIGHTED_POOL[Math.floor(Math.random() * WEIGHTED_POOL.length)];
}

// ──────────────────────────────────────────────
// k6 options — 20-second realistic smoke test
// ──────────────────────────────────────────────
export const options = {
  stages: [
    { duration: '5s',  target: MAX_VUS },  // ramp up
    { duration: '10s', target: MAX_VUS },  // steady state
    { duration: '5s',  target: 0       },  // ramp down
  ],
  thresholds: {
    http_req_failed:                               ['rate<0.02'],
    'http_req_duration{scenario:register}':        ['p(95)<2000', 'p(99)<4000'],
    'http_req_duration{scenario:verify_iap}':      ['p(95)<3000', 'p(99)<6000'],
    'http_req_duration{scenario:mock_create_sub}': ['p(95)<500',  'p(99)<1000'],
    'apple_iap_valid_success_rate':                ['rate>0.95'],
    'apple_iap_invalid_reject_rate':               ['rate>0.95'],
  },
};

// ──────────────────────────────────────────────
// Custom metrics
// ──────────────────────────────────────────────
const appleValidRate    = new Rate('apple_iap_valid_success_rate');
const appleInvalidRate  = new Rate('apple_iap_invalid_reject_rate');
const iapTrend          = new Trend('apple_iap_verify_duration_ms', true);
const scenarioCounter   = new Counter('apple_iap_scenario_calls');
const registerErrors    = new Counter('apple_register_errors');
const mockErrors        = new Counter('apple_mock_errors');

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────
const JSON_HEADERS = { 'Content-Type': 'application/json' };

function authHeaders(token) {
  return { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` };
}

// Register a new iOS user and return JWT
function register() {
  const id = uuidv4().replace(/-/g, '').slice(0, 20);
  const res = http.post(`${BASE_URL}/v1/auth/register`, JSON.stringify({
    platform_user_id: `k6_ios_${id}`,
    device_id:        `ios_dev_${id}`,
    platform:         'ios',
    app_version:      '2.0.0',
  }), { headers: JSON_HEADERS, tags: { scenario: 'register' } });

  const ok = check(res, {
    'register: 2xx':              r => r.status >= 200 && r.status < 300,
    'register: has access_token': r => {
      try { return !!JSON.parse(r.body).data.access_token; } catch { return false; }
    },
  });

  if (!ok) { registerErrors.add(1); return null; }
  try { return JSON.parse(res.body).data.access_token; } catch { return null; }
}

// Create a subscription on the Apple mock → returns receiptToken or null
function mockCreateSub() {
  const res = http.post(
    `${APPLE_MOCK_URL}/subs`,
    JSON.stringify({ productId: PRODUCT_ID }),
    { headers: JSON_HEADERS, tags: { scenario: 'mock_create_sub' } },
  );

  const ok = check(res, {
    'mock create sub: 200': r => r.status === 200,
    'mock create sub: has token': r => {
      try {
        const b = JSON.parse(r.body);
        return !!(b.latestReceiptInfo && b.latestReceiptInfo.length > 0);
      } catch { return false; }
    },
  });

  if (!ok) { mockErrors.add(1); return null; }

  try {
    // latestReceiptInfo[0].transactionId is the receipt token
    const b = JSON.parse(res.body);
    return b.latestReceiptInfo[0].transactionId;
  } catch { mockErrors.add(1); return null; }
}

// Mutate subscription state on the Apple mock
function mockMutate(receiptToken, mutate) {
  switch (mutate) {
    case 'cancel': {
      const res = http.post(
        `${APPLE_MOCK_URL}/subs/${receiptToken}/cancel`,
        null,
        { headers: JSON_HEADERS, tags: { scenario: 'mock_mutate' } },
      );
      check(res, { 'mock cancel: ok': r => r.status === 200 });
      break;
    }
    case 'expire': {
      // Set Apple status to 21006 (SubscriptionExpired)
      const res = http.post(
        `${APPLE_MOCK_URL}/subs/${receiptToken}`,
        JSON.stringify({ status: APPLE_STATUS.EXPIRED }),
        { headers: JSON_HEADERS, tags: { scenario: 'mock_mutate' } },
      );
      check(res, { 'mock expire: ok': r => r.status === 200 });
      break;
    }
    case 'cancel_then_expire': {
      http.post(`${APPLE_MOCK_URL}/subs/${receiptToken}/cancel`, null, {
        headers: JSON_HEADERS, tags: { scenario: 'mock_mutate' },
      });
      http.post(
        `${APPLE_MOCK_URL}/subs/${receiptToken}`,
        JSON.stringify({ status: APPLE_STATUS.EXPIRED }),
        { headers: JSON_HEADERS, tags: { scenario: 'mock_mutate' } },
      );
      break;
    }
    case 'refund': {
      // Refund first transaction: renew to get a transaction ID, then refund
      const subRes = http.post(
        `${APPLE_MOCK_URL}/verifyReceipt`,
        JSON.stringify({ 'receipt-data': receiptToken }),
        { headers: JSON_HEADERS, tags: { scenario: 'mock_mutate' } },
      );
      try {
        const b = JSON.parse(subRes.body);
        const txId = b.latest_receipt_info && b.latest_receipt_info[0]
          ? b.latest_receipt_info[0].transaction_id
          : null;
        if (txId) {
          http.post(
            `${APPLE_MOCK_URL}/subs/${receiptToken}/refund/${txId}`,
            null,
            { headers: JSON_HEADERS, tags: { scenario: 'mock_mutate' } },
          );
        }
      } catch { /* ignore */ }
      break;
    }
  }
}

// Send Apple receipt to backend for verification
function verifyIAP(jwtToken, receiptData, scenario) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/v1/verify/iap`,
    JSON.stringify({
      platform:     'apple',
      product_id:   PRODUCT_ID,
      receipt_data: receiptData,
    }),
    {
      headers:          authHeaders(jwtToken),
      responseCallback: http.expectedStatuses({ min: 200, max: 299 }, 400, 422),
      tags:             { scenario: 'verify_iap', iap_label: scenario.label },
    },
  );
  iapTrend.add(Date.now() - start);
  scenarioCounter.add(1, { label: scenario.label });

  let parsed = null;
  try { parsed = JSON.parse(res.body); } catch {}

  if (scenario.expect === 200) {
    const valid = check(res, {
      [`[${scenario.label}] status 200`]:       r => r.status === 200,
      [`[${scenario.label}] has subscription`]: r => parsed && parsed.data && parsed.data.subscription_id,
    });
    appleValidRate.add(valid ? 1 : 0);
  } else {
    const rejected = check(res, {
      [`[${scenario.label}] rejected 4xx`]: r => r.status >= 400 && r.status < 500,
      [`[${scenario.label}] has error`]:    r => parsed && (parsed.error || parsed.message),
    });
    appleInvalidRate.add(rejected ? 1 : 0);
  }
}

// ──────────────────────────────────────────────
// setup() — runs once before all VUs
// ──────────────────────────────────────────────
export function setup() {
  // Verify backend is up
  const backendHealth = http.get(`${BASE_URL}/health`);
  check(backendHealth, { 'backend health': r => r.status === 200 });

  // Verify Apple mock is up
  const mockHealth = http.get(`${APPLE_MOCK_URL}/`);
  check(mockHealth, { 'apple mock is up': r => r.status === 200 });

  // Ensure the plan exists in the mock (clear & seed plans if mock supports it)
  // The mock loads plans from tmp/plans.json at startup — no HTTP seeding needed.
}

// ──────────────────────────────────────────────
// Main VU iteration
// ──────────────────────────────────────────────
export default function () {
  let jwtToken;
  group('register', () => {
    jwtToken = register();
  });

  if (!jwtToken) { sleep(1); return; }

  const scenario = pickScenario();
  let receiptData;

  if (scenario.needsMock) {
    // Create a real subscription on the Apple mock
    let receiptToken;
    group('mock_prepare', () => {
      receiptToken = mockCreateSub();
      if (receiptToken && scenario.mutate) {
        mockMutate(receiptToken, scenario.mutate);
      }
    });

    if (!receiptToken) { sleep(0.5); return; }
    receiptData = receiptToken;
  } else {
    // Garbage receipt — Apple mock returns 21002 → backend returns 422
    receiptData = `invalid_receipt_${uuidv4().replace(/-/g, '').slice(0, 16)}`;
  }

  group('verify_iap', () => {
    verifyIAP(jwtToken, receiptData, scenario);
  });

  sleep(Math.random() * 0.3 + 0.05);
}

// ──────────────────────────────────────────────
// Summary report
// ──────────────────────────────────────────────
export function handleSummary(data) {
  const m = data.metrics;

  const fmt  = (v, unit = '') => v !== undefined ? `${v.toFixed ? v.toFixed(2) : v}${unit}` : 'n/a';
  const rate = key => m[key] ? fmt(m[key].values.rate * 100, '%') : 'n/a';
  const p95  = key => m[key] ? fmt(m[key].values['p(95)'], 'ms') : 'n/a';
  const p99  = key => m[key] ? fmt(m[key].values['p(99)'], 'ms') : 'n/a';
  const cnt  = key => m[key] ? m[key].values.count : 0;

  const report = `
╔══════════════════════════════════════════════════════════════╗
║       Apple IAP Chaos Load Test — 20s Realistic Summary      ║
╚══════════════════════════════════════════════════════════════╝

  Apple mock        : ${APPLE_MOCK_URL}
  Backend           : ${BASE_URL}
  Product ID        : ${PRODUCT_ID}

  Total IAP calls   : ${cnt('apple_iap_scenario_calls')}
  Register errors   : ${cnt('apple_register_errors')}
  Mock errors       : ${cnt('apple_mock_errors')}
  HTTP error rate   : ${rate('http_req_failed')}

  Scenario mix (target):
    50% active · 15% canceled_active · 15% invalid_receipt
     8% expired · 7% canceled_expired · 5% refunded

  ┌─ Latency ──────────────────────────────────────────────────┐
  │  Register       p95: ${p95('http_req_duration{scenario:register}')}     p99: ${p99('http_req_duration{scenario:register}')}
  │  Mock create    p95: ${p95('http_req_duration{scenario:mock_create_sub}')}     p99: ${p99('http_req_duration{scenario:mock_create_sub}')}
  │  Verify IAP     p95: ${p95('http_req_duration{scenario:verify_iap}')}     p99: ${p99('http_req_duration{scenario:verify_iap}')}
  │  IAP trend      p95: ${p95('apple_iap_verify_duration_ms')}     p99: ${p99('apple_iap_verify_duration_ms')}
  └────────────────────────────────────────────────────────────┘

  ┌─ Business metrics ─────────────────────────────────────────┐
  │  Valid receipt success rate  : ${rate('apple_iap_valid_success_rate')}
  │  Invalid receipt reject rate : ${rate('apple_iap_invalid_reject_rate')}
  └────────────────────────────────────────────────────────────┘
`;

  return {
    stdout: report,
    'tests/load/results/iap_apple_chaos_summary.json': JSON.stringify(data, null, 2),
  };
}
