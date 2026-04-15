package httpapi

import (
	"net/http"

	authroutes "finance-backend/internal/httpapi/auth"
	authmiddleware "finance-backend/internal/httpapi/middleware"

	"finance-backend/internal/auth"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type healthResponse struct {
	Status string `json:"status"`
}

func NewRouter(authService *auth.Service) http.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)

	router.Get("/", rootHandler)
	router.Get("/health", healthHandler)

	router.Route("/v1", func(r chi.Router) {
		if authService != nil {
			authroutes.RegisterRoutes(r, authroutes.HandlerDependencies{
				AuthService:    authService,
				AuthMiddleware: authmiddleware.NewAuth(authService),
			})
		}
	})

	return router
}

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "finance-backend running"})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
