package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	authuc "github.com/zenfulcode/zencial/internal/usecase/auth"
)

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

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account and return authentication tokens
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

	output, appErr := h.authService.Register(r.Context(), authuc.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.AuthToResponse(output.User, output.TokenPair))
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password, returns tokens
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

	output, appErr := h.authService.Login(r.Context(), authuc.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.AuthToResponse(output.User, output.TokenPair))
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchange a valid refresh token for a new token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} httputil.Response{data=dto.TokenResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}

	output, appErr := h.authService.RefreshToken(r.Context(), authuc.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.TokenPairToResponse(output.TokenPair))
}

// Logout godoc
// @Summary      Logout
// @Description  Invalidate the refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LogoutRequest true "Refresh token to invalidate"
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}

	appErr := h.authService.Logout(r.Context(), authuc.LogoutInput{
		RefreshToken: req.RefreshToken,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
