package v1

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
)

const maxUploadSize = 50 << 20 // 50 MB

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
