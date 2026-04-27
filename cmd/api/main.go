package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"finance-backend/internal/alerts"
	"finance-backend/internal/auth"
	"finance-backend/internal/budget"
	"finance-backend/internal/category"
	"finance-backend/internal/config"
	"finance-backend/internal/dashboard"
	"finance-backend/internal/database"
	"finance-backend/internal/debt"
	"finance-backend/internal/exportcsv"
	"finance-backend/internal/mail"
	"finance-backend/internal/media"
	"finance-backend/internal/notifications"
	"finance-backend/internal/reports"
	"finance-backend/internal/server"
	"finance-backend/internal/storage"
	"finance-backend/internal/transaction"
	"finance-backend/internal/wallet"

	"github.com/joho/godotenv"
)

type dashboardExpenseProvider struct {
	repo dashboard.Repository
}

func (p dashboardExpenseProvider) ExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]budget.ExpenseItem, error) {
	items, err := p.repo.ExpenseByCategory(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	result := make([]budget.ExpenseItem, 0, len(items))
	for _, item := range items {
		result = append(result, budget.ExpenseItem{Category: item.Category, Amount: item.Amount})
	}

	return result, nil
}

func main() {
	_ = godotenv.Load()

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

	walletRepo := wallet.NewMySQLWalletRepository(db)
	walletService := wallet.NewService(walletRepo)

	categoryRepo := category.NewMySQLCategoryRepository(db)
	categoryService := category.NewService(categoryRepo)

	debtRepo := debt.NewMySQLDebtRepository(db)
	debtService := debt.NewService(debtRepo, walletService)

	notificationsRepo := notifications.NewMySQLNotificationsRepository(db)
	var pushSender notifications.PushSender
	if sender, err := notifications.NewFirebaseSender(startupCtx, notifications.FirebasePushConfig{
		ProjectID:       cfg.Push.FirebaseProjectID,
		CredentialsJSON: cfg.Push.FirebaseCredentialsJSON,
		CredentialsPath: cfg.Push.FirebaseCredentialsPath,
	}); err != nil {
		log.Printf("firebase push disabled: %v", err)
	} else {
		pushSender = sender
	}
	notificationsService := notifications.NewService(notificationsRepo, pushSender)

	alertsRepo := alerts.NewMySQLAlertsRepository(db)
	alertsService := alerts.NewService(alertsRepo)

	dashboardRepo := dashboard.NewMySQLDashboardRepository(db)
	budgetRepo := budget.NewMySQLRepository(db)
	budgetService := budget.NewService(budgetRepo, dashboardExpenseProvider{repo: dashboardRepo})
	dashboardService := dashboard.NewService(dashboardRepo, walletService, alertsService, notificationsService, budgetService)

	reportsRepo := reports.NewMySQLReportsRepository(db)
	reportsService := reports.NewService(reportsRepo, walletService)

	fileStorage := storage.NewLocalStorage(cfg.Storage.UploadDir)
	mediaService := media.NewService(fileStorage)

	txRepo := transaction.NewMySQLTransactionRepository(db)
	txService := transaction.NewService(txRepo, walletService)

	exportService := exportcsv.NewService(txService, debtService, reportsService)

	if cfg.Runtime.Mode == "worker" || os.Getenv("APP_MODE") == "worker" {
		worker := notifications.NewWorker(notificationsService, alertsService, notificationsRepo, cfg.Runtime.NotificationCronSpec)
		workerCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		if err := worker.Run(workerCtx); err != nil && err != context.Canceled {
			log.Fatal(err)
		}
		return
	}

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: server.NewRouter(authService, txService, walletService, categoryService, debtService, dashboardService, reportsService, alertsService, notificationsService, mediaService, fileStorage, cfg.Storage.UploadDir, budgetService, exportService),
	}

	log.Printf("server listening on :%s", cfg.Server.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
