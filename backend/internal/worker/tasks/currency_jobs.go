package tasks

import (
	"context"

	"github.com/bivex/paywall-iap/internal/domain/service"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// CurrencyJobs contains currency-related background jobs
type CurrencyJobs struct {
	currencyService *service.CurrencyRateService
	logger          *zap.Logger
}

// NewCurrencyJobs creates a new currency jobs handler
func NewCurrencyJobs(
	currencyService *service.CurrencyRateService,
	logger *zap.Logger,
) *CurrencyJobs {
	return &CurrencyJobs{
		currencyService: currencyService,
		logger:          logger,
	}
}

// UpdateExchangeRatesArgs represents arguments for the exchange rate update job
type UpdateExchangeRatesArgs struct {
	// No arguments needed for this job
}

// UpdateExchangeRates updates currency exchange rates from ECB
func (j *CurrencyJobs) UpdateExchangeRates(ctx context.Context, _ *river.Job[UpdateExchangeRatesArgs]) error {
	j.logger.Info("Starting currency exchange rate update")

	if err := j.currencyService.UpdateRates(ctx); err != nil {
		j.logger.Error("Failed to update exchange rates", zap.Error(err))
		return err
	}

	j.logger.Info("Currency exchange rates updated successfully")
	return nil
}

// GetSupportedCurrenciesArgs represents arguments for getting supported currencies
type GetSupportedCurrenciesArgs struct {
	// No arguments needed
}

// SupportedCurrenciesResult represents the result of getting supported currencies
type SupportedCurrenciesResult struct {
	Currencies []string
}

// GetSupportedCurrencies returns all supported currency codes
func (j *CurrencyJobs) GetSupportedCurrencies(ctx context.Context, _ *river.Job[GetSupportedCurrenciesArgs]) (SupportedCurrenciesResult, error) {
	currencies := j.currencyService.GetSupportedCurrencies()

	return SupportedCurrenciesResult{
		Currencies: currencies,
	}, nil
}
