# Deployment Guide

This guide covers deploying and running the FX-Aware Settlement Engine in production environments.

## Table of Contents

1. [Production Checklist](#production-checklist)
2. [Building for Production](#building-for-production)
3. [Environment Configuration](#environment-configuration)
4. [FX Rate Provider Setup](#fx-rate-provider-setup)
5. [Monitoring and Logging](#monitoring-and-logging)
6. [Performance Tuning](#performance-tuning)
7. [Security Considerations](#security-considerations)
8. [Operational Procedures](#operational-procedures)

## Production Checklist

Before deploying to production, ensure:

- [ ] Go 1.24.5 or higher is installed
- [ ] All tests pass: `make test`
- [ ] Code is formatted: `make fmt`
- [ ] Linter passes: `make vet`
- [ ] Real FX provider is configured (not MockProvider)
- [ ] API keys are stored securely (not in code)
- [ ] Log levels are set appropriately
- [ ] Monitoring is configured
- [ ] Backup procedures are in place
- [ ] Rollback plan is documented

## Building for Production

### Standard Build

```bash
# Build optimized binary
go build -o settlement \
  -ldflags="-s -w" \
  cmd/settlement/main.go
```

The `-ldflags="-s -w"` flag strips debug information for a smaller binary.

### Cross-Platform Builds

Build for different platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o settlement-linux \
  -ldflags="-s -w" \
  cmd/settlement/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o settlement-macos \
  -ldflags="-s -w" \
  cmd/settlement/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o settlement.exe \
  -ldflags="-s -w" \
  cmd/settlement/main.go
```

### Version Tagging

Include version information in builds:

```bash
VERSION=$(git describe --tags --always)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

go build -o settlement \
  -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" \
  cmd/settlement/main.go
```

## Environment Configuration

### Configuration File

Create a production configuration file `config.yaml`:

```yaml
# FX Rate Provider Configuration
fx_provider:
  type: "openexchangerates"  # or "currencylayer", "internal"
  api_key_env: "FX_API_KEY"  # Environment variable containing API key
  base_url: "https://openexchangerates.org/api"
  timeout: "30s"
  retry_attempts: 3

# File Paths
input:
  directory: "/data/transactions"
  pattern: "*.csv"

output:
  directory: "/data/settlements"
  format: "settlement_{date}.csv"

# Logging
logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json or text
  output: "/var/log/settlement/settlement.log"

# Performance
performance:
  max_concurrent_fx_requests: 10
  batch_size: 1000
```

### Environment Variables

Set required environment variables:

```bash
# FX Provider API Key
export FX_API_KEY="your-api-key-here"

# Optional: Override config file location
export SETTLEMENT_CONFIG="/etc/settlement/config.yaml"

# Optional: Database connection (future)
export DB_CONNECTION_STRING="postgresql://user:pass@localhost/settlements"
```

### Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.24.5-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o settlement \
    -ldflags="-s -w" \
    cmd/settlement/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/settlement .

# Create directories for data
RUN mkdir -p /data/transactions /data/settlements

# Non-root user
RUN addgroup -S settlement && adduser -S settlement -G settlement
USER settlement

ENTRYPOINT ["./settlement"]
CMD ["--config", "/etc/settlement/config.yaml"]
```

Build and run:

```bash
# Build image
docker build -t settlement-engine:latest .

# Run container
docker run -d \
  --name settlement \
  -v /path/to/config.yaml:/etc/settlement/config.yaml:ro \
  -v /path/to/transactions:/data/transactions:ro \
  -v /path/to/settlements:/data/settlements \
  -e FX_API_KEY="${FX_API_KEY}" \
  settlement-engine:latest
```

## FX Rate Provider Setup

### OpenExchangeRates

1. Sign up at https://openexchangerates.org/
2. Get your API key
3. Create implementation:

```go
// internal/fxrate/openexchange_provider.go
package fxrate

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/ignacio/solara-settlement/internal/domain"
    "github.com/shopspring/decimal"
)

type OpenExchangeProvider struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func NewOpenExchangeProvider(apiKey string) *OpenExchangeProvider {
    return &OpenExchangeProvider{
        apiKey:  apiKey,
        baseURL: "https://openexchangerates.org/api",
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (p *OpenExchangeProvider) GetRate(currency domain.Currency, date time.Time) (decimal.Decimal, error) {
    // Implementation details...
    // Call API, parse response, return rate
}
```

### Rate Limiting

Implement rate limiting to avoid exceeding API quotas:

```go
import "golang.org/x/time/rate"

type RateLimitedProvider struct {
    provider Provider
    limiter  *rate.Limiter
}

func NewRateLimitedProvider(provider Provider, requestsPerSecond int) *RateLimitedProvider {
    return &RateLimitedProvider{
        provider: provider,
        limiter:  rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
    }
}

func (p *RateLimitedProvider) GetRate(currency domain.Currency, date time.Time) (decimal.Decimal, error) {
    if err := p.limiter.Wait(context.Background()); err != nil {
        return decimal.Zero, err
    }
    return p.provider.GetRate(currency, date)
}
```

### Caching FX Rates

Cache rates to reduce API calls:

```go
import (
    "sync"
    "time"
)

type CachedProvider struct {
    provider Provider
    cache    map[string]cachedRate
    mu       sync.RWMutex
    ttl      time.Duration
}

type cachedRate struct {
    rate      decimal.Decimal
    timestamp time.Time
}

func NewCachedProvider(provider Provider, ttl time.Duration) *CachedProvider {
    return &CachedProvider{
        provider: provider,
        cache:    make(map[string]cachedRate),
        ttl:      ttl,
    }
}

func (p *CachedProvider) GetRate(currency domain.Currency, date time.Time) (decimal.Decimal, error) {
    key := fmt.Sprintf("%s-%s", currency, date.Format("2006-01-02"))

    // Check cache
    p.mu.RLock()
    if cached, found := p.cache[key]; found {
        if time.Since(cached.timestamp) < p.ttl {
            p.mu.RUnlock()
            return cached.rate, nil
        }
    }
    p.mu.RUnlock()

    // Fetch from provider
    rate, err := p.provider.GetRate(currency, date)
    if err != nil {
        return decimal.Zero, err
    }

    // Store in cache
    p.mu.Lock()
    p.cache[key] = cachedRate{
        rate:      rate,
        timestamp: time.Now(),
    }
    p.mu.Unlock()

    return rate, nil
}
```

## Monitoring and Logging

### Structured Logging

Use structured logging for better observability:

```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("processing transaction",
    zap.String("transaction_id", tx.ID),
    zap.String("supplier_id", tx.SupplierID),
    zap.String("currency", string(tx.Currency)),
    zap.String("amount", tx.OriginalAmount.String()),
)
```

### Metrics

Track key metrics:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    transactionsProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "settlement_transactions_processed_total",
            Help: "Total number of transactions processed",
        },
        []string{"supplier_id", "currency", "status"},
    )

    settlementAmount = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "settlement_amount_usd",
            Help: "Settlement amounts in USD",
        },
        []string{"supplier_id"},
    )

    fxRateLookupDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "fx_rate_lookup_duration_seconds",
            Help: "Time spent fetching FX rates",
        },
    )
)

func init() {
    prometheus.MustRegister(transactionsProcessed)
    prometheus.MustRegister(settlementAmount)
    prometheus.MustRegister(fxRateLookupDuration)
}
```

### Health Checks

Implement health check endpoint:

```go
func HealthCheck() error {
    // Check FX provider connectivity
    if err := fxProvider.Ping(); err != nil {
        return fmt.Errorf("fx provider unhealthy: %w", err)
    }

    // Check file system access
    if err := checkFileAccess("/data/transactions"); err != nil {
        return fmt.Errorf("transaction directory inaccessible: %w", err)
    }

    return nil
}
```

## Performance Tuning

### Concurrent FX Rate Lookups

Process FX rate lookups in parallel:

```go
func (e *Engine) calculateSupplierSettlement(supplierID string, transactions []*domain.Transaction) (*domain.SupplierSettlement, error) {
    settlement := domain.NewSupplierSettlement(supplierID, supplierName)

    // Create worker pool
    type result struct {
        line domain.SettlementLine
        err  error
    }

    results := make(chan result, len(transactions))
    semaphore := make(chan struct{}, 10) // Max 10 concurrent requests

    var wg sync.WaitGroup
    for _, tx := range transactions {
        wg.Add(1)
        go func(tx *domain.Transaction) {
            defer wg.Done()
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release

            usdAmount, fxRate, err := e.fxService.ConvertToUSD(tx)
            results <- result{
                line: domain.SettlementLine{
                    Transaction: tx,
                    FXRate:      fxRate,
                    USDAmount:   usdAmount,
                },
                err: err,
            }
        }(tx)
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    for res := range results {
        if res.err != nil {
            return nil, res.err
        }
        settlement.AddLine(res.line)
    }

    return settlement, nil
}
```

### Batch Processing

For very large files, process in batches:

```bash
# Split large file into batches
split -l 1000 large_transactions.csv batch_

# Process each batch
for batch in batch_*; do
    ./settlement --input $batch --output settlements_$(basename $batch).csv
done

# Combine results
cat settlements_batch_* > final_settlements.csv
```

## Security Considerations

### API Key Management

- Never commit API keys to version control
- Use environment variables or secret management systems
- Rotate keys regularly
- Use different keys for dev/staging/prod

### Input Validation

- Validate all CSV input before processing
- Sanitize file paths to prevent directory traversal
- Set file size limits to prevent resource exhaustion

### Output Protection

- Set appropriate file permissions (chmod 600)
- Encrypt sensitive settlement reports
- Use secure transfer methods (SFTP, HTTPS)

## Operational Procedures

### Daily Settlement Run

```bash
#!/bin/bash
# daily_settlement.sh

DATE=$(date +%Y-%m-%d)
INPUT_DIR="/data/transactions"
OUTPUT_DIR="/data/settlements"
LOG_FILE="/var/log/settlement/settlement_${DATE}.log"

echo "Starting settlement process for ${DATE}" >> ${LOG_FILE}

# Run settlement
./settlement \
  --input "${INPUT_DIR}/transactions_${DATE}.csv" \
  --output "${OUTPUT_DIR}/settlements_${DATE}.csv" \
  2>&1 | tee -a ${LOG_FILE}

if [ $? -eq 0 ]; then
    echo "Settlement completed successfully" >> ${LOG_FILE}
    # Archive input file
    gzip "${INPUT_DIR}/transactions_${DATE}.csv"
    mv "${INPUT_DIR}/transactions_${DATE}.csv.gz" /data/archive/
else
    echo "Settlement failed" >> ${LOG_FILE}
    # Send alert
    curl -X POST https://alerts.example.com/webhook \
      -d "Settlement process failed on ${DATE}"
fi
```

Schedule with cron:

```cron
# Run daily at 2 AM
0 2 * * * /opt/settlement/daily_settlement.sh
```

### Backup and Restore

```bash
# Backup settlements
tar -czf settlements_backup_$(date +%Y%m%d).tar.gz /data/settlements/

# Upload to S3
aws s3 cp settlements_backup_*.tar.gz s3://backups/settlements/

# Restore from backup
aws s3 cp s3://backups/settlements/settlements_backup_20240115.tar.gz .
tar -xzf settlements_backup_20240115.tar.gz -C /data/
```

### Rollback Procedure

If a settlement run produces incorrect results:

1. Stop the settlement process
2. Restore from last known good backup
3. Investigate the issue (check logs, FX rates, input data)
4. Fix the issue
5. Re-run the settlement for affected dates
6. Validate output before distributing

## Troubleshooting

### Common Issues

**Issue**: FX rate provider timeout

```bash
# Check API connectivity
curl -v https://openexchangerates.org/api/latest.json?app_id=YOUR_KEY

# Increase timeout in config
fx_provider:
  timeout: "60s"
```

**Issue**: Out of memory for large files

```bash
# Process in smaller batches
# Or increase memory limits
docker run --memory=4g settlement-engine:latest
```

**Issue**: Incorrect settlement amounts

```bash
# Verify FX rates are correct
# Check for decimal precision issues
# Review transaction filtering logic
```

## Support

For production issues:
1. Check logs: `/var/log/settlement/`
2. Review metrics dashboard
3. Contact engineering team
4. Open incident ticket with details

---

**Last Updated**: 2026-02-24
**Version**: 1.0.0
