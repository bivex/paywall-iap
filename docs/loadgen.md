# Load Generator

`backend/cmd/loadgen` — subscription load generator that drives the full
registration + IAP verify flow against live mock servers.

## What it does

For each simulated user:
1. POST /v1/auth/register  — creates a real user in the DB
2. Obtains a JWT access token from the response
3. For Android — builds a valid JSON receipt with a random purchase token
   and one of the Google mock scenarios
4. For iOS — calls the Apple mock POST /subs to create a real receipt token,
   then uses it in the verify call
5. POST /v1/verify/iap  — verifies the receipt, persists subscription

## Prerequisites

Both mock servers must be running:

```
apple-iap-mock  :9090   (docker-compose.local.yml or tests/apple-of-my-iap)
google-billing-mock :8090  (docker-compose.local.yml or tests/google-billing-mock)
API             :8081
```

Start the full local stack:

```bash
docker compose -f infra/docker-compose/docker-compose.dev.yml up -d
```

## Usage

```bash
cd backend

# 200 users, 20 workers, mixed iOS + Android
go run ./cmd/loadgen \
  --app-id 8ef6793c-b439-404f-93d6-4663e0b12b78 \
  --users 200 \
  --workers 20 \
  --platform mixed

# Android only, 500 users
go run ./cmd/loadgen \
  --app-id 8ef6793c-b439-404f-93d6-4663e0b12b78 \
  --users 500 \
  --workers 50 \
  --platform android

# Dry-run (prints requests, no network calls)
go run ./cmd/loadgen \
  --app-id 8ef6793c-b439-404f-93d6-4663e0b12b78 \
  --users 10 \
  --dry-run
```

## Flags

| Flag           | Default                    | Description                          |
|----------------|----------------------------|--------------------------------------|
| --app-id       | (required)                 | App UUID sent as X-App-ID header     |
| --api          | http://localhost:8081      | API base URL                         |
| --apple-mock   | http://localhost:9090      | Apple IAP mock server URL            |
| --users        | 100                        | Total users to generate              |
| --workers      | 10                         | Concurrent goroutines                |
| --platform     | mixed                      | ios / android / mixed                |
| --dry-run      | false                      | Print requests without sending       |

## Scenarios

Android (Google mock):

| Scenario       | Token prefix              | Expected result |
|----------------|---------------------------|-----------------|
| active_monthly | valid_active_user_123     | 200 active      |
| active_annual  | valid_active_annual_456   | 200 active      |
| expired        | expired_user_789          | 422             |
| canceled_active| canceled_active_user_001  | 200 active      |
| canceled_user  | canceled_user_002         | 422             |
| pending        | pending_user_003          | 422             |

iOS (Apple mock):
Each user gets a fresh receipt created via POST /subs against the Apple mock.
Product IDs: `com.mothsalt.game1.premium.monthly`, `com.mothsalt.game1.premium.yearly`.
All Apple mock receipts return `active`.

## Sample output

```
loadgen: 200 users, 20 workers, platform=mixed, api=http://localhost:8081
[user 2] ios ios_sub → active (http 200)
[user 14] android active_monthly → active (http 200)
[user 0] android expired →  (http 422)
...

========== LOAD GEN RESULTS ==========
Duration:       209ms
Registered:     200 ok / 0 failed
IAP verified:   200 ok / 0 failed
  active:       141
  expired:      0
  canceled:     0
  pending:      0
Throughput:     1914.1 req/s
=======================================
```

422 responses for expired/pending/canceled scenarios are expected — the API
correctly rejects invalid receipts. They are not counted as failures.
