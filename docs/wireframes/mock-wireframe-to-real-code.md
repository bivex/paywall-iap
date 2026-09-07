# 🚀 MockWireframeToRealCode.md — Путь от Wireframe до Production

> Этот документ описывает весь путь от StarUML wireframes и мок-данных до полностью работающего Admin Dashboard с реальной БД, реальным API и реальным фронтендом. Каждый раздел — отдельный урок.

---

## 📑 Содержание

1. [Старт: от .mdj к реальному проекту](#1-старт)
2. [Google IAP Mock: git submodule + Docker](#2-google-iap-mock)
3. [Интеграционные тесты с мок-сервером](#3-интеграционные-тесты)
4. [Admin Dashboard UI: shadcn MCP подход](#4-admin-dashboard-ui)
5. [Замена mock-данных на реальный API](#5-реальный-api)
6. [Database Seeding: реалистичные тестовые данные](#6-database-seeding)
7. [Docker: порты, сети и compose](#7-docker)
8. [Миграции: dirty flag и порядок запуска](#8-миграции)
9. [Компонентный апгрейд: Table vs список](#9-компоненты)
10. [Ключевые технические инсайты](#10-инсайты)
11. [Архитектура финального результата](#11-архитектура)

---

## 1. Старт

### Откуда начинали

В `docs/` уже лежал `Wireframes_Rethink.md` — подробный spec с ASCII-макетами каждого экрана. В StarUML был `.mdj` файл на 508 виджетов и 20 экранов. Но всё это было **статической документацией** — ни один экран не был реализован.

```
docs/Wireframes_Rethink.mdj   → 508 виджетов, 20 экранов (только картинки)
backend/                       → Go API, реальная БД, но /dashboard/metrics возвращал заглушки
frontend/                      → Next.js, страница /dashboard/default с захардкоженными числами
```

### Цель

Пройти весь путь: **Wireframe → Компонент → Реальные данные → Работающий Dashboard**.

---

## 2. Google IAP Mock: git submodule + Docker

### Задача

Тестировать Google In-App Purchase верификацию без реального Google API.

### Решение: git submodule

```bash
git submodule add https://github.com/bivex/google-billing-mock tests/google-billing-mock
git submodule update --init --recursive
```

**Урок**: submodule — правильный способ подключать внешние сервисы-заглушки. Они версионируются вместе с проектом, но обновляются независимо.

### Dockerfile для mock-сервера

```dockerfile
# tests/google-billing-mock/deploy/Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mock-server ./cmd/server

FROM alpine:3.19
COPY --from=builder /app/mock-server /mock-server
EXPOSE 8080
CMD ["./mock-server"]
```

### docker-compose интеграция

```yaml
# infra/docker-compose/docker-compose.local.yml
google-billing-mock:
  build:
    context: ../../tests/google-billing-mock
    dockerfile: deploy/Dockerfile
  ports:
    - "8090:8080"
  healthcheck:
    test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
    interval: 5s
    retries: 5
```

**Урок**: мок-сервис — полноправный контейнер в compose. API-сервис указывает на него через `GOOGLE_IAP_BASE_URL=http://google-billing-mock:8080` (внутренний DNS Docker).

### Изменение в Go-верификаторе

```go
// internal/infrastructure/external/iap/google_verifier.go
baseURL := os.Getenv("GOOGLE_IAP_BASE_URL") // "http://google-billing-mock:8080" в dev
// В prod: реальный Google API endpoint
```

---

## 3. Интеграционные тесты с мок-сервером

### Проблема: build failed

```
# github.com/bivex/paywall-iap/tests/testutil
tests/testutil/test_server.go:48:54: not enough arguments in call to handlers.NewAuthHandler
        have (*command.RegisterCommand, *middleware.JWTMiddleware)
        want (*command.RegisterCommand, *command.AdminLoginCommand, *middleware.JWTMiddleware)
```

**Урок**: при добавлении новых зависимостей в хендлеры — обязательно обновлять тестовые хелперы (`testutil/test_server.go`). Тесты — первые, кто сломается при API-изменениях.

### Fix

```go
// tests/testutil/test_server.go
authHandler := handlers.NewAuthHandler(registerCmd, adminLoginCmd, jwtMiddleware)
//                                                  ↑ добавили
```

### Запуск интеграционных тестов

```bash
cd backend && GOOGLE_IAP_BASE_URL=http://localhost:8090 \
  go test -tags=integration -run TestGoogleMock ./tests/integration/ -v
```

**Урок**: интеграционные тесты запускать с `build tag` (`-tags=integration`) — они не должны попадать в обычный `go test ./...`.

---

## 4. Admin Dashboard UI: shadcn MCP подход

### Проблема: какой компонент выбрать?

Вместо того чтобы гуглить или угадывать — использовали **shadcn MCP** для поиска паттернов:

```
shadcn-get_item_examples_from_registries("chart-area-interactive")
→ точный код AreaChart с Select toggle, ChartContainer/ChartTooltip/ChartLegend

shadcn-get_item_examples_from_registries("chart-pie-donut-active")
→ точный код PieChart donut с активным Sector и center Label
```

**Урок**: shadcn MCP возвращает **рабочий код с реальными import paths** (`@/components/ui/chart`, не `recharts` напрямую). Копировать оттуда безопаснее, чем с docs.shadcn.ui — там могут быть устаревшие примеры.

### MRR Trend Chart

```tsx
// _components/mrr-trend-chart.tsx
// Паттерн: chart-area-interactive
import { AreaChart, Area, XAxis, CartesianGrid } from "recharts";
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart";
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select";

// Props вместо хардкода:
export function MrrTrendChart({ data, activeSubs }: {
  data: MonthlyMRR[];
  activeSubs: number;
}) { ... }
```

### Sub Status Chart

```tsx
// _components/sub-status-chart.tsx
// Паттерн: chart-pie-donut-active
// activeShape через render prop — более живой визуально
<Pie activeShape={renderActiveShape} data={pieData} ... />
```

### KPI Cards структура

```tsx
// Badge + TrendingUp icon — паттерн из badge-demo
<Badge variant="secondary" className="bg-emerald-100 text-emerald-800">
  <TrendingUp className="h-3 w-3" />
  +8.3%
</Badge>
```

**Урок**: shadcn MCP — это оракул паттернов. Не изобретать велосипед — спрашивать MCP.

---

## 5. Замена mock-данных на реальный API

### Исходное состояние

```tsx
// ❌ До: page.tsx
const ACTIVE_USERS = 14205;
const MRR = 45230;
// ... всё захардкожено
```

### Шаг 1: Расширить AnalyticsRepository

В `backend/internal/domain/repository/analytics_repository.go` добавили 4 новых типа и 5 методов:

```go
type MonthlyMRR struct {
    Month string
    MRR   float64
}

type SubscriptionStatusCounts struct {
    Active      int
    GracePeriod int
    Cancelled   int
    Expired     int
}

type WebhookProviderHealth struct {
    Provider    string
    TotalEvents int
    Unprocessed int
    LastEventAt *time.Time
}

type AuditLogEntry struct {
    Time   time.Time
    Action string
    Detail string
}
```

**Урок**: начинать с типов домена. Не с SQL, не с JSON, а с Go-структурами — они диктуют контракт.

### Шаг 2: SQL реализация

Самый сложный запрос — MRR Trend через `generate_series`:

```sql
-- GetMRRTrend: последние N месяцев
WITH months AS (
    SELECT generate_series(
        date_trunc('month', NOW() - INTERVAL '5 months'),
        date_trunc('month', NOW()),
        '1 month'::interval
    ) AS month
)
SELECT
    to_char(m.month, 'YYYY-MM') AS month,
    COALESCE(SUM(t.amount), 0) AS mrr
FROM months m
LEFT JOIN transactions t ON date_trunc('month', t.created_at) = m.month
    AND t.status = 'success'
LEFT JOIN subscriptions s ON t.subscription_id = s.id
    AND s.plan_type = 'monthly'
GROUP BY m.month
ORDER BY m.month
```

**Урок**: `generate_series` — незаменим для time-series данных с заполнением пустых месяцев через `LEFT JOIN`. Без него месяцы без транзакций просто пропадут из графика.

### Шаг 3: Handler возвращает полный payload

```go
// handlers/admin.go GetDashboardMetrics
type dashboardResponse struct {
    ActiveUsers   int64                     `json:"active_users"`
    ActiveSubs    int64                     `json:"active_subs"`
    MRR           float64                   `json:"mrr"`
    ARR           float64                   `json:"arr"`
    ChurnRisk     int64                     `json:"churn_risk"`
    MRRTrend      []repo.MonthlyMRR         `json:"mrr_trend"`
    StatusCounts  repo.SubscriptionStatusCounts `json:"status_counts"`
    AuditLog      []repo.AuditLogEntry      `json:"audit_log"`
    WebhookHealth []repo.WebhookProviderHealth `json:"webhook_health"`
    LastUpdated   time.Time                 `json:"last_updated"`
}
```

### Шаг 4: Next.js Server Action

```typescript
// frontend/src/actions/dashboard.ts
export async function getDashboardMetrics(): Promise<DashboardMetrics | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return null; // не авторизован → FALLBACK в page.tsx

  const res = await fetch(`${BACKEND_URL}/v1/admin/dashboard/metrics`, {
    headers: { Authorization: `Bearer ${token}` },
    next: { revalidate: 60 }, // ISR: обновляем в фоне каждые 60 секунд
  });
  if (!res.ok) return null;
  return res.json();
}
```

**Урок**: `next: { revalidate: 60 }` на fetch — это ISR (Incremental Static Regeneration). Страница отдаётся из кэша мгновенно, а в фоне обновляется. Идеально для dashboard.

### Шаг 5: page.tsx как async Server Component

```tsx
// ✅ После: page.tsx
export default async function DashboardPage() {
  const [t, metrics] = await Promise.all([
    getTranslations("dashboard"),
    getDashboardMetrics(), // реальные данные из API
  ]);

  const d = metrics ?? FALLBACK; // нет данных → нули вместо краша

  return (
    <MrrTrendChart data={d.mrr_trend} activeSubs={d.active_subs} />
  );
}
```

**Урок**: `Promise.all` для параллельных запросов — не ждать переводы пока не загрузятся данные.

---

## 6. Database Seeding: реалистичные тестовые данные

### Зачем

После подключения реального API — всё нули. БД пустая. Нужны тестовые данные, максимально похожие на prod.

### Структура `scripts/seed_dev_data.sql`

```sql
-- 1. Тестовые пользователи (50 штук)
INSERT INTO users (platform_user_id, platform, app_version, email, ...)
SELECT
    'test_user_' || i,
    CASE WHEN i % 3 = 0 THEN 'ios' WHEN i % 3 = 1 THEN 'android' ELSE 'web' END,
    ...
FROM generate_series(1, 50) i;

-- 2. Подписки с разными статусами
-- 30 active, 5 grace, 8 cancelled, 7 expired

-- 3. Транзакции за 6 месяцев (для MRR Trend)
INSERT INTO transactions (user_id, subscription_id, amount, status, created_at)
SELECT ..., NOW() - (random() * 180 || ' days')::interval
FROM generate_series(1, 300) i;

-- 4. Webhook events (2 Google unprocessed → оранжевый badge)
-- 5. Admin audit log (5 записей)
```

**Урок**: семантика `ON CONFLICT DO NOTHING` — seed можно запускать несколько раз безопасно.

### Критический баг с LIMIT/OFFSET

```sql
-- ❌ Неправильно: LIMIT 35 OFFSET 30 = возьми 35 строк начиная с 31-й
-- ✅ Правильно:  LIMIT 5  OFFSET 30 = возьми 5 строк начиная с 31-й
```

**Урок**: `LIMIT N OFFSET M` — это не "бери строки с M по N", это "пропусти M строк, возьми следующие N".

### Запуск seed: хост vs контейнер

```bash
# ❌ Не работает: файл ищется ВНУТРИ контейнера
docker exec docker-compose-db-1 psql -U postgres -d iap_db -f scripts/seed_dev_data.sql

# ✅ Работает: stdin redirect с хоста
docker exec -i docker-compose-db-1 psql -U postgres -d iap_db < scripts/seed_dev_data.sql
```

**Урок**: `docker exec -f` ищет файл внутри контейнера. Для файлов с хоста — всегда stdin через `<`.

### Уникальный constraint подписок

```sql
-- Только ОДНА активная подписка на пользователя:
CREATE UNIQUE INDEX idx_subscriptions_one_active
ON subscriptions (user_id)
WHERE status = 'active' AND deleted_at IS NULL;
```

Это сломало seed — попытка создать 30 активных подписок для 30 пользователей, но каждому пользователю нужна уникальная активная. Решение: убедиться, что каждый `user_id` используется только один раз для активных подписок.

---

## 7. Docker: порты, сети и compose

### Проблема: port is already allocated

```
Error: Bind for 0.0.0.0:8081 failed: port is already allocated
```

`run_dev.sh` пытался угадать свободный порт через `lsof`, но не учитывал порты, занятые самим Docker (например, предыдущий запуск оставил `8082` за Docker daemon).

### Диагностика

```bash
# Что занимает порты
lsof -i -sTCP:LISTEN -nP | grep ':808'
# → httpd: 8080, com.docker: 8082

# Что запущено в Docker
docker ps --format 'table {{.Names}}\t{{.Ports}}'
```

### Правило двух маппингов

compose-файл должен иметь:
- `api` → `8081:8080` (хост:контейнер)
- `google-billing-mock` → `8090:8080` (хост:контейнер)

Нельзя, чтобы два контейнера слушали один и тот же порт хоста.

### Сети в Docker

```
docker-compose_internal  ← backend network (api, db, redis, mock)
frontend_default         ← frontend network

paywall-frontend         ← подключён к ОБОИМ
```

Фронтенд обращается к API через `http://api:8080` (внутренний Docker DNS). Это работает потому что frontend подключён к `docker-compose_internal`.

**Урок**: `BACKEND_URL=http://api:8080` — правильный env для production-like Docker setup. `localhost:8080` — только для разработки без Docker.

### Sed + macOS = боль

```bash
# ❌ GNU sed синтаксис (Linux):
sed -i.bak "0,/pattern/{s|old|new|}"  # не работает на macOS

# ✅ Python — универсальный способ:
python3 - "$FILE" "$OLD" "$NEW" <<'PYEOF'
import sys, re
path, old, new = sys.argv[1], sys.argv[2], sys.argv[3]
content = open(path).read()
content = re.sub(pattern, replacement, content, count=1, flags=re.DOTALL)
open(path, 'w').write(content)
PYEOF
```

**Урок**: для сложных in-place замен в скриптах на macOS — Python надёжнее `sed`.

---

## 8. Миграции: dirty flag и порядок запуска

### Проблема: dirty migration

```sql
SELECT version, dirty FROM schema_migrations;
-- version | dirty
-- --------|-------
-- 15      | t      ← застряли здесь!
```

Migrator упал посередине migration 15 из-за SQL-ошибки (функция в partial index не была `IMMUTABLE`). После этого все миграции блокируются.

### Fix dirty flag

```sql
UPDATE schema_migrations SET dirty = false WHERE version = 15;
```

Затем вручную запустить пропущенные миграции 16–19.

### Причина падения: non-IMMUTABLE function в index

```sql
-- ❌ Упало:
CREATE INDEX idx ON table WHERE next_retry_at <= now();
-- now() — не IMMUTABLE, нельзя использовать в partial index predicate

-- ✅ Fix в run_dev.sh (автозамена):
sed -i "s/WHERE status = 'pending' AND next_retry_at <= now()/WHERE status = 'pending'/g"
```

**Урок**: в PostgreSQL функции в `WHERE` clause индекса должны быть `IMMUTABLE`. `now()`, `current_timestamp` — `STABLE`, не подходят.

### admin_credentials — отдельная таблица

```sql
-- users: хранит профиль, платформу, роль — без пароля!
-- admin_credentials: user_id + password_hash (bcrypt) — только для admin/superadmin

CREATE TABLE admin_credentials (
    user_id     UUID PRIMARY KEY REFERENCES users(id),
    password_hash TEXT NOT NULL,
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);
```

**Урок**: разделение user profile и credentials — хорошая практика. Обычные mobile users никогда не имеют записи в `admin_credentials`.

### Bcrypt hash: всегда генерировать динамически

```go
// ❌ Хардкодить hash в SQL seed — он может быть для другого пароля
// ✅ Генерировать на лету:
hash, _ := bcrypt.GenerateFromPassword([]byte("admin12345"), bcrypt.DefaultCost)
```

Нашли правильный путь: Go-программа для генерации хеша → вставка в SQL:

```bash
cd backend && cat > /tmp/genhash.go << 'EOF'
package main
import ("fmt"; "golang.org/x/crypto/bcrypt")
func main() {
    h, _ := bcrypt.GenerateFromPassword([]byte("admin12345"), bcrypt.DefaultCost)
    fmt.Println(string(h))
}
EOF
go run /tmp/genhash.go
# → $2a$10$RVW7PiO3zH...
```

---

## 9. Компонентный апгрейд: Table vs список

### Было: plain list с Separator

```tsx
// ❌ До: простой текстовый список
d.audit_log.map((entry, i) => (
  <div>
    <span>[{time}]</span>
    <span>{entry.Action} · {entry.Detail}</span>
    {i < length - 1 && <Separator />}
  </div>
))
```

Проблема: всё в одну строку, трудно читать, нет визуальной иерархии.

### Стало: shadcn Table + color-coded Badge

```tsx
// ✅ После: AuditLogTable
// Три чёткие колонки: Time | Action (Badge) | Details
// Иконки по типу действия из lucide-react
// Цвета Badge:
// Grant   → emerald (зелёный)
// Revoke  → red (красный)
// Refund  → blue (синий)
// Dunning → orange (оранжевый)
// Pricing → violet (фиолетовый)
```

### Функция getActionMeta

```tsx
function getActionMeta(action: string): ActionMeta {
  const a = action.toLowerCase();
  if (a.includes("grant"))   return { icon: <ShieldCheck />, badge: "bg-emerald-100..." };
  if (a.includes("revoke"))  return { icon: <XCircle />,     badge: "bg-red-100..." };
  if (a.includes("refund"))  return { icon: <DollarSign />,  badge: "bg-blue-100..." };
  if (a.includes("retry"))   return { icon: <RefreshCw />,   badge: "bg-orange-100..." };
  if (a.includes("pricing")) return { icon: <Settings />,    badge: "bg-violet-100..." };
  return { icon: <Activity />, badge: "bg-muted..." }; // fallback
}
```

**Урок**: классификация через `string.includes()` — быстрый и достаточно надёжный подход для audit log actions. Не нужен enum, если сервер присылает строки.

### shadcn MCP для выбора компонента

```
shadcn-search_items_in_registries("timeline activity feed") → nothing
shadcn-get_item_examples_from_registries("table demo")      → Table + TableHeader + TableBody
shadcn-get_item_examples_from_registries("badge demo")      → Badge с иконками
```

**Урок**: если точный поиск не дал результата — искать более общий компонент (table вместо timeline). Admin UI строится из базовых блоков.

---

## 10. Ключевые технические инсайты

### 🔑 Инсайт 1: Next.js Server Action + Cookie

```typescript
// Цепочка авторизации на сервере:
// 1. Пользователь логинится → POST /v1/admin/auth/login
// 2. Сервер ставит HttpOnly cookie: "admin_access_token"
// 3. Server Action читает cookie серверно:
const cookieStore = await cookies(); // next/headers
const token = cookieStore.get("admin_access_token")?.value;
// 4. Передаёт в запрос к backend
```

**Преимущество**: токен никогда не попадает в JavaScript клиента — HttpOnly cookie. XSS не сможет украсть токен.

### 🔑 Инсайт 2: FALLBACK вместо краша

```tsx
const FALLBACK = {
  active_users: 0,
  mrr: 0,
  audit_log: [] as AuditLogEntry[],
  // ...
};

const d = metrics ?? FALLBACK;
// Если API недоступен — показываем нули, не Error 500
```

**Урок**: dashboard должен грацировать при недоступности API. Нули лучше белого экрана.

### 🔑 Инсайт 3: `as const` ломает типизацию

```typescript
// ❌ Неправильно:
const FALLBACK = { audit_log: [] } as const;
// Тип: readonly [] — нельзя передать как AuditLogEntry[]

// ✅ Правильно:
const FALLBACK = { audit_log: [] as AuditLogEntry[] };
```

### 🔑 Инсайт 4: Docker internal DNS

```yaml
# backend service видит mock как "google-billing-mock:8080"
# НЕ как "localhost:8090" (это хостовый порт)
environment:
  - GOOGLE_IAP_BASE_URL=http://google-billing-mock:8080
```

Docker Compose автоматически создаёт DNS-записи по именам сервисов в пределах одной сети.

### 🔑 Инсайт 5: generate_series для time series

```sql
-- Всегда использовать generate_series для временны́х рядов
-- иначе месяцы без данных пропадают из графика
WITH months AS (
    SELECT generate_series(...) AS month
)
SELECT m.month, COALESCE(SUM(amount), 0)
FROM months m
LEFT JOIN transactions t ON date_trunc('month', t.created_at) = m.month
GROUP BY m.month
```

### 🔑 Инсайт 6: пропасть между wireframe и production

Wireframe показывал `ACTIVE USERS: 14,205` — реальная БД показала `0`. Путь к реальным данным занял:

1. Расширить repo interface (новые методы)
2. Реализовать SQL (с edge cases)
3. Обновить service layer (делегирование)
4. Переписать handler (полный payload)
5. Создать server action (Next.js)
6. Обновить page (async, props)
7. Обновить компоненты (убрать моки)
8. Seed DB (реалистичные данные)
9. Починить миграции (dirty flag)
10. Создать admin user (credentials)

**Это нормально.** Wireframe → Production — это всегда 10 шагов, а не 1.

### 🔑 Инсайт 7: admin users в mobile-first схеме

Схема `users` создавалась для mobile (iOS/Android). Добавление admin-пользователей потребовало:

```sql
-- CHECK constraint обновлён:
platform IN ('ios', 'android', 'web')  -- добавили 'web' для admin

-- Отдельная таблица credentials:
admin_credentials (user_id, password_hash)
-- Обычные users не имеют записи — никакого пароля в users таблице
```

---

## 11. Архитектура финального результата

### Стек

```
Frontend (Next.js 14, App Router)
│
├── /dashboard/default (async Server Component)
│   ├── getDashboardMetrics() ← Server Action
│   │   ├── reads admin_access_token cookie
│   │   └── fetch GET /v1/admin/dashboard/metrics
│   │       └── next: { revalidate: 60 }  ← ISR
│   │
│   ├── <MrrTrendChart data={mrr_trend} />  ← AreaChart
│   ├── <SubStatusChart counts={status} />  ← PieChart donut
│   └── <AuditLogTable entries={audit_log} /> ← Table + Badge
│
Backend (Go, Gin, pgxpool)
│
├── GET /v1/admin/dashboard/metrics
│   ├── AdminMiddleware (JWT verify + role check)
│   └── GetDashboardMetrics handler
│       ├── CountUsers (SQLC)
│       ├── GetActiveSubscriptionCount (SQLC)
│       ├── GetMRRTrend (custom SQL + generate_series)
│       ├── GetSubscriptionStatusCounts (custom SQL)
│       ├── GetWebhookHealthByProvider (GROUP BY provider)
│       └── GetRecentAuditLog (JOIN admin_audit_log)
│
├── POST /v1/admin/auth/login
│   └── AdminLoginCommand (bcrypt verify → JWT issue)
│
Database (PostgreSQL 17)
│
├── users + admin_credentials
├── subscriptions + transactions
├── webhook_events
└── admin_audit_log

Docker (dev stack)
├── docker-compose-api-1          :8081→:8080
├── docker-compose-db-1           :5432
├── docker-compose-redis-1        :6379
├── docker-compose-google-billing-mock-1  :8090→:8080
├── docker-compose-worker-1
└── paywall-frontend              :3000
```

### Файлы изменённые/созданные в этой сессии

| Файл | Что сделано |
|------|-------------|
| `tests/google-billing-mock/` | git submodule: Google IAP mock сервер |
| `infra/docker-compose/docker-compose.local.yml` | добавлен google-billing-mock сервис |
| `backend/internal/domain/repository/analytics_repository.go` | 4 новых типа, 5 новых методов |
| `backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go` | SQL реализации (generate_series, GROUP BY) |
| `backend/internal/domain/service/analytics_service.go` | делегирующие методы |
| `backend/internal/interfaces/http/handlers/admin.go` | полный dashboard payload |
| `frontend/src/actions/dashboard.ts` | Server Action + типы + ISR cache |
| `frontend/src/app/(main)/dashboard/default/page.tsx` | async Server Component |
| `frontend/src/app/(main)/dashboard/default/_components/mrr-trend-chart.tsx` | AreaChart (shadcn паттерн) |
| `frontend/src/app/(main)/dashboard/default/_components/sub-status-chart.tsx` | PieChart donut (shadcn паттерн) |
| `frontend/src/app/(main)/dashboard/default/_components/audit-log-table.tsx` | Table + color-coded Badge |
| `scripts/seed_dev_data.sql` | 50 users, 50 subs, 300 txns, webhooks, audit log |
| `run_dev.sh` | cold-start скрипт: Docker check, migrations, seed, health |

### Итоговая статистика

| Метрика | Значение |
|---------|---------|
| Компонентов создано | 3 (MrrTrendChart, SubStatusChart, AuditLogTable) |
| Backend методов добавлено | 10 (repo + service + handler) |
| SQL запросов написано | 5 (MRR trend, status counts, webhook health, audit log, churn risk) |
| Строк в seed SQL | ~230 |
| DB записей засеяно | 51 users, 50 subs, 225 txns, 42 webhooks, 5 audit entries |
| Docker сервисов в стеке | 6 |

---

## 🎯 Главный вывод

> **Wireframe — это гипотеза. Production — это доказательство.**

Путь от `14,205` в мокапе до реального числа в БД — это не "имплементация", это **верификация всей архитектуры**: схемы данных, API-контракта, аутентификации, кэширования, Docker-сети и UI-компонентов.

Каждый "0" на дашборде — это не баг, это задача, которую нужно пройти насквозь:
```
0 → нет seed данных → нет миграции → dirty flag → нет admin_credentials → нет токена → нет cookie → нет данных с API
```

**Инструментарий этой сессии:**
- `shadcn MCP` — поиск паттернов без гугла
- `generate_series` — time-series без пустых месяцев
- `ON CONFLICT DO NOTHING` — идемпотентный seed
- `next: { revalidate: 60 }` — ISR для dashboard
- `cookies()` from `next/headers` — серверная авторизация
- `python3` вместо `sed` — кроссплатформенные скрипты
