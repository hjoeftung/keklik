package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrInvalidToken    = errors.New("invalid or expired token")
)

// AccountID is the unique identifier for an authenticated identity.
type AccountID string

// Account represents an authenticated identity, separate from the family-domain FamilyMember.
// The link between an Account and a FamilyMember is the shared GoogleSubjectID.
type Account struct {
	ID              AccountID
	GoogleSubjectID string
	Email           string
	CreatedAt       time.Time
}

// AccountRepository persists and retrieves Account records.
type AccountRepository interface {
	Save(ctx context.Context, account Account) error
	FindByID(ctx context.Context, id AccountID) (Account, error)
	FindByGoogleSubjectID(ctx context.Context, googleSubjectID string) (Account, error)
	// Upsert inserts the account or, on a google_subject_id conflict, returns the existing row.
	// It is safe to call concurrently for the same subject ID.
	Upsert(ctx context.Context, account Account) (Account, error)
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
