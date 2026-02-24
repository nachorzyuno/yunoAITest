# Test Data

This directory contains sample data files for testing and demonstrating the FX-Aware Settlement Engine.

## Files

### transactions.csv
Generated test data with 250+ transactions including:
- Multiple suppliers (SUP-001 through SUP-010)
- All supported currencies (ARS, BRL, COP, MXN)
- Various transaction types (authorization, capture, refund)
- Different statuses (completed, pending, failed)
- Realistic amounts and time distribution
- Edge cases (zero amounts, large amounts, etc.)

### sample_settlement.csv
Example output showing the expected settlement report format with:
- Detail rows for each processed transaction
- Summary rows with totals per supplier
- FX rate application and USD conversion
- Transaction counts and net settlement amounts

## Generating Test Data

To regenerate the transactions.csv file:

```bash
go run scripts/generate_testdata.go --output testdata/transactions.csv
```

The generator creates realistic test data with:
- Configurable number of transactions (default: 250)
- Random but deterministic data (seeded for reproducibility)
- Balanced distribution across:
  - Suppliers (10 suppliers)
  - Currencies (ARS: 25%, BRL: 30%, COP: 25%, MXN: 20%)
  - Transaction types (80% captures, 20% refunds)
  - Statuses (85% completed, 10% pending, 5% failed)
- Time range: Last 90 days
- Edge cases included automatically

## Running a Test

Process the test data:

```bash
# Using go run
go run cmd/settlement/main.go \
  --input testdata/transactions.csv \
  --output testdata/output_settlements.csv

# Or using make
make run
```

## Expected Output Format

The settlement report contains two types of rows:

### Detail Row Example
```csv
SUP-001,DETAIL,TXN-001,capture,5000.00,BRL,0.2045,1022.50,2024-01-15T10:30:00Z,,
```

Columns:
- `supplier_id`: SUP-001
- `type`: DETAIL
- `transaction_id`: TXN-001
- `transaction_type`: capture
- `original_amount`: 5000.00
- `currency`: BRL
- `fx_rate`: 0.2045
- `usd_amount`: 1022.50
- `timestamp`: 2024-01-15T10:30:00Z

### Summary Row Example
```csv
SUP-001,SUMMARY,,,,,,,,1636.00,3
```

Columns:
- `supplier_id`: SUP-001
- `type`: SUMMARY
- `total_usd`: 1636.00
- `transaction_count`: 3

## Understanding the Test Data

### Currency Distribution
- **ARS** (Argentine Peso): ~25% of transactions
- **BRL** (Brazilian Real): ~30% of transactions
- **COP** (Colombian Peso): ~25% of transactions
- **MXN** (Mexican Peso): ~20% of transactions

### Transaction Types
- **Capture**: 80% - Money coming in
- **Refund**: 20% - Money going back
- **Authorization**: Tracked but not settled (not in final report)

### Status Distribution
- **Completed**: 85% - Included in settlement
- **Pending**: 10% - Excluded from settlement
- **Failed**: 5% - Excluded from settlement

### Typical FX Rates (Base Rates with ±2% Daily Volatility)
- **ARS**: ~0.0012 per USD
- **BRL**: ~0.20 per USD
- **COP**: ~0.00025 per USD
- **MXN**: ~0.055 per USD

Example calculation:
```
Transaction: 5000 BRL capture on 2024-01-15
FX Rate: 0.2045 (base rate of 0.20 with daily volatility)
USD Amount: 5000 × 0.2045 = 1022.50 USD
```

## Validating Output

After running the settlement engine, validate the output:

1. **Check supplier grouping**: All transactions for a supplier should be together
2. **Verify detail rows**: Each completed capture/refund should have a detail row
3. **Validate summary rows**: Total should equal sum of detail row USD amounts
4. **Check transaction counts**: Should match number of detail rows per supplier
5. **Verify FX rates**: Should be within ±2% of base rate
6. **Confirm decimal precision**: No floating-point rounding errors

## Edge Cases in Test Data

The generated test data includes:
- **Small amounts**: Testing decimal precision
- **Large amounts**: Testing number formatting
- **Refunds**: Testing negative amounts in settlement
- **Same-day transactions**: Testing date-based grouping
- **Multiple currencies per supplier**: Testing currency mixing
- **Failed transactions**: Should be excluded
- **Pending transactions**: Should be excluded
- **Authorizations**: Should be excluded from settlement

## Manual Testing Scenarios

### Scenario 1: Single Currency, Single Supplier
```csv
transaction_id,supplier_id,type,original_amount,currency,timestamp,status
TXN-001,SUP-999,capture,1000.00,BRL,2024-01-15T10:00:00Z,completed
```

Expected output:
- 1 detail row with USD conversion
- 1 summary row with total

### Scenario 2: Multiple Currencies, Single Supplier
```csv
transaction_id,supplier_id,type,original_amount,currency,timestamp,status
TXN-001,SUP-999,capture,1000.00,BRL,2024-01-15T10:00:00Z,completed
TXN-002,SUP-999,capture,100000.00,ARS,2024-01-15T11:00:00Z,completed
```

Expected output:
- 2 detail rows with different FX rates
- 1 summary row with combined total

### Scenario 3: Capture and Refund
```csv
transaction_id,supplier_id,type,original_amount,currency,timestamp,status
TXN-001,SUP-999,capture,1000.00,BRL,2024-01-15T10:00:00Z,completed
TXN-002,SUP-999,refund,-200.00,BRL,2024-01-16T10:00:00Z,completed
```

Expected output:
- 2 detail rows (one positive, one negative)
- 1 summary row with net amount (capture - refund)

## Troubleshooting

### "No transactions found"
- Check that the input file exists
- Verify CSV format is correct (header row + data rows)
- Ensure there are completed capture/refund transactions

### "Invalid currency"
- Supported currencies: ARS, BRL, COP, MXN
- Check for typos or extra spaces
- Currency codes must be uppercase

### "Invalid timestamp"
- Must be in RFC3339 format: `2024-01-15T10:30:00Z`
- Must include timezone (Z for UTC or offset like +00:00)
- Cannot be in the future

### Decimal precision issues
- Amounts should use decimal notation: `1000.50`
- No currency symbols: `$1000` is invalid
- No thousand separators: `1,000.00` is invalid
- Use plain numbers: `1000.50` is correct
