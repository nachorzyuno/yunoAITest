package fxrate

import (
	"fmt"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
)

// Service provides foreign exchange rate conversion functionality
type Service struct {
	provider Provider
}

// NewService creates a new FX rate service with the given provider
func NewService(provider Provider) *Service {
	return &Service{
		provider: provider,
	}
}

// ConvertToUSD converts the given transaction to USD using the appropriate
// exchange rate for the transaction's date and currency
func (s *Service) ConvertToUSD(tx *domain.Transaction) (decimal.Decimal, decimal.Decimal, error) {
	// Validate the currency
	if err := tx.Currency.Validate(); err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("invalid currency: %w", err)
	}

	// Get the FX rate for the transaction date
	rate, err := s.provider.GetRate(tx.Currency, tx.Timestamp)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("failed to get FX rate: %w", err)
	}

	// Convert to USD
	usdAmount := tx.OriginalAmount.Mul(rate)

	return usdAmount, rate, nil
}

// GetRate retrieves the FX rate for a given currency and date
// This method is used for volatility detection and other rate comparisons
func (s *Service) GetRate(currency domain.Currency, date time.Time) (decimal.Decimal, error) {
	return s.provider.GetRate(currency, date)
}
