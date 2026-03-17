package command

import (
	"context"
	"fmt"

	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/google/uuid"
)

// TrackSessionCommand increments the user's session counter
type TrackSessionCommand struct {
	userRepo repository.UserRepository
}

func NewTrackSessionCommand(userRepo repository.UserRepository) *TrackSessionCommand {
	return &TrackSessionCommand{userRepo: userRepo}
}

// Execute increments session count and returns the new count
func (c *TrackSessionCommand) Execute(ctx context.Context, userID string) (int, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid user ID", domainErrors.ErrInvalidInput)
	}

	count, err := c.userRepo.IncrementSessionCount(ctx, userUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to increment session: %w", err)
	}

	return count, nil
}
