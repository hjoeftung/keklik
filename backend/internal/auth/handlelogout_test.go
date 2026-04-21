package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/auth"
)

func TestHandleLogout_RevokesAllTokensForAccount(t *testing.T) {
	t.Parallel()

	accountID := auth.AccountID("acc-logout")
	refreshTokens := &inMemoryRefreshTokenRepository{}
	refreshTokens.tokens = append(refreshTokens.tokens,
		auth.RefreshToken{Token: "tok-1", AccountID: accountID, ExpiresAt: time.Now().Add(time.Hour)},
		auth.RefreshToken{Token: "tok-2", AccountID: accountID, ExpiresAt: time.Now().Add(time.Hour)},
		auth.RefreshToken{Token: "tok-other", AccountID: "other-account", ExpiresAt: time.Now().Add(time.Hour)},
	)

	h := auth.NewHandleLogoutHandler(refreshTokens)
	if err := h.Handle(context.Background(), auth.HandleLogoutCommand{AccountID: accountID}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, tok := range []string{"tok-1", "tok-2"} {
		stored, _ := refreshTokens.FindByToken(context.Background(), tok)
		if stored.RevokedAt == nil {
			t.Errorf("expected token %q to be revoked after logout", tok)
		}
	}

	// Tokens belonging to other accounts must be unaffected.
	other, _ := refreshTokens.FindByToken(context.Background(), "tok-other")
	if other.RevokedAt != nil {
		t.Error("expected other-account token to remain active after logout")
	}
}
