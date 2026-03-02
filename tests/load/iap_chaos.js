/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-03-02 07:18
 * Last Updated: 2026-03-02 07:18
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

/**
 * IAP Chaos Load Test — 20-second realistic scenario mix
 *
 * Simulates concurrent Android IAP verifications through the google-billing-mock
 * server with a production-realistic scenario distribution.
 *
 * Scenario weights mirror real-world IAP analytics:
 *   50% valid_active       → 200  active auto-renewing subscription
 *   15% canceled_active    → 200  canceled but still within paid period (VALID — must accept)
 *   15% expired            → 422  churned user
 *    8% canceled           → 422  voluntary cancel + already expired
 *    5% pending            → 422  billing issue (payment_state=0)
 *    4% invalid            → 410→422  replay / fraud attempt
 *    3% product_valid      → 422  one-time purchase token on subscriptions endpoint
 *
 * Note on go-iap library:
 *   go-iap/playstore wraps androidpublisher/v3. Our GoogleVerifier uses
 *   androidpublisher/v3 directly (with option.WithEndpoint for mock testing),
 *   which allows custom base URL redirection that go-iap/playstore doesn't expose.
 *   go-iap is only used for iOS/Apple receipt verification (AppleVerifier).
 *   IAPAdapter.VerifyReceipt() is intentionally broken — only platform-specific
 *   adapters (AppleVerifierAdapter, AndroidVerifierAdapter) are valid.
 *
 * Usage:
 *   k6 run tests/load/iap_chaos.js
 *   k6 run tests/load/iap_chaos.js --env BASE_URL=http://localhost:8081
 *   k6 run tests/load/iap_chaos.js --env VUS=100
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// ──────────────────────────────────────────────
// Config
// ──────────────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';
const PACKAGE   = __ENV.PACKAGE  || 'com.yourapp';
const PRODUCT   = __ENV.PRODUCT  || 'com.yourapp.premium_monthly';
const MAX_VUS   = parseInt(__ENV.VUS || '100', 10);

// ──────────────────────────────────────────────
// Realistic production scenario distribution
// ──────────────────────────────────────────────
const SCENARIOS = [
  // [prefix, expectedStatus, label, weight]
  // 50% active — most users have valid, auto-renewing subscriptions
  { prefix: 'valid_active_',    expect: 200, label: 'active',           weight: 50 },
  // 15% canceled_active — user canceled but still within paid period (MUST accept)
  { prefix: 'canceled_active_', expect: 200, label: 'canceled_paid',    weight: 15 },
  // 15% expired — churned users (payment_state=0 or expiry in past)
  { prefix: 'expired_',         expect: 422, label: 'expired',          weight: 15 },
  //  8% canceled+expired — voluntary cancel AND already past expiry
  { prefix: 'canceled_',        expect: 422, label: 'canceled_expired', weight:  8 },
  //  5% pending — billing issue, payment not yet received (payment_state=0)
  { prefix: 'pending_',         expect: 422, label: 'pending',          weight:  5 },
  //  4% invalid — replay attack / stolen token (mock returns 410, backend maps → 422)
  { prefix: 'invalid_',         expect: 422, label: 'invalid_token',    weight:  4 },
  //  3% product_valid — one-time purchase token mistakenly sent to subscriptions endpoint
  { prefix: 'product_valid_',   expect: 422, label: 'product_sub_404',  weight:  3 },
];

// Build weighted random picker
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
    http_req_failed:                              ['rate<0.01'],
    'http_req_duration{scenario:register}':       ['p(95)<2000', 'p(99)<4000'],
    'http_req_duration{scenario:verify_iap}':     ['p(95)<3000', 'p(99)<6000'],
    'iap_valid_success_rate':                     ['rate>0.95'],
    'iap_invalid_reject_rate':                    ['rate>0.95'],
  },
};

// ──────────────────────────────────────────────
// Custom metrics
// ──────────────────────────────────────────────
const iapValidSuccessRate  = new Rate('iap_valid_success_rate');
const iapInvalidRejectRate = new Rate('iap_invalid_reject_rate');
const iapTrend             = new Trend('iap_verify_duration_ms', true);
const scenarioCounter      = new Counter('iap_scenario_calls');
const registerErrors       = new Counter('register_errors');

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────
const JSON_HEADERS = { 'Content-Type': 'application/json' };

function authHeaders(token) {
  return { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` };
}

function register() {
  const id = uuidv4().replace(/-/g, '').slice(0, 20);
  const payload = JSON.stringify({
    platform_user_id: `k6_${id}`,
    device_id:        `dev_${id}`,
    platform:         'android',
    app_version:      '1.0.0',
    // no email — exercises partial unique index for empty-email users
  });

  const res = http.post(`${BASE_URL}/v1/auth/register`, payload, {
    headers: JSON_HEADERS,
    tags:    { scenario: 'register' },
  });

  const ok = check(res, {
    'register: status 2xx': r => r.status >= 200 && r.status < 300,
    'register: has access_token': r => {
      try { return !!JSON.parse(r.body).data.access_token; } catch { return false; }
    },
  });

  if (!ok || res.status < 200 || res.status >= 300) {
    registerErrors.add(1);
    return null;
  }

  return JSON.parse(res.body).data.access_token;
}

function verifyIAP(token, scenario) {
  const purchaseToken = `${scenario.prefix}${uuidv4().replace(/-/g, '').slice(0, 12)}`;

  const receiptData = JSON.stringify({
    packageName:   PACKAGE,
    productId:     PRODUCT,
    purchaseToken: purchaseToken,
    type:          'subscription',
  });

  const body = JSON.stringify({
    platform:     'android',
    product_id:   PRODUCT,
    receipt_data: receiptData,
  });

  const start = Date.now();
  const res = http.post(`${BASE_URL}/v1/verify/iap`, body, {
    headers: authHeaders(token),
    responseCallback: http.expectedStatuses({ min: 200, max: 299 }, 400, 410, 422),
    tags:    { scenario: 'verify_iap', iap_label: scenario.label },
  });
  iapTrend.add(Date.now() - start);
  scenarioCounter.add(1, { label: scenario.label });

  let parsed = null;
  try { parsed = JSON.parse(res.body); } catch {}

  if (scenario.expect === 200) {
    const valid = check(res, {
      [`[${scenario.label}] status 200`]:       r => r.status === 200,
      [`[${scenario.label}] has subscription`]: r => parsed && parsed.data && parsed.data.subscription_id,
    });
    iapValidSuccessRate.add(valid ? 1 : 0);
  } else {
    const rejected = check(res, {
      [`[${scenario.label}] rejected (4xx)`]: r => r.status >= 400 && r.status < 500,
      [`[${scenario.label}] has error field`]:r => parsed && parsed.error,
    });
    iapInvalidRejectRate.add(rejected ? 1 : 0);
  }
}

// ──────────────────────────────────────────────
// Main VU iteration
// ──────────────────────────────────────────────
export default function () {
  let token;
  group('register', () => {
    token = register();
  });

  if (!token) {
    sleep(1);
    return;
  }

  const scenario = pickScenario();

  group('verify_iap', () => {
    verifyIAP(token, scenario);
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
║          IAP Chaos Load Test — 20s Realistic Summary         ║
╚══════════════════════════════════════════════════════════════╝

  Total IAP calls     : ${cnt('iap_scenario_calls')}
  Register errors     : ${cnt('register_errors')}
  HTTP error rate     : ${rate('http_req_failed')}

  Scenario mix (target): 50% active · 15% canceled_paid · 15% expired
                          8% canceled_expired · 5% pending · 4% invalid · 3% product_sub

  ┌─ Latency ──────────────────────────────────────────────┐
  │  Register  p95: ${p95('http_req_duration{scenario:register}')}    p99: ${p99('http_req_duration{scenario:register}')}
  │  Verify    p95: ${p95('http_req_duration{scenario:verify_iap}')}    p99: ${p99('http_req_duration{scenario:verify_iap}')}
  │  IAP trend p95: ${p95('iap_verify_duration_ms')}    p99: ${p99('iap_verify_duration_ms')}
  └────────────────────────────────────────────────────────┘

  ┌─ Business metrics ─────────────────────────────────────┐
  │  Valid receipt success rate : ${rate('iap_valid_success_rate')}
  │  Invalid receipt reject rate: ${rate('iap_invalid_reject_rate')}
  └────────────────────────────────────────────────────────┘
`;

  return {
    stdout: report,
    'tests/load/results/iap_chaos_summary.json': JSON.stringify(data, null, 2),
  };
}

