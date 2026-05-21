package cast

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UploadPictureInput holds the fields needed to upload a cast member's picture.
type UploadPictureInput struct {
	ID          uuid.UUID
	Body        io.Reader
	ContentType string
	Ext         string // e.g. ".jpg", ".png", ".webp", ".gif"
	CallerID    uuid.UUID
	CallerRoles []entity.UserRole
}

// UploadPictureOutput holds the updated cast member after a picture upload.
type UploadPictureOutput struct {
	Cast *entity.Cast
}

// UploadPicture uploads a picture for a cast member and persists the S3 key.
// Publishers may only upload pictures for cast on their own videos; admins are unrestricted.
func (s *Service) UploadPicture(ctx context.Context, input *UploadPictureInput) (*UploadPictureOutput, *apperror.AppError) {
	if s.storage == nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "storage not configured", nil)
	}

	c, err := s.castRepo.GetByID(ctx, input.ID)
	if err != nil {
		s.log.Error("getting cast for picture upload", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get cast member", err)
	}
	if c == nil {
		return nil, apperror.NotFound(apperror.CodeCastNotFound, "cast member not found", domain.ErrCastNotFound)
	}

	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) {
		video, err := s.videoRepo.GetByID(ctx, c.VideoID)
		if err != nil {
			s.log.Error("getting video for cast ownership check", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
		}
		if video == nil || video.UploadedBy != input.CallerID {
			return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", domain.ErrVideoOwnershipRequired)
		}
	}

	newKey := fmt.Sprintf("cast/%s/picture%s", c.ID, input.Ext)

	if _, err := s.storage.Upload(ctx, newKey, input.Body, input.ContentType); err != nil {
		s.log.Error("uploading cast picture", "error", err)
		return nil, apperror.Internal(apperror.CodeStorageError, "failed to upload picture", err)
	}

	oldKey := c.PictureKey
	c.PictureKey = newKey

	if err := s.castRepo.Update(ctx, c); err != nil {
		// Best-effort cleanup: delete the newly uploaded file so storage stays consistent.
		if delErr := s.storage.Delete(ctx, newKey); delErr != nil {
			s.log.Error("cleaning up picture after DB failure", "key", newKey, "error", delErr)
		}
		s.log.Error("updating cast picture key", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update cast member", err)
	}

	// Remove old picture from storage only after the DB update succeeds and the key differs.
	if oldKey != "" && oldKey != newKey {
		if delErr := s.storage.Delete(ctx, oldKey); delErr != nil {
			s.log.Error("deleting old cast picture", "key", oldKey, "error", delErr)
		}
	}

	s.resolvePictureURL(c)
	return &UploadPictureOutput{Cast: c}, nil
}
