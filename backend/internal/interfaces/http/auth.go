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
	"net/url"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"golang.org/x/oauth2"

	"github.com/google/uuid"
	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/sleep"
)

// babyAccessChecker verifies the caller is a family member of a given baby.
type babyAccessChecker interface {
	CheckBabyAccess(ctx context.Context, accountID auth.AccountID, babyID sleep.BabyID) (sleep.FamilyMemberID, error)
}

const oauthStateCookieName = "oauth_state"
const oauthStateCookieMaxAge = 300 // 5 minutes

type googleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

type accountIDResponse struct {
	AccountID string `json:"account_id"`
}

type testLoginRequest struct {
	Identifier string `json:"identifier"`
}

const accessCookieName = "keklik_access"
const refreshCookieName = "keklik_refresh"

type authCookieConfig struct {
	Secure          bool
	AccessDuration  time.Duration
	RefreshDuration time.Duration
}

func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, cfg authCookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    accessToken,
		MaxAge:   int(cfg.AccessDuration.Seconds()),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		MaxAge:   int(cfg.RefreshDuration.Seconds()),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	// Evict any stale cookie left over from the previous Path:/auth/refresh era.
	// Browsers send the more-specific path first, so the old cookie would shadow the new one.
	http.SetCookie(w, &http.Cookie{Name: refreshCookieName, Value: "", MaxAge: -1, Path: "/auth/refresh"})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: accessCookieName, Value: "", MaxAge: -1, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: refreshCookieName, Value: "", MaxAge: -1, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: refreshCookieName, Value: "", MaxAge: -1, Path: "/auth/refresh"})
}

// oauthStartHandler generates a random state, stores it in a signed cookie, and
// redirects the client to Google's authorisation page.
//
// @Summary  Start Google OAuth flow
// @Tags     auth
// @Success  307  {string}  string  "Redirect to Google authorisation page"
// @Router   /auth/google/start [get]
func oauthStartHandler(w http.ResponseWriter, r *http.Request, cfg *oauth2.Config, stateSecret string, secureCookie bool) {
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "failed to generate state"))
		return
	}
	nonce := base64.RawURLEncoding.EncodeToString(stateBytes)
	state := buildState(nonce, time.Now().Unix())

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    signState(state, stateSecret),
		MaxAge:   oauthStateCookieMaxAge,
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.Redirect(w, r, cfg.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// oauthCallbackHandler verifies the state, exchanges the code for a Google token,
// fetches the user's identity from Google, then resolves or provisions an internal
// Account and redirects the browser to the frontend with session cookies.
//
// @Summary   Google OAuth callback
// @Tags      auth
// @Param     code   query  string  true  "Authorization code returned by Google"
// @Param     state  query  string  true  "State value for CSRF verification"
// @Success   302
// @Failure   302
// @Router    /auth/google/callback [get]
func oauthCallbackHandler(
	w http.ResponseWriter,
	r *http.Request,
	cfg *oauth2.Config,
	stateSecret string,
	frontendURL string,
	h *auth.HandleOAuthCallbackHandler,
	cookieCfg authCookieConfig,
) {
	redirectError := func(code string) {
		dest, _ := url.Parse(frontendURL + "/")
		q := dest.Query()
		q.Set("error", code)
		dest.RawQuery = q.Encode()
		http.Redirect(w, r, dest.String(), http.StatusFound)
	}

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		redirectError("google_error")
		return
	}

	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || !verifyState(state, cookie.Value, stateSecret) {
		redirectError("invalid_state")
		return
	}
	if !stateTimestampValid(state, time.Now().Unix()) {
		redirectError("state_expired")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		redirectError("missing_code")
		return
	}

	token, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		slog.ErrorContext(r.Context(), "oauth_token_exchange_failed", "error", err)
		redirectError("token_exchange_failed")
		return
	}

	userInfo, err := fetchGoogleUserInfo(r.Context(), cfg, token)
	if err != nil {
		slog.ErrorContext(r.Context(), "oauth_identity_fetch_failed", "error", err)
		redirectError("identity_fetch_failed")
		return
	}

	result, err := h.Handle(r.Context(), auth.HandleOAuthCallbackCommand{
		GoogleSubjectID: userInfo.Sub,
		Email:           userInfo.Email,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "oauth_account_resolution_failed", "error", err)
		redirectError("account_resolution_failed")
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

	setAuthCookies(w, result.AccessToken, result.RefreshToken, cookieCfg)

	dest, _ := url.Parse(frontendURL + "/auth/callback")
	q := dest.Query()
	q.Set("account_id", string(result.Account.ID))
	dest.RawQuery = q.Encode()
	http.Redirect(w, r, dest.String(), http.StatusFound)
}

// testLoginHandler issues a regular application session for a test-only identity.
// Only available when ENABLE_TEST_AUTH=true.
//
// @Summary   Test login (dev/staging only)
// @Tags      auth
// @Accept    json
// @Produce   json
// @Param     body  body      testLoginRequest  true  "Test login credentials"
// @Success   200   {object}  accountIDResponse
// @Failure   400   {object}  errorResponse
// @Router    /auth/test/login [post]
func testLoginHandler(w http.ResponseWriter, r *http.Request, enabled bool, h *auth.HandleTestLoginHandler, cfg authCookieConfig) {
	if !enabled {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h == nil {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "test auth is unavailable"))
		return
	}

	var req testLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid JSON body"))
		return
	}

	result, err := h.Handle(r.Context(), auth.HandleTestLoginCommand{Identifier: req.Identifier})
	if err != nil {
		writeError(w, r, apperror.New(apperror.CodeInvalidArgument, err.Error()))
		return
	}

	setAuthCookies(w, result.AccessToken, result.RefreshToken, cfg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(accountIDResponse{AccountID: string(result.Account.ID)})
}

// refreshTokenHandler exchanges a valid refresh token for a new access token and rotated
// refresh token.
//
// @Summary   Refresh access token
// @Tags      auth
// @Success   204
// @Failure   401   {object}  errorResponse
// @Router    /auth/refresh [post]
func refreshTokenHandler(w http.ResponseWriter, r *http.Request, h *auth.HandleRefreshTokenHandler, cfg authCookieConfig) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		slog.WarnContext(r.Context(), "refresh_cookie_missing")
		clearAuthCookies(w)
		writeError(w, r, apperror.New(apperror.CodeUnauthenticated, "invalid or expired refresh token"))
		return
	}

	result, err := h.Handle(r.Context(), auth.HandleRefreshTokenCommand{Token: cookie.Value})
	if err != nil {
		slog.WarnContext(r.Context(), "refresh_token_invalid", "reason", err.Error())
		clearAuthCookies(w)
		writeError(w, r, apperror.New(apperror.CodeUnauthenticated, "invalid or expired refresh token"))
		return
	}

	setAuthCookies(w, result.AccessToken, result.RefreshToken, cfg)
	w.WriteHeader(http.StatusNoContent)
}

// logoutHandler revokes all refresh tokens for the authenticated account.
//
// @Summary   Logout
// @Tags      auth
// @Security  BearerAuth
// @Success   204
// @Failure   401   {object}  errorResponse
// @Router    /auth/logout [post]
func logoutHandler(w http.ResponseWriter, r *http.Request, h *auth.HandleLogoutHandler) {
	accountID, ok := auth.AccountIDFromContext(r.Context())
	if !ok {
		writeError(w, r, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	if err := h.Handle(r.Context(), auth.HandleLogoutCommand{AccountID: accountID}); err != nil {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "logout failed"))
		return
	}

	clearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

// requireBabyAccess is middleware that extracts {baby_id} from the request path, verifies
// the authenticated caller is a family member of that baby, and stores (baby_id, member_id)
// in the request context. Must be applied after requireAuth.
func requireBabyAccess(checker babyAccessChecker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := auth.AccountIDFromContext(r.Context())
		if !ok {
			writeError(w, r, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
			return
		}

		rawBabyID := r.PathValue("baby_id")
		if _, err := uuid.Parse(rawBabyID); err != nil {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid baby_id"))
			return
		}
		babyID := sleep.BabyID(rawBabyID)

		memberID, err := checker.CheckBabyAccess(r.Context(), accountID, babyID)
		if err != nil {
			var appErr apperror.AppError
			if asErr, ok2 := err.(apperror.AppError); ok2 {
				appErr = asErr
			} else {
				appErr = apperror.Wrap(apperror.CodeInternalError, "unexpected error", err)
			}
			writeError(w, r, appErr)
			return
		}

		ctx := withBabyContext(r.Context(), babyID, memberID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireAuth is middleware that validates the keklik_access cookie and attaches
// the account ID to the request context.
// Protected handlers retrieve the account ID via auth.AccountIDFromContext.
func requireAuth(validator auth.TokenValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(accessCookieName)
		if err != nil {
			writeError(w, r, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
			return
		}

		identity, err := validator.Validate(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, r, apperror.New(apperror.CodeUnauthenticated, "invalid or expired session"))
			return
		}

		next.ServeHTTP(w, r.WithContext(auth.WithAccountID(r.Context(), identity.AccountID)))
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

// buildState returns "nonce.base64(ts)" embedding the Unix timestamp in the state value.
func buildState(nonce string, ts int64) string {
	tsEncoded := base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(ts, 10)))
	return nonce + "." + tsEncoded
}

// stateTimestampValid returns true if the timestamp embedded in state is within 5 minutes of now.
func stateTimestampValid(state string, now int64) bool {
	parts := strings.SplitN(state, ".", 2)
	if len(parts) != 2 {
		return false
	}
	tsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	ts, err := strconv.ParseInt(string(tsBytes), 10, 64)
	if err != nil {
		return false
	}
	diff := now - ts
	return diff >= 0 && diff <= oauthStateCookieMaxAge
}
