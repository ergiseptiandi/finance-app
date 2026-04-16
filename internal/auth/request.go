package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

type loginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceName string `json:"device_name"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceName   string `json:"device_name"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type registerRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceName string `json:"device_name"`
}

type updateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func decodeLoginInput(r *http.Request) (LoginInput, error) {
	var request loginRequest
	if err := decodeJSON(r, &request); err != nil {
		return LoginInput{}, err
	}

	return LoginInput{
		Email:      strings.TrimSpace(request.Email),
		Password:   request.Password,
		DeviceName: strings.TrimSpace(request.DeviceName),
	}, nil
}

func decodeRefreshInput(r *http.Request) (string, string, error) {
	var request refreshRequest
	if err := decodeJSON(r, &request); err != nil {
		return "", "", err
	}

	return strings.TrimSpace(request.RefreshToken), strings.TrimSpace(request.DeviceName), nil
}

func decodeLogoutInput(r *http.Request) (string, error) {
	var request logoutRequest
	if err := decodeJSON(r, &request); err != nil {
		return "", err
	}

	return strings.TrimSpace(request.RefreshToken), nil
}

func decodeRegisterInput(r *http.Request) (RegisterInput, error) {
	var request registerRequest
	if err := decodeJSON(r, &request); err != nil {
		return RegisterInput{}, err
	}

	return RegisterInput{
		Name:       strings.TrimSpace(request.Name),
		Email:      strings.TrimSpace(request.Email),
		Password:   request.Password,
		DeviceName: strings.TrimSpace(request.DeviceName),
	}, nil
}

func decodeUpdateProfileInput(r *http.Request) (UpdateProfileInput, error) {
	var request updateProfileRequest
	if err := decodeJSON(r, &request); err != nil {
		return UpdateProfileInput{}, err
	}

	return UpdateProfileInput{
		Name:  strings.TrimSpace(request.Name),
		Email: strings.TrimSpace(request.Email),
	}, nil
}

func decodeChangePasswordInput(r *http.Request) (ChangePasswordInput, error) {
	var request changePasswordRequest
	if err := decodeJSON(r, &request); err != nil {
		return ChangePasswordInput{}, err
	}

	return ChangePasswordInput{
		OldPassword: request.OldPassword,
		NewPassword: request.NewPassword,
	}, nil
}

func decodeForgotPasswordInput(r *http.Request) (ForgotPasswordInput, error) {
	var request forgotPasswordRequest
	if err := decodeJSON(r, &request); err != nil {
		return ForgotPasswordInput{}, err
	}

	return ForgotPasswordInput{
		Email: strings.TrimSpace(request.Email),
	}, nil
}

func decodeResetPasswordInput(r *http.Request) (ResetPasswordInput, error) {
	var request resetPasswordRequest
	if err := decodeJSON(r, &request); err != nil {
		return ResetPasswordInput{}, err
	}

	return ResetPasswordInput{
		Token:       strings.TrimSpace(request.Token),
		NewPassword: request.NewPassword,
	}, nil
}
