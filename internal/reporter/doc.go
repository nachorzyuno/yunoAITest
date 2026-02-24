// Package reporter generates CSV settlement reports from calculated settlement data.
//
// This package provides functionality to:
//   - Format settlement data into CSV output
//   - Generate detail rows for each transaction with FX conversion info
//   - Generate summary rows showing total settlements per supplier
//   - Write reports to files with proper CSV formatting
//   - Handle decimal precision formatting for monetary values
//
// The CSV report format includes two types of rows:
//
// Detail Rows (one per transaction):
//   - supplier_id: Supplier identifier
//   - type: "DETAIL"
//   - transaction_id: Original transaction ID
//   - transaction_type: capture or refund
//   - original_amount: Amount in local currency
//   - currency: Currency code (ARS, BRL, COP, MXN)
//   - fx_rate: Applied exchange rate
//   - usd_amount: Converted amount in USD
//   - timestamp: Transaction date/time
//
// Summary Rows (one per supplier):
//   - supplier_id: Supplier identifier
//   - type: "SUMMARY"
//   - total_usd: Net settlement amount in USD
//   - transaction_count: Number of transactions
//
// Usage:
//
//	reporter := reporter.NewCSVReporter()
//	err := reporter.WriteReport("settlements.csv", settlements)
//	if err != nil {
//	    log.Fatal(err)
//	}
package reporter
