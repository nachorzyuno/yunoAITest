package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

// Supplier defines a supplier with business characteristics
type Supplier struct {
	ID           string
	Name         string
	TargetTxns   int
	Currencies   []string
	RefundRate   float64 // 0.0 to 1.0
	AmountRanges map[string][2]float64
}

// Transaction represents a single transaction
type Transaction struct {
	ID             string
	SupplierID     string
	Type           string
	OriginalAmount float64
	Currency       string
	Timestamp      time.Time
	Status         string
	RelatedAuthID  string // For captures and refunds
}

var (
	outputPath = flag.String("output", "testdata/transactions.csv", "Output CSV file path")
	seed       = flag.Int64("seed", 42, "Random seed for reproducible data")
)

// Suppliers configuration matching the requirements
var suppliers = []Supplier{
	{
		ID:         "SUP001",
		Name:       "Hotel Marriott Buenos Aires",
		TargetTxns: 60,
		Currencies: []string{"ARS", "BRL", "MXN"}, // Multi-currency
		RefundRate: 0.10,                          // Normal 10% refund rate
		AmountRanges: map[string][2]float64{
			"ARS": {10000, 500000},
			"BRL": {500, 15000},
			"MXN": {1000, 40000},
		},
	},
	{
		ID:         "SUP002",
		Name:       "Airline LATAM",
		TargetTxns: 55,
		Currencies: []string{"BRL"}, // Mostly single currency
		RefundRate: 0.10,
		AmountRanges: map[string][2]float64{
			"BRL": {500, 15000},
		},
	},
	{
		ID:         "SUP003",
		Name:       "Car Rental Hertz Mexico",
		TargetTxns: 40,
		Currencies: []string{"MXN"},
		RefundRate: 0.10,
		AmountRanges: map[string][2]float64{
			"MXN": {1000, 40000},
		},
	},
	{
		ID:         "SUP004",
		Name:       "Hotel Copacabana Rio",
		TargetTxns: 35,
		Currencies: []string{"BRL"},
		RefundRate: 0.10,
		AmountRanges: map[string][2]float64{
			"BRL": {500, 15000},
		},
	},
	{
		ID:         "SUP005",
		Name:       "Tour Operator Colombia",
		TargetTxns: 25,
		Currencies: []string{"COP"},
		RefundRate: 0.10,
		AmountRanges: map[string][2]float64{
			"COP": {100000, 5000000},
		},
	},
	{
		ID:         "SUP006",
		Name:       "Beach Resort Cancun",
		TargetTxns: 30,
		Currencies: []string{"MXN"},
		RefundRate: 0.10,
		AmountRanges: map[string][2]float64{
			"MXN": {1000, 40000},
		},
	},
	{
		ID:         "SUP007",
		Name:       "Hostel Palermo",
		TargetTxns: 3,                  // Edge case: very low volume
		Currencies: []string{"ARS"},
		RefundRate: 0.60, // Edge case: HIGH refund rate >50%
		AmountRanges: map[string][2]float64{
			"ARS": {10000, 100000},
		},
	},
}

func main() {
	flag.Parse()

	// Set random seed for reproducibility
	rand.Seed(*seed)

	log.Printf("Generating test data with seed %d...", *seed)

	// Generate all transactions
	transactions := generateTransactions()

	log.Printf("Generated %d total transactions", len(transactions))

	// Print statistics
	printStatistics(transactions)

	// Write to CSV
	if err := writeCSV(*outputPath, transactions); err != nil {
		log.Fatalf("Failed to write CSV: %v", err)
	}

	log.Printf("Successfully wrote %d transactions to %s", len(transactions), *outputPath)
}

func generateTransactions() []Transaction {
	var allTransactions []Transaction
	txnCounter := 1

	// Start date: 2024-01-01
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Generate transactions for each supplier
	for _, supplier := range suppliers {
		log.Printf("Generating %d transactions for %s (%s)...", supplier.TargetTxns, supplier.ID, supplier.Name)

		// Spread transactions over 30 days
		for i := 0; i < supplier.TargetTxns; i++ {
			// Random day within 30 days
			dayOffset := rand.Intn(30)
			// Random hour and minute
			hour := rand.Intn(24)
			minute := rand.Intn(60)
			authTimestamp := baseDate.AddDate(0, 0, dayOffset).Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)

			// Select currency (randomly if multiple currencies)
			currency := supplier.Currencies[rand.Intn(len(supplier.Currencies))]

			// Generate amount within range
			amountRange := supplier.AmountRanges[currency]
			amount := amountRange[0] + rand.Float64()*(amountRange[1]-amountRange[0])
			amount = float64(int(amount*100)) / 100 // Round to 2 decimals

			// Create authorization transaction
			authID := fmt.Sprintf("TXN%03d", txnCounter)
			txnCounter++

			// Determine authorization status: 95% completed, 5% failed
			authStatus := "completed"
			if rand.Float64() < 0.05 {
				authStatus = "failed"
			}

			auth := Transaction{
				ID:             authID,
				SupplierID:     supplier.ID,
				Type:           "authorization",
				OriginalAmount: amount,
				Currency:       currency,
				Timestamp:      authTimestamp,
				Status:         authStatus,
			}
			allTransactions = append(allTransactions, auth)

			// If authorization failed, skip capture/refund
			if authStatus == "failed" {
				continue
			}

			// 85% of successful authorizations → captures
			// 15% remain uncaptured (some pending, some completed but not captured)
			shouldCapture := rand.Float64() < 0.85

			if !shouldCapture {
				// Some uncaptured authorizations stay "pending"
				if rand.Float64() < 0.5 {
					auth.Status = "pending"
					allTransactions[len(allTransactions)-1] = auth // Update the last added auth
				}
				continue
			}

			// Create capture (same day or +1-2 days later)
			captureDelay := time.Duration(rand.Intn(3)) * 24 * time.Hour
			captureTimestamp := authTimestamp.Add(captureDelay).Add(time.Duration(rand.Intn(300)) * time.Minute)

			captureID := fmt.Sprintf("TXN%03d", txnCounter)
			txnCounter++

			capture := Transaction{
				ID:             captureID,
				SupplierID:     supplier.ID,
				Type:           "capture",
				OriginalAmount: amount,
				Currency:       currency,
				Timestamp:      captureTimestamp,
				Status:         "completed",
				RelatedAuthID:  authID,
			}
			allTransactions = append(allTransactions, capture)

			// Determine if this capture should be refunded based on supplier refund rate
			shouldRefund := rand.Float64() < supplier.RefundRate

			if shouldRefund {
				// Create refund (3-7 days after capture)
				refundDelay := time.Duration(3+rand.Intn(5)) * 24 * time.Hour
				refundTimestamp := captureTimestamp.Add(refundDelay).Add(time.Duration(rand.Intn(300)) * time.Minute)

				refundID := fmt.Sprintf("TXN%03d", txnCounter)
				txnCounter++

				refund := Transaction{
					ID:             refundID,
					SupplierID:     supplier.ID,
					Type:           "refund",
					OriginalAmount: amount,
					Currency:       currency,
					Timestamp:      refundTimestamp,
					Status:         "completed",
					RelatedAuthID:  captureID,
				}
				allTransactions = append(allTransactions, refund)
			}
		}
	}

	// Sort transactions by timestamp for realistic ordering
	sortTransactionsByTimestamp(allTransactions)

	return allTransactions
}

func sortTransactionsByTimestamp(transactions []Transaction) {
	// Simple bubble sort (sufficient for this dataset size)
	n := len(transactions)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if transactions[j].Timestamp.After(transactions[j+1].Timestamp) {
				transactions[j], transactions[j+1] = transactions[j+1], transactions[j]
			}
		}
	}
}

func writeCSV(filepath string, transactions []Transaction) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"transaction_id", "supplier_id", "type", "original_amount", "currency", "timestamp", "status"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write transactions
	for _, txn := range transactions {
		record := []string{
			txn.ID,
			txn.SupplierID,
			txn.Type,
			fmt.Sprintf("%.2f", txn.OriginalAmount),
			txn.Currency,
			txn.Timestamp.Format(time.RFC3339),
			txn.Status,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

func printStatistics(transactions []Transaction) {
	// Count transactions by supplier
	supplierCounts := make(map[string]int)
	// Count by type
	typeCounts := make(map[string]int)
	// Count by currency
	currencyCounts := make(map[string]int)
	// Count by status
	statusCounts := make(map[string]int)
	// Count captures and refunds per supplier
	supplierCaptures := make(map[string]int)
	supplierRefunds := make(map[string]int)

	for _, txn := range transactions {
		supplierCounts[txn.SupplierID]++
		typeCounts[txn.Type]++
		currencyCounts[txn.Currency]++
		statusCounts[txn.Status]++

		if txn.Type == "capture" && txn.Status == "completed" {
			supplierCaptures[txn.SupplierID]++
		}
		if txn.Type == "refund" && txn.Status == "completed" {
			supplierRefunds[txn.SupplierID]++
		}
	}

	fmt.Println("\n=== TRANSACTION STATISTICS ===\n")

	fmt.Println("Transactions per Supplier:")
	for _, supplier := range suppliers {
		fmt.Printf("  %s (%s): %d transactions\n", supplier.ID, supplier.Name, supplierCounts[supplier.ID])
		if supplierCaptures[supplier.ID] > 0 {
			refundRate := float64(supplierRefunds[supplier.ID]) / float64(supplierCaptures[supplier.ID]) * 100
			fmt.Printf("    -> Captures: %d, Refunds: %d (%.1f%% refund rate)\n",
				supplierCaptures[supplier.ID], supplierRefunds[supplier.ID], refundRate)
		}
	}

	fmt.Println("\nTransaction Types:")
	for txnType, count := range typeCounts {
		fmt.Printf("  %s: %d\n", txnType, count)
	}

	fmt.Println("\nCurrency Distribution:")
	for currency, count := range currencyCounts {
		percentage := float64(count) / float64(len(transactions)) * 100
		fmt.Printf("  %s: %d (%.1f%%)\n", currency, count, percentage)
	}

	fmt.Println("\nStatus Distribution:")
	for status, count := range statusCounts {
		fmt.Printf("  %s: %d\n", status, count)
	}

	fmt.Println("\n=== EDGE CASES INCLUDED ===")
	fmt.Println("1. SUP007 (Hostel Palermo): Low volume with HIGH refund rate (>50%)")
	fmt.Println("2. SUP001 (Hotel Marriott): Multi-currency transactions (ARS, BRL, MXN)")
	fmt.Println("3. SUP002 (Airline LATAM): Single currency focus (BRL)")
	fmt.Println("4. Failed authorizations: Included (~5% of authorizations)")
	fmt.Println("5. Pending authorizations: Included (~7-8% remain pending/uncaptured)")
	fmt.Println("6. Realistic transaction flow: authorization → capture (+0-2 days) → refund (+3-7 days)")
	fmt.Printf("\nTotal transactions: %d\n", len(transactions))
}
