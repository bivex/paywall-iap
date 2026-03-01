package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CurrencyConversionRewardStrategy wraps rewards with currency conversion to USD
type CurrencyConversionRewardStrategy struct {
	baseStrategy   RewardStrategy
	currencyService *CurrencyRateService
	logger         *zap.Logger
}

// RewardWithCurrency represents a reward with currency metadata
type RewardWithCurrency struct {
	Value          float64
	Currency       string
	ConvertedValue float64 // USD equivalent
	OriginalValue  float64
	OriginalCurrency string
}

// NewCurrencyConversionRewardStrategy creates a new currency conversion reward strategy
func NewCurrencyConversionRewardStrategy(
	baseStrategy RewardStrategy,
	currencyService *CurrencyRateService,
	logger *zap.Logger,
) *CurrencyConversionRewardStrategy {
	return &CurrencyConversionRewardStrategy{
		baseStrategy:    baseStrategy,
		currencyService: currencyService,
		logger:          logger,
	}
}

// CalculateReward converts the reward to USD before applying the base strategy
func (s *CurrencyConversionRewardStrategy) CalculateReward(
	ctx context.Context,
	baseReward float64,
	arm Arm,
	userContext UserContext,
) (float64, error) {
	// Extract currency from metadata if available
	currency := s.getCurrencyFromContext(userContext)

	// Convert to USD
	rewardUSD, err := s.currencyService.ConvertToUSD(ctx, baseReward, currency)
	if err != nil {
		s.logger.Warn("Failed to convert currency, using original value",
			zap.String("currency", currency),
			zap.Float64("original_value", baseReward),
			zap.Error(err),
		)
		// Fall back to original value if conversion fails
		rewardUSD = baseReward
	}

	s.logger.Debug("Currency conversion applied",
		zap.Float64("original_value", baseReward),
		zap.String("original_currency", currency),
		zap.Float64("converted_value", rewardUSD),
		zap.String("target_currency", "USD"),
	)

	// Apply base strategy calculation
	if s.baseStrategy != nil {
		return s.baseStrategy.CalculateReward(ctx, rewardUSD, arm, userContext)
	}

	return rewardUSD, nil
}

// GetType returns the strategy type
func (s *CurrencyConversionRewardStrategy) GetType() string {
	return "currency_conversion"
}

// getCurrencyFromContext extracts currency from user context or metadata
func (s *CurrencyConversionRewardStrategy) getCurrencyFromContext(userContext UserContext) string {
	// Try to get currency from user's country
	if userContext.Country != "" {
		currency := s.countryToCurrency(userContext.Country)
		if currency != "" {
			return currency
		}
	}

	// Try to get from custom metadata
	if currencyVal, ok := userContext.CustomFeatures["currency"].(string); ok {
		return currencyVal
	}

	// Default to USD
	return "USD"
}

// countryToCurrency maps country codes to currency codes
func (s *CurrencyConversionRewardStrategy) countryToCurrency(countryCode string) string {
	currencyMap := map[string]string{
		"US": "USD",
		"CA": "CAD",
		"GB": "GBP",
		"EU": "EUR",
		"DE": "EUR",
		"FR": "EUR",
		"IT": "EUR",
		"ES": "EUR",
		"NL": "EUR",
		"BE": "EUR",
		"AT": "EUR",
		"IE": "EUR",
		"PT": "EUR",
		"GR": "EUR",
		"FI": "EUR",
		"JP": "JPY",
		"AU": "AUD",
		"CH": "CHF",
		"CN": "CNY",
		"IN": "INR",
		"BR": "BRL",
		"KR": "KRW",
		"MX": "MXN",
		"RU": "RUB",
		"ZA": "ZAR",
		"SG": "SGD",
		"HK": "HKD",
		"NO": "NOK",
		"SE": "SEK",
		"DK": "DKK",
		"PL": "PLN",
		"CZ": "CZK",
		"HU": "HUF",
		"RO": "RON",
		"BG": "BGN",
		"TR": "TRY",
		"IL": "ILS",
		"TH": "THB",
		"MY": "MYR",
		"ID": "IDR",
		"PH": "PHP",
		"VN": "VND",
		"NZ": "NZD",
		"AE": "AED",
		"SA": "SAR",
		"NG": "NGN",
		"EG": "EGP",
	}

	return currencyMap[countryCode]
}

// RecordRewardWithCurrency records a reward with currency metadata
func (s *CurrencyConversionRewardStrategy) RecordRewardWithCurrency(
	ctx context.Context,
	experimentID, armID, userID uuid.UUID,
	rewardValue float64,
	currency string,
	metadata map[string]interface{},
) (*RewardWithCurrency, error) {
	// Convert to USD
	convertedValue, err := s.currencyService.ConvertToUSD(ctx, rewardValue, currency)
	if err != nil {
		s.logger.Warn("Failed to convert currency for recording",
			zap.String("currency", currency),
			zap.Float64("original_value", rewardValue),
			zap.Error(err),
		)
		// Continue with original value if conversion fails
		convertedValue = rewardValue
	}

	reward := &RewardWithCurrency{
		Value:           rewardValue,
		Currency:        currency,
		ConvertedValue:  convertedValue,
		OriginalValue:   rewardValue,
		OriginalCurrency: currency,
	}

	// Add currency metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["original_currency"] = currency
	metadata["original_revenue"] = rewardValue
	metadata["converted_revenue_usd"] = convertedValue

	s.logger.Info("Reward recorded with currency conversion",
		zap.String("experiment_id", experimentID.String()),
		zap.String("arm_id", armID.String()),
		zap.String("user_id", userID.String()),
		zap.Float64("original_value", rewardValue),
		zap.String("original_currency", currency),
		zap.Float64("converted_value", convertedValue),
	)

	return reward, nil
}

// GetConversionRate returns the current conversion rate for a currency
func (s *CurrencyConversionRewardStrategy) GetConversionRate(
	ctx context.Context,
	currency string,
) (float64, error) {
	return s.currencyService.GetRate(ctx, currency)
}

// GetAllSupportedCurrencies returns all supported currency codes
func (s *CurrencyConversionRewardStrategy) GetAllSupportedCurrencies() []string {
	return s.currencyService.GetSupportedCurrencies()
}

// EstimateRevenueUSD estimates revenue in USD from multiple currencies
func (s *CurrencyConversionRewardStrategy) EstimateRevenueUSD(
	ctx context.Context,
	rewards []RewardWithCurrency,
) (float64, error) {
	totalUSD := 0.0

	for _, reward := range rewards {
		if reward.Currency == "USD" {
			totalUSD += reward.Value
		} else {
			converted, err := s.currencyService.ConvertToUSD(ctx, reward.Value, reward.Currency)
			if err != nil {
				s.logger.Warn("Failed to convert reward",
					zap.String("currency", reward.Currency),
					zap.Float64("value", reward.Value),
					zap.Error(err),
				)
				continue
			}
			totalUSD += converted
		}
	}

	return totalUSD, nil
}

// ValidateCurrency checks if a currency code is supported
func (s *CurrencyConversionRewardStrategy) ValidateCurrency(currency string) error {
	supported := s.currencyService.GetSupportedCurrencies()
	for _, c := range supported {
		if c == currency {
			return nil
		}
	}
	return fmt.Errorf("unsupported currency: %s", currency)
}
