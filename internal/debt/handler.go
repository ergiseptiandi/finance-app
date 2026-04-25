package debt

import (
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	domainauth "finance-backend/internal/auth"
	"finance-backend/internal/server/routeinfo"

	"github.com/go-chi/chi/v5"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (domainauth.AccessTokenClaims, bool)
}

type FileStorage interface {
	SaveMultipartFile(relativeDir string, fileHeader *multipart.FileHeader) (string, error)
	Delete(publicPath string) error
}

type HandlerDependencies struct {
	DebtService    *Service
	AuthMiddleware Middleware
	Storage        FileStorage
}

type handler struct {
	svc            *Service
	authMiddleware Middleware
	storage        FileStorage
}

func Definitions() []routeinfo.RouteInfo {
	return []routeinfo.RouteInfo{
		{Method: http.MethodPost, Path: "/v1/debts", Summary: "Create debt", Protected: true},
		{Method: http.MethodGet, Path: "/v1/debts", Summary: "Get debt list", Protected: true},
		{Method: http.MethodGet, Path: "/v1/debts/{id}", Summary: "Get debt detail", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/debts/{id}", Summary: "Update debt", Protected: true},
		{Method: http.MethodDelete, Path: "/v1/debts/{id}", Summary: "Delete debt", Protected: true},
		{Method: http.MethodPost, Path: "/v1/debts/{id}/payments", Summary: "Create debt payment", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/debts/{id}/payments/{paymentId}", Summary: "Update debt payment", Protected: true},
		{Method: http.MethodGet, Path: "/v1/debts/{id}/payments", Summary: "Get payment history", Protected: true},
		{Method: http.MethodGet, Path: "/v1/debts/{id}/installments", Summary: "Get debt installments", Protected: true},
		{Method: http.MethodPatch, Path: "/v1/debts/{id}/installments/{installmentId}/paid", Summary: "Mark installment as paid", Protected: true},
	}
}

func RegisterRoutes(r chi.Router, deps HandlerDependencies) {
	h := handler{
		svc:            deps.DebtService,
		authMiddleware: deps.AuthMiddleware,
		storage:        deps.Storage,
	}

	r.Route("/debts", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Get("/{id}", h.detail)
		r.Patch("/{id}", h.update)
		r.Delete("/{id}", h.delete)
		r.Post("/{id}/payments", h.createPayment)
		r.Patch("/{id}/payments/{paymentId}", h.updatePayment)
		r.Get("/{id}/payments", h.paymentHistory)
		r.Get("/{id}/installments", h.installments)
		r.Patch("/{id}/installments/{installmentId}/paid", h.markInstallmentPaid)
	})
}

func (h handler) userID(r *http.Request) (int64, bool) {
	return parseDebtUserID(r, h.authMiddleware)
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

	detail, err := h.svc.Create(r.Context(), userID, input)
	if err != nil {
		writeDebtError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "Success Create", detail)
}

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeDebtError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) detail(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseDebtID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}

	item, err := h.svc.Detail(r.Context(), userID, id)
	if err != nil {
		writeDebtError(w, err)
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

	id, err := parseDebtID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}

	input, err := decodeUpdateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), userID, id, input)
	if err != nil {
		writeDebtError(w, err)
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

	id, err := parseDebtID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		writeDebtError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Delete", map[string]string{"status": "deleted"})
}

func (h handler) createPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	debtID, err := parseDebtID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}

	input, fileHeader, err := parseCreatePaymentMultipart(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	publicPath, err := h.storage.SaveMultipartFile(filepath.Join("debt-payments", strconv.FormatInt(userID, 10), strconv.FormatInt(debtID, 10)), fileHeader)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save proof image")
		return
	}

	input.ProofImage = publicPath
	payment, err := h.svc.CreatePayment(r.Context(), userID, debtID, input)
	if err != nil {
		_ = h.storage.Delete(publicPath)
		writeDebtError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "Success Create", payment)
}

func (h handler) updatePayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	debtID, err := parseDebtID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}

	paymentID, err := parsePaymentID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment id")
		return
	}

	input, fileHeader, err := parseUpdatePaymentMultipart(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if fileHeader != nil {
		publicPath, err := h.storage.SaveMultipartFile(filepath.Join("debt-payments", strconv.FormatInt(userID, 10), strconv.FormatInt(debtID, 10)), fileHeader)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save proof image")
			return
		}
		input.ProofImage = &publicPath
	}

	payment, err := h.svc.UpdatePayment(r.Context(), userID, debtID, paymentID, input)
	if err != nil {
		if input.ProofImage != nil {
			_ = h.storage.Delete(*input.ProofImage)
		}
		writeDebtError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Update", payment)
}

func (h handler) paymentHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	debtID, err := parseDebtID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}

	items, err := h.svc.PaymentHistory(r.Context(), userID, debtID)
	if err != nil {
		writeDebtError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) installments(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	debtID, err := parseDebtID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}

	items, err := h.svc.Installments(r.Context(), userID, debtID)
	if err != nil {
		writeDebtError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Get", items)
}

func (h handler) markInstallmentPaid(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	debtID, err := parseDebtID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}

	installmentID, err := parseInstallmentID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid installment id")
		return
	}

	input, err := decodeMarkPaidInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.MarkInstallmentPaid(r.Context(), userID, debtID, installmentID, input.PaidAt)
	if err != nil {
		writeDebtError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "Success Update", item)
}

func writeDebtError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrNoInstallment):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInsufficientBalance):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "positive number") || strings.Contains(err.Error(), "greater than zero") || strings.Contains(err.Error(), "cannot change") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
