package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMoney(t *testing.T) {
	m := NewMoney(799, "USD")
	assert.Equal(t, int64(799), m.Amount)
	assert.Equal(t, "USD", m.Currency)
}

func TestMoney_String(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency string
		want     string
	}{
		{
			name:     "standard amount",
			amount:   799,
			currency: "USD",
			want:     "USD 7.99",
		},
		{
			name:     "zero amount",
			amount:   0,
			currency: "USD",
			want:     "USD 0.00",
		},
		{
			name:     "whole dollars",
			amount:   1000,
			currency: "EUR",
			want:     "EUR 10.00",
		},
		{
			name:     "single cent",
			amount:   1,
			currency: "USD",
			want:     "USD 0.01",
		},
		{
			name:     "large amount",
			amount:   999999,
			currency: "GBP",
			want:     "GBP 9999.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMoney(tt.amount, tt.currency)
			assert.Equal(t, tt.want, m.String())
		})
	}
}

func TestMoney_IsZero(t *testing.T) {
	tests := []struct {
		name   string
		amount int64
		want   bool
	}{
		{
			name:   "zero amount",
			amount: 0,
			want:   true,
		},
		{
			name:   "positive amount",
			amount: 100,
			want:   false,
		},
		{
			name:   "negative amount",
			amount: -50,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMoney(tt.amount, "USD")
			assert.Equal(t, tt.want, m.IsZero())
		})
	}
}

func TestMoney_Add_SameCurrency(t *testing.T) {
	tests := []struct {
		name string
		a    Money
		b    Money
		want Money
	}{
		{
			name: "basic addition",
			a:    NewMoney(100, "USD"),
			b:    NewMoney(200, "USD"),
			want: NewMoney(300, "USD"),
		},
		{
			name: "add zero",
			a:    NewMoney(500, "USD"),
			b:    NewMoney(0, "USD"),
			want: NewMoney(500, "USD"),
		},
		{
			name: "both zero",
			a:    NewMoney(0, "EUR"),
			b:    NewMoney(0, "EUR"),
			want: NewMoney(0, "EUR"),
		},
		{
			name: "add negative",
			a:    NewMoney(500, "USD"),
			b:    NewMoney(-200, "USD"),
			want: NewMoney(300, "USD"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Add(tt.b)
			assert.Equal(t, tt.want.Amount, result.Amount)
			assert.Equal(t, tt.want.Currency, result.Currency)
		})
	}
}

func TestMoney_Add_DifferentCurrencies_Panics(t *testing.T) {
	usd := NewMoney(100, "USD")
	eur := NewMoney(200, "EUR")

	assert.Panics(t, func() {
		usd.Add(eur)
	}, "adding different currencies should panic")
}
