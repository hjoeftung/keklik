package auth

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const testAuthSubjectPrefix = "test:"
const testAuthEmailDomain = "test.local"

// HandleTestLoginCommand describes the requested test-only identity.
type HandleTestLoginCommand struct {
	Identifier string
}

// HandleTestLoginHandler resolves or provisions a test-only Account and issues a JWT.
type HandleTestLoginHandler struct {
	accounts   AccountRepository
	signingKey string
	duration   time.Duration
}

// NewHandleTestLoginHandler returns a handler backed by the given account repository.
func NewHandleTestLoginHandler(accounts AccountRepository, signingKey string, duration time.Duration) *HandleTestLoginHandler {
	return &HandleTestLoginHandler{accounts: accounts, signingKey: signingKey, duration: duration}
}

// Handle resolves or provisions a test-only account using a deterministic subject ID
// derived from the caller-provided identifier, then issues a signed JWT.
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

	token, err := IssueJWT(account.ID, h.signingKey, h.duration)
	if err != nil {
		return HandleOAuthCallbackResult{}, err
	}

	return HandleOAuthCallbackResult{Account: account, Token: token}, nil
}

func testAuthEmail(identifier string) string {
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "@", "-")
	localPart := replacer.Replace(strings.ToLower(identifier))
	return localPart + "@" + testAuthEmailDomain
}
