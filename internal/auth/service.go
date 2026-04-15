package auth

import (
	"context"
	"database/sql"
	"strconv"
	"time"
)

type Service struct {
	users         UserRepository
	refreshTokens RefreshTokenRepository
	tokens        *TokenManager
}

func NewService(users UserRepository, refreshTokens RefreshTokenRepository, tokens *TokenManager) *Service {
	return &Service{
		users:         users,
		refreshTokens: refreshTokens,
		tokens:        tokens,
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
