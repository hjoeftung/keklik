// Package main is the entry point for the Keklik HTTP API server.
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
//
//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go --dir ../.. --output ../../docs --outputTypes yaml
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
	"github.com/hjoeftung/keklik/internal/infrastructure"
	httpapi "github.com/hjoeftung/keklik/internal/interfaces/http"
	"github.com/hjoeftung/keklik/internal/sleep"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	infrastructure.SetupLogger()

	config, err := infrastructure.LoadConfig(getenv)
	if err != nil {
		slog.Error("configuration error", "err", err)
		return err
	}

	db, err := infrastructure.OpenDB(config.Database.URL)
	if err != nil {
		slog.Error("database connection error", "err", err)
		return err
	}
	defer db.Close()

	if err := infrastructure.RunMigrations(db); err != nil {
		slog.Error("migration error", "err", err)
		return err
	}

	refreshTokenRepo := infrastructure.NewPostgresRefreshTokenRepository(db)
	familyRepo := infrastructure.NewPostgresFamilyRepository(db)
	familyMemberRepo := infrastructure.NewPostgresFamilyMemberRepository(db)
	accountRepo := infrastructure.NewPostgresAccountRepository(db)
	nightWindowRepo := infrastructure.NewPostgresNightWindowRepository(db)
	sleepSessionRepo := infrastructure.NewPostgresSleepSessionRepository(db)
	babyAccessChecker := infrastructure.NewPostgresBabyAccessChecker(db)

	createFamily := family.NewCreateFamilyHandler(familyRepo)
	getFamily := family.NewGetFamilyHandler(familyRepo)
	createInviteLink := family.NewCreateFamilyInviteLinkHandler(
		familyRepo,
		familyMemberRepo,
		config.App.BaseURL,
		config.App.InviteLinkLifetime,
	)
	joinFamilyByInvite := family.NewJoinFamilyByInviteLinkHandler(familyRepo, familyMemberRepo)
	revokeInviteLink := family.NewRevokeInviteLinkHandler(familyRepo, familyMemberRepo)
	transactor := infrastructure.NewPostgresTransactor(db)
	setNightWindow := sleep.NewSetNightWindowHandler(nightWindowRepo, transactor)
	startSleep := sleep.NewStartSleepHandler(sleepSessionRepo)
	stopSleep := sleep.NewStopSleepHandler(sleepSessionRepo)
	editSleepSession := sleep.NewEditSleepSessionHandler(sleepSessionRepo)
	deleteSleepSession := sleep.NewDeleteSleepSessionHandler(sleepSessionRepo)
	getSleepHistory := sleep.NewGetSleepHistoryHandler(sleepSessionRepo, nightWindowRepo)
	jwtValidator := auth.NewJWTValidator(config.Auth.JWTSigningKey)
	oauthCallback := auth.NewHandleOAuthCallbackHandler(accountRepo, refreshTokenRepo, config.Auth.JWTSigningKey, config.Auth.AccessTokenDuration, config.Auth.RefreshTokenDuration)
	testLogin := auth.NewHandleTestLoginHandler(accountRepo, refreshTokenRepo, config.Auth.JWTSigningKey, config.Auth.AccessTokenDuration, config.Auth.RefreshTokenDuration)
	refreshTokenHandler := auth.NewHandleRefreshTokenHandler(refreshTokenRepo, config.Auth.JWTSigningKey, config.Auth.AccessTokenDuration, config.Auth.RefreshTokenDuration)
	logoutHandler := auth.NewHandleLogoutHandler(refreshTokenRepo)

	server := httpapi.NewServer(config, httpapi.Dependencies{
		Accounts:           accountRepo,
		Validator:          jwtValidator,
		OAuthCallback:      oauthCallback,
		TestLogin:          testLogin,
		RefreshToken:       refreshTokenHandler,
		Logout:             logoutHandler,
		CreateFamily:       createFamily,
		GetFamily:          getFamily,
		CreateInviteLink:   createInviteLink,
		RevokeInviteLink:   revokeInviteLink,
		JoinFamilyByInvite: joinFamilyByInvite,
		BabyAccess:         babyAccessChecker,
		SetNightWindow:     setNightWindow,
		StartSleep:         startSleep,
		StopSleep:          stopSleep,
		EditSleepSession:   editSleepSession,
		DeleteSleepSession: deleteSleepSession,
		GetSleepHistory:    getSleepHistory,
	})

	slog.Info("starting HTTP server", "addr", config.Address())

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown failed", "err", err)
			return err
		}
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
			return err
		}
	}
	return err
}
