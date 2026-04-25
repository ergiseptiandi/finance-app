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
	nowFunc         = time.Now
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
	if input.SalaryReminderEnabled != nil {
		updated.SalaryReminderEnabled = *input.SalaryReminderEnabled
	}
	if input.SalaryReminderTime != nil {
		updated.SalaryReminderTime = normalizeClock(*input.SalaryReminderTime, updated.SalaryReminderTime)
	}
	if input.SalaryReminderDaysBefore != nil {
		if *input.SalaryReminderDaysBefore < 0 {
			return Settings{}, ErrInvalidInput
		}
		updated.SalaryReminderDaysBefore = *input.SalaryReminderDaysBefore
	}
	if input.SalaryDay != nil {
		if *input.SalaryDay < 1 || *input.SalaryDay > 31 {
			return Settings{}, ErrInvalidInput
		}
		updated.SalaryDay = *input.SalaryDay
	}
	if input.BudgetAmount != nil {
		if *input.BudgetAmount < 0 {
			return Settings{}, ErrInvalidInput
		}
		updated.BudgetAmount = *input.BudgetAmount
	}
	if input.BudgetWarningEnabled != nil {
		updated.BudgetWarningEnabled = *input.BudgetWarningEnabled
	}
	if input.BudgetWarningThreshold != nil {
		if *input.BudgetWarningThreshold < 1 || *input.BudgetWarningThreshold > 100 {
			return Settings{}, ErrInvalidInput
		}
		updated.BudgetWarningThreshold = *input.BudgetWarningThreshold
	}
	if input.WeeklySummaryEnabled != nil {
		updated.WeeklySummaryEnabled = *input.WeeklySummaryEnabled
	}
	if input.WeeklySummaryDay != nil {
		if *input.WeeklySummaryDay < 0 || *input.WeeklySummaryDay > 6 {
			return Settings{}, ErrInvalidInput
		}
		updated.WeeklySummaryDay = *input.WeeklySummaryDay
	}
	if input.LargeTransactionEnabled != nil {
		updated.LargeTransactionEnabled = *input.LargeTransactionEnabled
	}
	if input.LargeTransactionThreshold != nil {
		if *input.LargeTransactionThreshold < 0 {
			return Settings{}, ErrInvalidInput
		}
		updated.LargeTransactionThreshold = *input.LargeTransactionThreshold
	}
	if input.GoalReminderEnabled != nil {
		updated.GoalReminderEnabled = *input.GoalReminderEnabled
	}
	if input.GoalReminderDaysBefore != nil {
		if *input.GoalReminderDaysBefore < 0 {
			return Settings{}, ErrInvalidInput
		}
		updated.GoalReminderDaysBefore = *input.GoalReminderDaysBefore
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

	now := nowFunc()
	items := make([]Notification, 0, 8)

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

	if settings.SalaryReminderEnabled {
		if item, err := s.generateSalaryReminder(ctx, userID, settings, now); err != nil {
			return nil, err
		} else if item != nil {
			items = append(items, *item)
		}
	}

	if settings.WeeklySummaryEnabled {
		if item, err := s.generateWeeklySummary(ctx, userID, settings, now); err != nil {
			return nil, err
		} else if item != nil {
			items = append(items, *item)
		}
	}

	if settings.GoalReminderEnabled {
		if item, err := s.generateGoalReminder(ctx, userID, settings, now); err != nil {
			return nil, err
		} else if item != nil {
			items = append(items, *item)
		}
	}

	return items, nil
}

func (s *Service) MarkRead(ctx context.Context, userID, id int64) (Notification, error) {
	return s.repo.MarkNotificationRead(ctx, userID, id, nowFunc())
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
		Title:          "Pengingat Pengeluaran Harian",
		Message:        "Yuk catat pengeluaran hari ini! Mencatat keuangan secara rutin membantu kamu mengelola uang lebih baik.",
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

	message := fmt.Sprintf("Kamu memiliki %d tagihan yang perlu dibayar segera.", summary.Count)
	if summary.NextDueAt != nil {
		message = fmt.Sprintf("%s Tagihan terdekat jatuh tempo pada %s.", message, summary.NextDueAt.Format("02 January 2006"))
	}

	item := Notification{
		UserID:         userID,
		Kind:           ReminderKindDebtPayment,
		Title:          "Pengingat Pembayaran Utang",
		Message:        message,
		Data:           notificationData(ReminderKindDebtPayment),
		DeliveryStatus: DeliveryStatusPending,
		ScheduledFor:   scheduledFor,
		DedupeKey:      dedupeKey,
	}

	return s.storeAndPush(ctx, settings, item)
}

func (s *Service) generateSalaryReminder(ctx context.Context, userID int64, settings Settings, now time.Time) (*Notification, error) {
	nextSalaryDate := nextSalaryDate(now, settings.SalaryDay)
	reminderDate := nextSalaryDate.AddDate(0, 0, -settings.SalaryReminderDaysBefore)
	if now.Before(combineDateAndClock(reminderDate, settings.SalaryReminderTime)) {
		return nil, nil
	}

	dedupeKey := fmt.Sprintf("salary-reminder:%s:%d", nextSalaryDate.Format("2006-01"), settings.SalaryReminderDaysBefore)
	existing, err := s.repo.FindNotificationByDedupeKey(ctx, userID, dedupeKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, nil
	}

	item := Notification{
		UserID:         userID,
		Kind:           ReminderKindSalary,
		Title:          "Pengingat Gaji",
		Message:        fmt.Sprintf("Gaji kamu akan cair tanggal %s. Jangan lupa catat pemasukan ya!", nextSalaryDate.Format("02 January 2006")),
		Data:           notificationData(ReminderKindSalary),
		DeliveryStatus: DeliveryStatusPending,
		ScheduledFor:   combineDateAndClock(reminderDate, settings.SalaryReminderTime),
		DedupeKey:      dedupeKey,
	}

	return s.storeAndPush(ctx, settings, item)
}

func (s *Service) generateWeeklySummary(ctx context.Context, userID int64, settings Settings, now time.Time) (*Notification, error) {
	// Check if today is the configured summary day
	if int(now.Weekday()) != settings.WeeklySummaryDay {
		return nil, nil
	}

	dedupeKey := fmt.Sprintf("weekly-summary:%s", now.Format("2006-01-02"))
	existing, err := s.repo.FindNotificationByDedupeKey(ctx, userID, dedupeKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, nil
	}

	// Get weekly summary data from repository
	summary, err := s.repo.WeeklySummary(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("Ringkasan minggu ini: Total pengeluaran Rp %.2f dan pemasukan Rp %.2f.", summary.TotalExpense, summary.TotalIncome)
	if summary.NetSavings > 0 {
		message = fmt.Sprintf("%s Tabungan bersih: Rp %.2f. Keren!", message, summary.NetSavings)
	} else if summary.NetSavings < 0 {
		message = fmt.Sprintf("%s Pengeluaran melebihi pemasukan sebesar Rp %.2f. Ayo lebih hemat minggu depan!", message, -summary.NetSavings)
	}

	item := Notification{
		UserID:         userID,
		Kind:           ReminderKindWeeklySummary,
		Title:          "Ringkasan Keuangan Mingguan",
		Message:        message,
		Data:           notificationData(ReminderKindWeeklySummary),
		DeliveryStatus: DeliveryStatusPending,
		ScheduledFor:   now,
		DedupeKey:      dedupeKey,
	}

	return s.storeAndPush(ctx, settings, item)
}

func (s *Service) generateGoalReminder(ctx context.Context, userID int64, settings Settings, now time.Time) (*Notification, error) {
	// Get upcoming goal deadlines
	goals, err := s.repo.UpcomingGoals(ctx, userID, settings.GoalReminderDaysBefore)
	if err != nil {
		return nil, err
	}

	if len(goals) == 0 {
		return nil, nil
	}

	items := make([]Notification, 0, len(goals))
	for _, goal := range goals {
		dedupeKey := fmt.Sprintf("goal-reminder:%d:%s", goal.ID, now.Format("2006-01-02"))
		existing, err := s.repo.FindNotificationByDedupeKey(ctx, userID, dedupeKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			continue
		}

		daysLeft := int(goal.Deadline.Sub(startOfDay(now)).Hours() / 24)
		message := fmt.Sprintf("Target tabungan '%s' akan jatuh tempo dalam %d hari.", goal.Name, daysLeft)
		progress := float64(0)
		if goal.TargetAmount > 0 {
			progress = (goal.CurrentAmount / goal.TargetAmount) * 100
		}
		message = fmt.Sprintf("%s Progress: %.0f%% (Rp %.2f / Rp %.2f).", message, progress, goal.CurrentAmount, goal.TargetAmount)

		item := Notification{
			UserID:         userID,
			Kind:           ReminderKindGoalReminder,
			Title:          "Pengingat Target Tabungan",
			Message:        message,
			Data:           notificationData(ReminderKindGoalReminder),
			DeliveryStatus: DeliveryStatusPending,
			ScheduledFor:   now,
			DedupeKey:      dedupeKey,
		}

		stored, err := s.storeAndPush(ctx, settings, item)
		if err != nil {
			return nil, err
		}
		items = append(items, *stored)
	}

	if len(items) == 0 {
		return nil, nil
	}

	// Return the first item (for compatibility with existing code)
	return &items[0], nil
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

		// If the token is invalid/expired, clear it so we don't keep
		// sending to a dead token. The mobile app will sync a fresh
		// token next time it opens.
		if IsTokenInvalid(err) {
			_ = s.repo.ClearPushToken(ctx, settings.UserID)
		}
	} else {
		sentAt := nowFunc()
		item.DeliveryStatus = DeliveryStatusSent
		item.SentAt = &sentAt
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
	if _, err := parseClock(item.SalaryReminderTime); err != nil {
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
		SalaryReminderEnabled:         true,
		SalaryReminderTime:            "08:00",
		SalaryReminderDaysBefore:      1,
		SalaryDay:                     25,
		BudgetAmount:                  0,
		BudgetWarningEnabled:          true,
		BudgetWarningThreshold:        80,
		WeeklySummaryEnabled:          true,
		WeeklySummaryDay:              0, // Sunday
		LargeTransactionEnabled:       true,
		LargeTransactionThreshold:     1000000, // Rp 1,000,000
		GoalReminderEnabled:           true,
		GoalReminderDaysBefore:        7,
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
	case ReminderKindSalary:
		return map[string]string{
			"kind":  string(kind),
			"type":  string(kind),
			"route": "/transactions?type=income",
		}
	case ReminderKindBudgetWarning:
		return map[string]string{
			"kind":  string(kind),
			"type":  string(kind),
			"route": "/reports",
		}
	case ReminderKindWeeklySummary:
		return map[string]string{
			"kind":  string(kind),
			"type":  string(kind),
			"route": "/reports",
		}
	case ReminderKindLargeTransaction:
		return map[string]string{
			"kind":  string(kind),
			"type":  string(kind),
			"route": "/activity",
		}
	case ReminderKindGoalReminder:
		return map[string]string{
			"kind":  string(kind),
			"type":  string(kind),
			"route": "/reports",
		}
	default:
		return map[string]string{
			"kind": string(kind),
			"type": string(kind),
		}
	}
}

func nextSalaryDate(now time.Time, salaryDay int) time.Time {
	if salaryDay < 1 {
		salaryDay = 1
	}

	current := salaryDateInMonth(now, salaryDay)
	if now.Before(current) {
		return current
	}

	return salaryDateInMonth(now.AddDate(0, 1, 0), salaryDay)
}

func salaryDateInMonth(ref time.Time, salaryDay int) time.Time {
	y, m, _ := ref.Date()
	lastDay := lastDayOfMonth(y, m, ref.Location())
	if salaryDay > lastDay {
		salaryDay = lastDay
	}
	return time.Date(y, m, salaryDay, 0, 0, 0, 0, ref.Location())
}

func lastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
	return firstOfNextMonth.AddDate(0, 0, -1).Day()
}

func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
