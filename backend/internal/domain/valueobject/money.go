package valueobject

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidAmount   = errors.New("amount must be non-negative")
	ErrInvalidCurrency = errors.New("invalid currency code")
)

// Money represents a monetary value
type Money struct {
	Amount   float64
	Currency string // ISO 4217 currency code (e.g., "USD", "EUR")
}

// NewMoney creates a new Money value object
func NewMoney(amount float64, currency string) (*Money, error) {
	if amount < 0 {
		return nil, fmt.Errorf("%w: %f", ErrInvalidAmount, amount)
	}
	if !isValidCurrency(currency) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCurrency, currency)
	}
	return &Money{
		Amount:   amount,
		Currency: currency,
	}, nil
}

// isValidCurrency checks if the currency code is valid (3 letters)
func isValidCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, c := range currency {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}

// String returns a string representation of the money
func (m *Money) String() string {
	return fmt.Sprintf("%.2f %s", m.Amount, m.Currency)
}

// IsZero returns true if the amount is zero
func (m *Money) IsZero() bool {
	return m.Amount == 0
}

// Add adds another Money value to this one
func (m *Money) Add(other *Money) (*Money, error) {
	if m.Currency != other.Currency {
		return nil, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}
	return NewMoney(m.Amount+other.Amount, m.Currency)
}
