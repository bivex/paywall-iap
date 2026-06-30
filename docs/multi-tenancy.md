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

### ⚠️ Вариант 2 — Добавить `app_id` в схему

Серьёзная миграция: затрагивает все 32 таблицы + все уникальные индексы
+ всю бизнес-логику в Go + роутинг API (JWT должен содержать `app_id`).

Примерный объём изменений:
```sql
-- В каждую ключевую таблицу:
ALTER TABLE users         ADD COLUMN app_id UUID NOT NULL;
ALTER TABLE subscriptions ADD COLUMN app_id UUID NOT NULL;
ALTER TABLE pricing_tiers ADD COLUMN app_id UUID NOT NULL;
ALTER TABLE ab_tests       ADD COLUMN app_id UUID NOT NULL;

-- Пересоздать уникальные индексы с учётом app_id:
DROP INDEX idx_users_platform_user_id_unique;
CREATE UNIQUE INDEX ON users(app_id, platform_user_id);

DROP INDEX ON pricing_tiers(name);
CREATE UNIQUE INDEX ON pricing_tiers(app_id, name);
```

**Плюсы:** один инстанс на все приложения.  
**Минусы:** очень большой рефакторинг, высокий риск регрессий.

---

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
