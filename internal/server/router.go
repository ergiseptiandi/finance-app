package server

import (
	"net/http"

	"finance-backend/internal/server/routeinfo"

	"finance-backend/internal/auth"
	"finance-backend/internal/category"
	"finance-backend/internal/transaction"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type healthResponse struct {
	Status string `json:"status"`
}

func NewRouter(authService *auth.Service, txService *transaction.Service, categoryService *category.Service) http.Handler {
	router := chi.NewRouter()
	catalog := newRouteCatalog()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)

	registerBaseRoutes(router, catalog)

	router.Route("/v1", func(r chi.Router) {
		registerAuthRoutes(r, catalog, authService)
		registerTransactionRoutes(r, catalog, authService, txService)
		registerCategoryRoutes(r, catalog, authService, categoryService)
	})

	registerDocsRoutes(router, catalog)
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/openapi.json", Summary: "OpenAPI specification"})
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/docs", Summary: "Swagger UI"})

	return router
}
