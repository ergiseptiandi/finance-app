package server

import (
	"finance-backend/internal/auth"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerAuthRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service) {
	if authService == nil {
		return
	}

	auth.RegisterRoutes(r, auth.HandlerDependencies{
		AuthService:    authService,
		AuthMiddleware: authmiddleware.NewAuth(authService),
	})

	for _, route := range auth.Definitions() {
		catalog.Add(route)
	}
}
