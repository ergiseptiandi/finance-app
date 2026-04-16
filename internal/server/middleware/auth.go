package middleware

import (
	"context"
	"net/http"
	"strings"

	domainauth "finance-backend/internal/auth"
)

type contextKey string

const accessClaimsContextKey contextKey = "access_claims"

type Auth struct {
	authService *domainauth.Service
}

func NewAuth(authService *domainauth.Service) *Auth {
	return &Auth{
		authService: authService,
	}
}

func (m *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		claims, err := m.authService.ParseAccessToken(strings.TrimSpace(parts[1]))
		if err != nil {
			writeAuthError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), accessClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Auth) GetAccessClaims(r *http.Request) (domainauth.AccessTokenClaims, bool) {
	claims, ok := r.Context().Value(accessClaimsContextKey).(domainauth.AccessTokenClaims)
	return claims, ok
}
