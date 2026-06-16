package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/domain/entity"
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
	cdnURLs      mapper.ThumbnailURLBuilder
	validator    *validator.Validator
}

// NewVideoHandler creates a new VideoHandler.
func NewVideoHandler(videoService *videouc.Service, subService *subscriptionuc.Service, storageSvc storage.StorageService, cdnURLs mapper.ThumbnailURLBuilder) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
		subService:   subService,
		storage:      storageSvc,
		cdnURLs:      cdnURLs,
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

// InitiateUpload godoc
// @Summary      Initiate video upload
// @Description  Returns a presigned PUT URL the admin client uses to upload the video binary directly to object storage. Bypasses CDN body-size limits. Follow up with POST /videos to commit the metadata.
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        body body dto.InitiateUploadRequest true "File metadata"
// @Success      200 {object} httputil.Response{data=dto.InitiateUploadResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/uploads [post]
func (h *VideoHandler) InitiateUpload(w http.ResponseWriter, r *http.Request) {
	var req dto.InitiateUploadRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil),
			errs,
		)
		return
	}

	out, appErr := h.videoService.InitiateUpload(r.Context(), &videouc.InitiateUploadInput{
		FileName:    req.FileName,
		ContentType: req.ContentType,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, dto.InitiateUploadResponse{
		UploadURL: out.UploadURL,
		ObjectKey: out.ObjectKey,
		ExpiresAt: out.ExpiresAt.Format(time.RFC3339),
	})
}

// CompleteUpload godoc
// @Summary      Commit a video upload
// @Description  Finalizes a video upload after the client has PUT the binary to the presigned URL returned by /videos/uploads. Verifies the object exists and creates the metadata record (admin only).
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        body body dto.CompleteUploadRequest true "Upload metadata"
// @Success      201 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos [post]
func (h *VideoHandler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	var req dto.CompleteUploadRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil),
			errs,
		)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	genreIDs := make([]uuid.UUID, 0, len(req.GenreIDs))
	for _, gid := range req.GenreIDs {
		parsed, err := uuid.Parse(gid)
		if err != nil {
			httputil.BadRequest(w, apperror.CodeValidationFailed, "invalid genre ID: "+gid)
			return
		}
		genreIDs = append(genreIDs, parsed)
	}

	video, appErr := h.videoService.CompleteUpload(r.Context(), &videouc.CompleteUploadInput{
		ObjectKey:        req.ObjectKey,
		Title:            req.Title,
		Description:      req.Description,
		Creator:          req.Creator,
		ContentRating:    req.ContentRating,
		GenreIDs:         genreIDs,
		UploadedBy:       userID,
		MinimumPlanLevel: req.MinimumPlanLevel,
		DurationSeconds:  req.DurationSeconds,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
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
	httputil.Success(w, http.StatusOK, mapper.VideoToResponseWithAccess(r.Context(), video, h.cdnURLs, planLevel))
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
	httputil.SuccessWithMeta(w, mapper.VideosToResponseWithAccess(r.Context(), videos, h.cdnURLs, planLevel), &httputil.Meta{
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
// @Param        submission_status query string false "Filter by submission status"
// @Param        content_rating query string false "Filter by content rating"
// @Param        title query string false "Search by title (substring match)"
// @Param        genre_id query string false "Filter by genre UUID" format(uuid)
// @Success      200 {object} httputil.Response{data=[]dto.VideoResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos [get]
func (h *VideoHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.AdminVideoFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	// genre_id is filtered via an EXISTS join, not the column whitelist.
	var genreID *uuid.UUID
	if g := r.URL.Query().Get("genre_id"); g != "" {
		parsed, perr := uuid.Parse(g)
		if perr != nil {
			httputil.BadRequest(w, apperror.CodeBadRequest, "invalid genre_id")
			return
		}
		genreID = &parsed
	}

	videos, total, appErr := h.videoService.ListAdmin(r.Context(), &fs, genreID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.VideosToResponse(r.Context(), videos, h.cdnURLs), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// Stats godoc
// @Summary      Catalog stats (admin)
// @Description  Returns platform-wide catalog aggregates: video counts grouped by status and submission status, plus a per-genre title count (admin only).
// @Tags         videos
// @Produce      json
// @Success      200 {object} httputil.Response{data=dto.AdminContentStatsResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/stats [get]
func (h *VideoHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, appErr := h.videoService.Stats(r.Context())
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoStatsToResponse(stats))
}

// BulkUpdate godoc
// @Summary      Bulk update videos (admin)
// @Description  Reassign category and/or change rating for multiple videos in a single request (admin only). Returns succeeded and failed entries.
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        body body dto.BulkUpdateVideosRequest true "Video IDs and fields to update"
// @Success      200 {object} httputil.Response{data=dto.BulkResultResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/bulk-update [post]
func (h *VideoHandler) BulkUpdate(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkUpdateVideosRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil),
			errs,
		)
		return
	}

	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeValidationFailed, err.Error())
		return
	}

	// A present (even empty) genre_ids array replaces the category set; an
	// absent one leaves categories untouched.
	var genreIDs []uuid.UUID
	if req.GenreIDs != nil {
		genreIDs, err = parseUUIDs(req.GenreIDs)
		if err != nil {
			httputil.BadRequest(w, apperror.CodeValidationFailed, err.Error())
			return
		}
	}

	callerID, _ := middleware.GetUserID(r.Context())
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	result, appErr := h.videoService.BulkUpdate(r.Context(), ids, videouc.BulkUpdateInput{
		GenreIDs:      genreIDs,
		ContentRating: req.ContentRating,
		CallerID:      callerID,
		CallerRoles:   callerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.BulkResultToResponse(result))
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

	callerID, _ := middleware.GetUserID(r.Context())
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	video, appErr := h.videoService.Update(r.Context(), &videouc.UpdateInput{
		ID:               id,
		Title:            req.Title,
		Description:      req.Description,
		Creator:          req.Creator,
		ContentRating:    req.ContentRating,
		GenreIDs:         genreIDs,
		MinimumPlanLevel: req.MinimumPlanLevel,
		CallerID:         callerID,
		CallerRoles:      callerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
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

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
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

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
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

	thumbCallerID, _ := middleware.GetUserID(r.Context())
	thumbCallerRoles, _ := middleware.GetUserRoles(r.Context())

	video, appErr := h.videoService.UpdateThumbnail(r.Context(), &videouc.UpdateThumbnailInput{
		VideoID:     id,
		File:        file,
		FileName:    header.Filename,
		ContentType: contentType,
		CallerID:    thumbCallerID,
		CallerRoles: thumbCallerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
}

// PurgeOrphans godoc
// @Summary      Purge orphaned video records and/or storage objects
// @Description  Phase 1 (always): hard-deletes DB rows whose storage_key file is absent in S3. Phase 2 (opt-in via include_s3_orphans): deletes S3 objects not referenced by any DB row. Use dry_run=true to preview without committing changes.
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        body body dto.PurgeOrphansRequest false "Purge options"
// @Success      200 {object} httputil.Response{data=dto.PurgeOrphansResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/purge-orphans [post]
func (h *VideoHandler) PurgeOrphans(w http.ResponseWriter, r *http.Request) {
	var req dto.PurgeOrphansRequest
	// Body is optional — ignore decode errors (empty body is valid).
	_ = httputil.DecodeJSON(r, &req)

	out, appErr := h.videoService.PurgeOrphans(r.Context(), videouc.PurgeOrphansInput{
		IncludeS3Orphans: req.IncludeS3Orphans,
		DryRun:           req.DryRun,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.PurgeOrphansToResponse(out, req.DryRun))
}

// ListOwned godoc
// @Summary      List publisher's own videos
// @Description  Return a paginated list of all videos uploaded by the authenticated publisher (any status)
// @Tags         publisher
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Param        sort query string false "Sort field (e.g. -created_at)"
// @Success      200 {object} httputil.Response{data=[]dto.VideoResponse,meta=httputil.Meta}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos [get]
func (h *VideoHandler) ListOwned(w http.ResponseWriter, r *http.Request) {
	uploaderID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	fs, err := filter.FromRequest(r, postgres.VideoFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	videos, total, appErr := h.videoService.ListOwned(r.Context(), uploaderID, &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.VideosToResponse(r.Context(), videos, h.cdnURLs), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// GetOwned godoc
// @Summary      Get publisher's own video
// @Description  Return a single video owned by the authenticated publisher, regardless of status. Admins may fetch any video.
// @Tags         publisher
// @Produce      json
// @Param        id path string true "Video ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id} [get]
func (h *VideoHandler) GetOwned(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	video, appErr := h.videoService.GetByID(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	callerRoles, _ := middleware.GetUserRoles(r.Context())
	if !entity.HasRole(callerRoles, entity.RoleAdmin) && video.UploadedBy != callerID {
		httputil.Error(w, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil))
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
}

// PublishOwned godoc
// @Summary      Publish publisher's own video
// @Description  Transition a publisher-owned video to the published state. Publishers can only publish their own videos.
// @Tags         publisher
// @Produce      json
// @Param        id path string true "Video ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/publish [post]
func (h *VideoHandler) PublishOwned(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	// Ownership check: non-admins may only publish their own videos.
	callerRoles, _ := middleware.GetUserRoles(r.Context())
	if !entity.HasRole(callerRoles, entity.RoleAdmin) {
		video, appErr := h.videoService.GetByID(r.Context(), id)
		if appErr != nil {
			httputil.Error(w, appErr)
			return
		}
		if video.UploadedBy != callerID {
			httputil.Error(w, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil))
			return
		}
	}

	video, appErr := h.videoService.Publish(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
}

// DeleteOwned godoc
// @Summary      Archive publisher's own video
// @Description  Soft-delete a publisher-owned video. Publishers can only archive their own videos.
// @Tags         publisher
// @Param        id path string true "Video ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id} [delete]
func (h *VideoHandler) DeleteOwned(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	// Ownership check: non-admins may only archive their own videos.
	callerRolesD, _ := middleware.GetUserRoles(r.Context())
	if !entity.HasRole(callerRolesD, entity.RoleAdmin) {
		video, appErr := h.videoService.GetByID(r.Context(), id)
		if appErr != nil {
			httputil.Error(w, appErr)
			return
		}
		if video.UploadedBy != callerID {
			httputil.Error(w, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil))
			return
		}
	}

	if appErr := h.videoService.Delete(r.Context(), id); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Preflight godoc
// @Summary      Pre-flight checklist
// @Description  Returns the submission readiness checklist for a video.
// @Tags         videos
// @Produce      json
// @Param        id path string true "Video ID"
// @Success      200 {object} httputil.Response{data=dto.PreflightResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/preflight [get]
func (h *VideoHandler) Preflight(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	result, appErr := h.videoService.Preflight(r.Context(), id, callerID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	items := make([]dto.PreflightItemResponse, len(result.Items))
	for i, item := range result.Items {
		items[i] = dto.PreflightItemResponse{
			Key:     item.Key,
			Label:   item.Label,
			Passed:  item.Passed,
			Blocker: item.Blocker,
		}
	}

	httputil.Success(w, http.StatusOK, dto.PreflightResponse{
		VideoID:    result.VideoID,
		ReadyCount: result.ReadyCount,
		TotalCount: result.TotalCount,
		Blockers:   result.Blockers,
		Items:      items,
	})
}

// Submit godoc
// @Summary      Submit video for review
// @Description  Locks a video and queues it for moderation. Validates pre-flight before locking.
// @Tags         videos
// @Produce      json
// @Param        id path string true "Video ID"
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/submit [post]
func (h *VideoHandler) Submit(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	video, appErr := h.videoService.SubmitForReview(r.Context(), id, callerID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
}

// ApproveSubmission godoc
// @Summary      Approve submission
// @Description  Approves a video submission and triggers the publish flow. Admin only.
// @Tags         videos
// @Produce      json
// @Param        id path string true "Video ID"
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/{id}/approve [post]
func (h *VideoHandler) ApproveSubmission(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	video, appErr := h.videoService.ApproveSubmission(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
}

// RejectSubmission godoc
// @Summary      Reject submission
// @Description  Rejects a video submission with moderator notes. Admin only.
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        id   path string                       true "Video ID"
// @Param        body body dto.RejectSubmissionRequest  true "Rejection reason"
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/{id}/reject [post]
func (h *VideoHandler) RejectSubmission(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req dto.RejectSubmissionRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil), errs)
		return
	}

	video, appErr := h.videoService.RejectSubmission(r.Context(), id, req.Notes)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
}

// GetFeatured godoc
// @Summary      List featured videos
// @Description  Return a paginated list of admin-designated featured published videos.
// @Tags         videos
// @Produce      json
// @Param        page     query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.VideoResponse,meta=httputil.Meta}
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /videos/featured [get]
func (h *VideoHandler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.VideoFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	videos, total, appErr := h.videoService.ListFeatured(r.Context(), &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	planLevel := h.resolveUserPlanLevel(r.Context())
	httputil.SuccessWithMeta(w, mapper.VideosToResponseWithAccess(r.Context(), videos, h.cdnURLs, planLevel), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// SetFeatured godoc
// @Summary      Feature a video
// @Description  Mark a published video as featured on the home page (admin only).
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        id   path string                    true "Video ID" format(uuid)
// @Param        body body dto.SetFeaturedRequest    false "Optional featured description"
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/{id}/feature [post]
func (h *VideoHandler) SetFeatured(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req dto.SetFeaturedRequest
	// Body is optional — ignore decode errors (empty body is valid).
	_ = httputil.DecodeJSON(r, &req)

	if appErr := h.videoService.SetFeatured(r.Context(), id, req.FeaturedDescription); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	video, appErr := h.videoService.GetByID(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
}

// UnsetFeatured godoc
// @Summary      Unfeature a video
// @Description  Remove a video from the featured list (admin only).
// @Tags         videos
// @Param        id path string true "Video ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/videos/{id}/feature [delete]
func (h *VideoHandler) UnsetFeatured(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	if appErr := h.videoService.UnsetFeatured(r.Context(), id); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ModerationQueue godoc
// @Summary      Moderation queue
// @Description  Returns videos pending moderation review. Without a submission_status filter, defaults to submitted and under_review. Admin only.
// @Tags         videos
// @Produce      json
// @Param        submission_status query string false "Filter by submission status" Enums(submitted, under_review, approved, rejected)
// @Success      200 {object} httputil.Response{data=[]dto.VideoResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/moderation/queue [get]
func (h *VideoHandler) ModerationQueue(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.VideoFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid filter parameters")
		return
	}

	// Default the queue to actionable submissions when no explicit
	// submission_status filter was requested.
	hasSubmissionFilter := false
	for _, c := range fs.Conditions {
		if c.DBColumn == "v.submission_status" {
			hasSubmissionFilter = true
			break
		}
	}
	if !hasSubmissionFilter {
		fs.Conditions = append(fs.Conditions, filter.Condition{
			DBColumn: "v.submission_status",
			Op:       filter.OpIn,
			Values: []any{
				string(entity.SubmissionStatusSubmitted),
				string(entity.SubmissionStatusUnderReview),
			},
		})
	}

	videos, total, appErr := h.videoService.List(r.Context(), &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.VideosToResponse(r.Context(), videos, h.cdnURLs),
		&httputil.Meta{Total: total},
	)
}
