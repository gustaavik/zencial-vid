package audit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

type listFakeRepo struct {
	logs  []entity.AuditLog
	total int64
	err   error
}

func (r *listFakeRepo) Create(_ context.Context, _ *entity.AuditLog) error { return nil }
func (r *listFakeRepo) List(_ context.Context, _ *filter.FilterSet) ([]entity.AuditLog, int64, error) {
	return r.logs, r.total, r.err
}

func TestService_List_HappyPath(t *testing.T) {
	repo := &listFakeRepo{
		logs: []entity.AuditLog{
			{ID: uuid.New(), EventName: "user.logged_in", EntityType: "user", OccurredAt: time.Now().UTC()},
		},
		total: 42,
	}
	svc := NewService(repo, newQuietLogger())

	got, total, appErr := svc.List(context.Background(), &filter.FilterSet{})
	require.Nil(t, appErr)
	assert.Equal(t, int64(42), total)
	assert.Len(t, got, 1)
}

func TestService_List_RepoError(t *testing.T) {
	repo := &listFakeRepo{err: errors.New("db down")}
	svc := NewService(repo, newQuietLogger())

	_, _, appErr := svc.List(context.Background(), &filter.FilterSet{})
	require.NotNil(t, appErr)
	assert.Equal(t, apperror.CodeInternalError, appErr.Code)
}
