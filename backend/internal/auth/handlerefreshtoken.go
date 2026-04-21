package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// HandleRefreshTokenCommand carries the opaque refresh token string from the client.
type HandleRefreshTokenCommand struct {
	Token string
}

// HandleRefreshTokenResult carries the new access and refresh tokens after a successful rotation.
type HandleRefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
}

// HandleRefreshTokenHandler validates a refresh token, issues a new access JWT, and rotates
// the refresh token (revoke-old, persist-new). If a revoked token is presented it signals
// replay, and all tokens for the account are revoked.
type HandleRefreshTokenHandler struct {
	refreshTokens   RefreshTokenRepository
	signingKey      string
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewHandleRefreshTokenHandler returns a handler backed by the given repository.
func NewHandleRefreshTokenHandler(
	refreshTokens RefreshTokenRepository,
	signingKey string,
	accessDuration, refreshDuration time.Duration,
) *HandleRefreshTokenHandler {
	return &HandleRefreshTokenHandler{
		refreshTokens:   refreshTokens,
		signingKey:      signingKey,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

// Handle validates the refresh token, rotates it, and returns a new access + refresh token pair.
func (h *HandleRefreshTokenHandler) Handle(ctx context.Context, cmd HandleRefreshTokenCommand) (HandleRefreshTokenResult, error) {
	if cmd.Token == "" {
		return HandleRefreshTokenResult{}, ErrInvalidRefreshToken
	}

	existing, err := h.refreshTokens.FindByToken(ctx, cmd.Token)
	if err != nil {
		return HandleRefreshTokenResult{}, ErrInvalidRefreshToken
	}

	// Replay detection: a revoked token being re-used means the token family was leaked.
	// Revoke all tokens for the account and reject the request.
	if existing.RevokedAt != nil {
		_ = h.refreshTokens.RevokeAllForAccount(ctx, existing.AccountID)
		return HandleRefreshTokenResult{}, ErrInvalidRefreshToken
	}

	if !existing.ExpiresAt.After(now()) {
		return HandleRefreshTokenResult{}, ErrInvalidRefreshToken
	}

	accessToken, err := IssueJWT(existing.AccountID, h.signingKey, h.accessDuration)
	if err != nil {
		return HandleRefreshTokenResult{}, fmt.Errorf("issue access token: %w", err)
	}

	if err := h.refreshTokens.Revoke(ctx, cmd.Token); err != nil {
		return HandleRefreshTokenResult{}, fmt.Errorf("revoke old refresh token: %w", err)
	}

	newToken := RefreshToken{
		Token:     uuid.New().String(),
		AccountID: existing.AccountID,
		ExpiresAt: now().Add(h.refreshDuration),
	}
	if err := h.refreshTokens.Save(ctx, newToken); err != nil {
		return HandleRefreshTokenResult{}, fmt.Errorf("save new refresh token: %w", err)
	}

	return HandleRefreshTokenResult{
		AccessToken:  accessToken,
		RefreshToken: newToken.Token,
	}, nil
}
