// Package processor handles CSV transaction data ingestion and validation.
//
// This package provides functionality to:
//   - Parse CSV files containing transaction data
//   - Validate transaction fields (IDs, amounts, currencies, timestamps)
//   - Transform CSV rows into domain.Transaction entities
//   - Handle data quality issues and provide clear error messages
//
// The processor ensures that only valid, well-formed transaction data
// enters the settlement calculation pipeline. It validates:
//   - Required fields are present and non-empty
//   - Amounts are valid decimal numbers
//   - Currencies are supported (ARS, BRL, COP, MXN)
//   - Timestamps are in RFC3339 format and not in the future
//   - Transaction types and statuses are valid enum values
//
// Usage:
//
//	processor := processor.NewCSVProcessor()
//	transactions, err := processor.ProcessFile("transactions.csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
package processor
