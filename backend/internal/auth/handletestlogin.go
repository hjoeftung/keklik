package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const testAuthSubjectPrefix = "test:"
const testAuthEmailDomain = "test.local"

// HandleTestLoginCommand describes the requested test-only identity.
type HandleTestLoginCommand struct {
	Identifier string
}

// HandleTestLoginHandler resolves or provisions a test-only Account and issues a token pair.
type HandleTestLoginHandler struct {
	accounts        AccountRepository
	refreshTokens   RefreshTokenRepository
	signingKey      string
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewHandleTestLoginHandler returns a handler backed by the given repositories.
func NewHandleTestLoginHandler(
	accounts AccountRepository,
	refreshTokens RefreshTokenRepository,
	signingKey string,
	accessDuration, refreshDuration time.Duration,
) *HandleTestLoginHandler {
	return &HandleTestLoginHandler{
		accounts:        accounts,
		refreshTokens:   refreshTokens,
		signingKey:      signingKey,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

// Handle resolves or provisions a test-only account using a deterministic subject ID
// derived from the caller-provided identifier, then issues an access JWT and refresh token.
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

func testAuthEmail(identifier string) string {
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "@", "-")
	localPart := replacer.Replace(strings.ToLower(identifier))
	return localPart + "@" + testAuthEmailDomain
}
