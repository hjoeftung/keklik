package auth

import (
	"context"
	"fmt"
	"strings"
)

const testAuthSubjectPrefix = "test:"
const testAuthEmailDomain = "test.local"

// HandleTestLoginCommand describes the requested test-only identity.
type HandleTestLoginCommand struct {
	Identifier string
}

// HandleTestLoginHandler resolves or provisions a test-only Account and issues a session.
type HandleTestLoginHandler struct {
	accounts AccountRepository
	sessions SessionRepository
}

// NewHandleTestLoginHandler returns a handler backed by the given repositories.
func NewHandleTestLoginHandler(accounts AccountRepository, sessions SessionRepository) *HandleTestLoginHandler {
	return &HandleTestLoginHandler{accounts: accounts, sessions: sessions}
}

// Handle resolves or provisions a test-only account using a deterministic subject ID
// derived from the caller-provided identifier, then issues a normal application session.
func (h *HandleTestLoginHandler) Handle(ctx context.Context, cmd HandleTestLoginCommand) (HandleOAuthCallbackResult, error) {
	identifier := strings.TrimSpace(cmd.Identifier)
	if identifier == "" {
		return HandleOAuthCallbackResult{}, fmt.Errorf("identifier must not be empty")
	}

	subjectID := testAuthSubjectPrefix + identifier
	account, err := findOrCreateAccount(ctx, h.accounts, subjectID, testAuthEmail(identifier))
	if err != nil {
		return HandleOAuthCallbackResult{}, err
	}

	session, err := issueSession(ctx, h.sessions, account.ID, now)
	if err != nil {
		return HandleOAuthCallbackResult{}, err
	}

	return HandleOAuthCallbackResult{Account: account, Session: session}, nil
}

func testAuthEmail(identifier string) string {
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "@", "-")
	localPart := replacer.Replace(strings.ToLower(identifier))
	return localPart + "@" + testAuthEmailDomain
}
