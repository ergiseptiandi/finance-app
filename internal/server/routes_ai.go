package server

import (
	"finance-backend/internal/ai"
	"finance-backend/internal/auth"
	"finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerAIRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, aiService *ai.Service) {
	if aiService == nil {
		return
	}

	deps := ai.HandlerDependencies{
		AIService:      aiService,
		AuthMiddleware: middleware.NewAuth(authService),
		AuthService:    authService,
	}

	for _, def := range ai.Definitions() {
		catalog.Add(def)
	}

	ai.RegisterRoutes(r, deps)
}
