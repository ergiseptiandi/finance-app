package ai

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ChatContext struct {
	SalaryDay   int    `json:"salary_day,omitempty"`
	PeriodStart string `json:"period_start,omitempty"`
	PeriodEnd   string `json:"period_end,omitempty"`
	PeriodMode  string `json:"period_mode,omitempty"`
}

type chatRequest struct {
	Message string       `json:"message"`
	Context *ChatContext `json:"context,omitempty"`
}

type ChatInput struct {
	Message string
	Context *ChatContext
}

func decodeChatInput(r *http.Request) (ChatInput, error) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return ChatInput{}, err
	}
	return ChatInput{
		Message: strings.TrimSpace(req.Message),
		Context: req.Context,
	}, nil
}
