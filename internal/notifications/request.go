package notifications

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

func parseNotificationsUserID(r *http.Request, middleware Middleware) (int64, bool) {
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

func decodeUpdateSettingsInput(r *http.Request) (UpdateSettingsInput, error) {
	var input UpdateSettingsInput
	if err := decodeJSON(r, &input); err != nil {
		return UpdateSettingsInput{}, err
	}

	return input, nil
}

func parseNotificationFilter(r *http.Request) (NotificationFilter, error) {
	filter := NotificationFilter{}
	query := r.URL.Query()

	if value := strings.TrimSpace(query.Get("kind")); value != "" {
		filter.Kind = &value
	}

	if value := strings.TrimSpace(query.Get("read")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return NotificationFilter{}, err
		}
		filter.Read = &parsed
	}

	return filter, nil
}
