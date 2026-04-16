package server

import (
	"finance-backend/internal/auth"
	"finance-backend/internal/dashboard"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerDashboardRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, dashboardService *dashboard.Service) {
	if authService == nil || dashboardService == nil {
		return
	}

	dashboard.RegisterRoutes(r, dashboard.HandlerDependencies{
		DashboardService: dashboardService,
		AuthMiddleware:   authmiddleware.NewAuth(authService),
	})

	for _, route := range dashboard.Definitions() {
		catalog.Add(route)
	}
}
