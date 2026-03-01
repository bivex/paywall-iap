# Advanced Thompson Sampling Extensions - Design Document

**Date:** 2026-03-01
**Author:** Claude Opus 4.6
**Status:** Design Approved
**Context:** Growth Layer - Multi-Armed Bandit Enhancements

---

## Overview

Extend the existing Thompson Sampling implementation with four advanced features:
1. **Currency Conversion** - Convert rewards from multiple currencies to USD
2. **Contextual Bandit (LinUCB)** - Personalize based on user attributes
3. **Sliding Window** - Only consider recent samples for non-stationarity
4. **Delayed Feedback** - Handle conversions that happen days later

**Approach:** Modular/plugin system where each feature is independently enableable per experiment.

---

## Table of Contents

1. [Architecture: Modular Plugin System](#architecture-modular-plugin-system)
2. [Currency Conversion Module](#section-1-currency-conversion-module)
3. [Contextual Bandit Module](#section-2-contextual-bandit-module-linucb)
4. [Sliding Window Module](#section-3-sliding-window-module)
5. [Delayed Feedback Module](#section-4-delayed-feedback-module)
6. [Multi-Objective Hybrid System](#section-5-multi-objective-hybrid-system)
7. [Integration & Plugin System](#section-6-integration--plugin-system)
8. [Production Features](#section-7-production-features)
9. [Implementation Summary](#section-8-implementation-summary)

---

## Architecture: Modular Plugin System

Each advanced feature is implemented as an independent "strategy" that can be composed with the base `ThompsonSamplingBandit`.

```go
// Plugin interfaces
type RewardStrategy interface {
    CalculateReward(ctx context.Context, baseReward float64, arm Arm, userContext UserContext) (float64, error)
    GetType() string
}

type SelectionStrategy interface {
    SelectArm(ctx context.Context, arms []Arm, userContext UserContext) (*Arm, error)
    GetName() string
}

type WindowStrategy interface {
    GetArmStats(ctx context.Context, armID uuid.UUID) (*ArmStats, error)
    RecordEvent(ctx context.Context, armID uuid.UUID, event RewardEvent) error
}
```

---

## Section 1: Currency Conversion Module

### Purpose
Convert rewards from multiple currencies to USD for unified optimization.

### Components

1. **CurrencyRateService** - Fetches and caches exchange rates
   - Updates hourly from ECB API (free, reliable)
   - Stores rates in Redis with 1-hour TTL
   - Fallback to hardcoded rates if API fails

2. **CurrencyConversionRewardStrategy** - Wraps base rewards
   - Converts reward.Value to USD before recording
   - Tracks original currency for analytics

### Data Schema

```sql
-- Currency rates table
CREATE TABLE currency_rates (
    id SERIAL PRIMARY KEY,
    base_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    target_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(18,6) NOT NULL,
    source VARCHAR(50),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(base_currency, target_currency)
);

-- Add to ab_test_arm_stats
ALTER TABLE ab_test_arm_stats
ADD COLUMN revenue_usd DECIMAL(18,2) DEFAULT 0,
ADD COLUMN original_currency VARCHAR(3),
ADD COLUMN original_revenue DECIMAL(18,2);
```

### Flow

```
Reward (EUR, 9.99)
    ↓
CurrencyConversionRewardStrategy.CalculateReward()
    ↓
Fetch rate: EUR→USD (0.92)
    ↓
Convert: 9.99 × 0.92 = 9.19 USD
    ↓
Record: Revenue=9.19, Original=9.99 EUR
```

---

## Section 2: Contextual Bandit Module (LinUCB)

### Purpose
Factor in user attributes (country, device, LTV) for personalized arm selection.

### Components

1. **UserContext** - Captures relevant user features
2. **LinUCBSelectionStrategy** - Linear Upper Confidence Bound
   - Maintains feature matrix per arm
   - Calculates UCB = prediction + exploration_bonus
   - Updates model with each reward

### Data Schema

```sql
-- User context cache
CREATE TABLE bandit_user_context (
    user_id UUID PRIMARY KEY,
    country VARCHAR(3),
    device VARCHAR(20),
    app_version VARCHAR(20),
    days_since_install INT,
    total_spent DECIMAL(12,2),
    last_purchase_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- LinUCB model parameters
CREATE TABLE bandit_arm_context_model (
    arm_id UUID PRIMARY KEY,
    dimension INT NOT NULL DEFAULT 20,
    matrix_a JSONB NOT NULL,
    vector_b JSONB NOT NULL,
    theta JSONB NOT NULL,
    samples_count BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Algorithm

```
UCB = θ^T × x + α × sqrt(x^T × A^(-1) × x)

Where:
- θ = model parameters (learned)
- x = feature vector (user context)
- A = design matrix (updated per reward)
- α = exploration parameter (0.3)
```

---

## Section 3: Sliding Window Module

### Purpose
Only consider recent samples (e.g., last 1000 events) to handle non-stationary user behavior.

### Components

1. **SlidingWindowStrategy** - Alternative to full-history aggregation
   - Configurable window size (events or time-based)
   - Uses Redis Sorted Sets for O(log N) operations
   - Automatic cleanup of out-of-window events

### Data Schema

```sql
-- Window configuration
ALTER TABLE ab_tests
ADD COLUMN window_type VARCHAR(10) CHECK (window_type IN ('events', 'time', 'none')),
ADD COLUMN window_size INT DEFAULT 1000,
ADD COLUMN window_min_samples INT DEFAULT 100;

-- Event tracking
CREATE TABLE bandit_window_events (
    id BIGSERIAL PRIMARY KEY,
    experiment_id UUID NOT NULL,
    arm_id UUID NOT NULL,
    user_id UUID NOT NULL,
    event_type VARCHAR(10) NOT NULL,
    reward_value DECIMAL(12,2),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Redis Structure

```
bandit:window:{experiment_id}:{arm_id} → Sorted Set (score=timestamp)
bandit:window:stats:{experiment_id}:{arm_id} → Hash (cached stats)
```

---

## Section 4: Delayed Feedback Module

### Purpose
Handle conversions that happen days/weeks after initial paywall view.

### Components

1. **PendingRewardQueue** - Tracks pending conversions
2. **Background Worker** - Processes queue and links to actual purchases

### Data Schema

```sql
-- Pending reward tracking
CREATE TABLE bandit_pending_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL,
    arm_id UUID NOT NULL,
    user_id UUID NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    converted BOOLEAN DEFAULT FALSE,
    conversion_value DECIMAL(12,2),
    conversion_currency VARCHAR(3),
    converted_at TIMESTAMPTZ,
    processed_at TIMESTAMPTZ
);

-- Conversion links
CREATE TABLE bandit_conversion_links (
    pending_id UUID NOT NULL,
    transaction_id UUID NOT NULL,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pending_id, transaction_id)
);
```

---

## Section 5: Multi-Objective Hybrid System

### Purpose
Support multiple optimization objectives (conversion rate, LTV, normalized revenue) with configurable weighting.

### Components

1. **ObjectiveType** - Conversion, LTV, Revenue, or Hybrid
2. **HybridObjectiveStrategy** - Weighted combination of objectives

### Data Schema

```sql
-- Experiment objectives
ALTER TABLE ab_tests
ADD COLUMN objective_type VARCHAR(20) NOT NULL DEFAULT 'conversion',
ADD COLUMN objective_weights JSONB,
ADD COLUMN price_normalization BOOLEAN DEFAULT FALSE;

-- Per-objective statistics
CREATE TABLE bandit_arm_objective_stats (
    id BIGSERIAL PRIMARY KEY,
    arm_id UUID NOT NULL,
    objective_type VARCHAR(20) NOT NULL,
    alpha DECIMAL(10,2) DEFAULT 1.0,
    beta DECIMAL(10,2) DEFAULT 1.0,
    samples BIGINT DEFAULT 0,
    conversions BIGINT DEFAULT 0,
    total_revenue DECIMAL(18,2) DEFAULT 0,
    avg_ltv DECIMAL(12,2),
    UNIQUE(arm_id, objective_type)
);
```

### Score Calculation

- **Conversion**: Standard Thompson Sampling
- **LTV**: Expected Value = P(conversion) × AvgLTV
- **Revenue**: Normalized Revenue = P(conv) × (Revenue / Price)
- **Hybrid**: Weighted combination of above

---

## Section 6: Integration & Plugin System

### Main Orchestrator

```go
type AdvancedBanditEngine struct {
    base              *ThompsonSamplingBandit
    rewardStrategy    RewardStrategy
    selectionStrategy SelectionStrategy
    windowStrategy    WindowStrategy
    delayedStrategy   *DelayedRewardStrategy
    currencyService   *CurrencyRateService
}

func NewAdvancedBanditEngine(config ExperimentConfig) *AdvancedBanditEngine
```

### Configuration Per Experiment

```go
type ExperimentConfig struct {
    ID               uuid.UUID
    ObjectiveType    ObjectiveType
    ObjectiveWeights map[string]float64
    WindowConfig     *WindowConfig
    EnableContextual bool
    EnableDelayed    bool
    EnableCurrency   bool
}
```

---

## Section 7: Production Features

### Monitoring Metrics

```go
type BanditMetrics struct {
    Regret             float64  // Cumulative regret vs best arm
    ExplorationRate    float64  // How often we switch arms
    ConvergenceGAP     float64  // Performance gap best/worst
    BalanceIndex       float64  // Distribution evenness
    WindowUtilization  float64  // For sliding window
    PendingRewards     int64    // Unprocessed delayed rewards
}
```

### Warmup Strategies

- **Uniform**: Random selection until min_samples reached
- **Informed**: Use historical priors for initialization
- **Thompson**: Use bandit from start (default)

### Auto-rebalancing

Periodically check convergence and recommend stopping when winner exceeds confidence threshold.

---

## Section 8: Implementation Summary

### Files to Create

1. `internal/domain/service/currency_service.go`
2. `internal/domain/service/linucb_strategy.go`
3. `internal/domain/service/sliding_window_strategy.go`
4. `internal/domain/service/delayed_reward_strategy.go`
5. `internal/domain/service/hybrid_objective_strategy.go`
6. `internal/domain/service/advanced_bandit_engine.go`
7. `internal/worker/tasks/currency_jobs.go`
8. `internal/worker/tasks/bandit_maintenance_jobs.go`
9. `backend/migrations/017_create_bandit_advanced.up.sql`
10. `internal/interfaces/http/handlers/bandit_advanced.go`

### Files to Modify

1. `internal/domain/service/bandit_service.go` - Add plugin interfaces
2. `internal/infrastructure/persistence/repository/bandit_repository.go` - Add new methods
3. `cmd/api/main.go` - Wire up dependencies
4. `cmd/worker/main.go` - Register worker tasks

### Dependencies

- **Redis**: Sliding window (Sorted Sets), currency rates caching
- **External API**: ECB exchange rates (HTTPS)
- **Linear algebra**: Simple matrix operations for LinUCB

---

## Implementation Phases

**Phase 1: Core Infrastructure** (Week 1)
- Plugin interfaces and base orchestrator
- Database migrations
- Configuration system

**Phase 2: Currency Conversion** (Week 1)
- Currency rate service
- Conversion strategy
- Worker for rate updates

**Phase 3: Multi-Objective System** (Week 2)
- Objective stats tracking
- Hybrid scoring
- API endpoints

**Phase 4: Sliding Window** (Week 2)
- Window strategy implementation
- Redis Sorted Set operations
- Cleanup jobs

**Phase 5: Contextual Bandit** (Week 3)
- User context tracking
- LinUCB implementation
- Feature engineering

**Phase 6: Delayed Feedback** (Week 3)
- Pending reward queue
- Background processing
- Transaction linking

**Phase 7: Production Features** (Week 4)
- Monitoring and metrics
- Grafana dashboard
- Warmup strategies
- Auto-rebalancing

---

**END OF DESIGN DOCUMENT**
