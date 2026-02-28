package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type userRepositoryImpl struct {
	queries *generated.Queries
}

// NewUserRepository creates a new user repository implementation
func NewUserRepository(queries *generated.Queries) repository.UserRepository {
	return &userRepositoryImpl{queries: queries}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	params := generated.CreateUserParams{
		PlatformUserID: user.PlatformUserID,
		DeviceID:       &user.DeviceID,
		Platform:       string(user.Platform),
		AppVersion:     user.AppVersion,
		Email:          user.Email,
	}

	row, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = row.ID
	return nil
}

func (r *userRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", domainErrors.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *userRepositoryImpl) GetByPlatformID(ctx context.Context, platformUserID string) (*entity.User, error) {
	row, err := r.queries.GetUserByPlatformID(ctx, platformUserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", domainErrors.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", domainErrors.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	params := generated.UpdateUserLTVParams{
		ID:  user.ID,
		Ltv: user.LTV,
	}

	_, err := r.queries.UpdateUserLTV(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) ExistsByPlatformID(ctx context.Context, platformUserID string) (bool, error) {
	_, err := r.queries.GetUserByPlatformID(ctx, platformUserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return true, nil
}

func (r *userRepositoryImpl) mapToEntity(row generated.User) *entity.User {
	var deviceID string
	if row.DeviceID != nil {
		deviceID = *row.DeviceID
	}

	var ltvUpdatedAt time.Time
	if row.LtvUpdatedAt != nil {
		ltvUpdatedAt = *row.LtvUpdatedAt
	}

	return &entity.User{
		ID:             row.ID,
		PlatformUserID: row.PlatformUserID,
		DeviceID:       deviceID,
		Platform:       entity.Platform(row.Platform),
		AppVersion:     row.AppVersion,
		Email:          row.Email,
		LTV:            row.Ltv,
		LTVUpdatedAt:   ltvUpdatedAt,
		CreatedAt:      row.CreatedAt,
		DeletedAt:      row.DeletedAt,
	}
}
