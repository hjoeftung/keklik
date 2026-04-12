package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const sessionDuration = 30 * 24 * time.Hour

// HandleOAuthCallbackCommand holds the verified Google identity from the OAuth callback.
type HandleOAuthCallbackCommand struct {
	GoogleSubjectID string
	Email           string
}

// HandleOAuthCallbackResult holds the resolved account and the new session token.
type HandleOAuthCallbackResult struct {
	Account Account
	Session Session
}

// HandleOAuthCallbackHandler resolves or provisions an internal Account for a Google identity
// and issues a new session token.
type HandleOAuthCallbackHandler struct {
	accounts AccountRepository
	sessions SessionRepository
}

// NewHandleOAuthCallbackHandler returns a handler backed by the given repositories.
func NewHandleOAuthCallbackHandler(accounts AccountRepository, sessions SessionRepository) *HandleOAuthCallbackHandler {
	return &HandleOAuthCallbackHandler{accounts: accounts, sessions: sessions}
}

// Handle looks up the account by Google subject ID, creating one if it does not exist,
// then issues a new session. The link between the resulting Account and a family-domain
// FamilyMember is resolved separately by callers using the Account.GoogleSubjectID.
func (h *HandleOAuthCallbackHandler) Handle(ctx context.Context, cmd HandleOAuthCallbackCommand) (HandleOAuthCallbackResult, error) {
	if cmd.GoogleSubjectID == "" {
		return HandleOAuthCallbackResult{}, fmt.Errorf("google subject ID must not be empty")
	}

	account, err := h.accounts.FindByGoogleSubjectID(ctx, cmd.GoogleSubjectID)
	if err != nil {
		if !errors.Is(err, ErrAccountNotFound) {
			return HandleOAuthCallbackResult{}, fmt.Errorf("look up account: %w", err)
		}

		account = Account{
			ID:              AccountID(uuid.New().String()),
			GoogleSubjectID: cmd.GoogleSubjectID,
			Email:           cmd.Email,
		}
		if err := h.accounts.Save(ctx, account); err != nil {
			return HandleOAuthCallbackResult{}, fmt.Errorf("save account: %w", err)
		}
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return HandleOAuthCallbackResult{}, fmt.Errorf("generate session token: %w", err)
	}

	session := Session{
		Token:     SessionToken(base64.RawURLEncoding.EncodeToString(tokenBytes)),
		AccountID: account.ID,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	if err := h.sessions.Save(ctx, session); err != nil {
		return HandleOAuthCallbackResult{}, fmt.Errorf("save session: %w", err)
	}

	return HandleOAuthCallbackResult{Account: account, Session: session}, nil
}
