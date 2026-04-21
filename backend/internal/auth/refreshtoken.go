package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrInvalidRefreshToken  = errors.New("invalid or expired refresh token")
)

// RefreshToken is a long-lived opaque credential used to obtain new access tokens.
type RefreshToken struct {
	Token     string
	AccountID AccountID
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// RefreshTokenRepository persists and retrieves RefreshToken records.
type RefreshTokenRepository interface {
	Save(ctx context.Context, token RefreshToken) error
	FindByToken(ctx context.Context, token string) (RefreshToken, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllForAccount(ctx context.Context, accountID AccountID) error
}
