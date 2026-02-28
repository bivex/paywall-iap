import http from 'k6/http';
import { check, sleep } from 'k6';
import { RateLimiter } from 'k6/x/http';

// Growth Layer Bandit API Load Test
//
// Tests bandit assignment API under load
// Target: P99 latency < 50ms for assignments
//
// Usage: k6 run tests/load/bandit_bench.js
//   --env VUS=100
//   --env DURATION=5m
//   --env BASE_URL=http://localhost:8080

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const VUS = parseInt(__ENV.VUS || '100');
const DURATION = __ENV.DURATION || '5m';

// Test configuration
export const options = {
  vus: VUS,
  duration: DURATION,
  thresholds: {
    // Bandit assignment latency requirements
    'http_req_duration{type:bandit_assign}': ['p(95)<100', 'p(99)<200'],
    'http_req_duration{type:bandit_reward}': ['p(95)<150', 'p(99)<300'],

    // Error rates
    'http_req_failed{type:bandit_assign}': ['rate<0.01'], // <1% failure
    'http_req_failed{type:bandit_reward}': ['rate<0.01'],

    // Throughput
    'http_reqs{type:bandit_assign}': ['rate>100'], // 100 req/sec minimum
  },
};

const EXPERIMENT_ID = '00000000-0000-0000-0000-000000000001'; // Test experiment ID

// Track metrics
const metrics = {
  assignments: new Map([
    ['total', 0],
    ['success', 0],
    ['error', 0],
    ['cache_hit', 0],
    ['cache_miss', 0],
  ]),
  rewards: new Map([
    ['total', 0],
    ['success', 0],
    ['error', 0],
  ]),
};

// Setup function - runs once before test
export function setup() {
  // Verify API is accessible
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'API health check': (r) => r.status === 200,
  });

  // Verify bandit service is accessible
  const banditHealthRes = http.get(`${BASE_URL}/v1/bandit/health`);
  check(banditHealthRes, {
    'Bandit service health check': (r) => r.status === 200,
  });
}

// Main test - VU logic
export default function() {
  const scenario = Math.random();

  if (scenario < 0.7) {
    // 70%: Assignment flow
    testBanditAssign();
  } else if (scenario < 0.9) {
    // 20%: Reward tracking
    testBanditReward();
  } else {
    // 10%: Statistics query
    testBanditStatistics();
  }

  // Small think time between iterations
  sleep(Math.random() * 100);
}

function testBanditAssign() {
  const start = Date.now();

  // Generate random user ID
  const userId = `user_${Math.random().toString(36).substring(7)}_${__VU}`;

  const payload = {
    experiment_id: EXPERIMENT_ID,
    user_id: userId,
  };

  const params = {
    headers: { 'Content-Type': 'application/json' },
    tags: { type: 'bandit_assign' },
  };

  const res = http.post(
    `${BASE_URL}/v1/bandit/assign`,
    JSON.stringify(payload),
    params
  );

  const duration = Date.now() - start;

  // Track metrics
  metrics.assignments.set('total', metrics.assignments.get('total') + 1);

  if (check(res, { 'Assignment successful': (r) => r.status === 200 })) {
    metrics.assignments.set('success', metrics.assignments.get('success') + 1);

    // Check if assignment was from cache (faster response)
    if (duration < 20) {
      metrics.assignments.set('cache_hit', metrics.assignments.get('cache_hit') + 1);
    } else {
      metrics.assignments.set('cache_miss', metrics.assignments.get('cache_miss') + 1);
    }
  } else {
    metrics.assignments.set('error', metrics.assignments.get('error') + 1);
  }
}

function testBanditReward() {
  // Use previously assigned arm (simulated)
  const armId = `arm_${Math.floor(Math.random() * 3)}`;
  const userId = `user_${__VU}_${Date.now()}`;

  const payload = {
    experiment_id: EXPERIMENT_ID,
    arm_id: armId,
    user_id: userId,
    reward: Math.random() > 0.3 ? 9.99 : 0.0, // 70% conversion rate
  };

  const params = {
    headers: { 'Content-Type': 'application/json' },
    tags: { type: 'bandit_reward' },
  };

  const res = http.post(
    `${BASE_URL}/v1/bandit/reward`,
    JSON.stringify(payload),
    params
  );

  metrics.rewards.set('total', metrics.rewards.get('total') + 1);

  if (check(res, { 'Reward recorded': (r) => r.status === 200 })) {
    metrics.rewards.set('success', metrics.rewards.get('success') + 1);
  } else {
    metrics.rewards.set('error', metrics.rewards.get('error') + 1);
  }
}

function testBanditStatistics() {
  const params = {
    tags: { type: 'bandit_stats' },
  };

  const res = http.get(
    `${BASE_URL}/v1/bandit/statistics?experiment_id=${EXPERIMENT_ID}&win_probs=true`,
    params
  );

  check(res, {
    'Statistics retrieved': (r) => r.status === 200,
  });
}

// Teardown function - runs once after test
export function teardown() {
  // Log final metrics
  console.log('\n=== Bandit Load Test Results ===');
  console.log(`Assignments: ${metrics.assignments.get('total')} total`);
  console.log(`  Success: ${metrics.assignments.get('success')} (${(metrics.assignments.get('success') / metrics.assignments.get('total') * 100).toFixed(1)}%)`);
  console.log(`  Errors: ${metrics.assignments.get('error')} (${(metrics.assignments.get('error') / metrics.assignments.get('total') * 100).toFixed(2)}%)`);
  console.log(`  Cache hits: ${metrics.assignments.get('cache_hit')} (${(metrics.assignments.get('cache_hit') / metrics.assignments.get('success') * 100).toFixed(1)}%)`);
  console.log(`\nRewards: ${metrics.rewards.get('total')} total`);
  console.log(`  Success: ${metrics.rewards.get('success')} (${(metrics.rewards.get('success') / metrics.rewards.get('total') * 100).toFixed(1)}%)`);
  console.log(`  Errors: ${metrics.rewards.get('error')} (${(metrics.rewards.get('error') / metrics.rewards.get('total') * 100).toFixed(2)}%)`);
}

// Stress test - higher load with spike detection
export function stressTest() {
  // Spike test: 10x normal load for 30 seconds
  const spikeDuration = 30;
  const normalRate = 10;
  const spikeRate = 100;

  console.log(`Starting stress test: ${normalRate} → ${spikeRate} → ${normalRate}`);

  // Phase 1: Normal load
  console.log(`Phase 1: Normal load (${normalRate} RPS)`);
  executeLoad(normalRate, 60);

  // Phase 2: Spike
  console.log(`Phase 2: Spike (${spikeRate} RPS)`);
  executeLoad(spikeRate, spikeDuration);

  // Phase 3: Recovery
  console.log(`Phase 3: Recovery (${normalRate} RPS)`);
  executeLoad(normalRate, 60);
}

function executeLoad(rps, duration) {
  const startTime = Date.now();
  const interval = 1000 / rps; // ms between requests

  while (Date.now() - startTime < duration * 1000) {
    testBanditAssign();
    sleep(interval);
  }
}

// Soak test - sustained load over extended period
export function soakTest() {
  console.log('Starting soak test (2 hours)');

  const startTime = Date.now();
  const duration = 2 * 60 * 60 * 1000; // 2 hours
  let iteration = 0;

  while (Date.now() - startTime < duration) {
    testBanditAssign();
    testBanditReward();

    // Report metrics every 5 minutes
    iteration++;
    if (iteration % 300 === 0) { // ~5 min with 1 sec intervals
      const elapsed = ((Date.now() - startTime) / 60000).toFixed(0);
      console.log(`[${elapsed}min] Assignments: ${metrics.assignments.get('total')}, Errors: ${metrics.assignments.get('error')}`);
    }

    sleep(1000);
  }
}

// Cache hit rate test
export function testCacheHitRate() {
  console.log('Testing cache hit rate with sticky assignments');

  const userId = `cache_test_user`;
  const iterations = 100;

  // First assignment - should cache
  const firstStart = Date.now();
  let firstDuration;

  for (let i = 0; i < iterations; i++) {
    const payload = {
      experiment_id: EXPERIMENT_ID,
      user_id: userId,
    };

    const res = http.post(
      `${BASE_URL}/v1/bandit/assign`,
      JSON.stringify(payload),
      { headers: { 'Content-Type': 'application/json' } }
    );

    check(res, { 'Assignment successful': (r) => r.status === 200 });

    if (i === 0) {
      firstDuration = Date.now() - firstStart;
    }

    sleep(10);
  }

  const cacheHits = metrics.assignments.get('cache_miss');
  const hitRate = ((iterations - cacheHits) / iterations) * 100;

  console.log(`Cache hit rate: ${hitRate.toFixed(1)}% (${iterations - cacheHits}/${iterations})`);
  console.log(`First assignment: ${firstDuration}ms`);

  // Assert cache hit rate > 95% for sticky assignments
  if (hitRate < 95) {
    console.warn(`WARNING: Cache hit rate below 95%: ${hitRate.toFixed(1)}%`);
  }
}
