// Package main is the entry point for the Keklik HTTP API server.
//
//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go --dir ../.. --output ../../docs --outputTypes yaml
//
// @title       Keklik API
// @version     1.0
// @description Baby sleep tracking API for families.
//
// @host        localhost:8080
// @BasePath    /
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Session token obtained from the auth endpoints.
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
	"github.com/hjoeftung/keklik/internal/infrastructure"
	httpapi "github.com/hjoeftung/keklik/internal/interfaces/http"
	"github.com/hjoeftung/keklik/internal/sleep"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := infrastructure.LoadConfig()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	db, err := infrastructure.OpenDB(config.Database.URL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer db.Close()

	if err := infrastructure.RunMigrations(db); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	familyRepo := infrastructure.NewPostgresFamilyRepository(db)
	familyMemberRepo := infrastructure.NewPostgresFamilyMemberRepository(db)
	accountRepo := infrastructure.NewPostgresAccountRepository(db)
	sleepProfileRepo := infrastructure.NewPostgresSleepProfileRepository(db)
	sleepSessionRepo := infrastructure.NewPostgresSleepSessionRepository(db)
	babyAccessChecker := infrastructure.NewPostgresBabyAccessChecker(db)

	createFamily := family.NewCreateFamilyHandler(familyRepo)
	getFamily := family.NewGetFamilyHandler(familyRepo, familyMemberRepo)
	createInviteLink := family.NewCreateFamilyInviteLinkHandler(
		familyRepo,
		familyMemberRepo,
		config.App.BaseURL,
		config.App.InviteLinkLifetime,
	)
	joinFamilyByInvite := family.NewJoinFamilyByInviteLinkHandler(familyRepo, familyMemberRepo)
	transactor := infrastructure.NewPostgresTransactor(db)
	createSleepProfile := sleep.NewCreateSleepProfileHandler(sleepProfileRepo, sleepSessionRepo, sleepSessionRepo, transactor)
	startSleep := sleep.NewStartSleepHandler(sleepSessionRepo)
	stopSleep := sleep.NewStopSleepHandler(sleepSessionRepo, sleepProfileRepo)
	editSleepSession := sleep.NewEditSleepSessionHandler(sleepSessionRepo, sleepProfileRepo)
	deleteSleepSession := sleep.NewDeleteSleepSessionHandler(sleepSessionRepo)
	getSleepHistory := sleep.NewGetSleepHistoryHandler(sleepSessionRepo, sleepProfileRepo)
	getDashboardSummary := sleep.NewGetDashboardSummaryHandler(sleepProfileRepo, sleepSessionRepo)
	const tokenDuration = 30 * 24 * time.Hour
	jwtValidator := auth.NewJWTValidator(config.Auth.JWTSigningKey)
	oauthCallback := auth.NewHandleOAuthCallbackHandler(accountRepo, config.Auth.JWTSigningKey, tokenDuration)
	testLogin := auth.NewHandleTestLoginHandler(accountRepo, config.Auth.JWTSigningKey, tokenDuration)

	server := httpapi.NewServer(config, httpapi.Dependencies{
		Accounts:           accountRepo,
		Validator:          jwtValidator,
		OAuthCallback:      oauthCallback,
		TestLogin:          testLogin,
		CreateFamily:       createFamily,
		GetFamily:          getFamily,
		CreateInviteLink:   createInviteLink,
		JoinFamilyByInvite: joinFamilyByInvite,
		BabyAccess:         babyAccessChecker,
		CreateSleepProfile: createSleepProfile,
		StartSleep:         startSleep,
		StopSleep:          stopSleep,
		EditSleepSession:   editSleepSession,
		DeleteSleepSession: deleteSleepSession,
		GetSleepHistory:     getSleepHistory,
		GetDashboardSummary: getDashboardSummary,
	})

	log.Printf("starting HTTP server on %s", config.Address())

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown failed: %v", err)
		}
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}
}
