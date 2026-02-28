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
	UserID         uuid.UUID
	SubscriptionID uuid.UUID
	Amount         float64
	Currency       string
	Status         TransactionStatus
	ReceiptHash    string
	ProviderTxID   string
	CreatedAt      time.Time
}

// NewTransaction creates a new transaction entity
func NewTransaction(userID, subscriptionID uuid.UUID, amount float64, currency string) *Transaction {
	return &Transaction{
		ID:             uuid.New(),
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Amount:         amount,
		Currency:       currency,
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
