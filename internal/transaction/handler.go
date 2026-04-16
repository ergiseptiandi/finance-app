package transaction

import (
	"errors"
	"net/http"

	domainauth "finance-backend/internal/auth"
	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (domainauth.AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	TransactionService *Service
	AuthMiddleware     Middleware
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodPost, Path: "/v1/transactions", Summary: "Create transaction", Protected: true},
		{Method: http.MethodGet, Path: "/v1/transactions", Summary: "List transactions", Protected: true},
		{Method: http.MethodGet, Path: "/v1/transactions/summary", Summary: "Get transactions summary", Protected: true},
		{Method: http.MethodGet, Path: "/v1/transactions/{id}", Summary: "Get transaction detail", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/transactions/{id}", Summary: "Update transaction", Protected: true},
		{Method: http.MethodDelete, Path: "/v1/transactions/{id}", Summary: "Delete transaction", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.TransactionService,
		authMiddleware: deps.AuthMiddleware,
	}

	r.Route("/transactions", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Get("/summary", h.summary)
		r.Get("/{id}", h.get)
		r.Patch("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	})
}

func (h handler) getUserID(r *http.Request) (int64, bool) {
	return parseTransactionUserID(r, h.authMiddleware)
}

func (h handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input, err := decodeCreateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	txn, err := h.svc.Create(r.Context(), userID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, txn)
}

func (h handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseTransactionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	txn, err := h.svc.Get(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, txn)
}

func (h handler) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseTransactionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	input, err := decodeUpdateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	txn, err := h.svc.Update(r.Context(), id, userID, input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, txn)
}

func (h handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseTransactionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	err = h.svc.Delete(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter := parseListFilter(r)

	list, err := h.svc.List(r.Context(), userID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h handler) summary(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.svc.Summary(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
