package notifications

import (
	"context"
	"time"
)

type WeeklySummaryData struct {
	TotalExpense float64 `json:"total_expense"`
	TotalIncome  float64 `json:"total_income"`
	NetSavings   float64 `json:"net_savings"`
}

type GoalData struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	TargetAmount   float64   `json:"target_amount"`
	CurrentAmount  float64   `json:"current_amount"`
	Deadline       time.Time `json:"deadline"`
}

type Repository interface {
	GetSettings(ctx context.Context, userID int64) (*Settings, error)
	UpsertSettings(ctx context.Context, settings Settings) (Settings, error)
	ClearPushToken(ctx context.Context, userID int64) error
	ListUserIDs(ctx context.Context) ([]int64, error)
	ListNotifications(ctx context.Context, userID int64, filter NotificationFilter) ([]Notification, error)
	GetNotificationByID(ctx context.Context, userID, id int64) (Notification, error)
	MarkNotificationRead(ctx context.Context, userID, id int64, readAt time.Time) (Notification, error)
	UpsertNotification(ctx context.Context, item Notification) (Notification, error)
	FindNotificationByDedupeKey(ctx context.Context, userID int64, dedupeKey string) (*Notification, error)
	DebtReminderSummary(ctx context.Context, userID int64, cutoff time.Time) (ReminderSummary, error)
	WeeklySummary(ctx context.Context, userID int64, now time.Time) (WeeklySummaryData, error)
	UpcomingGoals(ctx context.Context, userID int64, daysBefore int) ([]GoalData, error)
}
