package auth

import "time"

type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"`
}

type LoginInput struct {
	Email      string
	Password   string
	DeviceName string
}

type AuthResult struct {
	User  User      `json:"user"`
	Token TokenPair `json:"token"`
}

type RegisterInput struct {
	Name       string
	Email      string
	Password   string
	DeviceName string
}

type UpdateProfileInput struct {
	Name  string
	Email string
}

type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
}

type ForgotPasswordInput struct {
	Email string
}

type ResetPasswordInput struct {
	Token       string
	NewPassword string
}
