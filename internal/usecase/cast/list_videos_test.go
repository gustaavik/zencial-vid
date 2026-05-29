package cast

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
)

func TestListVideos(t *testing.T) {
	castID := uuid.New()
	videoID := uuid.New()
	now := time.Now().UTC()

	makeCredit := func() entity.VideoCast {
		return entity.VideoCast{
			ID:        uuid.New(),
			VideoID:   videoID,
			CastID:    castID,
			Role:      "actor",
			SortOrder: 0,
			Video: &entity.Video{
				ID:    videoID,
				Title: "Test Video",
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	cases := []struct {
		name         string
		input        ListVideosInput
		setupCast    func() *mockCastRepo
		setupVidCast func() *mockVideoCastRepo
		wantErr      string
		wantErrCode  string
		wantTotal    int
		wantCredits  int
		wantPage     int
		wantPerPage  int
	}{
		{
			name:  "happy path - returns paginated credits",
			input: ListVideosInput{CastID: castID, Page: 1, PerPage: 10},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return newCastMember(castID), nil
					},
				}
			},
			setupVidCast: func() *mockVideoCastRepo {
				return &mockVideoCastRepo{
					listByCastFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]entity.VideoCast, int, error) {
						return []entity.VideoCast{makeCredit()}, 1, nil
					},
				}
			},
			wantTotal:   1,
			wantCredits: 1,
			wantPage:    1,
			wantPerPage: 10,
		},
		{
			name:  "empty result - zero total, empty slice",
			input: ListVideosInput{CastID: castID, Page: 1, PerPage: 20},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return newCastMember(castID), nil
					},
				}
			},
			setupVidCast: func() *mockVideoCastRepo {
				return &mockVideoCastRepo{
					listByCastFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]entity.VideoCast, int, error) {
						return nil, 0, nil
					},
				}
			},
			wantTotal:   0,
			wantCredits: 0,
			wantPage:    1,
			wantPerPage: 20,
		},
		{
			name:  "page and per_page clamped to defaults when zero",
			input: ListVideosInput{CastID: castID, Page: 0, PerPage: 0},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return newCastMember(castID), nil
					},
				}
			},
			setupVidCast: func() *mockVideoCastRepo {
				return &mockVideoCastRepo{
					listByCastFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]entity.VideoCast, int, error) {
						return nil, 0, nil
					},
				}
			},
			wantPage:    1,
			wantPerPage: 20, // DefaultPerPage
		},
		{
			name:  "per_page clamped to max when too large",
			input: ListVideosInput{CastID: castID, Page: 1, PerPage: 999},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return newCastMember(castID), nil
					},
				}
			},
			setupVidCast: func() *mockVideoCastRepo {
				return &mockVideoCastRepo{
					listByCastFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]entity.VideoCast, int, error) {
						return nil, 0, nil
					},
				}
			},
			wantPage:    1,
			wantPerPage: 100, // MaxPerPage
		},
		{
			name:  "cast not found - returns 404",
			input: ListVideosInput{CastID: castID, Page: 1, PerPage: 20},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return nil, nil
					},
				}
			},
			setupVidCast: func() *mockVideoCastRepo { return &mockVideoCastRepo{} },
			wantErr:      "cast member not found",
			wantErrCode:  apperror.CodeCastNotFound,
		},
		{
			name:  "GetByID repo error - returns 500",
			input: ListVideosInput{CastID: castID, Page: 1, PerPage: 20},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return nil, errors.New("db error")
					},
				}
			},
			setupVidCast: func() *mockVideoCastRepo { return &mockVideoCastRepo{} },
			wantErr:      "failed to get cast member",
			wantErrCode:  apperror.CodeInternalError,
		},
		{
			name:  "ListByCast repo error - returns 500",
			input: ListVideosInput{CastID: castID, Page: 1, PerPage: 20},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return newCastMember(castID), nil
					},
				}
			},
			setupVidCast: func() *mockVideoCastRepo {
				return &mockVideoCastRepo{
					listByCastFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]entity.VideoCast, int, error) {
						return nil, 0, errors.New("db error")
					},
				}
			},
			wantErr:     "failed to list videos",
			wantErrCode: apperror.CodeInternalError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestService(tc.setupCast(), tc.setupVidCast(), &mockVideoRepo{}, nil)

			out, appErr := svc.ListVideos(context.Background(), tc.input)

			if tc.wantErr != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.wantErrCode, appErr.Code)
				assert.Contains(t, appErr.Message, tc.wantErr)
				return
			}

			require.Nil(t, appErr)
			require.NotNil(t, out)
			assert.Equal(t, tc.wantPage, out.Page)
			assert.Equal(t, tc.wantPerPage, out.PerPage)
			assert.Equal(t, tc.wantTotal, out.Total)
			assert.Len(t, out.Credits, tc.wantCredits)
		})
	}
}
