package entity

import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatus string

const (
	TransactionStatusSuccess  TransactionStatus = "success"
	TransactionStatusFailed   TransactionStatus = "failed"
	TransactionStatusRefunded TransactionStatus = "refunded"
)

type Transaction struct {
	ID             uuid.UUID
	AppID          uuid.UUID
	UserID         uuid.UUID
	SubscriptionID uuid.UUID
	Amount         float64
	Currency       string
	Status         TransactionStatus
	ReceiptHash    string
	ProviderTxID   string
	CreatedAt      time.Time
}

// NewTransactionParams contains parameters for creating a new transaction entity.
type NewTransactionParams struct {
	AppID          uuid.UUID
	UserID         uuid.UUID
	SubscriptionID uuid.UUID
	Amount         float64
	Currency       string
}

// NewTransaction creates a new transaction entity
func NewTransaction(p NewTransactionParams) *Transaction {
	return &Transaction{
		ID:             uuid.New(),
		AppID:          p.AppID,
		UserID:         p.UserID,
		SubscriptionID: p.SubscriptionID,
		Amount:         p.Amount,
		Currency:       p.Currency,
		Status:         TransactionStatusSuccess,
		CreatedAt:      time.Now(),
	}
}

// IsSuccessful returns true if the transaction was successful
func (t *Transaction) IsSuccessful() bool {
	return t.Status == TransactionStatusSuccess
}

// IsFailed returns true if the transaction failed
func (t *Transaction) IsFailed() bool {
	return t.Status == TransactionStatusFailed
}
