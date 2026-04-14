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

// NewServer wires the HTTP transport and returns a ready-to-start server.
func NewServer(
	config infrastructure.Config,
	accounts auth.AccountRepository,
	sessions auth.SessionRepository,
	oauthCallback *auth.HandleOAuthCallbackHandler,
	createFamily *family.CreateFamilyHandler,
	sleepCtx sleepContextResolver,
	createSleepProfile *sleep.CreateSleepProfileHandler,
	startSleep *sleep.StartSleepHandler,
) *http.Server {
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

	// Protected endpoints — wrapped with requireAuth middleware.
	protected := http.NewServeMux()
	protected.HandleFunc("POST /families", func(w http.ResponseWriter, r *http.Request) {
		createFamilyHandler(w, r, createFamily)
	})
	protected.HandleFunc("POST /sleep-profiles", func(w http.ResponseWriter, r *http.Request) {
		createSleepProfileHandler(w, r, createSleepProfile)
	})
	protected.HandleFunc("POST /sleep-sessions", func(w http.ResponseWriter, r *http.Request) {
		startSleepHandler(w, r, sleepCtx, startSleep)
	})
	mux.Handle("/", requireAuth(accounts, sessions, protected))

	return &http.Server{
		Addr:    config.Address(),
		Handler: mux,
	}
}

func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(writer).Encode(healthResponse{Status: "ok"})
}
