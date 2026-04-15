package auth

import (
	"context"
	"time"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id int64) (User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, params CreateRefreshTokenParams) error
	FindActiveByHash(ctx context.Context, tokenHash string) (RefreshToken, error)
	Rotate(ctx context.Context, currentTokenHash string, nextToken CreateRefreshTokenParams) error
	RevokeByHash(ctx context.Context, tokenHash string) error
}

type CreateRefreshTokenParams struct {
	UserID     int64
	TokenHash  string
	DeviceName string
	ExpiresAt  time.Time
}
