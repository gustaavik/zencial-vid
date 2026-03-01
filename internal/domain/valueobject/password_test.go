package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHashedPassword(t *testing.T) {
	hash := "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01234"
	p := NewHashedPassword(hash)

	assert.Equal(t, hash, p.String())
	assert.False(t, p.IsZero())
}

func TestHashedPassword_IsZero(t *testing.T) {
	tests := []struct {
		name string
		hash string
		want bool
	}{
		{"empty hash is zero", "", true},
		{"non-empty hash is not zero", "somehash", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewHashedPassword(tt.hash)
			assert.Equal(t, tt.want, p.IsZero())
		})
	}
}

func TestHashedPassword_String(t *testing.T) {
	hash := "my-hashed-password"
	p := NewHashedPassword(hash)
	assert.Equal(t, hash, p.String())
}
