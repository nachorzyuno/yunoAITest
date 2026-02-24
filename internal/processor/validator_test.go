package processor

import (
	"testing"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	assert.NotNil(t, validator)
}

func TestValidator_Validate_ValidTransaction(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.50),
		Currency:       domain.USD,
		Timestamp:      time.Now().Add(-1 * time.Hour),
		Status:         domain.Completed,
	}

	err := validator.Validate(tx)
	assert.NoError(t, err)
}

func TestValidator_Validate_InvalidCurrency(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.50),
		Currency:       domain.Currency("EUR"),
		Timestamp:      time.Now().Add(-1 * time.Hour),
		Status:         domain.Completed,
	}

	err := validator.Validate(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid currency")
}

func TestValidator_Validate_NegativeAmount(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(-50.00),
		Currency:       domain.USD,
		Timestamp:      time.Now().Add(-1 * time.Hour),
		Status:         domain.Completed,
	}

	err := validator.Validate(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be positive")
}

func TestValidator_Validate_ZeroAmount(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.Zero,
		Currency:       domain.USD,
		Timestamp:      time.Now().Add(-1 * time.Hour),
		Status:         domain.Completed,
	}

	err := validator.Validate(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be positive")
}

func TestValidator_Validate_InvalidType(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.TransactionType("invalid"),
		OriginalAmount: decimal.NewFromFloat(100.50),
		Currency:       domain.USD,
		Timestamp:      time.Now().Add(-1 * time.Hour),
		Status:         domain.Completed,
	}

	err := validator.Validate(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction type")
}

func TestValidator_Validate_InvalidStatus(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.50),
		Currency:       domain.USD,
		Timestamp:      time.Now().Add(-1 * time.Hour),
		Status:         domain.TransactionStatus("invalid"),
	}

	err := validator.Validate(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction status")
}

func TestValidator_Validate_FutureDate(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.50),
		Currency:       domain.USD,
		Timestamp:      time.Now().Add(24 * time.Hour),
		Status:         domain.Completed,
	}

	err := validator.Validate(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timestamp cannot be in the future")
}

func TestValidator_Validate_ZeroTimestamp(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.50),
		Currency:       domain.USD,
		Timestamp:      time.Time{},
		Status:         domain.Completed,
	}

	err := validator.Validate(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timestamp cannot be zero")
}

func TestValidator_Validate_EmptyTransactionID(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "",
		SupplierID:     "sup123",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.50),
		Currency:       domain.USD,
		Timestamp:      time.Now().Add(-1 * time.Hour),
		Status:         domain.Completed,
	}

	err := validator.Validate(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction ID cannot be empty")
}

func TestValidator_Validate_EmptySupplierID(t *testing.T) {
	validator := NewValidator()

	tx := &domain.Transaction{
		ID:             "tx001",
		SupplierID:     "",
		Type:           domain.Capture,
		OriginalAmount: decimal.NewFromFloat(100.50),
		Currency:       domain.USD,
		Timestamp:      time.Now().Add(-1 * time.Hour),
		Status:         domain.Completed,
	}

	err := validator.Validate(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "supplier ID cannot be empty")
}

func TestValidator_ValidateBatch(t *testing.T) {
	validator := NewValidator()
	validTime := time.Now().Add(-1 * time.Hour)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100.50),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup456",
			Type:           domain.Refund,
			OriginalAmount: decimal.NewFromFloat(50.00),
			Currency:       domain.BRL,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	err := validator.ValidateBatch(transactions)
	assert.NoError(t, err)
}

func TestValidator_ValidateBatch_WithInvalidTransaction(t *testing.T) {
	validator := NewValidator()
	validTime := time.Now().Add(-1 * time.Hour)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100.50),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup456",
			Type:           domain.Refund,
			OriginalAmount: decimal.NewFromFloat(-50.00), // Invalid negative amount
			Currency:       domain.BRL,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	err := validator.ValidateBatch(transactions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction 2")
	assert.Contains(t, err.Error(), "amount must be positive")
}
