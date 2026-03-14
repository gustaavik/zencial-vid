package v1

import (
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/postgres"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	videouc "github.com/zenfulcode/zencial/internal/usecase/video"
)

// VideoHandler handles video HTTP requests.
type VideoHandler struct {
	videoService *videouc.Service
	storage      storage.StorageService
	validator    *validator.Validator
}

// NewVideoHandler creates a new VideoHandler.
func NewVideoHandler(videoService *videouc.Service, storage storage.StorageService) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
		storage:      storage,
		validator:    validator.New(),
	}
}

// Upload handles video file upload with metadata via multipart form.
func (h *VideoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (limit: 500MB)
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "video file is required")
		return
	}
	defer file.Close()

	// Optional thumbnail file
	var thumbnailReader io.Reader
	var thumbnailFileName string
	var thumbnailContentType string
	thumbnailFile, thumbnailHeader, thumbnailErr := r.FormFile("thumbnail")
	if thumbnailErr == nil {
		defer thumbnailFile.Close()
		thumbnailReader = thumbnailFile
		thumbnailFileName = thumbnailHeader.Filename
		thumbnailContentType = thumbnailHeader.Header.Get("Content-Type")
		if thumbnailContentType == "" {
			thumbnailContentType = "image/jpeg"
		}
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	// Parse genre IDs from repeated form values
	var genreIDs []uuid.UUID
	for _, gid := range r.MultipartForm.Value["genre_ids"] {
		parsed, err := uuid.Parse(gid)
		if err != nil {
			httputil.BadRequest(w, apperror.CodeValidationFailed, "invalid genre ID: "+gid)
			return
		}
		genreIDs = append(genreIDs, parsed)
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}

	title := r.FormValue("title")
	if title == "" {
		httputil.BadRequest(w, apperror.CodeValidationFailed, "title is required")
		return
	}

	video, appErr := h.videoService.Upload(r.Context(), videouc.UploadInput{
		Title:                title,
		Description:          r.FormValue("description"),
		Creator:              r.FormValue("creator"),
		ContentRating:        r.FormValue("content_rating"),
		Quality:              r.FormValue("quality"),
		GenreIDs:             genreIDs,
		File:                 file,
		FileName:             header.Filename,
		ContentType:          contentType,
		FileSize:             header.Size,
		UploadedBy:           userID,
		Thumbnail:            thumbnailReader,
		ThumbnailFileName:    thumbnailFileName,
		ThumbnailContentType: thumbnailContentType,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.VideoToResponse(video, h.storage))
}

// GetByID returns a video by ID.
func (h *VideoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	video, appErr := h.videoService.GetByID(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(video, h.storage))
}

// ListPublished returns a paginated list of published videos (public endpoint).
func (h *VideoHandler) ListPublished(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.VideoFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	videos, total, appErr := h.videoService.ListPublished(r.Context(), fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.VideosToResponse(videos, h.storage), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// ListAll returns a paginated list of all videos (admin endpoint).
func (h *VideoHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.VideoFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	videos, total, appErr := h.videoService.List(r.Context(), fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.VideosToResponse(videos, h.storage), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// Update updates video metadata.
func (h *VideoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req dto.UpdateVideoRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil),
			errors,
		)
		return
	}

	// Parse genre IDs from string to UUID
	var genreIDs []uuid.UUID
	if req.GenreIDs != nil {
		genreIDs = make([]uuid.UUID, len(req.GenreIDs))
		for i, gid := range req.GenreIDs {
			parsed, err := uuid.Parse(gid)
			if err != nil {
				httputil.BadRequest(w, apperror.CodeValidationFailed, "invalid genre ID: "+gid)
				return
			}
			genreIDs[i] = parsed
		}
	}

	video, appErr := h.videoService.Update(r.Context(), videouc.UpdateInput{
		ID:            id,
		Title:         req.Title,
		Description:   req.Description,
		Creator:       req.Creator,
		ContentRating: req.ContentRating,
		Quality:       req.Quality,
		GenreIDs:      genreIDs,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(video, h.storage))
}

// Publish sets a video's status to published.
func (h *VideoHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	video, appErr := h.videoService.Publish(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(video, h.storage))
}

// Unarchive restores a soft-deleted video back to draft status.
func (h *VideoHandler) Unarchive(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	video, appErr := h.videoService.Unarchive(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(video, h.storage))
}

// Delete removes a video.
func (h *VideoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	appErr := h.videoService.Delete(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Stream returns a presigned URL for streaming a video.
func (h *VideoHandler) Stream(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	output, appErr := h.videoService.GetStreamURL(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.StreamToResponse(output))
}

// UploadThumbnail handles updating a video's thumbnail image.
func (h *VideoHandler) UploadThumbnail(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "thumbnail file is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	video, appErr := h.videoService.UpdateThumbnail(r.Context(), videouc.UpdateThumbnailInput{
		VideoID:     id,
		File:        file,
		FileName:    header.Filename,
		ContentType: contentType,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(video, h.storage))
}
