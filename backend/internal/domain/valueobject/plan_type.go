package valueobject

import (
	"errors"
)

var (
	ErrInvalidPlanType = errors.New("invalid plan type")
)

type PlanType string

const (
	PlanMonthly  PlanType = "monthly"
	PlanAnnual   PlanType = "annual"
	PlanLifetime PlanType = "lifetime"
)

// NewPlanType creates a new PlanType value object
func NewPlanType(planType string) (PlanType, error) {
	pt := PlanType(planType)
	switch pt {
	case PlanMonthly, PlanAnnual, PlanLifetime:
		return pt, nil
	default:
		return "", ErrInvalidPlanType
	}
}

// String returns the string representation of the plan type
func (p PlanType) String() string {
	return string(p)
}

// IsValid returns true if the plan type is valid
func (p PlanType) IsValid() bool {
	switch p {
	case PlanMonthly, PlanAnnual, PlanLifetime:
		return true
	default:
		return false
	}
}
