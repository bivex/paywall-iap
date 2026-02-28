package testutil

import (
	"context"
	"fmt"
)

// RunMigrationsOnContainer runs all schema migrations on the test database pool
// It re-uses the existing inline schema approach from testutil.go but via the container pool
func RunMigrationsOnContainer(ctx context.Context, tc *TestDBContainer) error {
	return RunMigrations(ctx, tc.Pool)
}

// SeedTestData inserts seed data into the test database
func SeedTestData(ctx context.Context, tc *TestDBContainer) error {
	_, err := tc.Pool.Exec(ctx, `
		INSERT INTO users (platform_user_id, device_id, platform, app_version, email)
		VALUES ('seed-platform-user', 'seed-device', 'ios', '1.0.0', 'seed@example.com')
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed test data: %w", err)
	}
	return nil
}
