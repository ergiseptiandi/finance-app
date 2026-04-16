package server

import (
	"net/http"

	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

func registerBaseRoutes(router chi.Router, catalog routeCatalog) {
	router.Get("/", rootHandler)
	router.Get("/health", healthHandler)

	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/", Summary: "Root service status"})
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/health", Summary: "Health check"})
	catalog.Add(routeinfo.RouteInfo{Method: http.MethodGet, Path: "/routes", Summary: "List registered API routes"})
}

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "finance-backend running"})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
