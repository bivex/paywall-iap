package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/password9090/paywall-iap/internal/domain/entity"
)

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	// Create creates a new transaction
	Create(ctx context.Context, transaction *entity.Transaction) error

	// GetByID retrieves a transaction by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)

	// GetByUserID retrieves transactions for a user with pagination
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Transaction, error)

	// GetBySubscriptionID retrieves transactions for a subscription
	GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.Transaction, error)

	// CheckDuplicateReceipt checks if a receipt has already been processed
	CheckDuplicateReceipt(ctx context.Context, receiptHash string) (bool, error)
}
