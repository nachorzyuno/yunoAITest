# Architecture Documentation

## Overview

The FX-Aware Settlement Engine is built using a clean, layered architecture that separates concerns and promotes maintainability. This document provides a detailed explanation of the system architecture, design patterns, and key decisions.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                                │
│                    (cmd/settlement)                              │
│  • Argument parsing                                             │
│  • Configuration loading                                        │
│  • Error handling & user feedback                              │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     v
┌─────────────────────────────────────────────────────────────────┐
│                    Application Layer                             │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐           │
│  │  Processor  │→ │  Settlement  │→ │  Reporter   │           │
│  │    (CSV)    │  │    Engine    │  │    (CSV)    │           │
│  └─────────────┘  └──────────────┘  └─────────────┘           │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     v
┌─────────────────────────────────────────────────────────────────┐
│                     Domain Layer                                 │
│  • Transaction    • Currency      • Settlement                  │
│  • Supplier       • Business Rules                              │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     v
┌─────────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                            │
│  • FX Rate Provider (Mock/Real)                                 │
│  • File I/O                                                     │
│  • External APIs (future)                                       │
└─────────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. CLI Layer (`cmd/settlement`)

**Purpose**: Entry point for the application, handles user interaction.

**Responsibilities**:
- Parse command-line arguments
- Load configuration
- Orchestrate the settlement workflow
- Handle top-level errors
- Provide user feedback

**Key Files**:
- `main.go`: Application entry point

**Design Pattern**: Facade pattern - provides a simple interface to complex subsystems.

### 2. Application Layer (`internal/processor`, `internal/settlement`, `internal/reporter`)

**Purpose**: Implements the business workflow and use cases.

#### Processor (`internal/processor`)

**Responsibilities**:
- Read CSV files
- Parse transaction data
- Validate input data
- Transform CSV rows to domain entities
- Report parsing errors with context

**Key Validations**:
- Required fields present
- Amounts are valid decimals
- Currencies are supported
- Timestamps are valid RFC3339
- Transaction types/statuses are valid enums

#### Settlement Engine (`internal/settlement`)

**Responsibilities**:
- Group transactions by supplier
- Apply FX rates to convert to USD
- Calculate net settlement amounts
- Generate settlement line items
- Aggregate totals per supplier

**Business Rules**:
- Only completed captures and refunds are settled
- Pending and failed transactions are excluded
- Authorizations are tracked but not settled
- Refunds reduce the settlement total
- Each transaction uses historical FX rate from its date

#### Reporter (`internal/reporter`)

**Responsibilities**:
- Format settlement data as CSV
- Generate detail rows (one per transaction)
- Generate summary rows (one per supplier)
- Write output files
- Handle decimal formatting

**Output Format**:
- Detail rows: Full transaction breakdown with FX info
- Summary rows: Aggregated totals per supplier

### 3. Domain Layer (`internal/domain`)

**Purpose**: Defines core business entities and rules.

**Entities**:

#### Currency
```go
type Currency string

const (
    ARS Currency = "ARS" // Argentine Peso
    BRL Currency = "BRL" // Brazilian Real
    COP Currency = "COP" // Colombian Peso
    MXN Currency = "MXN" // Mexican Peso
    USD Currency = "USD" // US Dollar
)
```

**Purpose**: Type-safe currency representation with validation.

#### Transaction
```go
type Transaction struct {
    ID             string
    SupplierID     string
    Type           TransactionType
    OriginalAmount decimal.Decimal
    Currency       Currency
    Timestamp      time.Time
    Status         TransactionStatus
}
```

**Purpose**: Represents a financial transaction with all necessary attributes.

**Key Methods**:
- `Validate()`: Ensures transaction data is valid
- `IsSettleable()`: Determines if transaction should be in settlement

#### Settlement
```go
type SupplierSettlement struct {
    SupplierID       string
    Lines            []SettlementLine
    TotalCapturesUSD decimal.Decimal
    TotalRefundsUSD  decimal.Decimal
    NetAmountUSD     decimal.Decimal
    TransactionCount int
}
```

**Purpose**: Aggregates all transactions for a supplier into a settlement report.

**Design Pattern**: Aggregate pattern - maintains consistency of related entities.

### 4. Infrastructure Layer (`internal/fxrate`)

**Purpose**: Provides external services and data sources.

#### FX Rate Service
```go
type Provider interface {
    GetRate(currency Currency, date time.Time) (decimal.Decimal, error)
}
```

**Purpose**: Abstracts FX rate retrieval, allowing different implementations.

**Current Implementation**: MockProvider with simulated rates
**Future Implementations**:
- OpenExchangeRates API
- CurrencyLayer API
- Internal FX service

**Design Pattern**: Strategy pattern - allows swapping implementations.

## Data Flow

### Settlement Processing Pipeline

```
1. Read CSV File
   │
   v
2. Parse & Validate Transactions
   │
   v
3. Filter Settleable Transactions
   │ (completed captures & refunds only)
   v
4. Group by Supplier
   │
   v
5. For Each Transaction:
   │ a. Get historical FX rate
   │ b. Convert to USD
   │ c. Create settlement line
   │
   v
6. Aggregate by Supplier
   │ a. Sum captures
   │ b. Sum refunds
   │ c. Calculate net (captures - refunds)
   │
   v
7. Generate CSV Report
   │ a. Write detail rows
   │ b. Write summary rows
   │
   v
8. Output File
```

## Design Decisions

### 1. Decimal Precision

**Decision**: Use `github.com/shopspring/decimal` for all monetary values.

**Rationale**:
- Floating-point arithmetic (`float64`) introduces rounding errors
- Financial calculations require exact precision
- Decimal type maintains accuracy for monetary operations
- Meets regulatory and audit requirements

**Example Issue Avoided**:
```go
// With float64 (BAD):
var total float64 = 0.1 + 0.2  // Result: 0.30000000000000004

// With decimal (GOOD):
total := decimal.NewFromFloat(0.1).Add(decimal.NewFromFloat(0.2))  // Result: 0.3
```

### 2. CSV Format

**Decision**: Use CSV for both input and output.

**Rationale**:
- Universal format understood by all financial systems
- Easy to inspect and validate manually
- Compatible with Excel, Google Sheets, databases
- Simple to integrate with existing systems
- Human-readable for debugging

**Alternatives Considered**:
- JSON: More complex, less universal in finance
- XML: Verbose, harder to read
- Protobuf: Binary format, not human-readable
- Database: Requires additional infrastructure

### 3. Mock FX Provider

**Decision**: Implement simulated FX rates with daily volatility.

**Rationale**:
- No external API dependencies for testing
- Deterministic results for validation
- Simulates real market conditions (±2% daily variance)
- Easy to swap with real provider in production

**Implementation**:
```go
// Base rates
ARS: 0.0012 per USD
BRL: 0.20 per USD
COP: 0.00025 per USD
MXN: 0.055 per USD

// Daily volatility: ±2% based on date
rate = baseRate × (1 + sin(daysSinceEpoch) × 0.02)
```

### 4. Transaction Filtering

**Decision**: Only settle completed captures and refunds.

**Rationale**:
- **Authorizations**: Represent intent, not actual funds transfer
- **Pending**: Transaction not finalized yet
- **Failed**: Transaction didn't complete successfully
- **Captures/Refunds**: Actual money movement requiring settlement

**Business Logic**:
```go
func (t *Transaction) IsSettleable() bool {
    return (t.Type == Capture || t.Type == Refund) &&
           t.Status == Completed
}
```

### 5. Package Structure

**Decision**: Use internal packages with domain-driven design.

**Rationale**:
- `internal/`: Prevents external packages from importing
- Domain layer: Core business logic independent of infrastructure
- Clear separation of concerns
- Easy to test and maintain
- Supports future extension

**Benefits**:
- Domain logic is pure and testable
- Infrastructure can be swapped (e.g., different FX providers)
- Clear boundaries between layers
- Easy to understand for new developers

## Error Handling Strategy

### Validation Errors
- Fail fast at input validation
- Provide clear error messages with context
- Include field names and values in errors

### Processing Errors
- Continue processing other transactions on non-fatal errors
- Log errors with transaction IDs
- Report errors at end of processing

### Fatal Errors
- Stop processing immediately
- Clean up resources
- Return clear error message to user

## Testing Strategy

### Unit Tests
- Test each package independently
- Mock dependencies (e.g., FX provider)
- Test edge cases and error conditions
- Aim for >80% coverage

### Integration Tests
- Test full pipeline (CSV → Settlement → CSV)
- Use test fixtures with known outputs
- Validate end-to-end behavior

### Test Data
- Generated programmatically for consistency
- Includes edge cases automatically
- Reproducible (seeded random generation)

## Performance Considerations

### Current Scale
- Designed for 250-10,000 transactions per run
- In-memory processing (no database required)
- O(n) time complexity for most operations

### Future Optimizations
- Stream processing for large files (>100k transactions)
- Parallel FX rate lookups
- Database-backed processing for millions of transactions
- Batch processing with checkpointing

## Extensibility Points

### 1. Add New FX Provider
```go
type RealFXProvider struct {
    apiKey string
    client *http.Client
}

func (p *RealFXProvider) GetRate(currency Currency, date time.Time) (decimal.Decimal, error) {
    // Call external API
}
```

### 2. Add New Currency
- Add to `Currency` enum
- Add validation
- Add base rate to provider
- Update documentation

### 3. Add New Report Format
- Implement new reporter (JSON, XML, etc.)
- Add CLI flag for format selection
- Maintain same data structure

### 4. Add Database Storage
- Implement repository pattern
- Add database package
- Maintain same domain entities

## Security Considerations

### Current Implementation
- No authentication/authorization (single-user CLI)
- File system access only
- No network calls (mock provider)

### Future Enhancements
- API key management for real FX provider
- Audit logging
- Input sanitization for untrusted data
- Rate limiting for external API calls

## Monitoring & Observability

### Future Additions
- Structured logging with levels
- Metrics (transactions processed, errors, latency)
- Tracing for distributed operations
- Health checks for external dependencies

## Conclusion

The architecture prioritizes:
1. **Correctness**: Financial precision with decimal arithmetic
2. **Maintainability**: Clear separation of concerns
3. **Testability**: Pure domain logic, mockable dependencies
4. **Extensibility**: Easy to add currencies, providers, formats
5. **Simplicity**: Minimal dependencies, straightforward flow

This design supports the current use case while providing a solid foundation for future enhancements.
