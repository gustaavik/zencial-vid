package auth

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionTokenService_GenerateProducesUniqueTokens(t *testing.T) {
	svc := NewSessionTokenService()

	seen := make(map[string]struct{})
	for i := 0; i < 1000; i++ {
		token, hash, err := svc.Generate()
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		// 32 bytes base64url (no padding) -> 43 chars.
		assert.Equal(t, 43, len(token))
		// Hex SHA-256 -> 64 chars.
		assert.Equal(t, 64, len(hash))
		_, err = base64.RawURLEncoding.DecodeString(token)
		require.NoError(t, err, "token must decode as base64url")
		_, err = hex.DecodeString(hash)
		require.NoError(t, err, "hash must be hex")
		if _, dup := seen[token]; dup {
			t.Fatalf("duplicate token at iter %d", i)
		}
		seen[token] = struct{}{}
	}
}

func TestSessionTokenService_HashIsDeterministic(t *testing.T) {
	svc := NewSessionTokenService()

	token, hash, err := svc.Generate()
	require.NoError(t, err)

	assert.Equal(t, hash, svc.Hash(token), "hashing the same token must yield the stored hash")
	assert.NotEqual(t, hash, svc.Hash(token+"x"), "hashes must differ for different inputs")
}
