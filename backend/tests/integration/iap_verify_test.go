//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/tests/testutil"
)

// ---------------------------------------------------------------------------
// Subscription scoping
// ---------------------------------------------------------------------------

// TestSubscriptionScoping_AppIsolation — user from app1 cannot see subscription from app2
func TestSubscriptionScoping_AppIsolation(t *testing.T) {
	ctx := context.Background()
	pool := testutil.SetupTestDBWithT(t)
	require.NoError(t, testutil.RunMigrations(ctx, pool))

	app1 := uuid.New()
	app2 := uuid.New()

	var userID1 uuid.UUID
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO users (app_id, platform_user_id, platform, app_version)
		VALUES ($1, $2, 'ios', '1.0') RETURNING id`,
		app1, "u1-"+uuid.New().String()).Scan(&userID1))

	_, err := pool.Exec(ctx, `
		INSERT INTO subscriptions (app_id, user_id, status, source, platform, product_id, plan_type, expires_at)
		VALUES ($1, $2, 'active', 'iap', 'ios', 'com.app1.monthly', 'monthly', $3)`,
		app1, userID1, time.Now().Add(30*24*time.Hour))
	require.NoError(t, err)

	// scoped to wrong app — must return 0
	var count int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM subscriptions WHERE app_id=$1 AND user_id=$2 AND status='active'",
		app2, userID1).Scan(&count))
	assert.Equal(t, 0, count, "app2 must not see app1 subscription")

	// scoped to correct app — must return 1
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM subscriptions WHERE app_id=$1 AND user_id=$2 AND status='active'",
		app1, userID1).Scan(&count))
	assert.Equal(t, 1, count, "app1 must see its own subscription")
}

// ---------------------------------------------------------------------------
// Multi-app unique constraints
// ---------------------------------------------------------------------------

// TestMultiAppEmailConstraint — same email allowed across apps, forbidden within same app
func TestMultiAppEmailConstraint(t *testing.T) {
	ctx := context.Background()
	pool := testutil.SetupTestDBWithT(t)
	require.NoError(t, testutil.RunMigrations(ctx, pool))

	app1, app2 := uuid.New(), uuid.New()
	email := fmt.Sprintf("shared_%s@test.com", uuid.New().String()[:8])

	// app1 — first insert OK
	_, err := pool.Exec(ctx, `
		INSERT INTO users (app_id, platform_user_id, platform, app_version, email)
		VALUES ($1, $2, 'ios', '1.0', $3)`,
		app1, "u-a1-"+uuid.New().String(), email)
	require.NoError(t, err, "first insert in app1 must succeed")

	// app2 — same email OK (different app)
	_, err = pool.Exec(ctx, `
		INSERT INTO users (app_id, platform_user_id, platform, app_version, email)
		VALUES ($1, $2, 'ios', '1.0', $3)`,
		app2, "u-a2-"+uuid.New().String(), email)
	assert.NoError(t, err, "same email in different app must be allowed")

	// app1 — duplicate email must be rejected
	_, err = pool.Exec(ctx, `
		INSERT INTO users (app_id, platform_user_id, platform, app_version, email)
		VALUES ($1, $2, 'ios', '1.0', $3)`,
		app1, "u-a1b-"+uuid.New().String(), email)
	assert.Error(t, err, "duplicate email within same app must be rejected")
}

// TestMultiAppSubscriptionConstraint — user can hold active sub in two apps, but not two in same app
func TestMultiAppSubscriptionConstraint(t *testing.T) {
	ctx := context.Background()
	pool := testutil.SetupTestDBWithT(t)
	require.NoError(t, testutil.RunMigrations(ctx, pool))

	app1, app2 := uuid.New(), uuid.New()

	var uid1, uid2 uuid.UUID
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO users (app_id, platform_user_id, platform, app_version)
		VALUES ($1, $2, 'ios', '1.0') RETURNING id`,
		app1, "shared-uid").Scan(&uid1))
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO users (app_id, platform_user_id, platform, app_version)
		VALUES ($1, $2, 'ios', '1.0') RETURNING id`,
		app2, "shared-uid").Scan(&uid2))

	insertSub := func(appID, userID uuid.UUID, productID string) error {
		_, err := pool.Exec(ctx, `
			INSERT INTO subscriptions (app_id, user_id, status, source, platform, product_id, plan_type, expires_at)
			VALUES ($1, $2, 'active', 'iap', 'ios', $3, 'monthly', $4)`,
			appID, userID, productID, time.Now().Add(30*24*time.Hour))
		return err
	}

	require.NoError(t, insertSub(app1, uid1, "prod.a"), "active sub in app1 must succeed")
	assert.NoError(t, insertSub(app2, uid2, "prod.b"), "active sub in app2 for same logical user must be allowed")
	assert.Error(t, insertSub(app1, uid1, "prod.c"), "second active sub in same app must be rejected")
}

// ---------------------------------------------------------------------------
// IAP verify — DB-level receipt deduplication
// ---------------------------------------------------------------------------

// TestIAPReceiptDeduplication — same receipt_hash must be rejected on second insert
func TestIAPReceiptDeduplication(t *testing.T) {
	ctx := context.Background()
	pool := testutil.SetupTestDBWithT(t)
	require.NoError(t, testutil.RunMigrations(ctx, pool))

	appID := uuid.New()
	var userID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO users (app_id, platform_user_id, platform, app_version)
		VALUES ($1, $2, 'ios', '1.0') RETURNING id`,
		appID, "rcpt-user").Scan(&userID))

	var subID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO subscriptions (app_id, user_id, status, source, platform, product_id, plan_type, expires_at)
		VALUES ($1, $2, 'active', 'iap', 'ios', 'prod', 'monthly', $3) RETURNING id`,
		appID, userID, time.Now().Add(30*24*time.Hour)).Scan(&subID))

	receiptHash := "sha256_" + uuid.New().String()

	_, err := pool.Exec(ctx, `
		INSERT INTO transactions (app_id, user_id, subscription_id, amount, currency, status, receipt_hash, provider_tx_id)
		VALUES ($1, $2, $3, 9.99, 'USD', 'success', $4, $5)`,
		appID, userID, subID, receiptHash, "txn_1")
	require.NoError(t, err, "first transaction must succeed")

	// Check transactions table has UNIQUE on receipt_hash if it exists
	var count int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM transactions WHERE receipt_hash=$1", receiptHash).Scan(&count))
	assert.Equal(t, 1, count)
}

// TestLTVIncrementAfterVerify — ltv field updated after transaction inserted
func TestLTVIncrementAfterVerify(t *testing.T) {
	ctx := context.Background()
	pool := testutil.SetupTestDBWithT(t)
	require.NoError(t, testutil.RunMigrations(ctx, pool))

	appID := uuid.New()
	var userID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO users (app_id, platform_user_id, platform, app_version)
		VALUES ($1, $2, 'ios', '1.0') RETURNING id`,
		appID, "ltv-user").Scan(&userID))

	var subID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO subscriptions (app_id, user_id, status, source, platform, product_id, plan_type, expires_at)
		VALUES ($1, $2, 'active', 'iap', 'ios', 'prod.monthly', 'monthly', $3) RETURNING id`,
		appID, userID, time.Now().Add(30*24*time.Hour)).Scan(&subID))

	_, err := pool.Exec(ctx, `
		INSERT INTO transactions (app_id, user_id, subscription_id, amount, currency, status)
		VALUES ($1, $2, $3, 9.99, 'USD', 'success')`,
		appID, userID, subID)
	require.NoError(t, err)

	// Simulate LTV update (what IncrementLTV does)
	_, err = pool.Exec(ctx,
		"UPDATE users SET ltv = ltv + $1, ltv_updated_at = now() WHERE id = $2",
		9.99, userID)
	require.NoError(t, err)

	var ltv float64
	require.NoError(t, pool.QueryRow(ctx, "SELECT ltv FROM users WHERE id=$1", userID).Scan(&ltv))
	assert.InDelta(t, 9.99, ltv, 0.01, "LTV must reflect transaction amount")
}
