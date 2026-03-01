package repository

import (
	"context"
	"fmt"

	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type adminCredentialRepositoryImpl struct {
	queries *generated.Queries
}

// NewAdminCredentialRepository creates a new admin credential repository.
func NewAdminCredentialRepository(queries *generated.Queries) domainRepo.AdminCredentialRepository {
	return &adminCredentialRepositoryImpl{queries: queries}
}

func (r *adminCredentialRepositoryImpl) SetPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	_, err := r.queries.UpsertAdminCredential(ctx, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to set admin password: %w", err)
	}
	return nil
}

func (r *adminCredentialRepositoryImpl) GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error) {
	cred, err := r.queries.GetAdminCredential(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("no credentials found for user")
		}
		return "", fmt.Errorf("failed to get admin credential: %w", err)
	}
	return cred.PasswordHash, nil
}
