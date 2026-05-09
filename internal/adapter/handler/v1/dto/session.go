package dto

// SessionResponse represents a single user session in API responses. The
// Current flag is only set on /me/sessions; admin endpoints always return
// false because the admin viewing the page is not the session owner.
type SessionResponse struct {
	ID                string `json:"id"`
	UserID            string `json:"user_id"`
	DeviceName        string `json:"device_name"`
	UserAgent         string `json:"user_agent"`
	IPAddress         string `json:"ip_address"`
	CreatedAt         string `json:"created_at"`
	LastActivityAt    string `json:"last_activity_at"`
	ExpiresAt         string `json:"expires_at"`
	AbsoluteExpiresAt string `json:"absolute_expires_at"`
	IsCurrent         bool   `json:"is_current"`
}

// RevokeOthersResponse reports the number of sessions revoked.
type RevokeOthersResponse struct {
	RevokedCount int64 `json:"revoked_count"`
}
