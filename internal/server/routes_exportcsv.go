package server

import (
	"finance-backend/internal/auth"
	"finance-backend/internal/exportcsv"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerExportCSVRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, exportService *exportcsv.Service) {
	if authService == nil || exportService == nil {
		return
	}

	exportcsv.RegisterRoutes(r, exportcsv.HandlerDependencies{
		ExportService:  exportService,
		AuthMiddleware: authmiddleware.NewAuth(authService),
	})

	for _, route := range exportcsv.Definitions() {
		catalog.Add(route)
	}
}
