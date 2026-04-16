package salary

import (
	"errors"
	"net/http"
	"strings"

	domainauth "finance-backend/internal/auth"
	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (domainauth.AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	SalaryService  *Service
	AuthMiddleware Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodPost, Path: "/v1/salaries", Summary: "Create salary record", Protected: true},
		{Method: http.MethodGet, Path: "/v1/salaries", Summary: "Get salary history", Protected: true},
		{Method: http.MethodGet, Path: "/v1/salaries/current", Summary: "Get current salary", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/salaries/{id}", Summary: "Update salary", Protected: true},
		{Method: http.MethodDelete, Path: "/v1/salaries/{id}", Summary: "Delete salary", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/salaries/schedule", Summary: "Set salary date", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.SalaryService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/salaries", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Post("/", h.create)
		r.Get("/", h.history)
		r.Get("/current", h.current)
		r.Patch("/{id}", h.update)
		r.Delete("/{id}", h.delete)
		r.Patch("/schedule", h.setSchedule)
	})
}

func (h handler) userID(r *http.Request) (int64, bool) {
	return parseSalaryUserID(r, h.authMiddleware)
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

	record, err := h.svc.Create(r.Context(), userID, input)
	if err != nil {
		writeSalaryError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "Success Create", record)
}

func (h handler) history(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.History(r.Context(), userID)
	if err != nil {
		writeSalaryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) current(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	item, err := h.svc.Current(r.Context(), userID)
	if err != nil {
		writeSalaryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseSalaryID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid salary id")
		return
	}

	input, err := decodeUpdateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	record, err := h.svc.Update(r.Context(), id, userID, input)
	if err != nil {
		writeSalaryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Update", record)
}

func (h handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseSalaryID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid salary id")
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		writeSalaryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Delete", map[string]string{"status": "deleted"})
}

func (h handler) setSchedule(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input, err := decodeSetSalaryDayInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	schedule, err := h.svc.SetSalaryDay(r.Context(), userID, input.SalaryDay)
	if err != nil {
		writeSalaryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Set Salary Date", schedule)
}

func writeSalaryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		if errors.Is(err, ErrInvalidInput) || strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "greater than zero") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
