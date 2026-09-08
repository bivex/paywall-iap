package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/bivex/paywall-iap/internal/application/dto"
	appMiddleware "github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/google/uuid"
)

// RegisterCommand handles user registration
type RegisterCommand struct {
	userRepo      repository.UserRepository
	jwtMiddleware *appMiddleware.JWTMiddleware
}

// NewRegisterCommand creates a new register command
func NewRegisterCommand(userRepo repository.UserRepository, jwtMiddleware *appMiddleware.JWTMiddleware) *RegisterCommand {
	return &RegisterCommand{
		userRepo:      userRepo,
		jwtMiddleware: jwtMiddleware,
	}
}

func validateRegisterRequest(req *dto.RegisterRequest) (uuid.UUID, error) {
	if req.Platform != "ios" && req.Platform != "android" {
		return uuid.Nil, fmt.Errorf("%w: invalid platform", domainErrors.ErrInvalidPlatform)
	}
	if containsNullByte(req.PlatformUserID) || containsNullByte(req.DeviceID) || containsNullByte(req.AppVersion) || containsNullByte(req.Email) {
		return uuid.Nil, fmt.Errorf("invalid request: text fields must not contain null bytes")
	}

	if req.AppID != "" {
		parsed, err := uuid.Parse(req.AppID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid app_id: %w", err)
		}
		return parsed, nil
	}
	return uuid.Nil, nil
}

func (c *RegisterCommand) checkUserAvailability(ctx context.Context, req *dto.RegisterRequest, appID uuid.UUID) error {
	var exists bool
	var err error
	if appID != uuid.Nil {
		exists, err = c.userRepo.ExistsByPlatformIDAndApp(ctx, req.PlatformUserID, appID)
	} else {
		exists, err = c.userRepo.ExistsByPlatformID(ctx, req.PlatformUserID)
	}
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return fmt.Errorf("%w: user already exists", domainErrors.ErrUserAlreadyExists)
	}

	if strings.TrimSpace(req.Email) != "" {
		userByEmail, err := c.userRepo.GetByEmail(ctx, req.Email)
		if err == nil && userByEmail != nil {
			return fmt.Errorf("%w: user already exists", domainErrors.ErrUserAlreadyExists)
		}
		if err != nil && !strings.Contains(err.Error(), domainErrors.ErrUserNotFound.Error()) {
			return fmt.Errorf("failed to check user email: %w", err)
		}
	}
	return nil
}

// Execute executes the register command
func (c *RegisterCommand) Execute(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	appID, err := validateRegisterRequest(req)
	if err != nil {
		return nil, err
	}

	if err := c.checkUserAvailability(ctx, req, appID); err != nil {
		return nil, err
	}

	// Create user entity
	user := entity.NewUser(
		req.PlatformUserID,
		req.DeviceID,
		entity.Platform(req.Platform),
		req.AppVersion,
		req.Email,
		appID,
	)

	// Save user
	if err := c.userRepo.Create(ctx, user); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key value") {
			return nil, fmt.Errorf("%w: user already exists", domainErrors.ErrUserAlreadyExists)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT tokens (embed app_id when present)
	accessToken, refreshToken, err := c.jwtMiddleware.GenerateTokenPair(user.ID.String(), req.AppID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &dto.RegisterResponse{
		UserID:       user.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
	}, nil
}

func containsNullByte(value string) bool {
	return strings.ContainsRune(value, '\x00')
}
