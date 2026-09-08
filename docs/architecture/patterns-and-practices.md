# Backend Architecture Patterns & Engineering Practices

## 1. Architectural Style: Clean Architecture & DDD

The backend service adheres to Clean Architecture and Domain-Driven Design (DDD) principles:

```
[cmd/api, cmd/worker]
       │
       ▼
[internal/interfaces/http] (HTTP Handlers, Middleware, DTOs)
       │
       ▼
[internal/application]    (Commands, Queries)
       │
       ▼
[internal/domain]         (Entities, Value Objects, Domain Services, Repository Interfaces)
       ▲
       │
[internal/infrastructure] (SQLC Persistence, Redis Cache, External IAP / Matomo Adapters)
```

### Layer Constraints & Rules
- **Domain Layer (`internal/domain`)**: Pure Go business logic. Defines domain entities, value objects, domain services, and repository contracts (interfaces). Has zero dependencies on HTTP frameworks, SQL drivers, or external clients.
- **Application Layer (`internal/application`)**: Orchestrates use cases via the **Command Pattern** (`Execute(ctx, req)`) and **Query Pattern**.
- **Infrastructure Layer (`internal/infrastructure`)**: Implements repository interfaces using `sqlc` (`pgxpool`), Redis caching, and external billing/analytics providers.
- **Interfaces Layer (`internal/interfaces`)**: Gin HTTP handlers, routes, and middleware adapters.
- **Worker Layer (`internal/worker/tasks`)**: Asynq background task handlers and distributed workers.

---

## 2. Active Design Patterns

DPX static pattern analysis monitors and verifies design pattern compliance across the entire codebase.

### 2.1 Command Pattern (`application/command`)
Encapsulates executable write operations into self-contained use cases.
- `RegisterCommand` — User registration and auth token provisioning.
- `VerifyIAPCommand` — Platform receipt validation, subscription upsert, and analytics forwarding.
- `ResolveGracePeriodCommand`, `CreateGracePeriodCommand`, `CancelSubscriptionCommand`, etc.

**Signature standard:**
```go
type Command interface {
    Execute(ctx context.Context, req *dto.Request) (*dto.Response, error)
}
```

### 2.2 Strategy Pattern (`domain/service`)
Used for algorithmic flexibility in Multi-Armed Bandit (MAB) optimization and pricing exploration:
- `SelectionStrategy`: Concrete strategies include `ThompsonSamplingBandit` (Bayesian Beta distribution) and `LinUCBSelectionStrategy` (contextual bandits).
- `RewardStrategy`: Concrete strategies include `DelayedRewardStrategy` (two-phase conversion), `HybridObjectiveStrategy` (multi-objective optimization), and `CurrencyConversionRewardStrategy` (multi-currency normalization).
- `WindowStrategy`: `SlidingWindowStrategy` for time-decayed sample windows.

### 2.3 Factory Method Pattern (`Factory Method: 105 instances`)
All services, handlers, repositories, value objects, and entities expose constructor functions (`New<Component>`), enforcing invariant validation at creation time:
- `entity.NewUser(p NewUserParams) *User`
- `entity.NewSubscription(p NewSubscriptionParams) *Subscription`
- `entity.NewWinbackOffer(p NewWinbackOfferParams) *WinbackOffer`
- `service.NewThompsonSamplingBandit(...)`

### 2.4 Struct Embedding & Composition Over Inheritance (`14 instances`)
To eliminate "God Structs", monolithic structures were decomposed into focused components and unified via Go anonymous struct embedding:
- `TaskHandlers` embeds `WebhookTaskHandler`, `LTVTaskHandler`, `AnalyticsTaskHandler`, etc.
- `PostgresBanditRepository` embeds `PostgresBanditArmRepository`, `PostgresBanditAssignmentRepository`, and `PostgresBanditConversionRepository`.
- `AdminHandler` embeds `AdminSubscriptionHandler`, `AdminUserHandler`, and `AdminMetricsHandler`.
- `AdvancedBanditEngine` embeds `BanditRewardEngine`, `BanditWindowEngine`, and `BanditObjectiveEngine`.

### 2.5 Facade Pattern (`11 instances`)
High-level facades provide a cohesive entry point for complex multi-engine subsystems:
- `AdvancedBanditEngine` serves as the facade for selection, reward processing, sliding windows, and objective sync.
- `LTVService` coordinates transactions, cohort aggregates, and predictive models.
- `MatomoClient` wraps HTTP serialization, retry policies, and tracking endpoints.

---

## 3. SOLID & Clean Code Adherence

### 3.1 Single Responsibility Principle (SRP)
- Structs are constrained to single domain responsibilities.
- Monolithic structs with high method/field counts were split into modular components.
- **Current SRP violations:** `0` (clean).

### 3.2 Open-Closed Principle (OCP)
- MAB engines and subscription processors are open for extension via interfaces (`SelectionStrategy`, `RewardStrategy`, `DelayedConversionProcessor`) and closed for modification.

### 3.3 Interface Segregation Principle (ISP)
- Repository interfaces are broken down into granular consumer-driven roles:
  - `UserReader`, `UserWriter`, `UserPurchaseChannelUpdater`
  - `SubscriptionReader`, `SubscriptionStatusUpdater`, `SubscriptionExpiryUpdater`
- Fat interfaces (such as sqlc's monolith) are composed from smaller contracts.

### 3.4 Dependency Inversion Principle (DIP)
- Functions and constructors depend on interface abstractions rather than concrete pointers.
- Verified by DPX: **46 clean DIP adherence points** across core domain services.

### 3.5 KISS & Parameter Objects
- Functions avoid long parameter lists (≥ 5 arguments).
- Functions with multiple parameters are encapsulated in typed parameter objects:
  - `NewUserParams`, `NewSubscriptionParams`, `NewWinbackOfferParams`, `NewTransactionParams`
  - `VerifyIAPCommandParams`, `MatomoTrackEventParams`, `TrackImpressionParams`, `RewardWithEventParams`
- Benefits: prevents argument ordering bugs, self-documenting code, zero-cost backward-compatible extension.
- **Current KISS violations:** `0` (clean).

---

## 4. Concurrency & Reliability Guarantees

### 4.1 Context Propagation (`484 instances`)
- Every I/O bound function, repository method, and service call receives `ctx context.Context` as its first parameter.
- Timeouts and cancellations propagate cleanly from HTTP middleware and Asynq workers down to PostgreSQL queries (`pgxpool`) and Redis calls.

### 4.2 Distributed Locking
- Subscription state mutations and concurrent webhook/receipt verifications acquire Redis distributed locks (`acquireSubscriptionLock`) to prevent double-spending or conflicting state transitions.

### 4.3 Database Transactions & Pessimistic Locking
- Multi-step state transitions (such as pending bandit reward finalization) execute in database transactions with `SELECT ... FOR UPDATE` row-level locks to eliminate race conditions.

### 4.4 Background Job Isolation
- Heavy computational tasks (cohort aggregation, LTV recalculation, Matomo fallback batch flushing, expired offer reconciliation) run asynchronously via Asynq worker pools.

---

## 5. Architectural Audits & Tooling

To run the automated DPX pattern detection scan:

```bash
# Run DPX scan excluding tests, outputting JSON, HTML dashboard, and Markdown
/Volumes/External/Code/DPX-Go/.venv/bin/python -m pattern_detector.adapters.inbound.cli.main scan \
  backend \
  --no-tests \
  -J docs/reports/backend_dpx.json \
  -H docs/reports/backend_dpx.html \
  -M docs/reports/backend_dpx.md
```

### Current Metric Snapshot
- **Files Scanned:** `148`
- **Total Architecture Findings:** `676`
- **Violations / Smells:** `0` (0 God Structs, 0 Long Parameter Lists, 0 High Coupling)
- **Active Patterns:** `630`
- **Clean Inversions (DIP):** `46`
