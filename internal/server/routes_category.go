package server

import (
	"finance-backend/internal/auth"
	"finance-backend/internal/category"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerCategoryRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, categoryService *category.Service) {
	if authService == nil || categoryService == nil {
		return
	}

	category.RegisterRoutes(r, category.HandlerDependencies{
		CategoryService: categoryService,
		AuthMiddleware:  authmiddleware.NewAuth(authService),
	})

	for _, route := range category.Definitions() {
		catalog.Add(route)
	}
}
