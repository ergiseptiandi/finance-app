package ai

import (
	"context"
	"errors"
)

var ErrChatLimitExceeded = errors.New("chat limit exceeded")

type Repository interface {
	GetChatCount(ctx context.Context, userID int64) (int, error)
	IncrementChatCount(ctx context.Context, userID int64) error
}
