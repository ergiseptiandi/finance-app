package ai

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	domainauth "finance-backend/internal/auth"
	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (domainauth.AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	AIService      *Service
	AuthMiddleware Middleware
	AuthService    *domainauth.Service
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
	authService    *domainauth.Service
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodPost, Path: "/v1/ai/chat", Summary: "Send message to AI financial analyst", Protected: true},
		{Method: http.MethodGet, Path: "/v1/ai/usage", Summary: "Get AI chat usage info", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.AIService,
		authMiddleware: deps.AuthMiddleware,
		authService:    deps.AuthService,
	}

	r.Route("/ai", func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequireAuth)
		r.Post("/chat", h.chat)
		r.Get("/usage", h.usage)
	})
}

func (h handler) chat(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authMiddleware.GetAccessClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.authService.GetUserByClaims(r.Context(), claims)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	message, err := decodeChatInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	reply, err := h.svc.Analyze(r.Context(), userID, user.Name, message)
	if err != nil {
		log.Printf("[ai] analyze error for user %d: %v", userID, err)
		if errors.Is(err, ErrChatLimitExceeded) {
			writeError(w, http.StatusTooManyRequests, "Batas chat AI anda sudah habis. Silakan hubungi admin untuk menambah kuota.")
			return
		}
		writeError(w, http.StatusInternalServerError, "Gagal mendapatkan analisis. Coba lagi nanti.")
		return
	}

	writeJSON(w, http.StatusOK, "Success", AnalysisResponse{Reply: reply})
}

func (h handler) usage(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.authMiddleware.GetAccessClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	usage, err := h.svc.GetUsage(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, "Success", usage)
}
