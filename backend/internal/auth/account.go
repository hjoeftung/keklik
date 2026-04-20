package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrSessionNotFound = errors.New("session not found")
)

// AccountID is the unique identifier for an authenticated identity.
type AccountID string

// SessionToken is an opaque bearer token that identifies an active session.
type SessionToken string

// Account represents an authenticated identity, separate from the family-domain FamilyMember.
// The link between an Account and a FamilyMember is the shared GoogleSubjectID.
type Account struct {
	ID              AccountID
	GoogleSubjectID string
	Email           string
	CreatedAt       time.Time
}

// Session holds a bearer token tied to an account, with an expiry.
type Session struct {
	Token     SessionToken
	AccountID AccountID
	ExpiresAt time.Time
}

// IsExpired reports whether the session has passed its expiry time.
func (s Session) IsExpired(now time.Time) bool {
	return !s.ExpiresAt.After(now)
}

// AccountRepository persists and retrieves Account records.
type AccountRepository interface {
	Save(ctx context.Context, account Account) error
	FindByID(ctx context.Context, id AccountID) (Account, error)
	FindByGoogleSubjectID(ctx context.Context, googleSubjectID string) (Account, error)
}

// SessionRepository persists and retrieves Session records.
type SessionRepository interface {
	Save(ctx context.Context, session Session) error
	FindByToken(ctx context.Context, token SessionToken) (Session, error)
}

// Identity is the verified result of a successful token validation.
//
// A JWT implementation must encode these two claims:
//   - "account_id": the internal AccountID (UUID string)
//   - "exp": Unix timestamp matching ExpiresAt
//
// Signing-key rotation should use a "kid" (key-ID) header paired with a
// key store so old tokens remain verifiable during rollover.
type Identity struct {
	AccountID AccountID
	ExpiresAt time.Time
}

// TokenValidator validates a raw bearer token and returns the associated Identity.
// The returned error is non-nil when the token is missing, invalid, or expired.
type TokenValidator interface {
	Validate(ctx context.Context, token string) (Identity, error)
}
