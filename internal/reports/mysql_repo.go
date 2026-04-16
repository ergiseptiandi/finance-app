package reports

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type MySQLReportsRepository struct {
	db *sql.DB
}

func NewMySQLReportsRepository(db *sql.DB) *MySQLReportsRepository {
	return &MySQLReportsRepository{db: db}
}

func (r *MySQLReportsRepository) ExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]CategoryExpense, error) {
	const query = `
		SELECT category_name, COALESCE(SUM(amount), 0) AS amount
		FROM (
			SELECT category AS category_name, amount
			FROM transactions
			WHERE user_id = ? AND type = 'expense' AND date >= ? AND date < ?
			UNION ALL
			SELECT 'Debt Payment' AS category_name, p.amount
			FROM debt_payments p
			JOIN debts d ON d.id = p.debt_id
			WHERE d.user_id = ? AND p.payment_date >= ? AND p.payment_date < ?
		) sources
		GROUP BY category_name
		ORDER BY amount DESC, category_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, start, end, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CategoryExpense, 0)
	for rows.Next() {
		var item CategoryExpense
		if err := rows.Scan(&item.Category, &item.Amount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLReportsRepository) ExpenseTrendByMonth(ctx context.Context, userID int64, start, end time.Time) (map[string]float64, error) {
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

func (r *MySQLReportsRepository) AllTimeIncome(ctx context.Context, userID int64) (float64, error) {
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

func (r *MySQLReportsRepository) AllTimeExpense(ctx context.Context, userID int64) (float64, error) {
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

func (r *MySQLReportsRepository) ExpenseBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
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

func normalizeMySQLError(err error) error {
	return fmt.Errorf("reports query: %w", err)
}

var _ = errors.New
