package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secretKey       []byte
	issuer          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type AccessTokenClaims struct {
	TokenType string `json:"token_type"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	jwt.RegisteredClaims
}

func NewTokenManager(secret, issuer string, accessTokenTTL, refreshTokenTTL time.Duration) *TokenManager {
	return &TokenManager{
		secretKey:       []byte(secret),
		issuer:          issuer,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (m *TokenManager) GenerateAccessToken(user User) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.accessTokenTTL)
	claims := AccessTokenClaims{
		TokenType: "access",
		Email:     user.Email,
		Name:      user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			Issuer:    m.issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

func (m *TokenManager) ParseAccessToken(rawToken string) (AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		return m.secretKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		if isJWTExpired(err) {
			return AccessTokenClaims{}, ErrExpiredToken
		}

		return AccessTokenClaims{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid || claims.TokenType != "access" {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	return *claims, nil
}

func (m *TokenManager) GenerateRefreshToken() (string, string, time.Time, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", time.Time{}, err
	}

	rawToken := hex.EncodeToString(tokenBytes)
	tokenHash := HashToken(rawToken)
	expiresAt := time.Now().Add(m.refreshTokenTTL)

	return rawToken, tokenHash, expiresAt, nil
}

func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func isJWTExpired(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet)
}
