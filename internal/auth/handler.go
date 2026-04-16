package auth

import (
	"net/http"
	"strings"

	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	AuthService    *Service
	AuthMiddleware Middleware
}

type handler struct {
	authService    *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodPost, Path: "/v1/auth/register", Summary: "Register a new user"},
		{Method: http.MethodPost, Path: "/v1/auth/login", Summary: "Authenticate bootstrap user and issue token pair"},
		{Method: http.MethodPost, Path: "/v1/auth/refresh", Summary: "Rotate refresh token and issue a new token pair"},
		{Method: http.MethodPost, Path: "/v1/auth/logout", Summary: "Revoke refresh token"},
		{Method: http.MethodPost, Path: "/v1/auth/forgot-password", Summary: "Request a password reset link"},
		{Method: http.MethodPost, Path: "/v1/auth/reset-password", Summary: "Reset password using token"},
		{Method: http.MethodGet, Path: "/v1/auth/me", Summary: "Get current authenticated user", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/auth/profile", Summary: "Update current authenticated user's profile", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/auth/password", Summary: "Change current authenticated user's password", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		authService:    deps.AuthService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
		r.Post("/refresh", h.refresh)
		r.Post("/logout", h.logout)
		r.Post("/forgot-password", h.forgotPassword)
		r.Post("/reset-password", h.resetPassword)

		r.Group(func(r chi.Router) {
			r.Use(h.authMiddleware.RequireAuth)
			r.Get("/me", h.me)
			r.Patch("/profile", h.updateProfile)
			r.Patch("/password", h.changePassword)
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

	result, err := h.authService.Login(r.Context(), LoginInput{
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

	writeJSON(w, http.StatusOK, map[string]User{"user": user})
}

func (h handler) register(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name       string `json:"name"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		DeviceName string `json:"device_name"`
	}

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Register(r.Context(), RegisterInput{
		Name:       strings.TrimSpace(request.Name),
		Email:      strings.TrimSpace(request.Email),
		Password:   request.Password,
		DeviceName: strings.TrimSpace(request.DeviceName),
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

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

	updatedUser, err := h.authService.UpdateProfile(r.Context(), user.ID, UpdateProfileInput{
		Name:  strings.TrimSpace(request.Name),
		Email: strings.TrimSpace(request.Email),
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]User{"user": updatedUser})
}

func (h handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var request struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

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

	if err := h.authService.ChangePassword(r.Context(), user.ID, ChangePasswordInput{
		OldPassword: request.OldPassword,
		NewPassword: request.NewPassword,
	}); err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "password_changed"})
}

func (h handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.authService.ForgotPassword(r.Context(), ForgotPasswordInput{
		Email: strings.TrimSpace(request.Email),
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	// We return the token for debugging/development
	writeJSON(w, http.StatusOK, map[string]string{"status": "email_sent", "reset_token": token})
}

func (h handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.authService.ResetPassword(r.Context(), ResetPasswordInput{
		Token:       strings.TrimSpace(request.Token),
		NewPassword: request.NewPassword,
	}); err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "password_reset"})
}
