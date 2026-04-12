package main

import (
	"log"
	"os"

	"github.com/hjoeftung/keklik/internal/infrastructure"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := infrastructure.OpenDB(databaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer db.Close()

	if err := infrastructure.RunMigrations(db); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	log.Println("migrations applied successfully")
}
