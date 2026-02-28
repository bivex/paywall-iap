//go:build e2e

package e2e

import (
	"os"
	"testing"
)

// TestMain is the entry point for E2E tests
func TestMain(m *testing.M) {
	// If any global setup is required, it goes here

	// Run all E2E tests
	code := m.Run()

	// If any global teardown is required, it goes here

	os.Exit(code)
}
