package salary

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	domainauth "finance-backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}

func parseSalaryID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

func parseSalaryUserID(r *http.Request, middleware Middleware) (int64, bool) {
	claims, ok := middleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	return parseSalaryUserIDFromClaims(claims)
}

func parseSalaryUserIDFromClaims(claims domainauth.AccessTokenClaims) (int64, bool) {
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

	input.Note = strings.TrimSpace(input.Note)
	return input, nil
}

func decodeUpdateInput(r *http.Request) (UpdateInput, error) {
	var input UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		return UpdateInput{}, err
	}

	if input.Note != nil {
		trimmed := strings.TrimSpace(*input.Note)
		input.Note = &trimmed
	}

	return input, nil
}

func decodeSetSalaryDayInput(r *http.Request) (SetSalaryDayInput, error) {
	var input SetSalaryDayInput
	if err := decodeJSON(r, &input); err != nil {
		return SetSalaryDayInput{}, err
	}

	return input, nil
}
