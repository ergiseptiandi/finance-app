package server

import (
	"encoding/json"
	"net/http"

	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func registerDocsRoutes(router chi.Router, catalog routeCatalog) {
	router.Get("/routes", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string][]routeinfo.RouteInfo{"routes": catalog.List()})
	})

	router.Get("/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(buildOpenAPISpec(catalog.List()))
	})

	router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusTemporaryRedirect)
	})

	router.Handle("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/openapi.json"),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DefaultModelsExpandDepth(-1),
	))
}
