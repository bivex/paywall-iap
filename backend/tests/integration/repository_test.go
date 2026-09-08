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

	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	queries := generated.New(dbContainer.Pool)
	userRepo := infrarepo.NewUserRepository(queries)

	t.Run("Create and GetUserByID", func(t *testing.T) {
		user := entity.NewUser(entity.NewUserParams{
			PlatformUserID: "platform-user-123-" + uuid.New().String(),
			DeviceID:       "device-123",
			Platform:       entity.PlatformiOS,
			AppVersion:     "1.0.0",
			Email:          "test_" + uuid.New().String()[:8] + "@example.com",
			AppID:          uuid.Nil,
		})

		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := userRepo.GetByPlatformID(ctx, user.PlatformUserID)
		require.NoError(t, err)
		assert.Equal(t, user.PlatformUserID, retrieved.PlatformUserID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("GetByPlatformID", func(t *testing.T) {
		platformUserID := "platform-user-" + uuid.New().String()
		user := entity.NewUser(entity.NewUserParams{
			PlatformUserID: platformUserID,
			DeviceID:       "device-456",
			Platform:       entity.PlatformAndroid,
			AppVersion:     "1.0.0",
			Email:          "test2_" + uuid.New().String()[:8] + "@example.com",
			AppID:          uuid.Nil,
		})

		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := userRepo.GetByPlatformID(ctx, platformUserID)
		require.NoError(t, err)
		assert.Equal(t, platformUserID, retrieved.PlatformUserID)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		email := "test_" + uuid.New().String() + "@example.com"
		user := entity.NewUser(entity.NewUserParams{
			PlatformUserID: "platform-user-" + uuid.New().String(),
			DeviceID:       "device-789",
			Platform:       entity.PlatformiOS,
			AppVersion:     "1.0.0",
			Email:          email,
			AppID:          uuid.Nil,
		})

		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := userRepo.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, email, retrieved.Email)
	})

	t.Run("SoftDeleteUser", func(t *testing.T) {
		user := entity.NewUser(entity.NewUserParams{
			PlatformUserID: "platform-user-del-" + uuid.New().String(),
			DeviceID:       "device-delete",
			Platform:       entity.PlatformiOS,
			AppVersion:     "1.0.0",
			Email:          "delete_" + uuid.New().String()[:8] + "@example.com",
			AppID:          uuid.Nil,
		})
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := userRepo.GetByPlatformID(ctx, user.PlatformUserID)
		require.NoError(t, err)

		err = userRepo.SoftDelete(ctx, retrieved.ID)
		require.NoError(t, err)

		_, err = userRepo.GetByID(ctx, retrieved.ID)
		assert.Error(t, err)
	})

	t.Run("ExistsByPlatformID", func(t *testing.T) {
		platformUserID := "platform-exists-" + uuid.New().String()
		user := entity.NewUser(entity.NewUserParams{
			PlatformUserID: platformUserID,
			DeviceID:       "device-exists",
			Platform:       entity.PlatformiOS,
			AppVersion:     "1.0.0",
			Email:          "exists_" + uuid.New().String()[:8] + "@example.com",
			AppID:          uuid.Nil,
		})
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

	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	queries := generated.New(dbContainer.Pool)
	userRepo := infrarepo.NewUserRepository(queries)
	subRepo := infrarepo.NewSubscriptionRepository(queries)

	user := entity.NewUser(entity.NewUserParams{
		PlatformUserID: "platform-user-sub-" + uuid.New().String(),
		DeviceID:       "device-sub",
		Platform:       entity.PlatformiOS,
		AppVersion:     "1.0.0",
		Email:          "sub_" + uuid.New().String()[:8] + "@example.com",
		AppID:          uuid.Nil,
	})
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	dbUser, err := userRepo.GetByPlatformID(ctx, user.PlatformUserID)
	require.NoError(t, err)

	t.Run("Create and GetSubscriptionByID", func(t *testing.T) {
		sub := entity.NewSubscription(entity.NewSubscriptionParams{
			UserID:    dbUser.ID,
			Source:    entity.SourceIAP,
			Platform:  "ios",
			ProductID: "com.app.premium",
			PlanType:  entity.PlanMonthly,
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		})

		err := subRepo.Create(ctx, sub)
		require.NoError(t, err)

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
		sub := entity.NewSubscription(entity.NewSubscriptionParams{
			UserID:    dbUser.ID,
			Source:    entity.SourceIAP,
			Platform:  "ios",
			ProductID: "com.app.premium",
			PlanType:  entity.PlanMonthly,
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		})
		err := subRepo.Create(ctx, sub)
		require.NoError(t, err)

		retrieved, err := subRepo.GetActiveByUserID(ctx, dbUser.ID)
		require.NoError(t, err)

		err = subRepo.UpdateStatus(ctx, retrieved.ID, entity.StatusCancelled)
		require.NoError(t, err)

		updated, err := subRepo.GetByID(ctx, retrieved.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusCancelled, updated.Status)
	})

	t.Run("CancelSubscription", func(t *testing.T) {
		cancelUser := entity.NewUser(entity.NewUserParams{
			PlatformUserID: "platform-cancel-" + uuid.New().String(),
			DeviceID:       "device-cancel",
			Platform:       entity.PlatformiOS,
			AppVersion:     "1.0.0",
			Email:          "cancel_" + uuid.New().String()[:8] + "@example.com",
			AppID:          uuid.Nil,
		})
		err := userRepo.Create(ctx, cancelUser)
		require.NoError(t, err)

		dbCancelUser, err := userRepo.GetByPlatformID(ctx, cancelUser.PlatformUserID)
		require.NoError(t, err)

		sub := entity.NewSubscription(entity.NewSubscriptionParams{
			UserID:    dbCancelUser.ID,
			Source:    entity.SourceIAP,
			Platform:  "ios",
			ProductID: "com.app.premium",
			PlanType:  entity.PlanMonthly,
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		})
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

	dbContainer, err := testutil.SetupTestDBContainer(ctx, t)
	require.NoError(t, err)
	defer dbContainer.Teardown(ctx, t)

	err = testutil.RunMigrations(ctx, dbContainer.Pool)
	require.NoError(t, err)

	queries := generated.New(dbContainer.Pool)
	userRepo := infrarepo.NewUserRepository(queries)
	subRepo := infrarepo.NewSubscriptionRepository(queries)
	txRepo := infrarepo.NewTransactionRepository(queries)

	user := entity.NewUser(entity.NewUserParams{
		PlatformUserID: "platform-user-tx-" + uuid.New().String(),
		DeviceID:       "device-tx",
		Platform:       entity.PlatformiOS,
		AppVersion:     "1.0.0",
		Email:          "tx_" + uuid.New().String()[:8] + "@example.com",
		AppID:          uuid.Nil,
	})
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	dbUser, err := userRepo.GetByPlatformID(ctx, user.PlatformUserID)
	require.NoError(t, err)

	sub := entity.NewSubscription(entity.NewSubscriptionParams{
		UserID:    dbUser.ID,
		Source:    entity.SourceIAP,
		Platform:  "ios",
		ProductID: "com.app.premium",
		PlanType:  entity.PlanMonthly,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})
	err = subRepo.Create(ctx, sub)
	require.NoError(t, err)

	dbSub, err := subRepo.GetActiveByUserID(ctx, dbUser.ID)
	require.NoError(t, err)

	t.Run("Create and GetTransactionByID", func(t *testing.T) {
		tx := entity.NewTransaction(entity.NewTransactionParams{
			AppID:          dbUser.AppID,
			UserID:         dbUser.ID,
			SubscriptionID: dbSub.ID,
			Amount:         9.99,
			Currency:       "USD",
		})
		tx.Status = entity.TransactionStatusSuccess
		tx.ReceiptHash = "sha256_test_hash_" + uuid.New().String()
		tx.ProviderTxID = "provider_tx_123"

		err := txRepo.Create(ctx, tx)
		require.NoError(t, err)

		txs, err := txRepo.GetByUserID(ctx, dbUser.ID, 10, 0)
		require.NoError(t, err)
		assert.True(t, len(txs) >= 1)
	})

	t.Run("CheckDuplicateReceipt", func(t *testing.T) {
		receiptHash := "sha256_duplicate_test_" + uuid.New().String()

		tx1 := entity.NewTransaction(entity.NewTransactionParams{
			AppID:          dbUser.AppID,
			UserID:         dbUser.ID,
			SubscriptionID: dbSub.ID,
			Amount:         9.99,
			Currency:       "USD",
		})
		tx1.ReceiptHash = receiptHash
		err := txRepo.Create(ctx, tx1)
		require.NoError(t, err)

		isDuplicate, err := txRepo.CheckDuplicateReceipt(ctx, receiptHash)
		require.NoError(t, err)
		assert.True(t, isDuplicate)

		isDuplicate, err = txRepo.CheckDuplicateReceipt(ctx, "sha256_non_existent_"+uuid.New().String())
		require.NoError(t, err)
		assert.False(t, isDuplicate)
	})
}
