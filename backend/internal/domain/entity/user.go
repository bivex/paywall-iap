package entity

import (
	"time"

	"github.com/google/uuid"
)

type Platform string

const (
	PlatformiOS     Platform = "ios"
	PlatformAndroid Platform = "android"
)

const (
	RoleUser       = "user"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "superadmin"
)

const (
	PurchaseChannelIAP    = "iap"
	PurchaseChannelStripe = "stripe"
	PurchaseChannelWeb    = "web"
)

type User struct {
	ID              uuid.UUID
	PlatformUserID  string
	DeviceID        string
	Platform        Platform
	AppVersion      string
	Email           string
	LTV             float64
	LTVUpdatedAt    time.Time
	Role            string
	CreatedAt       time.Time
	DeletedAt       *time.Time
	PurchaseChannel *string  // "iap", "stripe", "web", or nil
	SessionCount    int
	HasViewedAds    bool
}

// NewUser creates a new user entity
func NewUser(platformUserID, deviceID string, platform Platform, appVersion, email string) *User {
	return &User{
		ID:             uuid.New(),
		PlatformUserID: platformUserID,
		DeviceID:       deviceID,
		Platform:       platform,
		AppVersion:     appVersion,
		Email:          email,
		LTV:            0,
		Role:           RoleUser,
		CreatedAt:      time.Now(),
	}
}

// IsDeleted returns true if the user has been soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// HasEmail returns true if the user has an email address
func (u *User) HasEmail() bool {
	return u.Email != ""
}

// IsAdmin returns true if the user has admin or superadmin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// HasPurchasedViaIAP returns true if the user's first purchase was via IAP
func (u *User) HasPurchasedViaIAP() bool {
	return u.PurchaseChannel != nil && *u.PurchaseChannel == PurchaseChannelIAP
}

// ShouldShowD2CButton returns true if the D2C button should be shown.
// Per Google/Apple policy, don't show D2C steering to IAP users.
func (u *User) ShouldShowD2CButton() bool {
	return u.PurchaseChannel == nil || *u.PurchaseChannel != PurchaseChannelIAP
}
