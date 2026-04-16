package server

import (
	"finance-backend/internal/auth"
	"finance-backend/internal/salary"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerSalaryRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, salaryService *salary.Service) {
	if authService == nil || salaryService == nil {
		return
	}

	salary.RegisterRoutes(r, salary.HandlerDependencies{
		SalaryService:  salaryService,
		AuthMiddleware: authmiddleware.NewAuth(authService),
	})

	for _, route := range salary.Definitions() {
		catalog.Add(route)
	}
}
