package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
)

const oauthStateCookieName = "oauth_state"
const oauthStateCookieMaxAge = 300 // 5 minutes

type googleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

type authSessionResponse struct {
	Token     string `json:"token"`
	AccountID string `json:"account_id"`
}

type testLoginRequest struct {
	Identifier string `json:"identifier"`
}

// oauthStartHandler generates a random state, stores it in a signed cookie, and
// redirects the client to Google's authorisation page.
func oauthStartHandler(w http.ResponseWriter, r *http.Request, cfg *oauth2.Config, stateSecret string) {
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		writeError(w, apperror.New(apperror.CodeInternalError, "failed to generate state"))
		return
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    signState(state, stateSecret),
		MaxAge:   oauthStateCookieMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.Redirect(w, r, cfg.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// oauthCallbackHandler verifies the state, exchanges the code for a Google token,
// fetches the user's identity from Google, then resolves or provisions an internal
// Account and issues a session token.
func oauthCallbackHandler(
	w http.ResponseWriter,
	r *http.Request,
	cfg *oauth2.Config,
	stateSecret string,
	h *auth.HandleOAuthCallbackHandler,
) {
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, fmt.Sprintf("google oauth error: %s", errParam)))
		return
	}

	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || !verifyState(state, cookie.Value, stateSecret) {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "invalid oauth state"))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "missing authorization code"))
		return
	}

	token, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "failed to exchange authorization code"))
		return
	}

	userInfo, err := fetchGoogleUserInfo(r.Context(), cfg, token)
	if err != nil {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "failed to retrieve google identity"))
		return
	}

	result, err := h.Handle(r.Context(), auth.HandleOAuthCallbackCommand{
		GoogleSubjectID: userInfo.Sub,
		Email:           userInfo.Email,
	})
	if err != nil {
		writeError(w, apperror.New(apperror.CodeInternalError, "failed to resolve account"))
		return
	}

	// Clear the state cookie now that it has been consumed.
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	writeAuthSessionResponse(w, result)
}

// testLoginHandler issues a regular application session for a test-only identity.
func testLoginHandler(w http.ResponseWriter, r *http.Request, enabled bool, h *auth.HandleTestLoginHandler) {
	if !enabled {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h == nil {
		writeError(w, apperror.New(apperror.CodeInternalError, "test auth is unavailable"))
		return
	}

	var req testLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(apperror.CodeInvalidArgument, "invalid JSON body"))
		return
	}

	result, err := h.Handle(r.Context(), auth.HandleTestLoginCommand{Identifier: req.Identifier})
	if err != nil {
		writeError(w, apperror.New(apperror.CodeInvalidArgument, err.Error()))
		return
	}

	writeAuthSessionResponse(w, result)
}

// requireAuth is middleware that validates the Bearer session token in the Authorization
// header, loads the associated Account, and attaches it to the request context.
// Protected handlers retrieve the account via auth.AccountFromContext.
func requireAuth(accounts auth.AccountRepository, sessions auth.SessionRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		if !strings.HasPrefix(bearer, "Bearer ") {
			writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
			return
		}

		token := auth.SessionToken(strings.TrimPrefix(bearer, "Bearer "))

		session, err := sessions.FindByToken(r.Context(), token)
		if err != nil || session.IsExpired(time.Now()) {
			writeError(w, apperror.New(apperror.CodeUnauthenticated, "invalid or expired session"))
			return
		}

		account, err := accounts.FindByID(r.Context(), session.AccountID)
		if err != nil {
			writeError(w, apperror.New(apperror.CodeUnauthenticated, "account not found"))
			return
		}

		next.ServeHTTP(w, r.WithContext(auth.WithAccount(r.Context(), account)))
	})
}

// fetchGoogleUserInfo calls Google's userinfo endpoint using the given OAuth2 token.
func fetchGoogleUserInfo(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (googleUserInfo, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return googleUserInfo{}, fmt.Errorf("userinfo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return googleUserInfo{}, fmt.Errorf("userinfo returned %d: %s", resp.StatusCode, body)
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return googleUserInfo{}, fmt.Errorf("decode userinfo: %w", err)
	}

	return info, nil
}

func writeAuthSessionResponse(w http.ResponseWriter, result auth.HandleOAuthCallbackResult) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(authSessionResponse{
		Token:     string(result.Session.Token),
		AccountID: string(result.Account.ID),
	})
}

// signState returns "state.HMAC" where HMAC is computed over state using secret.
func signState(state, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(state))
	return state + "." + hex.EncodeToString(mac.Sum(nil))
}

// verifyState checks that signed == signState(state, secret).
func verifyState(state, signed, secret string) bool {
	expected := signState(state, secret)
	return hmac.Equal([]byte(signed), []byte(expected))
}
