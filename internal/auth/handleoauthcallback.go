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

var now = time.Now

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

	account, err := findOrCreateAccount(ctx, h.accounts, cmd.GoogleSubjectID, cmd.Email)
	if err != nil {
		return HandleOAuthCallbackResult{}, err
	}

	session, err := issueSession(ctx, h.sessions, account.ID, now)
	if err != nil {
		return HandleOAuthCallbackResult{}, err
	}

	return HandleOAuthCallbackResult{Account: account, Session: session}, nil
}

func findOrCreateAccount(ctx context.Context, accounts AccountRepository, subjectID, email string) (Account, error) {
	account, err := accounts.FindByGoogleSubjectID(ctx, subjectID)
	if err == nil {
		return account, nil
	}
	if !errors.Is(err, ErrAccountNotFound) {
		return Account{}, fmt.Errorf("look up account: %w", err)
	}

	account = Account{
		ID:              AccountID(uuid.New().String()),
		GoogleSubjectID: subjectID,
		Email:           email,
	}
	if err := accounts.Save(ctx, account); err != nil {
		return Account{}, fmt.Errorf("save account: %w", err)
	}

	return account, nil
}

func issueSession(ctx context.Context, sessions SessionRepository, accountID AccountID, clock func() time.Time) (Session, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return Session{}, fmt.Errorf("generate session token: %w", err)
	}

	session := Session{
		Token:     SessionToken(base64.RawURLEncoding.EncodeToString(tokenBytes)),
		AccountID: accountID,
		ExpiresAt: clock().Add(sessionDuration),
	}
	if err := sessions.Save(ctx, session); err != nil {
		return Session{}, fmt.Errorf("save session: %w", err)
	}

	return session, nil
}
