package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/application/dto"
	appMiddleware "github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
)

type registerRepoStub struct {
	existsCalled bool
	createCalled bool
	userByEmail  *entity.User
	emailErr     error
	createErr    error
}

func (r *registerRepoStub) Create(context.Context, *entity.User) error {
	r.createCalled = true
	return r.createErr
}
func (r *registerRepoStub) GetByID(context.Context, uuid.UUID) (*entity.User, error) { return nil, nil }
func (r *registerRepoStub) GetByPlatformID(context.Context, string) (*entity.User, error) {
	return nil, nil
}
func (r *registerRepoStub) GetByEmail(context.Context, string) (*entity.User, error) {
	return r.userByEmail, r.emailErr
}
func (r *registerRepoStub) Update(context.Context, *entity.User) error  { return nil }
func (r *registerRepoStub) SoftDelete(context.Context, uuid.UUID) error { return nil }
func (r *registerRepoStub) ExistsByPlatformID(context.Context, string) (bool, error) {
	r.existsCalled = true
	return false, nil
}
func (r *registerRepoStub) UpdatePurchaseChannel(context.Context, uuid.UUID, string) error {
	return nil
}
func (r *registerRepoStub) UpdateEmail(context.Context, uuid.UUID, string) error { return nil }
func (r *registerRepoStub) IncrementSessionCount(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (r *registerRepoStub) UpdateHasViewedAds(context.Context, uuid.UUID, bool) error { return nil }

func TestRegisterCommand_RejectsNullBytesBeforeRepositoryAccess(t *testing.T) {
	repo := &registerRepoStub{}
	cmd := NewRegisterCommand(repo, appMiddleware.NewJWTMiddleware("test-secret", nil, time.Minute))

	_, err := cmd.Execute(context.Background(), &dto.RegisterRequest{
		PlatformUserID: "bad\x00user",
		DeviceID:       "device-1",
		Platform:       "ios",
		AppVersion:     "1.0.0",
		Email:          "user@example.com",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "null bytes")
	require.False(t, repo.existsCalled)
}

func TestRegisterCommand_RejectsDuplicateEmailBeforeCreate(t *testing.T) {
	repo := &registerRepoStub{userByEmail: entity.NewUser("existing-user", "device-1", entity.PlatformiOS, "1.0.0", "user@example.com")}
	cmd := NewRegisterCommand(repo, appMiddleware.NewJWTMiddleware("test-secret", nil, time.Minute))

	_, err := cmd.Execute(context.Background(), &dto.RegisterRequest{
		PlatformUserID: "new-user",
		DeviceID:       "device-2",
		Platform:       "ios",
		AppVersion:     "1.0.0",
		Email:          "user@example.com",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, domainErrors.ErrUserAlreadyExists)
	require.False(t, repo.createCalled)
}

func TestRegisterCommand_MapsDuplicateCreateErrorToUserAlreadyExists(t *testing.T) {
	repo := &registerRepoStub{emailErr: errors.New("user not found: " + domainErrors.ErrUserNotFound.Error()), createErr: errors.New("failed to create user: ERROR: duplicate key value violates unique constraint \"users_email_unique\" (SQLSTATE 23505)")}
	cmd := NewRegisterCommand(repo, appMiddleware.NewJWTMiddleware("test-secret", nil, time.Minute))

	_, err := cmd.Execute(context.Background(), &dto.RegisterRequest{
		PlatformUserID: "new-user",
		DeviceID:       "device-2",
		Platform:       "ios",
		AppVersion:     "1.0.0",
		Email:          "user@example.com",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, domainErrors.ErrUserAlreadyExists)
	require.True(t, repo.createCalled)
}
