package auth

import (
	"net/http"

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
	input, err := decodeLoginInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Login(r.Context(), input)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h handler) refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, deviceName, err := decodeRefreshInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Refresh(r.Context(), refreshToken, deviceName)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h handler) logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := decodeLogoutInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.authService.Logout(r.Context(), refreshToken); err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "logged_out"})
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

	writeJSON(w, http.StatusOK, userResponse{User: user})
}

func (h handler) register(w http.ResponseWriter, r *http.Request) {
	input, err := decodeRegisterInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Register(r.Context(), input)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	input, err := decodeUpdateProfileInput(r)
	if err != nil {
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

	updatedUser, err := h.authService.UpdateProfile(r.Context(), user.ID, input)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userResponse{User: updatedUser})
}

func (h handler) changePassword(w http.ResponseWriter, r *http.Request) {
	input, err := decodeChangePasswordInput(r)
	if err != nil {
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

	if err := h.authService.ChangePassword(r.Context(), user.ID, input); err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "password_changed"})
}

func (h handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	input, err := decodeForgotPasswordInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.authService.ForgotPassword(r.Context(), input)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, forgotPasswordResponse{
		Status:     "email_sent",
		ResetToken: token,
	})
}

func (h handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	input, err := decodeResetPasswordInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.authService.ResetPassword(r.Context(), input); err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "password_reset"})
}
