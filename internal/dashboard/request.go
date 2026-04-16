package dashboard

import (
	"net/http"
	"strconv"
	"strings"
)

func parseDashboardUserID(r *http.Request, middleware Middleware) (int64, bool) {
	claims, ok := middleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	id, err := strconv.ParseInt(strings.TrimSpace(claims.Subject), 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}
