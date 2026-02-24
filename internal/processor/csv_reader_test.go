package processor

import (
	"strings"
	"testing"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCSVReader(t *testing.T) {
	reader := NewCSVReader()
	assert.NotNil(t, reader)
	assert.Equal(t, 7, len(reader.expectedHeaders))
}

func TestCSVReader_Read_ValidData(t *testing.T) {
	csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup123,capture,100.50,USD,2024-01-15T10:30:00Z,completed
tx002,sup456,refund,50.25,BRL,2024-01-16T14:20:00Z,completed`

	reader := NewCSVReader()
	transactions, err := reader.Read(strings.NewReader(csvData))

	require.NoError(t, err)
	require.Equal(t, 2, len(transactions))

	// Verify first transaction
	tx1 := transactions[0]
	assert.Equal(t, "tx001", tx1.ID)
	assert.Equal(t, "sup123", tx1.SupplierID)
	assert.Equal(t, domain.Capture, tx1.Type)
	assert.True(t, tx1.OriginalAmount.Equal(decimal.NewFromFloat(100.50)))
	assert.Equal(t, domain.USD, tx1.Currency)
	assert.Equal(t, domain.Completed, tx1.Status)

	// Verify second transaction
	tx2 := transactions[1]
	assert.Equal(t, "tx002", tx2.ID)
	assert.Equal(t, "sup456", tx2.SupplierID)
	assert.Equal(t, domain.Refund, tx2.Type)
	assert.True(t, tx2.OriginalAmount.Equal(decimal.NewFromFloat(50.25)))
	assert.Equal(t, domain.BRL, tx2.Currency)
	assert.Equal(t, domain.Completed, tx2.Status)
}

func TestCSVReader_Read_InvalidHeader(t *testing.T) {
	tests := []struct {
		name    string
		csvData string
		errMsg  string
	}{
		{
			name:    "wrong column name",
			csvData: "transaction_id,supplier_id,type,amount,currency,timestamp,status",
			errMsg:  "invalid header at column 4",
		},
		{
			name:    "missing column",
			csvData: "transaction_id,supplier_id,type,original_amount,currency,timestamp",
			errMsg:  "expected 7 columns",
		},
		{
			name:    "extra column",
			csvData: "transaction_id,supplier_id,type,original_amount,currency,timestamp,status,extra",
			errMsg:  "expected 7 columns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewCSVReader()
			_, err := reader.Read(strings.NewReader(tt.csvData))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestCSVReader_Read_InvalidAmount(t *testing.T) {
	csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup123,capture,invalid,USD,2024-01-15T10:30:00Z,completed`

	reader := NewCSVReader()
	_, err := reader.Read(strings.NewReader(csvData))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid amount")
}

func TestCSVReader_Read_InvalidTimestamp(t *testing.T) {
	csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup123,capture,100.50,USD,2024-01-15,completed`

	reader := NewCSVReader()
	_, err := reader.Read(strings.NewReader(csvData))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid timestamp")
}

func TestCSVReader_Read_MissingColumns(t *testing.T) {
	csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup123,capture,100.50,USD`

	reader := NewCSVReader()
	_, err := reader.Read(strings.NewReader(csvData))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrong number of fields")
}

func TestCSVReader_Read_EmptyFile(t *testing.T) {
	csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status`

	reader := NewCSVReader()
	transactions, err := reader.Read(strings.NewReader(csvData))

	require.NoError(t, err)
	assert.Equal(t, 0, len(transactions))
}

func TestCSVReader_Read_MultipleTransactions(t *testing.T) {
	csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup123,capture,100.50,USD,2024-01-15T10:30:00Z,completed
tx002,sup123,refund,20.00,USD,2024-01-16T14:20:00Z,completed
tx003,sup456,capture,5000.00,ARS,2024-01-17T09:15:00Z,completed
tx004,sup456,capture,200.00,BRL,2024-01-18T11:45:00Z,pending
tx005,sup789,refund,100.00,MXN,2024-01-19T16:30:00Z,failed`

	reader := NewCSVReader()
	transactions, err := reader.Read(strings.NewReader(csvData))

	require.NoError(t, err)
	require.Equal(t, 5, len(transactions))

	// Verify different currencies
	assert.Equal(t, domain.USD, transactions[0].Currency)
	assert.Equal(t, domain.ARS, transactions[2].Currency)
	assert.Equal(t, domain.BRL, transactions[3].Currency)
	assert.Equal(t, domain.MXN, transactions[4].Currency)

	// Verify different statuses
	assert.Equal(t, domain.Completed, transactions[0].Status)
	assert.Equal(t, domain.Pending, transactions[3].Status)
	assert.Equal(t, domain.Failed, transactions[4].Status)
}

func TestCSVReader_Read_DecimalPrecision(t *testing.T) {
	csvData := `transaction_id,supplier_id,type,original_amount,currency,timestamp,status
tx001,sup123,capture,100.123456789,USD,2024-01-15T10:30:00Z,completed`

	reader := NewCSVReader()
	transactions, err := reader.Read(strings.NewReader(csvData))

	require.NoError(t, err)
	require.Equal(t, 1, len(transactions))

	// Verify decimal precision is preserved
	expected, _ := decimal.NewFromString("100.123456789")
	assert.True(t, transactions[0].OriginalAmount.Equal(expected))
}
