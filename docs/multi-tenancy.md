# Multi-Tenancy: Архитектурная заметка

> Дата анализа: 2026-07-01  
> Автор: анализ схемы БД (миграции 001–032)

---

## Вывод: paywall-iap — строго single-tenant

Ни в одной таблице нет колонки `app_id`, `tenant_id`, `organization_id`
или любого другого разделителя приложений. Каждый развёрнутый инстанс
обслуживает **одно приложение**.

### Затронутые таблицы

| Таблица | Проблема при shared-deploy |
|---|---|
| `users` | `platform_user_id UNIQUE` — один AppleID/GoogleID из двух разных игр создаст конфликт |
| `subscriptions` | FK только на `user_id`, нет изоляции по приложению |
| `pricing_tiers` | `name UNIQUE` — глобальный namespace, нельзя иметь `vip_monthly` в двух играх независимо |
| `ab_tests` | нет `app_id`, эксперименты перемешаются |
| `admin_credentials` | один инстанс — один набор администраторов |

---

## Варианты если нужно несколько приложений

### ✅ Вариант 1 — Один инстанс на приложение (рекомендуется)

Деплоишь отдельный Docker stack с отдельной БД на каждую игру.
Инфраструктура уже поддерживает это через `docker-compose`.

```
game-1/  →  paywall-iap stack (port 8081, db: iap_db_game1)
game-2/  →  paywall-iap stack (port 8082, db: iap_db_game2)
```

**Плюсы:** нет изменений в коде, полная изоляция данных, независимые A/B тесты.  
**Минусы:** больше инфраструктуры для обслуживания.

---

### ✅ Вариант 2 — Добавить `app_id` в схему **[РЕАЛИЗОВАН]**

**Миграция:** [`033_add_multi_tenancy.up.sql`](../migrations/033_add_multi_tenancy.up.sql)  
**Откат:** [`033_add_multi_tenancy.down.sql`](../migrations/033_add_multi_tenancy.down.sql)

Добавлена таблица `apps` и колонка `app_id UUID NOT NULL` во все таблицы с
пользовательскими данными. Существующие строки бэк-филлятся sentinel-приложением
`com.mothsalt.legacy` — миграция безопасна на живой БД.

#### Затронутые таблицы

| Таблица | Изменение |
|---|---|
| `users` | `app_id` + переделан UNIQUE на `(app_id, platform_user_id)` и `(app_id, email)` |
| `pricing_tiers` | `app_id` + UNIQUE на `(app_id, name)` |
| `ab_tests` | `app_id` |
| `webhook_events` | `app_id` + UNIQUE на `(app_id, provider, event_id)` |
| `analytics_aggregates` | `app_id` + UNIQUE на `(app_id, metric_name, metric_date, dimensions)` |
| `matomo_staged_events` | `app_id` |
| `bandit_user_context` | `app_id` + PK переделан на `(app_id, user_id)` |

#### Таблицы без app_id (глобальные / инфра)
`admin_credentials`, `admin_audit_log`, `admin_settings`,
`automation_job_run_log`, `currency_rates`,
`experiment_lifecycle_audit_log`, `experiment_automation_decision_log`

#### Сидированные приложения Mothsalt

```sql
-- Автоматически вставляются миграцией:
com.mothsalt.game1  →  iOS
com.mothsalt.game2  →  iOS
com.mothsalt.game3  →  iOS
com.mothsalt.game4  →  Android
com.mothsalt.game5  →  Android
```

#### Применить миграцию

```bash
# Через Docker (рекомендуется)
docker exec -i <db_container> psql -U postgres -d iap_db \
  < backend/migrations/033_add_multi_tenancy.up.sql

# Или через Go migrator
cd backend && go run ./cmd/migrate up
```

**Плюсы:** один инстанс на все приложения, полная изоляция данных.  
**Минусы:** Go-слой и роутинг JWT ещё требуют доработки (см. ниже).

#### Что ещё нужно доделать в Go после миграции

- [ ] `app_id` в JWT claims (`bundle_id` → lookup `apps.id`)
- [ ] Middleware: извлекать `app_id` из JWT и прокидывать в `context`
- [ ] Все репозитории: добавить `WHERE app_id = $1` к каждому запросу
- [ ] Admin UI: фильтр по приложению на всех страницах
- [ ] SDK (`PaywallBanditClient`): передавать `bundle_id` при регистрации



### 🔵 Вариант 3 — PostgreSQL schemas

Изоляция на уровне PG без изменения Go-кода.
Каждое приложение получает свою схему (`game1`, `game2`) в одной БД.

```sql
CREATE SCHEMA game1;
CREATE SCHEMA game2;
-- Запускать миграции в каждой схеме отдельно:
SET search_path = game1; -- затем применить все *.up.sql
SET search_path = game2; -- затем применить все *.up.sql
```

Требует: параметр `search_path` в строке подключения (`?search_path=game1`)
или отдельные `DATABASE_URL` на инстанс воркера/API для каждой схемы.

**Плюсы:** один PostgreSQL сервер, изоляция без изменения кода.  
**Минусы:** сложнее в обслуживании, мониторинге и бэкапах.

---

## Как это отражается на SDK

В `PaywallBanditClient` и `BackendReceiptValidator` есть единственный
параметр `apiBaseUrl`. Архитектура SDK уже **подразумевает «один бэкенд
на одно приложение»** — это полностью совпадает с Вариантом 1.

```dart
// Каждое приложение передаёт свой URL при инициализации:
final bandit = PaywallBanditClient(
  apiBaseUrl: 'https://paywall.my-game-1.com',  // инстанс для игры 1
  getAuthToken: () => authService.currentToken,
);
```

**Рекомендация:** придерживаться Варианта 1 до тех пор, пока количество
приложений не сделает обслуживание нескольких инстансов неудобным.
