package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
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
	matomoClient     *matomo.Client
	cohortWorker     CohortWorker
	subscriptionRepo SubscriptionRepository
	transactionRepo  domainRepo.TransactionRepository
	userRepo         domainRepo.UserRepository
	logger           *zap.Logger
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

// LTVServiceParams encapsulates dependencies for LTVService
type LTVServiceParams struct {
	MatomoClient     *matomo.Client
	CohortWorker     CohortWorker
	SubscriptionRepo SubscriptionRepository
	TransactionRepo  domainRepo.TransactionRepository
	Logger           *zap.Logger
}

// NewLTVService creates a new LTV service
func NewLTVService(params LTVServiceParams) *LTVService {
	return &LTVService{
		matomoClient:     params.MatomoClient,
		cohortWorker:     params.CohortWorker,
		subscriptionRepo: params.SubscriptionRepo,
		transactionRepo:  params.TransactionRepo,
		logger:           params.Logger,
	}
}

// WithUserRepo sets the user repository for LTV updates (optional, enables DB-backed UpdateUserLTV)
func (s *LTVService) WithUserRepo(userRepo domainRepo.UserRepository) *LTVService {
	s.userRepo = userRepo
	return s
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

func (s *LTVService) populateActualLTV(estimates *LTVEstimates, subs []Subscription) {
	if len(subs) == 0 {
		return
	}
	firstSub := subs[0]
	daysSinceFirst := int(time.Since(firstSub.CreatedAt).Hours() / 24)

	if daysSinceFirst >= 30 {
		estimates.LTV30 = s.getRevenueInPeriod(subs, 30)
		estimates.Factors["actual_30day"] = estimates.LTV30
	}
	if daysSinceFirst >= 90 {
		estimates.LTV90 = s.getRevenueInPeriod(subs, 90)
		estimates.Factors["actual_90day"] = estimates.LTV90
	}
	if daysSinceFirst >= 365 {
		estimates.LTV365 = s.getRevenueInPeriod(subs, 365)
		estimates.Factors["actual_365day"] = estimates.LTV365
	}
}

func (s *LTVService) populatePredictedLTV(ctx context.Context, userID uuid.UUID, estimates *LTVEstimates) {
	if estimates.LTV30 == 0 {
		if ltv30, err := s.predictLTVFromCohorts(ctx, userID, 30); err == nil {
			estimates.LTV30 = ltv30
			estimates.Factors["predicted_30day"] = ltv30
		}
	}
	if estimates.LTV90 == 0 {
		if ltv90, err := s.predictLTVFromCohorts(ctx, userID, 90); err == nil {
			estimates.LTV90 = ltv90
			estimates.Factors["predicted_90day"] = ltv90
		}
	}
	if estimates.LTV365 == 0 {
		if ltv365, err := s.predictLTVFromCohorts(ctx, userID, 365); err == nil {
			estimates.LTV365 = ltv365
			estimates.Factors["predicted_365day"] = ltv365
		}
	}
}

func (s *LTVService) loadSubscriptionsAndRevenue(ctx context.Context, userID uuid.UUID) ([]Subscription, float64, error) {
	if s.subscriptionRepo == nil {
		return nil, 0, nil
	}
	subs, err := s.subscriptionRepo.GetUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	totalRevenue, err := s.subscriptionRepo.GetTotalRevenue(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total revenue: %w", err)
	}
	return subs, totalRevenue, nil
}

// CalculateLTV calculates LTV estimates for a user
func (s *LTVService) CalculateLTV(ctx context.Context, userID uuid.UUID) (*LTVEstimates, error) {
	subs, totalRevenue, err := s.loadSubscriptionsAndRevenue(ctx, userID)
	if err != nil {
		return nil, err
	}

	estimates := &LTVEstimates{
		UserID:       userID.String(),
		LTVLifetime:  totalRevenue,
		CalculatedAt: time.Now(),
		Method:       "cohort_based",
		Factors:      make(map[string]float64),
	}

	s.populateActualLTV(estimates, subs)
	s.populatePredictedLTV(ctx, userID, estimates)

	estimates.Confidence = s.calculateConfidence(estimates, subs)

	s.logger.Debug("Calculated LTV",
		zap.String("user_id", userID.String()),
		zap.Float64("ltv30", estimates.LTV30),
		zap.Float64("ltv90", estimates.LTV90),
		zap.Float64("ltv365", estimates.LTV365),
		zap.Float64("confidence", estimates.Confidence),
	)

	return estimates, nil
}

// predictLTVFromCohorts predicts LTV using cohort data
func (s *LTVService) predictLTVFromCohorts(ctx context.Context, userID uuid.UUID, days int) (float64, error) {
	if s.cohortWorker != nil {
		ltvMap, err := s.cohortWorker.CalculateLTVFromCohorts(ctx, userID)
		if err != nil {
			s.logger.Warn("Failed to get cohort LTV, using default",
				zap.String("user_id", userID.String()),
				zap.Error(err),
			)
			return s.getDefaultLTV(days), nil
		}
		if val, ok := lookupCohortLTV(ltvMap, days); ok {
			return val, nil
		}
	}

	return s.getDefaultLTV(days), nil
}

func lookupCohortLTV(ltvMap map[string]float64, days int) (float64, bool) {
	key := fmt.Sprintf("ltv%d", days)
	val, ok := ltvMap[key]
	return val, ok
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

func accumulateCohortMetricRevenue(cohortLTV *CohortLTV, metric CohortMetrics) {
	if metric.CohortSize <= 0 {
		return
	}
	cohortSize := float64(metric.CohortSize)
	if rev30, ok := metric.Revenue["day30"]; ok {
		cohortLTV.LTV30 += rev30 / cohortSize
	}
	if rev90, ok := metric.Revenue["day90"]; ok {
		cohortLTV.LTV90 += rev90 / cohortSize
	}
	if rev365, ok := metric.Revenue["day365"]; ok {
		cohortLTV.LTV365 += rev365 / cohortSize
	}
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
		accumulateCohortMetricRevenue(cohortLTV, metric)
	}

	// Average across all metrics
	numMetrics := float64(len(metrics))
	cohortLTV.LTV30 /= numMetrics
	cohortLTV.LTV90 /= numMetrics
	cohortLTV.LTV365 /= numMetrics

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
	s.logger.Debug("User LTV updated",
		zap.String("user_id", userID.String()),
		zap.Float64("amount", amount),
	)
	if s.userRepo != nil {
		if err := s.userRepo.IncrementLTV(ctx, userID, amount); err != nil {
			s.logger.Warn("Failed to persist LTV increment",
				zap.String("user_id", userID.String()),
				zap.Float64("amount", amount),
				zap.Error(err),
			)
		}
	}
	return nil
}

// GetSegmentedLTV calculates LTV for user segments based on real platform data
func (s *LTVService) GetSegmentedLTV(ctx context.Context, segment string, period int) (map[string]float64, error) {
	rows, err := s.transactionRepo.GetSegmentedLTV(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get segmented LTV: %w", err)
	}

	if segment != "" && segment != "all" {
		if val, ok := rows[segment]; ok {
			return map[string]float64{segment: val}, nil
		}
		return map[string]float64{segment: 0}, nil
	}

	return rows, nil
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
