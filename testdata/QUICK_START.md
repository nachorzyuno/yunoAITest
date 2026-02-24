# Quick Start Guide - Test Data Generator

## Generate Test Data (Default)

```bash
cd /Users/ignacio/yunoAITest
go run scripts/generate_testdata.go
```

Output: `testdata/transactions.csv` (478 transactions)

## Generate with Custom Parameters

```bash
# Custom output location
go run scripts/generate_testdata.go --output my_data.csv

# Different random seed for varied data
go run scripts/generate_testdata.go --seed 99

# Both custom output and seed
go run scripts/generate_testdata.go --seed 12345 --output testdata/custom.csv
```

## View Generated Data

```bash
# View first 20 transactions
head -20 testdata/transactions.csv

# Count total transactions
wc -l testdata/transactions.csv

# Show specific supplier
grep "SUP001" testdata/transactions.csv

# Show only captures
grep "capture" testdata/transactions.csv

# Show only refunds
grep "refund" testdata/transactions.csv

# Count by status
cut -d',' -f7 testdata/transactions.csv | sort | uniq -c
```

## Key Numbers

- **Total:** 478 transactions
- **Suppliers:** 7 (SUP001 - SUP007)
- **Date Range:** 2024-01-01 to 2024-02-06
- **Settlement-relevant:** 230 transactions (203 captures + 27 refunds)
- **Ignored:** 248 transactions (all authorizations, failed, pending)

## Edge Cases in Data

1. **SUP007:** Only 7 transactions, 33.3% refund rate (highest)
2. **SUP001:** Multi-currency (ARS, BRL, MXN)
3. **SUP002:** Single currency only (100% BRL)
4. **Failed:** 11 failed authorizations
5. **Pending:** 14 pending authorizations

## Settlement Engine Testing

### Expected Processing
```bash
# Count processable transactions
grep -E "(capture|refund)" testdata/transactions.csv | grep "completed" | wc -l
# Result: 230
```

### Expected Ignoring
```bash
# Count ignored transactions
grep "authorization" testdata/transactions.csv | wc -l
# Result: 248 (all authorizations ignored)
```

## Quick Stats Commands

```bash
# Transactions per supplier
for sup in SUP001 SUP002 SUP003 SUP004 SUP005 SUP006 SUP007; do
  echo "$sup: $(grep $sup testdata/transactions.csv | wc -l)"
done

# Currency distribution
cut -d',' -f5 testdata/transactions.csv | tail -n +2 | sort | uniq -c

# Status distribution
cut -d',' -f7 testdata/transactions.csv | tail -n +2 | sort | uniq -c

# Transaction types
cut -d',' -f3 testdata/transactions.csv | tail -n +2 | sort | uniq -c
```

## Documentation

- **Full README:** `scripts/README.md`
- **Data Summary:** `testdata/DATA_SUMMARY.md`
- **Validation Report:** `testdata/VALIDATION_REPORT.md`
- **This Guide:** `testdata/QUICK_START.md`

## Script Help

```bash
go run scripts/generate_testdata.go --help
```

## Reproducibility

Same seed = same data every time:
```bash
go run scripts/generate_testdata.go --seed 42
# Always generates identical 478 transactions
```
