package auth_test

import (
	"context"
	"testing"

	"github.com/hjoeftung/keklik/internal/auth"
)

func TestHandleTestLoginCreatesAccountAndToken(t *testing.T) {
	t.Parallel()

	accounts := &inMemoryAccountRepository{}
	refreshTokens := &inMemoryRefreshTokenRepository{}
	h := auth.NewHandleTestLoginHandler(accounts, refreshTokens, testSigningKey, testAccessTokenDuration, testRefreshTokenDuration)

	result, err := h.Handle(context.Background(), auth.HandleTestLoginCommand{Identifier: "qa-user"})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if result.Account.GoogleSubjectID != "test:qa-user" {
		t.Fatalf("expected test subject ID %q, got %q", "test:qa-user", result.Account.GoogleSubjectID)
	}

	if result.Account.Email != "qa-user@test.local" {
		t.Fatalf("expected test email %q, got %q", "qa-user@test.local", result.Account.Email)
	}

	if result.AccessToken == "" {
		t.Fatal("expected a non-empty access token")
	}

	validator := auth.NewJWTValidator(testSigningKey)
	identity, err := validator.Validate(context.Background(), result.AccessToken)
	if err != nil {
		t.Fatalf("issued token failed validation: %v", err)
	}
	if identity.AccountID != result.Account.ID {
		t.Fatalf("token account mismatch: got %q want %q", identity.AccountID, result.Account.ID)
	}

	if result.RefreshToken == "" {
		t.Fatal("expected a non-empty refresh token")
	}

	if len(accounts.saved) != 1 {
		t.Fatalf("expected 1 saved account, got %d", len(accounts.saved))
	}
}

func TestHandleTestLoginReusesExistingAccount(t *testing.T) {
	t.Parallel()

	existing := auth.Account{
		ID:              auth.AccountID("existing-id"),
		GoogleSubjectID: "test:qa-user",
		Email:           "qa-user@test.local",
	}
	accounts := &inMemoryAccountRepository{saved: []auth.Account{existing}}
	refreshTokens := &inMemoryRefreshTokenRepository{}
	h := auth.NewHandleTestLoginHandler(accounts, refreshTokens, testSigningKey, testAccessTokenDuration, testRefreshTokenDuration)

	result, err := h.Handle(context.Background(), auth.HandleTestLoginCommand{Identifier: "qa-user"})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if result.Account.ID != existing.ID {
		t.Fatalf("expected existing account ID %q, got %q", existing.ID, result.Account.ID)
	}

	if len(accounts.saved) != 1 {
		t.Fatalf("expected no new account to be saved, got %d total", len(accounts.saved))
	}
}

func TestHandleTestLoginRejectsEmptyIdentifier(t *testing.T) {
	t.Parallel()

	h := auth.NewHandleTestLoginHandler(&inMemoryAccountRepository{}, &inMemoryRefreshTokenRepository{}, testSigningKey, testAccessTokenDuration, testRefreshTokenDuration)

	_, err := h.Handle(context.Background(), auth.HandleTestLoginCommand{})
	if err == nil {
		t.Fatal("expected error for empty identifier")
	}

	if got := err.Error(); got != "identifier must not be empty" {
		t.Fatalf("unexpected error: %q", got)
	}
}
