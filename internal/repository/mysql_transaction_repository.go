package repository

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"

	"finance-backend/internal/transaction"
)

type MySQLTransactionRepository struct {
	db *sql.DB
}

func NewMySQLTransactionRepository(db *sql.DB) *MySQLTransactionRepository {
	return &MySQLTransactionRepository{db: db}
}

func (r *MySQLTransactionRepository) Create(ctx context.Context, txn transaction.Transaction) (int64, error) {
	const query = `
		INSERT INTO transactions (user_id, type, category, amount, date, description)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, txn.UserID, txn.Type, txn.Category, txn.Amount, txn.Date, txn.Description)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *MySQLTransactionRepository) GetByID(ctx context.Context, id int64, userID int64) (transaction.Transaction, error) {
	const query = `
		SELECT id, user_id, type, category, amount, date, description, created_at, updated_at
		FROM transactions
		WHERE id = ? AND user_id = ?
		LIMIT 1;
	`
	
	var txn transaction.Transaction
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&txn.ID,
		&txn.UserID,
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
			return transaction.Transaction{}, transaction.ErrNotFound
		}
		return transaction.Transaction{}, err
	}
	
	if description.Valid {
		txn.Description = description.String
	}
	return txn, nil
}

func (r *MySQLTransactionRepository) Update(ctx context.Context, txn transaction.Transaction) error {
	const query = `
		UPDATE transactions
		SET type = ?, category = ?, amount = ?, date = ?, description = ?
		WHERE id = ? AND user_id = ?
	`
	res, err := r.db.ExecContext(ctx, query, txn.Type, txn.Category, txn.Amount, txn.Date, txn.Description, txn.ID, txn.UserID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return transaction.ErrNotFound
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
		return transaction.ErrNotFound
	}
	return nil
}

func (r *MySQLTransactionRepository) FindAll(ctx context.Context, userID int64, filter transaction.ListFilter) (transaction.PaginatedList, error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("FROM transactions WHERE user_id = ?")
	
	args := []interface{}{userID}
	
	if filter.StartDate != nil {
		queryBuilder.WriteString(" AND date >= ?")
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != nil {
		queryBuilder.WriteString(" AND date <= ?")
		args = append(args, filter.EndDate)
	}
	if filter.Category != nil && *filter.Category != "" {
		queryBuilder.WriteString(" AND category = ?")
		args = append(args, *filter.Category)
	}
	if filter.Type != nil && *filter.Type != "" {
		queryBuilder.WriteString(" AND type = ?")
		args = append(args, *filter.Type)
	}
	
	whereClause := queryBuilder.String()
	
	var total int64
	countQuery := "SELECT COUNT(*) " + whereClause
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return transaction.PaginatedList{}, err
	}
	
	selectQuery := "SELECT id, user_id, type, category, amount, date, description, created_at, updated_at " + 
				   whereClause + " ORDER BY date DESC, id DESC LIMIT ? OFFSET ?"
	
	offset := (filter.Page - 1) * filter.PerPage
	args = append(args, filter.PerPage, offset)
	
	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return transaction.PaginatedList{}, err
	}
	defer rows.Close()
	
	var items []transaction.Transaction
	for rows.Next() {
		var txn transaction.Transaction
		var desc sql.NullString
		if err := rows.Scan(
			&txn.ID, &txn.UserID, &txn.Type, &txn.Category, &txn.Amount, &txn.Date, &desc, &txn.CreatedAt, &txn.UpdatedAt,
		); err != nil {
			return transaction.PaginatedList{}, err
		}
		if desc.Valid {
			txn.Description = desc.String
		}
		items = append(items, txn)
	}
	
	if err := rows.Err(); err != nil {
		return transaction.PaginatedList{}, err
	}
	
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(filter.PerPage)))
	}
	
	return transaction.PaginatedList{
		Data:       items,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (r *MySQLTransactionRepository) GetSummary(ctx context.Context, userID int64) (transaction.Summary, error) {
	const query = `
		SELECT type, SUM(amount)
		FROM transactions
		WHERE user_id = ?
		GROUP BY type
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return transaction.Summary{}, err
	}
	defer rows.Close()

	var sum transaction.Summary
	for rows.Next() {
		var txnType string
		var amount float64
		if err := rows.Scan(&txnType, &amount); err != nil {
			return transaction.Summary{}, err
		}
		if txnType == string(transaction.TypeIncome) {
			sum.TotalIncome = amount
		} else if txnType == string(transaction.TypeExpense) {
			sum.TotalExpense = amount
		}
	}
	if err := rows.Err(); err != nil {
		return transaction.Summary{}, err
	}
	
	sum.Balance = sum.TotalIncome - sum.TotalExpense
	return sum, nil
}
