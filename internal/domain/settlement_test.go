package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewSupplierSettlement(t *testing.T) {
	settlement := NewSupplierSettlement("sup123", "Test Supplier")

	assert.Equal(t, "sup123", settlement.SupplierID)
	assert.Equal(t, "Test Supplier", settlement.SupplierName)
	assert.Equal(t, 0, len(settlement.Lines))
	assert.True(t, settlement.TotalCapturesUSD.Equal(decimal.Zero))
	assert.True(t, settlement.TotalRefundsUSD.Equal(decimal.Zero))
	assert.True(t, settlement.NetAmountUSD.Equal(decimal.Zero))
	assert.Equal(t, 0, settlement.TransactionCount)
}

func TestSupplierSettlement_AddLine(t *testing.T) {
	settlement := NewSupplierSettlement("sup123", "Test Supplier")

	// Add a capture
	captureTx := &Transaction{
		ID:             "tx1",
		SupplierID:     "sup123",
		Type:           Capture,
		OriginalAmount: decimal.NewFromFloat(100),
		Currency:       USD,
		Timestamp:      time.Now(),
		Status:         Completed,
	}
	captureLine := SettlementLine{
		Transaction: captureTx,
		FXRate:      decimal.NewFromFloat(1.0),
		USDAmount:   decimal.NewFromFloat(100),
	}
	settlement.AddLine(captureLine)

	assert.Equal(t, 1, len(settlement.Lines))
	assert.Equal(t, 1, settlement.TransactionCount)
	assert.True(t, settlement.TotalCapturesUSD.Equal(decimal.NewFromFloat(100)))
	assert.True(t, settlement.TotalRefundsUSD.Equal(decimal.Zero))
	assert.True(t, settlement.NetAmountUSD.Equal(decimal.NewFromFloat(100)))

	// Add a refund
	refundTx := &Transaction{
		ID:             "tx2",
		SupplierID:     "sup123",
		Type:           Refund,
		OriginalAmount: decimal.NewFromFloat(30),
		Currency:       USD,
		Timestamp:      time.Now(),
		Status:         Completed,
	}
	refundLine := SettlementLine{
		Transaction: refundTx,
		FXRate:      decimal.NewFromFloat(1.0),
		USDAmount:   decimal.NewFromFloat(30),
	}
	settlement.AddLine(refundLine)

	assert.Equal(t, 2, len(settlement.Lines))
	assert.Equal(t, 2, settlement.TransactionCount)
	assert.True(t, settlement.TotalCapturesUSD.Equal(decimal.NewFromFloat(100)))
	assert.True(t, settlement.TotalRefundsUSD.Equal(decimal.NewFromFloat(30)))
	assert.True(t, settlement.NetAmountUSD.Equal(decimal.NewFromFloat(70)))
}

func TestSupplierSettlement_MultipleTransactions(t *testing.T) {
	settlement := NewSupplierSettlement("sup123", "Test Supplier")

	// Add multiple captures and refunds
	captures := []decimal.Decimal{
		decimal.NewFromFloat(100),
		decimal.NewFromFloat(200),
		decimal.NewFromFloat(150),
	}
	refunds := []decimal.Decimal{
		decimal.NewFromFloat(30),
		decimal.NewFromFloat(50),
	}

	for i, amt := range captures {
		tx := &Transaction{
			ID:             "cap" + string(rune(i)),
			SupplierID:     "sup123",
			Type:           Capture,
			OriginalAmount: amt,
			Currency:       USD,
			Timestamp:      time.Now(),
			Status:         Completed,
		}
		line := SettlementLine{
			Transaction: tx,
			FXRate:      decimal.NewFromFloat(1.0),
			USDAmount:   amt,
		}
		settlement.AddLine(line)
	}

	for i, amt := range refunds {
		tx := &Transaction{
			ID:             "ref" + string(rune(i)),
			SupplierID:     "sup123",
			Type:           Refund,
			OriginalAmount: amt,
			Currency:       USD,
			Timestamp:      time.Now(),
			Status:         Completed,
		}
		line := SettlementLine{
			Transaction: tx,
			FXRate:      decimal.NewFromFloat(1.0),
			USDAmount:   amt,
		}
		settlement.AddLine(line)
	}

	assert.Equal(t, 5, settlement.TransactionCount)
	assert.True(t, settlement.TotalCapturesUSD.Equal(decimal.NewFromFloat(450)))
	assert.True(t, settlement.TotalRefundsUSD.Equal(decimal.NewFromFloat(80)))
	assert.True(t, settlement.NetAmountUSD.Equal(decimal.NewFromFloat(370)))
}
