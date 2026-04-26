package budget

import (
	"context"
	"database/sql"
	"errors"

	mysql "github.com/go-sql-driver/mysql"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Create(ctx context.Context, userID int64, categoryID int64, monthlyAmount float64) (Goal, error) {
	const query = `
		INSERT INTO budget_goals (user_id, category_id, monthly_amount)
		VALUES (?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, userID, categoryID, monthlyAmount)
	if err != nil {
		return Goal{}, normalizeMySQLError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Goal{}, err
	}
	return r.GetByID(ctx, userID, id)
}

func (r *MySQLRepository) Update(ctx context.Context, userID int64, item Goal) (Goal, error) {
	const query = `
		UPDATE budget_goals
		SET category_id = ?, monthly_amount = ?
		WHERE id = ? AND user_id = ?
	`
	result, err := r.db.ExecContext(ctx, query, item.CategoryID, item.MonthlyAmount, item.ID, userID)
	if err != nil {
		return Goal{}, normalizeMySQLError(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return Goal{}, err
	}
	if rows == 0 {
		return Goal{}, ErrNotFound
	}
	return r.GetByID(ctx, userID, item.ID)
}

func (r *MySQLRepository) GetByID(ctx context.Context, userID, id int64) (Goal, error) {
	const query = `
		SELECT g.id, g.user_id, g.category_id, c.name, c.type, g.monthly_amount, g.created_at, g.updated_at
		FROM budget_goals g
		JOIN categories c ON c.id = g.category_id
		WHERE g.id = ? AND g.user_id = ?
		LIMIT 1
	`

	var item Goal
	if err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.CategoryID,
		&item.CategoryName,
		&item.CategoryType,
		&item.MonthlyAmount,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Goal{}, ErrNotFound
		}
		return Goal{}, err
	}

	return item, nil
}

func (r *MySQLRepository) Delete(ctx context.Context, userID, id int64) error {
	const query = `DELETE FROM budget_goals WHERE id = ? AND user_id = ?`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MySQLRepository) List(ctx context.Context, userID int64) ([]Goal, error) {
	const query = `
		SELECT g.id, g.user_id, g.category_id, c.name, c.type, g.monthly_amount, g.created_at, g.updated_at
		FROM budget_goals g
		JOIN categories c ON c.id = g.category_id
		WHERE g.user_id = ?
		ORDER BY c.type ASC, c.name ASC, g.id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Goal, 0)
	for rows.Next() {
		var item Goal
		if err := rows.Scan(&item.ID, &item.UserID, &item.CategoryID, &item.CategoryName, &item.CategoryType, &item.MonthlyAmount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MySQLRepository) GetCategory(ctx context.Context, userID, categoryID int64) (CategoryInfo, error) {
	const query = `
		SELECT id, name, type
		FROM categories
		WHERE id = ? AND user_id = ?
		LIMIT 1
	`

	var item CategoryInfo
	if err := r.db.QueryRowContext(ctx, query, categoryID, userID).Scan(&item.ID, &item.Name, &item.Type); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CategoryInfo{}, ErrNotFound
		}
		return CategoryInfo{}, err
	}

	return item, nil
}

func normalizeMySQLError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return errors.New("budget goal already exists")
	}
	return err
}
