package server

import (
	"finance-backend/internal/alerts"
	"finance-backend/internal/auth"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerAlertsRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, alertsService *alerts.Service) {
	if authService == nil || alertsService == nil {
		return
	}

	alerts.RegisterRoutes(r, alerts.HandlerDependencies{
		AlertService:   alertsService,
		AuthMiddleware: authmiddleware.NewAuth(authService),
	})

	for _, route := range alerts.Definitions() {
		catalog.Add(route)
	}
}
