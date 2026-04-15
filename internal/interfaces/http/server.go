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
	CreateInviteLink   *family.CreateFamilyInviteLinkHandler
	JoinFamilyByInvite *family.JoinFamilyByInviteLinkHandler
	SleepCtx           sleepContextResolver
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
	createInviteLink := deps.CreateInviteLink
	joinFamilyByInvite := deps.JoinFamilyByInvite
	sleepCtx := deps.SleepCtx
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

	// Protected endpoints — wrapped with requireAuth middleware.
	protected := http.NewServeMux()
	protected.HandleFunc("POST /families", func(w http.ResponseWriter, r *http.Request) {
		createFamilyHandler(w, r, createFamily)
	})
	protected.HandleFunc("POST /families/invite-links", func(w http.ResponseWriter, r *http.Request) {
		createFamilyInviteLinkHandler(w, r, createInviteLink)
	})
	protected.HandleFunc("POST /families/join-by-invite-link", func(w http.ResponseWriter, r *http.Request) {
		joinFamilyByInviteLinkHandler(w, r, joinFamilyByInvite)
	})
	protected.HandleFunc("POST /sleep-profiles", func(w http.ResponseWriter, r *http.Request) {
		createSleepProfileHandler(w, r, createSleepProfile)
	})
	protected.HandleFunc("POST /sleep-sessions/active", func(w http.ResponseWriter, r *http.Request) {
		startSleepHandler(w, r, sleepCtx, startSleep)
	})
	protected.HandleFunc("DELETE /sleep-sessions/active", func(w http.ResponseWriter, r *http.Request) {
		stopSleepHandler(w, r, sleepCtx, stopSleep)
	})
	protected.HandleFunc("PATCH /sleep-sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		editSleepSessionHandler(w, r, sleepCtx, editSleepSession)
	})
	protected.HandleFunc("DELETE /sleep-sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		deleteSleepSessionHandler(w, r, sleepCtx, deleteSleepSession)
	})
	protected.HandleFunc("GET /sleep-sessions", func(w http.ResponseWriter, r *http.Request) {
		getSleepHistoryHandler(w, r, sleepCtx, getSleepHistory)
	})
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
