package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"finance-backend/internal/auth"
	"finance-backend/internal/helpers"
)

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, message string, payload any) {
	helpers.WriteJSON(w, status, helpers.SuccessResponse(payload, message, strconv.Itoa(status)))
}

func writeError(w http.ResponseWriter, status int, message string) {
	helpers.WriteJSON(w, status, helpers.ErrorResponse(message, strconv.Itoa(status)))
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, auth.ErrInvalidToken), errors.Is(err, auth.ErrExpiredToken):
		writeError(w, http.StatusUnauthorized, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
