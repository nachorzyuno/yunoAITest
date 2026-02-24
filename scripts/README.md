# Test Data Generator

## Overview

The `generate_testdata.go` script creates realistic transaction data for testing the settlement engine. It generates 478 transactions across 7 suppliers over a 30-day period (January 1-30, 2024).

## Usage

```bash
# Generate with default settings
go run scripts/generate_testdata.go

# Specify custom output path
go run scripts/generate_testdata.go --output testdata/custom_transactions.csv

# Use different random seed for different data
go run scripts/generate_testdata.go --seed 12345
```

## Generated Data Summary

### Total Transactions: 478

**Transaction Types:**
- Authorizations: 248
- Captures: 203 (85% of successful authorizations)
- Refunds: 27 (10% of captures, with SUP007 >50%)

**Transaction Status:**
- Completed: 453
- Failed: 11 (failed authorizations)
- Pending: 14 (uncaptured authorizations)

**Currency Distribution:**
- BRL: 214 transactions (44.8%)
- MXN: 159 transactions (33.3%)
- ARS: 53 transactions (11.1%)
- COP: 52 transactions (10.9%)

### Suppliers

| Supplier ID | Name | Transactions | Captures | Refunds | Refund Rate | Currencies |
|-------------|------|--------------|----------|---------|-------------|------------|
| SUP001 | Hotel Marriott Buenos Aires | 115 | 49 | 6 | 12.2% | ARS, BRL, MXN |
| SUP002 | Airline LATAM | 102 | 42 | 5 | 11.9% | BRL |
| SUP003 | Car Rental Hertz Mexico | 75 | 32 | 3 | 9.4% | MXN |
| SUP004 | Hotel Copacabana Rio | 66 | 26 | 5 | 19.2% | BRL |
| SUP005 | Tour Operator Colombia | 52 | 22 | 5 | 22.7% | COP |
| SUP006 | Beach Resort Cancun | 61 | 29 | 2 | 6.9% | MXN |
| SUP007 | Hostel Palermo | 7 | 3 | 1 | 33.3% | ARS |

## Edge Cases Included

1. **SUP007 (Hostel Palermo)**: Very low volume (only 3 authorizations) with HIGH refund rate (33.3%, targeting >50% but random variance applies)

2. **SUP001 (Hotel Marriott)**: Multi-currency transactions across ARS (46), BRL (46), and MXN (23)

3. **SUP002 (Airline LATAM)**: Single currency focus - 100% BRL transactions (102 total)

4. **Failed Authorizations**: ~4.4% of authorizations fail (11 failed out of 248 total authorizations)

5. **Pending Authorizations**: ~5.6% remain pending/uncaptured (14 out of 248)

6. **Realistic Transaction Flow**:
   - Authorization created first
   - Capture occurs 0-2 days after authorization
   - Refunds occur 3-7 days after capture
   - All transactions sorted chronologically

## Transaction Amount Ranges

| Currency | Min Amount | Max Amount | Use Case |
|----------|------------|------------|----------|
| ARS | 10,000 | 500,000 | Hotel rooms, flights |
| BRL | 500 | 15,000 | Tours, rentals |
| COP | 100,000 | 5,000,000 | Large bookings |
| MXN | 1,000 | 40,000 | Various services |

## CSV Format

The generated CSV file has the following columns:

```csv
transaction_id,supplier_id,type,original_amount,currency,timestamp,status
```

Example:
```csv
TXN001,SUP001,authorization,50000.00,ARS,2024-01-01T10:00:00Z,completed
TXN002,SUP001,capture,50000.00,ARS,2024-01-01T10:30:00Z,completed
```

**Column Details:**
- `transaction_id`: Unique identifier (TXN001, TXN002, ...)
- `supplier_id`: Supplier identifier (SUP001-SUP007)
- `type`: Transaction type (authorization, capture, refund)
- `original_amount`: Amount with 2 decimal places
- `currency`: ISO currency code (ARS, BRL, COP, MXN)
- `timestamp`: RFC3339 format (2024-01-15T10:30:00Z)
- `status`: Transaction status (completed, pending, failed)

## Reproducibility

The script uses a fixed random seed (default: 42) to ensure reproducible data generation. Running the script multiple times with the same seed will produce identical results.

To generate different data, use a different seed:
```bash
go run scripts/generate_testdata.go --seed 99
```

## Settlement Engine Integration

This data is designed to test the settlement engine which:
- Processes only "capture" and "refund" types with "completed" status
- Ignores authorizations and failed/pending transactions
- Applies FX conversion for ARS, BRL, COP, MXN → USD

Expected settlement calculation:
- Total captures (completed): 203 transactions
- Total refunds (completed): 27 transactions
- Net settlement = Captures - Refunds (per supplier, after FX conversion)
