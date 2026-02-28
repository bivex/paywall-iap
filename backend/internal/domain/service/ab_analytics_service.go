package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ABTestEvent represents an A/B test event for analytics
type ABTestEvent struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FlagID    string
	Variant   string
	Event     string // "exposed", "converted", "revenue"
	Value     float64
	Timestamp time.Time
}

// ABAnalyticsService tracks A/B test analytics
type ABAnalyticsService struct {
	// In production, this would write to analytics database
	// For now, we'll use in-memory storage
	events []*ABTestEvent
}

// NewABAnalyticsService creates a new A/B analytics service
func NewABAnalyticsService() *ABAnalyticsService {
	return &ABAnalyticsService{
		events: make([]*ABTestEvent, 0),
	}
}

// TrackExposure tracks when a user is exposed to an A/B test variant
func (s *ABAnalyticsService) TrackExposure(ctx context.Context, userID uuid.UUID, flagID, variant string) error {
	event := &ABTestEvent{
		ID:        uuid.New(),
		UserID:    userID,
		FlagID:    flagID,
		Variant:   variant,
		Event:     "exposed",
		Timestamp: time.Now(),
	}

	s.events = append(s.events, event)
	return nil
}

// TrackConversion tracks when a user converts
func (s *ABAnalyticsService) TrackConversion(ctx context.Context, userID uuid.UUID, flagID, variant string) error {
	event := &ABTestEvent{
		ID:        uuid.New(),
		UserID:    userID,
		FlagID:    flagID,
		Variant:   variant,
		Event:     "converted",
		Timestamp: time.Now(),
	}

	s.events = append(s.events, event)
	return nil
}

// TrackRevenue tracks revenue from a user
func (s *ABAnalyticsService) TrackRevenue(ctx context.Context, userID uuid.UUID, flagID, variant string, amount float64) error {
	event := &ABTestEvent{
		ID:        uuid.New(),
		UserID:    userID,
		FlagID:    flagID,
		Variant:   variant,
		Event:     "revenue",
		Value:     amount,
		Timestamp: time.Now(),
	}

	s.events = append(s.events, event)
	return nil
}

// GetConversionRate returns the conversion rate for a variant
func (s *ABAnalyticsService) GetConversionRate(flagID, variant string) float64 {
	var exposed, converted int

	for _, event := range s.events {
		if event.FlagID == flagID && event.Variant == variant {
			if event.Event == "exposed" {
				exposed++
			} else if event.Event == "converted" {
				converted++
			}
		}
	}

	if exposed == 0 {
		return 0
	}

	return float64(converted) / float64(exposed) * 100
}

// GetAverageRevenue returns the average revenue per user for a variant
func (s *ABAnalyticsService) GetAverageRevenue(flagID, variant string) float64 {
	var exposed int
	var totalRevenue float64

	for _, event := range s.events {
		if event.FlagID == flagID && event.Variant == variant {
			if event.Event == "exposed" {
				exposed++
			} else if event.Event == "revenue" {
				totalRevenue += event.Value
			}
		}
	}

	if exposed == 0 {
		return 0
	}

	return totalRevenue / float64(exposed)
}

// GetVariantStats returns statistics for all variants of a flag
func (s *ABAnalyticsService) GetVariantStats(flagID string) map[string]map[string]float64 {
	stats := make(map[string]map[string]float64)

	// Get unique variants
	variants := make(map[string]bool)
	for _, event := range s.events {
		if event.FlagID == flagID {
			variants[event.Variant] = true
		}
	}

	// Calculate stats for each variant
	for variant := range variants {
		stats[variant] = map[string]float64{
			"conversion_rate": s.GetConversionRate(flagID, variant),
			"avg_revenue":     s.GetAverageRevenue(flagID, variant),
		}
	}

	return stats
}
