package valueobject

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

// Slug represents a URL-safe identifier.
type Slug struct {
	value string
}

// NewSlug creates a Slug from a raw string by normalizing it.
func NewSlug(raw string) (Slug, error) {
	if raw == "" {
		return Slug{}, fmt.Errorf("slug cannot be empty")
	}
	slug := generateSlug(raw)
	if slug == "" {
		return Slug{}, fmt.Errorf("slug is empty after normalization")
	}
	return Slug{value: slug}, nil
}

// SlugFromTrusted creates a Slug from a trusted source without re-normalizing.
func SlugFromTrusted(value string) Slug {
	return Slug{value: value}
}

// String returns the slug string.
func (s Slug) String() string {
	return s.value
}

// IsZero reports whether the slug is empty.
func (s Slug) IsZero() bool {
	return s.value == ""
}

func generateSlug(input string) string {
	// Normalize unicode characters
	normalized := norm.NFKD.String(input)

	// Remove non-ASCII characters
	var b strings.Builder
	for _, r := range normalized {
		if r < unicode.MaxASCII {
			b.WriteRune(r)
		}
	}

	// Convert to lowercase
	lower := strings.ToLower(b.String())

	// Replace non-alphanumeric characters with hyphens
	slug := slugRegex.ReplaceAllString(lower, "-")

	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}
