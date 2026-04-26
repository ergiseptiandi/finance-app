package budget

import (
	"context"
	"time"
)

type CategoryInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ExpenseItem struct {
	Category string
	Amount   float64
}

type ExpenseProvider interface {
	ExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]ExpenseItem, error)
}

type Repository interface {
	Create(ctx context.Context, userID int64, categoryID int64, monthlyAmount float64) (Goal, error)
	Update(ctx context.Context, userID int64, item Goal) (Goal, error)
	GetByID(ctx context.Context, userID, id int64) (Goal, error)
	Delete(ctx context.Context, userID, id int64) error
	List(ctx context.Context, userID int64) ([]Goal, error)
	GetCategory(ctx context.Context, userID, categoryID int64) (CategoryInfo, error)
}
