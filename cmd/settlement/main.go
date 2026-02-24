package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/ignacio/solara-settlement/internal/fxrate"
	"github.com/ignacio/solara-settlement/internal/processor"
	"github.com/ignacio/solara-settlement/internal/reporter"
	"github.com/ignacio/solara-settlement/internal/settlement"
	"github.com/shopspring/decimal"
)

func main() {
	// Parse command-line flags
	inputPath := flag.String("input", "", "Path to input CSV file (required)")
	outputPath := flag.String("output", "", "Path to output CSV file (required)")
	startDateStr := flag.String("start-date", "", "Start date for filtering (YYYY-MM-DD format, optional)")
	endDateStr := flag.String("end-date", "", "End date for filtering (YYYY-MM-DD format, optional)")
	flag.Parse()

	// Validate flags
	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if *outputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --output flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse date flags
	startDate, err := parseDateFlag(*startDateStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid start-date format: %v\n", err)
		fmt.Fprintln(os.Stderr, "Expected format: YYYY-MM-DD (e.g., 2024-01-15)")
		os.Exit(1)
	}

	endDate, err := parseDateFlag(*endDateStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid end-date format: %v\n", err)
		fmt.Fprintln(os.Stderr, "Expected format: YYYY-MM-DD (e.g., 2024-01-15)")
		os.Exit(1)
	}

	// Run the settlement process
	if err := runSettlement(*inputPath, *outputPath, startDate, endDate); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runSettlement(inputPath, outputPath string, startDate, endDate time.Time) error {
	fmt.Printf("Reading transactions from: %s\n", inputPath)

	// Initialize components
	csvReader := processor.NewCSVReader()
	validator := processor.NewValidator()
	fxProvider := fxrate.NewMockProvider()
	fxService := fxrate.NewService(fxProvider)
	engine := settlement.NewEngine(fxService)
	csvWriter := reporter.NewCSVWriter()

	// Step 1: Read transactions from CSV
	transactions, err := csvReader.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	fmt.Printf("Loaded %d transactions\n", len(transactions))

	// Step 1.5: Filter by date range (if provided)
	if !startDate.IsZero() || !endDate.IsZero() {
		fmt.Printf("Filtering transactions from %s to %s\n",
			formatDate(startDate), formatDate(endDate))
		transactions = filterByDateRange(transactions, startDate, endDate)
		fmt.Printf("Filtered to %d transactions\n", len(transactions))
	}

	// Step 2: Validate transactions
	if err := validator.ValidateBatch(transactions); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Println("Validation passed")

	// Step 3: Calculate settlements
	settlements, err := engine.Calculate(transactions)
	if err != nil {
		return fmt.Errorf("settlement calculation failed: %w", err)
	}

	fmt.Printf("Calculated settlements for %d suppliers\n", len(settlements))

	// Step 4: Generate report
	if err := csvWriter.WriteFile(outputPath, settlements); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Report written to: %s\n", outputPath)

	// Print summary statistics
	printSummary(settlements)

	return nil
}

func printSummary(settlements []*domain.SupplierSettlement) {
	fmt.Println("\n=== Settlement Summary ===")
	fmt.Printf("Total Suppliers: %d\n", len(settlements))

	totalSettled := decimal.Zero
	totalTransactions := 0

	for _, s := range settlements {
		totalSettled = totalSettled.Add(s.NetAmountUSD)
		totalTransactions += s.TransactionCount
	}

	fmt.Printf("Total Transactions Processed: %d\n", totalTransactions)
	fmt.Printf("Total Net Amount (USD): $%s\n", totalSettled.StringFixed(2))

	// Print per-supplier breakdown
	fmt.Println("\nPer-Supplier Breakdown:")
	for _, s := range settlements {
		fmt.Printf("  %s (%s): $%s (%d transactions)\n",
			s.SupplierID,
			s.SupplierName,
			s.NetAmountUSD.StringFixed(2),
			s.TransactionCount,
		)
	}

	// Print warnings summary
	printWarningsSummary(settlements)
}

// parseDateFlag parses a date string in YYYY-MM-DD format
// Returns zero time if the string is empty (optional flag not provided)
func parseDateFlag(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", dateStr)
}

// filterByDateRange filters transactions to those within the specified date range
// If startDate is zero, no lower bound is applied
// If endDate is zero, no upper bound is applied
func filterByDateRange(transactions []*domain.Transaction, startDate, endDate time.Time) []*domain.Transaction {
	if startDate.IsZero() && endDate.IsZero() {
		return transactions // No filtering
	}

	filtered := make([]*domain.Transaction, 0)
	for _, tx := range transactions {
		// Check start date boundary (inclusive)
		if !startDate.IsZero() && tx.Timestamp.Before(startDate) {
			continue
		}

		// Check end date boundary (inclusive - end of day)
		if !endDate.IsZero() {
			endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
			if tx.Timestamp.After(endOfDay) {
				continue
			}
		}

		filtered = append(filtered, tx)
	}
	return filtered
}

// formatDate formats a time.Time for display, handling zero times
func formatDate(t time.Time) string {
	if t.IsZero() {
		return "unspecified"
	}
	return t.Format("2006-01-02")
}

// printWarningsSummary prints any warnings detected during settlement processing
func printWarningsSummary(settlements []*domain.SupplierSettlement) {
	hasWarnings := false
	for _, s := range settlements {
		if len(s.Warnings) > 0 {
			if !hasWarnings {
				fmt.Println("\n⚠️  Warnings Detected:")
				hasWarnings = true
			}
			fmt.Printf("  %s (%s): %s\n",
				s.SupplierID,
				s.SupplierName,
				strings.Join(s.Warnings, ", "),
			)
		}
	}
	if hasWarnings {
		fmt.Println("\n⚠️  Review the settlement report for detailed warning information.")
	}
}
