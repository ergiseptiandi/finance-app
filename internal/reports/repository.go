package reports

import (
	"context"
	"time"
)

type Repository interface {
	ExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]CategoryExpense, error)
	ExpenseTrendByMonth(ctx context.Context, userID int64, start, end time.Time) (map[string]float64, error)
	AllTimeIncome(ctx context.Context, userID int64) (float64, error)
	AllTimeExpense(ctx context.Context, userID int64) (float64, error)
	ExpenseBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error)
}
