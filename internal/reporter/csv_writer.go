package reporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/ignacio/solara-settlement/internal/domain"
)

// CSVWriter writes settlement reports to CSV format
type CSVWriter struct{}

// NewCSVWriter creates a new CSV writer
func NewCSVWriter() *CSVWriter {
	return &CSVWriter{}
}

// WriteFile writes settlements to a CSV file
func (w *CSVWriter) WriteFile(filePath string, settlements []*domain.SupplierSettlement) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.Write(file, settlements)
}

// Write writes settlements to an io.Writer
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
