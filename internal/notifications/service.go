package notifications

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("notification not found")
	ErrInvalidInput = errors.New("invalid notification input")
)

type Service struct {
	repo       Repository
	pushSender PushSender
}

func NewService(repo Repository, pushSender PushSender) *Service {
	if pushSender == nil {
		pushSender = NoopPushSender{}
	}

	return &Service{repo: repo, pushSender: pushSender}
}

func (s *Service) GetSettings(ctx context.Context, userID int64) (Settings, error) {
	item, err := s.repo.GetSettings(ctx, userID)
	if err != nil {
		return Settings{}, err
	}
	if item == nil {
		return defaultSettings(userID), nil
	}
	return *item, nil
}

func (s *Service) UpdateSettings(ctx context.Context, userID int64, input UpdateSettingsInput) (Settings, error) {
	current, err := s.GetSettings(ctx, userID)
	if err != nil {
		return Settings{}, err
	}

	updated := current
	updated.UserID = userID

	if input.Enabled != nil {
		updated.Enabled = *input.Enabled
	}
	if input.DailyExpenseReminderEnabled != nil {
		updated.DailyExpenseReminderEnabled = *input.DailyExpenseReminderEnabled
	}
	if input.DailyExpenseReminderTime != nil {
		updated.DailyExpenseReminderTime = normalizeClock(*input.DailyExpenseReminderTime, updated.DailyExpenseReminderTime)
	}
	if input.DebtPaymentReminderEnabled != nil {
		updated.DebtPaymentReminderEnabled = *input.DebtPaymentReminderEnabled
	}
	if input.DebtPaymentReminderTime != nil {
		updated.DebtPaymentReminderTime = normalizeClock(*input.DebtPaymentReminderTime, updated.DebtPaymentReminderTime)
	}
	if input.DebtPaymentReminderDaysBefore != nil {
		if *input.DebtPaymentReminderDaysBefore < 0 {
			return Settings{}, ErrInvalidInput
		}
		updated.DebtPaymentReminderDaysBefore = *input.DebtPaymentReminderDaysBefore
	}
	if input.PushToken != nil {
		updated.PushToken = strings.TrimSpace(*input.PushToken)
	}

	if err := validateSettings(updated); err != nil {
		return Settings{}, err
	}

	return s.repo.UpsertSettings(ctx, updated)
}

func (s *Service) List(ctx context.Context, userID int64, filter NotificationFilter) ([]Notification, error) {
	return s.repo.ListNotifications(ctx, userID, filter)
}

func (s *Service) Generate(ctx context.Context, userID int64) ([]Notification, error) {
	settings, err := s.GetSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return []Notification{}, nil
	}

	now := time.Now()
	items := make([]Notification, 0, 3)

	if settings.DailyExpenseReminderEnabled {
		if item, err := s.generateDailyExpenseReminder(ctx, userID, settings, now); err != nil {
			return nil, err
		} else if item != nil {
			items = append(items, *item)
		}
	}

	if settings.DebtPaymentReminderEnabled {
		if item, err := s.generateDebtReminder(ctx, userID, settings, now); err != nil {
			return nil, err
		} else if item != nil {
			items = append(items, *item)
		}
	}

	return items, nil
}

func (s *Service) MarkRead(ctx context.Context, userID, id int64) (Notification, error) {
	return s.repo.MarkNotificationRead(ctx, userID, id, time.Now())
}

func (s *Service) generateDailyExpenseReminder(ctx context.Context, userID int64, settings Settings, now time.Time) (*Notification, error) {
	dedupeKey := fmt.Sprintf("daily-expense:%s", now.Format("2006-01-02"))
	existing, err := s.repo.FindNotificationByDedupeKey(ctx, userID, dedupeKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, nil
	}

	scheduledFor := combineDateAndClock(now, settings.DailyExpenseReminderTime)
	if now.Before(scheduledFor) {
		return nil, nil
	}

	item := Notification{
		UserID:         userID,
		Kind:           ReminderKindDailyExpense,
		Title:          "Daily expense reminder",
		Message:        "Jangan lupa input pengeluaran hari ini.",
		Data:           notificationData(ReminderKindDailyExpense),
		DeliveryStatus: DeliveryStatusPending,
		ScheduledFor:   scheduledFor,
		DedupeKey:      dedupeKey,
	}

	return s.storeAndPush(ctx, settings, item)
}

func (s *Service) generateDebtReminder(ctx context.Context, userID int64, settings Settings, now time.Time) (*Notification, error) {
	cutoff := startOfDay(now).AddDate(0, 0, settings.DebtPaymentReminderDaysBefore+1)
	summary, err := s.repo.DebtReminderSummary(ctx, userID, cutoff)
	if err != nil {
		return nil, err
	}
	if summary.Count == 0 {
		return nil, nil
	}

	dedupeKey := fmt.Sprintf("debt-reminder:%s:%d", now.Format("2006-01-02"), settings.DebtPaymentReminderDaysBefore)
	existing, err := s.repo.FindNotificationByDedupeKey(ctx, userID, dedupeKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, nil
	}

	scheduledFor := combineDateAndClock(now, settings.DebtPaymentReminderTime)
	if now.Before(scheduledFor) {
		return nil, nil
	}

	message := fmt.Sprintf("Ada %d tagihan debt dengan total %.2f yang perlu dibayar.", summary.Count, summary.Amount)
	if summary.NextDueAt != nil {
		message = fmt.Sprintf("%s Jatuh tempo terdekat: %s.", message, summary.NextDueAt.Format("2006-01-02"))
	}

	item := Notification{
		UserID:         userID,
		Kind:           ReminderKindDebtPayment,
		Title:          "Debt payment reminder",
		Message:        message,
		Data:           notificationData(ReminderKindDebtPayment),
		DeliveryStatus: DeliveryStatusPending,
		ScheduledFor:   scheduledFor,
		DedupeKey:      dedupeKey,
	}

	return s.storeAndPush(ctx, settings, item)
}

func (s *Service) storeAndPush(ctx context.Context, settings Settings, item Notification) (*Notification, error) {
	item = normalizeNotification(item)
	token := strings.TrimSpace(settings.PushToken)
	if token == "" {
		item.DeliveryStatus = DeliveryStatusSkipped
		stored, err := s.repo.UpsertNotification(ctx, item)
		if err != nil {
			return nil, err
		}
		return &stored, nil
	}

	if err := s.pushSender.Send(ctx, token, PushMessage{
		Title: item.Title,
		Body:  item.Message,
		Data:  item.Data,
	}); err != nil {
		item.DeliveryStatus = DeliveryStatusFailed
	} else {
		item.DeliveryStatus = DeliveryStatusSent
	}

	stored, err := s.repo.UpsertNotification(ctx, item)
	if err != nil {
		return nil, err
	}

	return &stored, nil
}

func validateSettings(item Settings) error {
	if _, err := parseClock(item.DailyExpenseReminderTime); err != nil {
		return ErrInvalidInput
	}
	if _, err := parseClock(item.DebtPaymentReminderTime); err != nil {
		return ErrInvalidInput
	}
	return nil
}

func defaultSettings(userID int64) Settings {
	return Settings{
		UserID:                        userID,
		Enabled:                       true,
		DailyExpenseReminderEnabled:   true,
		DailyExpenseReminderTime:      "20:00",
		DebtPaymentReminderEnabled:    true,
		DebtPaymentReminderTime:       "09:00",
		DebtPaymentReminderDaysBefore: 3,
	}
}

func normalizeClock(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func parseClock(value string) (time.Duration, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid clock")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, fmt.Errorf("invalid clock")
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("invalid clock")
	}

	return time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute, nil
}

func combineDateAndClock(date time.Time, clock string) time.Time {
	dur, err := parseClock(clock)
	if err != nil {
		return date
	}
	return startOfDay(date).Add(dur)
}

func normalizeNotification(item Notification) Notification {
	if item.Type == "" {
		item.Type = string(item.Kind)
	}
	if item.Data == nil {
		item.Data = notificationData(item.Kind)
	}
	item.Read = item.ReadAt != nil
	return item
}

func notificationData(kind ReminderKind) map[string]string {
	switch kind {
	case ReminderKindDailyExpense:
		return map[string]string{
			"kind":  string(kind),
			"type":  string(kind),
			"route": "/activity",
		}
	case ReminderKindDebtPayment:
		return map[string]string{
			"kind":  string(kind),
			"type":  string(kind),
			"route": "/debts",
		}
	default:
		return map[string]string{
			"kind": string(kind),
			"type": string(kind),
		}
	}
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
