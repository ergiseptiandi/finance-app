package ai

import (
	"encoding/json"
	"net/http"
	"strings"
)

type chatRequest struct {
	Message string `json:"message"`
}

func decodeChatInput(r *http.Request) (string, error) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", err
	}
	return strings.TrimSpace(req.Message), nil
}
