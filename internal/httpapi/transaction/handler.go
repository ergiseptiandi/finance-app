package transaction

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	domainauth "finance-backend/internal/auth"
	"finance-backend/internal/httpapi/routeinfo"
	"finance-backend/internal/transaction"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (domainauth.AccessTokenClaims, bool)
}

type HandlerDependencies struct {
	TransactionService *transaction.Service
	AuthMiddleware     Middleware
}

type handler struct {
	svc            *transaction.Service
	authMiddleware Middleware
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
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
	claims, ok := h.authMiddleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (h handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input transaction.CreateInput
	if err := decodeJSON(r, &input); err != nil {
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

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	txn, err := h.svc.Get(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, transaction.ErrNotFound) {
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

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	var input transaction.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	txn, err := h.svc.Update(r.Context(), id, userID, input)
	if err != nil {
		if errors.Is(err, transaction.ErrNotFound) {
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

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	err = h.svc.Delete(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, transaction.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	var filter transaction.ListFilter

	if s := q.Get("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			filter.StartDate = &t
		}
	}
	if s := q.Get("end_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			filter.EndDate = &t
		}
	}
	if s := q.Get("category"); s != "" {
		filter.Category = &s
	}
	if s := q.Get("type"); s != "" {
		t := transaction.Type(s)
		filter.Type = &t
	}
	if pageStr := q.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = p
		}
	}
	if perPageStr := q.Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil {
			filter.PerPage = pp
		}
	}

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
