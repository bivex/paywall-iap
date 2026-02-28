# Design: Implement All TODOs

**Date:** 2026-02-28
**Approach:** SQL-first (Approach B) — extend sqlc schema, generate typed code, implement all 17 TODOs

---

## Scope

17 TODO items across 6 files:

| File | TODOs |
|------|-------|
| `auth.go` | Refresh token validation + rotation |
| `admin.go` | Grant subscription, revoke subscription, list users, health check, add services |
| `webhook.go` | Stripe event storage, Apple JWS verification + storage, Google signature + storage |
| `tasks.go` | LTV update, analytics computation, webhook processing, notification sending, Lago sync, grace period expiration |
| `google_verifier.go` | Google Play Developer API integration (proper stub) |
| `logger.go` | Sentry SDK initialization (proper stub) |

---

## Architecture

### Database Layer (sqlc)

Extend `schema.sql` with three tables currently only in migrations:
- `webhook_events` — idempotent inbox for Stripe/Apple/Google events
- `analytics_aggregates` — pre-computed daily metrics (MRR, ARR, active count)
- `grace_periods` — failed renewal grace periods

New SQL queries added to existing query files:
- `users.sql`: `ListUsers`, `CountUsers`
- `transactions.sql`: `GetLTVByUserID` (SUM of successful amounts)
- New `webhook_events.sql`: `InsertWebhookEvent` (ON CONFLICT DO NOTHING), `GetUnprocessedWebhookEvents`, `MarkWebhookEventProcessed`
- New `analytics.sql`: `UpsertAnalyticsAggregate`
- New `grace_periods.sql`: `GetExpiredGracePeriods`, `UpdateGracePeriodStatus`, `GetActiveGracePeriodByUserID`

Run `sqlc generate` to produce typed Go code.

### Auth — Refresh Token Rotation

1. Parse refresh token string with `JWTMiddleware.ParseToken` → get `JTI`, `UserID`, `ExpiresAt`
2. Check Redis blocklist: if JTI blocked → 401
3. Issue new access token: `GenerateAccessToken(userID)` → `(accessToken, accessJTI, err)`
4. Rotate refresh token: `GenerateRefreshToken(userID)` → `(refreshToken, refreshJTI, err)`
5. Revoke old refresh token: `RevokeToken(ctx, oldJTI, remainingTTL)`
6. Return `RefreshTokenResponse{AccessToken, RefreshToken, ExpiresIn}`

### Admin Handler

Inject into `AdminHandler`:
- `subscriptionRepo domain.SubscriptionRepository`
- `userRepo domain.UserRepository`
- `dbPinger pool.Pinger` (for health)
- `redisPinger *redis.Client` (for health)

| Endpoint | Implementation |
|----------|---------------|
| `GrantSubscription` | Parse `ExpiresAt` string → time.Time; create `entity.Subscription`; call `subscriptionRepo.Create` |
| `RevokeSubscription` | Get active sub by userID; call `subscriptionRepo.Cancel` |
| `ListUsers` | Call `queries.ListUsers(ctx, limit, offset)`; return paginated JSON |
| `GetHealth` | Ping DB + Redis; return component statuses |

Update `NewAdminHandler` and `cmd/api/main.go` wiring.

### Webhooks — Idempotent Inbox

All three webhooks follow the same pattern:
1. Verify signature (provider-specific)
2. Parse `event_id` and `event_type` from body
3. `InsertWebhookEvent(provider, event_type, event_id, payload)` with `ON CONFLICT DO NOTHING`
4. Return 200 OK (idempotent — duplicate events are silently dropped)

**Stripe**: body is JSON with `id` (event_id) and `type` (event_type) fields
**Apple**: body is a JWS compact token; decode middle segment (payload) as base64 JSON to extract `notificationType` and `notificationUUID`; verify signature using `golang-jwt/jwt` with Apple's JWKS
**Google**: body is Pub/Sub push message; `message.data` is base64-encoded JSON with `subscriptionNotification.notificationType`; verify `Authorization` JWT bearer token is from Google

### Worker Tasks — Full DB-backed Logic

| Task | Implementation |
|------|---------------|
| **LTV Update** | `GetLTVByUserID(userID)` → sum; `UpdateUserLTV(userID, ltv)` |
| **Analytics** | For prior day: count active subs, sum revenue, compute MRR; `UpsertAnalyticsAggregate` per metric |
| **Webhook Processing** | `GetUnprocessedWebhookEvents()` → dispatch based on `provider+event_type`; `MarkWebhookEventProcessed` |
| **Grace Period Expiration** | `GetExpiredGracePeriods()` → for each: `CancelSubscription`, `UpdateGracePeriodStatus('expired')` |
| **Notification Sending** | Structured stub with FCM/APNs-ready interface; log intent, no external call |
| **Lago Sync** | Structured stub with Lago API-ready interface; log intent, no external call |

### External SDKs

**Sentry** (`github.com/getsentry/sentry-go`):
```go
sentry.Init(sentry.ClientOptions{
    Dsn:         cfg.DSN,
    Environment: cfg.Environment,
    Release:     cfg.Release,
})
defer sentry.Flush(2 * time.Second)
```
Attach `zapcore` Sentry hook so ERROR+ logs are captured.

**Google Play** (`golang.org/x/oauth2/google` + `google.golang.org/api/androidpublisher/v3`):
- If `serviceAccountJSON == ""`: return mock response (existing behaviour)
- Else: parse service account JSON, create OAuth2 client, call `androidpublisher.New(client).Purchases.Subscriptions.Get(packageName, productID, purchaseToken).Do()`
- Map response fields to `VerifyResponse`

---

## Dependency Changes

```
go get github.com/getsentry/sentry-go
go get golang.org/x/oauth2
go get google.golang.org/api/androidpublisher/v3
```

sqlc regenerated after schema + query additions.

---

## Error Handling

- All DB errors wrapped with `fmt.Errorf("context: %w", err)`
- Webhook signature failures → 401 (don't reveal details)
- Admin operations on non-existent users/subs → 404
- Grace period task errors logged, continue to next record (don't fail entire batch)

---

## What Is NOT Implemented

- Lago sync: kept as a structured, logged stub (no Lago credentials configured)
- Push notifications: kept as a structured, logged stub (no FCM/APNs credentials configured)
- Apple JWS full certificate chain validation (uses header kid lookup, not pinned cert)
