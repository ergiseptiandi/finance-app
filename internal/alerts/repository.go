package alerts

import "context"

type Repository interface {
	UpsertAlert(ctx context.Context, alert Alert) (Alert, error)
	ListAlerts(ctx context.Context, userID int64, filter AlertListFilter) ([]Alert, error)
	MarkAlertRead(ctx context.Context, userID, alertID int64, isRead bool) (Alert, error)
	CurrentMonthExpense(ctx context.Context, userID int64) (float64, error)
	TodayExpense(ctx context.Context, userID int64) (float64, error)
}
