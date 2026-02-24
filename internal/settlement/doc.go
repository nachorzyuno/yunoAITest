// Package settlement implements the core settlement calculation engine.
//
// This package provides functionality to:
//   - Group transactions by supplier ID
//   - Apply historical FX rates to convert amounts to USD
//   - Calculate net settlement amounts (captures minus refunds)
//   - Generate settlement line items for detailed reporting
//   - Aggregate totals per supplier
//
// The settlement engine processes only "settleable" transactions:
//   - Completed captures: Add to supplier's total
//   - Completed refunds: Subtract from supplier's total
//   - Pending/failed transactions: Excluded from settlement
//   - Authorizations: Tracked but not settled (intent vs. actual funds)
//
// All calculations use decimal.Decimal arithmetic to ensure financial
// precision and avoid floating-point rounding errors.
//
// Usage:
//
//	engine := settlement.NewEngine(fxService)
//	settlements, err := engine.Calculate(transactions)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, s := range settlements {
//	    fmt.Printf("Supplier %s: $%.2f USD\n", s.SupplierID, s.NetAmountUSD)
//	}
package settlement
