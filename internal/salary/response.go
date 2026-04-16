package salary

import (
	"net/http"
	"strconv"

	"finance-backend/internal/helpers"
)

func writeJSON(w http.ResponseWriter, status int, message string, payload any) {
	helpers.WriteJSON(w, status, helpers.SuccessResponse(payload, message, strconv.Itoa(status)))
}

func writeError(w http.ResponseWriter, status int, message string) {
	helpers.WriteJSON(w, status, helpers.ErrorResponse(message, strconv.Itoa(status)))
}
