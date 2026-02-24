package settlement

import (
	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/ignacio/solara-settlement/internal/fxrate"
	"github.com/shopspring/decimal"
)

// CalculateVolatility compares FX rates between authorization and capture transactions
// to detect significant currency fluctuations
//
// Returns:
//   - hasVolatility: true if variance exceeds 5%
//   - variance: the calculated variance as a percentage
//   - error: any error that occurred during rate lookup
func CalculateVolatility(
	authTx *domain.Transaction,
	captureTx *domain.Transaction,
	fxService *fxrate.Service,
) (hasVolatility bool, variance decimal.Decimal, err error) {
	// Get FX rate at authorization time
	authRate, err := fxService.GetRate(authTx.Currency, authTx.Timestamp)
	if err != nil {
		return false, decimal.Zero, err
	}

	// Get FX rate at capture time
	captureRate, err := fxService.GetRate(captureTx.Currency, captureTx.Timestamp)
	if err != nil {
		return false, decimal.Zero, err
	}

	// Calculate variance as percentage: abs((captureRate - authRate) / authRate) * 100
	if authRate.IsZero() {
		// Avoid division by zero
		return false, decimal.Zero, nil
	}

	variance = captureRate.Sub(authRate).Div(authRate).Abs().Mul(decimal.NewFromInt(100))

	// Flag if variance exceeds 5%
	threshold := decimal.NewFromInt(5)
	hasVolatility = variance.GreaterThan(threshold)

	return hasVolatility, variance, nil
}

// DetectVolatilityForSettlement checks for FX volatility across all auth/capture pairs
// for a supplier's transactions
//
// This function matches authorization transactions with their corresponding captures
// by currency and checks if FX rate variance exceeds 5%
func DetectVolatilityForSettlement(
	settlement *domain.SupplierSettlement,
	fxService *fxrate.Service,
) bool {
	// If no authorization transactions, no volatility to check
	if len(settlement.AuthTransactions) == 0 {
		return false
	}

	// Build a map of authorizations by currency
	authsByCurrency := make(map[domain.Currency][]*domain.Transaction)
	for _, auth := range settlement.AuthTransactions {
		authsByCurrency[auth.Currency] = append(authsByCurrency[auth.Currency], auth)
	}

	// Check each settlement line (capture/refund) against authorizations
	for _, line := range settlement.Lines {
		tx := line.Transaction

		// Only check captures (refunds are already completed money movements)
		if tx.Type != domain.Capture {
			continue
		}

		// Find matching authorizations for this currency
		authsForCurrency := authsByCurrency[tx.Currency]
		if len(authsForCurrency) == 0 {
			continue
		}

		// Check volatility against the most recent authorization for this currency
		// (In a real system, we'd match by a specific auth-capture relationship)
		for _, auth := range authsForCurrency {
			// Only compare if auth came before capture
			if auth.Timestamp.Before(tx.Timestamp) {
				hasVolatility, _, err := CalculateVolatility(auth, tx, fxService)
				if err != nil {
					// Log error but continue processing
					continue
				}
				if hasVolatility {
					return true
				}
			}
		}
	}

	return false
}
