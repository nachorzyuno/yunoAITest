package settlement

import (
	"testing"
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewAggregator(t *testing.T) {
	agg := NewAggregator()
	assert.NotNil(t, agg)
}

func TestAggregator_GroupBySupplier(t *testing.T) {
	agg := NewAggregator()
	validTime := time.Now().Add(-1 * time.Hour)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup123",
			Type:           domain.Refund,
			OriginalAmount: decimal.NewFromFloat(20),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx003",
			SupplierID:     "sup456",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(200),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	grouped := agg.GroupBySupplier(transactions)

	assert.Equal(t, 2, len(grouped))
	assert.Equal(t, 2, len(grouped["sup123"]))
	assert.Equal(t, 1, len(grouped["sup456"]))
}

func TestAggregator_GroupBySupplier_FiltersPendingAndFailed(t *testing.T) {
	agg := NewAggregator()
	validTime := time.Now().Add(-1 * time.Hour)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup123",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(50),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Pending,
		},
		{
			ID:             "tx003",
			SupplierID:     "sup123",
			Type:           domain.Refund,
			OriginalAmount: decimal.NewFromFloat(20),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Failed,
		},
	}

	grouped := agg.GroupBySupplier(transactions)

	assert.Equal(t, 1, len(grouped))
	assert.Equal(t, 1, len(grouped["sup123"]), "should only include completed transactions")
	assert.Equal(t, "tx001", grouped["sup123"][0].ID)
}

func TestAggregator_GroupBySupplier_EmptyInput(t *testing.T) {
	agg := NewAggregator()

	grouped := agg.GroupBySupplier([]*domain.Transaction{})

	assert.Equal(t, 0, len(grouped))
}

func TestAggregator_GroupBySupplier_MultipleSuppliers(t *testing.T) {
	agg := NewAggregator()
	validTime := time.Now().Add(-1 * time.Hour)

	transactions := []*domain.Transaction{
		{
			ID:             "tx001",
			SupplierID:     "sup001",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(100),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx002",
			SupplierID:     "sup002",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(200),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
		{
			ID:             "tx003",
			SupplierID:     "sup003",
			Type:           domain.Capture,
			OriginalAmount: decimal.NewFromFloat(300),
			Currency:       domain.USD,
			Timestamp:      validTime,
			Status:         domain.Completed,
		},
	}

	grouped := agg.GroupBySupplier(transactions)

	assert.Equal(t, 3, len(grouped))
	assert.Equal(t, 1, len(grouped["sup001"]))
	assert.Equal(t, 1, len(grouped["sup002"]))
	assert.Equal(t, 1, len(grouped["sup003"]))
}
