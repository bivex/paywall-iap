package command

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/application/dto"
	appMiddleware "github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/domain/entity"
)

type registerRepoStub struct{ existsCalled bool }

func (r *registerRepoStub) Create(context.Context, *entity.User) error               { return nil }
func (r *registerRepoStub) GetByID(context.Context, uuid.UUID) (*entity.User, error) { return nil, nil }
func (r *registerRepoStub) GetByPlatformID(context.Context, string) (*entity.User, error) {
	return nil, nil
}
func (r *registerRepoStub) GetByEmail(context.Context, string) (*entity.User, error) { return nil, nil }
func (r *registerRepoStub) Update(context.Context, *entity.User) error               { return nil }
func (r *registerRepoStub) SoftDelete(context.Context, uuid.UUID) error              { return nil }
func (r *registerRepoStub) ExistsByPlatformID(context.Context, string) (bool, error) {
	r.existsCalled = true
	return false, nil
}

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
