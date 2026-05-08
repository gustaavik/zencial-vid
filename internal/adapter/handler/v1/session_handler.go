package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/postgres"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/pagination"
	sessionuc "github.com/zenfulcode/zencial/internal/usecase/session"
)

// SessionHandler exposes session-management endpoints for both /me/sessions
// (the calling user) and /admin/users/{id}/sessions (admin acting on a user).
type SessionHandler struct {
	sessions *sessionuc.Service
}

// NewSessionHandler creates a SessionHandler.
func NewSessionHandler(svc *sessionuc.Service) *SessionHandler {
	return &SessionHandler{sessions: svc}
}

// ListMine godoc
// @Summary      List my active sessions
// @Description  Returns the calling user's active sessions, with the current one flagged.
// @Tags         sessions
// @Produce      json
// @Success      200 {object} httputil.Response{data=[]dto.SessionResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/sessions [get]
func (h *SessionHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	currentSessionID, _ := middleware.GetSessionID(r.Context())

	fs, err := filter.FromRequest(r, postgres.SessionFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	out, appErr := h.sessions.ListMine(r.Context(), userID, &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.SessionsToResponse(out.Sessions, currentSessionID),
		pagination.NewMeta(fs.Pagination.Page, fs.Pagination.PerPage, out.Total),
	)
}

// RevokeMine godoc
// @Summary      Revoke a session of mine
// @Tags         sessions
// @Produce      json
// @Param        sessionID path string true "Session ID"
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/sessions/{sessionID} [delete]
func (h *SessionHandler) RevokeMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	sessionID, err := httputil.URLParamUUID(r, "sessionID")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}
	if appErr := h.sessions.RevokeMine(r.Context(), userID, sessionID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RevokeOthers godoc
// @Summary      Sign out from all other devices
// @Tags         sessions
// @Produce      json
// @Success      200 {object} httputil.Response{data=dto.RevokeOthersResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/sessions/revoke-others [post]
func (h *SessionHandler) RevokeOthers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	currentSessionID, ok := middleware.GetSessionID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "missing session context")
		return
	}
	out, appErr := h.sessions.RevokeOthers(r.Context(), userID, currentSessionID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.RevokeOthersToResponse(out))
}

// AdminListByUser godoc
// @Summary      Admin: list a user's sessions
// @Tags         admin
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} httputil.Response{data=[]dto.SessionResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users/{id}/sessions [get]
func (h *SessionHandler) AdminListByUser(w http.ResponseWriter, r *http.Request) {
	targetUserID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}
	fs, err := filter.FromRequest(r, postgres.SessionFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}
	out, appErr := h.sessions.AdminList(r.Context(), targetUserID, &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.SuccessWithMeta(w,
		mapper.SessionsToResponse(out.Sessions, uuid.Nil),
		pagination.NewMeta(fs.Pagination.Page, fs.Pagination.PerPage, out.Total),
	)
}

// AdminRevoke godoc
// @Summary      Admin: revoke a specific session
// @Tags         admin
// @Produce      json
// @Param        sessionID path string true "Session ID"
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/sessions/{sessionID} [delete]
func (h *SessionHandler) AdminRevoke(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	sessionID, err := httputil.URLParamUUID(r, "sessionID")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}
	if appErr := h.sessions.AdminRevoke(r.Context(), sessionID, actorID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AdminRevokeAll godoc
// @Summary      Admin: revoke all sessions for a user
// @Tags         admin
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} httputil.Response{data=dto.RevokeOthersResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users/{id}/sessions/revoke-all [post]
func (h *SessionHandler) AdminRevokeAll(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	targetUserID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}
	out, appErr := h.sessions.AdminRevokeAll(r.Context(), targetUserID, actorID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.RevokeOthersToResponse(out))
}
