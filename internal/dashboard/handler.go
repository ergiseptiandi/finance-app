package dashboard

import (
	"errors"
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
		{Method: http.MethodGet, Path: "/v1/dashboard/budget-vs-actual", Summary: "Get budget vs actual data", Protected: true},
		{Method: http.MethodGet, Path: "/v1/dashboard/category-breakdown", Summary: "Get category breakdown", Protected: true},
		{Method: http.MethodGet, Path: "/v1/dashboard/upcoming-bills", Summary: "Get upcoming bills", Protected: true},
		{Method: http.MethodGet, Path: "/v1/dashboard/top-merchants", Summary: "Get top merchants", Protected: true},
		{Method: http.MethodGet, Path: "/v1/dashboard/insights", Summary: "Get dashboard insights", Protected: true},
		{Method: http.MethodGet, Path: "/v1/dashboard/goals-progress", Summary: "Get goals progress", Protected: true},
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
		r.Get("/budget-vs-actual", h.budgetVsActual)
		r.Get("/category-breakdown", h.categoryBreakdown)
		r.Get("/upcoming-bills", h.upcomingBills)
		r.Get("/top-merchants", h.topMerchants)
		r.Get("/insights", h.insights)
		r.Get("/goals-progress", h.goalsProgress)
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

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.Summary(r.Context(), userID, filter)
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

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.svc.DailySpending(r.Context(), userID, filter)
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

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.svc.MonthlySpending(r.Context(), userID, filter)
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

func (h handler) budgetVsActual(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	budgetAmount, err := parseOptionalFloatQuery(r, "budget_amount")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.BudgetVsActual(r.Context(), userID, filter, budgetAmount)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) categoryBreakdown(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.svc.CategoryBreakdown(r.Context(), userID, filter)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) upcomingBills(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	days, err := parsePositiveIntQuery(r, "days", 30)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.svc.UpcomingBills(r.Context(), userID, days)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) topMerchants(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.svc.TopMerchants(r.Context(), userID, filter)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) insights(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.svc.Insights(r.Context(), userID, filter)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) goalsProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.svc.GoalsProgress(r.Context(), userID, filter)
	if err != nil {
		writeDashboardError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func writeDashboardError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
