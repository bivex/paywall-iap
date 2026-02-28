package repository

import (
	"context"
	"fmt"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type transactionRepositoryImpl struct {
	queries *generated.Queries
}

// NewTransactionRepository creates a new transaction repository implementation
func NewTransactionRepository(queries *generated.Queries) repository.TransactionRepository {
	return &transactionRepositoryImpl{queries: queries}
}

func (r *transactionRepositoryImpl) Create(ctx context.Context, txn *entity.Transaction) error {
	params := generated.CreateTransactionParams{
		UserID:         txn.UserID,
		SubscriptionID: txn.SubscriptionID,
		Amount:         txn.Amount,
		Currency:       txn.Currency,
		Status:         string(txn.Status),
		ReceiptHash:    &txn.ReceiptHash,
		ProviderTxID:   &txn.ProviderTxID,
	}

	_, err := r.queries.CreateTransaction(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *transactionRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	row, err := r.queries.GetTransactionByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %w", domainErrors.ErrTransactionNotFound)
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *transactionRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Transaction, error) {
	params := generated.GetTransactionsByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	rows, err := r.queries.GetTransactionsByUserID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	transactions := make([]*entity.Transaction, len(rows))
	for i, row := range rows {
		transactions[i] = r.mapToEntity(row)
	}

	return transactions, nil
}

func (r *transactionRepositoryImpl) GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.Transaction, error) {
	// For now, return empty array
	// In full implementation, add query to GetTransactionsBySubscriptionID
	return []*entity.Transaction{}, nil
}

func (r *transactionRepositoryImpl) CheckDuplicateReceipt(ctx context.Context, receiptHash string) (bool, error) {
	_, err := r.queries.CheckDuplicateReceipt(ctx, &receiptHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check duplicate receipt: %w", err)
	}

	return true, nil
}

func (r *transactionRepositoryImpl) mapToEntity(row generated.Transaction) *entity.Transaction {
	var receiptHash, providerTxID string
	if row.ReceiptHash != nil {
		receiptHash = *row.ReceiptHash
	}
	if row.ProviderTxID != nil {
		providerTxID = *row.ProviderTxID
	}

	return &entity.Transaction{
		ID:             row.ID,
		UserID:         row.UserID,
		SubscriptionID: row.SubscriptionID,
		Amount:         row.Amount,
		Currency:       row.Currency,
		Status:         entity.TransactionStatus(row.Status),
		ReceiptHash:    receiptHash,
		ProviderTxID:   providerTxID,
		CreatedAt:      row.CreatedAt,
	}
}
