-- =============================================================
-- Dev seed data — realistic dashboard fixtures
-- Run: psql $DATABASE_URL -f scripts/seed_dev_data.sql
--   or: docker exec docker-compose-db-1 psql -U postgres -d iap_db \
--         -f /scripts/seed_dev_data.sql
-- Safe to re-run (ON CONFLICT DO NOTHING throughout)
-- =============================================================

BEGIN;

-- ─── 1. USERS (50 regular test users) ────────────────────────────────────────
INSERT INTO users (id, platform_user_id, platform, app_version, email, role, ltv, created_at)
SELECT
  gen_random_uuid(),
  'usr_seed_' || i,
  (ARRAY['ios','ios','ios','android','android','web'])[1 + (i % 6)],
  (ARRAY['2.1.0','2.2.0','2.3.0','3.0.0'])[1 + (i % 4)],
  'user' || i || '@seed.example.com',
  'user',
  ROUND((9.99 + (i % 10) * 10)::numeric, 2),
  now() - ((180 - i * 3) || ' days')::interval
FROM generate_series(1, 50) AS i
ON CONFLICT DO NOTHING;

-- ─── 2. SUBSCRIPTIONS ────────────────────────────────────────────────────────

-- 2a. Active (users 1-30)
INSERT INTO subscriptions (id, user_id, status, source, platform, product_id, plan_type, expires_at, auto_renew, created_at, updated_at)
SELECT
  gen_random_uuid(),
  u.id,
  'active',
  (ARRAY['iap','iap','stripe'])[1 + (u.rn % 3)],
  u.platform,
  CASE WHEN u.rn % 2 = 0 THEN 'pro_monthly' ELSE 'pro_annual' END,
  CASE WHEN u.rn % 2 = 0 THEN 'monthly'     ELSE 'annual'     END,
  now() + interval '30 days',
  true,
  now() - ((180 - u.rn * 5) || ' days')::interval,
  now()
FROM (
  SELECT id, platform, row_number() OVER (ORDER BY created_at) AS rn
  FROM users WHERE email LIKE '%@seed.example.com'
  ORDER BY created_at LIMIT 30
) u
ON CONFLICT DO NOTHING;

-- 2b. Grace period (users 31-35, 5 rows)
INSERT INTO subscriptions (id, user_id, status, source, platform, product_id, plan_type, expires_at, auto_renew, created_at, updated_at)
SELECT
  gen_random_uuid(),
  u.id,
  'grace',
  'iap',
  u.platform,
  'pro_monthly',
  'monthly',
  now() - interval '3 days',
  true,
  now() - interval '35 days',
  now() - interval '3 days'
FROM (
  SELECT id, platform FROM users WHERE email LIKE '%@seed.example.com'
  ORDER BY created_at LIMIT 5 OFFSET 30
) u
ON CONFLICT DO NOTHING;

-- 2c. Cancelled (users 36-43, 8 rows)
INSERT INTO subscriptions (id, user_id, status, source, platform, product_id, plan_type, expires_at, auto_renew, created_at, updated_at)
SELECT
  gen_random_uuid(),
  u.id,
  'cancelled',
  (ARRAY['iap','stripe'])[1 + (row_number() OVER (ORDER BY u.created_at) % 2)],
  u.platform,
  'pro_annual',
  'annual',
  now() - interval '10 days',
  false,
  now() - interval '90 days',
  now() - interval '10 days'
FROM (
  SELECT id, platform, created_at FROM users WHERE email LIKE '%@seed.example.com'
  ORDER BY created_at LIMIT 8 OFFSET 35
) u
ON CONFLICT DO NOTHING;

-- 2d. Expired (users 44-50, 7 rows)
INSERT INTO subscriptions (id, user_id, status, source, platform, product_id, plan_type, expires_at, auto_renew, created_at, updated_at)
SELECT
  gen_random_uuid(),
  u.id,
  'expired',
  'iap',
  u.platform,
  'pro_monthly',
  'monthly',
  now() - interval '20 days',
  false,
  now() - interval '60 days',
  now() - interval '20 days'
FROM (
  SELECT id, platform FROM users WHERE email LIKE '%@seed.example.com'
  ORDER BY created_at LIMIT 7 OFFSET 43
) u
ON CONFLICT DO NOTHING;

-- ─── 3. TRANSACTIONS (MRR trend over 6 months) ───────────────────────────────
-- Monthly renewals for active/grace subscriptions, one per month × 6 months
INSERT INTO transactions (id, user_id, subscription_id, amount, currency, status, provider_tx_id, created_at)
SELECT
  gen_random_uuid(),
  s.user_id,
  s.id,
  CASE WHEN s.plan_type = 'monthly' THEN 9.99 ELSE 99.99 END,
  'USD',
  'success',
  'txn_seed_' || s.id::text || '_m' || m.mo,
  date_trunc('month', now()) - ((5 - m.mo) || ' months')::interval
    + ((floor(random() * 27) + 1)::int || ' days')::interval
FROM subscriptions s
CROSS JOIN generate_series(0, 5) AS m(mo)
WHERE s.status IN ('active', 'grace')
  AND s.user_id IN (SELECT id FROM users WHERE email LIKE '%@seed.example.com')
ON CONFLICT DO NOTHING;

-- Single historical transaction for cancelled/expired subs
INSERT INTO transactions (id, user_id, subscription_id, amount, currency, status, provider_tx_id, created_at)
SELECT
  gen_random_uuid(),
  s.user_id,
  s.id,
  CASE WHEN s.plan_type = 'monthly' THEN 9.99 ELSE 99.99 END,
  'USD',
  'success',
  'txn_seed_hist_' || s.id::text,
  s.created_at + interval '1 hour'
FROM subscriptions s
WHERE s.status IN ('cancelled', 'expired')
  AND s.user_id IN (SELECT id FROM users WHERE email LIKE '%@seed.example.com')
ON CONFLICT DO NOTHING;

-- A few refunded transactions
INSERT INTO transactions (id, user_id, subscription_id, amount, currency, status, provider_tx_id, created_at)
SELECT
  gen_random_uuid(),
  s.user_id,
  s.id,
  9.99,
  'USD',
  'refunded',
  'txn_seed_refund_' || s.id::text,
  now() - interval '5 days'
FROM subscriptions s
WHERE s.status = 'cancelled'
  AND s.user_id IN (SELECT id FROM users WHERE email LIKE '%@seed.example.com')
LIMIT 3
ON CONFLICT DO NOTHING;

-- ─── 4. WEBHOOK EVENTS ───────────────────────────────────────────────────────
-- Stripe: 20 events, all processed
INSERT INTO webhook_events (provider, event_type, event_id, payload, processed_at, created_at)
SELECT
  'stripe',
  (ARRAY['payment_intent.succeeded','customer.subscription.updated','invoice.paid'])[1 + (i % 3)],
  'evt_stripe_seed_' || i,
  ('{"amount":' || (999 + i * 100) || '}')::jsonb,
  now() - (i || ' hours')::interval,
  now() - (i || ' hours')::interval
FROM generate_series(1, 20) AS i
ON CONFLICT DO NOTHING;

-- Apple: 12 events, all processed
INSERT INTO webhook_events (provider, event_type, event_id, payload, processed_at, created_at)
SELECT
  'apple',
  (ARRAY['DID_RENEW','DID_CHANGE_RENEWAL_STATUS','CANCEL'])[1 + (i % 3)],
  'evt_apple_seed_' || i,
  '{"notification_type":"DID_RENEW","auto_renew_product_id":"pro_monthly"}'::jsonb,
  now() - (i || ' hours')::interval,
  now() - (i || ' hours')::interval
FROM generate_series(1, 12) AS i
ON CONFLICT DO NOTHING;

-- Google: 10 events — 8 processed, 2 unprocessed (pending retry)
INSERT INTO webhook_events (provider, event_type, event_id, payload, processed_at, created_at)
SELECT
  'google',
  'subscriptionNotification',
  'evt_google_seed_' || i,
  '{"version":"1.0","packageName":"com.example.app"}'::jsonb,
  CASE WHEN i <= 8 THEN now() - (i || ' hours')::interval ELSE NULL END,
  now() - (i || ' hours')::interval
FROM generate_series(1, 10) AS i
ON CONFLICT DO NOTHING;

-- ─── 5. ADMIN AUDIT LOG ──────────────────────────────────────────────────────
INSERT INTO admin_audit_log (admin_id, action, target_type, details, created_at)
SELECT
  (SELECT id FROM users WHERE role IN ('admin','superadmin') ORDER BY created_at LIMIT 1),
  t.action,
  t.target_type,
  t.details::jsonb,
  t.created_at
FROM (VALUES
  ('grant_subscription',   'user',         '{"product_id":"pro_annual","plan_type":"annual","expires_at":"2027-01-01T00:00:00Z"}', now() - interval '20 minutes'),
  ('updated_pricing_tier', 'plan',         '{"tier":"Pro Annual","old_price":49.99,"new_price":39.99}',                             now() - interval '1 hour'),
  ('refund_transaction',   'transaction',  '{"transaction_id":"txn_8821","amount":49.99,"reason":"customer_request"}',             now() - interval '2 hours'),
  ('revoke_subscription',  'user',         '{"reason":"policy_violation"}',                                                        now() - interval '5 hours'),
  ('auto_retry_dunning',   'subscription', '{"attempt":2,"max_attempts":4}',                                                       now() - interval '8 hours')
) AS t(action, target_type, details, created_at)
ON CONFLICT DO NOTHING;

COMMIT;

-- ─── Verify ───────────────────────────────────────────────────────────────────
SELECT
  (SELECT COUNT(*) FROM users)                                          AS total_users,
  (SELECT COUNT(*) FROM subscriptions WHERE status = 'active')         AS active_subs,
  (SELECT COUNT(*) FROM subscriptions WHERE status = 'grace')          AS grace_subs,
  (SELECT COUNT(*) FROM subscriptions WHERE status = 'cancelled')      AS cancelled_subs,
  (SELECT COUNT(*) FROM subscriptions WHERE status = 'expired')        AS expired_subs,
  (SELECT COUNT(*) FROM transactions WHERE status = 'success')         AS txn_success,
  (SELECT COUNT(*) FROM webhook_events WHERE processed_at IS NULL)     AS unprocessed_webhooks,
  (SELECT COUNT(*) FROM admin_audit_log)                               AS audit_entries;
