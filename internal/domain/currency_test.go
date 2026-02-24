package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrency_Validate(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		wantErr  bool
	}{
		{"valid ARS", ARS, false},
		{"valid BRL", BRL, false},
		{"valid COP", COP, false},
		{"valid MXN", MXN, false},
		{"valid USD", USD, false},
		{"invalid EUR", Currency("EUR"), true},
		{"invalid empty", Currency(""), true},
		{"invalid lowercase", Currency("ars"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.currency.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCurrency_String(t *testing.T) {
	assert.Equal(t, "ARS", ARS.String())
	assert.Equal(t, "USD", USD.String())
}
