package httpapi

import (
	"net/http"

	authroutes "finance-backend/internal/httpapi/auth"
	authmiddleware "finance-backend/internal/httpapi/middleware"
	"finance-backend/internal/httpapi/routeinfo"

	"finance-backend/internal/auth"
	"finance-backend/internal/transaction"
	txroutes "finance-backend/internal/httpapi/transaction"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type healthResponse struct {
	Status string `json:"status"`
}

func NewRouter(authService *auth.Service, txService *transaction.Service) http.Handler {
	router := chi.NewRouter()
	catalog := newRouteCatalog()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)

	router.Get("/", rootHandler)
	router.Get("/health", healthHandler)
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/", Summary: "Root service status"})
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/health", Summary: "Health check"})
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/routes", Summary: "List registered API routes"})

	router.Route("/v1", func(r chi.Router) {
		if authService != nil {
			authroutes.RegisterRoutes(r, authroutes.HandlerDependencies{
				AuthService:    authService,
				AuthMiddleware: authmiddleware.NewAuth(authService),
			})
			for _, route := range authroutes.Definitions() {
				catalog.Add(route)
			}
		}

		if txService != nil && authService != nil {
			txroutes.RegisterRoutes(r, txroutes.HandlerDependencies{
				TransactionService: txService,
				AuthMiddleware:     authmiddleware.NewAuth(authService),
			})
			for _, route := range txroutes.Definitions() {
				catalog.Add(route)
			}
		}
	})

	registerDocsRoutes(router, catalog)
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/openapi.json", Summary: "OpenAPI specification"})
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/docs", Summary: "Swagger UI"})

	return router
}

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "finance-backend running"})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
