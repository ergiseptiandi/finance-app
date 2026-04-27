package exportcsv

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"finance-backend/internal/auth"
	"finance-backend/internal/debt"
	"finance-backend/internal/reports"
	"finance-backend/internal/server/routeinfo"
	"finance-backend/internal/transaction"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (auth.AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	ExportService  *Service
	AuthMiddleware Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodGet, Path: "/v1/exports/csv", Summary: "Export CSV data", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.ExportService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/exports", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Get("/csv", h.exportCSV)
	})
}

func (h handler) exportCSV(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	scope, err := parseScope(r.URL.Query().Get("scope"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	period, err := parsePeriod(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.svc.Export(r.Context(), userID, scope, period)
	if err != nil {
		if errors.Is(err, transaction.ErrInvalidInput) || errors.Is(err, reports.ErrInvalidInput) || errors.Is(err, debt.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+result.FileName+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Export-Partial", strconv.FormatBool(result.Partial))
	w.Header().Set("X-Export-Record-Count", strconv.Itoa(result.RecordCount))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.CSV)
}

func (h handler) userID(r *http.Request) (int64, bool) {
	claims, ok := h.authMiddleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}

func parseScope(value string) (Scope, error) {
	scope := Scope(strings.TrimSpace(value))

	switch scope {
	case ScopeTransactions, ScopeDebts, ScopeReports:
		return scope, nil
	default:
		return "", errors.New("scope must be transactions, debts, or reports")
	}
}

func parsePeriod(r *http.Request) (Period, error) {
	query := r.URL.Query()
	month := strings.TrimSpace(query.Get("month"))
	startDate := strings.TrimSpace(query.Get("start_date"))
	endDate := strings.TrimSpace(query.Get("end_date"))

	if month != "" {
		if startDate != "" || endDate != "" {
			return Period{}, errors.New("month cannot be combined with start_date or end_date")
		}

		parsedMonth, err := time.ParseInLocation("2006-01", month, time.Local)
		if err != nil {
			return Period{}, errors.New("month must use format YYYY-MM")
		}

		startOfMonth := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, time.Local)
		endOfMonth := startOfMonth.AddDate(0, 1, 0).AddDate(0, 0, -1)
		return Period{
			Month:     startOfMonth.Format("2006-01"),
			StartDate:  &startOfMonth,
			EndDate:    &endOfMonth,
			Label:      startOfMonth.Format("2006-01"),
			HasFilters: true,
		}, nil
	}

	if startDate == "" && endDate == "" {
		return Period{}, nil
	}

	if startDate == "" || endDate == "" {
		return Period{}, errors.New("start_date and end_date must be provided together")
	}

	parsedStartDate, err := time.ParseInLocation("2006-01-02", startDate, time.Local)
	if err != nil {
		return Period{}, errors.New("start_date must use format YYYY-MM-DD")
	}

	parsedEndDate, err := time.ParseInLocation("2006-01-02", endDate, time.Local)
	if err != nil {
		return Period{}, errors.New("end_date must use format YYYY-MM-DD")
	}

	parsedStartDate = time.Date(parsedStartDate.Year(), parsedStartDate.Month(), parsedStartDate.Day(), 0, 0, 0, 0, time.Local)
	parsedEndDate = time.Date(parsedEndDate.Year(), parsedEndDate.Month(), parsedEndDate.Day(), 0, 0, 0, 0, time.Local)

	if parsedEndDate.Before(parsedStartDate) {
		return Period{}, errors.New("end_date must be greater than or equal to start_date")
	}

	if parsedEndDate.After(parsedStartDate.AddDate(1, 0, 0)) {
		return Period{}, errors.New("date range cannot exceed 1 year")
	}

	return Period{
		StartDate:  &parsedStartDate,
		EndDate:    &parsedEndDate,
		Label:      parsedStartDate.Format("2006-01-02") + "_to_" + parsedEndDate.Format("2006-01-02"),
		HasFilters: true,
	}, nil
}
