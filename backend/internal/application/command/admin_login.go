package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/bivex/paywall-iap/internal/application/dto"
	appMiddleware "github.com/bivex/paywall-iap/internal/application/middleware"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

// AdminLoginCommand handles email+password login for admin users.
type AdminLoginCommand struct {
	userRepo      repository.UserRepository
	credRepo      repository.AdminCredentialRepository
	jwtMiddleware *appMiddleware.JWTMiddleware
}

// NewAdminLoginCommand creates a new AdminLoginCommand.
func NewAdminLoginCommand(
	userRepo repository.UserRepository,
	credRepo repository.AdminCredentialRepository,
	jwtMiddleware *appMiddleware.JWTMiddleware,
) *AdminLoginCommand {
	return &AdminLoginCommand{
		userRepo:      userRepo,
		credRepo:      credRepo,
		jwtMiddleware: jwtMiddleware,
	}
}

// Execute validates credentials and returns JWT tokens.
func (c *AdminLoginCommand) Execute(ctx context.Context, req *dto.AdminLoginRequest) (*dto.AdminLoginResponse, error) {
	// 1. Find user by email
	user, err := c.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domainErrors.ErrUserNotFound) {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}

	// 2. Must be admin or superadmin
	if !user.IsAdmin() {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 3. Get stored password hash
	hash, err := c.credRepo.GetPasswordHash(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 4. Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 5. Generate tokens with role
	accessToken, _, err := c.jwtMiddleware.GenerateAccessTokenWithRole(user.ID.String(), user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := c.jwtMiddleware.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AdminLoginResponse{
		UserID:       user.ID.String(),
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(c.jwtMiddleware.AccessTTL().Seconds()),
	}, nil
}
