package server

import (
	"finance-backend/internal/auth"
	authmiddleware "finance-backend/internal/server/middleware"
	"finance-backend/internal/transaction"

	"github.com/go-chi/chi/v5"
)

func registerTransactionRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, txService *transaction.Service) {
	if authService == nil || txService == nil {
		return
	}

	transaction.RegisterRoutes(r, transaction.HandlerDependencies{
		TransactionService: txService,
		AuthMiddleware:     authmiddleware.NewAuth(authService),
	})

	for _, route := range transaction.Definitions() {
		catalog.Add(route)
	}
}
