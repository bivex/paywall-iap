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

type User struct {
	ID            uuid.UUID
	PlatformUserID string
	DeviceID      string
	Platform      Platform
	AppVersion    string
	Email         string
	LTV           float64
	LTVUpdatedAt  time.Time
	CreatedAt     time.Time
	DeletedAt     *time.Time
}

// NewUser creates a new user entity
func NewUser(platformUserID, deviceID string, platform Platform, appVersion, email string) *User {
	return &User{
		ID:            uuid.New(),
		PlatformUserID: platformUserID,
		DeviceID:      deviceID,
		Platform:      platform,
		AppVersion:    appVersion,
		Email:         email,
		LTV:           0,
		CreatedAt:     time.Now(),
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
