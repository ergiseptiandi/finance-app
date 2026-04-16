package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type MySQLDashboardRepository struct {
	db *sql.DB
}

func NewMySQLDashboardRepository(db *sql.DB) *MySQLDashboardRepository {
	return &MySQLDashboardRepository{db: db}
}

func (r *MySQLDashboardRepository) AllTimeIncome(ctx context.Context, userID int64) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM (
			SELECT amount FROM transactions WHERE user_id = ? AND type = 'income'
			UNION ALL
			SELECT amount FROM salary_records WHERE user_id = ?
		) sources
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID, userID).Scan(&value); err != nil {
		return 0, err
	}
	return value, nil
}

func (r *MySQLDashboardRepository) AllTimeExpense(ctx context.Context, userID int64) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM (
			SELECT amount FROM transactions WHERE user_id = ? AND type = 'expense'
			UNION ALL
			SELECT p.amount
			FROM debt_payments p
			JOIN debts d ON d.id = p.debt_id
			WHERE d.user_id = ?
		) sources
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID, userID).Scan(&value); err != nil {
		return 0, err
	}
	return value, nil
}

func (r *MySQLDashboardRepository) IncomeBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM (
			SELECT amount FROM transactions
			WHERE user_id = ? AND type = 'income' AND date >= ? AND date < ?
			UNION ALL
			SELECT amount FROM salary_records
			WHERE user_id = ? AND paid_at >= ? AND paid_at < ?
		) sources
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID, start, end, userID, start, end).Scan(&value); err != nil {
		return 0, err
	}
	return value, nil
}

func (r *MySQLDashboardRepository) ExpenseBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
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

func (r *MySQLDashboardRepository) ExpenseByDay(ctx context.Context, userID int64, start, end time.Time) (map[string]float64, error) {
	const query = `
		SELECT day_key, COALESCE(SUM(amount), 0)
		FROM (
			SELECT DATE_FORMAT(date, '%Y-%m-%d') AS day_key, amount
			FROM transactions
			WHERE user_id = ? AND type = 'expense' AND date >= ? AND date < ?
			UNION ALL
			SELECT DATE_FORMAT(p.payment_date, '%Y-%m-%d') AS day_key, p.amount
			FROM debt_payments p
			JOIN debts d ON d.id = p.debt_id
			WHERE d.user_id = ? AND p.payment_date >= ? AND p.payment_date < ?
		) sources
		GROUP BY day_key
	`

	rows, err := r.db.QueryContext(ctx, query, userID, start, end, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[string]float64)
	for rows.Next() {
		var key string
		var amount float64
		if err := rows.Scan(&key, &amount); err != nil {
			return nil, err
		}
		values[key] = amount
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func (r *MySQLDashboardRepository) ExpenseByMonth(ctx context.Context, userID int64, start, end time.Time) (map[string]float64, error) {
	const query = `
		SELECT month_key, COALESCE(SUM(amount), 0)
		FROM (
			SELECT DATE_FORMAT(date, '%Y-%m') AS month_key, amount
			FROM transactions
			WHERE user_id = ? AND type = 'expense' AND date >= ? AND date < ?
			UNION ALL
			SELECT DATE_FORMAT(p.payment_date, '%Y-%m') AS month_key, p.amount
			FROM debt_payments p
			JOIN debts d ON d.id = p.debt_id
			WHERE d.user_id = ? AND p.payment_date >= ? AND p.payment_date < ?
		) sources
		GROUP BY month_key
	`

	rows, err := r.db.QueryContext(ctx, query, userID, start, end, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[string]float64)
	for rows.Next() {
		var key string
		var amount float64
		if err := rows.Scan(&key, &amount); err != nil {
			return nil, err
		}
		values[key] = amount
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func (r *MySQLDashboardRepository) LatestSalaryAmount(ctx context.Context, userID int64) (float64, error) {
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
