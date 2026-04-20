package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/auth"
)

// --- test doubles ---

type inMemoryAccountRepository struct {
	saved []auth.Account
	err   error
}

func (r *inMemoryAccountRepository) Save(_ context.Context, a auth.Account) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, a)
	return nil
}

func (r *inMemoryAccountRepository) FindByID(_ context.Context, id auth.AccountID) (auth.Account, error) {
	for _, a := range r.saved {
		if a.ID == id {
			return a, nil
		}
	}
	return auth.Account{}, auth.ErrAccountNotFound
}

func (r *inMemoryAccountRepository) FindByGoogleSubjectID(_ context.Context, sub string) (auth.Account, error) {
	if r.err != nil {
		return auth.Account{}, r.err
	}
	for _, a := range r.saved {
		if a.GoogleSubjectID == sub {
			return a, nil
		}
	}
	return auth.Account{}, auth.ErrAccountNotFound
}

// --- helpers ---

const testSigningKey = "test-signing-key"
const testTokenDuration = 30 * 24 * time.Hour

func validCallbackCommand() auth.HandleOAuthCallbackCommand {
	return auth.HandleOAuthCallbackCommand{
		GoogleSubjectID: "google-sub-12345",
		Email:           "user@example.com",
	}
}

// --- tests ---

func TestHandleOAuthCallbackNewAccount(t *testing.T) {
	t.Parallel()

	accounts := &inMemoryAccountRepository{}
	h := auth.NewHandleOAuthCallbackHandler(accounts, testSigningKey, testTokenDuration)

	result, err := h.Handle(context.Background(), validCallbackCommand())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Account.GoogleSubjectID != "google-sub-12345" {
		t.Errorf("expected google subject ID %q, got %q", "google-sub-12345", result.Account.GoogleSubjectID)
	}
	if result.Account.Email != "user@example.com" {
		t.Errorf("expected email %q, got %q", "user@example.com", result.Account.Email)
	}
	if result.Account.ID == "" {
		t.Error("expected a non-empty account ID")
	}
	if len(accounts.saved) != 1 {
		t.Errorf("expected 1 saved account, got %d", len(accounts.saved))
	}
}

func TestHandleOAuthCallbackExistingAccount(t *testing.T) {
	t.Parallel()

	existing := auth.Account{
		ID:              auth.AccountID("existing-id"),
		GoogleSubjectID: "google-sub-12345",
		Email:           "old@example.com",
	}
	accounts := &inMemoryAccountRepository{saved: []auth.Account{existing}}
	h := auth.NewHandleOAuthCallbackHandler(accounts, testSigningKey, testTokenDuration)

	result, err := h.Handle(context.Background(), validCallbackCommand())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Account.ID != "existing-id" {
		t.Errorf("expected existing account ID, got %q", result.Account.ID)
	}
	// No new account should be persisted.
	if len(accounts.saved) != 1 {
		t.Errorf("expected 1 account (pre-existing), got %d", len(accounts.saved))
	}
}

func TestHandleOAuthCallbackTokenIssued(t *testing.T) {
	t.Parallel()

	accounts := &inMemoryAccountRepository{}
	h := auth.NewHandleOAuthCallbackHandler(accounts, testSigningKey, testTokenDuration)

	result, err := h.Handle(context.Background(), validCallbackCommand())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Token == "" {
		t.Error("expected a non-empty JWT token")
	}

	validator := auth.NewJWTValidator(testSigningKey)
	identity, err := validator.Validate(context.Background(), result.Token)
	if err != nil {
		t.Fatalf("issued token failed validation: %v", err)
	}
	if identity.AccountID != result.Account.ID {
		t.Errorf("token account ID mismatch: got %q, want %q", identity.AccountID, result.Account.ID)
	}
	if !identity.ExpiresAt.After(time.Now()) {
		t.Error("expected token to not be expired")
	}
}

func TestHandleOAuthCallbackEmptySubjectID(t *testing.T) {
	t.Parallel()
	h := auth.NewHandleOAuthCallbackHandler(&inMemoryAccountRepository{}, testSigningKey, testTokenDuration)
	cmd := auth.HandleOAuthCallbackCommand{GoogleSubjectID: "", Email: "user@example.com"}

	_, err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for empty subject ID")
	}
}

func TestHandleOAuthCallbackAccountLookupError(t *testing.T) {
	t.Parallel()

	lookupErr := errors.New("db failure")
	accounts := &inMemoryAccountRepository{err: lookupErr}
	h := auth.NewHandleOAuthCallbackHandler(accounts, testSigningKey, testTokenDuration)

	_, err := h.Handle(context.Background(), validCallbackCommand())
	if err == nil {
		t.Fatal("expected error when account lookup fails")
	}
}
