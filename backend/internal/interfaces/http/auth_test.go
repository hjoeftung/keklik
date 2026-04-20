package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/infrastructure"
)

const testSigningKey = "test-signing-key"
const testTokenDuration = 30 * 24 * time.Hour

// --- requireAuth middleware ---

func TestRequireAuth_MissingHeader(t *testing.T) {
	t.Parallel()

	handler := requireAuth(&stubAccountRepository{}, &stubTokenValidator{}, okHandler())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_WrongScheme(t *testing.T) {
	t.Parallel()

	handler := requireAuth(&stubAccountRepository{}, &stubTokenValidator{}, okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	t.Parallel()

	validator := &stubTokenValidator{err: auth.ErrInvalidToken}
	handler := requireAuth(&stubAccountRepository{}, validator, okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer unknown-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ExpiredSession(t *testing.T) {
	t.Parallel()

	validator := &stubTokenValidator{err: auth.ErrInvalidToken}
	handler := requireAuth(&stubAccountRepository{}, validator, okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ValidToken_AttachesAccountToContext(t *testing.T) {
	t.Parallel()

	account := auth.Account{ID: "acc-id", GoogleSubjectID: "google-sub"}
	validator := &stubTokenValidator{identity: auth.Identity{AccountID: "acc-id"}}
	accounts := &stubAccountRepository{account: account}

	var capturedAccount auth.Account
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAccount, _ = auth.AccountFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := requireAuth(accounts, validator, next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-tok")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if capturedAccount.ID != "acc-id" {
		t.Errorf("expected account ID %q in context, got %q", "acc-id", capturedAccount.ID)
	}
}

// --- state signing ---

func TestSignState_VerifyState_RoundTrip(t *testing.T) {
	t.Parallel()

	state := "random-nonce-abc123"
	secret := "super-secret"
	signed := signState(state, secret)

	if !verifyState(state, signed, secret) {
		t.Error("expected verifyState to return true for a valid signature")
	}
}

func TestVerifyState_WrongSecret(t *testing.T) {
	t.Parallel()

	signed := signState("state", "correct-secret")
	if verifyState("state", signed, "wrong-secret") {
		t.Error("expected verifyState to return false for wrong secret")
	}
}

func TestVerifyState_TamperedState(t *testing.T) {
	t.Parallel()

	signed := signState("original-state", "secret")
	if verifyState("tampered-state", signed, "secret") {
		t.Error("expected verifyState to return false for tampered state")
	}
}

func TestVerifyState_TamperedSignature(t *testing.T) {
	t.Parallel()

	state := "some-state"
	if verifyState(state, state+".deadbeef", "secret") {
		t.Error("expected verifyState to return false for forged signature")
	}
}

// --- OAuth start handler ---

func TestOAuthStartHandler_SetsStateCookieAndRedirects(t *testing.T) {
	t.Parallel()

	cfg := minimalOAuthConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/start", nil)
	oauthStartHandler(rec, req, cfg, "test-secret")

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location == "" {
		t.Fatal("expected a Location header")
	}

	var stateCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == oauthStateCookieName {
			stateCookie = c
			break
		}
	}
	if stateCookie == nil {
		t.Fatal("expected oauth_state cookie to be set")
	}
	if !stateCookie.HttpOnly {
		t.Error("expected oauth_state cookie to be HttpOnly")
	}
}

// --- OAuth callback handler: error cases that don't reach Google ---

func TestOAuthCallbackHandler_GoogleErrorParam(t *testing.T) {
	t.Parallel()

	cfg := minimalOAuthConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?error=access_denied", nil)
	oauthCallbackHandler(rec, req, cfg, "secret", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestOAuthCallbackHandler_MissingStateCookie(t *testing.T) {
	t.Parallel()

	cfg := minimalOAuthConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=some-state&code=code", nil)
	oauthCallbackHandler(rec, req, cfg, "secret", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestOAuthCallbackHandler_StateMismatch(t *testing.T) {
	t.Parallel()

	cfg := minimalOAuthConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=legit-state&code=code", nil)
	// Cookie contains a signature for a different state.
	req.AddCookie(&http.Cookie{
		Name:  oauthStateCookieName,
		Value: signState("different-state", "secret"),
	})
	oauthCallbackHandler(rec, req, cfg, "secret", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestTestLoginHandlerReturns404WhenDisabled(t *testing.T) {
	t.Parallel()

	server := NewServer(
		minimalServerConfig(),
		Dependencies{
			Accounts:  &stubAccountRepository{},
			Validator: &stubTokenValidator{},
			TestLogin: auth.NewHandleTestLoginHandler(&stubAccountRepository{}, testSigningKey, testTokenDuration),
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/test/login", bytes.NewBufferString(`{"identifier":"qa-user"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTestLoginHandlerReturnsSessionWhenEnabled(t *testing.T) {
	t.Parallel()

	accounts := &stubAccountRepository{}
	server := NewServer(
		infrastructure.Config{
			HTTP: infrastructure.HTTPConfig{Port: 8080},
			Auth: infrastructure.AuthConfig{EnableTestAuth: true, JWTSigningKey: testSigningKey},
		},
		Dependencies{
			Accounts:  accounts,
			Validator: &stubTokenValidator{},
			TestLogin: auth.NewHandleTestLoginHandler(accounts, testSigningKey, testTokenDuration),
		},
	)

	body, err := json.Marshal(testLoginRequest{Identifier: "qa-user"})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/test/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp authSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}

	if resp.AccountID == "" {
		t.Fatal("expected non-empty account_id")
	}

	if len(accounts.saved) != 1 {
		t.Fatalf("expected 1 saved account, got %d", len(accounts.saved))
	}

	if accounts.saved[0].GoogleSubjectID != "test:qa-user" {
		t.Fatalf("expected test subject ID %q, got %q", "test:qa-user", accounts.saved[0].GoogleSubjectID)
	}

}

func TestTestLoginHandlerRejectsBadJSON(t *testing.T) {
	t.Parallel()

	server := NewServer(
		infrastructure.Config{
			HTTP: infrastructure.HTTPConfig{Port: 8080},
			Auth: infrastructure.AuthConfig{EnableTestAuth: true},
		},
		Dependencies{
			Accounts:  &stubAccountRepository{},
			Validator: &stubTokenValidator{},
			TestLogin: auth.NewHandleTestLoginHandler(&stubAccountRepository{}, testSigningKey, testTokenDuration),
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/test/login", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// --- helpers ---

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// minimalOAuthConfig returns an oauth2.Config with enough fields set to build
// an AuthCodeURL without panicking. It does not point to real Google endpoints.
func minimalOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/auth/google/callback",
		Scopes:       []string{"openid", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}
}

func minimalServerConfig() infrastructure.Config {
	return infrastructure.Config{
		HTTP: infrastructure.HTTPConfig{Port: 8080},
	}
}
