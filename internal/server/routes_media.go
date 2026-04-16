package server

import (
	"finance-backend/internal/auth"
	"finance-backend/internal/media"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerMediaRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, mediaService *media.Service) {
	if authService == nil || mediaService == nil {
		return
	}

	media.RegisterRoutes(r, media.HandlerDependencies{
		MediaService:   mediaService,
		AuthMiddleware: authmiddleware.NewAuth(authService),
	})

	for _, route := range media.Definitions() {
		catalog.Add(route)
	}
}
