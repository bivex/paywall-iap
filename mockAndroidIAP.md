# mockAndroidIAP.md — Тестирование Android IAP через локальный мок

> Полное руководство: как протестировать покупку Android подписки локально без реального Google Play, используя `google-billing-mock`.

---

## Содержание

1. [Архитектура: как работает в проде vs локально](#1-архитектура)
2. [Как мок подключён к backend](#2-подключение-мока)
3. [Сценарии мока (token prefix)](#3-сценарии)
4. [Пошаговое тестирование: рецепт через /verify/iap](#4-тестирование-рецепта)
5. [Тестирование webhook (RTDN Pub/Sub)](#5-тестирование-webhook)
6. [Admin API мока: управление сценариями в рантайме](#6-admin-api-мока)
7. [Что происходит в продакшене](#7-продакшн)

---

## 1. Архитектура

### Продакшн vs Локально

```
ПРОДАКШН
────────
Android App
  │ purchaseToken от Google Play
  ↓
POST /v1/verify/iap  ──→  GoogleVerifier  ──→  androidpublisher.googleapis.com
                                                (реальный Google Play API)

ЛОКАЛЬНО
────────
Android App / curl
  │ purchaseToken с нужным prefix
  ↓
POST /v1/verify/iap  ──→  GoogleVerifier  ──→  http://google-billing-mock:8080
                          (baseURL задан)       (наш мок, порт 8090 снаружи)
```

### Два пути в продакшене

| Путь | Когда | Кто инициирует |
|------|-------|----------------|
| `POST /v1/verify/iap` | Первая покупка, Restore Purchases | Мобильное приложение |
| `POST /webhook/google` | Auto-renewal, Cancel, Refund, Hold | Google Pub/Sub RTDN |

---

## 2. Подключение мока

Мок подключён через env-переменную в `docker-compose.local.yml`:

```yaml
# infra/docker-compose/docker-compose.local.yml
api:
  environment:
    - GOOGLE_IAP_BASE_URL=http://google-billing-mock:8080
```

`GoogleVerifier` при наличии `baseURL`:
- **Пропускает OAuth** (не нужен service account JSON)
- Перенаправляет все запросы к `androidpublisher/v3/...` на мок

Мок запущен на `http://localhost:8090` (снаружи Docker).  
Backend обращается к нему через `http://google-billing-mock:8080` (внутри Docker network).

---

## 3. Сценарии мока (token prefix)

Мок определяет ответ по **prefix purchaseToken**. Конфиг: `tests/google-billing-mock/config/scenarios/default.json`

### Подписки (type: subscription)

| Prefix токена | Сценарий | Valid? | Auto-renew | Срок |
|---|---|:---:|:---:|---|
| `valid_active_...` | Активная подписка | ✅ | ✅ | +30 дней |
| `expired_...` | Истекла | ❌ | ❌ | -1 день |
| `canceled_...` | Отменена пользователем | ❌ | ❌ | -1 час |
| `pending_...` | Ожидание оплаты | ❌ | ❌ | +1 день |
| `invalid_...` | HTTP 410 — токен не существует | 💥 | — | — |

### Разовые покупки (type: product)

| Prefix токена | Сценарий | Valid? |
|---|---|:---:|
| `product_valid_...` | Успешная покупка | ✅ |
| `product_pending_...` | Ожидание оплаты | ❌ |

> **Любой другой prefix** → мок вернёт случайный сценарий или 404.

---

## 4. Тестирование рецепта через /verify/iap

### Шаг 1: Получить JWT пользователя

Мобильное приложение сначала регистрируется/авторизуется:

```bash
# Регистрация нового пользователя (как делает Android-приложение)
USER_TOKEN=$(curl -s -X POST http://localhost:8081/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "platform_user_id": "android_test_user_001",
    "device_id": "device_abc123",
    "platform": "android",
    "app_version": "1.0.0",
    "email": "testuser@example.com"
  }' | jq -r '.access_token')

echo "User JWT: $USER_TOKEN"
```

### Шаг 2: Отправить рецепт Android

`receiptData` — это JSON-строка с параметрами покупки, которую Google Play SDK возвращает приложению.

```bash
# ✅ Сценарий: успешная подписка (valid_active)
curl -X POST http://localhost:8081/v1/verify/iap \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "android",
    "productId": "com.yourapp.premium_monthly",
    "receiptData": "{\"packageName\":\"com.yourapp\",\"productId\":\"com.yourapp.premium_monthly\",\"purchaseToken\":\"valid_active_tok_abc123\",\"type\":\"subscription\"}"
  }'
```

Ожидаемый ответ:
```json
{
  "subscription_id": "uuid-...",
  "status": "active",
  "expires_at": "2026-04-02T05:00:00Z",
  "auto_renew": true,
  "plan_type": "monthly",
  "is_new": true
}
```

### Шаг 3: Проверить другие сценарии

```bash
# ❌ Истекшая подписка
curl -X POST http://localhost:8081/v1/verify/iap \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "android",
    "productId": "com.yourapp.premium_monthly",
    "receiptData": "{\"packageName\":\"com.yourapp\",\"productId\":\"com.yourapp.premium_monthly\",\"purchaseToken\":\"expired_tok_xyz\",\"type\":\"subscription\"}"
  }'
# Ожидается: 422 Unprocessable Entity — "receipt is invalid"

# 💥 Невалидный токен (HTTP 410 от мока)
curl -X POST http://localhost:8081/v1/verify/iap \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "android",
    "productId": "com.yourapp.premium_monthly",
    "receiptData": "{\"packageName\":\"com.yourapp\",\"productId\":\"com.yourapp.premium_monthly\",\"purchaseToken\":\"invalid_tok_000\",\"type\":\"subscription\"}"
  }'
# Ожидается: 422 — "failed to verify Google Play subscription"

# ✅ Разовая покупка (product)
curl -X POST http://localhost:8081/v1/verify/iap \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "android",
    "productId": "com.yourapp.coins_100",
    "receiptData": "{\"packageName\":\"com.yourapp\",\"productId\":\"com.yourapp.coins_100\",\"purchaseToken\":\"product_valid_tok_001\",\"type\":\"product\"}"
  }'
```

### Шаг 4: Убедиться в базе данных

```bash
docker exec docker-compose-db-1 psql -U postgres -d iap_db -c "
  SELECT u.email, s.status, s.plan_type, s.expires_at, t.provider_tx_id
  FROM subscriptions s
  JOIN users u ON u.id = s.user_id
  JOIN transactions t ON t.subscription_id = s.id
  ORDER BY s.created_at DESC
  LIMIT 5;
"
```

---

## 5. Тестирование webhook (RTDN Pub/Sub)

Google в проде отправляет уведомления через Google Cloud Pub/Sub.  
Локально симулируем через прямой POST.

### Notification types

| `notificationType` | Что означает |
|---|---|
| `1` | SUBSCRIPTION_RECOVERED — восстановлена после hold |
| `2` | SUBSCRIPTION_RENEWED — авто-продление |
| `3` | SUBSCRIPTION_CANCELED — отменена |
| `4` | SUBSCRIPTION_PURCHASED — первая покупка (редко в webhook) |
| `5` | SUBSCRIPTION_ON_HOLD — нет оплаты, grace period |
| `6` | SUBSCRIPTION_IN_GRACE_PERIOD — льготный период |
| `12` | SUBSCRIPTION_REVOKED — немедленный отзыв (возврат) |
| `13` | SUBSCRIPTION_EXPIRED — истекла |

### Пример: симуляция авто-продления

```bash
# 1. Создаём base64 payload (DeveloperNotification)
PAYLOAD=$(echo -n '{
  "version": "1.0",
  "packageName": "com.yourapp",
  "eventTimeMillis": "1700000000000",
  "subscriptionNotification": {
    "version": "1.0",
    "notificationType": 2,
    "purchaseToken": "valid_active_tok_abc123",
    "subscriptionId": "com.yourapp.premium_monthly"
  }
}' | base64)

# 2. Отправить как Pub/Sub push
curl -X POST http://localhost:8081/webhook/google \
  -H "Content-Type: application/json" \
  -d "{
    \"message\": {
      \"data\": \"$PAYLOAD\",
      \"messageId\": \"test-msg-$(date +%s)\"
    },
    \"subscription\": \"projects/test-project/subscriptions/iap-sub\"
  }"
```

### Пример: симуляция отмены подписки

```bash
PAYLOAD=$(echo -n '{
  "version": "1.0",
  "packageName": "com.yourapp",
  "eventTimeMillis": "1700000000000",
  "subscriptionNotification": {
    "version": "1.0",
    "notificationType": 3,
    "purchaseToken": "valid_active_tok_abc123",
    "subscriptionId": "com.yourapp.premium_monthly"
  }
}' | base64)

curl -X POST http://localhost:8081/webhook/google \
  -H "Content-Type: application/json" \
  -d "{\"message\":{\"data\":\"$PAYLOAD\",\"messageId\":\"cancel-test-001\"},\"subscription\":\"projects/test/subscriptions/iap\"}"
```

---

## 6. Admin API мока: управление сценариями в рантайме

Мок предоставляет Admin API на `http://localhost:8090`.

```bash
# Проверить здоровье мока
curl http://localhost:8090/health

# Список всех сценариев
curl http://localhost:8090/admin/scenarios | jq

# Добавить кастомный сценарий (например: grace period)
curl -X POST http://localhost:8090/admin/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "name": "grace_period_test",
    "token_prefix": "grace_",
    "type": "subscription",
    "purchase_state": 0,
    "payment_state": 0,
    "acknowledgement_state": 1,
    "auto_renewing": true,
    "expiry_offset_seconds": 259200
  }'

# Теперь токен grace_tok_001 вернёт grace period сценарий
```

---

## 7. Что происходит в продакшене

Для перехода на реальный Google Play нужно:

### 1. Создать Service Account в Google Cloud Console
```
Google Cloud Console → IAM → Service Accounts → Create
→ Роль: "Service Account Token Creator"
→ Скачать JSON ключ
```

### 2. Выдать права в Google Play Console
```
Google Play Console → Setup → API access
→ Link to Google Cloud project
→ Выдать Service Account права "Financial data + Orders and subscriptions"
```

### 3. Настроить Google Cloud Pub/Sub для RTDN
```
Google Cloud Console → Pub/Sub → Create Topic: "android-iap"
→ Create Subscription → Push → URL: https://yourdomain.com/webhook/google
→ Google Play Console → Monetization → Real-time developer notifications
→ Topic: projects/YOUR_PROJECT/topics/android-iap
```

### 4. Обновить env-переменные

```bash
# .env (продакшн)
GOOGLE_IAP_BASE_URL=          # пустая строка = реальный Google Play API
GOOGLE_KEY_JSON={"type":"service_account","project_id":"..."}  # JSON ключ
IAP_IS_PRODUCTION=true
GOOGLE_WEBHOOK_SECRET=        # IP-whitelist Google: 64.233.160.0/19 и др.
```

> При `GOOGLE_IAP_BASE_URL=""` `GoogleVerifier` автоматически переключается на реальный `androidpublisher.googleapis.com` с OAuth через service account.

---

## Быстрый чеклист для локального теста

```bash
# 1. Убедиться что всё запущено
docker ps | grep -E "api|mock|db"

# 2. Мок отвечает?
curl http://localhost:8090/health

# 3. Backend видит мок?
curl -s http://localhost:8081/v1/verify/iap | jq
# Должно вернуть 401 (не 404)

# 4. Полный тест покупки
USER_TOKEN=$(curl -s -X POST http://localhost:8081/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"platform_user_id":"test001","device_id":"dev001","platform":"android","app_version":"1.0.0"}' \
  | jq -r '.access_token')

curl -X POST http://localhost:8081/v1/verify/iap \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"platform":"android","productId":"com.yourapp.premium_monthly","receiptData":"{\"packageName\":\"com.yourapp\",\"productId\":\"com.yourapp.premium_monthly\",\"purchaseToken\":\"valid_active_tok_001\",\"type\":\"subscription\"}"}'
```
