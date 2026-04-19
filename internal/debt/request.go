package debt

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	domainauth "finance-backend/internal/auth"

	"github.com/go-chi/chi/v5"
)

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func parseDebtID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func parsePaymentID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "paymentId"), 10, 64)
}

func parseInstallmentID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "installmentId"), 10, 64)
}

func parseDebtUserID(r *http.Request, middleware Middleware) (int64, bool) {
	claims, ok := middleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	return parseDebtUserIDFromClaims(claims)
}

func parseDebtUserIDFromClaims(claims domainauth.AccessTokenClaims) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(claims.Subject), 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func decodeCreateInput(r *http.Request) (CreateInput, error) {
	var input CreateInput
	if err := decodeJSON(r, &input); err != nil {
		return CreateInput{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	return input, nil
}

func decodeUpdateInput(r *http.Request) (UpdateInput, error) {
	var input UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		return UpdateInput{}, err
	}
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}
	return input, nil
}

func decodeMarkPaidInput(r *http.Request) (MarkInstallmentPaidInput, error) {
	var input MarkInstallmentPaidInput
	if err := decodeJSON(r, &input); err != nil {
		return MarkInstallmentPaidInput{}, err
	}
	return input, nil
}

func parseCreatePaymentMultipart(r *http.Request) (CreatePaymentInput, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return CreatePaymentInput{}, nil, err
	}

	amount, err := strconv.ParseFloat(strings.TrimSpace(r.FormValue("amount")), 64)
	if err != nil {
		return CreatePaymentInput{}, nil, err
	}

	paymentDateStr := strings.TrimSpace(r.FormValue("payment_date"))
	paymentDate, err := time.Parse(time.RFC3339, paymentDateStr)
	if err != nil {
		return CreatePaymentInput{}, nil, err
	}

	var walletID *int64
	if walletIDStr := strings.TrimSpace(r.FormValue("wallet_id")); walletIDStr != "" {
		parsed, err := strconv.ParseInt(walletIDStr, 10, 64)
		if err != nil {
			return CreatePaymentInput{}, nil, err
		}
		walletID = &parsed
	}

	file, fileHeader, err := r.FormFile("proof_image")
	if err != nil {
		return CreatePaymentInput{}, nil, err
	}
	defer file.Close()

	return CreatePaymentInput{
		WalletID:    walletID,
		Amount:      amount,
		PaymentDate: paymentDate,
	}, fileHeader, nil
}

func parseUpdatePaymentMultipart(r *http.Request) (UpdatePaymentInput, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return UpdatePaymentInput{}, nil, err
	}

	var input UpdatePaymentInput
	if amountStr := strings.TrimSpace(r.FormValue("amount")); amountStr != "" {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return UpdatePaymentInput{}, nil, err
		}
		input.Amount = &amount
	}
	if paymentDateStr := strings.TrimSpace(r.FormValue("payment_date")); paymentDateStr != "" {
		paymentDate, err := time.Parse(time.RFC3339, paymentDateStr)
		if err != nil {
			return UpdatePaymentInput{}, nil, err
		}
		input.PaymentDate = &paymentDate
	}
	if walletIDStr := strings.TrimSpace(r.FormValue("wallet_id")); walletIDStr != "" {
		walletID, err := strconv.ParseInt(walletIDStr, 10, 64)
		if err != nil {
			return UpdatePaymentInput{}, nil, err
		}
		input.WalletID = &walletID
	}

	file, fileHeader, err := r.FormFile("proof_image")
	if err != nil && err != http.ErrMissingFile && !strings.Contains(err.Error(), "no such file") {
		return UpdatePaymentInput{}, nil, err
	}
	if err == nil {
		defer file.Close()
	}

	return input, fileHeader, nil
}
