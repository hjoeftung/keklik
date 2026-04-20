package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	AccountID string `json:"account_id"`
	jwt.RegisteredClaims
}

// IssueJWT creates a signed HS256 JWT encoding the given account ID and expiry.
func IssueJWT(accountID AccountID, signingKey string, duration time.Duration) (string, error) {
	expiresAt := now().Add(duration)
	claims := jwtClaims{
		AccountID: string(accountID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

// JWTValidator implements TokenValidator by verifying HMAC-SHA256 signed JWTs.
type JWTValidator struct {
	signingKey string
}

// NewJWTValidator returns a JWTValidator using the given signing key.
func NewJWTValidator(signingKey string) *JWTValidator {
	return &JWTValidator{signingKey: signingKey}
}

// Validate parses and verifies the token, returning the embedded Identity.
// Returns ErrInvalidToken for malformed, expired, or incorrectly signed tokens.
func (v *JWTValidator) Validate(_ context.Context, raw string) (Identity, error) {
	var claims jwtClaims
	_, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(v.signingKey), nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		return Identity{}, ErrInvalidToken
	}

	if claims.AccountID == "" {
		return Identity{}, ErrInvalidToken
	}

	return Identity{
		AccountID: AccountID(claims.AccountID),
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
