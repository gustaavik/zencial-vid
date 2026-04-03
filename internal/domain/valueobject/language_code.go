package valueobject

import (
	"fmt"
	"strings"
)

// LanguageCode represents an ISO 639-1 language code.
type LanguageCode struct {
	value string
}

// NewLanguageCode creates a LanguageCode after validation.
// The code is normalized to lowercase before storing.
func NewLanguageCode(code string) (LanguageCode, error) {
	if len(code) < 2 || len(code) > 5 {
		return LanguageCode{}, fmt.Errorf("language code must be 2-5 characters, got %q", code)
	}
	return LanguageCode{value: strings.ToLower(code)}, nil
}

// LanguageCodeFromTrusted creates a LanguageCode from a trusted source without validation.
func LanguageCodeFromTrusted(code string) LanguageCode {
	return LanguageCode{value: code}
}

// String returns the language code string.
func (lc LanguageCode) String() string {
	return lc.value
}

// IsZero reports whether the language code is empty.
func (lc LanguageCode) IsZero() bool {
	return lc.value == ""
}
