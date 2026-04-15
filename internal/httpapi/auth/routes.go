package auth

import (
	"net/http"
	"strings"

	domainauth "finance-backend/internal/auth"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (domainauth.AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	AuthService    *domainauth.Service
	AuthMiddleware Middleware
}

type handler struct {
	authService    *domainauth.Service
	authMiddleware Middleware
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		authService:    deps.AuthService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.login)
		r.Post("/refresh", h.refresh)
		r.Post("/logout", h.logout)

		r.Group(func(r chi.Router) {
			r.Use(h.authMiddleware.RequireAuth)
			r.Get("/me", h.me)
		})
	})
}

func (h handler) login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		DeviceName string `json:"device_name"`
	}

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Login(r.Context(), domainauth.LoginInput{
		Email:      strings.TrimSpace(request.Email),
		Password:   request.Password,
		DeviceName: strings.TrimSpace(request.DeviceName),
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h handler) refresh(w http.ResponseWriter, r *http.Request) {
	var request struct {
		RefreshToken string `json:"refresh_token"`
		DeviceName   string `json:"device_name"`
	}

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Refresh(r.Context(), strings.TrimSpace(request.RefreshToken), strings.TrimSpace(request.DeviceName))
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h handler) logout(w http.ResponseWriter, r *http.Request) {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.authService.Logout(r.Context(), strings.TrimSpace(request.RefreshToken)); err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h handler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authMiddleware.GetAccessClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.authService.GetUserByClaims(r.Context(), claims)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]domainauth.User{"user": user})
}
