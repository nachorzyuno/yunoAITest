package fxrate

import (
	"testing"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)
	assert.NotNil(t, service)
	assert.NotNil(t, service.provider)
}

func TestService_ConvertToUSD(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	tests := []struct {
		name           string
		tx             *domain.Transaction
		expectError    bool
		expectedMinUSD float64
		expectedMaxUSD float64
	}{
		{
			name: "BRL transaction",
			tx: &domain.Transaction{
				ID:             "tx1",
				SupplierID:     "sup1",
				Type:           domain.Capture,
				OriginalAmount: decimal.NewFromFloat(100),
				Currency:       domain.BRL,
				Timestamp:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				Status:         domain.Completed,
			},
			expectError:    false,
			expectedMinUSD: 19.6,  // 100 * 0.20 * 0.98 (with -2% volatility)
			expectedMaxUSD: 20.4,  // 100 * 0.20 * 1.02 (with +2% volatility)
		},
		{
			name: "ARS transaction",
			tx: &domain.Transaction{
				ID:             "tx2",
				SupplierID:     "sup1",
				Type:           domain.Capture,
				OriginalAmount: decimal.NewFromFloat(1000),
				Currency:       domain.ARS,
				Timestamp:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				Status:         domain.Completed,
			},
			expectError:    false,
			expectedMinUSD: 1.176, // 1000 * 0.0012 * 0.98
			expectedMaxUSD: 1.224, // 1000 * 0.0012 * 1.02
		},
		{
			name: "USD transaction",
			tx: &domain.Transaction{
				ID:             "tx3",
				SupplierID:     "sup1",
				Type:           domain.Capture,
				OriginalAmount: decimal.NewFromFloat(100),
				Currency:       domain.USD,
				Timestamp:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				Status:         domain.Completed,
			},
			expectError:    false,
			expectedMinUSD: 100.0,
			expectedMaxUSD: 100.0,
		},
		{
			name: "invalid currency",
			tx: &domain.Transaction{
				ID:             "tx4",
				SupplierID:     "sup1",
				Type:           domain.Capture,
				OriginalAmount: decimal.NewFromFloat(100),
				Currency:       domain.Currency("EUR"),
				Timestamp:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				Status:         domain.Completed,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usdAmount, rate, err := service.ConvertToUSD(tt.tx)

			if tt.expectError {
				assert.Error(t, err)
				assert.True(t, usdAmount.Equal(decimal.Zero))
				assert.True(t, rate.Equal(decimal.Zero))
			} else {
				assert.NoError(t, err)
				assert.True(t, rate.GreaterThan(decimal.Zero))
				assert.True(t, usdAmount.GreaterThan(decimal.Zero))

				// Check USD amount is within expected range
				minUSD := decimal.NewFromFloat(tt.expectedMinUSD)
				maxUSD := decimal.NewFromFloat(tt.expectedMaxUSD)
				assert.True(t, usdAmount.GreaterThanOrEqual(minUSD),
					"USD amount %s should be >= %s", usdAmount, minUSD)
				assert.True(t, usdAmount.LessThanOrEqual(maxUSD),
					"USD amount %s should be <= %s", usdAmount, maxUSD)
			}
		})
	}
}

func TestService_ConvertToUSD_Deterministic(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	tx := &domain.Transaction{
		ID:             "tx1",
		SupplierID:     "sup1",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100),
		Currency:       domain.BRL,
		Timestamp:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Status:         domain.Completed,
	}

	// Convert same transaction twice
	usd1, rate1, err1 := service.ConvertToUSD(tx)
	assert.NoError(t, err1)

	usd2, rate2, err2 := service.ConvertToUSD(tx)
	assert.NoError(t, err2)

	// Should get same results
	assert.True(t, usd1.Equal(usd2), "USD amounts should be equal")
	assert.True(t, rate1.Equal(rate2), "rates should be equal")
}
