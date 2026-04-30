package reports

import (
	"context"
	"time"
)

type TrendGroupBy string

const (
	TrendGroupByDay   TrendGroupBy = "day"
	TrendGroupByMonth TrendGroupBy = "month"
)

type Repository interface {
	ExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]CategoryExpense, error)
	TrendByPeriod(ctx context.Context, userID int64, start, end time.Time, groupBy TrendGroupBy) (map[string]TrendTotals, error)
	AllTimeIncome(ctx context.Context, userID int64) (float64, error)
	AllTimeExpense(ctx context.Context, userID int64) (float64, error)
	IncomeBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error)
	ExpenseBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error)
	ConsumptionExpenseBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error)
	DebtRepaymentBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error)
}
