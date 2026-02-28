package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	matomoClient "github.com/bivex/paywall-iap/internal/infrastructure/external/matomo"
)

// Task types for cohort processing
const (
	TypeCohortAggregation = "cohort:aggregate"
)

// CohortWorker handles cohort aggregation jobs
type CohortWorker struct {
	matomoClient *matomoClient.Client
	analyticsRepo AnalyticsRepository
	logger       *zap.Logger
}

// AnalyticsRepository defines the interface for storing analytics data
type AnalyticsRepository interface {
	StoreCohortData(ctx context.Context, data *CohortAggregate) error
	GetCohortData(ctx context.Context, metricName string, date time.Time) (*CohortAggregate, error)
}

// CohortAggregate represents aggregated cohort data
type CohortAggregate struct {
	ID          uuid.UUID
	MetricName  string
	MetricDate  time.Time
	CohortSize  int
	Retention   map[string]int    // day1, day7, day30 -> count
	Revenue     map[string]float64 // day1, day7, day30 -> revenue
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewCohortWorker creates a new cohort worker
func NewCohortWorker(
	matomoClient *matomoClient.Client,
	analyticsRepo AnalyticsRepository,
) *CohortWorker {
	return &CohortWorker{
		matomoClient: matomoClient,
		analyticsRepo: analyticsRepo,
		logger:       logging.Logger,
	}
}

// CohortAggregationPayload represents the job payload
type CohortAggregationPayload struct {
	Date         string `json:"date"`         // YYYY-MM-DD
	CohortPeriod string `json:"cohort_period"` // day, week, month
	Periods      int    `json:"periods"`      // Number of periods to analyze
}

// NewCohortAggregationTask creates a new cohort aggregation task
func NewCohortAggregationTask(date string, cohortPeriod string, periods int) (*asynq.Task, error) {
	payload := CohortAggregationPayload{
		Date:         date,
		CohortPeriod: cohortPeriod,
		Periods:      periods,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeCohortAggregation, data), nil
}

// HandleCohortAggregation processes cohort aggregation job
func (w *CohortWorker) HandleCohortAggregation(ctx context.Context, t *asynq.Task) error {
	var payload CohortAggregationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("Failed to unmarshal payload", zap.Error(err))
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Parse date
	date, err := time.Parse("2006-01-02", payload.Date)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	w.logger.Info("Processing cohort aggregation",
		zap.String("date", payload.Date),
		zap.String("period", payload.CohortPeriod),
		zap.Int("periods", payload.Periods),
	)

	// Fetch cohort data from Matomo
	matomoReq := matomoClient.CohortRequest{
		DateFrom:     date.AddDate(0, 0, -payload.Periods),
		DateTo:       date,
		CohortPeriod: payload.CohortPeriod,
	}

	cohortData, err := w.matomoClient.GetCohorts(ctx, matomoReq)
	if err != nil {
		w.logger.Error("Failed to fetch cohort data from Matomo", zap.Error(err))
		return fmt.Errorf("failed to fetch cohorts: %w", err)
	}

	// Aggregate data for each cohort
	for _, cohort := range cohortData.Cohorts {
		aggregate := &CohortAggregate{
			ID:         uuid.New(),
			MetricName: fmt.Sprintf("cohort_%s_%s", payload.CohortPeriod, cohort.Period),
			MetricDate: date,
			CohortSize: cohort.SampleSize,
			Retention:  make(map[string]int),
			Revenue:    make(map[string]float64),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// Extract retention data
		for day, count := range cohort.Retention {
			aggregate.Retention[day] = count
		}

		// Extract revenue data
		for day, rev := range cohort.Metrics {
			aggregate.Revenue[day] = rev
		}

		// Store in database
		if err := w.analyticsRepo.StoreCohortData(ctx, aggregate); err != nil {
			w.logger.Error("Failed to store cohort data",
				zap.String("cohort", cohort.Period),
				zap.Error(err),
			)
			continue
		}

		w.logger.Debug("Stored cohort aggregate",
			zap.String("metric_name", aggregate.MetricName),
			zap.Int("cohort_size", aggregate.CohortSize),
		)
	}

	w.logger.Info("Cohort aggregation completed",
		zap.String("date", payload.Date),
		zap.Int("cohorts_processed", len(cohortData.Cohorts)),
	)

	return nil
}

// ScheduleDailyCohortAggregation schedules daily cohort aggregation jobs
func ScheduleDailyCohortAggregation(scheduler *asynq.Scheduler, hour int, periods int) error {
	// Schedule for daily execution at specified hour
	cron := fmt.Sprintf("0 %d * * *", hour)

	payload := CohortAggregationPayload{
		Date:         time.Now().Format("2006-01-02"),
		CohortPeriod: "day",
		Periods:      periods,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeCohortAggregation, data)

	entryID, err := scheduler.Register(
		cron,
		task,
	)

	if err != nil {
		return fmt.Errorf("failed to register cohort aggregation: %w", err)
	}

	logging.Logger.Info("Scheduled daily cohort aggregation",
		zap.String("entry_id", entryID),
		zap.Int("hour", hour),
		zap.Int("periods", periods),
	)

	return nil
}

// RegisterCohortHandlers registers cohort task handlers with the Asynq mux
func RegisterCohortHandlers(mux *asynq.ServeMux, worker *CohortWorker) {
	mux.HandleFunc(TypeCohortAggregation, worker.HandleCohortAggregation)
}

// CalculateLTVFromCohorts calculates LTV estimates from cohort data
func (w *CohortWorker) CalculateLTVFromCohorts(ctx context.Context, userID uuid.UUID) (map[string]float64, error) {
	// Get user's cohort data (assuming user joined 30 days ago)
	joinDate := time.Now().AddDate(0, 0, -30)

	// Fetch cohort aggregates for the past 30 days
	metricName := "cohort_day_" + joinDate.Format("2006-01-02")
	cohortData, err := w.analyticsRepo.GetCohortData(ctx, metricName, joinDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cohort data: %w", err)
	}

	// Calculate LTV estimates
	ltv := make(map[string]float64)

	// LTV30 - average revenue from cohort over 30 days
	if rev30, ok := cohortData.Revenue["day30"]; ok && cohortData.CohortSize > 0 {
		ltv["ltv30"] = rev30 / float64(cohortData.CohortSize)
	}

	// LTV90 - extrapolate from 30-day data
	if rev30, ok := cohortData.Revenue["day30"]; ok {
		ltv["ltv90"] = rev30 * 3 // Simple 3x extrapolation
	}

	// LTV365 - extrapolate from 30-day data
	if rev30, ok := cohortData.Revenue["day30"]; ok {
		ltv["ltv365"] = rev30 * 12 // Simple 12x extrapolation
	}

	return ltv, nil
}

// GetCohortMetrics retrieves cohort metrics for a date range
func (w *CohortWorker) GetCohortMetrics(ctx context.Context, fromDate, toDate time.Time) ([]CohortMetrics, error) {
	metrics := make([]CohortMetrics, 0)

	// Fetch data for each day in the range
	for d := fromDate; !d.After(toDate); d = d.AddDate(0, 0, 1) {
		metricName := "cohort_day_" + d.Format("2006-01-02")
		data, err := w.analyticsRepo.GetCohortData(ctx, metricName, d)
		if err != nil {
			w.logger.Warn("Failed to get cohort data",
				zap.String("date", d.Format("2006-01-02")),
				zap.Error(err),
			)
			continue
		}

		metric := CohortMetrics{
			Date:        d,
			CohortSize:  data.CohortSize,
			Retention:   data.Retention,
			Revenue:     data.Revenue,
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// CohortMetrics represents aggregated cohort metrics
type CohortMetrics struct {
	Date       time.Time
	CohortSize int
	Retention  map[string]int
	Revenue    map[string]float64
}
