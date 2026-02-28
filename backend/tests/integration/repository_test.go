//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	infrarepo "github.com/bivex/paywall-iap/internal/infrastructure/persistence/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/tests/testutil"
)

func TestUserRepositoryIntegration(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	// Create real repository using sqlc queries
	queries := generated.New(dbContainer.Pool)
	userRepo := infrarepo.NewUserRepository(queries)

	t.Run("Create and GetUserByID", func(t *testing.T) {
		user := entity.NewUser("platform-user-123-"+uuid.New().String(), "device-123", entity.PlatformiOS, "1.0.0", "test_"+uuid.New().String()[:8]+"@example.com")

		// Create
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Get by Platform ID (since Create doesn't set the UUID from DB)
		retrieved, err := userRepo.GetByPlatformID(ctx, user.PlatformUserID)
		require.NoError(t, err)
		assert.Equal(t, user.PlatformUserID, retrieved.PlatformUserID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("GetByPlatformID", func(t *testing.T) {
		platformUserID := "platform-user-" + uuid.New().String()
		user := entity.NewUser(platformUserID, "device-456", entity.PlatformAndroid, "1.0.0", "test2_"+uuid.New().String()[:8]+"@example.com")

		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := userRepo.GetByPlatformID(ctx, platformUserID)
		require.NoError(t, err)
		assert.Equal(t, platformUserID, retrieved.PlatformUserID)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		email := "test_" + uuid.New().String() + "@example.com"
		user := entity.NewUser("platform-user-"+uuid.New().String(), "device-789", entity.PlatformiOS, "1.0.0", email)

		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := userRepo.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, email, retrieved.Email)
	})

	t.Run("SoftDeleteUser", func(t *testing.T) {
		user := entity.NewUser("platform-user-del-"+uuid.New().String(), "device-delete", entity.PlatformiOS, "1.0.0", "delete_"+uuid.New().String()[:8]+"@example.com")
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Get to find real ID
		retrieved, err := userRepo.GetByPlatformID(ctx, user.PlatformUserID)
		require.NoError(t, err)

		// Soft delete
		err = userRepo.SoftDelete(ctx, retrieved.ID)
		require.NoError(t, err)

		// Verify user is not found by normal queries
		_, err = userRepo.GetByID(ctx, retrieved.ID)
		assert.Error(t, err) // Should return not found
	})

	t.Run("ExistsByPlatformID", func(t *testing.T) {
		platformUserID := "platform-exists-" + uuid.New().String()
		user := entity.NewUser(platformUserID, "device-exists", entity.PlatformiOS, "1.0.0", "exists_"+uuid.New().String()[:8]+"@example.com")
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		exists, err := userRepo.ExistsByPlatformID(ctx, platformUserID)
		require.NoError(t, err)
		assert.True(t, exists)

		notExists, err := userRepo.ExistsByPlatformID(ctx, "non-existent-user")
		require.NoError(t, err)
		assert.False(t, notExists)
	})
}

func TestSubscriptionRepositoryIntegration(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	// Create repositories
	queries := generated.New(dbContainer.Pool)
	userRepo := infrarepo.NewUserRepository(queries)
	subRepo := infrarepo.NewSubscriptionRepository(queries)

	// Create test user
	user := entity.NewUser("platform-user-sub-"+uuid.New().String(), "device-sub", entity.PlatformiOS, "1.0.0", "sub_"+uuid.New().String()[:8]+"@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Get the actual user with DB-generated ID
	dbUser, err := userRepo.GetByPlatformID(ctx, user.PlatformUserID)
	require.NoError(t, err)

	t.Run("Create and GetSubscriptionByID", func(t *testing.T) {
		sub := entity.NewSubscription(dbUser.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))

		err := subRepo.Create(ctx, sub)
		require.NoError(t, err)

		// Get active sub to verify creation
		retrieved, err := subRepo.GetActiveByUserID(ctx, dbUser.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusActive, retrieved.Status)
	})

	t.Run("GetAccessCheck", func(t *testing.T) {
		hasAccess, err := subRepo.CanAccess(ctx, dbUser.ID)
		require.NoError(t, err)
		assert.True(t, hasAccess)
	})

	t.Run("UpdateSubscriptionStatus", func(t *testing.T) {
		sub := entity.NewSubscription(dbUser.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
		err := subRepo.Create(ctx, sub)
		require.NoError(t, err)

		// Get active sub to find real ID
		retrieved, err := subRepo.GetActiveByUserID(ctx, dbUser.ID)
		require.NoError(t, err)

		err = subRepo.UpdateStatus(ctx, retrieved.ID, entity.StatusCancelled)
		require.NoError(t, err)

		updated, err := subRepo.GetByID(ctx, retrieved.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusCancelled, updated.Status)
	})

	t.Run("CancelSubscription", func(t *testing.T) {
		// Create fresh user for this test
		cancelUser := entity.NewUser("platform-cancel-"+uuid.New().String(), "device-cancel", entity.PlatformiOS, "1.0.0", "cancel_"+uuid.New().String()[:8]+"@example.com")
		err := userRepo.Create(ctx, cancelUser)
		require.NoError(t, err)

		dbCancelUser, err := userRepo.GetByPlatformID(ctx, cancelUser.PlatformUserID)
		require.NoError(t, err)

		sub := entity.NewSubscription(dbCancelUser.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
		err = subRepo.Create(ctx, sub)
		require.NoError(t, err)

		retrieved, err := subRepo.GetActiveByUserID(ctx, dbCancelUser.ID)
		require.NoError(t, err)

		err = subRepo.Cancel(ctx, retrieved.ID)
		require.NoError(t, err)

		cancelled, err := subRepo.GetByID(ctx, retrieved.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusCancelled, cancelled.Status)
		assert.False(t, cancelled.AutoRenew)
	})
}

func TestTransactionRepositoryIntegration(t *testing.T) {
	ctx := context.Background()

	// Setup test database
	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	// Run migrations
	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	// Create repositories
	queries := generated.New(dbContainer.Pool)
	userRepo := infrarepo.NewUserRepository(queries)
	subRepo := infrarepo.NewSubscriptionRepository(queries)
	txRepo := infrarepo.NewTransactionRepository(queries)

	// Create test user and subscription
	user := entity.NewUser("platform-user-tx-"+uuid.New().String(), "device-tx", entity.PlatformiOS, "1.0.0", "tx_"+uuid.New().String()[:8]+"@example.com")
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	dbUser, err := userRepo.GetByPlatformID(ctx, user.PlatformUserID)
	require.NoError(t, err)

	sub := entity.NewSubscription(dbUser.ID, entity.SourceIAP, "ios", "com.app.premium", entity.PlanMonthly, time.Now().Add(30*24*time.Hour))
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	dbSub, err := subRepo.GetActiveByUserID(ctx, dbUser.ID)
	require.NoError(t, err)

	t.Run("Create and GetTransactionByID", func(t *testing.T) {
		tx := entity.NewTransaction(dbUser.ID, dbSub.ID, 9.99, "USD")
		tx.Status = entity.TransactionStatusSuccess
		tx.ReceiptHash = "sha256_test_hash_" + uuid.New().String()
		tx.ProviderTxID = "provider_tx_123"

		err := txRepo.Create(ctx, tx)
		require.NoError(t, err)

		// Get transactions by user ID
		txs, err := txRepo.GetByUserID(ctx, dbUser.ID, 10, 0)
		require.NoError(t, err)
		assert.True(t, len(txs) >= 1)
	})

	t.Run("CheckDuplicateReceipt", func(t *testing.T) {
		receiptHash := "sha256_duplicate_test_" + uuid.New().String()

		// First transaction
		tx1 := entity.NewTransaction(dbUser.ID, dbSub.ID, 9.99, "USD")
		tx1.ReceiptHash = receiptHash
		err := txRepo.Create(ctx, tx1)
		require.NoError(t, err)

		// Check duplicate
		isDuplicate, err := txRepo.CheckDuplicateReceipt(ctx, receiptHash)
		require.NoError(t, err)
		assert.True(t, isDuplicate)

		// Check non-existent receipt
		isDuplicate, err = txRepo.CheckDuplicateReceipt(ctx, "sha256_non_existent_"+uuid.New().String())
		require.NoError(t, err)
		assert.False(t, isDuplicate)
	})
}
