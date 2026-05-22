package cast

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func newActiveVideo(videoID, uploaderID uuid.UUID) *entity.Video {
	slug, _ := valueobject.NewSlug("test-video")
	return &entity.Video{
		ID:         videoID,
		UploadedBy: uploaderID,
		Slug:       slug,
		Status:     entity.VideoStatusDraft,
	}
}

func TestCreate_ArchivedCast(t *testing.T) {
	callerID := uuid.New()
	videoID := uuid.New()
	castID := uuid.New()
	roles := []entity.UserRole{entity.RoleAdmin}

	castRepo := &mockCastRepo{
		findOrCreateFn: func(_ context.Context, _ string) (*entity.Cast, error) {
			return newArchivedCastMember(castID), nil
		},
	}
	videoRepo := &mockVideoRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Video, error) {
			return newActiveVideo(id, callerID), nil
		},
	}

	svc := newTestService(castRepo, &mockVideoCastRepo{}, videoRepo, nil)
	out, appErr := svc.Create(context.Background(), &CreateInput{
		VideoID:     videoID,
		Name:        "Jane Doe",
		Role:        "actor",
		CallerID:    callerID,
		CallerRoles: roles,
	})

	require.NotNil(t, appErr)
	assert.Equal(t, apperror.CodeCastArchived, appErr.Code)
	assert.Nil(t, out)
}
