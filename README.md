# FX-Aware Settlement Engine

A robust, automated settlement processing system that handles multi-currency transactions with historical foreign exchange rate application. Built for the Solara Travel technical challenge to solve manual reconciliation inefficiencies.

## Overview

### The Problem

Solara Travel processes thousands of transactions daily across multiple currencies (ARS, BRL, COP, MXN). Manual reconciliation of these transactions with historical exchange rates is time-consuming, error-prone, and difficult to scale. Financial teams need an automated system that can:

- Ingest transaction data from multiple sources
- Apply accurate historical FX rates for each transaction date
- Generate settlement reports per supplier
- Ensure financial precision in all calculations

### The Solution

The FX-Aware Settlement Engine is a Golang-based system that automates the entire settlement workflow:

1. **Ingests** transaction data from CSV files with validation
2. **Converts** multi-currency amounts to USD using historical FX rates
3. **Calculates** net settlement amounts per supplier
4. **Generates** detailed CSV reports with transaction-level breakdowns

### Key Features

- **Multi-Currency Support**: Handles ARS, BRL, COP, and MXN with conversion to USD
- **Historical FX Rate Application**: Each transaction uses the FX rate from its transaction date
- **CSV Input/Output**: Easy integration with existing financial systems
- **Financial Precision**: Uses `decimal.Decimal` library to avoid floating-point errors
- **Transaction Type Support**: Processes authorizations, captures, and refunds
- **Comprehensive Validation**: Ensures data integrity before processing
- **Scalable Design**: Clean architecture supporting thousands of transactions

## Architecture

The project follows a clean, layered architecture with clear separation of concerns:

```
yunoAITest/
├── cmd/
│   └── settlement/          # CLI entry point and argument parsing
├── internal/
│   ├── domain/              # Core business entities (Transaction, Currency, Settlement)
│   ├── fxrate/              # FX rate service with mock provider
│   ├── processor/           # CSV ingestion and transaction validation
│   ├── settlement/          # Settlement calculation engine
│   └── reporter/            # CSV report generation
├── scripts/
│   └── generate_testdata.go # Test data generator (250+ transactions)
├── testdata/                # Sample CSV files and test fixtures
└── go.mod                   # Module dependencies
```

### Package Descriptions

- **`cmd/settlement`**: Command-line interface for running the settlement engine
- **`internal/domain`**: Core business models and domain logic (Transaction, Currency, Settlement entities)
- **`internal/fxrate`**: Foreign exchange rate service with provider interface and mock implementation
- **`internal/processor`**: CSV parsing, transaction validation, and data ingestion
- **`internal/settlement`**: Settlement calculation engine that aggregates transactions per supplier
- **`internal/reporter`**: CSV report generator producing settlement summaries and detail rows
- **`scripts`**: Utility scripts including realistic test data generation
- **`testdata`**: Sample input/output files for testing and demonstration

## Installation & Setup

### Prerequisites

- Go 1.24.5 or higher
- Git

### Installation Steps

```bash
# Clone the repository
git clone https://github.com/ignacio/solara-settlement.git
cd yunoAITest

# Install dependencies
go mod download

# Run tests to verify installation
go test ./...

# Build the CLI binary
go build -o settlement cmd/settlement/main.go
```

## Usage

### 1. Generate Test Data

Create realistic test data with 250+ transactions spanning multiple suppliers and currencies:

```bash
go run scripts/generate_testdata.go --output testdata/transactions.csv
```

This generates a CSV file with diverse transaction types, currencies, and edge cases.

### 2. Process Settlements

Run the settlement engine on your transaction data:

```bash
# Using go run
go run cmd/settlement/main.go \
  --input testdata/transactions.csv \
  --output settlements.csv

# Or using the built binary
./settlement --input testdata/transactions.csv --output settlements.csv
```

### 3. Review Settlement Report

The output file (`settlements.csv`) contains:
- Detail rows for each processed transaction with FX conversion
- Summary rows showing total settlement amounts per supplier

## Input CSV Format

The settlement engine expects a CSV file with the following columns:

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `transaction_id` | string | Unique transaction identifier | `TXN-001` |
| `supplier_id` | string | Supplier/merchant identifier | `SUP-123` |
| `type` | string | Transaction type: `authorization`, `capture`, or `refund` | `capture` |
| `original_amount` | decimal | Amount in local currency | `1250.50` |
| `currency` | string | Currency code: `ARS`, `BRL`, `COP`, or `MXN` | `BRL` |
| `timestamp` | RFC3339 | Transaction timestamp | `2024-01-15T10:30:00Z` |
| `status` | string | Transaction status: `pending`, `completed`, or `failed` | `completed` |

### Example Input

```csv
transaction_id,supplier_id,type,original_amount,currency,timestamp,status
TXN-001,SUP-123,capture,1250.50,BRL,2024-01-15T10:30:00Z,completed
TXN-002,SUP-123,refund,-100.00,BRL,2024-01-16T14:20:00Z,completed
TXN-003,SUP-456,capture,50000.00,ARS,2024-01-15T11:45:00Z,completed
```

### Important Notes on Input Data

- **Timestamps**: Must be in RFC3339 format with timezone information
- **Amounts**: Use decimal notation; negative amounts for refunds
- **Status**: Only `completed` transactions are included in settlements
- **Type**: `authorization` transactions are tracked but not settled (they represent intent, not final amounts)

## Output CSV Format

The settlement report contains two types of rows:

### Detail Rows

One row per processed transaction showing the FX conversion:

| Column | Description | Example |
|--------|-------------|---------|
| `supplier_id` | Supplier identifier | `SUP-123` |
| `type` | Always `DETAIL` | `DETAIL` |
| `transaction_id` | Original transaction ID | `TXN-001` |
| `transaction_type` | capture or refund | `capture` |
| `original_amount` | Amount in local currency | `1250.50` |
| `currency` | Local currency code | `BRL` |
| `fx_rate` | Applied exchange rate | `0.2045` |
| `usd_amount` | Converted USD amount | `255.78` |
| `timestamp` | Transaction date/time | `2024-01-15T10:30:00Z` |

### Summary Rows

One row per supplier showing total settlement:

| Column | Description | Example |
|--------|-------------|---------|
| `supplier_id` | Supplier identifier | `SUP-123` |
| `type` | Always `SUMMARY` | `SUMMARY` |
| `total_usd` | Net settlement in USD | `1234.56` |
| `transaction_count` | Number of transactions | `15` |

### Example Output

```csv
supplier_id,type,transaction_id,transaction_type,original_amount,currency,fx_rate,usd_amount,timestamp,total_usd,transaction_count
SUP-123,DETAIL,TXN-001,capture,1250.50,BRL,0.2045,255.78,2024-01-15T10:30:00Z,,
SUP-123,DETAIL,TXN-002,refund,-100.00,BRL,0.2045,-20.45,2024-01-16T14:20:00Z,,
SUP-123,SUMMARY,,,,,,,,235.33,2
SUP-456,DETAIL,TXN-003,capture,50000.00,ARS,0.0012,60.00,2024-01-15T11:45:00Z,,
SUP-456,SUMMARY,,,,,,,,60.00,1
```

## Design Decisions

### Mock FX Rate Provider

The current implementation uses simulated exchange rates for demonstration purposes:

- **ARS (Argentine Peso)**: 0.0012 per USD
- **BRL (Brazilian Real)**: 0.2045 per USD
- **COP (Colombian Peso)**: 0.00025 per USD
- **MXN (Mexican Peso)**: 0.0591 per USD

In production, this would be replaced with an integration to a real FX data provider (e.g., OpenExchangeRates, CurrencyLayer, or an internal service).

### CSV Format Choice

CSV was chosen for both input and output because:

- **Ubiquity**: Every financial system can produce/consume CSV
- **Simplicity**: Easy to inspect, debug, and validate
- **Integration**: Works seamlessly with Excel, Google Sheets, and data pipelines
- **Human-Readable**: Finance teams can review data without specialized tools

### Decimal Precision

The system uses the `github.com/shopspring/decimal` library instead of `float64` to:

- **Avoid rounding errors**: Floating-point arithmetic can introduce precision errors in financial calculations
- **Ensure accuracy**: Decimal arithmetic maintains exact precision for monetary values
- **Meet compliance**: Financial systems require precise calculations for auditing and regulatory compliance

### Transaction Processing Rules

1. **Only completed transactions are settled**: `pending` and `failed` transactions are excluded
2. **Authorizations are not settled**: They represent intent but not final amounts
3. **Captures and refunds are settled**: These represent actual money movement
4. **Refunds are negative**: Deducted from the supplier's settlement total
5. **Historical rates are applied**: Each transaction uses the FX rate from its transaction date

## Testing

### Run All Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage report
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Package Tests

```bash
# Test settlement calculation logic
go test ./internal/settlement/...

# Test FX rate service
go test ./internal/fxrate/...

# Test CSV processor
go test ./internal/processor/...
```

### Test Data

The `scripts/generate_testdata.go` script creates comprehensive test data including:

- 250+ transactions across multiple suppliers
- All supported currencies (ARS, BRL, COP, MXN)
- Mix of captures and refunds
- Various transaction statuses (completed, pending, failed)
- Edge cases (zero amounts, large amounts, same-day transactions)
- Time distribution across multiple months

## Challenge Criteria

This implementation addresses all requirements of the Solara Travel technical challenge:

- ✅ **Transaction Ingestion**: CSV parser with validation and error handling
- ✅ **Historical FX Rates**: Date-specific rate application per transaction
- ✅ **Settlement Calculation**: Accurate aggregation per supplier using decimal precision
- ✅ **CSV Report Generation**: Detailed and summary rows in standardized format
- ✅ **Test Data**: 250+ realistic transactions with edge cases
- ✅ **Code Quality**: Clean Go code following best practices
- ✅ **Testing**: Comprehensive unit tests for all business logic
- ✅ **Documentation**: Clear README, godoc comments, and examples

## Development

### Code Formatting

```bash
# Format all Go code
go fmt ./...

# Or use make
make fmt
```

### Linting

```bash
# Run Go vet
go vet ./...

# Or use make
make vet
```

### Building

```bash
# Build the CLI binary
go build -o settlement cmd/settlement/main.go

# Or use make
make build
```

## Contributing

Contributions are welcome! Please ensure your code follows these guidelines:

1. **Format your code**: Run `go fmt ./...` before committing
2. **Pass all tests**: Run `go test ./...` to verify functionality
3. **Run the linter**: Execute `go vet ./...` to catch common issues
4. **Add tests**: Include unit tests for new functionality
5. **Update documentation**: Keep the README and godoc comments current
6. **Use decimal types**: Always use `decimal.Decimal` for monetary calculations

## License

This project is part of the Solara Travel technical challenge.

## Contact

For questions or issues, please open a GitHub issue or contact the development team.

---

**Built with Go 1.24.5 | Uses `shopspring/decimal` for financial precision**
