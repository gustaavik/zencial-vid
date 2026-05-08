package session

import "context"

// PurgeExpired deletes session rows past their absolute expiry or revoked
// before now. Intended to be invoked by a periodic background job; errors are
// returned for caller logging since they are not user-facing.
func (s *Service) PurgeExpired(ctx context.Context) (int64, error) {
	return s.sessionRepo.DeleteExpired(ctx, s.clock.Now())
}
