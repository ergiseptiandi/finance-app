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
		SELECT category_name, COALESCE(SUM(amount), 0) AS amount, COUNT(*) AS transaction_count
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
		return nil, normalizeMySQLError(err)
	}
	defer rows.Close()

	items := make([]CategoryExpense, 0)
	for rows.Next() {
		var item CategoryExpense
		if err := rows.Scan(&item.Category, &item.Amount, &item.TransactionCount); err != nil {
			return nil, normalizeMySQLError(err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, normalizeMySQLError(err)
	}

	return items, nil
}

func (r *MySQLReportsRepository) TrendByPeriod(ctx context.Context, userID int64, start, end time.Time, groupBy TrendGroupBy) (map[string]TrendTotals, error) {
	format := "%Y-%m"
	if groupBy == TrendGroupByDay {
		format = "%Y-%m-%d"
	}

	query := fmt.Sprintf(`
		SELECT period_key, kind, COALESCE(SUM(amount), 0)
		FROM (
			SELECT DATE_FORMAT(date, '%s') AS period_key, 'income' AS kind, amount
			FROM transactions
			WHERE user_id = ? AND type = 'income' AND date >= ? AND date < ?
			UNION ALL
			SELECT DATE_FORMAT(date, '%s') AS period_key, 'expense' AS kind, amount
			FROM transactions
			WHERE user_id = ? AND type = 'expense' AND date >= ? AND date < ?
			UNION ALL
			SELECT DATE_FORMAT(p.payment_date, '%s') AS period_key, 'expense' AS kind, p.amount
			FROM debt_payments p
			JOIN debts d ON d.id = p.debt_id
			WHERE d.user_id = ? AND p.payment_date >= ? AND p.payment_date < ?
		) sources
		GROUP BY period_key, kind
	`, format, format, format)

	rows, err := r.db.QueryContext(ctx, query, userID, start, end, userID, start, end, userID, start, end)
	if err != nil {
		return nil, normalizeMySQLError(err)
	}
	defer rows.Close()

	values := make(map[string]TrendTotals)
	for rows.Next() {
		var key string
		var kind string
		var amount float64
		if err := rows.Scan(&key, &kind, &amount); err != nil {
			return nil, normalizeMySQLError(err)
		}

		total := values[key]
		switch kind {
		case "income":
			total.Income += amount
		case "expense":
			total.Expense += amount
		}
		values[key] = total
	}
	if err := rows.Err(); err != nil {
		return nil, normalizeMySQLError(err)
	}

	return values, nil
}

func (r *MySQLReportsRepository) AllTimeIncome(ctx context.Context, userID int64) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = ? AND type = 'income'
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&value); err != nil {
		return 0, normalizeMySQLError(err)
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
		return 0, normalizeMySQLError(err)
	}
	return value, nil
}

func (r *MySQLReportsRepository) IncomeBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = ? AND type = 'income' AND date >= ? AND date < ?
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID, start, end).Scan(&value); err != nil {
		return 0, normalizeMySQLError(err)
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
		return 0, normalizeMySQLError(err)
	}
	return value, nil
}

func normalizeMySQLError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("reports query: %w", err)
}

var _ = errors.New
