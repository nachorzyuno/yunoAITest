// Package domain provides core business entities and domain logic for the FX-Aware Settlement Engine.
//
// This package defines the fundamental types used throughout the settlement system, including:
//   - Currency: Multi-currency support (ARS, BRL, COP, MXN, USD)
//   - Transaction: Financial transaction representation with validation
//   - Settlement: Aggregated settlement calculations per supplier
//   - Supplier: Merchant/supplier entities
//
// All monetary amounts use decimal.Decimal to ensure financial precision and avoid
// floating-point arithmetic errors in calculations.
package domain

import "fmt"

// Currency represents a supported currency code
type Currency string

const (
	ARS Currency = "ARS" // Argentine Peso
	BRL Currency = "BRL" // Brazilian Real
	COP Currency = "COP" // Colombian Peso
	MXN Currency = "MXN" // Mexican Peso
	USD Currency = "USD" // US Dollar
)

// Validate checks if the currency is supported
func (c Currency) Validate() error {
	switch c {
	case ARS, BRL, COP, MXN, USD:
		return nil
	default:
		return fmt.Errorf("unsupported currency: %s", c)
	}
}

// String returns the string representation of the currency
func (c Currency) String() string {
	return string(c)
}
