package dashboard

import (
	"context"
	"database/sql"
	"time"
)

type MySQLDashboardRepository struct {
	db *sql.DB
}

func NewMySQLDashboardRepository(db *sql.DB) *MySQLDashboardRepository {
	return &MySQLDashboardRepository{db: db}
}

func (r *MySQLDashboardRepository) RefreshUserDebtStatuses(ctx context.Context, userID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	const overdueQuery = `
		UPDATE debt_installments di
		JOIN debts d ON d.id = di.debt_id
		SET di.status = 'overdue'
		WHERE d.user_id = ? AND di.status = 'pending' AND di.due_date < CURDATE()
	`
	if _, err := tx.ExecContext(ctx, overdueQuery, userID); err != nil {
		return err
	}

	const debtRefreshQuery = `
		UPDATE debts d
		SET
			paid_amount = COALESCE((
				SELECT SUM(di.amount)
				FROM debt_installments di
				WHERE di.debt_id = d.id AND di.status = 'paid'
			), 0),
			remaining_amount = GREATEST(d.total_amount - COALESCE((
				SELECT SUM(di.amount)
				FROM debt_installments di
				WHERE di.debt_id = d.id AND di.status = 'paid'
			), 0), 0),
			status = CASE
				WHEN COALESCE((
					SELECT SUM(di.amount)
					FROM debt_installments di
					WHERE di.debt_id = d.id AND di.status = 'paid'
				), 0) >= d.total_amount THEN 'paid'
				WHEN EXISTS (
					SELECT 1
					FROM debt_installments di
					WHERE di.debt_id = d.id AND di.status = 'overdue'
				) THEN 'overdue'
				ELSE 'pending'
			END
		WHERE d.user_id = ?
	`
	if _, err := tx.ExecContext(ctx, debtRefreshQuery, userID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MySQLDashboardRepository) AllTimeIncome(ctx context.Context, userID int64) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = ? AND type = 'income'
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&value); err != nil {
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
		FROM transactions
		WHERE user_id = ? AND type = 'income' AND date >= ? AND date < ?
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID, start, end).Scan(&value); err != nil {
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

func (r *MySQLDashboardRepository) ConsumptionExpenseBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = ? AND type = 'expense' AND date >= ? AND date < ?
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID, start, end).Scan(&value); err != nil {
		return 0, err
	}
	return value, nil
}

func (r *MySQLDashboardRepository) DebtRepaymentBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(p.amount), 0)
		FROM debt_payments p
		JOIN debts d ON d.id = p.debt_id
		WHERE d.user_id = ? AND p.payment_date >= ? AND p.payment_date < ?
	`

	var value float64
	if err := r.db.QueryRowContext(ctx, query, userID, start, end).Scan(&value); err != nil {
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

func (r *MySQLDashboardRepository) DebtOverview(ctx context.Context, userID int64, start, end time.Time) (DebtOverview, error) {
	const debtQuery = `
		SELECT
			COALESCE(SUM(d.total_amount), 0),
			COALESCE(SUM(d.paid_amount), 0),
			COALESCE(SUM(d.remaining_amount), 0),
			COUNT(*),
			COALESCE(SUM(CASE WHEN d.status IN ('pending', 'overdue') THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN d.status = 'overdue' THEN 1 ELSE 0 END), 0)
		FROM debts d
		WHERE d.user_id = ?
	`

	var overview DebtOverview
	if err := r.db.QueryRowContext(ctx, debtQuery, userID).Scan(
		&overview.TotalDebt,
		&overview.PaidDebt,
		&overview.RemainingDebt,
		&overview.TotalDebtCount,
		&overview.ActiveDebtCount,
		&overview.OverdueDebtCount,
	); err != nil {
		return DebtOverview{}, err
	}

	const installmentQuery = `
		SELECT
			COALESCE(SUM(CASE WHEN di.status = 'paid' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN di.status = 'overdue' THEN 1 ELSE 0 END), 0)
		FROM debt_installments di
		JOIN debts d ON d.id = di.debt_id
		WHERE d.user_id = ?
	`
	if err := r.db.QueryRowContext(ctx, installmentQuery, userID).Scan(
		&overview.PaidInstallments,
		&overview.OverdueInstallments,
	); err != nil {
		return DebtOverview{}, err
	}

	const upcomingQuery = `
		SELECT
			COALESCE(SUM(di.amount), 0),
			COALESCE(COUNT(*), 0)
		FROM debt_installments di
		JOIN debts d ON d.id = di.debt_id
		WHERE d.user_id = ? AND di.status IN ('pending', 'overdue') AND di.due_date >= ? AND di.due_date < ?
	`
	if err := r.db.QueryRowContext(ctx, upcomingQuery, userID, start, end).Scan(
		&overview.UpcomingDueAmount,
		&overview.UpcomingDueInstallments,
	); err != nil {
		return DebtOverview{}, err
	}

	return overview, nil
}

func (r *MySQLDashboardRepository) ExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]CategoryBreakdownItem, error) {
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
		return nil, err
	}
	defer rows.Close()

	items := make([]CategoryBreakdownItem, 0)
	for rows.Next() {
		var item CategoryBreakdownItem
		if err := rows.Scan(&item.Category, &item.Amount, &item.TransactionCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLDashboardRepository) UpcomingBills(ctx context.Context, userID int64, start, end time.Time) ([]UpcomingBill, error) {
	const query = `
		SELECT
			CONCAT(d.name, ' installment #', di.installment_no) AS bill_name,
			di.amount,
			DATE_FORMAT(di.due_date, '%Y-%m-%d') AS due_date,
			di.status,
			'debt' AS source_type
		FROM debt_installments di
		JOIN debts d ON d.id = di.debt_id
		WHERE d.user_id = ? AND di.status IN ('pending', 'overdue') AND di.due_date >= ? AND di.due_date < ?
		ORDER BY di.due_date ASC, d.name ASC, di.installment_no ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]UpcomingBill, 0)
	for rows.Next() {
		var item UpcomingBill
		if err := rows.Scan(&item.BillName, &item.Amount, &item.DueDate, &item.Status, &item.SourceType); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLDashboardRepository) TopMerchants(ctx context.Context, userID int64, start, end time.Time, limit int) ([]TopMerchant, error) {
	if limit <= 0 {
		limit = 5
	}

	const query = `
		SELECT merchant_name, COALESCE(SUM(amount), 0) AS amount, COUNT(*) AS transaction_count, DATE_FORMAT(MAX(tx_date), '%Y-%m-%d') AS last_transaction_date
		FROM (
			SELECT COALESCE(NULLIF(TRIM(description), ''), category) AS merchant_name, amount, date AS tx_date
			FROM transactions
			WHERE user_id = ? AND type = 'expense' AND date >= ? AND date < ?
			UNION ALL
			SELECT COALESCE(NULLIF(TRIM(d.name), ''), 'Debt Payment') AS merchant_name, p.amount, p.payment_date AS tx_date
			FROM debt_payments p
			JOIN debts d ON d.id = p.debt_id
			WHERE d.user_id = ? AND p.payment_date >= ? AND p.payment_date < ?
		) sources
		GROUP BY merchant_name
		ORDER BY amount DESC, transaction_count DESC, merchant_name ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, start, end, userID, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]TopMerchant, 0)
	for rows.Next() {
		var item TopMerchant
		if err := rows.Scan(&item.MerchantName, &item.Amount, &item.TransactionCount, &item.LastTransactionDate); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}
