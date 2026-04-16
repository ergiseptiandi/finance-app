package alerts

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type HandlerDependencies struct {
	AlertService   *Service
	AuthMiddleware Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodGet, Path: "/v1/alerts", Summary: "List alert history", Protected: true},
		{Method: http.MethodPost, Path: "/v1/alerts/evaluate", Summary: "Evaluate smart alerts", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/alerts/{id}/read", Summary: "Mark alert as read", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.AlertService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/alerts", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Get("/", h.list)
		r.Post("/evaluate", h.evaluate)
		r.Patch("/{id}/read", h.markRead)
	})
}

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseAlertListFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert filter")
		return
	}

	userID, ok := parseAlertsUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.List(r.Context(), userID, filter)
	if err != nil {
		writeAlertError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) evaluate(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseAlertsUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input, err := decodeEvaluateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.svc.Evaluate(r.Context(), userID, input)
	if err != nil {
		writeAlertError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Evaluate", items)
}

func (h handler) markRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseAlertsUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseAlertIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	item, err := h.svc.MarkRead(r.Context(), userID, id)
	if err != nil {
		writeAlertError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Update", item)
}

func writeAlertError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func parseAlertIDParam(r *http.Request) (int64, error) {
	value := strings.TrimSpace(chi.URLParam(r, "id"))
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid alert id")
	}

	return id, nil
}
