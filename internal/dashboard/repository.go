package dashboard

import (
	"context"
	"time"
)

type Repository interface {
	RefreshUserDebtStatuses(ctx context.Context, userID int64) error
	AllTimeIncome(ctx context.Context, userID int64) (float64, error)
	AllTimeExpense(ctx context.Context, userID int64) (float64, error)
	IncomeBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error)
	ExpenseBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error)
	ExpenseByDay(ctx context.Context, userID int64, start, end time.Time) (map[string]float64, error)
	ExpenseByMonth(ctx context.Context, userID int64, start, end time.Time) (map[string]float64, error)
	DebtOverview(ctx context.Context, userID int64, start, end time.Time) (DebtOverview, error)
	ExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]CategoryBreakdownItem, error)
	UpcomingBills(ctx context.Context, userID int64, start, end time.Time) ([]UpcomingBill, error)
	TopMerchants(ctx context.Context, userID int64, start, end time.Time, limit int) ([]TopMerchant, error)
}
