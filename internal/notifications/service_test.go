package notifications

import (
	"context"
	"errors"
	"testing"
	"time"
)

type notificationsRepoStub struct {
	getSettingsFn         func(context.Context, int64) (*Settings, error)
	upsertSettingsFn      func(context.Context, Settings) (Settings, error)
	clearPushTokenFn      func(context.Context, int64) error
	upsertNotificationFn  func(context.Context, Notification) (Notification, error)
	findNotificationFn    func(context.Context, int64, string) (*Notification, error)
	debtReminderSummaryFn func(context.Context, int64, time.Time) (ReminderSummary, error)
}

type pushSenderStub struct {
	sendFn func(context.Context, string, PushMessage) error
}

func (p pushSenderStub) Send(ctx context.Context, token string, message PushMessage) error {
	if p.sendFn != nil {
		return p.sendFn(ctx, token, message)
	}
	return nil
}

func (r notificationsRepoStub) GetSettings(ctx context.Context, userID int64) (*Settings, error) {
	if r.getSettingsFn != nil {
		return r.getSettingsFn(ctx, userID)
	}
	return nil, nil
}

func (r notificationsRepoStub) UpsertSettings(ctx context.Context, settings Settings) (Settings, error) {
	if r.upsertSettingsFn != nil {
		return r.upsertSettingsFn(ctx, settings)
	}
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

func (r notificationsRepoStub) WeeklySummary(ctx context.Context, userID int64, now time.Time) (WeeklySummaryData, error) {
	return WeeklySummaryData{}, nil
}

func (r notificationsRepoStub) UpcomingGoals(ctx context.Context, userID int64, daysBefore int) ([]GoalData, error) {
	return []GoalData{}, nil
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

func TestUpdateSettingsRefreshesPushTokenForExistingUser(t *testing.T) {
	newToken := "new-device-token"
	var stored Settings

	svc := NewService(notificationsRepoStub{
		getSettingsFn: func(context.Context, int64) (*Settings, error) {
			return &Settings{
				UserID:                      2,
				Enabled:                     true,
				DailyExpenseReminderEnabled: true,
				DailyExpenseReminderTime:    "20:00",
				DebtPaymentReminderEnabled:  true,
				DebtPaymentReminderTime:     "09:00",
				SalaryReminderEnabled:       true,
				SalaryReminderTime:          "08:00",
				SalaryReminderDaysBefore:    1,
				SalaryDay:                   25,
				PushToken:                   "old-device-token",
			}, nil
		},
		upsertSettingsFn: func(_ context.Context, settings Settings) (Settings, error) {
			stored = settings
			return settings, nil
		},
	}, nil)

	item, err := svc.UpdateSettings(context.Background(), 2, UpdateSettingsInput{
		PushToken: &newToken,
	})
	if err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}

	if stored.UserID != 2 {
		t.Fatalf("unexpected user id: %d", stored.UserID)
	}
	if stored.PushToken != newToken {
		t.Fatalf("expected push token %q, got %q", newToken, stored.PushToken)
	}
	if stored.DailyExpenseReminderTime != "20:00" {
		t.Fatalf("existing settings should be preserved, got %q", stored.DailyExpenseReminderTime)
	}
	if item.PushToken != newToken {
		t.Fatalf("unexpected returned token: %q", item.PushToken)
	}
}

func TestUpdateSettingsCreatesDefaultsWhenOnlyPushTokenProvided(t *testing.T) {
	newToken := "fresh-install-token"
	var stored Settings

	svc := NewService(notificationsRepoStub{
		getSettingsFn: func(context.Context, int64) (*Settings, error) {
			return nil, nil
		},
		upsertSettingsFn: func(_ context.Context, settings Settings) (Settings, error) {
			stored = settings
			return settings, nil
		},
	}, nil)

	item, err := svc.UpdateSettings(context.Background(), 7, UpdateSettingsInput{
		PushToken: &newToken,
	})
	if err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}

	if stored.UserID != 7 {
		t.Fatalf("unexpected user id: %d", stored.UserID)
	}
	if stored.PushToken != newToken {
		t.Fatalf("expected push token %q, got %q", newToken, stored.PushToken)
	}
	if !stored.Enabled {
		t.Fatal("expected default enabled=true")
	}
	if stored.DailyExpenseReminderTime != "20:00" {
		t.Fatalf("unexpected default daily reminder time: %q", stored.DailyExpenseReminderTime)
	}
	if stored.SalaryDay != 25 {
		t.Fatalf("unexpected default salary day: %d", stored.SalaryDay)
	}
	if item.PushToken != newToken {
		t.Fatalf("unexpected returned token: %q", item.PushToken)
	}
}

func TestGenerateMarksSuccessfulPushAsSent(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 21, 9, 30, 0, 0, time.UTC)
	}
	defer func() { nowFunc = originalNowFunc }()

	var stored Notification

	svc := NewService(notificationsRepoStub{
		getSettingsFn: func(context.Context, int64) (*Settings, error) {
			return &Settings{
				UserID:                      2,
				Enabled:                     true,
				DailyExpenseReminderEnabled: true,
				DailyExpenseReminderTime:    "09:00",
				DebtPaymentReminderEnabled:  false,
				SalaryReminderEnabled:       false,
				PushToken:                   "valid-device-token",
			}, nil
		},
		findNotificationFn: func(context.Context, int64, string) (*Notification, error) {
			return nil, nil
		},
		upsertNotificationFn: func(_ context.Context, item Notification) (Notification, error) {
			stored = item
			return item, nil
		},
	}, pushSenderStub{
		sendFn: func(_ context.Context, token string, message PushMessage) error {
			if token != "valid-device-token" {
				t.Fatalf("unexpected token: %s", token)
			}
			if message.Title != "Daily expense reminder" {
				t.Fatalf("unexpected title: %s", message.Title)
			}
			return nil
		},
	})

	items, err := svc.Generate(context.Background(), 2)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if stored.DeliveryStatus != DeliveryStatusSent {
		t.Fatalf("expected delivery status %s, got %s", DeliveryStatusSent, stored.DeliveryStatus)
	}
	if stored.SentAt == nil {
		t.Fatal("expected sent_at to be set")
	}
	if !stored.SentAt.Equal(nowFunc()) {
		t.Fatalf("unexpected sent_at: %s", stored.SentAt)
	}
}

func TestGenerateMarksFailedPushAsFailed(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 21, 9, 30, 0, 0, time.UTC)
	}
	defer func() { nowFunc = originalNowFunc }()

	var stored Notification

	svc := NewService(notificationsRepoStub{
		getSettingsFn: func(context.Context, int64) (*Settings, error) {
			return &Settings{
				UserID:                      2,
				Enabled:                     true,
				DailyExpenseReminderEnabled: true,
				DailyExpenseReminderTime:    "09:00",
				DebtPaymentReminderEnabled:  false,
				SalaryReminderEnabled:       false,
				PushToken:                   "bad-device-token",
			}, nil
		},
		findNotificationFn: func(context.Context, int64, string) (*Notification, error) {
			return nil, nil
		},
		upsertNotificationFn: func(_ context.Context, item Notification) (Notification, error) {
			stored = item
			return item, nil
		},
	}, pushSenderStub{
		sendFn: func(_ context.Context, token string, message PushMessage) error {
			return errors.New("fcm send failed")
		},
	})

	items, err := svc.Generate(context.Background(), 2)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if stored.DeliveryStatus != DeliveryStatusFailed {
		t.Fatalf("expected delivery status %s, got %s", DeliveryStatusFailed, stored.DeliveryStatus)
	}
	if stored.SentAt != nil {
		t.Fatal("expected sent_at to stay nil on failed push")
	}
}
