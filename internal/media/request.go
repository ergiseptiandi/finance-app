package media

import (
	"encoding/json"
	"net/http"
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

func parseMediaUserID(r *http.Request, middleware Middleware) (int64, bool) {
	claims, ok := middleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	id, err := ParseInt64(strings.TrimSpace(claims.Subject))
	if err != nil {
		return 0, false
	}

	return id, true
}
