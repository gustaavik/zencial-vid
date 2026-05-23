package v1

import (
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	authuc "github.com/zenfulcode/zencial/internal/usecase/auth"
)

// deviceNameHeader is the optional header clients can set to give a session
// a friendly label (e.g. "iPhone 15", "Chrome on macOS"). Falls back to
// the User-Agent header if absent.
const deviceNameHeader = "X-Device-Name"

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService *authuc.Service
	validator   *validator.Validator
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *authuc.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// extractSessionContext pulls device/network metadata from the inbound
// request to attach to the new session row. Relies on chi's RealIP
// middleware having canonicalised r.RemoteAddr from X-Forwarded-For.
func extractSessionContext(r *http.Request) authuc.SessionContext {
	return authuc.SessionContext{
		DeviceName: r.Header.Get(deviceNameHeader),
		UserAgent:  r.UserAgent(),
		IPAddress:  chiMiddleware.GetClientIP(r.Context()),
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account and return a session token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RegisterRequest true "Registration data"
// @Success      201 {object} httputil.Response{data=dto.AuthResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil),
			errors,
		)
		return
	}

	output, appErr := h.authService.Register(r.Context(), &authuc.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Session:  extractSessionContext(r),
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.AuthToResponse(output.User, output.Session, output.Token))
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password, returns a session token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LoginRequest true "Login credentials"
// @Success      200 {object} httputil.Response{data=dto.AuthResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil),
			errors,
		)
		return
	}

	output, appErr := h.authService.Login(r.Context(), &authuc.LoginInput{
		Email:    req.Email,
		Password: req.Password,
		Session:  extractSessionContext(r),
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.AuthToResponse(output.User, output.Session, output.Token))
}

// Logout godoc
// @Summary      Logout
// @Description  Revoke the session associated with the current bearer token
// @Tags         auth
// @Produce      json
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	sessionID, ok := middleware.GetSessionID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "missing session context")
		return
	}

	appErr := h.authService.Logout(r.Context(), authuc.LogoutInput{
		UserID:    userID,
		SessionID: sessionID,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
