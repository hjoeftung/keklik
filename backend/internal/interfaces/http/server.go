package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/time/rate"

	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
	"github.com/hjoeftung/keklik/internal/infrastructure"
	"github.com/hjoeftung/keklik/internal/sleep"
)

type healthResponse struct {
	Status string `json:"status"`
}

// Dependencies holds all handler and repository dependencies for the HTTP server.
type Dependencies struct {
	Accounts            auth.AccountRepository
	Validator           auth.TokenValidator
	OAuthCallback       *auth.HandleOAuthCallbackHandler
	TestLogin           *auth.HandleTestLoginHandler
	RefreshToken        *auth.HandleRefreshTokenHandler
	Logout              *auth.HandleLogoutHandler
	CreateFamily        *family.CreateFamilyHandler
	GetFamily           *family.GetFamilyHandler
	CreateInviteLink    *family.CreateFamilyInviteLinkHandler
	RevokeInviteLink    *family.RevokeInviteLinkHandler
	JoinFamilyByInvite  *family.JoinFamilyByInviteLinkHandler
	BabyAccess          babyAccessChecker
	CreateSleepProfile  *sleep.CreateSleepProfileHandler
	StartSleep          *sleep.StartSleepHandler
	StopSleep           *sleep.StopSleepHandler
	EditSleepSession    *sleep.EditSleepSessionHandler
	DeleteSleepSession  *sleep.DeleteSleepSessionHandler
	GetSleepHistory     *sleep.GetSleepHistoryHandler
	GetDashboardSummary *sleep.GetDashboardSummaryHandler
}

// NewServer wires the HTTP transport and returns a ready-to-start server.
func NewServer(config infrastructure.Config, deps Dependencies) *http.Server {
	accounts := deps.Accounts
	validator := deps.Validator
	oauthCallback := deps.OAuthCallback
	testLogin := deps.TestLogin
	refreshToken := deps.RefreshToken
	logout := deps.Logout
	createFamily := deps.CreateFamily
	getFamily := deps.GetFamily
	createInviteLink := deps.CreateInviteLink
	revokeInviteLink := deps.RevokeInviteLink
	joinFamilyByInvite := deps.JoinFamilyByInvite
	joinLimiter := newIPRateLimiter(rate.Every(time.Minute/5), 5)
	testLoginLimiter := newIPRateLimiter(rate.Every(time.Minute/5), 5)
	babyAccess := deps.BabyAccess
	createSleepProfile := deps.CreateSleepProfile
	startSleep := deps.StartSleep
	stopSleep := deps.StopSleep
	editSleepSession := deps.EditSleepSession
	deleteSleepSession := deps.DeleteSleepSession
	getSleepHistory := deps.GetSleepHistory
	getDashboardSummary := deps.GetDashboardSummary
	oauthCfg := &oauth2.Config{
		ClientID:     config.GoogleOAuth.ClientID,
		ClientSecret: config.GoogleOAuth.ClientSecret,
		RedirectURL:  config.GoogleOAuth.RedirectURL,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}
	stateSecret := config.GoogleOAuth.ClientSecret

	mux := http.NewServeMux()

	// Public endpoints.
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /auth/google/start", func(w http.ResponseWriter, r *http.Request) {
		oauthStartHandler(w, r, oauthCfg, stateSecret, !config.App.IsDev)
	})
	mux.HandleFunc("GET /auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		oauthCallbackHandler(w, r, oauthCfg, stateSecret, config.App.FrontendURL, oauthCallback)
	})
	mux.Handle("/auth/test/login", rateLimitByIP(testLoginLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testLoginHandler(w, r, config.Auth.EnableTestAuth, testLogin)
	})))
	mux.HandleFunc("POST /auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		refreshTokenHandler(w, r, refreshToken)
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
		logoutHandler(w, r, logout)
	})
	protected.HandleFunc("POST /families", func(w http.ResponseWriter, r *http.Request) {
		createFamilyHandler(w, r, createFamily)
	})
	protected.HandleFunc("GET /family", func(w http.ResponseWriter, r *http.Request) {
		getFamilyHandler(w, r, getFamily)
	})
	protected.HandleFunc("POST /families/invite-links", func(w http.ResponseWriter, r *http.Request) {
		createFamilyInviteLinkHandler(w, r, createInviteLink)
	})
	protected.HandleFunc("DELETE /families/invite-links/{token}", func(w http.ResponseWriter, r *http.Request) {
		revokeInviteLinkHandler(w, r, revokeInviteLink)
	})
	protected.Handle("POST /families/join-by-invite-link", rateLimitByIP(joinLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		joinFamilyByInviteLinkHandler(w, r, joinFamilyByInvite, accounts, config.Auth.EnableTestAuth)
	})))
	// Sleep endpoints — additionally wrapped with requireBabyAccess middleware.
	withBaby := func(h http.HandlerFunc) http.Handler {
		return requireBabyAccess(babyAccess, h)
	}
	protected.Handle("POST /babies/{baby_id}/sleep-profiles", withBaby(func(w http.ResponseWriter, r *http.Request) {
		createSleepProfileHandler(w, r, createSleepProfile)
	}))
	protected.Handle("POST /babies/{baby_id}/sleep-sessions/active", withBaby(func(w http.ResponseWriter, r *http.Request) {
		startSleepHandler(w, r, startSleep)
	}))
	protected.Handle("DELETE /babies/{baby_id}/sleep-sessions/active", withBaby(func(w http.ResponseWriter, r *http.Request) {
		stopSleepHandler(w, r, stopSleep)
	}))
	protected.Handle("PATCH /babies/{baby_id}/sleep-sessions/{id}", withBaby(func(w http.ResponseWriter, r *http.Request) {
		editSleepSessionHandler(w, r, editSleepSession)
	}))
	protected.Handle("DELETE /babies/{baby_id}/sleep-sessions/{id}", withBaby(func(w http.ResponseWriter, r *http.Request) {
		deleteSleepSessionHandler(w, r, deleteSleepSession)
	}))
	protected.Handle("GET /babies/{baby_id}/sleep-sessions", withBaby(func(w http.ResponseWriter, r *http.Request) {
		getSleepHistoryHandler(w, r, getSleepHistory)
	}))
	protected.Handle("GET /babies/{baby_id}/dashboard", withBaby(func(w http.ResponseWriter, r *http.Request) {
		getDashboardSummaryHandler(w, r, getDashboardSummary)
	}))
	mux.Handle("/", requireAuth(validator, protected))

	return &http.Server{
		Addr:    config.Address(),
		Handler: withRequestID(mux),
	}
}

// healthHandler returns a simple liveness check.
//
// @Summary  Health check
// @Tags     system
// @Produce  json
// @Success  200  {object}  healthResponse
// @Router   /healthz [get]
func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(writer).Encode(healthResponse{Status: "ok"})
}
