package alerts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type MySQLAlertsRepository struct {
	db *sql.DB
}

func NewMySQLAlertsRepository(db *sql.DB) *MySQLAlertsRepository {
	return &MySQLAlertsRepository{db: db}
}

func (r *MySQLAlertsRepository) UpsertAlert(ctx context.Context, alert Alert) (Alert, error) {
	const query = `
		INSERT INTO alerts (
			user_id, type, title, message, severity, metric_value, threshold_value, dedupe_key, is_read
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, FALSE)
		ON DUPLICATE KEY UPDATE
			type = VALUES(type),
			title = VALUES(title),
			message = VALUES(message),
			severity = VALUES(severity),
			metric_value = VALUES(metric_value),
			threshold_value = VALUES(threshold_value),
			is_read = FALSE
	`

	if _, err := r.db.ExecContext(ctx, query,
		alert.UserID,
		alert.Type,
		alert.Title,
		alert.Message,
		alert.Severity,
		alert.MetricValue,
		alert.ThresholdValue,
		alert.DedupeKey,
	); err != nil {
		return Alert{}, err
	}

	return r.getAlertByDedupeKey(ctx, alert.UserID, alert.DedupeKey)
}

func (r *MySQLAlertsRepository) ListAlerts(ctx context.Context, userID int64, filter AlertListFilter) ([]Alert, error) {
	builder := strings.Builder{}
	builder.WriteString(`
		SELECT id, user_id, type, title, message, severity, metric_value, threshold_value, dedupe_key, is_read, created_at, updated_at
		FROM alerts
		WHERE user_id = ?
	`)

	args := []any{userID}
	if filter.Type != nil && strings.TrimSpace(*filter.Type) != "" {
		builder.WriteString(" AND type = ?")
		args = append(args, strings.TrimSpace(*filter.Type))
	}
	if filter.Read != nil {
		builder.WriteString(" AND is_read = ?")
		args = append(args, *filter.Read)
	}
	builder.WriteString(" ORDER BY created_at DESC, id DESC")

	rows, err := r.db.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Alert, 0)
	for rows.Next() {
		item, err := scanAlert(rows)
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

func (r *MySQLAlertsRepository) MarkAlertRead(ctx context.Context, userID, alertID int64, isRead bool) (Alert, error) {
	const query = `
		UPDATE alerts
		SET is_read = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, isRead, alertID, userID)
	if err != nil {
		return Alert{}, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return Alert{}, err
	}
	if rows == 0 {
		return Alert{}, ErrNotFound
	}

	return r.getAlertByID(ctx, userID, alertID)
}

func (r *MySQLAlertsRepository) CurrentMonthExpense(ctx context.Context, userID int64) (float64, error) {
	start := startOfMonth(time.Now())
	end := start.AddDate(0, 1, 0)

	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM (
			SELECT amount FROM transactions
			WHERE user_id = ? AND type = 'expense' AND date >= ? AND date < ?
			UNION ALL
			SELECT p.amount
			FROM debt_payments p
			JOIN debts d ON d.id = p.debt_id
			WHERE d.user_id = ? AND p.payment_date >= ? AND p.payment_date < ?
		) sources
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID, start, end, userID, start, end).Scan(&value); err != nil {
		return 0, err
	}
	return value, nil
}

func (r *MySQLAlertsRepository) TodayExpense(ctx context.Context, userID int64) (float64, error) {
	start := startOfDay(time.Now())
	end := start.AddDate(0, 0, 1)

	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM (
			SELECT amount FROM transactions
			WHERE user_id = ? AND type = 'expense' AND date >= ? AND date < ?
			UNION ALL
			SELECT p.amount
			FROM debt_payments p
			JOIN debts d ON d.id = p.debt_id
			WHERE d.user_id = ? AND p.payment_date >= ? AND p.payment_date < ?
		) sources
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID, start, end, userID, start, end).Scan(&value); err != nil {
		return 0, err
	}
	return value, nil
}

func (r *MySQLAlertsRepository) LatestSalaryAmount(ctx context.Context, userID int64) (float64, error) {
	const query = `
		SELECT amount
		FROM salary_records
		WHERE user_id = ?
		ORDER BY paid_at DESC, id DESC
		LIMIT 1
	`

	var amount float64
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&amount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("latest salary amount: %w", err)
	}

	return amount, nil
}

func (r *MySQLAlertsRepository) getAlertByID(ctx context.Context, userID, alertID int64) (Alert, error) {
	const query = `
		SELECT id, user_id, type, title, message, severity, metric_value, threshold_value, dedupe_key, is_read, created_at, updated_at
		FROM alerts
		WHERE id = ? AND user_id = ?
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, alertID, userID)
	return scanAlertRow(row)
}

func (r *MySQLAlertsRepository) getAlertByDedupeKey(ctx context.Context, userID int64, dedupeKey string) (Alert, error) {
	const query = `
		SELECT id, user_id, type, title, message, severity, metric_value, threshold_value, dedupe_key, is_read, created_at, updated_at
		FROM alerts
		WHERE user_id = ? AND dedupe_key = ?
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, userID, dedupeKey)
	return scanAlertRow(row)
}

func scanAlert(rows *sql.Rows) (Alert, error) {
	var item Alert
	if err := rows.Scan(
		&item.ID,
		&item.UserID,
		&item.Type,
		&item.Title,
		&item.Message,
		&item.Severity,
		&item.MetricValue,
		&item.ThresholdValue,
		&item.DedupeKey,
		&item.IsRead,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return Alert{}, err
	}

	return item, nil
}

func scanAlertRow(row *sql.Row) (Alert, error) {
	var item Alert
	if err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Type,
		&item.Title,
		&item.Message,
		&item.Severity,
		&item.MetricValue,
		&item.ThresholdValue,
		&item.DedupeKey,
		&item.IsRead,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Alert{}, ErrNotFound
		}
		return Alert{}, err
	}

	return item, nil
}
