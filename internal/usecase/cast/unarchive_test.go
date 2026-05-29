package cast

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestUnarchive(t *testing.T) {
	adminRoles := []entity.UserRole{entity.RoleAdmin}
	publisherRoles := []entity.UserRole{entity.RolePublisher}

	cases := []struct {
		name      string
		roles     []entity.UserRole
		setupCast func(id uuid.UUID) *entity.Cast
		updateErr error
		wantErr   string
	}{
		{
			name:  "non-admin forbidden",
			roles: publisherRoles,
			setupCast: func(id uuid.UUID) *entity.Cast {
				return newArchivedCastMember(id)
			},
			wantErr: apperror.CodeForbidden,
		},
		{
			name:      "cast not found",
			roles:     adminRoles,
			setupCast: nil,
			wantErr:   apperror.CodeCastNotFound,
		},
		{
			name:  "cast not archived",
			roles: adminRoles,
			setupCast: func(id uuid.UUID) *entity.Cast {
				return newCastMember(id)
			},
			wantErr: apperror.CodeBadRequest,
		},
		{
			name:  "update failure",
			roles: adminRoles,
			setupCast: func(id uuid.UUID) *entity.Cast {
				return newArchivedCastMember(id)
			},
			updateErr: errors.New("db error"),
			wantErr:   apperror.CodeInternalError,
		},
		{
			name:  "success",
			roles: adminRoles,
			setupCast: func(id uuid.UUID) *entity.Cast {
				return newArchivedCastMember(id)
			},
			wantErr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			castID := uuid.New()

			castRepo := &mockCastRepo{
				getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Cast, error) {
					if tc.setupCast == nil {
						return nil, nil
					}
					return tc.setupCast(id), nil
				},
				updateFn: func(_ context.Context, _ *entity.Cast) error {
					return tc.updateErr
				},
			}

			svc := newTestService(castRepo, &mockVideoCastRepo{}, &mockVideoRepo{}, nil)
			out, appErr := svc.Unarchive(context.Background(), &UnarchiveInput{
				ID:          castID,
				CallerRoles: tc.roles,
			})

			if tc.wantErr == "" {
				require.Nil(t, appErr)
				require.NotNil(t, out)
				assert.Equal(t, entity.CastStatusActive, out.Cast.Status)
			} else {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.wantErr, appErr.Code)
				assert.Nil(t, out)
			}
		})
	}
}
