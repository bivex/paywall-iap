# Architecture

## Overview

Paywall-IAP is a multi-tenant in-app purchase backend with a Go API, async worker,
Next.js admin dashboard, PostgreSQL, and Redis.

```
Mobile SDK / Browser
      |
      v
  [Frontend]          Next.js 16 (Turbopack)   :3000
      |                admin dashboard
      v
  [API]               Go / Gin                 :8081 (host) → :8080 (container)
      |                REST, JWT auth, IAP verify
      |
      +──────────────> [Worker]    Go / Asynq   (background tasks)
      |                             dunning, notifications, Lago sync
      |
      +──────────────> [PostgreSQL]             :5432
      |                             main DB, migrations via sqlc
      |
      +──────────────> [Redis]                  :6379
                                    JWT blocklist, rate limit, Asynq queue

External services (real credentials via env):
  - Apple IAP          APPLE_MOCK_URL (dev mock :9090)
  - Google Play        GOOGLE_IAP_BASE_URL (dev mock :8090)
  - SendGrid           SENDGRID_API_KEY
  - FCM                FCM_SERVER_KEY
  - Lago               LAGO_API_URL / LAGO_API_KEY
  - Stripe / Paddle    webhook only (no outbound)
  - Sentry             SENTRY_DSN
```

## Services

| Service            | Port  | Image / Source                        |
|--------------------|-------|---------------------------------------|
| frontend           | 3000  | frontend/docker-compose.dev.yml       |
| api                | 8081  | backend/cmd/api                       |
| worker             | —     | backend/cmd/worker                    |
| migrator           | —     | backend/cmd/migrator                  |
| postgres           | 5432  | docker-compose.dev.yml                |
| redis              | 6379  | docker-compose.dev.yml                |
| apple-iap-mock     | 9090  | tests/apple-of-my-iap                 |
| google-billing-mock| 8090  | tests/google-billing-mock             |
| js-error-collector | 8088  | infra/                                |

## Backend layers

```
cmd/api          — Gin router, DI wiring, graceful shutdown
cmd/worker       — Asynq server, task handler registration
cmd/seed         — Admin user seeder
cmd/loadgen      — Subscription load generator (see docs/loadgen.md)

internal/
  domain/
    entity/          — pure structs (User, Subscription, Transaction …)
    service/         — business logic (LTV, dunning, notification …)
    repository/      — interfaces only
  application/
    command/         — use-cases (VerifyIAP, RegisterUser …)
    dto/             — request/response structs
  infrastructure/
    config/          — viper config + env bindings
    persistence/     — sqlc-generated queries, repository impls
    iap/             — Apple / Google IAP clients
  interfaces/
    http/
      handlers/      — Gin handlers (admin, auth, IAP, webhook …)
      middleware/    — JWT, rate limit, CORS
  worker/
    tasks/           — Asynq task handlers (dunning, notifications, Lago)
```

## Multi-tenancy

Each app is identified by `X-App-ID` header (UUID). Apps are rows in the `apps`
table. Settings, credentials, pricing tiers, experiments, and subscriptions are
all scoped per app.

Credentials (Apple, Google, Stripe, Paddle) are encrypted at rest with AES-256-GCM
using `APP_CREDENTIALS_KEY` (must be exactly 32 bytes). If the key is absent the
system runs in dev mode without encryption.

See [docs/multi-tenancy.md](multi-tenancy.md) for full details.

## Auth flow

```
POST /v1/auth/register   — device registers, gets JWT access + refresh tokens
POST /v1/auth/refresh    — refresh access token
POST /v1/admin/auth/login  — admin login (bcrypt), sets httpOnly cookies
POST /v1/admin/auth/logout — invalidates token in Redis blocklist
```

JWT access tokens expire in 15 min. Refresh tokens expire in 30 days.

## IAP verify flow

```
Client → POST /v1/verify/iap
  {platform, receipt_data, product_id, transaction_id}

  Android: receipt_data must be JSON
    {"packageName","productId","purchaseToken","type":"subscription"}

  iOS: receipt_data is the receipt token from Apple mock POST /subs
       or real App Store receipt (base64)

API → calls Apple/Google mock (dev) or real store API (prod)
    → persists subscription in DB
    → returns {subscription_id, status, expires_at}
```

## Task queue (Asynq)

Workers registered in `cmd/worker/main.go`:

| Task                    | Handler                  | Description                     |
|-------------------------|--------------------------|---------------------------------|
| subscription:expire     | HandleSubscriptionExpiry | mark expired subs               |
| subscription:renew      | HandleSubscriptionRenewal| attempt renewal                 |
| dunning:retry_payment   | HandleDunningRetry       | retry failed payments           |
| dunning:all_retries_failed | HandleAllRetriesFailed | final dunning notification      |
| notification:send       | HandleSendNotification   | FCM push via config credentials |
| lago:sync               | HandleSyncLago           | Lago billing REST sync          |

## Database

Migrations: `backend/migrations/*.up.sql`, applied by migrator container on startup.
Queries: generated by sqlc, source in `backend/internal/infrastructure/persistence/queries/`.
