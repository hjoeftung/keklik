package httpapi

import (
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

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
	Accounts           auth.AccountRepository
	Sessions           auth.SessionRepository
	OAuthCallback      *auth.HandleOAuthCallbackHandler
	TestLogin          *auth.HandleTestLoginHandler
	CreateFamily       *family.CreateFamilyHandler
	GetFamily          *family.GetFamilyHandler
	CreateInviteLink   *family.CreateFamilyInviteLinkHandler
	JoinFamilyByInvite *family.JoinFamilyByInviteLinkHandler
	BabyAccess         babyAccessChecker
	CreateSleepProfile *sleep.CreateSleepProfileHandler
	StartSleep         *sleep.StartSleepHandler
	StopSleep          *sleep.StopSleepHandler
	EditSleepSession   *sleep.EditSleepSessionHandler
	DeleteSleepSession *sleep.DeleteSleepSessionHandler
	GetSleepHistory    *sleep.GetSleepHistoryHandler
}

// NewServer wires the HTTP transport and returns a ready-to-start server.
func NewServer(config infrastructure.Config, deps Dependencies) *http.Server {
	accounts := deps.Accounts
	sessions := deps.Sessions
	oauthCallback := deps.OAuthCallback
	testLogin := deps.TestLogin
	createFamily := deps.CreateFamily
	getFamily := deps.GetFamily
	createInviteLink := deps.CreateInviteLink
	joinFamilyByInvite := deps.JoinFamilyByInvite
	babyAccess := deps.BabyAccess
	createSleepProfile := deps.CreateSleepProfile
	startSleep := deps.StartSleep
	stopSleep := deps.StopSleep
	editSleepSession := deps.EditSleepSession
	deleteSleepSession := deps.DeleteSleepSession
	getSleepHistory := deps.GetSleepHistory
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
		oauthStartHandler(w, r, oauthCfg, stateSecret)
	})
	mux.HandleFunc("GET /auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		oauthCallbackHandler(w, r, oauthCfg, stateSecret, oauthCallback)
	})
	mux.HandleFunc("/auth/test/login", func(w http.ResponseWriter, r *http.Request) {
		testLoginHandler(w, r, config.Auth.EnableTestAuth, testLogin)
	})
	mux.HandleFunc("GET /swagger/index.html", func(w http.ResponseWriter, r *http.Request) {
		swaggerUIHandler(w, r, config.App.EnableSwaggerUI)
	})
	mux.HandleFunc("GET /swagger/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLHandler(w, r, config.App.EnableSwaggerUI)
	})

	// Protected endpoints — wrapped with requireAuth middleware.
	protected := http.NewServeMux()
	protected.HandleFunc("POST /families", func(w http.ResponseWriter, r *http.Request) {
		createFamilyHandler(w, r, createFamily)
	})
	protected.HandleFunc("GET /family", func(w http.ResponseWriter, r *http.Request) {
		getFamilyHandler(w, r, getFamily)
	})
	protected.HandleFunc("POST /families/invite-links", func(w http.ResponseWriter, r *http.Request) {
		createFamilyInviteLinkHandler(w, r, createInviteLink)
	})
	protected.HandleFunc("POST /families/join-by-invite-link", func(w http.ResponseWriter, r *http.Request) {
		joinFamilyByInviteLinkHandler(w, r, joinFamilyByInvite)
	})
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
	mux.Handle("/", requireAuth(accounts, sessions, protected))

	return &http.Server{
		Addr:    config.Address(),
		Handler: mux,
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
