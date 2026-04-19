package server

import (
	"finance-backend/internal/auth"
	authmiddleware "finance-backend/internal/server/middleware"
	"finance-backend/internal/wallet"

	"github.com/go-chi/chi/v5"
)

func registerWalletRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, walletService *wallet.Service) {
	if authService == nil || walletService == nil {
		return
	}

	wallet.RegisterRoutes(r, wallet.HandlerDependencies{
		WalletService:  walletService,
		AuthMiddleware: authmiddleware.NewAuth(authService),
	})

	for _, route := range wallet.Definitions() {
		catalog.Add(route)
	}
}
