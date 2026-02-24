package processor

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
)

// CSVReader reads and parses transactions from CSV files
type CSVReader struct {
	expectedHeaders []string
}

// NewCSVReader creates a new CSV reader with expected header validation
func NewCSVReader() *CSVReader {
	return &CSVReader{
		expectedHeaders: []string{
			"transaction_id",
			"supplier_id",
			"type",
			"original_amount",
			"currency",
			"timestamp",
			"status",
		},
	}
}

// ReadFile reads transactions from a CSV file
func (r *CSVReader) ReadFile(filePath string) ([]*domain.Transaction, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return r.Read(file)
}

// Read reads transactions from an io.Reader
func (r *CSVReader) Read(reader io.Reader) ([]*domain.Transaction, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Validate header
	if err := r.validateHeader(header); err != nil {
		return nil, err
	}

	// Read all records
	var transactions []*domain.Transaction
	lineNum := 2 // Start at 2 (1 is header)

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read line %d: %w", lineNum, err)
		}

		tx, err := r.parseRecord(record, lineNum)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		transactions = append(transactions, tx)
		lineNum++
	}

	return transactions, nil
}

// validateHeader checks if the CSV header matches expected columns
func (r *CSVReader) validateHeader(header []string) error {
	if len(header) != len(r.expectedHeaders) {
		return fmt.Errorf("invalid header: expected %d columns, got %d", len(r.expectedHeaders), len(header))
	}

	for i, expected := range r.expectedHeaders {
		if header[i] != expected {
			return fmt.Errorf("invalid header at column %d: expected '%s', got '%s'", i+1, expected, header[i])
		}
	}

	return nil
}

// parseRecord converts a CSV record into a Transaction
func (r *CSVReader) parseRecord(record []string, lineNum int) (*domain.Transaction, error) {
	if len(record) != len(r.expectedHeaders) {
		return nil, fmt.Errorf("invalid number of columns: expected %d, got %d", len(r.expectedHeaders), len(record))
	}

	// Parse amount
	amount, err := decimal.NewFromString(record[3])
	if err != nil {
		return nil, fmt.Errorf("invalid amount '%s': %w", record[3], err)
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, record[5])
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp '%s': must be RFC3339 format: %w", record[5], err)
	}

	tx := &domain.Transaction{
		ID:             record[0],
		SupplierID:     record[1],
		Type:           domain.TransactionType(record[2]),
		OriginalAmount: amount,
		Currency:       domain.Currency(record[4]),
		Timestamp:      timestamp,
		Status:         domain.TransactionStatus(record[6]),
	}

	return tx, nil
}
