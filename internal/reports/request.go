package reports

import "net/http"

func parseReportsUserID(r *http.Request, middleware Middleware) (int64, bool) {
	claims, ok := middleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	return parseReportsUserIDFromClaims(claims.Subject)
}

func parseReportsUserIDFromClaims(subject string) (int64, bool) {
	return parseInt64(subject)
}

func parseInt64(value string) (int64, bool) {
	var id int64
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		id = id*10 + int64(ch-'0')
	}
	if id == 0 {
		return 0, false
	}
	return id, true
}
