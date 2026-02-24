package settlement

import (
	"testing"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/ignacio/solara-settlement/internal/fxrate"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.fxService)
	assert.NotNil(t, engine.aggregator)
}

func TestEngine_Calculate_SingleSupplier(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	validTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup123",
			Type:           domain.Refund,
			OriginalAmount: decimal.NewFromFloat(20),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	settlements, err := engine.Calculate(transactions)

	require.NoError(t, err)
	require.Equal(t, 1, len(settlements))

	settlement := settlements[0]
	assert.Equal(t, "sup123", settlement.SupplierID)
	assert.Equal(t, 2, settlement.TransactionCount)
	assert.Equal(t, 2, len(settlement.Lines))

	// USD amounts should be equal to original amounts (rate = 1.0)
	assert.True(t, settlement.TotalCapturesUSD.Equal(decimal.NewFromFloat(100)))
	assert.True(t, settlement.TotalRefundsUSD.Equal(decimal.NewFromFloat(20)))
	assert.True(t, settlement.NetAmountUSD.Equal(decimal.NewFromFloat(80)))
}

func TestEngine_Calculate_MultipleSuppliers(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	validTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup456",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(200),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	settlements, err := engine.Calculate(transactions)

	require.NoError(t, err)
	require.Equal(t, 2, len(settlements))
}

func TestEngine_Calculate_WithFXConversion(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	validTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.BRL,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	settlements, err := engine.Calculate(transactions)

	require.NoError(t, err)
	require.Equal(t, 1, len(settlements))

	settlement := settlements[0]
	assert.Equal(t, 1, settlement.TransactionCount)

	// BRL to USD conversion (base rate ~0.20, with volatility)
	// 100 BRL * ~0.20 = ~20 USD
	assert.True(t, settlement.TotalCapturesUSD.GreaterThan(decimal.NewFromFloat(19)))
	assert.True(t, settlement.TotalCapturesUSD.LessThan(decimal.NewFromFloat(21)))
}

func TestEngine_Calculate_OnlyRefunds(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	validTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Refund,
			OriginalAmount: decimal.NewFromFloat(50),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup123",
			Type:           domain.Refund,
			OriginalAmount: decimal.NewFromFloat(30),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	settlements, err := engine.Calculate(transactions)

	require.NoError(t, err)
	require.Equal(t, 1, len(settlements))

	settlement := settlements[0]
	assert.Equal(t, 2, settlement.TransactionCount)
	assert.True(t, settlement.TotalCapturesUSD.Equal(decimal.Zero))
	assert.True(t, settlement.TotalRefundsUSD.Equal(decimal.NewFromFloat(80)))
	assert.True(t, settlement.NetAmountUSD.Equal(decimal.NewFromFloat(-80)))
}

func TestEngine_Calculate_NoCaptures(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	validTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Refund,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	settlements, err := engine.Calculate(transactions)

	require.NoError(t, err)
	require.Equal(t, 1, len(settlements))

	settlement := settlements[0]
	assert.True(t, settlement.TotalCapturesUSD.Equal(decimal.Zero))
	assert.True(t, settlement.TotalRefundsUSD.Equal(decimal.NewFromFloat(100)))
	assert.True(t, settlement.NetAmountUSD.Equal(decimal.NewFromFloat(-100)))
}

func TestEngine_Calculate_SameDayTransactions(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	sameDay := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.BRL,
			Timestamp:      sameDay,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.BRL,
			Timestamp:      sameDay,
			Status:         domain.Completed,
		},
	}

	settlements, err := engine.Calculate(transactions)

	require.NoError(t, err)
	require.Equal(t, 1, len(settlements))

	settlement := settlements[0]
	assert.Equal(t, 2, settlement.TransactionCount)

	// Both transactions should have the same FX rate (same day)
	line1 := settlement.Lines[0]
	line2 := settlement.Lines[1]
	assert.True(t, line1.FXRate.Equal(line2.FXRate), "same day transactions should have same FX rate")
}

func TestEngine_Calculate_DifferentDayTransactions(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	day1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.BRL,
			Timestamp:      day1,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.BRL,
			Timestamp:      day2,
			Status:         domain.Completed,
		},
	}

	settlements, err := engine.Calculate(transactions)

	require.NoError(t, err)
	require.Equal(t, 1, len(settlements))

	settlement := settlements[0]
	assert.Equal(t, 2, settlement.TransactionCount)

	// Different day transactions should have different FX rates
	line1 := settlement.Lines[0]
	line2 := settlement.Lines[1]
	assert.False(t, line1.FXRate.Equal(line2.FXRate), "different day transactions should have different FX rates")
}

func TestEngine_Calculate_EmptyTransactions(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	settlements, err := engine.Calculate([]*domain.Transaction{})

	require.NoError(t, err)
	assert.Equal(t, 0, len(settlements))
}

func TestEngine_Calculate_MultipleCurrencies(t *testing.T) {
	provider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(provider)
	engine := NewEngine(fxService)

	validTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.BRL,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx003",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(1000),
			Currency:       domain.ARS,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	settlements, err := engine.Calculate(transactions)

	require.NoError(t, err)
	require.Equal(t, 1, len(settlements))

	settlement := settlements[0]
	assert.Equal(t, 3, settlement.TransactionCount)

	// Each currency should be converted appropriately
	assert.True(t, settlement.TotalCapturesUSD.GreaterThan(decimal.Zero))
}
