package middleware

import (
	"errors"
	"net/http"
	"strconv"

	domainauth "finance-backend/internal/auth"
	"finance-backend/internal/helpers"
)

func writeJSON(w http.ResponseWriter, status int, message string, payload any) {
	helpers.WriteJSON(w, status, helpers.SuccessResponse(payload, message, strconv.Itoa(status)))
}

func writeError(w http.ResponseWriter, status int, message string) {
	helpers.WriteJSON(w, status, helpers.ErrorResponse(message, strconv.Itoa(status)))
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domainauth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domainauth.ErrInvalidToken), errors.Is(err, domainauth.ErrExpiredToken):
		writeError(w, http.StatusUnauthorized, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
