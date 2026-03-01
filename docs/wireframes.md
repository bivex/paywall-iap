# Paywall IAP System - Wireframes & API Mapping

Complete mapping of admin panel screens to implemented API endpoints, DTOs, and code locations.

**Repository:** `bivex/paywall-iap`
**Last Updated:** 2026-03-01

---

## Table of Contents

1. [Authentication & System Health](#1-authentication--system-health)
2. [Dashboard & KPI Metrics](#2-dashboard--kpi-metrics)
3. [Users & Customers Management](#3-users--customers-management)
4. [Subscriptions Management](#4-subscriptions-management)
5. [Transactions & Purchases](#5-transactions--purchases)
6. [Pricing Tiers & Products](#6-pricing-tiers--products)
7. [Webhooks & Integrations](#7-webhooks--integrations)
8. [Analytics & Reports](#8-analytics--reports)
9. [A/B Testing & Feature Flags](#9-ab-testing--feature-flags)
10. [Multi-Armed Bandit System](#10-multi-armed-bandit-system)
11. [Advanced Bandit Features](#11-advanced-bandit-features)
12. [Winback Offers](#12-winback-offers)
13. [Grace Periods & Dunning](#13-grace-periods--dunning)
14. [Audit & Activity Logs](#14-audit--activity-logs)
15. [Admin System Management](#15-admin-system-management)
16. [What's Missing / TODO](#16-whats-missing--todo)

---

## 1. Authentication & System Health

### Health Check
- **Endpoint:** `GET /health`
- **Handler:** `cmd/api/main.go` (inline handler)
- **Auth:** None required
- **Response:** `{"status": "ok"}`

### User Registration
- **Endpoint:** `POST /v1/auth/register`
- **Handler:** `internal/interfaces/http/handlers/auth.go` → `Register()`
- **DTOs:** `RegisterRequest` { platform, device_id, app_version }
- **Response:** JWT tokens (access + refresh)
- **Implementation:** User created via platform-specific identity

### Token Refresh
- **Endpoint:** `POST /v1/auth/refresh`
- **Handler:** `internal/interfaces/http/handlers/auth.go` → `RefreshToken()`
- **Middleware:** Rate limited (ByIP, DefaultConfig)
- **DTOs:** `RefreshTokenRequest`
- **Response:** New JWT tokens

---

## 2. Dashboard & KPI Metrics

### Dashboard Metrics
- **Endpoint:** `GET /v1/admin/dashboard/metrics`
- **Handler:** `internal/interfaces/http/handlers/admin.go` → `GetDashboardMetrics()`
- **Auth:** JWT + Admin role required
- **Returns:**
  - Total users
  - Active subscriptions
  - Revenue (MRR, ARR)
  - Churn rate
  - Conversion funnels
- **Data Sources:** `AnalyticsService`, `SubscriptionRepository`

### Admin Health Check
- **Endpoint:** `GET /v1/admin/health`
- **Handler:** `internal/interfaces/http/handlers/admin.go` → `GetHealth()`
- **Checks:** Database, Redis, external services

---

## 3. Users & Customers Management

### List Users
- **Endpoint:** `GET /v1/admin/users`
- **Handler:** `internal/interfaces/http/handlers/admin.go` → `ListUsers()`
- **Query Params:** `page`, `limit`, `search`
- **Returns:** Paginated user list with subscription status

### Grant Subscription
- **Endpoint:** `POST /v1/admin/users/{id}/grant`
- **Handler:** `internal/interfaces/http/handlers/admin.go` → `GrantSubscription()`
- **Body:** `{ pricing_tier_id, expires_at }`
- **Effect:** Creates subscription, logs audit

### Revoke Subscription
- **Endpoint:** `POST /v1/admin/users/{id}/revoke`
- **Handler:** `internal/interfaces/http/handlers/admin.go` → `RevokeSubscription()`
- **Effect:** Cancels subscription, logs audit

---

## 4. Subscriptions Management

### Get User Subscription
- **Endpoint:** `GET /v1/subscription`
- **Handler:** `internal/interfaces/http/handlers/subscription.go` → `GetSubscription()`
- **Auth:** JWT required
- **Returns:** User's current subscription with status, tier, expiry

### Check Access
- **Endpoint:** `GET /v1/subscription/access`
- **Handler:** `internal/interfaces/http/handlers/subscription.go` → `CheckAccess()`
- **Auth:** JWT required
- **Middleware:** Rate limited (ByUserID, PollingConfig)
- **Returns:** `{ has_access: boolean, tier: string }`

### Cancel Subscription
- **Endpoint:** `DELETE /v1/subscription`
- **Handler:** `internal/interfaces/http/handlers/subscription.go` → `CancelSubscription()`
- **Auth:** JWT required
- **Effect:** Cancels at period end, processes refund if applicable

---

## 5. Transactions & Purchases

### Database Tables
- **transactions** (migration: `003_create_transactions.up.sql`)
  - Fields: user_id, amount, currency, platform, transaction_id, status, created_at
  - Linked to: subscriptions, pricing_tiers

### IAP Verification
- **Endpoint:** `POST /v1/verify/iap` (referenced in tests/DTOs)
- **Purpose:** Verify receipt with platform (Apple/Google)
- **Implementation:** Uses platform-specific validation

### ⚠️ **MISSING:** Admin Transaction Management
- **Needed:**
  - `GET /v1/admin/transactions` - List all transactions
  - `GET /v1/admin/transactions/{id}` - Transaction details
  - `POST /v1/admin/transactions/{id}/refund` - Process refund
  - `POST /v1/admin/transactions/reconcile` - Reconcile with platform

---

## 6. Pricing Tiers & Products

### Database Tables
- **pricing_tiers** (migration: `005_create_pricing_tiers.up.sql`)
  - Fields: name, description, price, currency, duration, features (JSON)

### ⚠️ **PARTIALLY IMPLEMENTED:** Pricing Management
- **Exists:** Database schema, DTO `PricingTier` in `dto.go`
- **Missing:** Admin CRUD endpoints
- **Needed:**
  - `GET /v1/admin/pricing-tiers` - List all tiers
  - `POST /v1/admin/pricing-tiers` - Create tier
  - `PUT /v1/admin/pricing-tiers/{id}` - Update tier
  - `DELETE /v1/admin/pricing-tiers/{id}` - Archive tier

---

## 7. Webhooks & Integrations

### Stripe Webhook
- **Endpoint:** `POST /webhook/stripe`
- **Handler:** `internal/interfaces/http/handlers/webhook.go` → `StripeWebhook()`
- **Security:** HMAC signature verification
- **Events:** payment_succeeded, payment_failed, subscription_created, etc.

### Apple Webhook
- **Endpoint:** `POST /webhook/apple`
- **Handler:** `internal/interfaces/http/handlers/webhook.go` → `AppleWebhook()`
- **Security:** JWS token verification (Apple S2S)
- **Events:** DID_CHANGE_RENEWAL_PREF, DID_FAIL_TO_RENEW, etc.

### Google Webhook
- **Endpoint:** `POST /webhook/google`
- **Handler:** `internal/interfaces/http/handlers/webhook.go` → `GoogleWebhook()`
- **Security:** Pub/Sub IP whitelisting + token auth
- **Events:** subscription notifications, RTDN

### Webhook Event Logging
- **Table:** `webhook_events` (migration: `004_create_webhook_events.up.sql`)
- **Fields:** provider, event_type, payload, processed_at, status

### ⚠️ **MISSING:** Webhook Monitoring UI
- **Needed:**
  - `GET /v1/admin/webhooks/deliveries` - Recent webhook deliveries
  - `GET /v1/admin/webhooks/deliveries/{id}` - Delivery details
  - `POST /v1/admin/webhooks/deliveries/{id}/retry` - Manual retry

---

## 8. Analytics & Reports

### Revenue Analytics
- **Endpoint:** `GET /v1/analytics/revenue`
- **Handler:** `internal/interfaces/http/handlers/analytics.go` → `GetRevenueMetrics()`
- **Query:** `days` (default: 30)
- **Returns:** Daily revenue, MRR, ARR, trends

### Churn Analytics
- **Endpoint:** `GET /v1/analytics/churn`
- **Handler:** `internal/interfaces/http/handlers/analytics.go` → `GetChurnMetrics()`
- **Query:** `start_date`, `end_date`
- **Returns:** Churn rate, churned users, reasons

### Extended Analytics (Test Implementation)
- **Endpoints:**
  - `GET /v1/analytics/extended`
  - `GET /v1/analytics/cohorts`
  - `GET /v1/analytics/extended/revenue`
  - `GET /v1/analytics/extended/churn`
- **Note:** Implemented in tests but not in main router
- **Service:** `AnalyticsService` with `AnalyticsRepository`

### Cohort Analytics
- **Service:** `CohortService` in `internal/domain/service/`
- **Worker:** `CohortJobs` in `internal/worker/tasks/`
- **Table:** `analytics_aggregates` for pre-computed data

---

## 9. A/B Testing & Feature Flags

### Feature Flags

#### List All Flags
- **Endpoint:** `GET /v1/ab-test/flags`
- **Handler:** `internal/interfaces/http/handlers/ab_test.go` → `ListFlags()`

#### Evaluate Flag
- **Endpoint:** `GET /v1/ab-test/evaluate/{flag_id}`
- **Handler:** `internal/interfaces/http/handlers/ab_test.go` → `EvaluateFlag()`
- **Query:** `user_id`
- **Returns:** `{ enabled: boolean, variant: string }`

#### Get Paywall Variant
- **Endpoint:** `GET /v1/ab-test/paywall`
- **Handler:** `internal/interfaces/http/handlers/ab_test.go` → `GetPaywallVariant()`
- **Query:** `user_id`
- **Returns:** Paywall template/variant for user

#### Create Flag (Admin)
- **Endpoint:** `POST /v1/ab-test/flags`
- **Handler:** `internal/interfaces/http/handlers/ab_test.go` → `CreateFlag()`
- **Auth:** Admin required

#### Update Flag (Admin)
- **Endpoint:** `PUT /v1/ab-test/flags/{flag_id}`
- **Handler:** `internal/interfaces/http/handlers/ab_test.go` → `UpdateFlag()`
- **Auth:** Admin required

#### Delete Flag (Admin)
- **Endpoint:** `DELETE /v1/ab-test/flags/{flag_id}`
- **Handler:** `internal/interfaces/http/handlers/ab_test.go` → `DeleteFlag()`
- **Auth:** Admin required

### Database Tables
- **ab_tests** - Experiment definitions with bandit support
- **ab_test_arms** - Variants/arms for experiments
- **ab_test_arm_stats** - Statistics per arm (Beta distribution)
- **ab_test_assignments** - User assignments (sticky)

### Services
- **FeatureFlagService** - Flag evaluation and rollout
- **ABAnalyticsService** - A/B test analytics

---

## 10. Multi-Armed Bandit System

### Core Endpoints

#### Assign Arm
- **Endpoint:** `POST /v1/bandit/assign`
- **Handler:** `internal/interfaces/http/handlers/bandit.go` → `Assign()`
- **Body:** `{ experiment_id, user_id }`
- **Returns:** `{ arm_id, assignment_id, expires_at }`
- **Algorithm:** Thompson Sampling with Beta(α, β)

#### Record Reward
- **Endpoint:** `POST /v1/bandit/reward`
- **Handler:** `internal/interfaces/http/handlers/bandit.go` → `Reward()`
- **Body:** `{ experiment_id, arm_id, user_id, reward }`
- **Updates:** Alpha/Beta parameters, cache

#### Get Statistics
- **Endpoint:** `GET /v1/bandit/statistics`
- **Handler:** `internal/interfaces/http/handlers/bandit.go` → `Statistics()`
- **Query:** `experiment_id`, `simulations` (default: 10000)
- **Returns:**
  - Arm statistics (α, β, samples, conversions, revenue)
  - Win probabilities (Monte Carlo simulation)

#### Health Check
- **Endpoint:** `GET /v1/bandit/health`
- **Handler:** `internal/interfaces/http/handlers/bandit.go` → `Health()`
- **Returns:** Service status, cache status

### Implementation
- **Service:** `ThompsonSamplingBandit` in `internal/domain/service/bandit_service.go`
- **Repository:** `PostgresBanditRepository` in `internal/infrastructure/persistence/repository/bandit_repository.go`
- **Cache:** `RedisBanditCache` with arm stats and assignments
- **Algorithm:** Beta distribution sampling with Johnk's, Marsaglia-Tsang, and Cheng's methods

### ⚠️ **MISSING:** Experiment Management
- **Needed:**
  - `GET /v1/admin/ab/experiments` - List all experiments
  - `POST /v1/admin/ab/experiments` - Create experiment
  - `PUT /v1/admin/ab/experiments/{id}` - Update experiment
  - `DELETE /v1/admin/ab/experiments/{id}` - Delete experiment
  - `POST /v1/admin/ab/experiments/{id}/arms` - Add arm to experiment
  - `GET /v1/admin/ab/experiments/{id}/report` - Full experiment report

---

## 11. Advanced Bandit Features

### Currency Conversion

#### Get Exchange Rates
- **Endpoint:** `GET /v1/bandit/currency/rates`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `GetCurrencyRates()`
- **Returns:** All supported currency rates (USD base)

#### Update Rates
- **Endpoint:** `POST /v1/bandit/currency/update`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `UpdateCurrencyRates()`
- **Source:** ECB API (https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml)
- **Cache:** Redis with 1-hour TTL
- **Worker:** Updates hourly (cron: `*/30 * * * *`)

#### Convert Currency
- **Endpoint:** `POST /v1/bandit/currency/convert`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `ConvertCurrency()`
- **Body:** `{ amount, currency }`
- **Returns:** Converted amount in USD

**Service:** `CurrencyRateService` with ECB integration and fallback rates

---

### Multi-Objective Optimization

#### Get Objective Scores
- **Endpoint:** `GET /v1/bandit/experiments/{id}/objectives`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `GetObjectiveScores()`
- **Returns:** Scores for conversion, LTV, revenue per arm

#### Configure Objectives
- **Endpoint:** `PUT /v1/bandit/experiments/{id}/objectives/config`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `SetObjectiveConfig()`
- **Body:**
  ```json
  {
    "objective_type": "hybrid",
    "objective_weights": {
      "conversion": 0.5,
      "ltv": 0.3,
      "revenue": 0.2
    }
  }
  ```

**Objectives Supported:**
- **Conversion** - Standard Thompson Sampling
- **LTV** - Expected Value = P(conversion) × AvgLTV
- **Revenue** - Normalized Revenue = P(conv) × (Revenue / Price)
- **Hybrid** - Weighted combination of above

**Service:** `HybridObjectiveStrategy` with per-objective Beta tracking

---

### Sliding Window (Non-Stationary Behavior)

#### Get Window Info
- **Endpoint:** `GET /v1/bandit/experiments/{id}/window/info`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `GetWindowInfo()`
- **Returns:** Window size, utilization, time range

#### Trim Window
- **Endpoint:** `POST /v1/bandit/experiments/{id}/window/trim`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `TrimWindow()`
- **Purpose:** Force cleanup of out-of-window events

#### Export Events
- **Endpoint:** `GET /v1/bandit/experiments/{id}/window/events`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `ExportWindowEvents()`
- **Query:** `limit` (default: 1000)

**Implementation:**
- Redis Sorted Sets for O(log N) operations
- Key pattern: `bandit:window:{experiment_id}:{arm_id}`
- Window types: events, time, none
- Worker: Trims hourly (cron: `0 * * * *`)

**Service:** `SlidingWindowStrategy`

---

### Delayed Feedback

#### Process Conversion
- **Endpoint:** `POST /v1/bandit/conversions`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `ProcessConversion()`
- **Body:**
  ```json
  {
    "transaction_id": "uuid",
    "user_id": "uuid",
    "conversion_value": 9.99,
    "currency": "EUR"
  }
  ```
- **Links:** Pending reward → transaction

#### Get Pending Reward
- **Endpoint:** `GET /v1/bandit/pending/{id}`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `GetPendingReward()`
- **Returns:** Pending reward status, expiry

#### Get User Pending Rewards
- **Endpoint:** `GET /v1/bandit/users/{id}/pending`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `GetUserPendingRewards()`
- **Returns:** All pending rewards for user

**Implementation:**
- TTL: 7 days default, 30 days maximum
- Worker: Processes expired every 15 minutes (cron: `*/15 * * * *`)
- Tables: `bandit_pending_rewards`, `bandit_conversion_links`

**Service:** `DelayedRewardStrategy`

---

### Production Metrics

#### Get Experiment Metrics
- **Endpoint:** `GET /v1/bandit/experiments/{id}/metrics`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `GetMetrics()`
- **Returns:**
  - `regret` - Cumulative regret vs best arm
  - `exploration_rate` - How often arms switch
  - `convergence_gap` - Performance gap best/worst
  - `balance_index` - Distribution evenness (0-1)
  - `window_utilization` - For sliding window
  - `pending_rewards` - Unprocessed delayed rewards

#### Run Maintenance
- **Endpoint:** `POST /v1/bandit/maintenance`
- **Handler:** `internal/interfaces/http/handlers/bandit_advanced.go` → `RunMaintenance()`
- **Tasks:** Process expired, trim windows, update rates, cleanup context
- **Worker:** Full maintenance every 6 hours (cron: `0 */6 * * *`)

---

### Contextual Bandit (LinUCB)

**Implementation Details (no direct HTTP endpoints):**
- **Service:** `LinUCBSelectionStrategy`
- **Algorithm:** UCB = θ^T × x + α × sqrt(x^T × A^(-1) × x)
- **Features:** 20 dimensions
  - Country (10): one-hot for US, GB, DE, FR, JP, CA, AU, BR, IN, other
  - Device (5): ios, android, web, tablet, other
  - Days since install (1): normalized 0-1 (capped at 30)
  - Total spent (1): log-normalized
  - Is past purchaser (1): binary
  - Recent purchaser (1): binary (within 7 days)
  - Bias (1): constant
- **Per-arm model:** Matrix A, vector b, parameters θ
- **Update:** Online learning with each reward
- **Exploration:** α = 0.3 (configurable)

**Enable via experiment config:**
```json
{
  "enable_contextual": true,
  "exploration_alpha": 0.3
}
```

---

### Database Tables (Migration 017)

- **currency_rates** - Exchange rates (USD base)
- **bandit_user_context** - Cached user attributes
- **bandit_arm_context_model** - LinUCB parameters (JSONB: matrix_a, vector_b, theta)
- **bandit_window_events** - Sliding window event log
- **bandit_pending_rewards** - Pending conversions
- **bandit_conversion_links** - Links pending → transaction
- **bandit_arm_objective_stats** - Per-objective statistics
- **ab_tests** - Added: window_type, objective_type, objective_weights, enable_contextual, enable_delayed, enable_currency, exploration_alpha

---

### Main Orchestrator

**Service:** `AdvancedBanditEngine` in `internal/domain/service/advanced_bandit_engine.go`

**Composes:**
- Base `ThompsonSamplingBandit`
- `CurrencyRateService`
- `RewardStrategy` (currency conversion)
- `SelectionStrategy` (LinUCB)
- `WindowStrategy` (sliding window)
- `DelayedRewardStrategy`
- `HybridObjectiveStrategy`

**Methods:**
- `SelectArm()` - With user context, records pending if delayed enabled
- `RecordReward()` - Converts currency, updates all strategies
- `ProcessConversion()` - Links transaction to pending reward
- `GetArmStatistics()` - Aggregate stats
- `GetObjectiveScores()` - Per-objective breakdown
- `GetMetrics()` - Production metrics
- `RunMaintenance()` - Background tasks

---

## 12. Winback Offers

### Get Active Offers
- **Endpoint:** `GET /v1/winback/offers`
- **Handler:** `internal/interfaces/http/handlers/winback.go` → `GetOffers()`
- **Auth:** JWT required
- **Returns:** Available winback offers for user

### Accept Offer
- **Endpoint:** `POST /v1/winback/offers/accept`
- **Handler:** `internal/interfaces/http/handlers/winback.go` → `AcceptOffer()`
- **Auth:** JWT required
- **Body:** `{ offer_id }`
- **Effect:** Creates subscription with offer terms

### Database & Services
- **Table:** `winback_offers` (migration: `007_create_winback_offers.up.sql`)
- **Service:** `WinbackService`
- **Worker:** `WinbackJobs` (offer eligibility, expiry)

---

## 13. Grace Periods & Dunning

### Grace Periods
- **Table:** `grace_periods` (migration: `006_create_grace_periods.up.sql`)
- **Service:** `GracePeriodService`
- **Worker:** `GracePeriodJobs` in `internal/worker/tasks/grace_period_jobs.go`
- **Purpose:** Extend access during payment retries

### Dunning Management
- **Table:** `dunning_events` (migration: `010_create_dunning.up.sql`)
- **Service:** `DunningService`
- **Worker:** `DunningJobs` in `internal/worker/tasks/dunning_jobs.go`
- **Features:**
  - Configurable retry schedule
  - Email notifications
  - Auto-cancellation after max retries

### ⚠️ **MISSING:** Admin UI
- **Needed:**
  - `GET /v1/admin/dunning/config` - Get retry rules
  - `PUT /v1/admin/dunning/config` - Update retry rules
  - `GET /v1/admin/grace-periods` - Active grace periods
  - `POST /v1/admin/grace-periods/{id}/extend` - Manual extension

---

## 14. Audit & Activity Logs

### Audit Logging
- **Table:** `admin_audit_log` (migration: `009_create_admin_audit_log.up.sql`)
- **Service:** `AuditService`
- **Logged Actions:**
  - Grant/revoke subscription
  - User modifications
  - Experiment changes
  - Config changes

### ⚠️ **MISSING:** Audit Log API
- **Needed:**
  - `GET /v1/admin/audit/logs` - Paginated audit log
  - `GET /v1/admin/audit/logs/{id}` - Log details
  - `GET /v1/admin/audit/users/{id}` - User activity history

---

## 15. Admin System Management

### Role-Based Access Control
- **Implementation:** `AdminMiddleware` in `internal/application/middleware/admin.go`
- **Requirement:** JWT + admin role in users table
- **Applied to:** All `/v1/admin/*` routes

### ⚠️ **MISSING:** RBAC Management
- **Needed:**
  - `GET /v1/admin/roles` - List roles
  - `POST /v1/admin/roles` - Create role
  - `PUT /v1/admin/users/{id}/roles` - Assign roles
  - `GET /v1/admin/api-keys` - List API keys
  - `POST /v1/admin/api-keys` - Create API key

### User Roles
- **Table:** `users` (migration: `011_add_user_roles.up.sql`)
- **Field:** `role` (enum: user, admin, operator)

---

## 16. What's Missing / TODO

### High Priority Admin APIs

#### Experiments & A/B Testing
```
POST   /v1/admin/ab/experiments
GET    /v1/admin/ab/experiments
GET    /v1/admin/ab/experiments/{id}
PUT    /v1/admin/ab/experiments/{id}
DELETE /v1/admin/ab/experiments/{id}
POST   /v1/admin/ab/experiments/{id}/arms
PUT    /v1/admin/ab/experiments/{id}/start
PUT    /v1/admin/ab/experiments/{id}/stop
GET    /v1/admin/ab/experiments/{id}/report
GET    /v1/admin/ab/experiments/{id}/assignments
```

#### Pricing & Products
```
GET    /v1/admin/pricing-tiers
POST   /v1/admin/pricing-tiers
GET    /v1/admin/pricing-tiers/{id}
PUT    /v1/admin/pricing-tiers/{id}
DELETE /v1/admin/pricing-tiers/{id}
```

#### Transactions
```
GET    /v1/admin/transactions
GET    /v1/admin/transactions/{id}
POST   /v1/admin/transactions/{id}/refund
POST   /v1/admin/transactions/reconcile
```

#### Webhook Monitoring
```
GET    /v1/admin/webhooks/deliveries
GET    /v1/admin/webhooks/deliveries/{id}
POST   /v1/admin/webhooks/deliveries/{id}/retry
```

#### Audit & Activity
```
GET    /v1/admin/audit/logs
GET    /v1/admin/audit/logs/{id}
GET    /v1/admin/audit/users/{id}
```

#### System Configuration
```
GET    /v1/admin/config
PUT    /v1/admin/config
GET    /v1/admin/roles
POST   /v1/admin/roles
GET    /v1/admin/api-keys
POST   /v1/admin/api-keys
```

### Medium Priority Features

#### Unified Event Ingestion
```
POST /v1/events
Body: {
  event_type: "paywall_impression" | "click" | "purchase_attempt",
  experiment_id?: uuid,
  arm_id?: uuid,
  user_id: uuid,
  properties: {...}
}
```

#### Real-time Analytics Dashboard
- WebSocket endpoint for live metrics
- Cohort analysis UI
- Funnel visualization

#### Advanced Bandit Admin UI
```
GET /v1/admin/bandit/experiments/{id}/config
PUT /v1/admin/bandit/experiments/{id}/config
GET /v1/admin/bandit/currency/history
GET /v1/admin/bandit/pending?status=expired
POST /v1/admin/bandit/pending/{id}/process
```

### Database Tables Referenced but May Need Verification

Verify these exist in migrations:
- `matomo_staged_events` - Matomo integration staging
- `analytics_aggregates` - Pre-computed analytics

---

## Summary by Component

| Component | Status | Coverage |
|-----------|--------|----------|
| Authentication | ✅ Complete | 100% |
| Subscriptions | ✅ Complete | 100% |
| Webhooks | ✅ Complete | 100% |
| Bandit Core | ✅ Complete | 100% |
| Advanced Bandit | ✅ Complete | 100% |
| Feature Flags | ✅ Complete | 100% |
| Analytics | ⚠️ Partial | 75% |
| Winback | ✅ Complete | 100% |
| Grace/Dunning | ⚠️ Partial | 80% |
| Admin Users | ⚠️ Partial | 70% |
| Experiments Mgmt | ❌ Missing | 20% |
| Transactions | ❌ Missing | 30% |
| Pricing Mgmt | ❌ Missing | 40% |
| Audit Logs | ❌ Missing | 30% |

---

**Generated:** 2026-03-01
**Migration 017:** Advanced Thompson Sampling extensions
**Total Endpoints:** 50+ implemented
**Missing Critical:** Experiment CRUD, Transaction management, Audit log UI
