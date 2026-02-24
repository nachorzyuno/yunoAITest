# Test Data Summary

Generated: 2024-02-24
Script: `scripts/generate_testdata.go`
Output: `testdata/transactions.csv`

## Quick Stats

- **Total Transactions:** 478
- **Date Range:** 2024-01-01 to 2024-02-06 (includes refunds up to 7 days after captures)
- **Suppliers:** 7 (SUP001 - SUP007)
- **Currencies:** 4 (ARS, BRL, COP, MXN)

## Transaction Breakdown

### By Type
```
authorization: 248 (51.9%)
capture:       203 (42.5%)
refund:         27 (5.6%)
```

### By Status
```
completed: 453 (94.8%)
pending:    14 (2.9%)
failed:     11 (2.3%)
```

### By Currency
```
BRL: 214 (44.8%) - Brazilian Real
MXN: 159 (33.3%) - Mexican Peso
ARS:  53 (11.1%) - Argentine Peso
COP:  52 (10.9%) - Colombian Peso
```

## Supplier Details

### SUP001 - Hotel Marriott Buenos Aires
- **Total Transactions:** 115
- **Captures:** 49 | **Refunds:** 6 | **Refund Rate:** 12.2%
- **Currencies:** ARS (46), BRL (46), MXN (23) - Multi-currency edge case
- **Edge Case:** Transactions across 3 different currencies

### SUP002 - Airline LATAM
- **Total Transactions:** 102
- **Captures:** 42 | **Refunds:** 5 | **Refund Rate:** 11.9%
- **Currencies:** BRL (102) - Single currency only
- **Edge Case:** 100% single currency (BRL)

### SUP003 - Car Rental Hertz Mexico
- **Total Transactions:** 75
- **Captures:** 32 | **Refunds:** 3 | **Refund Rate:** 9.4%
- **Currencies:** MXN only

### SUP004 - Hotel Copacabana Rio
- **Total Transactions:** 66
- **Captures:** 26 | **Refunds:** 5 | **Refund Rate:** 19.2%
- **Currencies:** BRL only

### SUP005 - Tour Operator Colombia
- **Total Transactions:** 52
- **Captures:** 22 | **Refunds:** 5 | **Refund Rate:** 22.7%
- **Currencies:** COP only
- **Note:** Large transaction amounts (100,000 - 5,000,000 COP)

### SUP006 - Beach Resort Cancun
- **Total Transactions:** 61
- **Captures:** 29 | **Refunds:** 2 | **Refund Rate:** 6.9%
- **Currencies:** MXN only

### SUP007 - Hostel Palermo (EDGE CASE)
- **Total Transactions:** 7
- **Captures:** 3 | **Refunds:** 1 | **Refund Rate:** 33.3%
- **Currencies:** ARS only
- **Edge Case:** Very low volume with high refund rate

## Transaction Flow Patterns

### Authorization → Capture
- **85% of successful authorizations** are captured
- Capture occurs **0-2 days** after authorization
- Example: Auth on Jan 1 → Capture on Jan 3

### Capture → Refund
- **~10% of captures** result in refunds (SUP007 has higher rate)
- Refund occurs **3-7 days** after capture
- Example: Capture on Jan 3 → Refund on Jan 9

### Failed/Pending Authorizations
- **5% failed** authorizations (authentication/fraud check failures)
- **7-8% pending** authorizations (never captured, awaiting confirmation)

## Sample Transaction Flow

**Example 1: Complete flow with refund (SUP002)**
```
TXN197: authorization, 10541.97 BRL, 2024-01-01T04:43:00Z, completed
TXN198: capture,       10541.97 BRL, 2024-01-03T06:49:00Z, completed
TXN199: refund,        10541.97 BRL, 2024-01-09T10:39:00Z, completed
```

**Example 2: Authorization without capture (SUP001)**
```
TXN061: authorization, 11921.25 BRL, 2024-01-01T03:51:00Z, pending
(No capture - remains pending)
```

**Example 3: Failed authorization (SUP002)**
```
TXN160: authorization, 936.57 BRL, 2024-01-01T16:48:00Z, failed
(No subsequent transactions)
```

## Edge Cases Included

1. **Low Volume High Refund Rate:** SUP007 has only 3 captures but 33.3% refund rate
2. **Multi-Currency:** SUP001 processes transactions in 3 currencies (ARS, BRL, MXN)
3. **Single Currency Focus:** SUP002 exclusively uses BRL (100% of transactions)
4. **Failed Authorizations:** 11 authorizations failed (2.3% failure rate)
5. **Pending Authorizations:** 14 authorizations remain pending/uncaptured
6. **Date Range:** Transactions span initial 30 days, refunds extend to Feb 6
7. **Realistic Timing:** Captures 0-2 days after auth, refunds 3-7 days after capture

## Settlement Engine Test Scenarios

### Expected Behavior
The settlement engine should:
1. **Process only:** `type=capture OR type=refund` AND `status=completed`
2. **Ignore:** All authorizations, failed transactions, pending transactions
3. **Calculate:** Net = (Sum of captures) - (Sum of refunds) per supplier
4. **Convert:** All amounts to USD using FX rates

### Expected Processable Transactions
- **Captures (completed):** 203 transactions
- **Refunds (completed):** 27 transactions
- **Ignored:** 248 authorizations + 11 failed + 14 pending = 273 transactions

### Test Assertions
- Verify only 230 transactions are processed (203 captures + 27 refunds)
- Verify 248 transactions are ignored (all authorizations)
- Verify each supplier's net settlement = captures - refunds
- Verify FX conversion applied correctly for each currency
- Verify SUP007 shows expected high refund impact on settlement

## Regenerating Data

To generate new test data with different random values:
```bash
go run scripts/generate_testdata.go --seed 99 --output testdata/new_transactions.csv
```

The seed value ensures reproducibility - same seed always produces identical data.
