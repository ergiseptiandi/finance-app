package media

import (
	"errors"
	"net/http"
	"strings"

	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type HandlerDependencies struct {
	MediaService   *Service
	AuthMiddleware Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodPost, Path: "/v1/media/upload", Summary: "Upload file", Protected: true},
		{Method: http.MethodGet, Path: "/v1/media/url", Summary: "Get file URL", Protected: true},
		{Method: http.MethodDelete, Path: "/v1/media", Summary: "Delete file", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.MediaService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/media", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Post("/upload", h.upload)
		r.Get("/url", h.getURL)
		r.Delete("/", h.delete)
	})
}

func (h handler) upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseMediaUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	file.Close()

	dir := strings.TrimSpace(r.FormValue("dir"))
	baseURL := requestBaseURL(r)

	item, err := h.svc.Upload(userID, dir, fileHeader, baseURL)
	if err != nil {
		writeMediaError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "Success Create", item)
}

func (h handler) getURL(w http.ResponseWriter, r *http.Request) {
	_, ok := parseMediaUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	publicPath := r.URL.Query().Get("path")
	item, err := h.svc.GetURL(publicPath, requestBaseURL(r))
	if err != nil {
		writeMediaError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) delete(w http.ResponseWriter, r *http.Request) {
	_, ok := parseMediaUserID(r, h.authMiddleware)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	publicPath := r.URL.Query().Get("path")
	if err := h.svc.Delete(publicPath); err != nil {
		writeMediaError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Delete", map[string]string{"status": "deleted"})
}

func writeMediaError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func requestBaseURL(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" && r.TLS != nil {
		scheme = "https"
	}
	return BaseURLFromRequest(SchemeFromHeaderOrTLS(r.TLS != nil, scheme), r.Host)
}
