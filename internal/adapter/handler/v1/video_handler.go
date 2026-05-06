package v1

import (
	"context"
	"fmt"
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
	subscriptionuc "github.com/zenfulcode/zencial/internal/usecase/subscription"
	videouc "github.com/zenfulcode/zencial/internal/usecase/video"
)

// VideoHandler handles video HTTP requests.
type VideoHandler struct {
	videoService *videouc.Service
	subService   *subscriptionuc.Service
	storage      storage.StorageService
	validator    *validator.Validator
}

// NewVideoHandler creates a new VideoHandler.
func NewVideoHandler(videoService *videouc.Service, subService *subscriptionuc.Service, storageSvc storage.StorageService) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
		subService:   subService,
		storage:      storageSvc,
		validator:    validator.New(),
	}
}

// resolveUserPlanLevel returns the requester's active plan level, or nil if
// unauthenticated, no active subscription, or lookup error (treat as locked).
func (h *VideoHandler) resolveUserPlanLevel(ctx context.Context) *int {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil
	}
	swp, appErr := h.subService.GetActiveByUserID(ctx, userID)
	if appErr != nil || swp == nil || swp.Plan == nil {
		return nil
	}
	level := swp.Plan.Level
	return &level
}

// Upload godoc
// @Summary      Upload video
// @Description  Upload a video file with metadata via multipart form (admin only). Optional thumbnail can be sent in the same request.
// @Tags         videos
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Video file"
// @Param        thumbnail formData file false "Thumbnail image"
// @Param        title formData string true "Video title"
// @Param        description formData string false "Video description"
// @Param        creator formData string false "Creator name"
// @Param        content_rating formData string false "Content rating (G, PG, PG13, R, NC17)"
// @Param        genre_ids formData []string false "Genre UUIDs (repeatable)" collectionFormat(multi)
// @Param        minimum_plan_level formData int false "Minimum plan level required to watch"
// @Success      201 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      413 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos [post]
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
	defer func() { _ = file.Close() }()

	// Optional thumbnail file
	var thumbnailReader io.Reader
	var thumbnailFileName string
	var thumbnailContentType string
	thumbnailFile, thumbnailHeader, thumbnailErr := r.FormFile("thumbnail")
	if thumbnailErr == nil {
		defer func() { _ = thumbnailFile.Close() }()
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

	// Parse optional minimum plan level
	var minimumPlanLevel *int
	if mpl := r.FormValue("minimum_plan_level"); mpl != "" {
		var level int
		if _, err := fmt.Sscanf(mpl, "%d", &level); err == nil {
			minimumPlanLevel = &level
		}
	}

	video, appErr := h.videoService.Upload(r.Context(), &videouc.UploadInput{
		Title:                title,
		Description:          r.FormValue("description"),
		Creator:              r.FormValue("creator"),
		ContentRating:        r.FormValue("content_rating"),
		GenreIDs:             genreIDs,
		MinimumPlanLevel:     minimumPlanLevel,
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

	httputil.Success(w, http.StatusCreated, mapper.VideoToResponse(r.Context(), video, h.storage))
}

// GetByID godoc
// @Summary      Get video by ID
// @Description  Return a single video by its UUID. With a valid bearer token the response includes is_accessible based on the user's plan.
// @Tags         videos
// @Produce      json
// @Param        id path string true "Video ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /videos/{id} [get]
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

	planLevel := h.resolveUserPlanLevel(r.Context())
	httputil.Success(w, http.StatusOK, mapper.VideoToResponseWithAccess(r.Context(), video, h.storage, planLevel))
}

// ListPublished godoc
// @Summary      List published videos
// @Description  Return a paginated list of published videos. With a valid bearer token, is_accessible is populated per item based on plan level.
// @Tags         videos
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Param        sort query string false "Sort field (e.g. -created_at)"
// @Param        genre_id query string false "Filter by genre UUID" format(uuid)
// @Success      200 {object} httputil.Response{data=[]dto.VideoResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /videos [get]
func (h *VideoHandler) ListPublished(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.VideoFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	videos, total, appErr := h.videoService.ListPublished(r.Context(), &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	planLevel := h.resolveUserPlanLevel(r.Context())
	httputil.SuccessWithMeta(w, mapper.VideosToResponseWithAccess(r.Context(), videos, h.storage, planLevel), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// ListAll godoc
// @Summary      List all videos (admin)
// @Description  Return a paginated list of videos in any status, including drafts and archived (admin only)
// @Tags         videos
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Param        sort query string false "Sort field (e.g. -created_at)"
// @Param        status query string false "Filter by status (draft, published, archived)"
// @Success      200 {object} httputil.Response{data=[]dto.VideoResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos [get]
func (h *VideoHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.VideoFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	videos, total, appErr := h.videoService.List(r.Context(), &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.VideosToResponse(r.Context(), videos, h.storage), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// Update godoc
// @Summary      Update video metadata
// @Description  Update an existing video's metadata (admin only)
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        id path string true "Video ID" format(uuid)
// @Param        body body dto.UpdateVideoRequest true "Video fields to update"
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{id} [put]
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

	video, appErr := h.videoService.Update(r.Context(), &videouc.UpdateInput{
		ID:               id,
		Title:            req.Title,
		Description:      req.Description,
		Creator:          req.Creator,
		ContentRating:    req.ContentRating,
		GenreIDs:         genreIDs,
		MinimumPlanLevel: req.MinimumPlanLevel,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.storage))
}

// Publish godoc
// @Summary      Publish video
// @Description  Transition a video to the published state (admin only)
// @Tags         videos
// @Produce      json
// @Param        id path string true "Video ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{id}/publish [post]
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

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.storage))
}

// Unarchive godoc
// @Summary      Unarchive video
// @Description  Restore a soft-deleted (archived) video back to draft status (admin only)
// @Tags         videos
// @Produce      json
// @Param        id path string true "Video ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{id}/unarchive [post]
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

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.storage))
}

// Delete godoc
// @Summary      Archive video
// @Description  Soft-delete a video by archiving it. Files are moved to a deleted/ prefix and the status changes to archived (admin only).
// @Tags         videos
// @Param        id path string true "Video ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{id} [delete]
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

// Stream godoc
// @Summary      Get video stream URL
// @Description  Return a presigned URL (or HLS manifest URL) the authenticated user can use to stream the video.
// @Tags         videos
// @Produce      json
// @Param        id path string true "Video ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.VideoStreamResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{id}/stream [get]
func (h *VideoHandler) Stream(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	output, appErr := h.videoService.GetStreamURL(r.Context(), id, userID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.StreamToResponse(output))
}

// BulkPublish godoc
// @Summary      Bulk publish videos
// @Description  Publish multiple videos in a single request (admin only). Returns succeeded and failed entries.
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        body body dto.BulkVideoIDsRequest true "Video IDs to publish"
// @Success      200 {object} httputil.Response{data=dto.BulkResultResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/bulk-publish [post]
func (h *VideoHandler) BulkPublish(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkVideoIDsRequest
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

	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeValidationFailed, err.Error())
		return
	}

	result, appErr := h.videoService.BulkPublish(r.Context(), ids)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.BulkResultToResponse(result))
}

// BulkDelete godoc
// @Summary      Bulk archive videos
// @Description  Archive multiple videos in a single request (admin only). Returns succeeded and failed entries.
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        body body dto.BulkVideoIDsRequest true "Video IDs to archive"
// @Success      200 {object} httputil.Response{data=dto.BulkResultResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/bulk-archive [post]
func (h *VideoHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkVideoIDsRequest
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

	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeValidationFailed, err.Error())
		return
	}

	result, appErr := h.videoService.BulkDelete(r.Context(), ids)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.BulkResultToResponse(result))
}

// BulkUnarchive godoc
// @Summary      Bulk unarchive videos
// @Description  Restore multiple archived videos in a single request (admin only). Returns succeeded and failed entries.
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        body body dto.BulkVideoIDsRequest true "Video IDs to unarchive"
// @Success      200 {object} httputil.Response{data=dto.BulkResultResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/bulk-unarchive [post]
func (h *VideoHandler) BulkUnarchive(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkVideoIDsRequest
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

	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeValidationFailed, err.Error())
		return
	}

	result, appErr := h.videoService.BulkUnarchive(r.Context(), ids)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.BulkResultToResponse(result))
}

// parseUUIDs converts a slice of string IDs to uuid.UUID.
func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid ID: %s", id)
		}
		result[i] = parsed
	}
	return result, nil
}

// UploadThumbnail godoc
// @Summary      Update video thumbnail
// @Description  Replace a video's thumbnail image via multipart form (admin only)
// @Tags         videos
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "Video ID" format(uuid)
// @Param        thumbnail formData file true "Thumbnail image"
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{id}/thumbnail [put]
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
	defer func() { _ = file.Close() }()

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

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.storage))
}
