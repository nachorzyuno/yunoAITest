package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Validate(t *testing.T) {
	validTime := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid transaction",
			tx: Transaction{
				ID:             "tx123",
				SupplierID:     "sup456",
				Type:           Capture,
				OriginalAmount: decimal.NewFromFloat(100.50),
				Currency:       USD,
				Timestamp:      validTime,
				Status:         Completed,
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			tx: Transaction{
				ID:             "",
				SupplierID:     "sup456",
				Type:           Capture,
				OriginalAmount: decimal.NewFromFloat(100.50),
				Currency:       USD,
				Timestamp:      validTime,
				Status:         Completed,
			},
			wantErr: true,
			errMsg:  "transaction ID cannot be empty",
		},
		{
			name: "empty supplier ID",
			tx: Transaction{
				ID:             "tx123",
				SupplierID:     "",
				Type:           Capture,
				OriginalAmount: decimal.NewFromFloat(100.50),
				Currency:       USD,
				Timestamp:      validTime,
				Status:         Completed,
			},
			wantErr: true,
			errMsg:  "supplier ID cannot be empty",
		},
		{
			name: "invalid type",
			tx: Transaction{
				ID:             "tx123",
				SupplierID:     "sup456",
				Type:           TransactionType("invalid"),
				OriginalAmount: decimal.NewFromFloat(100.50),
				Currency:       USD,
				Timestamp:      validTime,
				Status:         Completed,
			},
			wantErr: true,
			errMsg:  "invalid transaction type",
		},
		{
			name: "zero amount",
			tx: Transaction{
				ID:             "tx123",
				SupplierID:     "sup456",
				Type:           Capture,
				OriginalAmount: decimal.Zero,
				Currency:       USD,
				Timestamp:      validTime,
				Status:         Completed,
			},
			wantErr: true,
			errMsg:  "transaction amount must be positive",
		},
		{
			name: "negative amount",
			tx: Transaction{
				ID:             "tx123",
				SupplierID:     "sup456",
				Type:           Capture,
				OriginalAmount: decimal.NewFromFloat(-50.00),
				Currency:       USD,
				Timestamp:      validTime,
				Status:         Completed,
			},
			wantErr: true,
			errMsg:  "transaction amount must be positive",
		},
		{
			name: "invalid currency",
			tx: Transaction{
				ID:             "tx123",
				SupplierID:     "sup456",
				Type:           Capture,
				OriginalAmount: decimal.NewFromFloat(100.50),
				Currency:       Currency("EUR"),
				Timestamp:      validTime,
				Status:         Completed,
			},
			wantErr: true,
			errMsg:  "unsupported currency",
		},
		{
			name: "zero timestamp",
			tx: Transaction{
				ID:             "tx123",
				SupplierID:     "sup456",
				Type:           Capture,
				OriginalAmount: decimal.NewFromFloat(100.50),
				Currency:       USD,
				Timestamp:      time.Time{},
				Status:         Completed,
			},
			wantErr: true,
			errMsg:  "timestamp cannot be zero",
		},
		{
			name: "future timestamp",
			tx: Transaction{
				ID:             "tx123",
				SupplierID:     "sup456",
				Type:           Capture,
				OriginalAmount: decimal.NewFromFloat(100.50),
				Currency:       USD,
				Timestamp:      time.Now().Add(1 * time.Hour),
				Status:         Completed,
			},
			wantErr: true,
			errMsg:  "timestamp cannot be in the future",
		},
		{
			name: "invalid status",
			tx: Transaction{
				ID:             "tx123",
				SupplierID:     "sup456",
				Type:           Capture,
				OriginalAmount: decimal.NewFromFloat(100.50),
				Currency:       USD,
				Timestamp:      validTime,
				Status:         TransactionStatus("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid transaction status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransaction_IsSettleable(t *testing.T) {
	tests := []struct {
		name   string
		txType TransactionType
		status TransactionStatus
		want   bool
	}{
		{"capture completed", Capture, Completed, true},
		{"refund completed", Refund, Completed, true},
		{"capture pending", Capture, Pending, false},
		{"refund failed", Refund, Failed, false},
		{"invalid type completed", TransactionType("other"), Completed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := Transaction{
				Type:   tt.txType,
				Status: tt.status,
			}
			assert.Equal(t, tt.want, tx.IsSettleable())
		})
	}
}
