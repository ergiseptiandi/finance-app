package reports

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
	ReportsService *Service
	AuthMiddleware Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodGet, Path: "/v1/reports/expense-by-category", Summary: "Expense by category", Protected: true},
		{Method: http.MethodGet, Path: "/v1/reports/spending-trends", Summary: "Spending trends", Protected: true},
		{Method: http.MethodGet, Path: "/v1/reports/highest-spending-category", Summary: "Highest spending category", Protected: true},
		{Method: http.MethodGet, Path: "/v1/reports/average-daily-spending", Summary: "Average daily spending", Protected: true},
		{Method: http.MethodGet, Path: "/v1/reports/remaining-balance", Summary: "Remaining balance", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.ReportsService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/reports", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Get("/expense-by-category", h.expenseByCategory)
		r.Get("/spending-trends", h.spendingTrends)
		r.Get("/highest-spending-category", h.highestSpendingCategory)
		r.Get("/average-daily-spending", h.averageDailySpending)
		r.Get("/remaining-balance", h.remainingBalance)
	})
}

func (h handler) userID(r *http.Request) (int64, bool) {
	return parseReportsUserID(r, h.authMiddleware)
}

func (h handler) expenseByCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseReportsFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.ExpenseByCategory(r.Context(), userID, filter)
	if err != nil {
		writeReportsError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) spendingTrends(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseReportsFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.SpendingTrends(r.Context(), userID, filter)
	if err != nil {
		writeReportsError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) highestSpendingCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseReportsFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.HighestSpendingCategory(r.Context(), userID, filter)
	if err != nil {
		writeReportsError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) averageDailySpending(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseReportsFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.AverageDailySpending(r.Context(), userID, filter)
	if err != nil {
		writeReportsError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) remainingBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseReportsFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.RemainingBalance(r.Context(), userID, filter)
	if err != nil {
		writeReportsError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func writeReportsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
