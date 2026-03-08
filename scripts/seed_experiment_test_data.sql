-- =============================================================
-- Experiment + bandit fixture data for cold-start local testing
-- Run via scripts/seed_all_test_data.sh
-- Safe to re-run (UPSERTs / guarded inserts throughout)
-- =============================================================

BEGIN;

INSERT INTO currency_rates (base_currency, target_currency, rate, source, updated_at)
VALUES
  ('USD', 'EUR', 0.92, 'seed_all_test_data', now()),
  ('USD', 'GBP', 0.79, 'seed_all_test_data', now()),
  ('USD', 'JPY', 148.50, 'seed_all_test_data', now())
ON CONFLICT (base_currency, target_currency) DO UPDATE
SET rate = EXCLUDED.rate,
    source = EXCLUDED.source,
    updated_at = now();

INSERT INTO ab_tests (
  id, name, description, status, start_at, end_at,
  algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence,
  automation_policy,
  created_at, updated_at,
  window_type, window_size, window_min_samples,
  objective_type, objective_weights,
  price_normalization, enable_contextual, enable_delayed, enable_currency, exploration_alpha
)
VALUES
  (
    '10000000-0000-0000-0000-000000000001',
    'Seed Hybrid Paywall Experiment',
    'Running hybrid bandit fixture for Studio, Sliding Window, Delayed Feedback, and Objective dashboards.',
    'running',
    now() - interval '3 days',
    now() + interval '14 days',
    'thompson_sampling', true, 50, 0.95, 0.87,
    '{"enabled":true,"auto_start":true,"auto_complete":true,"complete_on_end_time":true,"complete_on_sample_size":false,"complete_on_confidence":false,"manual_override":false,"locked_until":null,"locked_by":null,"lock_reason":null}'::jsonb,
    now() - interval '3 days', now(),
    'events', 500, 50,
    'hybrid', '{"conversion":0.5,"ltv":0.3,"revenue":0.2}'::jsonb,
    true, true, true, true, 0.30
  ),
  (
    '10000000-0000-0000-0000-000000000002',
    'Seed Onboarding Copy Test',
    'Draft classic A/B experiment fixture for CRUD and arm-plan testing.',
    'draft',
    now() + interval '1 day',
    now() + interval '21 days',
    NULL, false, 100, 0.95, NULL,
    '{"enabled":false,"auto_start":false,"auto_complete":false,"complete_on_end_time":true,"complete_on_sample_size":false,"complete_on_confidence":false,"manual_override":false,"locked_until":null,"locked_by":null,"lock_reason":null}'::jsonb,
    now() - interval '1 day', now(),
    'none', 1000, 100,
    'conversion', NULL,
    false, false, false, false, 0.30
  ),
  (
    '10000000-0000-0000-0000-000000000003',
    'Seed Winback Offer Test',
    'Paused revenue-focused bandit fixture for status coverage and alternative arm telemetry.',
    'paused',
    now() - interval '10 days',
    now() + interval '10 days',
    'ucb', true, 75, 0.90, 0.64,
    '{"enabled":true,"auto_start":true,"auto_complete":false,"complete_on_end_time":true,"complete_on_sample_size":false,"complete_on_confidence":false,"manual_override":false,"locked_until":null,"locked_by":null,"lock_reason":null}'::jsonb,
    now() - interval '10 days', now(),
    'time', 72, 25,
    'revenue', NULL,
    true, false, false, true, 0.20
  ),
  (
    '10000000-0000-0000-0000-000000000004',
    'Seed Confirmable Winner Test',
    'Running bandit fixture with a confirmable winner recommendation for admin contract actions.',
    'running',
    now() - interval '6 days',
    now() + interval '8 days',
    'thompson_sampling', true, 20, 0.95, 0.98,
    '{"enabled":true,"auto_start":true,"auto_complete":true,"complete_on_end_time":true,"complete_on_sample_size":false,"complete_on_confidence":false,"manual_override":false,"locked_until":null,"locked_by":null,"lock_reason":null}'::jsonb,
    now() - interval '6 days', now(),
    'events', 250, 20,
    'conversion', NULL,
    false, false, false, false, 0.20
  )
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    start_at = EXCLUDED.start_at,
    end_at = EXCLUDED.end_at,
    algorithm_type = EXCLUDED.algorithm_type,
    is_bandit = EXCLUDED.is_bandit,
    min_sample_size = EXCLUDED.min_sample_size,
    confidence_threshold = EXCLUDED.confidence_threshold,
    winner_confidence = EXCLUDED.winner_confidence,
    automation_policy = EXCLUDED.automation_policy,
    updated_at = now(),
    window_type = EXCLUDED.window_type,
    window_size = EXCLUDED.window_size,
    window_min_samples = EXCLUDED.window_min_samples,
    objective_type = EXCLUDED.objective_type,
    objective_weights = EXCLUDED.objective_weights,
    price_normalization = EXCLUDED.price_normalization,
    enable_contextual = EXCLUDED.enable_contextual,
    enable_delayed = EXCLUDED.enable_delayed,
    enable_currency = EXCLUDED.enable_currency,
    exploration_alpha = EXCLUDED.exploration_alpha;

INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight, created_at, updated_at)
VALUES
  ('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'control_monthly', 'Baseline monthly paywall.', true, 0.34, now() - interval '3 days', now()),
  ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 'annual_anchor', 'Annual-first pricing anchor.', false, 0.33, now() - interval '3 days', now()),
  ('20000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 'lifetime_offer', 'Lifetime CTA with premium framing.', false, 0.33, now() - interval '3 days', now()),
  ('20000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000002', 'control_copy', 'Current onboarding copy.', true, 0.50, now() - interval '1 day', now()),
  ('20000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000002', 'benefit_led_copy', 'Benefit-led onboarding CTA copy.', false, 0.50, now() - interval '1 day', now()),
  ('20000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000003', 'control_winback', 'Baseline winback offer.', true, 0.40, now() - interval '10 days', now()),
  ('20000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000003', 'discount_20', '20 percent winback discount.', false, 0.30, now() - interval '10 days', now()),
  ('20000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000003', 'discount_40', '40 percent winback discount.', false, 0.30, now() - interval '10 days', now()),
  ('20000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000004', 'control_checkout', 'Baseline checkout flow.', true, 0.50, now() - interval '6 days', now()),
  ('20000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000004', 'winner_checkout', 'High-converting checkout variant.', false, 0.50, now() - interval '6 days', now())
ON CONFLICT (id) DO UPDATE
SET experiment_id = EXCLUDED.experiment_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_control = EXCLUDED.is_control,
    traffic_weight = EXCLUDED.traffic_weight,
    updated_at = now();

INSERT INTO ab_test_arm_stats (
  arm_id, alpha, beta, samples, conversions, revenue, avg_reward, updated_at, revenue_usd, original_currency, original_revenue
)
VALUES
  ('20000000-0000-0000-0000-000000000001', 22.00, 78.00, 120, 21, 260.00, 2.1667, now(), 260.00, 'USD', 260.00),
  ('20000000-0000-0000-0000-000000000002', 30.00, 70.00, 130, 29, 540.00, 4.1538, now(), 540.00, 'USD', 540.00),
  ('20000000-0000-0000-0000-000000000003', 36.00, 64.00, 140, 35, 940.00, 6.7143, now(), 940.00, 'USD', 940.00),
  ('20000000-0000-0000-0000-000000000006', 14.00, 36.00, 48, 13, 180.00, 3.7500, now(), 180.00, 'USD', 180.00),
  ('20000000-0000-0000-0000-000000000007', 18.00, 32.00, 50, 17, 265.00, 5.3000, now(), 265.00, 'USD', 265.00),
  ('20000000-0000-0000-0000-000000000008', 12.00, 28.00, 38, 11, 220.00, 5.7895, now(), 220.00, 'USD', 220.00),
  ('20000000-0000-0000-0000-000000000009', 20.00, 40.00, 58, 19, 420.00, 7.2414, now(), 420.00, 'USD', 420.00),
  ('20000000-0000-0000-0000-000000000010', 44.00, 16.00, 62, 43, 980.00, 15.8065, now(), 980.00, 'USD', 980.00)
ON CONFLICT (arm_id) DO UPDATE
SET alpha = EXCLUDED.alpha,
    beta = EXCLUDED.beta,
    samples = EXCLUDED.samples,
    conversions = EXCLUDED.conversions,
    revenue = EXCLUDED.revenue,
    avg_reward = EXCLUDED.avg_reward,
    updated_at = now(),
    revenue_usd = EXCLUDED.revenue_usd,
    original_currency = EXCLUDED.original_currency,
    original_revenue = EXCLUDED.original_revenue;

INSERT INTO users (id, platform_user_id, platform, app_version, email, role, ltv, created_at)
VALUES
  ('33333333-3333-4333-8333-000000000001', 'usr_contract_bandit_1', 'ios', '3.0.0', 'bandit-contract-1@seed.example.com', 'user', 49.99, now() - interval '14 days'),
  ('33333333-3333-4333-8333-000000000002', 'usr_contract_bandit_2', 'android', '3.0.0', 'bandit-contract-2@seed.example.com', 'user', 79.99, now() - interval '10 days'),
  ('33333333-3333-4333-8333-000000000003', 'usr_contract_bandit_3', 'ios', '3.0.0', 'bandit-contract-3@seed.example.com', 'user', 129.99, now() - interval '7 days')
ON CONFLICT (id) DO UPDATE
SET platform_user_id = EXCLUDED.platform_user_id,
    platform = EXCLUDED.platform,
    app_version = EXCLUDED.app_version,
    email = EXCLUDED.email,
    role = EXCLUDED.role,
    ltv = EXCLUDED.ltv;

INSERT INTO ab_tests (
  id, name, description, status, start_at, end_at,
  algorithm_type, is_bandit, min_sample_size, confidence_threshold, winner_confidence,
  automation_policy,
  created_at, updated_at,
  window_type, window_size, window_min_samples,
  objective_type, objective_weights,
  price_normalization, enable_contextual, enable_delayed, enable_currency, exploration_alpha
)
VALUES (
  '11111111-1111-4111-8111-111111111111',
  'Bandit Contract Fixture',
  'Schema-valid UUID contract fixture for bandit Schemathesis coverage.',
  'running',
  now() - interval '5 days',
  now() + interval '10 days',
  'thompson_sampling', true, 25, 0.95, NULL,
  '{"enabled":true,"auto_start":true,"auto_complete":false,"complete_on_end_time":true,"complete_on_sample_size":false,"complete_on_confidence":false,"manual_override":false,"locked_until":null,"locked_by":null,"lock_reason":null}'::jsonb,
  now() - interval '5 days', now(),
  'events', 250, 25,
  'hybrid', '{"conversion":0.5,"ltv":0.2,"revenue":0.3}'::jsonb,
  true, true, true, true, 0.25
)
ON CONFLICT (id) DO UPDATE
SET status = EXCLUDED.status,
    algorithm_type = EXCLUDED.algorithm_type,
    is_bandit = EXCLUDED.is_bandit,
    min_sample_size = EXCLUDED.min_sample_size,
    confidence_threshold = EXCLUDED.confidence_threshold,
    automation_policy = EXCLUDED.automation_policy,
    updated_at = now(),
    window_type = EXCLUDED.window_type,
    window_size = EXCLUDED.window_size,
    window_min_samples = EXCLUDED.window_min_samples,
    objective_type = EXCLUDED.objective_type,
    objective_weights = EXCLUDED.objective_weights,
    price_normalization = EXCLUDED.price_normalization,
    enable_contextual = EXCLUDED.enable_contextual,
    enable_delayed = EXCLUDED.enable_delayed,
    enable_currency = EXCLUDED.enable_currency,
    exploration_alpha = EXCLUDED.exploration_alpha;

INSERT INTO ab_test_arms (id, experiment_id, name, description, is_control, traffic_weight, created_at, updated_at)
VALUES
  ('22222222-2222-4222-8222-000000000001', '11111111-1111-4111-8111-111111111111', 'contract_control', 'Contract fixture control arm.', true, 0.34, now() - interval '5 days', now()),
  ('22222222-2222-4222-8222-000000000002', '11111111-1111-4111-8111-111111111111', 'contract_anchor', 'Contract fixture anchor arm.', false, 0.33, now() - interval '5 days', now()),
  ('22222222-2222-4222-8222-000000000003', '11111111-1111-4111-8111-111111111111', 'contract_lifetime', 'Contract fixture lifetime arm.', false, 0.33, now() - interval '5 days', now())
ON CONFLICT (id) DO UPDATE
SET experiment_id = EXCLUDED.experiment_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_control = EXCLUDED.is_control,
    traffic_weight = EXCLUDED.traffic_weight,
    updated_at = now();

INSERT INTO ab_test_arm_stats (arm_id, alpha, beta, samples, conversions, revenue, avg_reward, updated_at, revenue_usd, original_currency, original_revenue)
VALUES
  ('22222222-2222-4222-8222-000000000001', 24.00, 76.00, 100, 23, 310.00, 3.1000, now(), 310.00, 'USD', 310.00),
  ('22222222-2222-4222-8222-000000000002', 31.00, 69.00, 110, 30, 520.00, 4.7273, now(), 520.00, 'USD', 520.00),
  ('22222222-2222-4222-8222-000000000003', 37.00, 63.00, 120, 36, 890.00, 7.4167, now(), 890.00, 'USD', 890.00)
ON CONFLICT (arm_id) DO UPDATE
SET alpha = EXCLUDED.alpha,
    beta = EXCLUDED.beta,
    samples = EXCLUDED.samples,
    conversions = EXCLUDED.conversions,
    revenue = EXCLUDED.revenue,
    avg_reward = EXCLUDED.avg_reward,
    updated_at = now(),
    revenue_usd = EXCLUDED.revenue_usd,
    original_currency = EXCLUDED.original_currency,
    original_revenue = EXCLUDED.original_revenue;

INSERT INTO bandit_user_context (user_id, country, device, app_version, days_since_install, total_spent, last_purchase_at, updated_at)
VALUES
  ('33333333-3333-4333-8333-000000000001', 'US', 'ios', '3.0.0', 14, 49.99, now() - interval '2 days', now()),
  ('33333333-3333-4333-8333-000000000002', 'DE', 'android', '3.0.0', 10, 79.99, now() - interval '3 days', now()),
  ('33333333-3333-4333-8333-000000000003', 'GB', 'ios', '3.0.0', 7, 129.99, now() - interval '1 days', now())
ON CONFLICT (user_id) DO UPDATE
SET country = EXCLUDED.country,
    device = EXCLUDED.device,
    app_version = EXCLUDED.app_version,
    days_since_install = EXCLUDED.days_since_install,
    total_spent = EXCLUDED.total_spent,
    last_purchase_at = EXCLUDED.last_purchase_at,
    updated_at = now();

INSERT INTO bandit_arm_context_model (arm_id, dimension, matrix_a, vector_b, theta, samples_count, updated_at)
VALUES
  ('22222222-2222-4222-8222-000000000001', 20, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, 100, now()),
  ('22222222-2222-4222-8222-000000000002', 20, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, 110, now()),
  ('22222222-2222-4222-8222-000000000003', 20, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, 120, now())
ON CONFLICT (arm_id) DO UPDATE
SET dimension = EXCLUDED.dimension,
    matrix_a = EXCLUDED.matrix_a,
    vector_b = EXCLUDED.vector_b,
    theta = EXCLUDED.theta,
    samples_count = EXCLUDED.samples_count,
    updated_at = now();

INSERT INTO ab_test_assignments (id, experiment_id, user_id, arm_id, assigned_at, expires_at)
VALUES
  (gen_random_uuid(), '11111111-1111-4111-8111-111111111111', '33333333-3333-4333-8333-000000000001', '22222222-2222-4222-8222-000000000001', now() - interval '6 hours', now() + interval '18 hours'),
  (gen_random_uuid(), '11111111-1111-4111-8111-111111111111', '33333333-3333-4333-8333-000000000002', '22222222-2222-4222-8222-000000000002', now() - interval '5 hours', now() + interval '18 hours'),
  (gen_random_uuid(), '11111111-1111-4111-8111-111111111111', '33333333-3333-4333-8333-000000000003', '22222222-2222-4222-8222-000000000003', now() - interval '4 hours', now() + interval '18 hours')
ON CONFLICT (experiment_id, user_id) DO UPDATE
SET arm_id = EXCLUDED.arm_id,
    assigned_at = EXCLUDED.assigned_at,
    expires_at = EXCLUDED.expires_at;

INSERT INTO bandit_pending_rewards (id, experiment_id, arm_id, user_id, assigned_at, expires_at, converted, conversion_value, conversion_currency, converted_at, processed_at)
VALUES
  ('44444444-4444-4444-8444-000000000001', '11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-000000000001', '33333333-3333-4333-8333-000000000001', now() - interval '8 hours', now() + interval '12 hours', false, NULL, NULL, NULL, NULL),
  ('44444444-4444-4444-8444-000000000002', '11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-000000000002', '33333333-3333-4333-8333-000000000002', now() - interval '7 hours', now() + interval '12 hours', false, NULL, NULL, NULL, NULL),
  ('44444444-4444-4444-8444-000000000003', '11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-000000000003', '33333333-3333-4333-8333-000000000003', now() - interval '6 hours', now() + interval '12 hours', true, 89.99, 'USD', now() - interval '2 hours', now() - interval '90 minutes')
ON CONFLICT (id) DO UPDATE
SET arm_id = EXCLUDED.arm_id,
    user_id = EXCLUDED.user_id,
    assigned_at = EXCLUDED.assigned_at,
    expires_at = EXCLUDED.expires_at,
    converted = EXCLUDED.converted,
    conversion_value = EXCLUDED.conversion_value,
    conversion_currency = EXCLUDED.conversion_currency,
    converted_at = EXCLUDED.converted_at,
    processed_at = EXCLUDED.processed_at;

INSERT INTO bandit_conversion_links (pending_id, transaction_id)
VALUES ('44444444-4444-4444-8444-000000000003', '55555555-5555-4555-8555-000000000001')
ON CONFLICT (pending_id, transaction_id) DO NOTHING;

INSERT INTO bandit_arm_objective_stats (arm_id, objective_type, alpha, beta, samples, conversions, total_revenue, avg_ltv, updated_at)
VALUES
  ('22222222-2222-4222-8222-000000000001', 'conversion', 24.00, 76.00, 100, 23, 310.00, 36.00, now()),
  ('22222222-2222-4222-8222-000000000001', 'ltv',        19.00, 81.00, 100, 19, 310.00, 33.00, now()),
  ('22222222-2222-4222-8222-000000000001', 'revenue',    26.00, 74.00, 100, 23, 310.00, 36.00, now()),
  ('22222222-2222-4222-8222-000000000002', 'conversion', 31.00, 69.00, 110, 30, 520.00, 48.00, now()),
  ('22222222-2222-4222-8222-000000000002', 'ltv',        26.00, 74.00, 110, 25, 520.00, 45.00, now()),
  ('22222222-2222-4222-8222-000000000002', 'revenue',    34.00, 66.00, 110, 30, 520.00, 48.00, now()),
  ('22222222-2222-4222-8222-000000000003', 'conversion', 37.00, 63.00, 120, 36, 890.00, 71.00, now()),
  ('22222222-2222-4222-8222-000000000003', 'ltv',        32.00, 68.00, 120, 29, 890.00, 67.00, now()),
  ('22222222-2222-4222-8222-000000000003', 'revenue',    40.00, 60.00, 120, 36, 890.00, 71.00, now())
ON CONFLICT (arm_id, objective_type) DO UPDATE
SET alpha = EXCLUDED.alpha,
    beta = EXCLUDED.beta,
    samples = EXCLUDED.samples,
    conversions = EXCLUDED.conversions,
    total_revenue = EXCLUDED.total_revenue,
    avg_ltv = EXCLUDED.avg_ltv,
    updated_at = now();

INSERT INTO bandit_window_events (experiment_id, arm_id, user_id, event_type, reward_value, timestamp)
SELECT * FROM (VALUES
  ('11111111-1111-4111-8111-111111111111'::uuid, '22222222-2222-4222-8222-000000000001'::uuid, '33333333-3333-4333-8333-000000000001'::uuid, 'impression', NULL::numeric, now() - interval '9 hours'),
  ('11111111-1111-4111-8111-111111111111'::uuid, '22222222-2222-4222-8222-000000000002'::uuid, '33333333-3333-4333-8333-000000000002'::uuid, 'conversion', 89.99::numeric, now() - interval '7 hours'),
  ('11111111-1111-4111-8111-111111111111'::uuid, '22222222-2222-4222-8222-000000000003'::uuid, '33333333-3333-4333-8333-000000000003'::uuid, 'no_conversion', NULL::numeric, now() - interval '5 hours')
) AS seeded(experiment_id, arm_id, user_id, event_type, reward_value, timestamp)
WHERE NOT EXISTS (
  SELECT 1 FROM bandit_window_events WHERE experiment_id = '11111111-1111-4111-8111-111111111111'::uuid
);

WITH seeded_users AS (
  SELECT id, row_number() OVER (ORDER BY created_at, email) AS rn
  FROM users
  WHERE email LIKE '%@seed.example.com'
  ORDER BY created_at, email
  LIMIT 12
)
INSERT INTO bandit_user_context (
  user_id, country, device, app_version, days_since_install, total_spent, last_purchase_at, updated_at
)
SELECT
  id,
  (ARRAY['US','DE','GB','CA','BR','JP'])[1 + ((rn - 1) % 6)],
  (ARRAY['ios','android','web'])[1 + ((rn - 1) % 3)],
  (ARRAY['2.1.0','2.2.0','2.3.0','3.0.0'])[1 + ((rn - 1) % 4)],
  7 * rn,
  ROUND((14.99 + rn * 11)::numeric, 2),
  now() - ((rn * 4) || ' days')::interval,
  now()
FROM seeded_users
ON CONFLICT (user_id) DO UPDATE
SET country = EXCLUDED.country,
    device = EXCLUDED.device,
    app_version = EXCLUDED.app_version,
    days_since_install = EXCLUDED.days_since_install,
    total_spent = EXCLUDED.total_spent,
    last_purchase_at = EXCLUDED.last_purchase_at,
    updated_at = now();

INSERT INTO bandit_arm_context_model (arm_id, dimension, matrix_a, vector_b, theta, samples_count, updated_at)
VALUES
  ('20000000-0000-0000-0000-000000000001', 20, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, 120, now()),
  ('20000000-0000-0000-0000-000000000002', 20, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, 130, now()),
  ('20000000-0000-0000-0000-000000000003', 20, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, 140, now())
ON CONFLICT (arm_id) DO UPDATE
SET dimension = EXCLUDED.dimension,
    matrix_a = EXCLUDED.matrix_a,
    vector_b = EXCLUDED.vector_b,
    theta = EXCLUDED.theta,
    samples_count = EXCLUDED.samples_count,
    updated_at = now();

WITH seeded_users AS (
  SELECT id, row_number() OVER (ORDER BY created_at, email) AS rn
  FROM users
  WHERE email LIKE '%@seed.example.com'
  ORDER BY created_at, email
  LIMIT 12
)
INSERT INTO ab_test_assignments (id, experiment_id, user_id, arm_id, assigned_at, expires_at)
SELECT
  gen_random_uuid(),
  '10000000-0000-0000-0000-000000000001',
  id,
  CASE
    WHEN rn % 3 = 1 THEN '20000000-0000-0000-0000-000000000001'::uuid
    WHEN rn % 3 = 2 THEN '20000000-0000-0000-0000-000000000002'::uuid
    ELSE '20000000-0000-0000-0000-000000000003'::uuid
  END,
  now() - ((rn + 1) || ' hours')::interval,
  now() + interval '23 hours'
FROM seeded_users
ON CONFLICT (experiment_id, user_id) DO UPDATE
SET arm_id = EXCLUDED.arm_id,
    assigned_at = EXCLUDED.assigned_at,
    expires_at = EXCLUDED.expires_at;

WITH seeded_users AS (
  SELECT id, row_number() OVER (ORDER BY created_at, email) AS rn
  FROM users
  WHERE email LIKE '%@seed.example.com'
  ORDER BY created_at, email
  LIMIT 6
)
INSERT INTO ab_test_assignments (id, experiment_id, user_id, arm_id, assigned_at, expires_at)
SELECT
  gen_random_uuid(),
  '10000000-0000-0000-0000-000000000003',
  id,
  CASE
    WHEN rn <= 2 THEN '20000000-0000-0000-0000-000000000006'::uuid
    WHEN rn <= 4 THEN '20000000-0000-0000-0000-000000000007'::uuid
    ELSE '20000000-0000-0000-0000-000000000008'::uuid
  END,
  now() - ((rn + 3) || ' hours')::interval,
  now() + interval '12 hours'
FROM seeded_users
ON CONFLICT (experiment_id, user_id) DO UPDATE
SET arm_id = EXCLUDED.arm_id,
    assigned_at = EXCLUDED.assigned_at,
    expires_at = EXCLUDED.expires_at;

WITH pending_users AS (
  SELECT id, row_number() OVER (ORDER BY created_at, email) AS rn
  FROM users
  WHERE email LIKE '%@seed.example.com'
  ORDER BY created_at, email
  LIMIT 3
)
INSERT INTO bandit_pending_rewards (
  id, experiment_id, arm_id, user_id, assigned_at, expires_at,
  converted, conversion_value, conversion_currency, converted_at, processed_at
)
SELECT * FROM (
  SELECT
    CASE rn
      WHEN 1 THEN '30000000-0000-0000-0000-000000000001'::uuid
      WHEN 2 THEN '30000000-0000-0000-0000-000000000002'::uuid
      ELSE '30000000-0000-0000-0000-000000000003'::uuid
    END,
    '10000000-0000-0000-0000-000000000001'::uuid,
    CASE rn
      WHEN 1 THEN '20000000-0000-0000-0000-000000000003'::uuid
      WHEN 2 THEN '20000000-0000-0000-0000-000000000002'::uuid
      ELSE '20000000-0000-0000-0000-000000000001'::uuid
    END,
    id,
    now() - ((rn * 6) || ' hours')::interval,
    CASE WHEN rn = 1 THEN now() - interval '2 hours' ELSE now() + interval '18 hours' END,
    rn = 3,
    CASE WHEN rn = 3 THEN 79.99 ELSE NULL END,
    CASE WHEN rn = 3 THEN 'USD' ELSE NULL END,
    CASE WHEN rn = 3 THEN now() - interval '1 hour' ELSE NULL END,
    CASE WHEN rn = 3 THEN now() - interval '30 minutes' ELSE NULL END
  FROM pending_users
) AS seeded(id, experiment_id, arm_id, user_id, assigned_at, expires_at, converted, conversion_value, conversion_currency, converted_at, processed_at)
ON CONFLICT (id) DO UPDATE
SET arm_id = EXCLUDED.arm_id,
    user_id = EXCLUDED.user_id,
    assigned_at = EXCLUDED.assigned_at,
    expires_at = EXCLUDED.expires_at,
    converted = EXCLUDED.converted,
    conversion_value = EXCLUDED.conversion_value,
    conversion_currency = EXCLUDED.conversion_currency,
    converted_at = EXCLUDED.converted_at,
    processed_at = EXCLUDED.processed_at;

INSERT INTO bandit_conversion_links (pending_id, transaction_id)
VALUES ('30000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000001')
ON CONFLICT (pending_id, transaction_id) DO NOTHING;

INSERT INTO bandit_arm_objective_stats (
  arm_id, objective_type, alpha, beta, samples, conversions, total_revenue, avg_ltv, updated_at
)
VALUES
  ('20000000-0000-0000-0000-000000000001', 'conversion', 22.00, 78.00, 120, 21, 260.00, 34.00, now()),
  ('20000000-0000-0000-0000-000000000001', 'ltv',        18.00, 82.00, 120, 18, 260.00, 29.50, now()),
  ('20000000-0000-0000-0000-000000000001', 'revenue',    24.00, 76.00, 120, 21, 260.00, 34.00, now()),
  ('20000000-0000-0000-0000-000000000002', 'conversion', 30.00, 70.00, 130, 29, 540.00, 49.00, now()),
  ('20000000-0000-0000-0000-000000000002', 'ltv',        25.00, 75.00, 130, 24, 540.00, 44.50, now()),
  ('20000000-0000-0000-0000-000000000002', 'revenue',    33.00, 67.00, 130, 29, 540.00, 49.00, now()),
  ('20000000-0000-0000-0000-000000000003', 'conversion', 36.00, 64.00, 140, 35, 940.00, 72.00, now()),
  ('20000000-0000-0000-0000-000000000003', 'ltv',        30.00, 70.00, 140, 27, 940.00, 66.50, now()),
  ('20000000-0000-0000-0000-000000000003', 'revenue',    39.00, 61.00, 140, 35, 940.00, 72.00, now())
ON CONFLICT (arm_id, objective_type) DO UPDATE
SET alpha = EXCLUDED.alpha,
    beta = EXCLUDED.beta,
    samples = EXCLUDED.samples,
    conversions = EXCLUDED.conversions,
    total_revenue = EXCLUDED.total_revenue,
    avg_ltv = EXCLUDED.avg_ltv,
    updated_at = now();

WITH seeded_users AS (
  SELECT id, row_number() OVER (ORDER BY created_at, email) AS rn
  FROM users
  WHERE email LIKE '%@seed.example.com'
  ORDER BY created_at, email
  LIMIT 9
)
INSERT INTO bandit_window_events (experiment_id, arm_id, user_id, event_type, reward_value, timestamp)
SELECT
  '10000000-0000-0000-0000-000000000001'::uuid,
  CASE
    WHEN rn IN (1, 4, 7) THEN '20000000-0000-0000-0000-000000000001'::uuid
    WHEN rn IN (2, 5, 8) THEN '20000000-0000-0000-0000-000000000002'::uuid
    ELSE '20000000-0000-0000-0000-000000000003'::uuid
  END,
  id,
  CASE WHEN rn IN (2, 3, 5, 6, 9) THEN 'conversion' WHEN rn IN (1, 4, 7) THEN 'impression' ELSE 'no_conversion' END,
  CASE
    WHEN rn IN (2, 5) THEN 79.99
    WHEN rn IN (3, 6, 9) THEN 199.99
    ELSE NULL
  END,
  now() - ((10 - rn) || ' hours')::interval
FROM seeded_users
WHERE NOT EXISTS (
  SELECT 1 FROM bandit_window_events WHERE experiment_id = '10000000-0000-0000-0000-000000000001'::uuid
);

COMMIT;

SELECT
  (SELECT COUNT(*) FROM pricing_tiers WHERE deleted_at IS NULL) AS pricing_tiers,
  (SELECT COUNT(*) FROM users WHERE email LIKE '%@seed.example.com') AS seed_users,
  (SELECT COUNT(*) FROM ab_tests WHERE id::text LIKE '10000000-%') AS seed_experiments,
  (SELECT COUNT(*) FROM ab_test_arms WHERE experiment_id::text LIKE '10000000-%') AS seed_arms,
  (SELECT COUNT(*) FROM ab_test_arm_stats WHERE arm_id::text LIKE '20000000-%') AS seeded_arm_stats,
  (SELECT COUNT(*) FROM bandit_pending_rewards WHERE experiment_id = '10000000-0000-0000-0000-000000000001') AS pending_rewards,
  (SELECT COUNT(*) FROM bandit_window_events WHERE experiment_id = '10000000-0000-0000-0000-000000000001') AS window_events;