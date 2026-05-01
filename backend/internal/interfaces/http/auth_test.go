package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/infrastructure"
)

const testSigningKey = "test-signing-key"
const testAccessTokenDuration = 15 * time.Minute
const testRefreshTokenDuration = 30 * 24 * time.Hour

// --- requireAuth middleware ---

func TestRequireAuth_MissingCookie(t *testing.T) {
	t.Parallel()

	handler := requireAuth(&stubTokenValidator{}, okHandler())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	t.Parallel()

	validator := &stubTokenValidator{err: auth.ErrInvalidToken}
	handler := requireAuth(validator, okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: accessCookieName, Value: "unknown-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ExpiredSession(t *testing.T) {
	t.Parallel()

	validator := &stubTokenValidator{err: auth.ErrInvalidToken}
	handler := requireAuth(validator, okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: accessCookieName, Value: "expired-tok"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ValidToken_AttachesAccountIDToContext(t *testing.T) {
	t.Parallel()

	validator := &stubTokenValidator{identity: auth.Identity{AccountID: "acc-id"}}

	var capturedID auth.AccountID
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID, _ = auth.AccountIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := requireAuth(validator, next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: accessCookieName, Value: "valid-tok"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if capturedID != "acc-id" {
		t.Errorf("expected account ID %q in context, got %q", "acc-id", capturedID)
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

// --- state timestamp ---

func TestBuildState_StateTimestampValid_RoundTrip(t *testing.T) {
	t.Parallel()

	now := int64(1000000)
	state := buildState("nonce123", now)
	if !stateTimestampValid(state, now) {
		t.Error("expected stateTimestampValid to return true immediately after buildState")
	}
}

func TestStateTimestampValid_WithinWindow(t *testing.T) {
	t.Parallel()

	ts := int64(1000000)
	state := buildState("nonce", ts)
	if !stateTimestampValid(state, ts+299) {
		t.Error("expected state within 5-minute window to be valid")
	}
}

func TestStateTimestampValid_AtBoundary(t *testing.T) {
	t.Parallel()

	ts := int64(1000000)
	state := buildState("nonce", ts)
	if !stateTimestampValid(state, ts+300) {
		t.Error("expected state exactly at the 5-minute boundary to be valid")
	}
}

func TestStateTimestampValid_Expired(t *testing.T) {
	t.Parallel()

	ts := int64(1000000)
	state := buildState("nonce", ts)
	if stateTimestampValid(state, ts+301) {
		t.Error("expected state beyond 5-minute window to be invalid")
	}
}

func TestStateTimestampValid_FutureTimestamp(t *testing.T) {
	t.Parallel()

	ts := int64(1000000)
	state := buildState("nonce", ts)
	if stateTimestampValid(state, ts-1) {
		t.Error("expected state with future timestamp to be invalid")
	}
}

func TestStateTimestampValid_MalformedState(t *testing.T) {
	t.Parallel()

	if stateTimestampValid("no-dot-separator", 1000000) {
		t.Error("expected malformed state (no dot) to be invalid")
	}
	if stateTimestampValid("nonce.!!invalid-base64!!", 1000000) {
		t.Error("expected state with invalid base64 timestamp to be invalid")
	}
}

// --- OAuth start handler ---

func TestOAuthStartHandler_SetsStateCookieAndRedirects(t *testing.T) {
	t.Parallel()

	cfg := minimalOAuthConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/start", nil)
	oauthStartHandler(rec, req, cfg, "test-secret", false)

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

const testFrontendURL = "http://localhost:5173"

func TestOAuthCallbackHandler_GoogleErrorParam(t *testing.T) {
	t.Parallel()

	cfg := minimalOAuthConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?error=access_denied", nil)
	oauthCallbackHandler(rec, req, cfg, "secret", testFrontendURL, nil, authCookieConfig{})

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" {
		t.Fatal("expected Location header")
	}
	if !strings.Contains(loc, "error=google_error") {
		t.Errorf("expected error=google_error in redirect, got %q", loc)
	}
}

func TestOAuthCallbackHandler_MissingStateCookie(t *testing.T) {
	t.Parallel()

	cfg := minimalOAuthConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=some-state&code=code", nil)
	oauthCallbackHandler(rec, req, cfg, "secret", testFrontendURL, nil, authCookieConfig{})

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "error=invalid_state") {
		t.Errorf("expected error=invalid_state in redirect, got %q", loc)
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
	oauthCallbackHandler(rec, req, cfg, "secret", testFrontendURL, nil, authCookieConfig{})

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "error=invalid_state") {
		t.Errorf("expected error=invalid_state in redirect, got %q", loc)
	}
}

func TestTestLoginHandlerReturns401WhenDisabled(t *testing.T) {
	t.Parallel()

	server := NewServer(
		minimalServerConfig(),
		Dependencies{
			Accounts:  &stubAccountRepository{},
			Validator: &stubTokenValidator{},
			TestLogin: auth.NewHandleTestLoginHandler(&stubAccountRepository{}, &stubRefreshTokenRepository{}, testSigningKey, testAccessTokenDuration, testRefreshTokenDuration),
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/test/login", bytes.NewBufferString(`{"identifier":"qa-user"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
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
			TestLogin: auth.NewHandleTestLoginHandler(accounts, &stubRefreshTokenRepository{}, testSigningKey, testAccessTokenDuration, testRefreshTokenDuration),
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

	var resp accountIDResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.AccountID == "" {
		t.Fatal("expected non-empty account_id")
	}

	var accessCookie, refreshCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		switch c.Name {
		case accessCookieName:
			accessCookie = c
		case refreshCookieName:
			refreshCookie = c
		}
	}
	if accessCookie == nil || accessCookie.Value == "" {
		t.Fatal("expected keklik_access cookie to be set")
	}
	if refreshCookie == nil || refreshCookie.Value == "" {
		t.Fatal("expected keklik_refresh cookie to be set")
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
			TestLogin: auth.NewHandleTestLoginHandler(&stubAccountRepository{}, &stubRefreshTokenRepository{}, testSigningKey, testAccessTokenDuration, testRefreshTokenDuration),
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
