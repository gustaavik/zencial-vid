package dto

// CaptionResponse represents a caption/subtitle record.
type CaptionResponse struct {
	ID           string `json:"id"`
	VideoID      string `json:"video_id"`
	LanguageCode string `json:"language_code" example:"en"`
	Format       string `json:"format" example:"webvtt"`
	StorageKey   string `json:"storage_key,omitempty"`
	Status       string `json:"status" example:"auto_generated"`
	Source       string `json:"source" example:"auto"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// InitiateCaptionUploadRequest is the body for POST /publisher/videos/{id}/captions.
type InitiateCaptionUploadRequest struct {
	LanguageCode string `json:"language_code" validate:"required,min=2,max=10" example:"es"`
	Format       string `json:"format,omitempty" validate:"omitempty,oneof=webvtt srt" example:"webvtt"`
}

// InitiateCaptionUploadResponse contains the signed PUT URL for caption upload.
type InitiateCaptionUploadResponse struct {
	UploadURL    string `json:"upload_url"`
	ObjectKey    string `json:"object_key"`
	LanguageCode string `json:"language_code"`
	ExpiresAt    string `json:"expires_at"`
}

// RegisterCaptionRequest completes a caption upload by registering the storage key.
type RegisterCaptionRequest struct {
	LanguageCode string `json:"language_code" validate:"required,min=2,max=10"`
	Format       string `json:"format,omitempty" validate:"omitempty,oneof=webvtt srt"`
	StorageKey   string `json:"storage_key" validate:"required"`
	Source       string `json:"source,omitempty" validate:"omitempty,oneof=auto manual"`
}
