package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"finance-backend/internal/auth"
	"finance-backend/internal/config"
	"finance-backend/internal/database"
	"finance-backend/internal/mail"
	"finance-backend/internal/server"

	"finance-backend/internal/transaction"

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

	if err := cfg.ValidateForAPI(); err != nil {
		log.Fatal(err)
	}

	startupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.OpenMySQL(startupCtx, cfg.MySQL)
	if err != nil {
		log.Fatalf("mysql connection failed: %v", err)
	}
	defer db.Close()

	log.Printf("mysql connected to %s:%s/%s", cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Name)

	authRepo := auth.NewMySQLAuthRepository(db)
	passwordResetRepo := auth.NewMySQLPasswordResetRepository(db)
	tokenManager := auth.NewTokenManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.Issuer,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.RefreshTokenTTL,
	)

	var mailer mail.Sender
	if cfg.SMTP.Host != "" && cfg.SMTP.Port > 0 {
		mailer = mail.NewSMTPSender(
			cfg.SMTP.Host,
			cfg.SMTP.Port,
			cfg.SMTP.Username,
			cfg.SMTP.Password,
			cfg.SMTP.From,
		)
	}

	authService := auth.NewService(authRepo, authRepo, passwordResetRepo, tokenManager, mailer)

	txRepo := transaction.NewMySQLTransactionRepository(db)
	txService := transaction.NewService(txRepo)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: server.NewRouter(authService, txService),
	}

	log.Printf("server listening on :%s", cfg.Server.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
