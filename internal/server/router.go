package server

import (
	"net/http"

	"finance-backend/internal/server/routeinfo"

	"finance-backend/internal/ai"
	"finance-backend/internal/alerts"
	"finance-backend/internal/auth"
	"finance-backend/internal/budget"
	"finance-backend/internal/category"
	"finance-backend/internal/dashboard"
	"finance-backend/internal/debt"
	"finance-backend/internal/exportcsv"
	"finance-backend/internal/media"
	"finance-backend/internal/notifications"
	"finance-backend/internal/reports"
	"finance-backend/internal/transaction"
	"finance-backend/internal/wallet"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type healthResponse struct {
	Status string `json:"status"`
}

func NewRouter(authService *auth.Service, txService *transaction.Service, walletService *wallet.Service, categoryService *category.Service, debtService *debt.Service, dashboardService *dashboard.Service, reportsService *reports.Service, alertsService *alerts.Service, notificationsService *notifications.Service, mediaService *media.Service, debtStorage debt.FileStorage, uploadDir string, budgetService *budget.Service, exportService *exportcsv.Service, aiService *ai.Service) http.Handler {
	router := chi.NewRouter()
	catalog := newRouteCatalog()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)

	registerBaseRoutes(router, catalog)

		router.Route("/v1", func(r chi.Router) {
			registerAuthRoutes(r, catalog, authService)
			registerTransactionRoutes(r, catalog, authService, txService)
			registerWalletRoutes(r, catalog, authService, walletService)
			registerCategoryRoutes(r, catalog, authService, categoryService)
		registerDebtRoutes(r, catalog, authService, debtService, debtStorage)
		registerExportCSVRoutes(r, catalog, authService, exportService)
		registerBudgetRoutes(r, catalog, authService, budgetService)
			registerDashboardRoutes(r, catalog, authService, dashboardService)
		registerReportsRoutes(r, catalog, authService, reportsService)
		registerAlertsRoutes(r, catalog, authService, alertsService)
		registerNotificationsRoutes(r, catalog, authService, notificationsService)
		registerMediaRoutes(r, catalog, authService, mediaService)
		registerAIRoutes(r, catalog, authService, aiService)
	})

	if uploadDir != "" {
		router.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))
	}

	registerDocsRoutes(router, catalog)
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/openapi.json", Summary: "OpenAPI specification"})
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/docs", Summary: "Swagger UI"})

	return router
}
