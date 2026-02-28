package command

import (
	"context"
	"fmt"

	"github.com/bivex/paywall-iap/internal/application/dto"
	appMiddleware "github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/repository"
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

// Execute executes the register command
func (c *RegisterCommand) Execute(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Validate platform
	if req.Platform != "ios" && req.Platform != "android" {
		return nil, fmt.Errorf("%w: invalid platform", domainErrors.ErrInvalidPlatform)
	}

	// Check if user already exists
	exists, err := c.userRepo.ExistsByPlatformID(ctx, req.PlatformUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: user already exists", domainErrors.ErrUserAlreadyExists)
	}

	// Create user entity
	user := entity.NewUser(
		req.PlatformUserID,
		req.DeviceID,
		entity.Platform(req.Platform),
		req.AppVersion,
		req.Email,
	)

	// Save user
	if err := c.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT tokens
	accessToken, _, err := c.jwtMiddleware.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := c.jwtMiddleware.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.RegisterResponse{
		UserID:       user.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
	}, nil
}
