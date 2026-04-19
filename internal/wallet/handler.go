package wallet

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
	WalletService  *Service
	AuthMiddleware Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodPost, Path: "/v1/wallets", Summary: "Create wallet", Protected: true},
		{Method: http.MethodGet, Path: "/v1/wallets", Summary: "List wallets", Protected: true},
		{Method: http.MethodGet, Path: "/v1/wallets/summary", Summary: "Get wallet summary", Protected: true},
		{Method: http.MethodGet, Path: "/v1/wallets/{id}", Summary: "Get wallet detail", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/wallets/{id}", Summary: "Update wallet", Protected: true},
		{Method: http.MethodDelete, Path: "/v1/wallets/{id}", Summary: "Delete wallet", Protected: true},
		{Method: http.MethodPost, Path: "/v1/wallet-transfers", Summary: "Create wallet transfer", Protected: true},
		{Method: http.MethodGet, Path: "/v1/wallet-transfers", Summary: "List wallet transfers", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.WalletService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/wallets", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Get("/summary", h.summary)
		r.Get("/{id}", h.get)
		r.Patch("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	})

	r.Route("/wallet-transfers", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Post("/", h.createTransfer)
		r.Get("/", h.listTransfers)
	})
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
		writeWalletError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "Success Create", item)
}

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeWalletError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) summary(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	item, err := h.svc.Summary(r.Context(), userID)
	if err != nil {
		writeWalletError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", item)
}

func (h handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseWalletID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid wallet id")
		return
	}

	item, err := h.svc.GetByID(r.Context(), userID, id)
	if err != nil {
		writeWalletError(w, err)
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

	id, err := parseWalletID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid wallet id")
		return
	}

	input, err := decodeUpdateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), userID, id, input)
	if err != nil {
		writeWalletError(w, err)
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

	id, err := parseWalletID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid wallet id")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		writeWalletError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Delete", map[string]string{"status": "deleted"})
}

func (h handler) createTransfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input, err := decodeCreateTransferInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.CreateTransfer(r.Context(), userID, input)
	if err != nil {
		writeWalletError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "Success Create", item)
}

func (h handler) listTransfers(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.ListTransfers(r.Context(), userID)
	if err != nil {
		writeWalletError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func writeWalletError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
	default:
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "positive number") || strings.Contains(err.Error(), "greater than zero") || strings.Contains(err.Error(), "greater than or equal to zero") || strings.Contains(err.Error(), "must be different") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
