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

type inMemorySessionRepository struct {
	saved []auth.Session
	err   error
}

func (r *inMemorySessionRepository) Save(_ context.Context, s auth.Session) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, s)
	return nil
}

func (r *inMemorySessionRepository) FindByToken(_ context.Context, token auth.SessionToken) (auth.Session, error) {
	for _, s := range r.saved {
		if s.Token == token {
			return s, nil
		}
	}
	return auth.Session{}, auth.ErrSessionNotFound
}

// --- helpers ---

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
	sessions := &inMemorySessionRepository{}
	h := auth.NewHandleOAuthCallbackHandler(accounts, sessions)

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
	sessions := &inMemorySessionRepository{}
	h := auth.NewHandleOAuthCallbackHandler(accounts, sessions)

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

func TestHandleOAuthCallbackSessionIssued(t *testing.T) {
	t.Parallel()

	accounts := &inMemoryAccountRepository{}
	sessions := &inMemorySessionRepository{}
	h := auth.NewHandleOAuthCallbackHandler(accounts, sessions)

	result, err := h.Handle(context.Background(), validCallbackCommand())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Session.Token == "" {
		t.Error("expected a non-empty session token")
	}
	if result.Session.AccountID != result.Account.ID {
		t.Errorf("session account ID mismatch: got %q, want %q", result.Session.AccountID, result.Account.ID)
	}
	if result.Session.IsExpired(time.Now()) {
		t.Error("expected session to not be expired")
	}
	if len(sessions.saved) != 1 {
		t.Errorf("expected 1 saved session, got %d", len(sessions.saved))
	}
}

func TestHandleOAuthCallbackEmptySubjectID(t *testing.T) {
	t.Parallel()
	h := auth.NewHandleOAuthCallbackHandler(&inMemoryAccountRepository{}, &inMemorySessionRepository{})
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
	h := auth.NewHandleOAuthCallbackHandler(accounts, &inMemorySessionRepository{})

	_, err := h.Handle(context.Background(), validCallbackCommand())
	if err == nil {
		t.Fatal("expected error when account lookup fails")
	}
}

func TestHandleOAuthCallbackSessionSaveError(t *testing.T) {
	t.Parallel()

	accounts := &inMemoryAccountRepository{}
	sessions := &inMemorySessionRepository{err: errors.New("db failure")}
	h := auth.NewHandleOAuthCallbackHandler(accounts, sessions)

	_, err := h.Handle(context.Background(), validCallbackCommand())
	if err == nil {
		t.Fatal("expected error when session save fails")
	}
}
