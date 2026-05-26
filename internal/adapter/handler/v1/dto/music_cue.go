package dto

// MusicCueResponse represents a music cue in API responses.
type MusicCueResponse struct {
	ID                   string `json:"id"`
	VideoID              string `json:"video_id"`
	TimecodeSeconds      int    `json:"timecode_seconds" example:"252"`
	Title                string `json:"title" example:"Cold Open"`
	ComposerArtist       string `json:"composer_artist" example:"Sasha Lin"`
	UseType              string `json:"use_type" example:"original_score"`
	RightsStatus         string `json:"rights_status" example:"owned"`
	ClearanceDocumentKey string `json:"clearance_document_key,omitempty"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

// CreateMusicCueRequest is the body for POST /publisher/videos/{id}/music-cues.
type CreateMusicCueRequest struct {
	TimecodeSeconds int    `json:"timecode_seconds" validate:"gte=0"`
	Title           string `json:"title" validate:"required,min=1,max=500"`
	ComposerArtist  string `json:"composer_artist,omitempty" validate:"omitempty,max=500"`
	UseType         string `json:"use_type,omitempty" validate:"omitempty,oneof=original_score needle_drop sync_license background"`
	RightsStatus    string `json:"rights_status,omitempty" validate:"omitempty,oneof=owned pending_clearance cleared rejected"`
}

// UpdateMusicCueRequest is the body for PUT /publisher/videos/{id}/music-cues/{cueID}.
type UpdateMusicCueRequest struct {
	TimecodeSeconds *int    `json:"timecode_seconds,omitempty" validate:"omitempty,gte=0"`
	Title           *string `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	ComposerArtist  *string `json:"composer_artist,omitempty" validate:"omitempty,max=500"`
	UseType         *string `json:"use_type,omitempty" validate:"omitempty,oneof=original_score needle_drop sync_license background"`
	RightsStatus    *string `json:"rights_status,omitempty" validate:"omitempty,oneof=owned pending_clearance cleared rejected"`
}

// InitiateClearanceUploadResponse contains the signed PUT URL for a clearance document.
type InitiateClearanceUploadResponse struct {
	UploadURL string `json:"upload_url"`
	ObjectKey string `json:"object_key"`
	ExpiresAt string `json:"expires_at"`
}

// CompleteClearanceUploadRequest records the object key after a successful upload.
type CompleteClearanceUploadRequest struct {
	ObjectKey string `json:"object_key" validate:"required"`
}
