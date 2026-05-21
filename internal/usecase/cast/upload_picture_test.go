package cast

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestUploadPicture(t *testing.T) {
	callerID := uuid.New()
	castID := uuid.New()
	videoID := uuid.New()
	otherUserID := uuid.New()

	const uploadedURL = "https://cdn.example.com/cast/picture.jpg"

	body := func() io.Reader { return strings.NewReader("fake-image-bytes") }

	cases := []struct {
		name       string
		input      *UploadPictureInput
		setupCast  func() *mockCastRepo
		setupVideo func() *mockVideoRepo
		setupStore func() *mockStorageSvc
		wantErr    string
		wantPicKey string
		wantPicURL string
	}{
		{
			name: "happy path - picture uploaded and URL resolved",
			input: &UploadPictureInput{
				ID:          castID,
				Body:        body(),
				ContentType: "image/jpeg",
				Ext:         ".jpg",
				CallerID:    callerID,
				CallerRoles: []entity.UserRole{entity.RolePublisher},
			},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Cast, error) {
						c := newCastMember(castID, videoID)
						return c, nil
					},
					updateFn: func(_ context.Context, c *entity.Cast) error {
						return nil
					},
				}
			},
			setupVideo: func() *mockVideoRepo {
				return &mockVideoRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
						return newActiveVideo(videoID, callerID), nil
					},
				}
			},
			setupStore: func() *mockStorageSvc {
				return &mockStorageSvc{
					uploadFn: func(_ context.Context, key string, _ io.Reader, _ string) (string, error) {
						return uploadedURL, nil
					},
					deleteFn:    func(_ context.Context, _ string) error { return nil },
					publicURLFn: func(key string) string { return uploadedURL },
				}
			},
			wantPicKey: "cast/" + castID.String() + "/picture.jpg",
			wantPicURL: uploadedURL,
		},
		{
			name: "happy path - old picture with different ext is deleted",
			input: &UploadPictureInput{
				ID:          castID,
				Body:        body(),
				ContentType: "image/png",
				Ext:         ".png",
				CallerID:    callerID,
				CallerRoles: []entity.UserRole{entity.RolePublisher},
			},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						c := newCastMember(castID, videoID)
						c.PictureKey = "cast/" + castID.String() + "/picture.jpg" // old key, different ext
						return c, nil
					},
					updateFn: func(_ context.Context, _ *entity.Cast) error { return nil },
				}
			},
			setupVideo: func() *mockVideoRepo {
				return &mockVideoRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
						return newActiveVideo(videoID, callerID), nil
					},
				}
			},
			setupStore: func() *mockStorageSvc {
				deleted := ""
				return &mockStorageSvc{
					uploadFn: func(_ context.Context, key string, _ io.Reader, _ string) (string, error) {
						return uploadedURL, nil
					},
					deleteFn: func(_ context.Context, key string) error {
						deleted = key
						_ = deleted
						return nil
					},
					publicURLFn: func(key string) string { return uploadedURL },
				}
			},
			wantPicKey: "cast/" + castID.String() + "/picture.png",
			wantPicURL: uploadedURL,
		},
		{
			name: "admin bypasses ownership check",
			input: &UploadPictureInput{
				ID:          castID,
				Body:        body(),
				ContentType: "image/jpeg",
				Ext:         ".jpg",
				CallerID:    uuid.New(), // different user — but admin
				CallerRoles: []entity.UserRole{entity.RoleAdmin},
			},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return newCastMember(castID, videoID), nil
					},
					updateFn: func(_ context.Context, _ *entity.Cast) error { return nil },
				}
			},
			setupVideo: func() *mockVideoRepo {
				return &mockVideoRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
						return newActiveVideo(videoID, callerID), nil
					},
				}
			},
			setupStore: func() *mockStorageSvc {
				return &mockStorageSvc{
					uploadFn:    func(_ context.Context, _ string, _ io.Reader, _ string) (string, error) { return uploadedURL, nil },
					deleteFn:    func(_ context.Context, _ string) error { return nil },
					publicURLFn: func(_ string) string { return uploadedURL },
				}
			},
			wantPicKey: "cast/" + castID.String() + "/picture.jpg",
			wantPicURL: uploadedURL,
		},
		{
			name: "cast not found returns 404",
			input: &UploadPictureInput{
				ID:          castID,
				Body:        body(),
				ContentType: "image/jpeg",
				Ext:         ".jpg",
				CallerID:    callerID,
				CallerRoles: []entity.UserRole{entity.RolePublisher},
			},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return nil, nil
					},
				}
			},
			setupVideo: func() *mockVideoRepo { return &mockVideoRepo{} },
			setupStore: func() *mockStorageSvc { return &mockStorageSvc{} },
			wantErr:    apperror.CodeCastNotFound,
		},
		{
			name: "ownership denied returns 403",
			input: &UploadPictureInput{
				ID:          castID,
				Body:        body(),
				ContentType: "image/jpeg",
				Ext:         ".jpg",
				CallerID:    otherUserID, // not the owner
				CallerRoles: []entity.UserRole{entity.RolePublisher},
			},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return newCastMember(castID, videoID), nil
					},
				}
			},
			setupVideo: func() *mockVideoRepo {
				return &mockVideoRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
						return newActiveVideo(videoID, callerID), nil // owned by callerID, not otherUserID
					},
				}
			},
			setupStore: func() *mockStorageSvc { return &mockStorageSvc{} },
			wantErr:    apperror.CodeVideoOwnershipRequired,
		},
		{
			name: "storage upload failure returns storage error",
			input: &UploadPictureInput{
				ID:          castID,
				Body:        body(),
				ContentType: "image/jpeg",
				Ext:         ".jpg",
				CallerID:    callerID,
				CallerRoles: []entity.UserRole{entity.RolePublisher},
			},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return newCastMember(castID, videoID), nil
					},
				}
			},
			setupVideo: func() *mockVideoRepo {
				return &mockVideoRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
						return newActiveVideo(videoID, callerID), nil
					},
				}
			},
			setupStore: func() *mockStorageSvc {
				return &mockStorageSvc{
					uploadFn: func(_ context.Context, _ string, _ io.Reader, _ string) (string, error) {
						return "", errors.New("s3 error")
					},
					deleteFn:    func(_ context.Context, _ string) error { return nil },
					publicURLFn: func(_ string) string { return "" },
				}
			},
			wantErr: apperror.CodeStorageError,
		},
		{
			name: "storage not configured returns internal error",
			input: &UploadPictureInput{
				ID:          castID,
				Body:        body(),
				ContentType: "image/jpeg",
				Ext:         ".jpg",
				CallerID:    callerID,
				CallerRoles: []entity.UserRole{entity.RolePublisher},
			},
			setupCast: func() *mockCastRepo {
				return &mockCastRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Cast, error) {
						return newCastMember(castID, videoID), nil
					},
				}
			},
			setupVideo: func() *mockVideoRepo {
				return &mockVideoRepo{
					getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
						return newActiveVideo(videoID, callerID), nil
					},
				}
			},
			setupStore: nil, // no storage configured
			wantErr:    apperror.CodeInternalError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var storageSvc storage.StorageService
			if tc.setupStore != nil {
				storageSvc = tc.setupStore()
			}
			svc := newTestService(tc.setupCast(), tc.setupVideo(), storageSvc)

			out, appErr := svc.UploadPicture(context.Background(), tc.input)

			if tc.wantErr != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.wantErr, appErr.Code)
				return
			}

			require.Nil(t, appErr)
			require.NotNil(t, out)
			assert.Equal(t, tc.wantPicKey, out.Cast.PictureKey)
			assert.Equal(t, tc.wantPicURL, out.Cast.PictureURL)
		})
	}
}
