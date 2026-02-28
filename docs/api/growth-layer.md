# Growth Layer API Documentation

## Overview

The Growth Layer provides A/B testing, multi-armed bandit optimization, and analytics capabilities for the IAP system.

## Features

- **Multi-armed Bandit**: Thompson Sampling algorithm for automatic variant optimization
- **Matomo Integration**: Event tracking, cohort analysis, funnel analytics
- **LTV Prediction**: 30/90/365-day lifetime value estimates
- **Realtime Analytics**: Sub-second metric queries with Redis caching

---

## Bandit API

### Assign User to Variant

Assigns a user to an experiment arm using Thompson Sampling. Assignments are sticky for 24h.

**Request:**
```http
POST /v1/bandit/assign
Content-Type: application/json

{
  "experiment_id": "uuid",
  "user_id": "uuid"
}
```

**Response:**
```json
{
  "data": {
    "experiment_id": "uuid",
    "user_id": "uuid",
    "arm_id": "uuid",
    "is_new": false
  }
}
```

### Record Conversion/Reward

Records a reward (conversion or revenue) for an arm, updating the Thompson Sampling parameters.

**Request:**
```http
POST /v1/bandit/reward
Content-Type: application/json

{
  "experiment_id": "uuid",
  "arm_id": "uuid",
  "user_id": "uuid",
  "reward": 9.99
}
```

**Response:**
```json
{
  "data": {
    "experiment_id": "uuid",
    "arm_id": "uuid",
    "reward": 9.99,
    "updated": true
  }
}
```

### Get Experiment Statistics

Retrieves statistics for all arms in an experiment with optional win probabilities.

**Request:**
```http
GET /v1/bandit/statistics?experiment_id={uuid}&win_probs=true
```

**Response:**
```json
{
  "data": {
    "experiment_id": "uuid",
    "arms": [
      {
        "arm_id": "uuid",
        "alpha": 10.0,
        "beta": 5.0,
        "samples": 15,
        "conversions": 10,
        "revenue": 99.90,
        "avg_reward": 6.66,
        "conversion_rate": 0.667
      }
    ],
    "win_probabilities": {
      "arm_uuid_1": 0.75,
      "arm_uuid_2": 0.25
    }
  }
}
```

---

## Analytics API

### Get User LTV

Calculates and retrieves LTV estimates for a user (30/90/365-day).

**Request:**
```http
GET /api/v1/analytics/ltv?user_id={uuid}
```

**Response:**
```json
{
  "data": {
    "user_id": "uuid",
    "ltv30": 9.99,
    "ltv90": 29.97,
    "ltv365": 119.88,
    "ltv_lifetime": 9.99,
    "confidence": 0.8,
    "calculated_at": "2026-03-01T00:00:00Z",
    "method": "cohort_based",
    "factors": {
      "actual_30day": 9.99,
      "predicted_90day": 29.97
    }
  }
}
```

### Update LTV After Purchase

Updates LTV estimates after a new purchase.

**Request:**
```http
POST /api/v1/analytics/ltv
Content-Type: application/json

{
  "user_id": "uuid",
  "amount": 9.99
}
```

### Get Cohort LTV

Retrieves aggregate LTV metrics for a user cohort.

**Request:**
```http
GET /api/v1/analytics/cohort/ltv?cohort_date=2026-03-01
```

**Response:**
```json
{
  "data": {
    "cohort_date": "2026-03-01T00:00:00Z",
    "cohort_size": 1000,
    "ltv30": 9.99,
    "ltv90": 29.97,
    "ltv365": 119.88
  }
}
```

### Get Churn Risk

Predicts the probability of user churn.

**Request:**
```http
GET /api/v1/analytics/churn/risk?user_id={uuid}
```

**Response:**
```json
{
  "data": {
    "user_id": "uuid",
    "risk": 0.25,
    "risk_level": "low",
    "timestamp": "2026-03-01T00:00:00Z"
  }
}
```

### Get Funnel Analysis

Retrieves funnel analysis data from Matomo.

**Request:**
```http
GET /api/v1/analytics/funnels?funnel_id={id}&date_from=2026-03-01&date_to=2026-03-07&use_cache=true
```

**Response:**
```json
{
  "data": {
    "funnel_id": "purchase_funnel",
    "funnel_name": "Purchase Funnel",
    "steps": [
      {
        "step_id": "step1",
        "step_name": "View Product",
        "visitors": 1000,
        "dropoff": 0,
        "dropoff_rate": 0.0
      },
      {
        "step_id": "step2",
        "step_name": "Add to Cart",
        "visitors": 600,
        "dropoff": 400,
        "dropoff_rate": 0.4
      }
    ],
    "total_entries": 1000,
    "total_exits": 700,
    "conversion_rate": 0.30
  }
}
```

### Get Realtime Metrics

Retrieves current realtime metrics (30s cache).

**Request:**
```http
GET /api/v1/analytics/realtime?metrics=active_users
```

**Response:**
```json
{
  "data": [
    {
      "name": "active_users",
      "value": 5432.0,
      "timestamp": "2026-03-01T00:00:00Z",
      "tags": {
        "platform": "ios"
      }
    }
  ]
}
```

---

## Mobile SDK Integration

### React Native (TypeScript)

Install the Matomo SDK:

```typescript
import { MatomoAnalytics, initMatomo } from './infrastructure/analytics/matomo';

// Initialize
const matomo = initMatomo(
  {
    baseUrl: 'https://matomo.example.com',
    siteId: '1',
    batchSize: 20,
    flushInterval: 30000,
  },
  httpClient
);

await matomo.initialize();

// Track events
await matomo.trackEvent(
  'paywall',
  'shown',
  'premium_monthly',
  0,
  {
    experiment_id: 'exp_123',
    variant: 'control'
  }
);

// Track purchases
await matomo.trackPurchase(
  'order_123',
  9.99,
  [
    {
      sku: 'com.app.premium.monthly',
      name: 'Premium Monthly',
      price: 9.99,
      quantity: 1
    }
  ],
  {
    experiment_id: 'exp_123',
    variant: 'variant_a'
  }
);

// Flush manually (optional)
await matomo.flush();

// Get queue stats
const stats = matomo.getQueueStats();
console.log(stats.size, stats.pending, stats.failed);
```

---

## Performance Benchmarks

Run the k6 load test:

```bash
k6 run tests/load/bandit_bench.js \
  --env VUS=100 \
  --env DURATION=5m \
  --env BASE_URL=http://localhost:8080
```

**Target Metrics:**
- P99 latency < 50ms for assignments
- P99 latency < 200ms for rewards
- Error rate < 1%
- Throughput > 100 req/sec

**Cache Targets:**
- Assignment cache hit rate > 95%
- Realtime metrics: 30s TTL
- LTV data: 1h TTL
- Funnel data: 30min TTL

---

## Webhook Integration

Configure webhooks to track purchases for LTV updates:

```yaml
# Stripe webhook
POST /webhook/stripe
  → Updates LTV
  → Records reward for bandit arm

# Apple/Google webhooks
POST /webhook/apple
POST /webhook/google
  → Updates LTV
  → Records reward for bandit arm
```

---

## Database Schema

### ab_tests
Experiment configurations with bandit support

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| name | TEXT | Experiment name |
| algorithm_type | TEXT | thompson_sampling, ucb, epsilon_greedy |
| is_bandit | BOOLEAN | Uses bandit optimization |
| min_sample_size | INT | Minimum samples before declaring winner |
| confidence_threshold | NUMERIC | Statistical confidence level |

### ab_test_arms
Experiment arms (variants)

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| experiment_id | UUID | FK to ab_tests |
| name | TEXT | Arm name |
| is_control | BOOLEAN | Control arm flag |
| traffic_weight | NUMERIC | Initial traffic weight |

### ab_test_arm_stats
Thompson Sampling parameters

| Column | Type | Description |
|--------|------|-------------|
| arm_id | UUID | FK to ab_test_arms |
| alpha | NUMERIC | Successes + prior (Beta distribution) |
| beta | NUMERIC | Failures + prior |
| samples | INT | Total samples |
| conversions | INT | Conversions count |
| revenue | NUMERIC | Total revenue |

### matomo_staged_events
Event queue for async Matomo delivery

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| event_type | TEXT | event, ecommerce, custom |
| payload | JSONB | Event data |
| retry_count | INT | Retry counter |
| max_retries | INT | Max retries (default: 3) |
| next_retry_at | TIMESTAMPTZ | Exponential backoff time |
| status | TEXT | pending, processing, failed, sent |

---

## Configuration

### Environment Variables

```bash
# Bandit Configuration
BANDIT_MIN_SAMPLE_SIZE=100
BANDIT_CONFIDENCE_THRESHOLD=0.95
BANDIT_ASSIGNMENT_TTL=24h

# Matomo Configuration
MATOMO_BASE_URL=https://matomo.example.com
MATOMO_SITE_ID=1
MATOMO_TOKEN_AUTH=your_token_here
MATOMO_BATCH_SIZE=100
MATOMO_MAX_RETRIES=3

# Cache Configuration
REDIS_URL=redis://:password@localhost:6379
CACHE_TTL_REALTIME=30s
CACHE_TTL_LTV=1h
CACHE_TTL_FUNNEL=30m
```

---

## Deployment

### Docker Compose

```bash
docker-compose -f docker-compose.latency-optimized.yml up -d
```

### Run Migrations

```bash
cd backend
go run ./cmd/migrator migrate up
```

### Start Services

```bash
# API
go run ./cmd/api

# Worker (for Matomo event delivery)
go run ./cmd/worker
```

---

## Monitoring

### Key Metrics

- **Bandit Assignment Latency**: P99 < 50ms
- **Matomo Event Delivery**: P99 < 5s
- **Cache Hit Rate**: >95% for assignments
- **Event Loss Rate**: <0.01%

### Prometheus Queries

```promql
# Bandit assignment rate
rate(http_requests_total{endpoint="/v1/bandit/assign"}[5m])

# Cache hit rate
rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))

# Matomo queue size
matomo_staged_events_count{status="pending"}

# LTV calculation latency
histogram_quantile(0.99, ltv_calculation_duration_seconds)
```

---

## Troubleshooting

### Assignments Not Sticky

Check Redis connectivity and TTL settings:
```bash
redis-cli TTL "ab:assign:{experiment_id}:{user_id}"
```

### Matomo Events Not Delivered

Check staged events queue:
```sql
SELECT status, COUNT(*), AVG(retry_count)
FROM matomo_staged_events
GROUP BY status;
```

### High Cache Miss Rate

Verify Redis is configured correctly and keys are being set:
```bash
redis-cli KEYS "ab:*" | wc -l
```

### Slow LTV Calculations

Check cohort worker logs for Matomo API latency:
```
tail -f /var/log/worker.log | grep "GetCohorts"
```

---

## References

- [Thompson Sampling Technical Guide](../../plans/2026-03-01-thompson-sampling-technical-guide.md)
- [Growth Layer Design](../../plans/2026-03-01-growth-layer-design.md)
- [Matomo HTTP Tracking API](https://developer.matomo.org/api-reference/tracking-api)
