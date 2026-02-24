package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCSVWriter(t *testing.T) {
	writer := NewCSVWriter()
	assert.NotNil(t, writer)
}

func TestCSVWriter_Write_SingleSupplier(t *testing.T) {
	writer := NewCSVWriter()

	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.50),
		Currency:       domain.USD,
		Timestamp:      validTime,
		Status:         domain.Completed,
	}

	settlement := domain.NewSupplierSettlement("sup123", "Test Supplier")
	settlement.AddLine(domain.SettlementLine{
		Transaction: tx,
		FXRate:      decimal.NewFromFloat(1.0),
		USDAmount:   decimal.NewFromFloat(100.50),
	})

	var buf bytes.Buffer
	err := writer.Write(&buf, []*domain.SupplierSettlement{settlement})
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header + detail row + summary row
	assert.Equal(t, 3, len(lines))

	// Check header
	assert.Contains(t, lines[0], "supplier_id")
	assert.Contains(t, lines[0], "transaction_id")
	assert.Contains(t, lines[0], "type")

	// Check detail row
	assert.Contains(t, lines[1], "sup123")
	assert.Contains(t, lines[1], "Test Supplier")
	assert.Contains(t, lines[1], "tx001")
	assert.Contains(t, lines[1], "capture")
	assert.Contains(t, lines[1], "100.50")

	// Check summary row
	assert.Contains(t, lines[2], "sup123")
	assert.Contains(t, lines[2], "SUMMARY")
	assert.Contains(t, lines[2], "100.50") // total_captures_usd
}

func TestCSVWriter_Write_MultipleTransactions(t *testing.T) {
	writer := NewCSVWriter()

	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tx1 := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.00),
		Currency:       domain.USD,
		Timestamp:      validTime,
		Status:         domain.Completed,
	}

	tx2 := &domain.Transaction{
		ID:             "tx002",
		SupplierID:     "sup123",
		Type:           domain.Refund,
		OriginalAmount: decimal.NewFromFloat(30.00),
		Currency:       domain.USD,
		Timestamp:      validTime,
		Status:         domain.Completed,
	}

	settlement := domain.NewSupplierSettlement("sup123", "Test Supplier")
	settlement.AddLine(domain.SettlementLine{
		Transaction: tx1,
		FXRate:      decimal.NewFromFloat(1.0),
		USDAmount:   decimal.NewFromFloat(100.00),
	})
	settlement.AddLine(domain.SettlementLine{
		Transaction: tx2,
		FXRate:      decimal.NewFromFloat(1.0),
		USDAmount:   decimal.NewFromFloat(30.00),
	})

	var buf bytes.Buffer
	err := writer.Write(&buf, []*domain.SupplierSettlement{settlement})
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header + 2 detail rows + summary row
	assert.Equal(t, 4, len(lines))

	// Check summary row has correct totals
	summaryLine := lines[3]
	assert.Contains(t, summaryLine, "SUMMARY")
	assert.Contains(t, summaryLine, "100.00") // total_captures_usd
	assert.Contains(t, summaryLine, "30.00")  // total_refunds_usd
	assert.Contains(t, summaryLine, "70.00")  // net_amount_usd
	assert.Contains(t, summaryLine, "2")      // transaction_count
}

func TestCSVWriter_Write_MultipleSuppliers(t *testing.T) {
	writer := NewCSVWriter()

	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	// First supplier
	tx1 := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.00),
		Currency:       domain.USD,
		Timestamp:      validTime,
		Status:         domain.Completed,
	}

	settlement1 := domain.NewSupplierSettlement("sup123", "Supplier A")
	settlement1.AddLine(domain.SettlementLine{
		Transaction: tx1,
		FXRate:      decimal.NewFromFloat(1.0),
		USDAmount:   decimal.NewFromFloat(100.00),
	})

	// Second supplier
	tx2 := &domain.Transaction{
		ID:             "tx002",
		SupplierID:     "sup456",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(200.00),
		Currency:       domain.USD,
		Timestamp:      validTime,
		Status:         domain.Completed,
	}

	settlement2 := domain.NewSupplierSettlement("sup456", "Supplier B")
	settlement2.AddLine(domain.SettlementLine{
		Transaction: tx2,
		FXRate:      decimal.NewFromFloat(1.0),
		USDAmount:   decimal.NewFromFloat(200.00),
	})

	var buf bytes.Buffer
	err := writer.Write(&buf, []*domain.SupplierSettlement{settlement1, settlement2})
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header + (detail + summary) * 2 suppliers
	assert.Equal(t, 5, len(lines))

	// Check both suppliers are present
	assert.Contains(t, output, "sup123")
	assert.Contains(t, output, "sup456")
	assert.Contains(t, output, "Supplier A")
	assert.Contains(t, output, "Supplier B")
}

func TestCSVWriter_Write_WithFXConversion(t *testing.T) {
	writer := NewCSVWriter()

	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.00),
		Currency:       domain.BRL,
		Timestamp:      validTime,
		Status:         domain.Completed,
	}

	settlement := domain.NewSupplierSettlement("sup123", "Test Supplier")
	settlement.AddLine(domain.SettlementLine{
		Transaction: tx,
		FXRate:      decimal.NewFromFloat(0.20),
		USDAmount:   decimal.NewFromFloat(20.00),
	})

	var buf bytes.Buffer
	err := writer.Write(&buf, []*domain.SupplierSettlement{settlement})
	require.NoError(t, err)

	output := buf.String()

	// Check FX rate and conversion
	assert.Contains(t, output, "BRL")
	assert.Contains(t, output, "0.200000") // FX rate with 6 decimals
	assert.Contains(t, output, "20.00")    // USD amount
}

func TestCSVWriter_Write_EmptySettlements(t *testing.T) {
	writer := NewCSVWriter()

	var buf bytes.Buffer
	err := writer.Write(&buf, []*domain.SupplierSettlement{})
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should only have header
	assert.Equal(t, 1, len(lines))
	assert.Contains(t, lines[0], "supplier_id")
}

func TestCSVWriter_Write_DecimalPrecision(t *testing.T) {
	writer := NewCSVWriter()

	validTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(123.456789),
		Currency:       domain.USD,
		Timestamp:      validTime,
		Status:         domain.Completed,
	}

	settlement := domain.NewSupplierSettlement("sup123", "Test Supplier")
	settlement.AddLine(domain.SettlementLine{
		Transaction: tx,
		FXRate:      decimal.NewFromFloat(1.0),
		USDAmount:   decimal.NewFromFloat(123.456789),
	})

	var buf bytes.Buffer
	err := writer.Write(&buf, []*domain.SupplierSettlement{settlement})
	require.NoError(t, err)

	output := buf.String()

	// Amounts should be fixed to 2 decimal places
	assert.Contains(t, output, "123.46") // Rounded
}

func TestCSVWriter_Write_TimestampFormat(t *testing.T) {
	writer := NewCSVWriter()

	validTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.00),
		Currency:       domain.USD,
		Timestamp:      validTime,
		Status:         domain.Completed,
	}

	settlement := domain.NewSupplierSettlement("sup123", "Test Supplier")
	settlement.AddLine(domain.SettlementLine{
		Transaction: tx,
		FXRate:      decimal.NewFromFloat(1.0),
		USDAmount:   decimal.NewFromFloat(100.00),
	})

	var buf bytes.Buffer
	err := writer.Write(&buf, []*domain.SupplierSettlement{settlement})
	require.NoError(t, err)

	output := buf.String()

	// Check RFC3339 format
	assert.Contains(t, output, "2024-01-15T10:30:45Z")
}
