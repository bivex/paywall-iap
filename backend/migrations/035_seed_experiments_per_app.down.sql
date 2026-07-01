-- Migration: 035_seed_experiments_per_app.down.sql
-- Removes all seeded experiments (and their arms via CASCADE) for all 4 apps.

BEGIN;

DELETE FROM ab_tests
WHERE app_id IN (SELECT id FROM apps WHERE name IN (
    'com.mothsalt.game1',
    'com.mothsalt.game2',
    'com.mothsalt.game3',
    'com.mothsalt.game4'
))
AND name IN (
    'Paywall CTA — Game 1',
    'Price Point Bandit — Game 1',
    'Onboarding Paywall — Game 2',
    'Discount Bandit — Game 2',
    'Annual Upsell — Game 3',
    'Feature Gate Bandit — Game 3',
    'Play Pass Positioning — Game 4',
    'Bundle Offer Bandit — Game 4'
);

COMMIT;
