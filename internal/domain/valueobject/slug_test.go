package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "simple string",
			input: "Hello World",
			want:  "hello-world",
		},
		{
			name:  "already a slug",
			input: "hello-world",
			want:  "hello-world",
		},
		{
			name:  "uppercase string",
			input: "THE DARK KNIGHT",
			want:  "the-dark-knight",
		},
		{
			name:  "special characters replaced",
			input: "Hello, World! How's It Going?",
			want:  "hello-world-how-s-it-going",
		},
		{
			name:  "multiple spaces collapsed",
			input: "hello   world",
			want:  "hello-world",
		},
		{
			name:  "unicode characters normalized",
			input: "cafe\u0301",
			want:  "cafe",
		},
		{
			name:  "unicode accented characters",
			input: "El Ni\u00f1o",
			want:  "el-nino",
		},
		{
			name:  "numbers preserved",
			input: "Episode 42",
			want:  "episode-42",
		},
		{
			name:  "leading and trailing special chars trimmed",
			input: "---hello---",
			want:  "hello",
		},
		{
			name:    "empty string returns error",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only special characters returns error",
			input:   "!!!@@@###",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug, err := NewSlug(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, slug.IsZero())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, slug.String())
				assert.False(t, slug.IsZero())
			}
		})
	}
}

func TestSlugFromTrusted(t *testing.T) {
	slug := SlugFromTrusted("my-trusted-slug")
	assert.Equal(t, "my-trusted-slug", slug.String())
	assert.False(t, slug.IsZero())
}

func TestSlugFromTrusted_Empty(t *testing.T) {
	slug := SlugFromTrusted("")
	assert.Equal(t, "", slug.String())
	assert.True(t, slug.IsZero())
}

func TestSlug_IsZero(t *testing.T) {
	tests := []struct {
		name string
		slug Slug
		want bool
	}{
		{
			name: "zero value",
			slug: Slug{},
			want: true,
		},
		{
			name: "non-zero value",
			slug: SlugFromTrusted("test"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.slug.IsZero())
		})
	}
}
