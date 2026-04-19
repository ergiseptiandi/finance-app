package category

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"finance-backend/internal/auth"
	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (auth.AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	CategoryService *Service
	AuthMiddleware  Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodPost, Path: "/v1/categories", Summary: "Create category", Protected: true},
		{Method: http.MethodGet, Path: "/v1/categories", Summary: "List categories", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/categories/{id}", Summary: "Update category", Protected: true},
		{Method: http.MethodDelete, Path: "/v1/categories/{id}", Summary: "Delete category", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.CategoryService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/categories", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Patch("/{id}", h.update)
		r.Delete("/{id}", h.delete)
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
		writeCategoryError(w, err)
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

	id, err := parseCategoryID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	input, err := decodeUpdateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}

	item, err := h.svc.Update(r.Context(), userID, id, input)
	if err != nil {
		writeCategoryError(w, err)
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

	id, err := parseCategoryID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		writeCategoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Delete", map[string]string{"status": "deleted"})
}

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter := parseListFilter(r)

	items, err := h.svc.List(r.Context(), userID, filter)
	if err != nil {
		writeCategoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func writeCategoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	default:
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h handler) userID(r *http.Request) (int64, bool) {
	claims, ok := h.authMiddleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	id, err := strconv.ParseInt(strings.TrimSpace(claims.Subject), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}
