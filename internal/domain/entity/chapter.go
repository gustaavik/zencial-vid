package entity

import (
	"time"

	"github.com/google/uuid"
)

// ChapterSource indicates whether a chapter was auto-detected or manually entered.
type ChapterSource string

const (
	ChapterSourceAuto   ChapterSource = "auto"
	ChapterSourceManual ChapterSource = "manual"
)

// Chapter marks a named segment within a video.
type Chapter struct {
	ID            uuid.UUID
	VideoID       uuid.UUID
	StartTimeSecs int
	Title         string
	Source        ChapterSource
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewChapter creates a manual chapter entry for a video.
func NewChapter(videoID uuid.UUID, startTimeSecs int, title string) *Chapter {
	now := time.Now().UTC()
	return &Chapter{
		ID:            uuid.New(),
		VideoID:       videoID,
		StartTimeSecs: startTimeSecs,
		Title:         title,
		Source:        ChapterSourceManual,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
