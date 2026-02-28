package valueobject

import "fmt"

// Money represents a monetary amount with currency.
type Money struct {
	Amount   int64  // Amount in smallest currency unit (e.g., cents)
	Currency string // ISO 4217 currency code (e.g., "USD")
}

// NewMoney creates a Money value object.
func NewMoney(amount int64, currency string) Money {
	return Money{Amount: amount, Currency: currency}
}

// String returns a human-readable representation.
func (m Money) String() string {
	major := m.Amount / 100
	minor := m.Amount % 100
	return fmt.Sprintf("%s %d.%02d", m.Currency, major, minor)
}

// IsZero reports whether the money amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// Add returns a new Money with the sum of both amounts.
// Panics if currencies differ.
func (m Money) Add(other Money) Money {
	if m.Currency != other.Currency {
		panic(fmt.Sprintf("cannot add different currencies: %s and %s", m.Currency, other.Currency))
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}
}
