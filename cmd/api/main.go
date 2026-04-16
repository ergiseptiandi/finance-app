package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"finance-backend/internal/alerts"
	"finance-backend/internal/auth"
	"finance-backend/internal/category"
	"finance-backend/internal/config"
	"finance-backend/internal/dashboard"
	"finance-backend/internal/database"
	"finance-backend/internal/debt"
	"finance-backend/internal/mail"
	"finance-backend/internal/reports"
	"finance-backend/internal/salary"
	"finance-backend/internal/server"
	"finance-backend/internal/storage"
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

	categoryRepo := category.NewMySQLCategoryRepository(db)
	categoryService := category.NewService(categoryRepo)

	salaryRepo := salary.NewMySQLSalaryRepository(db)
	salaryService := salary.NewService(salaryRepo)

	debtRepo := debt.NewMySQLDebtRepository(db)
	debtService := debt.NewService(debtRepo)

	dashboardRepo := dashboard.NewMySQLDashboardRepository(db)
	dashboardService := dashboard.NewService(dashboardRepo)

	alertsRepo := alerts.NewMySQLAlertsRepository(db)
	alertsService := alerts.NewService(alertsRepo)

	reportsRepo := reports.NewMySQLReportsRepository(db)
	reportsService := reports.NewService(reportsRepo)

	fileStorage := storage.NewLocalStorage(cfg.Storage.UploadDir)

	txRepo := transaction.NewMySQLTransactionRepository(db)
	txService := transaction.NewService(txRepo)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: server.NewRouter(authService, txService, categoryService, salaryService, debtService, dashboardService, reportsService, alertsService, fileStorage, cfg.Storage.UploadDir),
	}

	log.Printf("server listening on :%s", cfg.Server.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
