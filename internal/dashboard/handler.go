package dashboard

import (
	"net/http"

	"finance-backend/internal/auth"
	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (auth.AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	DashboardService *Service
	AuthMiddleware   Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodGet, Path: "/v1/dashboard/summary", Summary: "Get dashboard summary", Protected: true},
		{Method: http.MethodGet, Path: "/v1/dashboard/daily-spending", Summary: "Get daily spending data", Protected: true},
		{Method: http.MethodGet, Path: "/v1/dashboard/monthly-spending", Summary: "Get monthly spending data", Protected: true},
		{Method: http.MethodGet, Path: "/v1/dashboard/comparison", Summary: "Get dashboard comparison", Protected: true},
		{Method: http.MethodGet, Path: "/v1/dashboard/expense-vs-salary", Summary: "Get expense percentage vs salary", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.DashboardService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/dashboard", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Get("/summary", h.summary)
		r.Get("/daily-spending", h.dailySpending)
		r.Get("/monthly-spending", h.monthlySpending)
		r.Get("/comparison", h.comparison)
		r.Get("/expense-vs-salary", h.expenseVsSalary)
	})
}

func (h handler) userID(r *http.Request) (int64, bool) {
	return parseDashboardUserID(r, h.authMiddleware)
}

func (h handler) summary(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	item, err := h.svc.Summary(r.Context(), userID)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) dailySpending(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.DailySpending(r.Context(), userID)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) monthlySpending(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.MonthlySpending(r.Context(), userID)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) comparison(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	item, err := h.svc.Comparison(r.Context(), userID)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) expenseVsSalary(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	item, err := h.svc.ExpenseVsSalary(r.Context(), userID)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func writeDashboardError(w http.ResponseWriter, err error) {
	_ = err
	writeError(w, http.StatusInternalServerError, "internal server error")
}
