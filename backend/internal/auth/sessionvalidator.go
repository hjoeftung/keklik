package auth

import (
	"context"
	"time"
)

// DBSessionValidator implements TokenValidator using the session database.
type DBSessionValidator struct {
	sessions SessionRepository
}

// NewDBSessionValidator returns a validator backed by the given SessionRepository.
func NewDBSessionValidator(sessions SessionRepository) *DBSessionValidator {
	return &DBSessionValidator{sessions: sessions}
}

// Validate looks up the token in the session store, checks expiry, and returns
// the associated Identity. Returns ErrSessionNotFound if the token is unknown or expired.
func (v *DBSessionValidator) Validate(ctx context.Context, token string) (Identity, error) {
	session, err := v.sessions.FindByToken(ctx, SessionToken(token))
	if err != nil {
		return Identity{}, err
	}
	if session.IsExpired(time.Now()) {
		return Identity{}, ErrSessionNotFound
	}
	return Identity{
		AccountID: session.AccountID,
		ExpiresAt: session.ExpiresAt,
	}, nil
}
