package settlement

import (
	"strings"
	"testing"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/ignacio/solara-settlement/internal/fxrate"
	"github.com/ignacio/solara-settlement/internal/processor"
	"github.com/ignacio/solara-settlement/internal/reporter"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndSettlementFlow tests the complete settlement workflow
func TestEndToEndSettlementFlow(t *testing.T) {
	// Step 1: Prepare test CSV data
	csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup123,capture,1000.00,BRL,2024-01-15T10:30:00Z,completed
tx002,sup123,refund,200.00,BRL,2024-01-16T14:20:00Z,completed
tx003,sup456,capture,50000.00,ARS,2024-01-15T11:45:00Z,completed
tx004,sup456,capture,500.00,MXN,2024-01-17T09:15:00Z,completed
tx005,sup789,capture,100000.00,COP,2024-01-18T16:30:00Z,completed`

	// Step 2: Parse CSV
	csvReader := processor.NewCSVReader()
	transactions, err := csvReader.Read(strings.NewReader(csvData))
	require.NoError(t, err)
	require.Equal(t, 5, len(transactions))

	// Step 3: Validate transactions
	validator := processor.NewValidator()
	err = validator.ValidateBatch(transactions)
	require.NoError(t, err)

	// Step 4: Calculate settlements
	fxProvider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(fxProvider)
	engine := NewEngine(fxService)

	settlements, err := engine.Calculate(transactions)
	require.NoError(t, err)
	require.Equal(t, 3, len(settlements))

	// Step 5: Generate CSV report
	csvWriter := reporter.NewCSVWriter()
	var output strings.Builder
	err = csvWriter.Write(&output, settlements)
	require.NoError(t, err)

	// Step 6: Verify output
	outputStr := output.String()
	assert.Contains(t, outputStr, "supplier_id")
	assert.Contains(t, outputStr, "sup123")
	assert.Contains(t, outputStr, "sup456")
	assert.Contains(t, outputStr, "sup789")
	assert.Contains(t, outputStr, "SUMMARY")

	// Verify settlement calculations
	var sup123Settlement *domain.SupplierSettlement
	for _, s := range settlements {
		if s.SupplierID == "sup123" {
			sup123Settlement = s
			break
		}
	}

	require.NotNil(t, sup123Settlement)
	assert.Equal(t, 2, sup123Settlement.TransactionCount)
	assert.True(t, sup123Settlement.TotalCapturesUSD.GreaterThan(decimal.Zero))
	assert.True(t, sup123Settlement.TotalRefundsUSD.GreaterThan(decimal.Zero))
	assert.True(t, sup123Settlement.NetAmountUSD.GreaterThan(decimal.Zero))
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("only completed transactions are settled", func(t *testing.T) {
		csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup123,capture,100.00,USD,2024-01-15T10:30:00Z,completed
tx002,sup123,capture,50.00,USD,2024-01-16T14:20:00Z,pending
tx003,sup123,capture,30.00,USD,2024-01-17T11:45:00Z,failed`

		csvReader := processor.NewCSVReader()
		transactions, err := csvReader.Read(strings.NewReader(csvData))
		require.NoError(t, err)

		fxProvider := fxrate.NewMockProvider()
		fxService := fxrate.NewService(fxProvider)
		engine := NewEngine(fxService)

		settlements, err := engine.Calculate(transactions)
		require.NoError(t, err)
		require.Equal(t, 1, len(settlements))

		// Only the completed transaction should be included
		assert.Equal(t, 1, settlements[0].TransactionCount)
		assert.True(t, settlements[0].TotalCapturesUSD.Equal(decimal.NewFromFloat(100.00)))
	})

	t.Run("all refunds result in negative net", func(t *testing.T) {
		csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup123,refund,100.00,USD,2024-01-15T10:30:00Z,completed
tx002,sup123,refund,50.00,USD,2024-01-16T14:20:00Z,completed`

		csvReader := processor.NewCSVReader()
		transactions, err := csvReader.Read(strings.NewReader(csvData))
		require.NoError(t, err)

		fxProvider := fxrate.NewMockProvider()
		fxService := fxrate.NewService(fxProvider)
		engine := NewEngine(fxService)

		settlements, err := engine.Calculate(transactions)
		require.NoError(t, err)
		require.Equal(t, 1, len(settlements))

		assert.True(t, settlements[0].NetAmountUSD.LessThan(decimal.Zero))
		assert.True(t, settlements[0].NetAmountUSD.Equal(decimal.NewFromFloat(-150.00)))
	})

	t.Run("empty transaction list", func(t *testing.T) {
		fxProvider := fxrate.NewMockProvider()
		fxService := fxrate.NewService(fxProvider)
		engine := NewEngine(fxService)

		settlements, err := engine.Calculate([]*domain.Transaction{})
		require.NoError(t, err)
		assert.Equal(t, 0, len(settlements))
	})

	t.Run("single transaction per supplier", func(t *testing.T) {
		csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup001,capture,100.00,USD,2024-01-15T10:30:00Z,completed
tx002,sup002,capture,200.00,USD,2024-01-16T14:20:00Z,completed
tx003,sup003,capture,300.00,USD,2024-01-17T11:45:00Z,completed`

		csvReader := processor.NewCSVReader()
		transactions, err := csvReader.Read(strings.NewReader(csvData))
		require.NoError(t, err)

		fxProvider := fxrate.NewMockProvider()
		fxService := fxrate.NewService(fxProvider)
		engine := NewEngine(fxService)

		settlements, err := engine.Calculate(transactions)
		require.NoError(t, err)
		assert.Equal(t, 3, len(settlements))

		for _, s := range settlements {
			assert.Equal(t, 1, s.TransactionCount)
		}
	})
}

// TestMultiCurrencyConversion tests FX conversion accuracy
func TestMultiCurrencyConversion(t *testing.T) {
	fxProvider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(fxProvider)
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
		{
			ID:             "tx004",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.MXN,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx005",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(10000),
			Currency:       domain.COP,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	settlements, err := engine.Calculate(transactions)
	require.NoError(t, err)
	require.Equal(t, 1, len(settlements))

	settlement := settlements[0]
	assert.Equal(t, 5, settlement.TransactionCount)

	// All currencies should be converted to USD
	// USD should have the highest USD value (1:1)
	// Total should be greater than USD amount alone
	assert.True(t, settlement.TotalCapturesUSD.GreaterThan(decimal.NewFromFloat(100)))
}

// TestDateSpecificRates tests that different dates produce different rates
func TestDateSpecificRates(t *testing.T) {
	fxProvider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(fxProvider)
	engine := NewEngine(fxService)

	day1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)

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

	// Verify different FX rates for different dates
	rate1 := settlement.Lines[0].FXRate
	rate2 := settlement.Lines[1].FXRate
	assert.False(t, rate1.Equal(rate2), "different dates should have different FX rates")
}

// TestComplexScenario tests a realistic complex scenario
func TestComplexScenario(t *testing.T) {
	csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,amazon,capture,1000.00,USD,2024-01-15T10:30:00Z,completed
tx002,amazon,capture,500.00,USD,2024-01-16T11:00:00Z,completed
tx003,amazon,refund,100.00,USD,2024-01-17T14:30:00Z,completed
tx004,mercadolibre,capture,5000.00,BRL,2024-01-15T09:15:00Z,completed
tx005,mercadolibre,capture,3000.00,BRL,2024-01-18T16:45:00Z,completed
tx006,mercadolibre,refund,500.00,BRL,2024-01-19T10:20:00Z,completed
tx007,rappi,capture,100000.00,COP,2024-01-15T08:00:00Z,completed
tx008,rappi,capture,50000.00,COP,2024-01-20T12:30:00Z,completed
tx009,rappi,refund,10000.00,COP,2024-01-21T15:45:00Z,completed
tx010,despegar,capture,50000.00,ARS,2024-01-15T13:20:00Z,completed
tx011,despegar,capture,75000.00,ARS,2024-01-22T11:10:00Z,completed`

	// Parse
	csvReader := processor.NewCSVReader()
	transactions, err := csvReader.Read(strings.NewReader(csvData))
	require.NoError(t, err)
	require.Equal(t, 11, len(transactions))

	// Validate
	validator := processor.NewValidator()
	err = validator.ValidateBatch(transactions)
	require.NoError(t, err)

	// Calculate
	fxProvider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(fxProvider)
	engine := NewEngine(fxService)

	settlements, err := engine.Calculate(transactions)
	require.NoError(t, err)
	require.Equal(t, 4, len(settlements))

	// Verify each supplier has correct transaction count
	supplierTxCounts := make(map[string]int)
	for _, s := range settlements {
		supplierTxCounts[s.SupplierID] = s.TransactionCount
	}

	assert.Equal(t, 3, supplierTxCounts["amazon"])
	assert.Equal(t, 3, supplierTxCounts["mercadolibre"])
	assert.Equal(t, 3, supplierTxCounts["rappi"])
	assert.Equal(t, 2, supplierTxCounts["despegar"])

	// Generate report
	csvWriter := reporter.NewCSVWriter()
	var output strings.Builder
	err = csvWriter.Write(&output, settlements)
	require.NoError(t, err)

	// Verify report contains all suppliers
	outputStr := output.String()
	assert.Contains(t, outputStr, "amazon")
	assert.Contains(t, outputStr, "mercadolibre")
	assert.Contains(t, outputStr, "rappi")
	assert.Contains(t, outputStr, "despegar")

	// Count SUMMARY rows (should be 4, one per supplier)
	summaryCount := strings.Count(outputStr, "SUMMARY")
	assert.Equal(t, 4, summaryCount)
}
