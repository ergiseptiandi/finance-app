package server

import (
	"finance-backend/internal/auth"
	"finance-backend/internal/debt"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerDebtRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, debtService *debt.Service, storage debt.FileStorage) {
	if authService == nil || debtService == nil || storage == nil {
		return
	}

	debt.RegisterRoutes(r, debt.HandlerDependencies{
		DebtService:    debtService,
		AuthMiddleware: authmiddleware.NewAuth(authService),
		Storage:        storage,
	})

	for _, route := range debt.Definitions() {
		catalog.Add(route)
	}
}
