package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var now = time.Now

// HandleOAuthCallbackCommand holds the verified Google identity from the OAuth callback.
type HandleOAuthCallbackCommand struct {
	GoogleSubjectID string
	Email           string
}

// HandleOAuthCallbackResult holds the resolved account and the new token pair.
type HandleOAuthCallbackResult struct {
	Account      Account
	AccessToken  string
	RefreshToken string
}

// HandleOAuthCallbackHandler resolves or provisions an internal Account for a Google identity
// and issues a short-lived access JWT plus a rotatable refresh token.
type HandleOAuthCallbackHandler struct {
	accounts        AccountRepository
	refreshTokens   RefreshTokenRepository
	signingKey      string
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewHandleOAuthCallbackHandler returns a handler backed by the given repositories.
func NewHandleOAuthCallbackHandler(
	accounts AccountRepository,
	refreshTokens RefreshTokenRepository,
	signingKey string,
	accessDuration, refreshDuration time.Duration,
) *HandleOAuthCallbackHandler {
	return &HandleOAuthCallbackHandler{
		accounts:        accounts,
		refreshTokens:   refreshTokens,
		signingKey:      signingKey,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

// Handle looks up the account by Google subject ID, creating one if it does not exist,
// then issues an access JWT and a new refresh token.
func (h *HandleOAuthCallbackHandler) Handle(ctx context.Context, cmd HandleOAuthCallbackCommand) (HandleOAuthCallbackResult, error) {
	if cmd.GoogleSubjectID == "" {
		return HandleOAuthCallbackResult{}, fmt.Errorf("google subject ID must not be empty")
	}

	account, err := findOrCreateAccount(ctx, h.accounts, cmd.GoogleSubjectID, cmd.Email)
	if err != nil {
		return HandleOAuthCallbackResult{}, err
	}

	accessToken, err := IssueJWT(account.ID, h.signingKey, h.accessDuration)
	if err != nil {
		return HandleOAuthCallbackResult{}, err
	}

	refreshToken := RefreshToken{
		Token:     uuid.New().String(),
		AccountID: account.ID,
		ExpiresAt: now().Add(h.refreshDuration),
	}
	if err := h.refreshTokens.Save(ctx, refreshToken); err != nil {
		return HandleOAuthCallbackResult{}, fmt.Errorf("save refresh token: %w", err)
	}

	return HandleOAuthCallbackResult{
		Account:      account,
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
	}, nil
}

func findOrCreateAccount(ctx context.Context, accounts AccountRepository, subjectID, email string) (Account, error) {
	account, err := accounts.Upsert(ctx, Account{
		ID:              AccountID(uuid.New().String()),
		GoogleSubjectID: subjectID,
		Email:           email,
	})
	if err != nil {
		return Account{}, fmt.Errorf("upsert account: %w", err)
	}
	return account, nil
}
