package httpapi

import (
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/time/rate"

	"github.com/hjoeftung/keklik/internal/infrastructure"
)

func addRoutes(mux *http.ServeMux, config infrastructure.Config, deps Dependencies) {
	joinLimiter := newIPRateLimiter(rate.Every(time.Minute/5), 5)
	testLoginLimiter := newIPRateLimiter(rate.Every(time.Minute/5), 5)
	oauthCfg := &oauth2.Config{
		ClientID:     config.GoogleOAuth.ClientID,
		ClientSecret: config.GoogleOAuth.ClientSecret,
		RedirectURL:  config.GoogleOAuth.RedirectURL,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}
	stateSecret := config.GoogleOAuth.ClientSecret
	cookieCfg := authCookieConfig{
		Secure:          !config.App.IsDev,
		AccessDuration:  config.Auth.AccessTokenDuration,
		RefreshDuration: config.Auth.RefreshTokenDuration,
	}

	// Public endpoints.
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /auth/google/start", func(w http.ResponseWriter, r *http.Request) {
		oauthStartHandler(w, r, oauthCfg, stateSecret, !config.App.IsDev)
	})
	mux.HandleFunc("GET /auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		oauthCallbackHandler(w, r, oauthCfg, stateSecret, config.App.FrontendURL, deps.OAuthCallback, cookieCfg)
	})
	mux.Handle("/auth/test/login", rateLimitByIP(testLoginLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testLoginHandler(w, r, config.Auth.EnableTestAuth, deps.TestLogin, cookieCfg)
	})))
	mux.HandleFunc("POST /auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		refreshTokenHandler(w, r, deps.RefreshToken, cookieCfg)
	})
	mux.HandleFunc("GET /swagger/index.html", func(w http.ResponseWriter, r *http.Request) {
		swaggerUIHandler(w, r, config.App.EnableSwaggerUI)
	})
	mux.HandleFunc("GET /swagger/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLHandler(w, r, config.App.EnableSwaggerUI)
	})

	// Protected endpoints — wrapped with requireAuth middleware.
	protected := http.NewServeMux()
	protected.HandleFunc("POST /auth/logout", func(w http.ResponseWriter, r *http.Request) {
		logoutHandler(w, r, deps.Logout)
	})
	protected.HandleFunc("POST /families", func(w http.ResponseWriter, r *http.Request) {
		createFamilyHandler(w, r, deps.CreateFamily)
	})
	protected.HandleFunc("GET /family", func(w http.ResponseWriter, r *http.Request) {
		getFamilyHandler(w, r, deps.GetFamily)
	})
	protected.HandleFunc("POST /families/invite-links", func(w http.ResponseWriter, r *http.Request) {
		createFamilyInviteLinkHandler(w, r, deps.CreateInviteLink)
	})
	protected.HandleFunc("DELETE /families/invite-links/{token}", func(w http.ResponseWriter, r *http.Request) {
		revokeInviteLinkHandler(w, r, deps.RevokeInviteLink)
	})
	protected.Handle("POST /families/join-by-invite-link", rateLimitByIP(joinLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		joinFamilyByInviteLinkHandler(w, r, deps.JoinFamilyByInvite, deps.Accounts, config.Auth.EnableTestAuth)
	})))

	// Sleep endpoints — additionally wrapped with requireBabyAccess middleware.
	withBaby := func(h http.HandlerFunc) http.Handler {
		return requireBabyAccess(deps.BabyAccess, h)
	}
	protected.Handle("POST /babies/{baby_id}/night-windows", withBaby(func(w http.ResponseWriter, r *http.Request) {
		setNightWindowHandler(w, r, deps.SetNightWindow)
	}))
	protected.Handle("POST /babies/{baby_id}/sleep-sessions", withBaby(func(w http.ResponseWriter, r *http.Request) {
		logPastSleepHandler(w, r, deps.LogPastSleep)
	}))
	protected.Handle("POST /babies/{baby_id}/sleep-sessions/active", withBaby(func(w http.ResponseWriter, r *http.Request) {
		startSleepHandler(w, r, deps.StartSleep)
	}))
	protected.Handle("DELETE /babies/{baby_id}/sleep-sessions/active", withBaby(func(w http.ResponseWriter, r *http.Request) {
		stopSleepHandler(w, r, deps.StopSleep)
	}))
	protected.Handle("PATCH /babies/{baby_id}/sleep-sessions/{id}", withBaby(func(w http.ResponseWriter, r *http.Request) {
		editSleepSessionHandler(w, r, deps.EditSleepSession)
	}))
	protected.Handle("DELETE /babies/{baby_id}/sleep-sessions/{id}", withBaby(func(w http.ResponseWriter, r *http.Request) {
		deleteSleepSessionHandler(w, r, deps.DeleteSleepSession)
	}))
	protected.Handle("GET /babies/{baby_id}/sleep-sessions", withBaby(func(w http.ResponseWriter, r *http.Request) {
		getSleepHistoryHandler(w, r, deps.GetSleepHistory)
	}))
	protected.Handle("GET /babies/{baby_id}/sleep-stats", withBaby(func(w http.ResponseWriter, r *http.Request) {
		getSleepStatsHandler(w, r, deps.GetSleepStats)
	}))
	mux.Handle("/", requireAuth(deps.Validator, protected))
}
