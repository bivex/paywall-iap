# App-ID Multi-Tenancy Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Propagate `app_id` from the frontend AppSelector through every admin API route to the backend, so all admin pages (users, subscriptions, transactions, experiments, pricing, winback, analytics) show data scoped to the selected app only.

**Architecture:**
- Frontend stores `selectedAppId` in Zustand (already done). All client-side fetches pass it as `?app_id=<uuid>` query param to Next.js API route handlers. Next.js proxy routes forward it to the backend as `X-App-ID` header (already trusted, avoids polluting URL on backend). Backend admin handlers read `X-App-ID`, validate it is a known UUID, and filter all DB queries by it.

**Tech Stack:** Next.js 16 (Turbopack), Zustand, Go (gin), sqlc, PostgreSQL

---

## Current State

- `selectedAppId` persisted in localStorage via Zustand store (`src/stores/app-store.ts`)
- AppSelector dropdown in sidebar — works, auto-selects first active app
- Backend `apps` table exists, `app_id` FK column on: users, subscriptions, pricing_tiers, ab_tests, webhook_events, bandit_user_context
- sqlc queries for users/subscriptions already have `WHERE app_id = $1`
- Admin handlers (pricing, experiments, winback, settings) do NOT yet filter by app_id
- Frontend route handlers do NOT yet forward app_id

## Scope

Pages that need app-scoped data:
1. User List (`/dashboard/users`) — users table
2. User 360 (`/dashboard/users/[id]`) — users + subscriptions
3. Subscriptions (`/dashboard/subscriptions`) — subscriptions table
4. Transactions (`/dashboard/transactions`) — transactions table
5. Experiments / A-B Tests (`/dashboard/experiments`) — ab_tests table
6. Pricing Tiers (`/dashboard/pricing`) — pricing_tiers table
7. Winback (`/dashboard/winback`) — dunning/winback table
8. Bandit Dashboard — bandit_user_context table
9. Analytics — analytics_aggregates table

Pages that do NOT need app_id (global):
- Admin settings, audit log, platform settings, webhooks config

---

## Task 1: Frontend — shared `apiFetch` utility with app_id injection

**Objective:** Single helper that all client components use for fetch; automatically appends `?app_id=` from Zustand store so no page has to do it manually.

**Files:**
- Create: `frontend/src/lib/api-fetch.ts`

**Step 1: Write the utility**

```ts
// frontend/src/lib/api-fetch.ts
import { useAppStore } from "@/stores/app-store";

export function buildUrl(path: string): string {
  const { selectedAppId } = useAppStore.getState();
  if (!selectedAppId) return path;
  const sep = path.includes("?") ? "&" : "?";
  return `${path}${sep}app_id=${encodeURIComponent(selectedAppId)}`;
}

export async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
  return fetch(buildUrl(path), init);
}
```

**Step 2: Smoke test in browser console**
Open DevTools → Console on any dashboard page:
```js
import("/api/admin/apps").then(console.log)
// Just verifying module loads — no runtime error
```

**Step 3: Commit**
```bash
git add frontend/src/lib/api-fetch.ts
git commit -m "feat(frontend): add apiFetch utility with app_id injection"
```

---

## Task 2: Frontend — Next.js route handlers forward X-App-ID to backend

**Objective:** All `/api/admin/*` route handlers read `app_id` from query params and forward to backend as `X-App-ID` header.

**Files:**
- Create: `frontend/src/lib/server/backend-fetch.ts` — shared server-side fetch helper

**Step 1: Write the server helper**

```ts
// frontend/src/lib/server/backend-fetch.ts
import { cookies } from "next/headers";
import { type NextRequest } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export async function backendFetch(
  path: string,
  req: NextRequest | null,
  init?: RequestInit
): Promise<Response> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) throw new Error("UNAUTHORIZED");

  const appId = req?.nextUrl?.searchParams?.get("app_id") ?? null;

  const headers: Record<string, string> = {
    ...(init?.headers as Record<string, string>),
    Authorization: `Bearer ${token}`,
  };
  if (appId) headers["X-App-ID"] = appId;

  return fetch(`${BACKEND_URL}${path}`, { ...init, headers, cache: "no-store" });
}
```

**Step 2: Update each route handler to use backendFetch**

For every file in `frontend/src/app/api/admin/` (except `apps/`), replace the inline fetch pattern:

Before:
```ts
const token = await getToken();
const res = await fetch(`${BACKEND_URL}/v1/admin/...`, {
  headers: { Authorization: `Bearer ${token}` },
});
```

After:
```ts
const res = await backendFetch("/v1/admin/...", req, {});
```

Route handlers that need updating (check each file):
- `bandit/dashboard/route.ts`
- `bandit/snapshot/route.ts`
- `delayed-feedback/*/route.ts` (5 files)
- `multi-objective/*/route.ts` (3 files)
- `sliding-window/*/route.ts` (4 files)
- `studio/*/route.ts` (2 files)

Note: `apps/route.ts` is global (no app_id filter needed — it lists all apps).

**Step 3: Verify route still works after refactor**
```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/admin/bandit/dashboard
# expect 401 (no cookie) — means route handler loaded correctly
```

**Step 4: Commit**
```bash
git add frontend/src/lib/server/backend-fetch.ts frontend/src/app/api/admin/
git commit -m "feat(frontend): forward X-App-ID header from all admin API route handlers"
```

---

## Task 3: Frontend — replace fetch calls in client pages with apiFetch

**Objective:** Every client page that currently calls `fetch("/api/admin/...")` directly uses `apiFetch` instead so app_id is appended automatically.

**Files to update** (search `fetch("/api/admin` in src):
```bash
grep -rn 'fetch("/api/admin' frontend/src/app --include="*.tsx" --include="*.ts" -l
```

For each file, replace:
```ts
fetch("/api/admin/bandit/dashboard")
```
with:
```ts
import { apiFetch } from "@/lib/api-fetch";
apiFetch("/api/admin/bandit/dashboard")
```

Also update pages in `apps-page-client.tsx` — but NOT the `/api/admin/apps` calls (those are global).

**Step 2: Verify AppSelector change triggers re-fetch**

Add a `useEffect` dependency on `selectedAppId` in each page that fetches on mount. Pattern:

```ts
const selectedAppId = useAppStore((s) => s.selectedAppId);

useEffect(() => {
  apiFetch("/api/admin/...").then(...)
}, [selectedAppId]); // re-fetch when app changes
```

**Step 3: Commit**
```bash
git add frontend/src/app
git commit -m "feat(frontend): use apiFetch in all client pages, re-fetch on app change"
```

---

## Task 4: Backend — app_id middleware for admin routes

**Objective:** Extract `X-App-ID` from request header, validate it is a valid UUID of a known app, store in gin context so handlers can read it without boilerplate.

**Files:**
- Create: `backend/internal/interfaces/http/middleware/app_id_middleware.go`

**Step 1: Write middleware**

```go
// backend/internal/interfaces/http/middleware/app_id_middleware.go
package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

const AppIDKey = "app_id"

// RequireAppID reads X-App-ID header, validates UUID, stores in context.
// Returns 400 if missing or malformed, 422 if not a valid UUID.
func RequireAppID() gin.HandlerFunc {
    return func(c *gin.Context) {
        raw := c.GetHeader("X-App-ID")
        if raw == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "X-App-ID header required"})
            c.Abort()
            return
        }
        id, err := uuid.Parse(raw)
        if err != nil {
            c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid X-App-ID"})
            c.Abort()
            return
        }
        c.Set(AppIDKey, id)
        c.Next()
    }
}

// GetAppID retrieves the validated app UUID from gin context.
func GetAppID(c *gin.Context) uuid.UUID {
    return c.MustGet(AppIDKey).(uuid.UUID)
}
```

**Step 2: Write test**

```go
// backend/internal/interfaces/http/middleware/app_id_middleware_test.go
func TestRequireAppID_Missing(t *testing.T) { ... }
func TestRequireAppID_Invalid(t *testing.T) { ... }
func TestRequireAppID_Valid(t *testing.T) { ... }
```

**Step 3: Build check**
```bash
cd backend && go build ./...
# expect: no errors
```

**Step 4: Commit**
```bash
git add backend/internal/interfaces/http/middleware/
git commit -m "feat(backend): app_id middleware — validate X-App-ID header, store in gin context"
```

---

## Task 5: Backend — apply middleware to app-scoped admin route groups

**Objective:** Register `RequireAppID` on all admin routes that need app scoping. Routes that are global (auth, settings, apps CRUD) are excluded.

**Files:**
- Modify: `backend/cmd/api/main.go` (or wherever `setupAdminRoutes` lives)

**Step 1: Find route registration**
```bash
grep -n "setupAdminRoutes\|adminGroup\|v1/admin" backend/cmd/api/main.go | head -30
```

**Step 2: Create sub-group with middleware**

```go
// app-scoped routes
appScoped := adminGroup.Group("/")
appScoped.Use(middleware.RequireAppID())
{
    appScoped.GET("/users", deps.usersHandler.List)
    appScoped.GET("/users/:id", deps.usersHandler.Get)
    appScoped.GET("/subscriptions", deps.subscriptionsHandler.List)
    appScoped.GET("/transactions", deps.transactionsHandler.List)
    appScoped.GET("/pricing-tiers", deps.pricingHandler.List)
    appScoped.POST("/pricing-tiers", deps.pricingHandler.Create)
    // ... etc for experiments, winback, bandit, analytics
}
```

**Step 3: Build check**
```bash
cd backend && go build ./...
```

**Step 4: Integration smoke test**
```bash
# Without X-App-ID — expect 400
curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8081/v1/admin/users

# With X-App-ID — expect 200
curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-App-ID: 00000000-0000-0000-0000-000000000001" \
  http://localhost:8081/v1/admin/users
```

**Step 5: Commit**
```bash
git add backend/cmd/api/main.go
git commit -m "feat(backend): apply RequireAppID middleware to app-scoped admin route group"
```

---

## Task 6: Backend — filter users and user-360 by app_id

**Objective:** `GET /v1/admin/users` and `GET /v1/admin/users/:id` filter by app_id from context.

**Files:**
- Modify: handler file for admin users (find with `grep -rn "admin/users\|ListUsers" backend/`)

**Step 1: Read current handler**
```bash
grep -rn "func.*List\|func.*Get" backend/internal/interfaces/http/handlers/ | grep -i user
```

**Step 2: Update List handler**

```go
func (h *UsersHandler) List(c *gin.Context) {
    appID := middleware.GetAppID(c)
    users, err := h.repo.ListByApp(c.Request.Context(), appID)
    // ...
}
```

**Step 3: Verify sqlc query exists**
Check `backend/internal/infrastructure/persistence/sqlc/queries/users.sql` for:
```sql
-- name: ListUsersByApp :many
SELECT ... FROM users WHERE app_id = $1 AND deleted_at IS NULL
```
If missing, add it and re-run `sqlc generate`.

**Step 4: Build + smoke test**
```bash
cd backend && go build ./... && go vet ./...
```

**Step 5: Commit**
```bash
git commit -m "feat(backend): filter admin users list by X-App-ID"
```

---

## Task 7: Backend — filter subscriptions and transactions by app_id

Same pattern as Task 6 for:
- `GET /v1/admin/subscriptions`
- `GET /v1/admin/transactions`

Both tables have `app_id` column and sqlc queries with `WHERE app_id = $1`.

**Commit:**
```bash
git commit -m "feat(backend): filter admin subscriptions and transactions by X-App-ID"
```

---

## Task 8: Backend — filter pricing tiers, experiments, winback by app_id

Same pattern for:
- `admin_pricing.go` — pricing_tiers
- `admin_experiments.go` — ab_tests
- `admin_winback.go` — winback/dunning table

Check each handler for the sqlc query name used, add `app_id` param.

**Commit:**
```bash
git commit -m "feat(backend): filter pricing, experiments, winback by X-App-ID"
```

---

## Task 9: Backend — filter bandit and analytics by app_id

- Bandit handlers — bandit_user_context (has app_id FK)
- Analytics aggregates — analytics_aggregates (has app_id column)

**Commit:**
```bash
git commit -m "feat(backend): filter bandit and analytics by X-App-ID"
```

---

## Task 10: Frontend — show app name in page headers

**Objective:** Each scoped page shows "Users — Mothsalt Game 4" so admin knows which app they're viewing.

**Files:**
- Create: `frontend/src/components/app-scope-badge.tsx` — small badge showing selected app display_name
- Modify: page header components for users, subscriptions, transactions, experiments, pricing

**Step 1: Write badge component**

```tsx
// frontend/src/components/app-scope-badge.tsx
"use client";
import { useAppStore, getSelectedApp } from "@/stores/app-store";
import { Badge } from "@/components/ui/badge";

export function AppScopeBadge() {
  const app = useAppStore(getSelectedApp);
  if (!app) return null;
  return <Badge variant="outline">{app.display_name}</Badge>;
}
```

**Step 2: Add to page headers**

In each page's `<h1>` section:
```tsx
<div className="flex items-center gap-2">
  <h1 className="text-xl font-semibold">Users</h1>
  <AppScopeBadge />
</div>
```

**Step 3: Commit**
```bash
git commit -m "feat(frontend): show selected app badge in scoped page headers"
```

---

## Task 11: Frontend — guard pages when no app selected

**Objective:** If `selectedAppId` is null (no apps exist yet), show a prompt to create an app instead of an empty/broken table.

**Files:**
- Create: `frontend/src/components/no-app-selected.tsx`

```tsx
export function NoAppSelected() {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-4 text-center">
      <p className="text-muted-foreground text-sm">No app selected.</p>
      <Button asChild variant="outline" size="sm">
        <a href="/dashboard/apps">Go to Apps</a>
      </Button>
    </div>
  );
}
```

Add at top of each scoped page client component:
```tsx
const selectedAppId = useAppStore((s) => s.selectedAppId);
if (!selectedAppId) return <NoAppSelected />;
```

**Commit:**
```bash
git commit -m "feat(frontend): show no-app-selected guard on scoped pages"
```

---

## Risks and open questions

1. **sqlc regeneration** — Tasks 6-9 may require adding new queries and running `sqlc generate`. Check if sqlc is available in the dev environment: `which sqlc`. If not, write raw Go SQL instead of using generated code.

2. **Backward compatibility** — After Task 5, existing frontend pages that don't yet pass `X-App-ID` will get 400. Implement Tasks 3 and 5 in the same deploy, or make the middleware return a fallback (legacy sentinel `00000000-0000-0000-0000-000000000001`) instead of 400 during migration. Recommended: fallback during migration, strict 400 after all pages are updated.

3. **Bandit/delayed-feedback handlers** — These are complex multi-file handlers. Read them carefully before touching; they may have their own repository pattern that differs from users/subscriptions.

4. **Analytics aggregates** — `app_id` column exists but may be nullable on old rows. Verify with `SELECT COUNT(*) FROM analytics_aggregates WHERE app_id IS NULL` before enforcing NOT NULL filter.

5. **No middleware.ts** — `proxy.ts` is not wired as Next.js middleware (it exports `proxy` not default). This is fine — the `X-App-ID` forwarding happens inside individual route handlers via `backendFetch`, not at middleware level.
