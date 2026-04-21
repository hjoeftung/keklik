package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/auth"
)

func newRefreshHandler(refreshTokens *inMemoryRefreshTokenRepository) *auth.HandleRefreshTokenHandler {
	return auth.NewHandleRefreshTokenHandler(refreshTokens, testSigningKey, testAccessTokenDuration, testRefreshTokenDuration)
}

func TestHandleRefreshTokenRotatesTokens(t *testing.T) {
	t.Parallel()

	refreshTokens := &inMemoryRefreshTokenRepository{}
	accountID := auth.AccountID("acc-1")
	oldToken := auth.RefreshToken{
		Token:     "old-token-uuid",
		AccountID: accountID,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	refreshTokens.tokens = append(refreshTokens.tokens, oldToken)

	h := newRefreshHandler(refreshTokens)
	result, err := h.Handle(context.Background(), auth.HandleRefreshTokenCommand{Token: "old-token-uuid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if result.RefreshToken == "old-token-uuid" {
		t.Error("expected a new refresh token, not the old one")
	}

	// Old token must be revoked.
	stored, _ := refreshTokens.FindByToken(context.Background(), "old-token-uuid")
	if stored.RevokedAt == nil {
		t.Error("expected old refresh token to be revoked after rotation")
	}

	// New token must be persisted and active.
	newStored, err := refreshTokens.FindByToken(context.Background(), result.RefreshToken)
	if err != nil {
		t.Fatalf("new refresh token not found in repo: %v", err)
	}
	if newStored.RevokedAt != nil {
		t.Error("expected new refresh token to not be revoked")
	}
	if newStored.AccountID != accountID {
		t.Errorf("expected new refresh token account ID %q, got %q", accountID, newStored.AccountID)
	}
}

func TestHandleRefreshToken_OldTokenRejectedAfterRotation(t *testing.T) {
	t.Parallel()

	refreshTokens := &inMemoryRefreshTokenRepository{}
	refreshTokens.tokens = append(refreshTokens.tokens, auth.RefreshToken{
		Token:     "rotate-me",
		AccountID: "acc-2",
		ExpiresAt: time.Now().Add(time.Hour),
	})

	h := newRefreshHandler(refreshTokens)

	// First use — valid.
	_, err := h.Handle(context.Background(), auth.HandleRefreshTokenCommand{Token: "rotate-me"})
	if err != nil {
		t.Fatalf("first refresh failed: %v", err)
	}

	// Second use of the now-rotated token — must fail.
	_, err = h.Handle(context.Background(), auth.HandleRefreshTokenCommand{Token: "rotate-me"})
	if err == nil {
		t.Fatal("expected error when reusing a rotated refresh token")
	}
}

func TestHandleRefreshToken_ReplayRevokesEntireFamily(t *testing.T) {
	t.Parallel()

	accountID := auth.AccountID("acc-replay")
	refreshTokens := &inMemoryRefreshTokenRepository{}

	// Seed an already-rotated (revoked) token and a still-active sibling.
	revokedAt := time.Now().Add(-time.Minute)
	refreshTokens.tokens = append(refreshTokens.tokens,
		auth.RefreshToken{Token: "old-revoked", AccountID: accountID, ExpiresAt: time.Now().Add(time.Hour), RevokedAt: &revokedAt},
		auth.RefreshToken{Token: "active-sibling", AccountID: accountID, ExpiresAt: time.Now().Add(time.Hour)},
	)

	h := newRefreshHandler(refreshTokens)
	_, err := h.Handle(context.Background(), auth.HandleRefreshTokenCommand{Token: "old-revoked"})
	if err == nil {
		t.Fatal("expected error on replay of revoked token")
	}

	// Active sibling must also be revoked.
	sibling, _ := refreshTokens.FindByToken(context.Background(), "active-sibling")
	if sibling.RevokedAt == nil {
		t.Error("expected active sibling to be revoked after replay detection")
	}
}

func TestHandleRefreshToken_ExpiredTokenRejected(t *testing.T) {
	t.Parallel()

	refreshTokens := &inMemoryRefreshTokenRepository{}
	refreshTokens.tokens = append(refreshTokens.tokens, auth.RefreshToken{
		Token:     "expired-token",
		AccountID: "acc-3",
		ExpiresAt: time.Now().Add(-time.Hour),
	})

	h := newRefreshHandler(refreshTokens)
	_, err := h.Handle(context.Background(), auth.HandleRefreshTokenCommand{Token: "expired-token"})
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
}

func TestHandleRefreshToken_UnknownTokenRejected(t *testing.T) {
	t.Parallel()

	h := newRefreshHandler(&inMemoryRefreshTokenRepository{})
	_, err := h.Handle(context.Background(), auth.HandleRefreshTokenCommand{Token: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown refresh token")
	}
}

func TestHandleRefreshToken_EmptyTokenRejected(t *testing.T) {
	t.Parallel()

	h := newRefreshHandler(&inMemoryRefreshTokenRepository{})
	_, err := h.Handle(context.Background(), auth.HandleRefreshTokenCommand{})
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}
