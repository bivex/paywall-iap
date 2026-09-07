#!/usr/bin/env bash
# scripts/seed_tiers.sh — seed demo pricing_tiers for Paywall Creator
# Works with direct DB access OR via Docker container.
#
# Usage (direct DB):
#   DATABASE_URL=postgresql://postgres:postgres@localhost:5432/iap_db ./scripts/seed_tiers.sh
#
# Usage (Docker container):
#   DB_CONTAINER=paywall-db-1 ./scripts/seed_tiers.sh

set -euo pipefail

DB_CONTAINER="${DB_CONTAINER:-paywall-db-1}"
DATABASE_URL="${DATABASE_URL:-}"
DB_NAME="${DB_NAME:-iap_db}"
DB_USER="${DB_USER:-postgres}"

need() {
  command -v "$1" >/dev/null || {
    echo "Error: required command not found: $1" >&2
    exit 1
  }
}

run_sql() {
  if [[ -n "$DATABASE_URL" ]]; then
    need psql
    psql "$DATABASE_URL"
  else
    need docker
    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME"
  fi
}

echo "🌱 Seeding pricing_tiers..."
if [[ -n "$DATABASE_URL" ]]; then
  echo "   Mode: direct DATABASE_URL"
else
  echo "   Mode: docker container ($DB_CONTAINER)"
fi

run_sql <<'SQL'
BEGIN;

INSERT INTO pricing_tiers (
  app_id,
  name,
  description,
  monthly_price,
  annual_price,
  lifetime_price,
  currency,
  features,
  is_active,
  updated_at,
  deleted_at
)
SELECT
  a.id AS app_id,
  t.name,
  t.description,
  t.monthly_price,
  t.annual_price,
  t.lifetime_price,
  t.currency,
  t.features::jsonb,
  t.is_active,
  now(),
  NULL
FROM (SELECT id FROM apps) a
CROSS JOIN (
  VALUES
    (
      'Starter',
      'For solo makers getting started with subscriptions',
      9.99::numeric,
      79.99::numeric,
      NULL::numeric,
      'USD',
      '["Ad-free experience", "Unlimited access", "Basic analytics", "Email support"]',
      true
    ),
    (
      'Growth',
      'For growing teams that need collaboration and analytics',
      29.99::numeric,
      239.99::numeric,
      NULL::numeric,
      'USD',
      '["Everything in Starter", "Team collaboration", "Advanced dashboards", "Priority support"]',
      true
    ),
    (
      'Scale',
      'For larger organizations with premium support needs',
      79.99::numeric,
      639.99::numeric,
      NULL::numeric,
      'USD',
      '["Everything in Growth", "Dedicated success manager", "Custom integrations", "SLA & onboarding"]',
      true
    ),
    (
      'No Ads Lifetime',
      'Lifetime Ad-free experience for Demo Game',
      NULL::numeric,
      NULL::numeric,
      2.99::numeric,
      'USD',
      '["No forced ads", "Bottom banner ads disabled", "Keep rewarded ads", "+500 gems bonus"]',
      true
    )
) AS t(name, description, monthly_price, annual_price, lifetime_price, currency, features, is_active)
ON CONFLICT (app_id, name) WHERE deleted_at IS NULL DO UPDATE
SET description   = EXCLUDED.description,
    monthly_price = EXCLUDED.monthly_price,
    annual_price  = EXCLUDED.annual_price,
    lifetime_price = EXCLUDED.lifetime_price,
    currency      = EXCLUDED.currency,
    features      = EXCLUDED.features,
    is_active     = EXCLUDED.is_active,
    updated_at    = now(),
    deleted_at    = NULL;

COMMIT;

SELECT
  name,
  monthly_price,
  annual_price,
  currency,
  is_active,
  jsonb_array_length(COALESCE(features, '[]'::jsonb)) AS feature_count
FROM pricing_tiers
WHERE deleted_at IS NULL
ORDER BY name;
SQL

echo ""
echo "✅ pricing_tiers seeded successfully"
echo "👉 Open: http://localhost:3000/dashboard/paywall-creator"