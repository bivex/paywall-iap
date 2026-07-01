-- Migration: 035_seed_experiments_per_app.up.sql
-- Purpose:   Seed demo A/B and Bandit experiments for each active Mothsalt app
--            so Experiment Studio and A/B Tests pages show real data.
-- Apps:      game1 (ios), game2 (ios), game3 (ios), game4 (android)

BEGIN;

-- ============================================================
-- Mothsalt Game 1 (com.mothsalt.game1)
-- ============================================================

WITH app AS (SELECT id FROM apps WHERE name = 'com.mothsalt.game1'),
     exp AS (
         INSERT INTO ab_tests (app_id, name, description, status, is_bandit, algorithm_type,
                               min_sample_size, confidence_threshold, start_at)
         SELECT app.id,
                'Paywall CTA — Game 1',
                'Classic A/B test: two paywall call-to-action variants for Game 1.',
                'running',
                false,
                NULL,
                200,
                0.95,
                now() - interval '7 days'
         FROM app
         RETURNING id, app_id
     )
INSERT INTO ab_test_arms (experiment_id, name, description, is_control, traffic_weight)
SELECT exp.id, 'Control',  'Original CTA: "Subscribe Now"', true,  0.50 FROM exp
UNION ALL
SELECT exp.id, 'Variant A', 'New CTA: "Start Free Trial"',  false, 0.50 FROM exp;

WITH app AS (SELECT id FROM apps WHERE name = 'com.mothsalt.game1'),
     exp AS (
         INSERT INTO ab_tests (app_id, name, description, status, is_bandit, algorithm_type,
                               min_sample_size, confidence_threshold, start_at)
         SELECT app.id,
                'Price Point Bandit — Game 1',
                'Thompson Sampling bandit optimising across three price points for Game 1.',
                'running',
                true,
                'thompson_sampling',
                150,
                0.90,
                now() - interval '3 days'
         FROM app
         RETURNING id, app_id
     )
INSERT INTO ab_test_arms (experiment_id, name, description, is_control, traffic_weight)
SELECT exp.id, '$4.99 / month',  'Baseline price',  true,  0.34 FROM exp
UNION ALL
SELECT exp.id, '$6.99 / month',  'Mid-tier price',  false, 0.33 FROM exp
UNION ALL
SELECT exp.id, '$9.99 / month',  'Premium price',   false, 0.33 FROM exp;

-- ============================================================
-- Mothsalt Game 2 (com.mothsalt.game2)
-- ============================================================

WITH app AS (SELECT id FROM apps WHERE name = 'com.mothsalt.game2'),
     exp AS (
         INSERT INTO ab_tests (app_id, name, description, status, is_bandit, algorithm_type,
                               min_sample_size, confidence_threshold, start_at)
         SELECT app.id,
                'Onboarding Paywall — Game 2',
                'A/B test: show paywall on session 1 vs session 3 for Game 2.',
                'running',
                false,
                NULL,
                300,
                0.95,
                now() - interval '14 days'
         FROM app
         RETURNING id, app_id
     )
INSERT INTO ab_test_arms (experiment_id, name, description, is_control, traffic_weight)
SELECT exp.id, 'Session 1 (control)', 'Show paywall on first session', true,  0.50 FROM exp
UNION ALL
SELECT exp.id, 'Session 3',           'Delay paywall to third session', false, 0.50 FROM exp;

WITH app AS (SELECT id FROM apps WHERE name = 'com.mothsalt.game2'),
     exp AS (
         INSERT INTO ab_tests (app_id, name, description, status, is_bandit, algorithm_type,
                               min_sample_size, confidence_threshold, start_at)
         SELECT app.id,
                'Discount Bandit — Game 2',
                'UCB bandit: test discount offer vs no-discount vs limited-time badge for Game 2.',
                'running',
                true,
                'ucb',
                200,
                0.90,
                now() - interval '5 days'
         FROM app
         RETURNING id, app_id
     )
INSERT INTO ab_test_arms (experiment_id, name, description, is_control, traffic_weight)
SELECT exp.id, 'No discount',         'Standard offer',              true,  0.34 FROM exp
UNION ALL
SELECT exp.id, '20% discount',        'Flat 20% off first month',    false, 0.33 FROM exp
UNION ALL
SELECT exp.id, 'Limited-time badge',  'Urgency badge, no discount',  false, 0.33 FROM exp;

-- ============================================================
-- Mothsalt Game 3 (com.mothsalt.game3)
-- ============================================================

WITH app AS (SELECT id FROM apps WHERE name = 'com.mothsalt.game3'),
     exp AS (
         INSERT INTO ab_tests (app_id, name, description, status, is_bandit, algorithm_type,
                               min_sample_size, confidence_threshold, start_at)
         SELECT app.id,
                'Annual Upsell — Game 3',
                'A/B test: annual plan prominence on paywall for Game 3.',
                'draft',
                false,
                NULL,
                250,
                0.95,
                NULL
         FROM app
         RETURNING id, app_id
     )
INSERT INTO ab_test_arms (experiment_id, name, description, is_control, traffic_weight)
SELECT exp.id, 'Monthly highlight (control)', 'Monthly plan shown first',  true,  0.50 FROM exp
UNION ALL
SELECT exp.id, 'Annual highlight',            'Annual plan shown first',   false, 0.50 FROM exp;

WITH app AS (SELECT id FROM apps WHERE name = 'com.mothsalt.game3'),
     exp AS (
         INSERT INTO ab_tests (app_id, name, description, status, is_bandit, algorithm_type,
                               min_sample_size, confidence_threshold, start_at)
         SELECT app.id,
                'Feature Gate Bandit — Game 3',
                'Epsilon-greedy bandit: which feature gate drives most conversions in Game 3.',
                'running',
                true,
                'epsilon_greedy',
                100,
                0.85,
                now() - interval '10 days'
         FROM app
         RETURNING id, app_id
     )
INSERT INTO ab_test_arms (experiment_id, name, description, is_control, traffic_weight)
SELECT exp.id, 'No gate (control)', 'No feature gating',           true,  0.34 FROM exp
UNION ALL
SELECT exp.id, 'PvP gate',          'Gate PvP behind subscription', false, 0.33 FROM exp
UNION ALL
SELECT exp.id, 'Cosmetics gate',    'Gate cosmetics only',          false, 0.33 FROM exp;

-- ============================================================
-- Mothsalt Game 4 (com.mothsalt.game4, android)
-- ============================================================

WITH app AS (SELECT id FROM apps WHERE name = 'com.mothsalt.game4'),
     exp AS (
         INSERT INTO ab_tests (app_id, name, description, status, is_bandit, algorithm_type,
                               min_sample_size, confidence_threshold, start_at)
         SELECT app.id,
                'Play Pass Positioning — Game 4',
                'A/B test: Google Play Pass badge vs standard paywall copy for Game 4.',
                'running',
                false,
                NULL,
                400,
                0.95,
                now() - interval '20 days'
         FROM app
         RETURNING id, app_id
     )
INSERT INTO ab_test_arms (experiment_id, name, description, is_control, traffic_weight)
SELECT exp.id, 'Standard copy (control)', 'Default paywall text',          true,  0.50 FROM exp
UNION ALL
SELECT exp.id, 'Play Pass badge',         'Emphasise Play Pass eligibility', false, 0.50 FROM exp;

WITH app AS (SELECT id FROM apps WHERE name = 'com.mothsalt.game4'),
     exp AS (
         INSERT INTO ab_tests (app_id, name, description, status, is_bandit, algorithm_type,
                               min_sample_size, confidence_threshold, start_at)
         SELECT app.id,
                'Bundle Offer Bandit — Game 4',
                'Thompson Sampling: single-app vs bundle offer conversion for Game 4.',
                'running',
                true,
                'thompson_sampling',
                180,
                0.90,
                now() - interval '6 days'
         FROM app
         RETURNING id, app_id
     )
INSERT INTO ab_test_arms (experiment_id, name, description, is_control, traffic_weight)
SELECT exp.id, 'Single app (control)', 'Game 4 subscription only',      true,  0.50 FROM exp
UNION ALL
SELECT exp.id, 'Bundle offer',         '3-game bundle at reduced price', false, 0.50 FROM exp;

COMMIT;
