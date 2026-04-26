package budget

import (
	"net/http"
	"strconv"

	"finance-backend/internal/helpers"
)

func writeJSON(w http.ResponseWriter, status int, message string, payload any) {
	helper := helpers.SuccessResponse(payload, message, strconv.Itoa(status))
	helpers.WriteJSON(w, status, helper)
}

func writeError(w http.ResponseWriter, status int, message string) {
	helpers.WriteJSON(w, status, helpers.ErrorResponse(message, strconv.Itoa(status)))
}
