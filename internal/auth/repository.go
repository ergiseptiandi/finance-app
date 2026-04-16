package auth

import (
	"context"
	"time"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id int64) (User, error)
	CreateUser(ctx context.Context, user User) (int64, error)
	UpdateUser(ctx context.Context, user User) error
	UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error
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

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token PasswordResetToken) error
	FindByHash(ctx context.Context, tokenHash string) (PasswordResetToken, error)
	Delete(ctx context.Context, id int64) error
	DeleteAllForUser(ctx context.Context, userID int64) error
}
