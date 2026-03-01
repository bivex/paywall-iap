# IAP Sandbox Setup Guide

Руководство по настройке sandbox-тестирования покупок (In-App Purchases) для iOS и Android.

---

## Архитектура верификации

```
Mobile App
    │  POST /v1/verify/iap  { platform, receipt_data, product_id }
    ▼
IAPHandler  (interfaces/http/handlers/iap.go)
    │
    ▼
VerifyIAPCommand  (application/command/verify_iap.go)
    │  выбирает верификатор по platform
    ├─── ios     → AppleVerifierAdapter
    │                 └── AppleVerifier.VerifyReceipt()
    │                         └── go-iap → sandbox.itunes.apple.com  (или mock)
    │
    └─── android → AndroidVerifierAdapter
                      └── GoogleVerifier.VerifyReceipt()
                              └── androidpublisher API  (или mock)
    │
    ▼  результат верификации
    ├── проверка дубликата (hash receipt → transactions)
    ├── создать / обновить Subscription
    └── записать Transaction
```

Оба адаптера (`AppleVerifierAdapter`, `AndroidVerifierAdapter`) живут в  
`backend/internal/infrastructure/external/iap/adapter.go`  
и реализуют интерфейс `command.IAPVerifier`.

**Mock-режим** срабатывает автоматически, если переменная окружения пустая:
- `APPLE_SHARED_SECRET` пустой → `AppleVerifier` возвращает `Valid: true` без HTTP-запроса
- `GOOGLE_SERVICE_ACCOUNT_JSON` пустой → `GoogleVerifier` возвращает `Valid: true` без HTTP-запроса

---

## Режимы работы

| Режим | Условие | Поведение |
|-------|---------|-----------|
| **Mock** | Ключи не заданы (пустые строки) | Верификатор возвращает `Valid: true` без запросов во внешние сервисы |
| **Sandbox** | Ключи заданы, `IAP_IS_PRODUCTION` не установлен | Запросы идут в тестовые эндпоинты Apple/Google |
| **Production** | Ключи заданы, `IAP_IS_PRODUCTION=true` | Запросы идут в боевые эндпоинты |

---

## iOS (Apple Sandbox)

### Шаг 1 — Получить App-Specific Shared Secret

1. Открыть [App Store Connect](https://appstoreconnect.apple.com)
2. Перейти: **Apps → [твоё приложение] → In-App Purchases → Manage**
3. В правом верхнем углу нажать **App-Specific Shared Secret → Generate**
4. Скопировать сгенерированный секрет

### Шаг 2 — Прописать секрет в бэкенд

В файле `backend/.env`:

```env
APPLE_SHARED_SECRET=<вставить скопированный секрет>
```

### Шаг 3 — Создать Sandbox-тестера

1. Открыть [App Store Connect](https://appstoreconnect.apple.com)
2. Перейти: **Users and Access → Sandbox → Testers → (+)**
3. Заполнить данные (email должен быть незарегистрирован в Apple)
4. Сохранить

### Шаг 4 — Авторизоваться на устройстве

1. На iOS-устройстве открыть **Settings → App Store**
2. Прокрутить вниз до раздела **Sandbox Account**
3. Нажать **Sign In** и войти под созданным тестовым аккаунтом

> **Важно:** Не заходить в sandbox-аккаунт через обычный `Sign in with Apple` —  
> только через раздел `Sandbox Account` в настройках.

### Шаг 5 — Убедиться, что бэкенд использует Sandbox

В `backend/internal/infrastructure/external/iap/verifier.go` верификатор инициализируется через:

```go
NewAppleVerifier(sharedSecret string, isProduction bool)
```

При `isProduction = false` запросы идут на `https://sandbox.itunes.apple.com/verifyReceipt`.  
Убедиться, что в точке инициализации передаётся `false`.

---

## Android (Google Play Sandbox)

### Шаг 1 — Привязать Google Play к Google Cloud проекту

1. Открыть [Google Play Console](https://play.google.com/console)
2. Перейти: **Setup → API access**
3. Нажать **Link to a Google Cloud Project** (или создать новый)
4. Подтвердить связь

### Шаг 2 — Создать сервисный аккаунт

1. В том же разделе **API access** нажать **View in Google Cloud Console**
2. Перейти: **IAM & Admin → Service Accounts → Create Service Account**
3. Задать имя (например, `iap-verifier`)
4. Назначить роль: **Editor** (или минимально необходимую)
5. Нажать **Create Key → JSON** — скачать файл

### Шаг 3 — Выдать права сервисному аккаунту в Play Console

1. Вернуться в Play Console → **Setup → API access**
2. Найти созданный сервисный аккаунт, нажать **Grant access**
3. Выдать разрешение **View financial data, orders, and cancellation survey responses**
4. Нажать **Apply**

### Шаг 4 — Прописать JSON-ключ в бэкенд

Содержимое скачанного JSON-файла положить в `backend/.env` одной строкой:

```env
GOOGLE_SERVICE_ACCOUNT_JSON={"type":"service_account","project_id":"your-project","private_key_id":"...","private_key":"-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----\n","client_email":"iap-verifier@your-project.iam.gserviceaccount.com","client_id":"...","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token"}
```

> Если JSON многострочный — убрать переносы строк или передать через переменную окружения в CI.

### Шаг 5 — Добавить тестовых покупателей

1. Play Console → **Setup → License Testing**
2. В поле **License testers** добавить Gmail-адреса тестировщиков
3. Сохранить

Пользователи из этого списка будут получать sandbox-покупки без реального списания средств.

### Шаг 6 — Прописать публичный ключ в мобильное приложение

1. Play Console → **Setup → App integrity → App signing**
2. Скопировать **SHA-256 certificate fingerprint**
3. Конвертировать в Base64 (если требуется `react-native-iap`)
4. В файле `mobile/.env`:

```env
ANDROID_PUBLIC_KEY=<base64_encoded_public_key>
```

---

## Проверка работы

### Локально (mock-режим, без ключей)

```bash
# Запустить бэкенд
cd backend && make run

# Отправить тестовый запрос верификации
curl -X POST http://localhost:8080/v1/subscriptions/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt_token>" \
  -d '{
    "platform": "ios",
    "receipt_data": "mock_receipt_data_1234567890",
    "product_id": "com.yourapp.premium.monthly"
  }'
```

Ожидаемый ответ при пустом `APPLE_SHARED_SECRET`: `Valid: true` (mock).

### Sandbox-режим (с ключами)

Те же шаги, но в `backend/.env` заполнены реальные ключи. При покупке через тестовый аккаунт бэкенд отправит запрос в sandbox Apple/Google и вернёт реальный результат верификации.

---

## Переменные окружения (итог)

### `backend/.env`

```env
# Apple IAP (оставить пустым для mock-режима)
APPLE_SHARED_SECRET=

# Google IAP (оставить пустым для mock-режима)
GOOGLE_SERVICE_ACCOUNT_JSON=

# Webhook secrets
STRIPE_WEBHOOK_SECRET=whsec_dummy
APPLE_WEBHOOK_SECRET=whsec_dummy
GOOGLE_WEBHOOK_SECRET=whsec_dummy
```

### `mobile/.env`

```env
API_BASE_URL=http://localhost:8080/v1

# Android (оставить пустым если не тестируется Android)
ANDROID_PUBLIC_KEY=
```
