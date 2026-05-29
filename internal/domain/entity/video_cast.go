package entity

import (
	"time"

	"github.com/google/uuid"
)

// CastDepartment groups contributors by production department.
type CastDepartment string

const (
	DepartmentPerformance    CastDepartment = "performance"
	DepartmentDirection      CastDepartment = "direction"
	DepartmentCinematography CastDepartment = "cinematography"
	DepartmentSound          CastDepartment = "sound"
	DepartmentPost           CastDepartment = "post"
	DepartmentProduction     CastDepartment = "production"
	DepartmentWriting        CastDepartment = "writing"
	DepartmentVFX            CastDepartment = "vfx"
)

// CastInviteStatus tracks whether a contributor has accepted their credit.
type CastInviteStatus string

const (
	CastInviteNotInvited CastInviteStatus = "not_invited"
	CastInvitePending    CastInviteStatus = "pending"
	CastInviteAccepted   CastInviteStatus = "accepted"
)

// VideoCast represents a cast member's credit for a specific video
// (role, department, sort position, and invite status within that video).
type VideoCast struct {
	ID           uuid.UUID
	VideoID      uuid.UUID
	CastID       uuid.UUID
	Role         string
	Department   CastDepartment
	InviteStatus CastInviteStatus
	InvitedEmail *string
	SortOrder    int
	Cast         *Cast  // populated by repository JOIN for ListByVideo direction
	Video        *Video // populated by repository JOIN for ListByCast direction
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewVideoCast creates a new VideoCast record linking a cast member to a video.
func NewVideoCast(videoID, castID uuid.UUID, role string, department CastDepartment, sortOrder int) *VideoCast {
	now := time.Now().UTC()
	return &VideoCast{
		ID:           uuid.New(),
		VideoID:      videoID,
		CastID:       castID,
		Role:         role,
		Department:   department,
		InviteStatus: CastInviteNotInvited,
		SortOrder:    sortOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
