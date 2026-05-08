package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurgeExpired(t *testing.T) {
	var captured time.Time
	repo := &mockSessionRepo{
		deleteExpiredFn: func(_ context.Context, before time.Time) (int64, error) {
			captured = before
			return 7, nil
		},
	}
	svc := newTestService(repo, nil)

	count, err := svc.PurgeExpired(context.Background())
	require.NoError(t, err)
	assert.EqualValues(t, 7, count)
	assert.Equal(t, fixedNow, captured, "should pass clock time as cutoff")
}
