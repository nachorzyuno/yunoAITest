package processor

import (
	"fmt"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
)

// Validator validates transaction data
type Validator struct{}

// NewValidator creates a new transaction validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a transaction
func (v *Validator) Validate(tx *domain.Transaction) error {
	if err := v.validateCurrency(tx.Currency); err != nil {
		return err
	}

	if err := v.validateAmount(tx.OriginalAmount); err != nil {
		return err
	}

	if err := v.validateType(tx.Type); err != nil {
		return err
	}

	if err := v.validateStatus(tx.Status); err != nil {
		return err
	}

	if err := v.validateDate(tx.Timestamp); err != nil {
		return err
	}

	if err := v.validateIDs(tx); err != nil {
		return err
	}

	return nil
}

// ValidateBatch validates a batch of transactions
func (v *Validator) ValidateBatch(transactions []*domain.Transaction) error {
	for i, tx := range transactions {
		if err := v.Validate(tx); err != nil {
			return fmt.Errorf("transaction %d (%s): %w", i+1, tx.ID, err)
		}
	}
	return nil
}

func (v *Validator) validateCurrency(currency domain.Currency) error {
	if err := currency.Validate(); err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}
	return nil
}

func (v *Validator) validateAmount(amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be positive, got %s", amount)
	}
	return nil
}

func (v *Validator) validateType(txType domain.TransactionType) error {
	switch txType {
	case domain.Capture, domain.Refund:
		return nil
	default:
		return fmt.Errorf("invalid transaction type: %s", txType)
	}
}

func (v *Validator) validateStatus(status domain.TransactionStatus) error {
	switch status {
	case domain.Completed, domain.Pending, domain.Failed:
		return nil
	default:
		return fmt.Errorf("invalid transaction status: %s", status)
	}
}

func (v *Validator) validateDate(timestamp time.Time) error {
	if timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	if timestamp.After(time.Now()) {
		return fmt.Errorf("timestamp cannot be in the future: %s", timestamp)
	}

	return nil
}

func (v *Validator) validateIDs(tx *domain.Transaction) error {
	if tx.ID == "" {
		return fmt.Errorf("transaction ID cannot be empty")
	}

	if tx.SupplierID == "" {
		return fmt.Errorf("supplier ID cannot be empty")
	}

	return nil
}
