package auth_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/auth"
)

// --- test doubles ---

type inMemoryAccountRepository struct {
	mu    sync.Mutex
	saved []auth.Account
	err   error
}

func (r *inMemoryAccountRepository) Save(_ context.Context, a auth.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, a)
	return nil
}

func (r *inMemoryAccountRepository) FindByID(_ context.Context, id auth.AccountID) (auth.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.saved {
		if a.ID == id {
			return a, nil
		}
	}
	return auth.Account{}, auth.ErrAccountNotFound
}

func (r *inMemoryAccountRepository) FindByGoogleSubjectID(_ context.Context, sub string) (auth.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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

func (r *inMemoryAccountRepository) Upsert(_ context.Context, a auth.Account) (auth.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return auth.Account{}, r.err
	}
	for _, existing := range r.saved {
		if existing.GoogleSubjectID == a.GoogleSubjectID {
			return existing, nil
		}
	}
	r.saved = append(r.saved, a)
	return a, nil
}

type inMemoryRefreshTokenRepository struct {
	mu     sync.Mutex
	tokens []auth.RefreshToken
	err    error
}

func (r *inMemoryRefreshTokenRepository) Save(_ context.Context, t auth.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	r.tokens = append(r.tokens, t)
	return nil
}

func (r *inMemoryRefreshTokenRepository) FindByToken(_ context.Context, token string) (auth.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, t := range r.tokens {
		if t.Token == token {
			return r.tokens[i], nil
		}
	}
	return auth.RefreshToken{}, auth.ErrRefreshTokenNotFound
}

func (r *inMemoryRefreshTokenRepository) Revoke(_ context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for i, t := range r.tokens {
		if t.Token == token {
			r.tokens[i].RevokedAt = &now
			return nil
		}
	}
	return auth.ErrRefreshTokenNotFound
}

func (r *inMemoryRefreshTokenRepository) RevokeAllForAccount(_ context.Context, accountID auth.AccountID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for i, t := range r.tokens {
		if t.AccountID == accountID && t.RevokedAt == nil {
			r.tokens[i].RevokedAt = &now
		}
	}
	return nil
}

// --- helpers ---

const testSigningKey = "test-signing-key"
const testAccessTokenDuration = 15 * time.Minute
const testRefreshTokenDuration = 30 * 24 * time.Hour

func validCallbackCommand() auth.HandleOAuthCallbackCommand {
	return auth.HandleOAuthCallbackCommand{
		GoogleSubjectID: "google-sub-12345",
		Email:           "user@example.com",
	}
}

func newCallbackHandler(accounts *inMemoryAccountRepository, refreshTokens *inMemoryRefreshTokenRepository) *auth.HandleOAuthCallbackHandler {
	return auth.NewHandleOAuthCallbackHandler(accounts, refreshTokens, testSigningKey, testAccessTokenDuration, testRefreshTokenDuration)
}

// --- tests ---

func TestHandleOAuthCallbackNewAccount(t *testing.T) {
	t.Parallel()

	accounts := &inMemoryAccountRepository{}
	refreshTokens := &inMemoryRefreshTokenRepository{}
	h := newCallbackHandler(accounts, refreshTokens)

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
	refreshTokens := &inMemoryRefreshTokenRepository{}
	h := newCallbackHandler(accounts, refreshTokens)

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
	refreshTokens := &inMemoryRefreshTokenRepository{}
	h := newCallbackHandler(accounts, refreshTokens)

	result, err := h.Handle(context.Background(), validCallbackCommand())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("expected a non-empty access token")
	}

	validator := auth.NewJWTValidator(testSigningKey)
	identity, err := validator.Validate(context.Background(), result.AccessToken)
	if err != nil {
		t.Fatalf("issued access token failed validation: %v", err)
	}
	if identity.AccountID != result.Account.ID {
		t.Errorf("token account ID mismatch: got %q, want %q", identity.AccountID, result.Account.ID)
	}
	if !identity.ExpiresAt.After(time.Now()) {
		t.Error("expected token to not be expired")
	}

	if result.RefreshToken == "" {
		t.Error("expected a non-empty refresh token")
	}
	if len(refreshTokens.tokens) != 1 {
		t.Errorf("expected 1 saved refresh token, got %d", len(refreshTokens.tokens))
	}
	if refreshTokens.tokens[0].AccountID != result.Account.ID {
		t.Errorf("refresh token account ID mismatch")
	}
}

func TestHandleOAuthCallbackEmptySubjectID(t *testing.T) {
	t.Parallel()
	h := newCallbackHandler(&inMemoryAccountRepository{}, &inMemoryRefreshTokenRepository{})
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
	h := newCallbackHandler(accounts, &inMemoryRefreshTokenRepository{})

	_, err := h.Handle(context.Background(), validCallbackCommand())
	if err == nil {
		t.Fatal("expected error when account lookup fails")
	}
}

func TestHandleOAuthCallbackConcurrentFirstTimeLogin(t *testing.T) {
	t.Parallel()

	accounts := &inMemoryAccountRepository{}
	refreshTokens := &inMemoryRefreshTokenRepository{}
	h := newCallbackHandler(accounts, refreshTokens)
	cmd := validCallbackCommand()

	const n = 20
	results := make([]auth.HandleOAuthCallbackResult, n)
	errs := make([]error, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = h.Handle(context.Background(), cmd)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d got error: %v", i, err)
		}
	}

	firstID := results[0].Account.ID
	for i, r := range results {
		if r.Account.ID != firstID {
			t.Errorf("goroutine %d resolved to account %q, want %q", i, r.Account.ID, firstID)
		}
	}

	accounts.mu.Lock()
	saved := len(accounts.saved)
	accounts.mu.Unlock()
	if saved != 1 {
		t.Errorf("expected exactly 1 saved account, got %d", saved)
	}
}
