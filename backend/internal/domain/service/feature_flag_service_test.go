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
		flag := ffService.CreateFlag(service.CreateFlagParams{ID: "test_flag", Name: "Test Flag", Enabled: true, RolloutPercent: 50, UserIDs: []string{}})

		assert.Equal(t, "test_flag", flag.ID)
		assert.Equal(t, "Test Flag", flag.Name)
		assert.True(t, flag.Enabled)
		assert.Equal(t, 50, flag.RolloutPercent)
	})

	t.Run("IsFeatureEnabled with 100% rollout", func(t *testing.T) {
		ffService.CreateFlag(service.CreateFlagParams{ID: "full_rollout", Name: "Full Rollout", Enabled: true, RolloutPercent: 100, UserIDs: []string{}})

		enabled, err := ffService.IsFeatureEnabled(ctx, "full_rollout", "user_123")
		require.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("IsFeatureEnabled with 0% rollout", func(t *testing.T) {
		ffService.CreateFlag(service.CreateFlagParams{ID: "zero_rollout", Name: "Zero Rollout", Enabled: true, RolloutPercent: 0, UserIDs: []string{}})

		enabled, err := ffService.IsFeatureEnabled(ctx, "zero_rollout", "user_123")
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("IsFeatureEnabled with specific user IDs", func(t *testing.T) {
		ffService.CreateFlag(service.CreateFlagParams{ID: "beta_flag", Name: "Beta Flag", Enabled: true, RolloutPercent: 0, UserIDs: []string{"beta_user_1", "beta_user_2"}})

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
		ffService.CreateFlag(service.CreateFlagParams{ID: "disabled_flag", Name: "Disabled Flag", Enabled: false, RolloutPercent: 100, UserIDs: []string{}})

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
		ffService.CreateFlag(service.CreateFlagParams{ID: "update_flag", Name: "Update Flag", Enabled: true, RolloutPercent: 50, UserIDs: []string{}})

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
		ffService.CreateFlag(service.CreateFlagParams{ID: "delete_flag", Name: "Delete Flag", Enabled: true, RolloutPercent: 50, UserIDs: []string{}})

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
		ffService.CreateFlag(service.CreateFlagParams{ID: "paywall_variant_test", Name: "Paywall Test", Enabled: true, RolloutPercent: 100, UserIDs: []string{}})
		variant, err = ffService.EvaluatePaywallTest(ctx, "user_123")
		require.NoError(t, err)
		assert.Equal(t, "variant_b", variant)
	})
}
