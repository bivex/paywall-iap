package service

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ErrCurrencyRateNotFound is returned when a currency rate is not available
var ErrCurrencyRateNotFound = errors.New("currency rate not found")

// CurrencyRateService manages currency exchange rates with caching
type CurrencyRateService struct {
	redisClient *redis.Client
	logger      *zap.Logger
	httpClient  *http.Client

	// Fallback rates used when external API fails
	fallbackRates map[string]float64
	rateMutex     sync.RWMutex

	// ECB API endpoint
	ecbAPIURL string
}

// ECBCurrencyRates represents the ECB daily exchange rate XML structure
type ECBCurrencyRates struct {
	XMLName xml.Name `xml:"Envelope"`
	Cube    struct {
		XMLName xml.Name `xml:"Cube"`
		Cube    struct {
			XMLName xml.Name `xml:"Cube"`
			Time    string   `xml:"time,attr"`
			Cube    []struct {
				Currency string  `xml:"currency,attr"`
				Rate     float64 `xml:"rate,attr"`
			} `xml:"Cube"`
		} `xml:"Cube"`
	} `xml:"Cube"`
}

// CurrencyRate represents a currency exchange rate
type CurrencyRate struct {
	BaseCurrency   string
	TargetCurrency string
	Rate           float64
	Source         string
	UpdatedAt      time.Time
}

// NewCurrencyRateService creates a new currency rate service
func NewCurrencyRateService(redisClient *redis.Client, logger *zap.Logger) *CurrencyRateService {
	return &CurrencyRateService{
		redisClient: redisClient,
		logger:      logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		ecbAPIURL: "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml",
		fallbackRates: map[string]float64{
			"EUR": 0.92,  // Euro to USD
			"GBP": 0.79,  // British Pound to USD
			"JPY": 149.50, // Japanese Yen to USD
			"CAD": 1.36,  // Canadian Dollar to USD
			"AUD": 1.53,  // Australian Dollar to USD
			"CHF": 0.88,  // Swiss Franc to USD
			"CNY": 7.19,  // Chinese Yuan to USD
			"INR": 83.12, // Indian Rupee to USD
			"BRL": 4.97,  // Brazilian Real to USD
			"KRW": 1330.0, // South Korean Won to USD
		},
	}
}

// ConvertToUSD converts an amount from the given currency to USD
func (s *CurrencyRateService) ConvertToUSD(ctx context.Context, amount float64, currency string) (float64, error) {
	if currency == "USD" || currency == "" {
		return amount, nil
	}

	rate, err := s.GetRate(ctx, currency)
	if err != nil {
		return 0, fmt.Errorf("failed to get rate for %s: %w", currency, err)
	}

	// If rate is already USD-based, multiply
	// ECB rates are EUR-based, so we need to convert
	// For simplicity, we store all rates as USD-based in cache
	convertedAmount := amount * rate

	s.logger.Debug("Currency conversion",
		zap.Float64("original_amount", amount),
		zap.String("original_currency", currency),
		zap.Float64("rate", rate),
		zap.Float64("converted_amount", convertedAmount),
	)

	return convertedAmount, nil
}

// GetRate retrieves the exchange rate for a currency (to USD)
func (s *CurrencyRateService) GetRate(ctx context.Context, currency string) (float64, error) {
	// Check Redis cache first
	cacheKey := fmt.Sprintf("currency:rate:%s:USD", currency)
	cachedRate, err := s.redisClient.Get(ctx, cacheKey).Float64()
	if err == nil {
		return cachedRate, nil
	}

	if err != redis.Nil {
		s.logger.Warn("Redis error when fetching rate", zap.Error(err))
	}

	// Try to fetch from ECB API
	rate, source, err := s.fetchRateFromECB(ctx, currency)
	if err != nil {
		s.logger.Warn("Failed to fetch from ECB API, using fallback",
			zap.String("currency", currency),
			zap.Error(err),
		)

		// Use fallback rate
		rate, ok := s.getFallbackRate(currency)
		if !ok {
			return 0, ErrCurrencyRateNotFound
		}
		source = "fallback"
	}

	// Cache the rate for 1 hour
	if err := s.redisClient.Set(ctx, cacheKey, rate, 1*time.Hour).Err(); err != nil {
		s.logger.Warn("Failed to cache currency rate", zap.Error(err))
	}

	s.logger.Info("Currency rate retrieved",
		zap.String("currency", currency),
		zap.Float64("rate", rate),
		zap.String("source", source),
	)

	return rate, nil
}

// fetchRateFromECB fetches exchange rates from the European Central Bank API
func (s *CurrencyRateService) fetchRateFromECB(ctx context.Context, currency string) (float64, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.ecbAPIURL, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to fetch rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read response: %w", err)
	}

	var ecbRates ECBCurrencyRates
	if err := xml.Unmarshal(body, &ecbRates); err != nil {
		return 0, "", fmt.Errorf("failed to parse XML: %w", err)
	}

	// ECB provides EUR-based rates
	// We need to convert to USD-based
	// EUR/USD rate from ECB (inverse)
	var eurToUsdRate float64 = 1.08 // Approximate, will be calculated

	// Find the target currency rate
	for _, cube := range ecbRates.Cube.Cube.Cube {
		if cube.Currency == "USD" {
			eurToUsdRate = cube.Rate
			break
		}
	}

	// Calculate target currency rate to USD
	for _, cube := range ecbRates.Cube.Cube.Cube {
		if cube.Currency == currency {
			// ECB rate is EUR to Currency
			// We want Currency to USD
			// Currency/USD = (EUR/USD) / (EUR/Currency)
			rate := eurToUsdRate / cube.Rate

			// Cache all fetched rates
			s.cacheFetchedRates(ctx, ecbRates, eurToUsdRate)

			return rate, "ecb", nil
		}
	}

	return 0, "", ErrCurrencyRateNotFound
}

// cacheFetchedRates caches all rates from an ECB response
func (s *CurrencyRateService) cacheFetchedRates(ctx context.Context, ecbRates ECBCurrencyRates, eurToUsdRate float64) {
	pipe := s.redisClient.Pipeline()

	for _, cube := range ecbRates.Cube.Cube.Cube {
		cacheKey := fmt.Sprintf("currency:rate:%s:USD", cube.Currency)
		rate := eurToUsdRate / cube.Rate
		pipe.Set(ctx, cacheKey, rate, 1*time.Hour)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Warn("Failed to cache fetched rates", zap.Error(err))
	}
}

// getFallbackRate retrieves a fallback exchange rate
func (s *CurrencyRateService) getFallbackRate(currency string) (float64, bool) {
	s.rateMutex.RLock()
	defer s.rateMutex.RUnlock()

	rate, ok := s.fallbackRates[currency]
	return rate, ok
}

// SetFallbackRate sets a fallback exchange rate
func (s *CurrencyRateService) SetFallbackRate(currency string, rate float64) {
	s.rateMutex.Lock()
	defer s.rateMutex.Unlock()

	s.fallbackRates[currency] = rate
	s.logger.Info("Fallback rate updated",
		zap.String("currency", currency),
		zap.Float64("rate", rate),
	)
}

// UpdateRates triggers an update of all exchange rates from the ECB API
func (s *CurrencyRateService) UpdateRates(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.ecbAPIURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var ecbRates ECBCurrencyRates
	if err := xml.Unmarshal(body, &ecbRates); err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	// Find EUR to USD rate
	var eurToUsdRate float64
	for _, cube := range ecbRates.Cube.Cube.Cube {
		if cube.Currency == "USD" {
			eurToUsdRate = cube.Rate
			break
		}
	}

	if eurToUsdRate == 0 {
		return errors.New("EUR to USD rate not found in ECB response")
	}

	// Cache all rates
	s.cacheFetchedRates(ctx, ecbRates, eurToUsdRate)

	s.logger.Info("Currency rates updated from ECB",
		zap.String("date", ecbRates.Cube.Cube.Time),
		zap.Int("currencies", len(ecbRates.Cube.Cube.Cube)),
	)

	return nil
}

// GetSupportedCurrencies returns a list of supported currencies
func (s *CurrencyRateService) GetSupportedCurrencies() []string {
	currencies := []string{"USD"} // USD is always supported

	s.rateMutex.RLock()
	for currency := range s.fallbackRates {
		currencies = append(currencies, currency)
	}
	s.rateMutex.RUnlock()

	return currencies
}

// HealthCheck verifies the service is operational
func (s *CurrencyRateService) HealthCheck(ctx context.Context) error {
	// Try to get a rate
	_, err := s.GetRate(ctx, "EUR")
	return err
}
