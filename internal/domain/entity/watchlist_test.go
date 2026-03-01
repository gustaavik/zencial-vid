package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWatchlistItem(t *testing.T) {
	userID := uuid.New()
	contentID := uuid.New()

	item := NewWatchlistItem(userID, contentID)

	require.NotNil(t, item)
	assert.NotEqual(t, uuid.Nil, item.ID)
	assert.Equal(t, userID, item.UserID)
	assert.Equal(t, contentID, item.ContentID)
	assert.Nil(t, item.Content, "Content should not be preloaded")
	assert.False(t, item.AddedAt.IsZero(), "AddedAt should be set")
}
