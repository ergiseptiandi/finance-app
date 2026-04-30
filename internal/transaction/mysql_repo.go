package transaction

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
)

type MySQLTransactionRepository struct {
	db *sql.DB
}

func NewMySQLTransactionRepository(db *sql.DB) *MySQLTransactionRepository {
	return &MySQLTransactionRepository{db: db}
}

func (r *MySQLTransactionRepository) Create(ctx context.Context, txn Transaction) (int64, error) {
	const query = `
		INSERT INTO transactions (user_id, wallet_id, type, category, amount, date, description)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, txn.UserID, txn.WalletID, txn.Type, txn.Category, txn.Amount, txn.Date, txn.Description)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *MySQLTransactionRepository) GetByID(ctx context.Context, id int64, userID int64) (Transaction, error) {
	const query = `
		SELECT id, user_id, wallet_id, type, category, amount, date, description, created_at, updated_at
		FROM transactions
		WHERE id = ? AND user_id = ?
		LIMIT 1;
	`

	var txn Transaction
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&txn.ID,
		&txn.UserID,
		&txn.WalletID,
		&txn.Type,
		&txn.Category,
		&txn.Amount,
		&txn.Date,
		&description,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Transaction{}, ErrNotFound
		}
		return Transaction{}, err
	}

	if description.Valid {
		txn.Description = description.String
	}
	return txn, nil
}

func (r *MySQLTransactionRepository) Update(ctx context.Context, txn Transaction) error {
	const query = `
		UPDATE transactions
		SET wallet_id = ?, type = ?, category = ?, amount = ?, date = ?, description = ?
		WHERE id = ? AND user_id = ?
	`
	res, err := r.db.ExecContext(ctx, query, txn.WalletID, txn.Type, txn.Category, txn.Amount, txn.Date, txn.Description, txn.ID, txn.UserID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MySQLTransactionRepository) Delete(ctx context.Context, id int64, userID int64) error {
	const query = `DELETE FROM transactions WHERE id = ? AND user_id = ?`
	res, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MySQLTransactionRepository) FindAll(ctx context.Context, userID int64, filter ListFilter) (PaginatedList, error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("FROM transactions WHERE user_id = ?")

	args := []interface{}{userID}
	applyTransactionFilters(&queryBuilder, &args, filter)

	whereClause := queryBuilder.String()

	var total int64
	countQuery := "SELECT COUNT(*) " + whereClause
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return PaginatedList{}, err
	}

	selectQuery := "SELECT id, user_id, wallet_id, type, category, amount, date, description, created_at, updated_at " +
		whereClause + " ORDER BY date DESC, id DESC LIMIT ? OFFSET ?"

	offset := (filter.Page - 1) * filter.PerPage
	args = append(args, filter.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return PaginatedList{}, err
	}
	defer rows.Close()

	var items []Transaction
	for rows.Next() {
		var txn Transaction
		var desc sql.NullString
		if err := rows.Scan(
			&txn.ID, &txn.UserID, &txn.WalletID, &txn.Type, &txn.Category, &txn.Amount, &txn.Date, &desc, &txn.CreatedAt, &txn.UpdatedAt,
		); err != nil {
			return PaginatedList{}, err
		}
		if desc.Valid {
			txn.Description = desc.String
		}
		items = append(items, txn)
	}

	if err := rows.Err(); err != nil {
		return PaginatedList{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(filter.PerPage)))
	}

	return PaginatedList{
		Data:       items,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (r *MySQLTransactionRepository) GetSummary(ctx context.Context, userID int64, filter ListFilter) (Summary, error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`SELECT type, COALESCE(SUM(amount), 0)
FROM (
	SELECT user_id, wallet_id, type, category, date, amount
	FROM transactions
	UNION ALL
	SELECT d.user_id AS user_id, p.wallet_id AS wallet_id, 'expense' AS type, 'Debt Payment' AS category, p.payment_date AS date, p.amount
	FROM debt_payments p
	JOIN debts d ON d.id = p.debt_id
) sources
WHERE user_id = ?`)

	args := []interface{}{userID}
	applyTransactionFilters(&queryBuilder, &args, filter)
	queryBuilder.WriteString(" GROUP BY type")

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return Summary{}, err
	}
	defer rows.Close()

	var sum Summary
	for rows.Next() {
		var txnType string
		var amount float64
		if err := rows.Scan(&txnType, &amount); err != nil {
			return Summary{}, err
		}
		if txnType == string(TypeIncome) {
			sum.TotalIncome = amount
		} else if txnType == string(TypeExpense) {
			sum.TotalExpense = amount
		}
	}
	if err := rows.Err(); err != nil {
		return Summary{}, err
	}

	// Get consumption expense (only transactions, no debt payments)
	consumptionQuery := strings.Builder{}
	consumptionQuery.WriteString(`SELECT COALESCE(SUM(amount), 0)
FROM transactions
WHERE user_id = ? AND type = 'expense'`)
	consumptionArgs := []interface{}{userID}
	if filter.StartDate != nil {
		consumptionQuery.WriteString(" AND date >= ?")
		consumptionArgs = append(consumptionArgs, *filter.StartDate)
	}
	if filter.EndDate != nil {
		consumptionQuery.WriteString(" AND date < ?")
		consumptionArgs = append(consumptionArgs, filter.EndDate.AddDate(0, 0, 1))
	}
	if filter.WalletID != nil && *filter.WalletID > 0 {
		consumptionQuery.WriteString(" AND wallet_id = ?")
		consumptionArgs = append(consumptionArgs, *filter.WalletID)
	}
	if filter.Category != nil && *filter.Category != "" {
		consumptionQuery.WriteString(" AND category = ?")
		consumptionArgs = append(consumptionArgs, *filter.Category)
	}
	_ = r.db.QueryRowContext(ctx, consumptionQuery.String(), consumptionArgs...).Scan(&sum.ConsumptionExpense)

	// Get debt repayment (only debt payments)
	debtQuery := strings.Builder{}
	debtQuery.WriteString(`SELECT COALESCE(SUM(p.amount), 0)
FROM debt_payments p
JOIN debts d ON d.id = p.debt_id
WHERE d.user_id = ?`)
	debtArgs := []interface{}{userID}
	if filter.StartDate != nil {
		debtQuery.WriteString(" AND p.payment_date >= ?")
		debtArgs = append(debtArgs, *filter.StartDate)
	}
	if filter.EndDate != nil {
		debtQuery.WriteString(" AND p.payment_date < ?")
		debtArgs = append(debtArgs, filter.EndDate.AddDate(0, 0, 1))
	}
	_ = r.db.QueryRowContext(ctx, debtQuery.String(), debtArgs...).Scan(&sum.DebtRepayment)

	sum.Balance = sum.TotalIncome - sum.ConsumptionExpense
	if sum.TotalIncome > 0 {
		sum.SavingsRate = math.Round(((sum.TotalIncome-sum.ConsumptionExpense)/sum.TotalIncome)*10000) / 100
		sum.ConsumptionRate = math.Round((sum.ConsumptionExpense/sum.TotalIncome)*10000) / 100
	}
	return sum, nil
}

func applyTransactionFilters(queryBuilder *strings.Builder, args *[]interface{}, filter ListFilter) {
	if filter.StartDate != nil {
		queryBuilder.WriteString(" AND date >= ?")
		*args = append(*args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		queryBuilder.WriteString(" AND date < ?")
		*args = append(*args, filter.EndDate.AddDate(0, 0, 1))
	}
	if filter.WalletID != nil && *filter.WalletID > 0 {
		queryBuilder.WriteString(" AND wallet_id = ?")
		*args = append(*args, *filter.WalletID)
	}
	if filter.Category != nil && *filter.Category != "" {
		queryBuilder.WriteString(" AND category = ?")
		*args = append(*args, *filter.Category)
	}
	if filter.Type != nil && *filter.Type != "" {
		queryBuilder.WriteString(" AND type = ?")
		*args = append(*args, *filter.Type)
	}
}
