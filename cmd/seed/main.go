package main

import (
	"context"
	"log"
	"time"

	"finance-backend/internal/config"
	"finance-backend/internal/database"
	"finance-backend/internal/seed"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("no .env file loaded; using system environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if err := cfg.ValidateForSeed(); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.OpenMySQL(ctx, cfg.MySQL)
	if err != nil {
		log.Fatalf("mysql connection failed: %v", err)
	}
	defer db.Close()

	if err := seed.UpsertBootstrapUser(ctx, db, cfg.Seed); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Printf("seeded user %s", cfg.Seed.UserEmail)
}
