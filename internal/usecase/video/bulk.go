package video

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// BulkFailure describes a single failure within a bulk operation.
type BulkFailure struct {
	ID    uuid.UUID
	Error string
}

// BulkResult holds the outcome of a bulk operation.
type BulkResult struct {
	Succeeded []uuid.UUID
	Failed    []BulkFailure
}

// BulkPublish publishes multiple videos, collecting per-video successes and failures.
func (s *Service) BulkPublish(ctx context.Context, ids []uuid.UUID) (*BulkResult, *apperror.AppError) {
	if len(ids) == 0 {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "at least one video ID is required", nil)
	}

	result := &BulkResult{}
	for _, id := range ids {
		if _, err := s.Publish(ctx, id); err != nil {
			result.Failed = append(result.Failed, BulkFailure{ID: id, Error: err.Message})
		} else {
			result.Succeeded = append(result.Succeeded, id)
		}
	}

	return result, nil
}

// BulkDelete archives multiple videos, collecting per-video successes and failures.
func (s *Service) BulkDelete(ctx context.Context, ids []uuid.UUID) (*BulkResult, *apperror.AppError) {
	if len(ids) == 0 {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "at least one video ID is required", nil)
	}

	result := &BulkResult{}
	for _, id := range ids {
		if err := s.Delete(ctx, id); err != nil {
			result.Failed = append(result.Failed, BulkFailure{ID: id, Error: err.Message})
		} else {
			result.Succeeded = append(result.Succeeded, id)
		}
	}

	return result, nil
}

// BulkUnarchive restores multiple archived videos, collecting per-video successes and failures.
func (s *Service) BulkUnarchive(ctx context.Context, ids []uuid.UUID) (*BulkResult, *apperror.AppError) {
	if len(ids) == 0 {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "at least one video ID is required", nil)
	}

	result := &BulkResult{}
	for _, id := range ids {
		if _, err := s.Unarchive(ctx, id); err != nil {
			result.Failed = append(result.Failed, BulkFailure{ID: id, Error: err.Message})
		} else {
			result.Succeeded = append(result.Succeeded, id)
		}
	}

	return result, nil
}
