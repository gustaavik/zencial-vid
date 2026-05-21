package cast

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// ListAllInput carries pagination parameters for the global cast list.
type ListAllInput struct {
	Page    int
	PerPage int
}

// ListAllOutput holds the paginated result.
type ListAllOutput struct {
	Members []entity.Cast
	Total   int
	Page    int
	PerPage int
}

// ListAll returns a paginated list of all cast members ordered by name.
func (s *Service) ListAll(ctx context.Context, input ListAllInput) (*ListAllOutput, *apperror.AppError) {
	p := valueobject.NewPagination(input.Page, input.PerPage)

	members, total, err := s.castRepo.ListAll(ctx, p.Offset(), p.Limit())
	if err != nil {
		s.log.Error("listing all cast", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list cast", err)
	}

	for i := range members {
		s.resolvePictureURL(&members[i])
	}

	return &ListAllOutput{
		Members: members,
		Total:   total,
		Page:    p.Page,
		PerPage: p.PerPage,
	}, nil
}
