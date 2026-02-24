package settlement

import (
	"fmt"
	"log"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/ignacio/solara-settlement/internal/fxrate"
)

// Engine orchestrates the settlement calculation process by applying FX rates
// and aggregating transactions per supplier. It coordinates between the FX rate
// service and the transaction aggregator to produce comprehensive settlement reports.
//
// The engine processes only "settleable" transactions (completed captures and refunds),
// applies historical FX rates based on transaction dates, and generates detailed
// settlement breakdowns per supplier.
type Engine struct {
	fxService  *fxrate.Service
	aggregator *Aggregator
}

// NewEngine creates a new settlement calculation engine with the provided FX rate service.
// The engine will use this service to convert all transaction amounts to USD.
func NewEngine(fxService *fxrate.Service) *Engine {
	return &Engine{
		fxService:  fxService,
		aggregator: NewAggregator(),
	}
}

// Calculate processes a list of transactions and generates settlement reports per supplier.
// The method:
//  1. Detects data anomalies (duplicate IDs, orphaned refunds)
//  2. Groups transactions by supplier ID
//  3. Filters only settleable transactions (completed captures/refunds)
//  4. Applies historical FX rates to convert amounts to USD
//  5. Aggregates totals per supplier
//  6. Runs anomaly detection (high refund rates, volatility)
//
// Returns a slice of SupplierSettlement entities, one per supplier, or an error if
// any transaction cannot be processed (e.g., FX rate unavailable).
func (e *Engine) Calculate(transactions []*domain.Transaction) ([]*domain.SupplierSettlement, error) {
	// STEP 1: Detect duplicate transaction IDs
	duplicates := DetectDuplicateIDs(transactions)
	if len(duplicates) > 0 {
		log.Printf("WARNING: Duplicate transaction IDs detected: %v", duplicates)
	}

	// STEP 2: Detect orphaned refunds
	orphans := DetectOrphanedRefunds(transactions)
	if len(orphans) > 0 {
		log.Printf("WARNING: Orphaned refunds detected (refunds without matching captures): %v", orphans)
	}

	// STEP 3: Group transactions by supplier (including authorizations for volatility detection)
	grouped := e.aggregator.GroupAllBySupplier(transactions)

	// STEP 4: Calculate settlements for each supplier
	var settlements []*domain.SupplierSettlement

	for supplierID, group := range grouped {
		settlement, err := e.calculateSupplierSettlement(supplierID, group.Settleable, group.Authorizations)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate settlement for supplier %s: %w", supplierID, err)
		}

		// STEP 5: Run anomaly detection on this settlement
		e.detectAnomalies(settlement)

		settlements = append(settlements, settlement)
	}

	return settlements, nil
}

// calculateSupplierSettlement calculates settlement for a single supplier
func (e *Engine) calculateSupplierSettlement(
	supplierID string,
	transactions []*domain.Transaction,
	authorizations []*domain.Transaction,
) (*domain.SupplierSettlement, error) {
	// Get supplier name (using ID for now; in production, this would query a supplier service)
	supplierName := fmt.Sprintf("Supplier %s", supplierID)
	settlement := domain.NewSupplierSettlement(supplierID, supplierName)

	// Store authorization transactions for volatility detection
	settlement.AuthTransactions = authorizations

	// Process each transaction
	for _, tx := range transactions {
		// Convert to USD
		usdAmount, fxRate, err := e.fxService.ConvertToUSD(tx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert transaction %s: %w", tx.ID, err)
		}

		// Create settlement line
		line := domain.SettlementLine{
			Transaction: tx,
			FXRate:      fxRate,
			USDAmount:   usdAmount,
		}

		// Add to settlement
		settlement.AddLine(line)
	}

	return settlement, nil
}

// detectAnomalies runs all anomaly detection checks on a settlement
func (e *Engine) detectAnomalies(settlement *domain.SupplierSettlement) {
	// Check for high refund rate (>20%)
	if DetectHighRefundRate(settlement) {
		settlement.Warnings = append(settlement.Warnings, AnomalyHighRefundRate)
	}

	// Check for FX volatility (>5% variance between auth and capture)
	if DetectVolatilityForSettlement(settlement, e.fxService) {
		settlement.VolatilityFlag = true
		settlement.Warnings = append(settlement.Warnings, AnomalyVolatility)
	}

	// Check for negative net (informational warning)
	if DetectNegativeNet(settlement) {
		settlement.Warnings = append(settlement.Warnings, AnomalyNegativeNet)
	}
}
