# Environment Variables

Variables marked REQUIRED must be set in production. Everything else has a working default.

## Server

| Variable     | Default | Description                         |
|--------------|---------|-------------------------------------|
| PORT         | 8080    | HTTP listen port (inside container) |
| ENV          | dev     | Environment name (dev/staging/prod) |
| LOG_LEVEL    | info    | Zap log level                       |
| CORS_ORIGINS | *       | Comma-separated allowed origins     |

## Database

| Variable          | Default                                      | Description          |
|-------------------|----------------------------------------------|----------------------|
| DATABASE_URL      | postgres://postgres:postgres@localhost:5432/iap_db | Full DSN       |
| DB_MAX_OPEN_CONNS | 25                                           | Max open connections |
| DB_MAX_IDLE_CONNS | 25                                           | Max idle connections |

## Redis

| Variable  | Default                  | Description             |
|-----------|--------------------------|-------------------------|
| REDIS_URL | redis://localhost:6379/0 | Redis connection string |

## Auth / Security

| Variable              | Default           | Description                                      |
|-----------------------|-------------------|--------------------------------------------------|
| JWT_SECRET            | REQUIRED (prod)   | HMAC-SHA256 signing key, min 32 chars            |
| JWT_ACCESS_EXPIRES    | 15m               | Access token TTL                                 |
| JWT_REFRESH_EXPIRES   | 720h              | Refresh token TTL (30 days)                      |
| APP_CREDENTIALS_KEY   | (dev mode)        | AES-256 key for store credentials, exactly 32 bytes. Absent = no encryption |

## IAP

| Variable            | Default                         | Description                            |
|---------------------|---------------------------------|----------------------------------------|
| IAP_IS_PRODUCTION   | false                           | Use real Apple/Google endpoints        |
| APPLE_MOCK_URL      | http://apple-iap-mock:9090      | Apple IAP server (dev mock)            |
| GOOGLE_IAP_BASE_URL | http://google-billing-mock:8080 | Google Play billing server (dev mock)  |

## External Billing — Lago

| Variable          | Default                   | Description             |
|-------------------|---------------------------|-------------------------|
| LAGO_API_URL      | https://api.getlago.com   | Lago API base URL       |
| LAGO_API_KEY      | REQUIRED (prod)           | Lago API key            |
| LAGO_WEBHOOK_SECRET | —                       | Lago webhook HMAC secret|

## Notifications — SendGrid

| Variable          | Default         | Description          |
|-------------------|-----------------|----------------------|
| SENDGRID_API_KEY  | REQUIRED (prod) | SendGrid API key     |
| EMAIL_FROM        | noreply@paywall.local | Sender email   |

## Notifications — FCM / APNs

| Variable       | Default         | Description                    |
|----------------|-----------------|--------------------------------|
| FCM_SERVER_KEY | REQUIRED (prod) | Firebase Cloud Messaging key   |
| APNS_KEY_ID    | REQUIRED (prod) | APNs key ID                    |
| APNS_TEAM_ID   | REQUIRED (prod) | Apple Developer Team ID        |
| APNS_KEY_FILE  | —               | Path to APNs .p8 private key   |

## Observability — Sentry

| Variable       | Default | Description               |
|----------------|---------|---------------------------|
| SENTRY_DSN     | —       | Sentry DSN (omit to skip) |
| SENTRY_ENV     | dev     | Environment tag           |
| SENTRY_RELEASE | —       | Release tag               |

## Production checklist

Minimum required vars for a production deployment:

```
DATABASE_URL
REDIS_URL
JWT_SECRET
APP_CREDENTIALS_KEY   (32 bytes)
IAP_IS_PRODUCTION=true
LAGO_API_KEY
LAGO_API_URL
SENDGRID_API_KEY
FCM_SERVER_KEY
APNS_KEY_ID
APNS_TEAM_ID
SENTRY_DSN
```

See `.env.example` at the repo root for a full template.
