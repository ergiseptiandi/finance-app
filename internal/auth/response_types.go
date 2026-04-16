package auth

type userResponse struct {
	User User `json:"user"`
}

type statusResponse struct {
	Status string `json:"status"`
}

type forgotPasswordResponse struct {
	Status     string `json:"status"`
	ResetToken string `json:"reset_token,omitempty"`
}
