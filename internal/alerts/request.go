package alerts

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"finance-backend/internal/auth"
)

type Middleware interface {
	RequireAuth(next http.Handler) http.Handler
	GetAccessClaims(r *http.Request) (auth.AccessTokenClaims, bool)
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}

func parseAlertsUserID(r *http.Request, middleware Middleware) (int64, bool) {
	claims, ok := middleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	id, err := strconv.ParseInt(strings.TrimSpace(claims.Subject), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

func decodeEvaluateInput(r *http.Request) (EvaluateInput, error) {
	var input EvaluateInput
	if err := decodeJSON(r, &input); err != nil {
		return EvaluateInput{}, err
	}

	return input, nil
}

func parseAlertListFilter(r *http.Request) (AlertListFilter, error) {
	query := r.URL.Query()
	filter := AlertListFilter{}

	if value := strings.TrimSpace(query.Get("type")); value != "" {
		filter.Type = &value
	}

	if value := strings.TrimSpace(query.Get("read")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return AlertListFilter{}, err
		}
		filter.Read = &parsed
	}

	return filter, nil
}
