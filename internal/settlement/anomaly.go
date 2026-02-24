package settlement

import (
	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
)

// Anomaly type constants
const (
	AnomalyHighRefundRate = "HIGH_REFUND_RATE"   // Refund rate > 20% of captures
	AnomalyVolatility     = "VOLATILITY_WARNING"  // FX rate variance > 5% between auth and capture
	AnomalyOrphanedRefund = "ORPHANED_REFUND"     // Refund without matching capture
	AnomalyDuplicateID    = "DUPLICATE_ID"        // Duplicate transaction ID
	AnomalyNegativeNet    = "NEGATIVE_NET"        // Informational: supplier owes money back
)

// DetectHighRefundRate checks if a supplier's refund rate exceeds 20% of captures
// Returns true if the refund rate is above the threshold
func DetectHighRefundRate(settlement *domain.SupplierSettlement) bool {
	if settlement.TotalCapturesUSD.IsZero() {
		// If no captures, cannot calculate refund rate
		return false
	}

	// Calculate refund rate as percentage: (refunds / captures) * 100
	refundRate := settlement.TotalRefundsUSD.Div(settlement.TotalCapturesUSD).Mul(decimal.NewFromInt(100))
	settlement.RefundRatePct = refundRate

	// Flag if refund rate exceeds 20%
	threshold := decimal.NewFromInt(20)
	return refundRate.GreaterThan(threshold)
}

// DetectOrphanedRefunds identifies refunds that don't have a matching capture
// Returns a list of orphaned transaction IDs
func DetectOrphanedRefunds(transactions []*domain.Transaction) []string {
	// Build a set of completed capture transaction IDs
	captures := make(map[string]bool)
	for _, tx := range transactions {
		if tx.Type == domain.Capture && tx.Status == domain.Completed {
			captures[tx.ID] = true
		}
	}

	// Find refunds without matching captures
	// In a real system, refunds would reference their capture ID
	// For this implementation, we'll check if there's at least one capture for the supplier
	supplierHasCaptures := make(map[string]bool)
	for _, tx := range transactions {
		if tx.Type == domain.Capture && tx.Status == domain.Completed {
			supplierHasCaptures[tx.SupplierID] = true
		}
	}

	orphans := make([]string, 0)
	for _, tx := range transactions {
		if tx.Type == domain.Refund && tx.Status == domain.Completed {
			// Check if this supplier has any captures
			if !supplierHasCaptures[tx.SupplierID] {
				orphans = append(orphans, tx.ID)
			}
		}
	}

	return orphans
}

// DetectDuplicateIDs identifies duplicate transaction IDs in the dataset
// Returns a list of duplicate transaction IDs
func DetectDuplicateIDs(transactions []*domain.Transaction) []string {
	seen := make(map[string]bool)
	duplicates := make([]string, 0)
	duplicateSet := make(map[string]bool) // To avoid reporting same ID multiple times

	for _, tx := range transactions {
		if seen[tx.ID] {
			// This is a duplicate
			if !duplicateSet[tx.ID] {
				duplicates = append(duplicates, tx.ID)
				duplicateSet[tx.ID] = true
			}
		}
		seen[tx.ID] = true
	}

	return duplicates
}

// DetectNegativeNet checks if a supplier has a negative net settlement
// This is informational rather than an error condition
func DetectNegativeNet(settlement *domain.SupplierSettlement) bool {
	return settlement.NetAmountUSD.LessThan(decimal.Zero)
}
