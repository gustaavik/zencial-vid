package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// RevokeOthersOutput reports the number of sessions revoked.
type RevokeOthersOutput struct {
	RevokedCount int64
}

// RevokeMine revokes a single session belonging to the calling user. Returns
// SESSION_NOT_FOUND if the session belongs to a different user — the response
// deliberately doesn't disambiguate from "doesn't exist" to avoid leaking
// existence of other users' session IDs.
func (s *Service) RevokeMine(
	ctx context.Context,
	userID, sessionID uuid.UUID,
) *apperror.AppError {
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		s.log.Error("getting session", "error", err, "session_id", sessionID)
		return apperror.Internal(apperror.CodeInternalError, "failed to revoke session", err)
	}
	if sess == nil || sess.UserID != userID {
		return apperror.NotFound(apperror.CodeSessionNotFound, "session not found", domain.ErrSessionNotFound)
	}

	now := s.clock.Now()
	if err := s.sessionRepo.Revoke(ctx, sess.ID, now); err != nil {
		s.log.Error("revoking session", "error", err, "session_id", sess.ID)
		return apperror.Internal(apperror.CodeInternalError, "failed to revoke session", err)
	}

	actor := userID
	if dispatchErr := s.dispatcher.Dispatch(event.SessionRevoked{
		UserID:    sess.UserID,
		SessionID: sess.ID,
		ActorID:   &actor,
		Timestamp: now,
	}); dispatchErr != nil {
		s.log.Error("dispatching session revoked event", "error", dispatchErr)
	}
	return nil
}

// RevokeOthers revokes all of the user's sessions except the given current
// session. Returns the number of sessions revoked.
func (s *Service) RevokeOthers(
	ctx context.Context,
	userID, currentSessionID uuid.UUID,
) (*RevokeOthersOutput, *apperror.AppError) {
	now := s.clock.Now()
	count, err := s.sessionRepo.RevokeOthersForUser(ctx, userID, currentSessionID, now)
	if err != nil {
		s.log.Error("revoking other sessions", "error", err, "user_id", userID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to revoke sessions", err)
	}
	return &RevokeOthersOutput{RevokedCount: count}, nil
}

// AdminRevoke revokes any session by ID. Authorization is enforced at the
// middleware layer.
func (s *Service) AdminRevoke(
	ctx context.Context,
	sessionID uuid.UUID,
	actorID uuid.UUID,
) *apperror.AppError {
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		s.log.Error("getting session", "error", err, "session_id", sessionID)
		return apperror.Internal(apperror.CodeInternalError, "failed to revoke session", err)
	}
	if sess == nil {
		return apperror.NotFound(apperror.CodeSessionNotFound, "session not found", domain.ErrSessionNotFound)
	}

	now := s.clock.Now()
	if err := s.sessionRepo.Revoke(ctx, sess.ID, now); err != nil {
		s.log.Error("revoking session", "error", err, "session_id", sess.ID)
		return apperror.Internal(apperror.CodeInternalError, "failed to revoke session", err)
	}

	if dispatchErr := s.dispatcher.Dispatch(event.SessionRevoked{
		UserID:    sess.UserID,
		SessionID: sess.ID,
		ActorID:   &actorID,
		Timestamp: now,
	}); dispatchErr != nil {
		s.log.Error("dispatching session revoked event", "error", dispatchErr)
	}
	return nil
}

// AdminRevokeAll revokes every active session for a target user.
func (s *Service) AdminRevokeAll(
	ctx context.Context,
	targetUserID, actorID uuid.UUID,
) (*RevokeOthersOutput, *apperror.AppError) {
	now := s.clock.Now()
	count, err := s.sessionRepo.RevokeAllForUser(ctx, targetUserID, now)
	if err != nil {
		s.log.Error("revoking all sessions", "error", err, "user_id", targetUserID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to revoke sessions", err)
	}

	if count > 0 {
		if dispatchErr := s.dispatcher.Dispatch(event.SessionRevoked{
			UserID:    targetUserID,
			ActorID:   &actorID,
			Timestamp: now,
		}); dispatchErr != nil {
			s.log.Error("dispatching session revoked event", "error", dispatchErr)
		}
	}
	return &RevokeOthersOutput{RevokedCount: count}, nil
}
