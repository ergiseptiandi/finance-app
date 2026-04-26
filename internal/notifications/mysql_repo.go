package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
)

type MySQLNotificationsRepository struct {
	db                 *sql.DB
	schemaMu           sync.RWMutex
	schemaChecked      bool
	hasSalaryDayColumn bool
}

func NewMySQLNotificationsRepository(db *sql.DB) *MySQLNotificationsRepository {
	return &MySQLNotificationsRepository{db: db}
}

func (r *MySQLNotificationsRepository) GetSettings(ctx context.Context, userID int64) (*Settings, error) {
	query := `
		SELECT user_id, enabled, daily_expense_reminder_enabled, daily_expense_reminder_time,
		       debt_payment_reminder_enabled, debt_payment_reminder_time, debt_payment_reminder_days_before,
		       salary_reminder_enabled, salary_reminder_time, salary_reminder_days_before, salary_day,
		       budget_amount, budget_warning_enabled, budget_warning_threshold,
		       weekly_summary_enabled, weekly_summary_day,
		       large_transaction_enabled, large_transaction_threshold,
		       goal_reminder_enabled, goal_reminder_days_before,
		       push_token, created_at, updated_at
		FROM notification_settings
		WHERE user_id = ?
		LIMIT 1
	`

	var item Settings
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&item.UserID,
		&item.Enabled,
		&item.DailyExpenseReminderEnabled,
		&item.DailyExpenseReminderTime,
		&item.DebtPaymentReminderEnabled,
		&item.DebtPaymentReminderTime,
		&item.DebtPaymentReminderDaysBefore,
		&item.SalaryReminderEnabled,
		&item.SalaryReminderTime,
		&item.SalaryReminderDaysBefore,
		&item.SalaryDay,
		&item.BudgetAmount,
		&item.BudgetWarningEnabled,
		&item.BudgetWarningThreshold,
		&item.WeeklySummaryEnabled,
		&item.WeeklySummaryDay,
		&item.LargeTransactionEnabled,
		&item.LargeTransactionThreshold,
		&item.GoalReminderEnabled,
		&item.GoalReminderDaysBefore,
		&item.PushToken,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

func (r *MySQLNotificationsRepository) UpsertSettings(ctx context.Context, settings Settings) (Settings, error) {
	args := []any{
		settings.UserID,
		settings.Enabled,
		settings.DailyExpenseReminderEnabled,
		settings.DailyExpenseReminderTime,
		settings.DebtPaymentReminderEnabled,
		settings.DebtPaymentReminderTime,
		settings.DebtPaymentReminderDaysBefore,
		settings.SalaryReminderEnabled,
		settings.SalaryReminderTime,
		settings.SalaryReminderDaysBefore,
		settings.SalaryDay,
		settings.BudgetAmount,
		settings.BudgetWarningEnabled,
		settings.BudgetWarningThreshold,
		settings.WeeklySummaryEnabled,
		settings.WeeklySummaryDay,
		settings.LargeTransactionEnabled,
		settings.LargeTransactionThreshold,
		settings.GoalReminderEnabled,
		settings.GoalReminderDaysBefore,
		settings.PushToken,
	}

	const query = `
		INSERT INTO notification_settings (
			user_id, enabled, daily_expense_reminder_enabled, daily_expense_reminder_time,
			debt_payment_reminder_enabled, debt_payment_reminder_time, debt_payment_reminder_days_before,
			salary_reminder_enabled, salary_reminder_time, salary_reminder_days_before, salary_day,
			budget_amount, budget_warning_enabled, budget_warning_threshold,
			weekly_summary_enabled, weekly_summary_day,
			large_transaction_enabled, large_transaction_threshold,
			goal_reminder_enabled, goal_reminder_days_before, push_token
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			enabled = VALUES(enabled),
			daily_expense_reminder_enabled = VALUES(daily_expense_reminder_enabled),
			daily_expense_reminder_time = VALUES(daily_expense_reminder_time),
			debt_payment_reminder_enabled = VALUES(debt_payment_reminder_enabled),
			debt_payment_reminder_time = VALUES(debt_payment_reminder_time),
			debt_payment_reminder_days_before = VALUES(debt_payment_reminder_days_before),
			salary_reminder_enabled = VALUES(salary_reminder_enabled),
			salary_reminder_time = VALUES(salary_reminder_time),
			salary_reminder_days_before = VALUES(salary_reminder_days_before),
			salary_day = VALUES(salary_day),
			budget_amount = VALUES(budget_amount),
			budget_warning_enabled = VALUES(budget_warning_enabled),
			budget_warning_threshold = VALUES(budget_warning_threshold),
			weekly_summary_enabled = VALUES(weekly_summary_enabled),
			weekly_summary_day = VALUES(weekly_summary_day),
			large_transaction_enabled = VALUES(large_transaction_enabled),
			large_transaction_threshold = VALUES(large_transaction_threshold),
			goal_reminder_enabled = VALUES(goal_reminder_enabled),
			goal_reminder_days_before = VALUES(goal_reminder_days_before),
			push_token = VALUES(push_token)
	`

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return Settings{}, err
	}

	stored, err := r.GetSettings(ctx, settings.UserID)
	if err != nil {
		return Settings{}, err
	}
	if stored == nil {
		return Settings{}, ErrNotFound
	}

	return *stored, nil
}

func (r *MySQLNotificationsRepository) ClearPushToken(ctx context.Context, userID int64) error {
	const query = `UPDATE notification_settings SET push_token = '' WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *MySQLNotificationsRepository) ListUserIDs(ctx context.Context) ([]int64, error) {
	const query = `
		SELECT id
		FROM users
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLNotificationsRepository) ListNotifications(ctx context.Context, userID int64, filter NotificationFilter) ([]Notification, error) {
	builder := strings.Builder{}
	builder.WriteString(`
		SELECT id, user_id, kind, title, message, delivery_status, scheduled_for, sent_at, read_at, dedupe_key, data, created_at, updated_at
		FROM notifications
		WHERE user_id = ?
	`)
	args := []any{userID}

	if filter.Kind != nil && strings.TrimSpace(*filter.Kind) != "" {
		builder.WriteString(" AND kind = ?")
		args = append(args, strings.TrimSpace(*filter.Kind))
	}
	if filter.Read != nil {
		if *filter.Read {
			builder.WriteString(" AND read_at IS NOT NULL")
		} else {
			builder.WriteString(" AND read_at IS NULL")
		}
	}
	builder.WriteString(" ORDER BY created_at DESC, id DESC")

	rows, err := r.db.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Notification, 0)
	for rows.Next() {
		item, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLNotificationsRepository) GetNotificationByID(ctx context.Context, userID, id int64) (Notification, error) {
	const query = `
		SELECT id, user_id, kind, title, message, delivery_status, scheduled_for, sent_at, read_at, dedupe_key, data, created_at, updated_at
		FROM notifications
		WHERE id = ? AND user_id = ?
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id, userID)
	return scanNotificationRow(row)
}

func (r *MySQLNotificationsRepository) MarkNotificationRead(ctx context.Context, userID, id int64, readAt time.Time) (Notification, error) {
	const query = `
		UPDATE notifications
		SET read_at = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, readAt, id, userID)
	if err != nil {
		return Notification{}, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return Notification{}, err
	}
	if rows == 0 {
		return Notification{}, ErrNotFound
	}

	return r.GetNotificationByID(ctx, userID, id)
}

func (r *MySQLNotificationsRepository) UpsertNotification(ctx context.Context, item Notification) (Notification, error) {
	const query = `
		INSERT INTO notifications (
			user_id, kind, title, message, delivery_status, scheduled_for, sent_at, read_at, dedupe_key, data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			kind = VALUES(kind),
			title = VALUES(title),
			message = VALUES(message),
			delivery_status = VALUES(delivery_status),
			scheduled_for = VALUES(scheduled_for),
			sent_at = VALUES(sent_at),
			read_at = VALUES(read_at),
			data = VALUES(data)
	`

	var sentAt any
	if item.SentAt != nil {
		sentAt = *item.SentAt
	}
	var readAt any
	if item.ReadAt != nil {
		readAt = *item.ReadAt
	}
	data, err := marshalNotificationData(item.Data)
	if err != nil {
		return Notification{}, err
	}

	if _, err := r.db.ExecContext(ctx, query,
		item.UserID,
		item.Kind,
		item.Title,
		item.Message,
		item.DeliveryStatus,
		item.ScheduledFor,
		sentAt,
		readAt,
		item.DedupeKey,
		data,
	); err != nil {
		return Notification{}, err
	}

	existing, err := r.FindNotificationByDedupeKey(ctx, item.UserID, item.DedupeKey)
	if err != nil {
		return Notification{}, err
	}
	if existing == nil {
		return Notification{}, ErrNotFound
	}

	return *existing, nil
}

func (r *MySQLNotificationsRepository) FindNotificationByDedupeKey(ctx context.Context, userID int64, dedupeKey string) (*Notification, error) {
	const query = `
		SELECT id, user_id, kind, title, message, delivery_status, scheduled_for, sent_at, read_at, dedupe_key, data, created_at, updated_at
		FROM notifications
		WHERE user_id = ? AND dedupe_key = ?
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, userID, dedupeKey)
	item, err := scanNotificationRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

func (r *MySQLNotificationsRepository) DebtReminderSummary(ctx context.Context, userID int64, cutoff time.Time) (ReminderSummary, error) {
	const query = `
		SELECT COALESCE(COUNT(i.id), 0), COALESCE(SUM(i.amount), 0), MIN(i.due_date)
		FROM debt_installments i
		JOIN debts d ON d.id = i.debt_id
		WHERE d.user_id = ?
		  AND i.status IN ('pending', 'overdue')
		  AND i.due_date <= ?
	`

	var summary ReminderSummary
	var nextDue sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, userID, cutoff).Scan(&summary.Count, &summary.Amount, &nextDue); err != nil {
		return ReminderSummary{}, err
	}
	if nextDue.Valid {
		summary.NextDueAt = &nextDue.Time
	}

	return summary, nil
}

func (r *MySQLNotificationsRepository) WeeklySummary(ctx context.Context, userID int64, now time.Time) (WeeklySummaryData, error) {
	weekStart := startOfDay(now).AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	const expenseQuery = `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = ? AND type = 'expense' AND date >= ? AND date < ?
	`
	const incomeQuery = `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = ? AND type = 'income' AND date >= ? AND date < ?
	`

	var summary WeeklySummaryData
	if err := r.db.QueryRowContext(ctx, expenseQuery, userID, weekStart, weekEnd).Scan(&summary.TotalExpense); err != nil {
		if strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "42S02") {
			return WeeklySummaryData{}, nil
		}
		return WeeklySummaryData{}, err
	}
	if err := r.db.QueryRowContext(ctx, incomeQuery, userID, weekStart, weekEnd).Scan(&summary.TotalIncome); err != nil {
		if strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "42S02") {
			return WeeklySummaryData{}, nil
		}
		return WeeklySummaryData{}, err
	}
	summary.NetSavings = summary.TotalIncome - summary.TotalExpense

	return summary, nil
}

func (r *MySQLNotificationsRepository) UpcomingGoals(ctx context.Context, userID int64, daysBefore int) ([]GoalData, error) {
	now := time.Now()
	deadline := startOfDay(now.AddDate(0, 0, daysBefore))
	const query = `
		SELECT id, name, target_amount, current_amount, deadline
		FROM savings_goals
		WHERE user_id = ?
		  AND deadline <= ?
		  AND deadline > ?
		  AND status = 'active'
		ORDER BY deadline ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, deadline, startOfDay(now))
	if err != nil {
		// Table doesn't exist yet, return empty list
		if strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "42S02") {
			return []GoalData{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var goals []GoalData
	for rows.Next() {
		var g GoalData
		if err := rows.Scan(&g.ID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Deadline); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}

	return goals, rows.Err()
}

func (r *MySQLNotificationsRepository) detectSalaryDayColumn(ctx context.Context) (bool, error) {
	r.schemaMu.RLock()
	if r.schemaChecked {
		hasColumn := r.hasSalaryDayColumn
		r.schemaMu.RUnlock()
		return hasColumn, nil
	}
	r.schemaMu.RUnlock()

	const probeQuery = `
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		  AND table_name = 'notification_settings'
		  AND column_name = 'salary_day'
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRowContext(ctx, probeQuery).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	hasColumn := err == nil
	r.schemaMu.Lock()
	r.schemaChecked = true
	r.hasSalaryDayColumn = hasColumn
	r.schemaMu.Unlock()

	return hasColumn, nil
}

func scanNotification(rows *sql.Rows) (Notification, error) {
	var item Notification
	var data sql.NullString
	if err := rows.Scan(
		&item.ID,
		&item.UserID,
		&item.Kind,
		&item.Title,
		&item.Message,
		&item.DeliveryStatus,
		&item.ScheduledFor,
		&item.SentAt,
		&item.ReadAt,
		&item.DedupeKey,
		&data,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return Notification{}, err
	}
	if err := applyNotificationData(&item, data); err != nil {
		return Notification{}, err
	}
	return normalizeNotification(item), nil
}

func scanNotificationRow(row *sql.Row) (Notification, error) {
	var item Notification
	var data sql.NullString
	if err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Kind,
		&item.Title,
		&item.Message,
		&item.DeliveryStatus,
		&item.ScheduledFor,
		&item.SentAt,
		&item.ReadAt,
		&item.DedupeKey,
		&data,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Notification{}, ErrNotFound
		}
		return Notification{}, err
	}
	if err := applyNotificationData(&item, data); err != nil {
		return Notification{}, err
	}
	return normalizeNotification(item), nil
}

func applyNotificationData(item *Notification, data sql.NullString) error {
	if !data.Valid || strings.TrimSpace(data.String) == "" {
		return nil
	}

	var parsed map[string]string
	if err := json.Unmarshal([]byte(data.String), &parsed); err != nil {
		return err
	}
	item.Data = parsed
	return nil
}

func marshalNotificationData(data map[string]string) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return string(raw), nil
}
