package httpapi

import (
	"net/http"

	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
	"github.com/hjoeftung/keklik/internal/infrastructure"
	"github.com/hjoeftung/keklik/internal/sleep"
)

// Dependencies holds all handler and repository dependencies for the HTTP server.
type Dependencies struct {
	Accounts           auth.AccountRepository
	Validator          auth.TokenValidator
	OAuthCallback      *auth.HandleOAuthCallbackHandler
	TestLogin          *auth.HandleTestLoginHandler
	RefreshToken       *auth.HandleRefreshTokenHandler
	Logout             *auth.HandleLogoutHandler
	CreateFamily       *family.CreateFamilyHandler
	GetFamily          *family.GetFamilyHandler
	CreateInviteLink   *family.CreateFamilyInviteLinkHandler
	RevokeInviteLink   *family.RevokeInviteLinkHandler
	JoinFamilyByInvite *family.JoinFamilyByInviteLinkHandler
	BabyAccess         babyAccessChecker
	SetNightWindow     *sleep.SetNightWindowHandler
	LogPastSleep       *sleep.LogPastSleepHandler
	StartSleep         *sleep.StartSleepHandler
	StopSleep          *sleep.StopSleepHandler
	EditSleepSession   *sleep.EditSleepSessionHandler
	DeleteSleepSession *sleep.DeleteSleepSessionHandler
	GetSleepHistory    *sleep.GetSleepHistoryHandler
}

// NewServer wires the HTTP transport and returns a ready-to-start server.
func NewServer(config infrastructure.Config, deps Dependencies) *http.Server {
	mux := http.NewServeMux()
	addRoutes(mux, config, deps)
	return &http.Server{
		Addr:    config.Address(),
		Handler: withCORS(config.App.FrontendURL, withRequestID(mux)),
	}
}
