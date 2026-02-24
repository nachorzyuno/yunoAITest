package domain

import (
	"github.com/shopspring/decimal"
)

// SettlementLine represents a single transaction in the settlement report
type SettlementLine struct {
	Transaction *Transaction
	FXRate      decimal.Decimal
	USDAmount   decimal.Decimal
}

// SupplierSettlement represents the aggregated settlement for a supplier
type SupplierSettlement struct {
	SupplierID        string
	SupplierName      string
	Lines             []SettlementLine
	TotalCapturesUSD  decimal.Decimal
	TotalRefundsUSD   decimal.Decimal
	NetAmountUSD      decimal.Decimal
	TransactionCount  int

	// Stretch goal fields for anomaly detection and multi-period analysis
	RefundRatePct    decimal.Decimal  // Refund rate as percentage of captures
	VolatilityFlag   bool              // True if >5% FX variance detected between auth and capture
	Warnings         []string          // List of warning codes (HIGH_REFUND_RATE, VOLATILITY_WARNING, etc.)
	AuthTransactions []*Transaction    // Authorization transactions for volatility comparison
}

// NewSupplierSettlement creates a new supplier settlement
func NewSupplierSettlement(supplierID, supplierName string) *SupplierSettlement {
	return &SupplierSettlement{
		SupplierID:       supplierID,
		SupplierName:     supplierName,
		Lines:            make([]SettlementLine, 0),
		TotalCapturesUSD: decimal.Zero,
		TotalRefundsUSD:  decimal.Zero,
		NetAmountUSD:     decimal.Zero,
		TransactionCount: 0,
		RefundRatePct:    decimal.Zero,
		VolatilityFlag:   false,
		Warnings:         make([]string, 0),
		AuthTransactions: make([]*Transaction, 0),
	}
}

// AddLine adds a settlement line and updates totals
func (s *SupplierSettlement) AddLine(line SettlementLine) {
	s.Lines = append(s.Lines, line)
	s.TransactionCount++

	switch line.Transaction.Type {
	case Capture:
		s.TotalCapturesUSD = s.TotalCapturesUSD.Add(line.USDAmount)
	case Refund:
		s.TotalRefundsUSD = s.TotalRefundsUSD.Add(line.USDAmount)
	}

	s.NetAmountUSD = s.TotalCapturesUSD.Sub(s.TotalRefundsUSD)
}
