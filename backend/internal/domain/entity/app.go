package entity

import (
	"time"

	"github.com/google/uuid"
)

// App represents a registered Mothsalt application.
type App struct {
	ID          uuid.UUID
	Name        string // reverse-dns, e.g. "com.mothsalt.game1"
	DisplayName string
	Platform    string // "ios", "android", "both"
	BundleID    string // App Store bundle ID / Google Play package name
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
