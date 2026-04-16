package helpers

import (
	"encoding/json"
	"net/http"
)

type FormatSuccess struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
	Data    any    `json:"Data,omitempty"`
}

type FormatError struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

func SuccessResponse(response any, message string, status string) FormatSuccess {
	return FormatSuccess{
		Status:  status,
		Message: message,
		Data:    response,
	}
}

func ErrorResponse(message string, status string) FormatError {
	return FormatError{
		Status:  status,
		Message: message,
	}
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
