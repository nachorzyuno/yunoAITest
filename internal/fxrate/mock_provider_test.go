package fxrate

import (
	"testing"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewMockProvider(t *testing.T) {
	provider := NewMockProvider()
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.baseRates)
	assert.Equal(t, 5, len(provider.baseRates))
}

func TestMockProvider_GetRate_USD(t *testing.T) {
	provider := NewMockProvider()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	rate, err := provider.GetRate(domain.USD, date)
	assert.NoError(t, err)
	assert.True(t, rate.Equal(decimal.NewFromFloat(1.0)))
}

func TestMockProvider_GetRate_SupportedCurrencies(t *testing.T) {
	provider := NewMockProvider()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		currency domain.Currency
		baseRate float64
	}{
		{"ARS", domain.ARS, 0.0012},
		{"BRL", domain.BRL, 0.20},
		{"COP", domain.COP, 0.00025},
		{"MXN", domain.MXN, 0.055},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := provider.GetRate(tt.currency, date)
			assert.NoError(t, err)
			assert.True(t, rate.GreaterThan(decimal.Zero))

			// Rate should be within ±2% of base rate
			baseRate := decimal.NewFromFloat(tt.baseRate)
			minRate := baseRate.Mul(decimal.NewFromFloat(0.98))
			maxRate := baseRate.Mul(decimal.NewFromFloat(1.02))

			assert.True(t, rate.GreaterThanOrEqual(minRate), "rate should be >= %s, got %s", minRate, rate)
			assert.True(t, rate.LessThanOrEqual(maxRate), "rate should be <= %s, got %s", maxRate, rate)
		})
	}
}

func TestMockProvider_GetRate_UnsupportedCurrency(t *testing.T) {
	provider := NewMockProvider()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	rate, err := provider.GetRate(domain.Currency("EUR"), date)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported currency")
	assert.True(t, rate.Equal(decimal.Zero))
}

func TestMockProvider_GetRate_DeterministicVolatility(t *testing.T) {
	provider := NewMockProvider()
	date1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	// Same date should give same rate
	rate1a, err := provider.GetRate(domain.BRL, date1)
	assert.NoError(t, err)

	rate1b, err := provider.GetRate(domain.BRL, date1)
	assert.NoError(t, err)
	assert.True(t, rate1a.Equal(rate1b), "same date should produce same rate")

	// Different dates should give different rates
	rate2, err := provider.GetRate(domain.BRL, date2)
	assert.NoError(t, err)
	assert.False(t, rate1a.Equal(rate2), "different dates should produce different rates")
}
