package auth

import (
	"context"
	"errors"
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

// HandleOAuthCallbackResult holds the resolved account and the new signed JWT.
type HandleOAuthCallbackResult struct {
	Account Account
	Token   string
}

// HandleOAuthCallbackHandler resolves or provisions an internal Account for a Google identity
// and issues a signed JWT.
type HandleOAuthCallbackHandler struct {
	accounts   AccountRepository
	signingKey string
	duration   time.Duration
}

// NewHandleOAuthCallbackHandler returns a handler backed by the given account repository.
func NewHandleOAuthCallbackHandler(accounts AccountRepository, signingKey string, duration time.Duration) *HandleOAuthCallbackHandler {
	return &HandleOAuthCallbackHandler{accounts: accounts, signingKey: signingKey, duration: duration}
}

// Handle looks up the account by Google subject ID, creating one if it does not exist,
// then issues a signed JWT.
func (h *HandleOAuthCallbackHandler) Handle(ctx context.Context, cmd HandleOAuthCallbackCommand) (HandleOAuthCallbackResult, error) {
	if cmd.GoogleSubjectID == "" {
		return HandleOAuthCallbackResult{}, fmt.Errorf("google subject ID must not be empty")
	}

	account, err := findOrCreateAccount(ctx, h.accounts, cmd.GoogleSubjectID, cmd.Email)
	if err != nil {
		return HandleOAuthCallbackResult{}, err
	}

	token, err := IssueJWT(account.ID, h.signingKey, h.duration)
	if err != nil {
		return HandleOAuthCallbackResult{}, err
	}

	return HandleOAuthCallbackResult{Account: account, Token: token}, nil
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
