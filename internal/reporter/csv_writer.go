package reporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/ignacio/solara-settlement/internal/domain"
)

// CSVWriter formats and writes settlement reports to CSV files.
// It generates two types of rows for each supplier:
//   - Detail rows: One per transaction, showing the FX conversion details
//   - Summary rows: One per supplier, showing aggregated totals
//
// The CSV output includes columns for transaction details, FX rates, converted amounts,
// and settlement totals, making it easy to import into spreadsheets or financial systems.
type CSVWriter struct{}

// NewCSVWriter creates a new CSV writer for settlement reports.
func NewCSVWriter() *CSVWriter {
	return &CSVWriter{}
}

// WriteFile writes settlement reports to a CSV file at the specified path.
// The file will be created or overwritten if it already exists.
// Returns an error if the file cannot be created or if writing fails.
func (w *CSVWriter) WriteFile(filePath string, settlements []*domain.SupplierSettlement) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.Write(file, settlements)
}

// Write writes settlement reports to an io.Writer in CSV format.
// This method is useful for testing or writing to non-file destinations.
// Returns an error if writing fails.
func (w *CSVWriter) Write(writer io.Writer, settlements []*domain.SupplierSettlement) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"supplier_id",
		"supplier_name",
		"transaction_id",
		"type",
		"timestamp",
		"original_amount",
		"original_currency",
		"fx_rate",
		"usd_amount",
		"total_captures_usd",
		"total_refunds_usd",
		"net_amount_usd",
		"transaction_count",
	}

	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write settlements
	for _, settlement := range settlements {
		if err := w.writeSettlement(csvWriter, settlement); err != nil {
			return err
		}
	}

	return nil
}

// writeSettlement writes a single supplier settlement
func (w *CSVWriter) writeSettlement(csvWriter *csv.Writer, settlement *domain.SupplierSettlement) error {
	// Write detail rows for each transaction
	for _, line := range settlement.Lines {
		record := []string{
			settlement.SupplierID,
			settlement.SupplierName,
			line.Transaction.ID,
			string(line.Transaction.Type),
			line.Transaction.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			line.Transaction.OriginalAmount.StringFixed(2),
			line.Transaction.Currency.String(),
			line.FXRate.StringFixed(6),
			line.USDAmount.StringFixed(2),
			"", // Empty for detail rows
			"", // Empty for detail rows
			"", // Empty for detail rows
			"", // Empty for detail rows
		}

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write detail row: %w", err)
		}
	}

	// Write summary row
	summaryRecord := []string{
		settlement.SupplierID,
		settlement.SupplierName,
		"",      // No transaction ID for summary
		"SUMMARY", // Type indicates summary row
		"",      // No timestamp for summary
		"",      // No original amount for summary
		"",      // No currency for summary
		"",      // No FX rate for summary
		"",      // No individual USD amount for summary
		settlement.TotalCapturesUSD.StringFixed(2),
		settlement.TotalRefundsUSD.StringFixed(2),
		settlement.NetAmountUSD.StringFixed(2),
		fmt.Sprintf("%d", settlement.TransactionCount),
	}

	if err := csvWriter.Write(summaryRecord); err != nil {
		return fmt.Errorf("failed to write summary row: %w", err)
	}

	return nil
}
