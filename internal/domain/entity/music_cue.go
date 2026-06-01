package entity

import (
	"time"

	"github.com/google/uuid"
)

// MusicUseType describes how a musical work is used.
type MusicUseType string

const (
	MusicUseOriginalScore MusicUseType = "original_score"
	MusicUseNeedleDrop    MusicUseType = "needle_drop"
	MusicUseSyncLicense   MusicUseType = "sync_license"
	MusicUseBackground    MusicUseType = "background"
)

// MusicRightsStatus tracks clearance state for a cue.
type MusicRightsStatus string

const (
	MusicRightsOwned            MusicRightsStatus = "owned"
	MusicRightsPendingClearance MusicRightsStatus = "pending_clearance"
	MusicRightsCleared          MusicRightsStatus = "cleared"
	MusicRightsRejected         MusicRightsStatus = "rejected"
)

// MusicCue records a single piece of music used in a video and its rights status.
type MusicCue struct {
	ID                   uuid.UUID
	VideoID              uuid.UUID
	TimecodeSeconds      int
	Title                string
	ComposerArtist       string
	UseType              MusicUseType
	RightsStatus         MusicRightsStatus
	ClearanceDocumentKey string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewMusicCue creates a new music cue entry for a video.
func NewMusicCue(videoID uuid.UUID, timecodeSeconds int, title, composerArtist string, useType MusicUseType) *MusicCue {
	now := time.Now().UTC()
	return &MusicCue{
		ID:              uuid.New(),
		VideoID:         videoID,
		TimecodeSeconds: timecodeSeconds,
		Title:           title,
		ComposerArtist:  composerArtist,
		UseType:         useType,
		RightsStatus:    MusicRightsOwned,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// IsBlockingSubmission reports whether this cue prevents submission.
func (m *MusicCue) IsBlockingSubmission() bool {
	return m.RightsStatus == MusicRightsPendingClearance
}

// AttachClearanceDocument records the S3 key for an uploaded clearance document.
func (m *MusicCue) AttachClearanceDocument(key string) {
	m.ClearanceDocumentKey = key
	m.RightsStatus = MusicRightsCleared
	m.UpdatedAt = time.Now().UTC()
}
