package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
)

// WinbackOfferStatus represents the status of a winback offer
type WinbackOfferStatus string

const (
	OfferStatusOffered  WinbackOfferStatus = "offered"
	OfferStatusAccepted WinbackOfferStatus = "accepted"
	OfferStatusExpired  WinbackOfferStatus = "expired"
	OfferStatusDeclined WinbackOfferStatus = "declined"
)

// WinbackOffer represents a discount offer to win back churned users
type WinbackOffer struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	CampaignID    string
	DiscountType  DiscountType
	DiscountValue float64
	Status        WinbackOfferStatus
	OfferedAt     time.Time
	ExpiresAt     time.Time
	AcceptedAt    *time.Time
	CreatedAt     time.Time
}

// NewWinbackOffer creates a new winback offer
func NewWinbackOffer(userID uuid.UUID, campaignID string, discountType DiscountType, discountValue float64, expiresAt time.Time) *WinbackOffer {
	now := time.Now()
	return &WinbackOffer{
		ID:            uuid.New(),
		UserID:        userID,
		CampaignID:    campaignID,
		DiscountType:  discountType,
		DiscountValue: discountValue,
		Status:        OfferStatusOffered,
		OfferedAt:     now,
		ExpiresAt:     expiresAt,
		CreatedAt:     now,
	}
}

// IsActive returns true if the offer is still active
func (o *WinbackOffer) IsActive() bool {
	return o.Status == OfferStatusOffered && o.ExpiresAt.After(time.Now())
}

// IsExpired returns true if the offer has expired
func (o *WinbackOffer) IsExpired() bool {
	return o.Status == OfferStatusExpired || o.ExpiresAt.Before(time.Now())
}

// IsAccepted returns true if the offer has been accepted
func (o *WinbackOffer) IsAccepted() bool {
	return o.Status == OfferStatusAccepted
}

// Accept marks the offer as accepted
func (o *WinbackOffer) Accept() error {
	if o.IsExpired() {
		return errors.New("cannot accept expired offer")
	}

	if o.Status == OfferStatusDeclined {
		return errors.New("cannot accept declined offer")
	}

	now := time.Now()
	o.Status = OfferStatusAccepted
	o.AcceptedAt = &now
	return nil
}

// Expire marks the offer as expired
func (o *WinbackOffer) Expire() error {
	o.Status = OfferStatusExpired
	return nil
}

// Decline marks the offer as declined
func (o *WinbackOffer) Decline() error {
	if o.Status == OfferStatusAccepted {
		return errors.New("cannot decline already accepted offer")
	}

	o.Status = OfferStatusDeclined
	return nil
}

// CalculateDiscountAmount calculates the discount amount for a given total
func (o *WinbackOffer) CalculateDiscountAmount(totalAmount float64) float64 {
	if !o.IsActive() {
		return 0
	}

	var discount float64
	switch o.DiscountType {
	case DiscountTypePercentage:
		discount = totalAmount * (o.DiscountValue / 100.0)
	case DiscountTypeFixed:
		discount = o.DiscountValue
	}

	// Discount cannot exceed total amount
	if discount > totalAmount {
		return totalAmount
	}

	return discount
}

// CalculateFinalAmount calculates the final amount after applying discount
func (o *WinbackOffer) CalculateFinalAmount(totalAmount float64) float64 {
	discount := o.CalculateDiscountAmount(totalAmount)
	return totalAmount - discount
}

// DaysUntilExpiry returns the number of days until the offer expires
func (o *WinbackOffer) DaysUntilExpiry() int {
	if o.IsExpired() {
		return 0
	}

	duration := time.Until(o.ExpiresAt)
	if duration < 0 {
		return 0
	}

	return int(duration.Hours() / 24)
}
