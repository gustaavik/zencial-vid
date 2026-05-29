package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/clock"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
)

const (
	userIDKey    contextKey = "user_id"
	userRolesKey contextKey = "user_roles"
	sessionIDKey contextKey = "session_id"
)

// SessionAuthenticator validates session tokens against the session repository
// on every request, slides idle expiry forward (debounced), and reloads the
// user so role/status changes take effect immediately without a per-token
// revocation list.
type SessionAuthenticator struct {
	sessionRepo repository.SessionRepository
	userRepo    repository.UserRepository
	tokens      auth.SessionTokenService
	clock       clock.Clock
	cfg         config.SessionConfig
	log         *slog.Logger
}

// NewSessionAuthenticator constructs the middleware factory.
func NewSessionAuthenticator(
	sessionRepo repository.SessionRepository,
	userRepo repository.UserRepository,
	tokens auth.SessionTokenService,
	clk clock.Clock,
	cfg config.SessionConfig,
	log *slog.Logger,
) *SessionAuthenticator {
	return &SessionAuthenticator{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		tokens:      tokens,
		clock:       clk,
		cfg:         cfg,
		log:         log,
	}
}

// Authenticate validates the bearer session token and stores the resolved
// user_id, user_role, and session_id in the request context. Rejects with 401
// for any failure mode (missing/malformed header, unknown token, revoked,
// expired, suspended user).
func (a *SessionAuthenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			httputil.Unauthorized(w, apperror.CodeUnauthorized, "missing authorization header")
			return
		}

		ctx, code, ok := a.resolve(r.Context(), token)
		if !ok {
			httputil.Unauthorized(w, code, unauthorizedMessage(code))
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthenticate parses the bearer session token if present but does
// not reject unauthenticated requests. Anonymous requests pass through with
// no auth context populated.
func (a *SessionAuthenticator) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		ctx, _, ok := a.resolve(r.Context(), token)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// resolve performs the full authentication pipeline and returns the augmented
// context plus an apperror code on failure. Caller decides whether to reject
// or pass through (Authenticate vs OptionalAuthenticate).
func (a *SessionAuthenticator) resolve(ctx context.Context, token string) (context.Context, string, bool) {
	hash := a.tokens.Hash(token)
	sess, err := a.sessionRepo.GetByTokenHash(ctx, hash)
	if err != nil {
		a.log.Error("session lookup failed", "error", err)
		return ctx, apperror.CodeInternalError, false
	}
	if sess == nil {
		return ctx, apperror.CodeInvalidToken, false
	}

	now := a.clock.Now()
	if sess.IsRevoked() {
		return ctx, apperror.CodeSessionRevoked, false
	}
	if sess.IsExpired(now) {
		return ctx, apperror.CodeSessionExpired, false
	}

	user, err := a.userRepo.GetByID(ctx, sess.UserID)
	if err != nil {
		a.log.Error("user lookup failed", "error", err, "user_id", sess.UserID)
		return ctx, apperror.CodeInternalError, false
	}
	if user == nil || !user.IsActive() {
		return ctx, apperror.CodeUnauthorized, false
	}

	if now.Sub(sess.LastActivityAt) >= a.cfg.SlideDebounce {
		slid := sess.Slide(now, a.cfg.IdleTimeout)
		if err := a.sessionRepo.UpdateActivity(ctx, slid.ID, slid.LastActivityAt, slid.IdleExpiresAt); err != nil {
			a.log.Warn("slide activity failed", "error", err, "session_id", sess.ID)
		}
	}

	ctx = context.WithValue(ctx, userIDKey, sess.UserID)
	ctx = context.WithValue(ctx, userRolesKey, user.Roles)
	ctx = context.WithValue(ctx, sessionIDKey, sess.ID)
	ctx = actor.WithActor(ctx, sess.UserID)
	return ctx, "", true
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func unauthorizedMessage(code string) string {
	switch code {
	case apperror.CodeSessionRevoked:
		return "session has been revoked"
	case apperror.CodeSessionExpired:
		return "session has expired"
	case apperror.CodeUnauthorized:
		return "authentication required"
	case apperror.CodeInternalError:
		return "authentication failed"
	default:
		return "invalid or expired token"
	}
}

// GetUserID retrieves the authenticated user's ID from context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

// GetUserRoles retrieves the authenticated user's roles from context.
func GetUserRoles(ctx context.Context) ([]entity.UserRole, bool) {
	roles, ok := ctx.Value(userRolesKey).([]entity.UserRole)
	return roles, ok
}

// CallerHasRole reports whether the authenticated user holds the given role.
func CallerHasRole(ctx context.Context, role entity.UserRole) bool {
	roles, ok := ctx.Value(userRolesKey).([]entity.UserRole)
	if !ok {
		return false
	}
	return entity.HasRole(roles, role)
}

// GetSessionID retrieves the authenticated session's ID from context.
func GetSessionID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(sessionIDKey).(uuid.UUID)
	return id, ok
}
