package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

func TestFeatureFlagService(t *testing.T) {
	ctx := context.Background()
	ffService := service.NewFeatureFlagService()

	t.Run("CreateFlag and GetFlag", func(t *testing.T) {
		flag := ffService.CreateFlag("test_flag", "Test Flag", true, 50, []string{})

		assert.Equal(t, "test_flag", flag.ID)
		assert.Equal(t, "Test Flag", flag.Name)
		assert.True(t, flag.Enabled)
		assert.Equal(t, 50, flag.RolloutPercent)
	})

	t.Run("IsFeatureEnabled with 100% rollout", func(t *testing.T) {
		ffService.CreateFlag("full_rollout", "Full Rollout", true, 100, []string{})

		enabled, err := ffService.IsFeatureEnabled(ctx, "full_rollout", "user_123")
		require.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("IsFeatureEnabled with 0% rollout", func(t *testing.T) {
		ffService.CreateFlag("zero_rollout", "Zero Rollout", true, 0, []string{})

		enabled, err := ffService.IsFeatureEnabled(ctx, "zero_rollout", "user_123")
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("IsFeatureEnabled with specific user IDs", func(t *testing.T) {
		ffService.CreateFlag("beta_flag", "Beta Flag", true, 0, []string{"beta_user_1", "beta_user_2"})

		// Beta user should have access
		enabled, err := ffService.IsFeatureEnabled(ctx, "beta_flag", "beta_user_1")
		require.NoError(t, err)
		assert.True(t, enabled)

		// Non-beta user should not have access
		enabled, err = ffService.IsFeatureEnabled(ctx, "beta_flag", "regular_user")
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("IsFeatureEnabled with disabled flag", func(t *testing.T) {
		ffService.CreateFlag("disabled_flag", "Disabled Flag", false, 100, []string{})

		enabled, err := ffService.IsFeatureEnabled(ctx, "disabled_flag", "user_123")
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("IsFeatureEnabled returns error for non-existent flag", func(t *testing.T) {
		enabled, err := ffService.IsFeatureEnabled(ctx, "non_existent", "user_123")
		assert.Error(t, err)
		assert.False(t, enabled)
	})

	t.Run("UpdateFlag updates existing flag", func(t *testing.T) {
		ffService.CreateFlag("update_flag", "Update Flag", true, 50, []string{})

		enabled := false
		rollout := 75
		err := ffService.UpdateFlag("update_flag", &enabled, &rollout, []string{"new_user"})
		require.NoError(t, err)

		flag, _ := ffService.GetFlag("update_flag")
		assert.False(t, flag.Enabled)
		assert.Equal(t, 75, flag.RolloutPercent)
		assert.Contains(t, flag.UserIDs, "new_user")
	})

	t.Run("DeleteFlag removes flag", func(t *testing.T) {
		ffService.CreateFlag("delete_flag", "Delete Flag", true, 50, []string{})

		err := ffService.DeleteFlag("delete_flag")
		require.NoError(t, err)

		_, err = ffService.GetFlag("delete_flag")
		assert.Error(t, err)
	})

	t.Run("EvaluatePaywallTest returns variants", func(t *testing.T) {
		// Without flag, should return control
		variant, err := ffService.EvaluatePaywallTest(ctx, "user_123")
		require.NoError(t, err)
		assert.Equal(t, "control", variant)

		// With flag enabled, should return variant_b
		ffService.CreateFlag("paywall_variant_test", "Paywall Test", true, 100, []string{})
		variant, err = ffService.EvaluatePaywallTest(ctx, "user_123")
		require.NoError(t, err)
		assert.Equal(t, "variant_b", variant)
	})
}
