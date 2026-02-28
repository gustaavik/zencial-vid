package valueobject

import (
	"fmt"
	"net/mail"
	"strings"
)

// Email represents a validated email address.
type Email struct {
	value string
}

// NewEmail creates an Email after validation.
func NewEmail(raw string) (Email, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}
	addr, err := mail.ParseAddress(normalized)
	if err != nil {
		return Email{}, fmt.Errorf("invalid email format: %w", err)
	}
	return Email{value: addr.Address}, nil
}

// EmailFromTrusted creates an Email from a trusted source (e.g., database)
// without re-validating. Use only when the value is known to be valid.
func EmailFromTrusted(value string) Email {
	return Email{value: value}
}

// String returns the email address as a string.
func (e Email) String() string {
	return e.value
}

// IsZero reports whether the email is empty.
func (e Email) IsZero() bool {
	return e.value == ""
}
