package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid email",
			input:   "user@example.com",
			want:    "user@example.com",
			wantErr: false,
		},
		{
			name:    "normalizes to lowercase",
			input:   "User@Example.COM",
			want:    "user@example.com",
			wantErr: false,
		},
		{
			name:    "trims whitespace",
			input:   "  user@example.com  ",
			want:    "user@example.com",
			wantErr: false,
		},
		{
			name:    "trims whitespace and normalizes case",
			input:   "  User@Example.COM  ",
			want:    "user@example.com",
			wantErr: false,
		},
		{
			name:    "empty email returns error",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace-only returns error",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "invalid format - no at sign",
			input:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "invalid format - no domain",
			input:   "user@",
			wantErr: true,
		},
		{
			name:    "invalid format - no local part",
			input:   "@example.com",
			wantErr: true,
		},
		{
			name:    "invalid format - double at",
			input:   "user@@example.com",
			wantErr: true,
		},
		{
			name:    "valid email with plus addressing",
			input:   "user+tag@example.com",
			want:    "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots",
			input:   "first.last@example.com",
			want:    "first.last@example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, email.IsZero(), "email should be zero value on error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, email.String())
				assert.False(t, email.IsZero())
			}
		})
	}
}

func TestEmailFromTrusted(t *testing.T) {
	email := EmailFromTrusted("trusted@example.com")
	assert.Equal(t, "trusted@example.com", email.String())
	assert.False(t, email.IsZero())
}

func TestEmailFromTrusted_Empty(t *testing.T) {
	email := EmailFromTrusted("")
	assert.Equal(t, "", email.String())
	assert.True(t, email.IsZero())
}

func TestEmail_IsZero(t *testing.T) {
	tests := []struct {
		name  string
		email Email
		want  bool
	}{
		{
			name:  "zero value email",
			email: Email{},
			want:  true,
		},
		{
			name:  "non-zero email",
			email: EmailFromTrusted("test@example.com"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.email.IsZero())
		})
	}
}

func TestEmail_String(t *testing.T) {
	email := EmailFromTrusted("hello@world.com")
	assert.Equal(t, "hello@world.com", email.String())
}
