// Package fxrate provides foreign exchange rate services for currency conversion.
//
// This package defines the Provider interface for retrieving historical exchange rates
// and includes implementations for converting transaction amounts from local currencies
// to USD. The current implementation uses a MockProvider with simulated rates for
// demonstration purposes, but the Provider interface can be implemented with real
// FX data sources (e.g., OpenExchangeRates, CurrencyLayer, or internal services).
//
// The Service type provides high-level conversion functionality that applies
// historical FX rates based on transaction dates, ensuring accurate settlement
// calculations that reflect market conditions at the time of each transaction.
package fxrate

import (
	"time"

	"github.com/ignacio/solara-settlement/internal/domain"
	"github.com/shopspring/decimal"
)

// Provider defines the interface for retrieving foreign exchange rates
type Provider interface {
	// GetRate returns the exchange rate for converting from the specified currency to USD
	// for the given date. Returns an error if the currency is not supported or if the
	// rate cannot be retrieved.
	GetRate(currency domain.Currency, date time.Time) (decimal.Decimal, error)
}
