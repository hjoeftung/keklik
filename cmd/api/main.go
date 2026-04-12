package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
	"github.com/hjoeftung/keklik/internal/infrastructure"
	httpapi "github.com/hjoeftung/keklik/internal/interfaces/http"
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
	accountRepo := infrastructure.NewPostgresAccountRepository(db)
	sessionRepo := infrastructure.NewPostgresSessionRepository(db)

	createFamily := family.NewCreateFamilyHandler(familyRepo)
	oauthCallback := auth.NewHandleOAuthCallbackHandler(accountRepo, sessionRepo)

	server := httpapi.NewServer(config, accountRepo, sessionRepo, oauthCallback, createFamily)

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
