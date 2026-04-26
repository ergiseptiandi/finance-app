package budget

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	domainauth "finance-backend/internal/auth"
	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (domainauth.AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	BudgetService *Service
	AuthMiddleware Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodGet, Path: "/v1/budgets/category-goals", Summary: "List monthly category budget goals", Protected: true},
		{Method: http.MethodPost, Path: "/v1/budgets/category-goals", Summary: "Create monthly category budget goal", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/budgets/category-goals/{id}", Summary: "Update monthly category budget goal", Protected: true},
		{Method: http.MethodDelete, Path: "/v1/budgets/category-goals/{id}", Summary: "Delete monthly category budget goal", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	if deps.BudgetService == nil || deps.AuthMiddleware == nil {
		return
	}

	h := handler{svc: deps.BudgetService, authMiddleware: deps.AuthMiddleware}

	r.Route("/budgets", func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequireAuth)
		r.Get("/category-goals", h.list)
		r.Post("/category-goals", h.create)
		r.Patch("/category-goals/{id}", h.update)
		r.Delete("/category-goals/{id}", h.delete)
	})
}

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	start, end := currentMonthRange()
	if month := r.URL.Query().Get("month"); month != "" {
		parsed, err := time.Parse("2006-01", month)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid month")
			return
		}
		start = time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, parsed.Location())
		end = start.AddDate(0, 1, 0)
	}

	items, summary, err := h.svc.List(r.Context(), userID, start, end)
	if err != nil {
		writeBudgetError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", map[string]any{
		"summary": summary,
		"items":   items,
	})
}

func (h handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input, err := decodeCreateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.Create(r.Context(), userID, input)
	if err != nil {
		writeBudgetError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "Success Create", item)
}

func (h handler) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget goal id")
		return
	}

	input, err := decodeUpdateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), userID, id, input)
	if err != nil {
		writeBudgetError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Update", item)
}

func (h handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget goal id")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		writeBudgetError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Delete", map[string]string{"status": "deleted"})
}

func (h handler) userID(r *http.Request) (int64, bool) {
	claims, ok := h.authMiddleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func writeBudgetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		if err != nil && (errors.Is(err, ErrInvalidInput) || containsBudgetValidationError(err.Error())) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func containsBudgetValidationError(message string) bool {
	return message != "" && (strings.Contains(message, "required") || strings.Contains(message, "greater than zero") || strings.Contains(message, "positive number") || strings.Contains(message, "expense categories"))
}

func currentMonthRange() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return start, start.AddDate(0, 1, 0)
}
