package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
)

const maxUploadSize = 1000 << 20 // 1000 MB 1024 * 1 024 * 1 000

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
	"video/mp4":  true,
	"video/webm": true,
}

// UploadHandler handles file upload HTTP requests.
type UploadHandler struct {
	storage storage.StorageService
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(s storage.StorageService) *UploadHandler {
	return &UploadHandler{storage: s}
}

// UploadResponse is the response for a successful upload.
type UploadResponse struct {
	URL         string `json:"url"`
	Key         string `json:"key"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// InitUploadRequest is the request body for initiating a presigned upload.
type InitUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// InitUploadResponse is the response for a presigned upload initiation.
type InitUploadResponse struct {
	UploadURL   string `json:"upload_url"`
	Key         string `json:"key"`
	PublicURL   string `json:"public_url"`
	ContentType string `json:"content_type"`
	ExpiresIn   int    `json:"expires_in_seconds"`
}

// InitUpload godoc
// @Summary      Get a presigned URL for direct upload to storage
// @Description  Returns a presigned PUT URL so the client can upload directly to S3/MinIO
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        body body InitUploadRequest true "Upload metadata"
// @Success      200 {object} httputil.Response{data=InitUploadResponse}
// @Security     BearerAuth
// @Router       /admin/upload/init [post]
func (h *UploadHandler) InitUpload(w http.ResponseWriter, r *http.Request) {
	var req InitUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "Invalid JSON body")
		return
	}

	if req.Filename == "" {
		httputil.BadRequest(w, apperror.CodeBadRequest, "filename is required")
		return
	}
	if req.Size <= 0 {
		httputil.BadRequest(w, apperror.CodeBadRequest, "size must be greater than zero")
		return
	}
	if req.Size > maxUploadSize {
		httputil.BadRequest(w, "FILE_TOO_LARGE", fmt.Sprintf("File exceeds maximum size of %d MB", maxUploadSize>>20))
		return
	}

	// Resolve content type from the request or fall back to extension
	contentType := req.ContentType
	if contentType == "" {
		ext := strings.ToLower(filepath.Ext(req.Filename))
		switch ext {
		case ".mp4":
			contentType = "video/mp4"
		case ".webm":
			contentType = "video/webm"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".webp":
			contentType = "image/webp"
		case ".gif":
			contentType = "image/gif"
		default:
			httputil.BadRequest(w, "INVALID_FILE_TYPE", "Cannot determine content type; provide content_type in the request")
			return
		}
	}

	if !allowedMimeTypes[contentType] {
		httputil.BadRequest(w, "INVALID_FILE_TYPE", fmt.Sprintf("File type %q is not allowed. Allowed: jpeg, png, webp, gif, mp4, webm", contentType))
		return
	}

	id := uuid.New().String()
	sanitizedName := sanitizeFilename(req.Filename)
	key := fmt.Sprintf("uploads/%s/%s", id, sanitizedName)

	presigned, err := h.storage.GeneratePresignedUploadURL(r.Context(), key, contentType, req.Size)
	if err != nil {
		httputil.BadRequest(w, "PRESIGN_FAILED", "Failed to generate upload URL: "+err.Error())
		return
	}

	httputil.Success(w, http.StatusOK, InitUploadResponse{
		UploadURL:   presigned.URL,
		Key:         presigned.Key,
		PublicURL:   presigned.PublicURL,
		ContentType: contentType,
		ExpiresIn:   int(presigned.ExpiresIn.Seconds()),
	})
}

// Upload godoc
// @Summary      Upload a file
// @Description  Upload an image or video file to storage
// @Tags         admin
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "File to upload"
// @Success      200 {object} httputil.Response{data=UploadResponse}
// @Security     BearerAuth
// @Router       /admin/upload [post]
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Limit the request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		httputil.BadRequest(w, "FILE_TOO_LARGE", fmt.Sprintf("File exceeds maximum size of %d MB", maxUploadSize>>20))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.BadRequest(w, "INVALID_FILE", "No file provided or file could not be read")
		return
	}
	defer file.Close()

	// Detect content type from first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		httputil.BadRequest(w, "INVALID_FILE", "Could not read file")
		return
	}
	contentType := http.DetectContentType(buf[:n])

	// Also check by extension as a fallback for video
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == ".mp4" && !strings.HasPrefix(contentType, "video/") {
		contentType = "video/mp4"
	}
	if ext == ".webm" && !strings.HasPrefix(contentType, "video/") {
		contentType = "video/webm"
	}
	if ext == ".webp" && contentType == "application/octet-stream" {
		contentType = "image/webp"
	}

	if !allowedMimeTypes[contentType] {
		httputil.BadRequest(w, "INVALID_FILE_TYPE", fmt.Sprintf("File type %q is not allowed. Allowed: jpeg, png, webp, gif, mp4, webm", contentType))
		return
	}

	// Seek back to the beginning after reading for detection
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			httputil.BadRequest(w, "INVALID_FILE", "Could not process file")
			return
		}
	}

	// Generate a unique storage key
	id := uuid.New().String()
	sanitizedName := sanitizeFilename(header.Filename)
	key := fmt.Sprintf("uploads/%s/%s", id, sanitizedName)

	// Upload to storage
	url, err := h.storage.Upload(r.Context(), key, file, contentType)
	if err != nil {
		httputil.BadRequest(w, "UPLOAD_FAILED", "Failed to upload file: "+err.Error())
		return
	}

	httputil.Success(w, http.StatusOK, UploadResponse{
		URL:         url,
		Key:         key,
		ContentType: contentType,
		Size:        header.Size,
	})
}

// sanitizeFilename cleans a filename for safe storage.
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	// Keep only alphanumeric, hyphens, underscores, dots
	var clean strings.Builder
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			clean.WriteRune(c)
		}
	}
	result := clean.String()
	if result == "" {
		result = "file"
	}
	return result
}
