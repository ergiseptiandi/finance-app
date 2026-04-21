package notifications

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type HandlerDependencies struct {
	NotificationService *Service
	AuthMiddleware      Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodGet, Path: "/v1/notifications", Summary: "List notifications", Protected: true},
		{Method: http.MethodGet, Path: "/v1/notifications/settings", Summary: "Get notification settings", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/notifications/settings", Summary: "Update notification settings", Protected: true},
		{Method: http.MethodPost, Path: "/v1/notifications/generate", Summary: "Generate scheduled reminders", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/notifications/{id}/read", Summary: "Mark notification as read", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.NotificationService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/notifications", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Get("/", h.list)
		r.Get("/settings", h.getSettings)
		r.Patch("/settings", h.updateSettings)
		r.Post("/generate", h.generate)
		r.Patch("/{id}/read", h.markRead)
	})
}

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseNotificationsUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter, err := parseNotificationFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification filter")
		return
	}

	items, err := h.svc.List(r.Context(), userID, filter)
	if err != nil {
		writeNotificationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) getSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseNotificationsUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	item, err := h.svc.GetSettings(r.Context(), userID)
	if err != nil {
		writeNotificationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseNotificationsUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input, err := decodeUpdateSettingsInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.UpdateSettings(r.Context(), userID, input)
	if err != nil {
		writeNotificationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Update", item)
}

func (h handler) generate(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseNotificationsUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.Generate(r.Context(), userID)
	if err != nil {
		writeNotificationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Generate", items)
}

func (h handler) markRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseNotificationsUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseNotificationID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	_, err = h.svc.MarkRead(r.Context(), userID, id)
	if err != nil {
		writeNotificationError(w, err)
		return
	}

	writeReadSuccess(w)
}

func writeNotificationError(w http.ResponseWriter, err error) {
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

func parseNotificationID(r *http.Request) (int64, error) {
	value := strings.TrimSpace(chi.URLParam(r, "id"))
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid notification id")
	}

	return id, nil
}
