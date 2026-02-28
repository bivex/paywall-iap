package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/infrastructure/external/matomo"
)

// CohortWorker defines the interface for cohort-based LTV calculations
type CohortWorker interface {
	CalculateLTVFromCohorts(ctx context.Context, userID uuid.UUID) (map[string]float64, error)
	GetCohortMetrics(ctx context.Context, fromDate, toDate time.Time) ([]CohortMetrics, error)
}

// CohortMetrics represents cohort analytics data
type CohortMetrics struct {
	CohortSize int                    `json:"cohort_size"`
	Retention  map[string]int         `json:"retention"`
	Revenue    map[string]float64     `json:"revenue"`
}

// LTVService handles Lifetime Value calculations and predictions
type LTVService struct {
	matomoClient      *matomo.Client
	cohortWorker      CohortWorker
	subscriptionRepo  SubscriptionRepository
	logger            *zap.Logger
}

// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
	GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
	GetTotalRevenue(ctx context.Context, userID uuid.UUID) (float64, error)
}

// Subscription represents a user subscription
type Subscription struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Status    string
	Revenue   float64
	CreatedAt time.Time
	EndDate   *time.Time
}

// NewLTVService creates a new LTV service
func NewLTVService(
	matomoClient *matomo.Client,
	cohortWorker CohortWorker,
	subscriptionRepo SubscriptionRepository,
	logger *zap.Logger,
) *LTVService {
	return &LTVService{
		matomoClient:     matomoClient,
		cohortWorker:     cohortWorker,
		subscriptionRepo: subscriptionRepo,
		logger:           logger,
	}
}

// LTVEstimates represents LTV predictions for different time horizons
type LTVEstimates struct {
	UserID        string            `json:"user_id"`
	LTV30         float64           `json:"ltv30"`
	LTV90         float64           `json:"ltv90"`
	LTV365        float64           `json:"ltv365"`
	LTVLifetime   float64           `json:"ltv_lifetime"`
	Confidence    float64           `json:"confidence"`
	CalculatedAt  time.Time         `json:"calculated_at"`
	Method        string            `json:"method"`
	Factors       map[string]float64 `json:"factors"`
}

// CalculateLTV calculates LTV estimates for a user
func (s *LTVService) CalculateLTV(ctx context.Context, userID uuid.UUID) (*LTVEstimates, error) {
	// Get user's subscription history
	subs, err := s.subscriptionRepo.GetUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	// Get total revenue to date
	totalRevenue, err := s.subscriptionRepo.GetTotalRevenue(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}

	// Calculate base LTV from actual data
	estimates := &LTVEstimates{
		UserID:       userID.String(),
		LTVLifetime:  totalRevenue,
		CalculatedAt: time.Now(),
		Method:       "cohort_based",
		Factors:      make(map[string]float64),
	}

	// If user has enough history, use actual data
	if len(subs) > 0 {
		firstSub := subs[0]
		daysSinceFirst := int(time.Since(firstSub.CreatedAt).Hours() / 24)

		if daysSinceFirst >= 30 {
			// Calculate 30-day LTV from actual data
			estimates.LTV30 = s.getRevenueInPeriod(subs, 30)
			estimates.Factors["actual_30day"] = estimates.LTV30
		}

		if daysSinceFirst >= 90 {
			// Calculate 90-day LTV from actual data
			estimates.LTV90 = s.getRevenueInPeriod(subs, 90)
			estimates.Factors["actual_90day"] = estimates.LTV90
		}

		if daysSinceFirst >= 365 {
			// Calculate 365-day LTV from actual data
			estimates.LTV365 = s.getRevenueInPeriod(subs, 365)
			estimates.Factors["actual_365day"] = estimates.LTV365
		}
	}

	// For missing time horizons, use cohort-based predictions
	if estimates.LTV30 == 0 {
		ltv30, err := s.predictLTVFromCohorts(ctx, userID, 30)
		if err == nil {
			estimates.LTV30 = ltv30
			estimates.Factors["predicted_30day"] = ltv30
		}
	}

	if estimates.LTV90 == 0 {
		ltv90, err := s.predictLTVFromCohorts(ctx, userID, 90)
		if err == nil {
			estimates.LTV90 = ltv90
			estimates.Factors["predicted_90day"] = ltv90
		}
	}

	if estimates.LTV365 == 0 {
		ltv365, err := s.predictLTVFromCohorts(ctx, userID, 365)
		if err == nil {
			estimates.LTV365 = ltv365
			estimates.Factors["predicted_365day"] = ltv365
		}
	}

	// Calculate confidence based on data availability
	confidence := s.calculateConfidence(estimates, subs)
	estimates.Confidence = confidence

	s.logger.Debug("Calculated LTV",
		zap.String("user_id", userID.String()),
		zap.Float64("ltv30", estimates.LTV30),
		zap.Float64("ltv90", estimates.LTV90),
		zap.Float64("ltv365", estimates.LTV365),
		zap.Float64("confidence", confidence),
	)

	return estimates, nil
}

// predictLTVFromCohorts predicts LTV using cohort data
func (s *LTVService) predictLTVFromCohorts(ctx context.Context, userID uuid.UUID, days int) (float64, error) {
	// Get user's join date (first subscription)
	subs, err := s.subscriptionRepo.GetUserSubscriptions(ctx, userID)
	if err != nil || len(subs) == 0 {
		// Use default LTV based on product pricing
		return s.getDefaultLTV(days), nil
	}

	// Get LTV from cohort worker
	ltvMap, err := s.cohortWorker.CalculateLTVFromCohorts(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get cohort LTV, using default",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return s.getDefaultLTV(days), nil
	}

	// Extract the requested time horizon
	switch days {
	case 30:
		return ltvMap["ltv30"], nil
	case 90:
		return ltvMap["ltv90"], nil
	case 365:
		return ltvMap["ltv365"], nil
	default:
		return s.getDefaultLTV(days), nil
	}
}

// getDefaultLTV returns default LTV estimates based on product pricing
func (s *LTVService) getDefaultLTV(days int) float64 {
	// Default pricing: $9.99/month
	monthlyPrice := 9.99

	switch days {
	case 30:
		return monthlyPrice
	case 90:
		return monthlyPrice * 3
	case 365:
		// Assume 10% monthly discount for annual
		return monthlyPrice * 12 * 0.9
	default:
		// Extrapolate
		return monthlyPrice * float64(days) / 30
	}
}

// getRevenueInPeriod calculates revenue within a specific time period from first subscription
func (s *LTVService) getRevenueInPeriod(subs []Subscription, days int) float64 {
	if len(subs) == 0 {
		return 0
	}

	firstSub := subs[0]
	cutoffDate := firstSub.CreatedAt.AddDate(0, 0, days)
	now := time.Now()

	// If cutoff is in the future, only count actual revenue so far
	if cutoffDate.After(now) {
		cutoffDate = now
	}

	var totalRevenue float64
	for _, sub := range subs {
		if sub.CreatedAt.After(cutoffDate) {
			continue
		}

		// Calculate revenue for this subscription within the period
		subRevenue := s.calculateSubscriptionRevenue(sub, firstSub.CreatedAt, cutoffDate)
		totalRevenue += subRevenue
	}

	return totalRevenue
}

// calculateSubscriptionRevenue calculates revenue for a subscription within a date range
func (s *LTVService) calculateSubscriptionRevenue(sub Subscription, startDate, endDate time.Time) float64 {
	// Simple calculation: revenue * number of billing cycles within period
	// Assuming monthly billing for now
	monthlyRevenue := sub.Revenue / 12 // Convert annual to monthly if needed

	// Count months in period
	months := 0
	currentDate := startDate

	for currentDate.Before(endDate) {
		// Check if subscription was active in this month
		if sub.Status == "active" || sub.Status == "grace" {
			months++
		}
		currentDate = currentDate.AddDate(0, 1, 0)
	}

	return float64(months) * monthlyRevenue
}

// calculateConfidence calculates confidence score based on data availability
func (s *LTVService) calculateConfidence(estimates *LTVEstimates, subs []Subscription) float64 {
	confidence := 0.0

	// Base confidence from actual data
	if estimates.LTV365 > 0 {
		confidence += 0.8 // High confidence with 365 days of data
	} else if estimates.LTV90 > 0 {
		confidence += 0.6 // Medium confidence with 90 days
	} else if estimates.LTV30 > 0 {
		confidence += 0.4 // Low confidence with only 30 days
	}

	// Boost confidence if user has multiple subscriptions
	if len(subs) > 1 {
		confidence += 0.1
	}

	// Boost confidence if user is currently active
	for _, sub := range subs {
		if sub.Status == "active" {
			confidence += 0.1
			break
		}
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// GetCohortLTV calculates LTV for an entire cohort
func (s *LTVService) GetCohortLTV(ctx context.Context, cohortDate time.Time) (*CohortLTV, error) {
	// Fetch cohort metrics from worker
	metrics, err := s.cohortWorker.GetCohortMetrics(ctx, cohortDate, cohortDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cohort metrics: %w", err)
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("no metrics found for cohort")
	}

	// Calculate aggregate LTV
	cohortLTV := &CohortLTV{
		CohortDate: cohortDate,
		CohortSize: 0,
		LTV30:      0,
		LTV90:      0,
		LTV365:     0,
	}

	for _, metric := range metrics {
		cohortLTV.CohortSize += metric.CohortSize

		// Average revenue per user
		if metric.CohortSize > 0 {
			if rev30, ok := metric.Revenue["day30"]; ok {
				cohortLTV.LTV30 += rev30 / float64(metric.CohortSize)
			}
			if rev90, ok := metric.Revenue["day90"]; ok {
				cohortLTV.LTV90 += rev90 / float64(metric.CohortSize)
			}
			if rev365, ok := metric.Revenue["day365"]; ok {
				cohortLTV.LTV365 += rev365 / float64(metric.CohortSize)
			}
		}
	}

	// Average across all metrics
	if len(metrics) > 0 {
		cohortLTV.LTV30 /= float64(len(metrics))
		cohortLTV.LTV90 /= float64(len(metrics))
		cohortLTV.LTV365 /= float64(len(metrics))
	}

	return cohortLTV, nil
}

// CohortLTV represents LTV metrics for a cohort
type CohortLTV struct {
	CohortDate time.Time `json:"cohort_date"`
	CohortSize int       `json:"cohort_size"`
	LTV30      float64   `json:"ltv30"`
	LTV90      float64   `json:"ltv90"`
	LTV365     float64   `json:"ltv365"`
}

// UpdateUserLTV updates LTV after a new purchase
func (s *LTVService) UpdateUserLTV(ctx context.Context, userID uuid.UUID, amount float64) error {
	// This would trigger a recalculation of LTV
	// For now, we'll just log the event
	s.logger.Debug("User LTV updated",
		zap.String("user_id", userID.String()),
		zap.Float64("amount", amount),
	)

	// TODO: Invalidate cached LTV for this user
	// TODO: Trigger async LTV recalculation

	return nil
}

// GetSegmentedLTV calculates LTV for user segments
func (s *LTVService) GetSegmentedLTV(ctx context.Context, segment string, period int) (map[string]float64, error) {
	// TODO: Implement segmentation based on user attributes
	// For now, return mock data

	segments := map[string]float64{
		"all_users":    29.99,
		"ios_premium":  35.99,
		"android_free": 9.99,
		"new_users":    15.00,
		"churned":      19.99,
	}

	if result, ok := segments[segment]; ok {
		return map[string]float64{segment: result}, nil
	}

	return segments, nil
}

// PredictChurnRisk predicts the likelihood of a user churning
func (s *LTVService) PredictChurnRisk(ctx context.Context, userID uuid.UUID) (float64, error) {
	subs, err := s.subscriptionRepo.GetUserSubscriptions(ctx, userID)
	if err != nil {
		return 0, err
	}

	// Simple churn risk model
	risk := 0.0

	if len(subs) == 0 {
		risk = 1.0 // No subscriptions = high risk
		return risk, nil
	}

	latestSub := subs[len(subs)-1]

	// Check subscription status
	if latestSub.Status == "cancelled" || latestSub.Status == "expired" {
		risk = 0.9
	} else if latestSub.Status == "grace" {
		risk = 0.7
	} else {
		risk = 0.1 // Active users have low risk
	}

	// Adjust based on subscription age
	daysSinceSub := int(time.Since(latestSub.CreatedAt).Hours() / 24)
	if daysSinceSub < 7 {
		risk += 0.1 // New users have slightly higher risk
	} else if daysSinceSub > 90 {
		risk -= 0.1 // Long-term users have lower risk
	}

	// Cap at 1.0
	if risk > 1.0 {
		risk = 1.0
	}
	if risk < 0 {
		risk = 0
	}

	return risk, nil
}
