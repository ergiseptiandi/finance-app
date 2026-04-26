package server

import (
	"finance-backend/internal/auth"
	"finance-backend/internal/budget"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerBudgetRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, budgetService *budget.Service) {
	if authService == nil || budgetService == nil {
		return
	}

	budget.RegisterRoutes(r, budget.HandlerDependencies{
		BudgetService:  budgetService,
		AuthMiddleware: authmiddleware.NewAuth(authService),
	})

	for _, route := range budget.Definitions() {
		catalog.Add(route)
	}
}
