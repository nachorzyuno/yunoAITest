# Quick Start Guide

Get up and running with the FX-Aware Settlement Engine in 5 minutes.

## Prerequisites

- Go 1.24.5 or higher installed
- Git
- Terminal/command line access

## Installation (2 minutes)

```bash
# Clone the repository
git clone https://github.com/ignacio/solara-settlement.git
cd yunoAITest

# Install dependencies
go mod download

# Verify installation
go test ./...
```

If all tests pass, you're ready to go!

## Your First Settlement (3 minutes)

### Step 1: Generate Test Data

```bash
make generate-data
```

This creates `testdata/transactions.csv` with 250+ sample transactions.

### Step 2: Run the Settlement Engine

```bash
make run
```

This processes the transactions and creates `settlements.csv`.

### Step 3: View the Results

Open `settlements.csv` in your favorite spreadsheet application or view in terminal:

```bash
head -20 settlements.csv
```

You'll see:
- Detail rows showing each transaction with FX conversion
- Summary rows showing total settlement per supplier

## Common Commands

```bash
# Run everything (format, lint, test, build)
make all

# Just run tests
make test

# Build the binary
make build

# Run with custom files
./settlement --input my_transactions.csv --output my_settlements.csv

# Clean up generated files
make clean

# View help
make help
```

## Understanding the Output

### Detail Row
```csv
SUP-001,DETAIL,TXN-123,capture,5000.00,BRL,0.2045,1022.50,2024-01-15T10:30:00Z,,
```

Reads as: "Supplier SUP-001 captured 5000.00 BRL (at rate 0.2045) = 1022.50 USD"

### Summary Row
```csv
SUP-001,SUMMARY,,,,,,,,1636.00,3
```

Reads as: "Supplier SUP-001 total settlement: 1636.00 USD from 3 transactions"

## Next Steps

- **Custom Data**: Create your own `transactions.csv` following the [format guide](README.md#input-csv-format)
- **Explore Code**: Check out the [architecture section](README.md#architecture)
- **Contributing**: Read [CONTRIBUTING.md](CONTRIBUTING.md) to make changes
- **Deep Dive**: Review the [full documentation](README.md)

## Quick Troubleshooting

### Build fails
```bash
# Clean and retry
make clean
go mod tidy
make build
```

### Tests fail
```bash
# Run with verbose output to see details
go test -v ./...
```

### Import errors
```bash
# Ensure you're in the project root
cd yunoAITest
go mod download
```

### CSV parsing errors
- Check that your CSV has the correct header row
- Verify timestamps are in RFC3339 format
- Ensure amounts are valid decimal numbers (no currency symbols)

## Example: Process Your Own Data

1. Create `my_transactions.csv`:
```csv
transaction_id,supplier_id,type,original_amount,currency,timestamp,status
TXN-001,MY-SUPPLIER,capture,1000.00,BRL,2024-01-15T10:00:00Z,completed
TXN-002,MY-SUPPLIER,refund,-100.00,BRL,2024-01-16T11:00:00Z,completed
```

2. Run the engine:
```bash
./settlement --input my_transactions.csv --output my_settlements.csv
```

3. Check the results:
```bash
cat my_settlements.csv
```

Expected output:
```csv
supplier_id,type,transaction_id,transaction_type,original_amount,currency,fx_rate,usd_amount,timestamp,total_usd,transaction_count
MY-SUPPLIER,DETAIL,TXN-001,capture,1000.00,BRL,0.2045,204.50,2024-01-15T10:00:00Z,,
MY-SUPPLIER,DETAIL,TXN-002,refund,-100.00,BRL,0.2045,-20.45,2024-01-16T11:00:00Z,,
MY-SUPPLIER,SUMMARY,,,,,,,,184.05,2
```

## That's It!

You've successfully:
- Installed the settlement engine
- Generated test data
- Processed settlements
- Viewed the results

Ready to learn more? Check out the [full README](README.md).

---

**Questions?** Open an issue on GitHub or check the [troubleshooting guide](testdata/README.md#troubleshooting).
