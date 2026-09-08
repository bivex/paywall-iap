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
	AppID           uuid.UUID
}

// NewUserParams contains parameters for creating a new user entity.
type NewUserParams struct {
	PlatformUserID string
	DeviceID       string
	Platform       Platform
	AppVersion     string
	Email          string
	AppID          uuid.UUID
}

// NewUser creates a new user entity
func NewUser(p NewUserParams) *User {
	return &User{
		ID:             uuid.New(),
		PlatformUserID: p.PlatformUserID,
		DeviceID:       p.DeviceID,
		Platform:       p.Platform,
		AppVersion:     p.AppVersion,
		Email:          p.Email,
		LTV:            0,
		Role:           RoleUser,
		CreatedAt:      time.Now(),
		AppID:          p.AppID,
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
