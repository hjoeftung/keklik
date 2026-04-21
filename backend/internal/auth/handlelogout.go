package auth

import (
	"context"
	"fmt"
)

// HandleLogoutCommand identifies the account whose refresh tokens should be revoked.
type HandleLogoutCommand struct {
	AccountID AccountID
}

// HandleLogoutHandler revokes all refresh tokens for an account.
type HandleLogoutHandler struct {
	refreshTokens RefreshTokenRepository
}

// NewHandleLogoutHandler returns a handler backed by the given repository.
func NewHandleLogoutHandler(refreshTokens RefreshTokenRepository) *HandleLogoutHandler {
	return &HandleLogoutHandler{refreshTokens: refreshTokens}
}

// Handle revokes all active refresh tokens for the account.
func (h *HandleLogoutHandler) Handle(ctx context.Context, cmd HandleLogoutCommand) error {
	if err := h.refreshTokens.RevokeAllForAccount(ctx, cmd.AccountID); err != nil {
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}
	return nil
}
