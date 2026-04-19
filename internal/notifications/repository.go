package notifications

import (
	"context"
	"time"
)

type Repository interface {
	GetSettings(ctx context.Context, userID int64) (*Settings, error)
	UpsertSettings(ctx context.Context, settings Settings) (Settings, error)
	ListUserIDs(ctx context.Context) ([]int64, error)
	ListNotifications(ctx context.Context, userID int64, filter NotificationFilter) ([]Notification, error)
	GetNotificationByID(ctx context.Context, userID, id int64) (Notification, error)
	MarkNotificationRead(ctx context.Context, userID, id int64, readAt time.Time) (Notification, error)
	UpsertNotification(ctx context.Context, item Notification) (Notification, error)
	FindNotificationByDedupeKey(ctx context.Context, userID int64, dedupeKey string) (*Notification, error)
	DebtReminderSummary(ctx context.Context, userID int64, cutoff time.Time) (ReminderSummary, error)
}
