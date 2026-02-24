package fxrate

import (
	"fmt"
	"math"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
)

// MockProvider implements Provider with simulated exchange rates
type MockProvider struct {
	baseRates map[domain.Currency]decimal.Decimal
}

// NewMockProvider creates a new mock FX rate provider with base rates
func NewMockProvider() *MockProvider {
	return &MockProvider{
		baseRates: map[domain.Currency]decimal.Decimal{
			domain.ARS: decimal.NewFromFloat(0.0012),   // Argentine Peso
			domain.BRL: decimal.NewFromFloat(0.20),     // Brazilian Real
			domain.COP: decimal.NewFromFloat(0.00025),  // Colombian Peso
			domain.MXN: decimal.NewFromFloat(0.055),    // Mexican Peso
			domain.USD: decimal.NewFromFloat(1.0),      // US Dollar
		},
	}
}

// GetRate returns a simulated exchange rate with daily volatility
// The rate varies by ±2% based on the date to simulate market fluctuations
func (m *MockProvider) GetRate(currency domain.Currency, date time.Time) (decimal.Decimal, error) {
	baseRate, exists := m.baseRates[currency]
	if !exists {
		return decimal.Zero, fmt.Errorf("unsupported currency: %s", currency)
	}

	// USD doesn't need conversion
	if currency == domain.USD {
		return baseRate, nil
	}

	// Apply date-based volatility (±2%)
	// Use date as seed for deterministic "randomness"
	daysSinceEpoch := date.Unix() / 86400
	volatility := math.Sin(float64(daysSinceEpoch)) * 0.02 // ±2%

	adjustedRate := baseRate.Mul(decimal.NewFromFloat(1.0 + volatility))

	return adjustedRate, nil
}
