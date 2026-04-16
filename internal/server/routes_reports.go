package server

import (
	"finance-backend/internal/auth"
	"finance-backend/internal/reports"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerReportsRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, reportsService *reports.Service) {
	if authService == nil || reportsService == nil {
		return
	}

	reports.RegisterRoutes(r, reports.HandlerDependencies{
		ReportsService: reportsService,
		AuthMiddleware: authmiddleware.NewAuth(authService),
	})

	for _, route := range reports.Definitions() {
		catalog.Add(route)
	}
}
