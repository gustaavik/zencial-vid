package genre

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// BulkCreateFailure describes a single failure within a bulk create operation.
type BulkCreateFailure struct {
	Slug  string
	Error string
}

// BulkCreateResult holds the outcome of a bulk create operation.
type BulkCreateResult struct {
	Succeeded []*entity.Genre
	Failed    []BulkCreateFailure
}

// BulkDeleteFailure describes a single failure within a bulk delete operation.
type BulkDeleteFailure struct {
	ID    uuid.UUID
	Error string
}

// BulkDeleteResult holds the outcome of a bulk delete operation.
type BulkDeleteResult struct {
	Succeeded []uuid.UUID
	Failed    []BulkDeleteFailure
}

// BulkCreate creates multiple genres, collecting per-genre successes and failures.
func (s *Service) BulkCreate(ctx context.Context, inputs []CreateInput) (*BulkCreateResult, *apperror.AppError) {
	if len(inputs) == 0 {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "at least one genre is required", nil)
	}

	result := &BulkCreateResult{}
	for _, input := range inputs {
		genre, err := s.Create(ctx, input)
		if err != nil {
			result.Failed = append(result.Failed, BulkCreateFailure{Slug: input.Slug, Error: err.Message})
		} else {
			result.Succeeded = append(result.Succeeded, genre)
		}
	}

	return result, nil
}

// BulkDelete removes multiple genres, collecting per-genre successes and failures.
func (s *Service) BulkDelete(ctx context.Context, ids []uuid.UUID) (*BulkDeleteResult, *apperror.AppError) {
	if len(ids) == 0 {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "at least one genre ID is required", nil)
	}

	result := &BulkDeleteResult{}
	for _, id := range ids {
		if err := s.Delete(ctx, id); err != nil {
			result.Failed = append(result.Failed, BulkDeleteFailure{ID: id, Error: err.Message})
		} else {
			result.Succeeded = append(result.Succeeded, id)
		}
	}

	return result, nil
}
