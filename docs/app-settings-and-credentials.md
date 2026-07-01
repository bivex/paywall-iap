# App Settings & Store Credentials

> Migration: `036_app_settings_and_credentials`  
> Date: 2026-07-01

Each app in the system has two configuration layers:

- **`apps.settings` (JSONB)** — non-sensitive per-app paywall behaviour
- **`app_credentials`** — sensitive store API keys, AES-256-GCM encrypted at rest

---

## App Settings

Stored as JSONB in `apps.settings`. Edited via **Admin Panel → Apps → ⚙ Configure → Settings tab**.

| Field | Type | Default | Description |
|---|---|---|---|
| `grace_period_days` | int (0–90) | `3` | Days after subscription expiry before access is revoked |
| `trial_enabled` | bool | `false` | Allow free trial for new users |
| `trial_days` | int (0–365) | `0` | Trial duration when `trial_enabled = true` |
| `default_currency` | string (ISO-4217) | `USD` | Default currency for pricing tiers |
| `store_environment` | `production` \| `sandbox` | `production` | Which store environment to use for receipt validation |
| `webhook_url` | string | `""` | Endpoint to receive subscription lifecycle events |
| `webhook_secret` | string | `""` | HMAC secret for webhook signature verification |
| `entitlements` | `map[product_id][]feature_key` | `{}` | Maps store product IDs to feature keys granted on purchase |

### Entitlements example

```json
{
  "com.mothsalt.game1.monthly": ["premium", "ads_free"],
  "com.mothsalt.game1.yearly":  ["premium", "ads_free", "cloud_save"]
}
```

### API

```
GET  /v1/admin/apps/:id/settings
PUT  /v1/admin/apps/:id/settings
```

PUT accepts partial updates — only provided fields are changed. Requires `X-App-ID` header and admin JWT.

---

## Store Credentials

Stored in the `app_credentials` table, one row per `(app_id, provider)`.

Sensitive fields (marked `_enc` in the DB column name) are encrypted with **AES-256-GCM** before insert and decrypted on read. The encryption key is set via the `APP_CREDENTIALS_KEY` environment variable (must be exactly 32 bytes).

If `APP_CREDENTIALS_KEY` is not set, the system operates in **dev mode** — credentials are stored as plaintext. Never deploy to production without this env var set.

### Supported providers

#### Apple App Store

| Field | Sensitive | Description |
|---|---|---|
| `apple_team_id` | no | 10-char Team ID from Apple Developer portal |
| `apple_key_id` | no | Key ID for App Store Connect API |
| `apple_bundle_id` | no | App bundle identifier |
| `apple_environment` | no | `production` or `sandbox` |
| `apple_shared_secret` | **yes** | Shared secret for receipt validation (legacy) |
| `apple_private_key` | **yes** | `.p8` private key for App Store Connect JWT auth |

#### Google Play

| Field | Sensitive | Description |
|---|---|---|
| `google_package_name` | no | App package name |
| `google_service_account` | **yes** | Full service account JSON from Google Cloud Console |

#### Stripe

| Field | Sensitive | Description |
|---|---|---|
| `stripe_publishable_key` | no | `pk_live_...` publishable key |
| `stripe_secret_key` | **yes** | `sk_live_...` secret key |
| `stripe_webhook_secret` | **yes** | `whsec_...` webhook signing secret |

#### Paddle

| Field | Sensitive | Description |
|---|---|---|
| `paddle_vendor_id` | no | Numeric vendor ID |
| `paddle_api_key` | **yes** | Paddle API key |
| `paddle_webhook_secret` | **yes** | Webhook secret for signature verification |

### API

```
GET    /v1/admin/apps/:id/credentials
PUT    /v1/admin/apps/:id/credentials
DELETE /v1/admin/apps/:id/credentials/:provider
```

GET response **never returns sensitive values**. Instead, boolean `_set` flags indicate whether a secret is configured:

```json
{
  "credentials": [
    {
      "provider": "apple",
      "apple_team_id": "ABCD1234EF",
      "apple_key_id": "ABCD1234EF",
      "apple_bundle_id": "com.mothsalt.game1",
      "apple_environment": "production",
      "apple_shared_secret_set": true,
      "apple_private_key_set": true
    }
  ]
}
```

PUT upserts by provider. Pass only the fields you want to update. Leave sensitive fields empty to keep existing values:

```json
{
  "provider": "apple",
  "apple_team_id": "NEWTEAMID1",
  "apple_shared_secret": ""
}
```

---

## Admin UI

Navigate to **Admin Panel → Apps**, then click the ⚙ icon on any app row.

The config page has two tabs:

- **Settings** — grace period, trial, currency, store environment, webhook, entitlements JSON editor
- **Store Credentials** — collapsible sections per provider with `Configured / Not set` badges for sensitive fields

Apps are also searchable via **⌘J** with direct links to the configure page.

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `APP_CREDENTIALS_KEY` | **production** | AES-256 key, exactly 32 bytes. Set via Docker secret or env. |

Generate a key:

```bash
openssl rand -hex 16   # 32 hex chars = 16 bytes — too short
openssl rand -base64 24 | tr -d '=' | cut -c1-32  # 32 ASCII chars
# or simply:
python3 -c "import secrets; print(secrets.token_hex(16))"
```
