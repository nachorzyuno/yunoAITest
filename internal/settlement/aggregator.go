package settlement

import (
	"github.com/ignacio/solara-settlement/internal/domain"
)

// Aggregator groups transactions by supplier and filters them based on settlement rules.
// Only settleable transactions (completed captures and refunds) are included in the grouping.
// Pending, failed, and authorization transactions are automatically filtered out.
type Aggregator struct{}

// NewAggregator creates a new transaction aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// SupplierTransactionGroup represents all transactions for a supplier,
// including both settleable transactions and authorizations
type SupplierTransactionGroup struct {
	Settleable     []*domain.Transaction // Completed captures and refunds
	Authorizations []*domain.Transaction // Authorization transactions for volatility detection
}

// GroupBySupplier groups transactions by supplier ID, including only settleable transactions.
// Returns a map where keys are supplier IDs and values are slices of their transactions.
// Only transactions where IsSettleable() returns true are included in the result.
func (a *Aggregator) GroupBySupplier(transactions []*domain.Transaction) map[string][]*domain.Transaction {
	grouped := make(map[string][]*domain.Transaction)

	for _, tx := range transactions {
		if tx.IsSettleable() {
			grouped[tx.SupplierID] = append(grouped[tx.SupplierID], tx)
		}
	}

	return grouped
}

// GroupAllBySupplier groups both settleable and authorization transactions by supplier.
// This is used for anomaly detection and volatility analysis.
// Returns a map where keys are supplier IDs and values contain both settleable and authorization transactions.
func (a *Aggregator) GroupAllBySupplier(transactions []*domain.Transaction) map[string]*SupplierTransactionGroup {
	grouped := make(map[string]*SupplierTransactionGroup)

	for _, tx := range transactions {
		// Initialize group if not exists
		if _, exists := grouped[tx.SupplierID]; !exists {
			grouped[tx.SupplierID] = &SupplierTransactionGroup{
				Settleable:     make([]*domain.Transaction, 0),
				Authorizations: make([]*domain.Transaction, 0),
			}
		}

		// Categorize transaction
		if tx.IsSettleable() {
			grouped[tx.SupplierID].Settleable = append(grouped[tx.SupplierID].Settleable, tx)
		} else if tx.Type == domain.Authorization && tx.Status == domain.Completed {
			// Track completed authorizations for volatility detection
			grouped[tx.SupplierID].Authorizations = append(grouped[tx.SupplierID].Authorizations, tx)
		}
	}

	return grouped
}
