package notifications

import (
	"context"
	"testing"
	"time"
)

type notificationsRepoStub struct {
	getSettingsFn         func(context.Context, int64) (*Settings, error)
	clearPushTokenFn      func(context.Context, int64) error
	upsertNotificationFn  func(context.Context, Notification) (Notification, error)
	findNotificationFn    func(context.Context, int64, string) (*Notification, error)
	debtReminderSummaryFn func(context.Context, int64, time.Time) (ReminderSummary, error)
}

func (r notificationsRepoStub) GetSettings(ctx context.Context, userID int64) (*Settings, error) {
	if r.getSettingsFn != nil {
		return r.getSettingsFn(ctx, userID)
	}
	return nil, nil
}

func (r notificationsRepoStub) UpsertSettings(ctx context.Context, settings Settings) (Settings, error) {
	return settings, nil
}

func (r notificationsRepoStub) ClearPushToken(ctx context.Context, userID int64) error {
	if r.clearPushTokenFn != nil {
		return r.clearPushTokenFn(ctx, userID)
	}
	return nil
}

func (r notificationsRepoStub) ListUserIDs(ctx context.Context) ([]int64, error) {
	return []int64{}, nil
}

func (r notificationsRepoStub) ListNotifications(ctx context.Context, userID int64, filter NotificationFilter) ([]Notification, error) {
	return []Notification{}, nil
}

func (r notificationsRepoStub) GetNotificationByID(ctx context.Context, userID, id int64) (Notification, error) {
	return Notification{}, nil
}

func (r notificationsRepoStub) MarkNotificationRead(ctx context.Context, userID, id int64, readAt time.Time) (Notification, error) {
	return Notification{}, nil
}

func (r notificationsRepoStub) UpsertNotification(ctx context.Context, item Notification) (Notification, error) {
	if r.upsertNotificationFn != nil {
		return r.upsertNotificationFn(ctx, item)
	}
	return item, nil
}

func (r notificationsRepoStub) FindNotificationByDedupeKey(ctx context.Context, userID int64, dedupeKey string) (*Notification, error) {
	if r.findNotificationFn != nil {
		return r.findNotificationFn(ctx, userID, dedupeKey)
	}
	return nil, nil
}

func (r notificationsRepoStub) DebtReminderSummary(ctx context.Context, userID int64, cutoff time.Time) (ReminderSummary, error) {
	if r.debtReminderSummaryFn != nil {
		return r.debtReminderSummaryFn(ctx, userID, cutoff)
	}
	return ReminderSummary{}, nil
}

func TestGenerateIncludesSalaryReminder(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 29, 8, 5, 0, 0, time.UTC)
	}
	defer func() { nowFunc = originalNowFunc }()

	svc := NewService(notificationsRepoStub{
		getSettingsFn: func(context.Context, int64) (*Settings, error) {
			return &Settings{
				UserID:                      2,
				Enabled:                     true,
				DailyExpenseReminderEnabled: false,
				DebtPaymentReminderEnabled:  false,
				SalaryReminderEnabled:       true,
				SalaryReminderTime:          "08:00",
				SalaryReminderDaysBefore:    1,
				SalaryDay:                   31,
			}, nil
		},
		findNotificationFn: func(context.Context, int64, string) (*Notification, error) {
			return nil, nil
		},
		upsertNotificationFn: func(_ context.Context, item Notification) (Notification, error) {
			return item, nil
		},
	}, nil)

	items, err := svc.Generate(context.Background(), 2)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.Kind != ReminderKindSalary {
		t.Fatalf("unexpected kind: %s", item.Kind)
	}
	if item.Type != string(ReminderKindSalary) {
		t.Fatalf("unexpected type: %s", item.Type)
	}
	if item.Data["route"] != "/transactions?type=income" {
		t.Fatalf("unexpected route: %v", item.Data)
	}
	if item.DeliveryStatus != DeliveryStatusSkipped {
		t.Fatalf("unexpected delivery status: %s", item.DeliveryStatus)
	}
	if item.DedupeKey != "salary-reminder:2026-04:1" {
		t.Fatalf("unexpected dedupe key: %s", item.DedupeKey)
	}
	if item.Message != "Jangan lupa catat pemasukan gaji tanggal 2026-04-30." {
		t.Fatalf("unexpected message: %s", item.Message)
	}
	if item.ScheduledFor.Format("2006-01-02 15:04") != "2026-04-29 08:00" {
		t.Fatalf("unexpected scheduled_for: %s", item.ScheduledFor.Format("2006-01-02 15:04"))
	}
}

func TestUpdateSettingsRejectsInvalidSalaryDay(t *testing.T) {
	svc := NewService(notificationsRepoStub{
		getSettingsFn: func(context.Context, int64) (*Settings, error) {
			return &Settings{UserID: 2, SalaryDay: 25}, nil
		},
	}, nil)

	invalid := 0
	_, err := svc.UpdateSettings(context.Background(), 2, UpdateSettingsInput{SalaryDay: &invalid})
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
