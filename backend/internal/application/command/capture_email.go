package command

import (
	"context"
	"fmt"

	"github.com/bivex/paywall-iap/internal/application/dto"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/google/uuid"
)

// CaptureEmailCommand stores a user's email before purchase
type CaptureEmailCommand struct {
	userRepo repository.UserRepository
}

func NewCaptureEmailCommand(userRepo repository.UserRepository) *CaptureEmailCommand {
	return &CaptureEmailCommand{userRepo: userRepo}
}

func (c *CaptureEmailCommand) Execute(ctx context.Context, userID string, req *dto.CaptureEmailRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("%w: invalid user ID", domainErrors.ErrInvalidInput)
	}

	// Verify user exists
	if _, err := c.userRepo.GetByID(ctx, userUUID); err != nil {
		return fmt.Errorf("user not found: %w", domainErrors.ErrUserNotFound)
	}

	if err := c.userRepo.UpdateEmail(ctx, userUUID, req.Email); err != nil {
		return fmt.Errorf("failed to save email: %w", err)
	}

	return nil
}
