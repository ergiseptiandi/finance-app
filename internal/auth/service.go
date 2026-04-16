package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"finance-backend/internal/mail"
)

type Service struct {
	users         UserRepository
	refreshTokens RefreshTokenRepository
	passwordReset PasswordResetTokenRepository
	tokens        *TokenManager
	mailer        mail.Sender
}

func NewService(users UserRepository, refreshTokens RefreshTokenRepository, passwordReset PasswordResetTokenRepository, tokens *TokenManager, mailer mail.Sender) *Service {
	return &Service{
		users:         users,
		refreshTokens: refreshTokens,
		passwordReset: passwordReset,
		tokens:        tokens,
		mailer:        mailer,
	}
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	if input.Email == "" || input.Password == "" {
		return AuthResult{}, ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return AuthResult{}, ErrInvalidCredentials
		}

		return AuthResult{}, err
	}

	if err := CheckPassword(user.PasswordHash, input.Password); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, user, input.DeviceName)
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken, deviceName string) (AuthResult, error) {
	if rawRefreshToken == "" {
		return AuthResult{}, ErrInvalidToken
	}

	storedToken, err := s.refreshTokens.FindActiveByHash(ctx, HashToken(rawRefreshToken))
	if err != nil {
		if err == sql.ErrNoRows {
			return AuthResult{}, ErrInvalidToken
		}

		return AuthResult{}, err
	}

	if storedToken.RevokedAt != nil {
		return AuthResult{}, ErrInvalidToken
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return AuthResult{}, ErrExpiredToken
	}

	user, err := s.users.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return AuthResult{}, err
	}

	accessToken, accessExpiresAt, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return AuthResult{}, err
	}

	nextRefreshToken, nextTokenHash, refreshExpiresAt, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}

	if deviceName == "" {
		deviceName = storedToken.DeviceName
	}

	if err := s.refreshTokens.Rotate(ctx, storedToken.TokenHash, CreateRefreshTokenParams{
		UserID:     user.ID,
		TokenHash:  nextTokenHash,
		DeviceName: deviceName,
		ExpiresAt:  refreshExpiresAt,
	}); err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User: user,
		Token: TokenPair{
			AccessToken:           accessToken,
			AccessTokenExpiresAt:  accessExpiresAt,
			RefreshToken:          nextRefreshToken,
			RefreshTokenExpiresAt: refreshExpiresAt,
			TokenType:             "Bearer",
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return ErrInvalidToken
	}

	return s.refreshTokens.RevokeByHash(ctx, HashToken(rawRefreshToken))
}

func (s *Service) ParseAccessToken(rawAccessToken string) (AccessTokenClaims, error) {
	return s.tokens.ParseAccessToken(rawAccessToken)
}

func (s *Service) GetUserByClaims(ctx context.Context, claims AccessTokenClaims) (User, error) {
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return User{}, ErrInvalidToken
	}

	return s.users.FindByID(ctx, userID)
}

func (s *Service) issueTokenPair(ctx context.Context, user User, deviceName string) (AuthResult, error) {
	accessToken, accessExpiresAt, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return AuthResult{}, err
	}

	refreshToken, refreshTokenHash, refreshExpiresAt, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}

	if err := s.refreshTokens.Create(ctx, CreateRefreshTokenParams{
		UserID:     user.ID,
		TokenHash:  refreshTokenHash,
		DeviceName: deviceName,
		ExpiresAt:  refreshExpiresAt,
	}); err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User: user,
		Token: TokenPair{
			AccessToken:           accessToken,
			AccessTokenExpiresAt:  accessExpiresAt,
			RefreshToken:          refreshToken,
			RefreshTokenExpiresAt: refreshExpiresAt,
			TokenType:             "Bearer",
		},
	}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	if input.Email == "" || input.Password == "" || input.Name == "" {
		return AuthResult{}, ErrInvalidCredentials // Generic for now, ideally specific validation error
	}

	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		return AuthResult{}, err
	}

	user := User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hashedPassword,
	}

	id, err := s.users.CreateUser(ctx, user)
	if err != nil {
		return AuthResult{}, err
	}

	user.ID = id
	return s.issueTokenPair(ctx, user, input.DeviceName)
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return User{}, err
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		user.Email = input.Email
	}

	if err := s.users.UpdateUser(ctx, user); err != nil {
		return User{}, err
	}

	return s.users.FindByID(ctx, userID)
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, input ChangePasswordInput) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := CheckPassword(user.PasswordHash, input.OldPassword); err != nil {
		return ErrInvalidCredentials
	}

	hashedPassword, err := HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	return s.users.UpdateUserPassword(ctx, userID, hashedPassword)
}

func (s *Service) ForgotPassword(ctx context.Context, input ForgotPasswordInput) (string, error) {
	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		// Do not leak if user does not exist
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	// In real setup, generate securely using token manager or crypto rand
	rawToken, tokenHash, _, err := s.tokens.GenerateRefreshToken() // We can reuse generation logic to give us a hex random string+hash
	if err != nil {
		return "", err
	}

	token := PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.passwordReset.Create(ctx, token); err != nil {
		return "", err
	}

	if s.mailer != nil {
		resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", rawToken)
		subject := "Password Reset Request"
		body := fmt.Sprintf("Hello,\r\n\r\nYou requested a password reset. Please click on the link below or copy it to your browser to reset your password:\r\n\r\n%s\r\n\r\nIf you did not request this, please ignore this email.\r\n", resetLink)
		
		if err := s.mailer.SendEmail(ctx, user.Email, subject, body); err != nil {
			// We can log the error, but we shouldn't fail the request completely
			fmt.Printf("Failed to send reset email to %s: %v\n", user.Email, err)
		} else {
			fmt.Printf("Sent reset email to %s\n", user.Email)
		}
	} else {
		// MOCK: Print raw token if no mailer is configured to allow local dev
		fmt.Printf("MOCK EMAIL => To: %s, Subj: Password Reset, Token: %s\n", user.Email, rawToken)
	}

	return rawToken, nil
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	tokenHash := HashToken(input.Token)

	token, err := s.passwordReset.FindByHash(ctx, tokenHash)
	if err != nil {
		return ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		s.passwordReset.Delete(ctx, token.ID)
		return ErrExpiredToken
	}

	hashedPassword, err := HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	if err := s.users.UpdateUserPassword(ctx, token.UserID, hashedPassword); err != nil {
		return err
	}

	_ = s.passwordReset.DeleteAllForUser(ctx, token.UserID)

	return nil
}
