package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	Capture TransactionType = "capture"
	Refund  TransactionType = "refund"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	Completed TransactionStatus = "completed"
	Pending   TransactionStatus = "pending"
	Failed    TransactionStatus = "failed"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID             string
	SupplierID     string
	Type           TransactionType
	OriginalAmount decimal.Decimal
	Currency       Currency
	Timestamp      time.Time
	Status         TransactionStatus
}

// Validate checks if the transaction is valid
func (t *Transaction) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("transaction ID cannot be empty")
	}
	if t.SupplierID == "" {
		return fmt.Errorf("supplier ID cannot be empty")
	}
	if err := t.ValidateType(); err != nil {
		return err
	}
	if t.OriginalAmount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("transaction amount must be positive")
	}
	if err := t.Currency.Validate(); err != nil {
		return err
	}
	if t.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}
	if t.Timestamp.After(time.Now()) {
		return fmt.Errorf("timestamp cannot be in the future")
	}
	if err := t.ValidateStatus(); err != nil {
		return err
	}
	return nil
}

// ValidateType checks if the transaction type is valid
func (t *Transaction) ValidateType() error {
	switch t.Type {
	case Capture, Refund:
		return nil
	default:
		return fmt.Errorf("invalid transaction type: %s", t.Type)
	}
}

// ValidateStatus checks if the transaction status is valid
func (t *Transaction) ValidateStatus() error {
	switch t.Status {
	case Completed, Pending, Failed:
		return nil
	default:
		return fmt.Errorf("invalid transaction status: %s", t.Status)
	}
}

// IsSettleable returns true if the transaction should be included in settlement
func (t *Transaction) IsSettleable() bool {
	return (t.Type == Capture || t.Type == Refund) && t.Status == Completed
}
