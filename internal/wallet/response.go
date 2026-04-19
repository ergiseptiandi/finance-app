package wallet

import (
	"encoding/json"
	"net/http"
)

type apiResponse struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
	Data    any    `json:"Data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, message string, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiResponse{Status: http.StatusText(status), Message: message, Data: payload})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, message, nil)
}
