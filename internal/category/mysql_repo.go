package category

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	mysql "github.com/go-sql-driver/mysql"
)

type MySQLCategoryRepository struct {
	db *sql.DB
}

func NewMySQLCategoryRepository(db *sql.DB) *MySQLCategoryRepository {
	return &MySQLCategoryRepository{db: db}
}

func (r *MySQLCategoryRepository) Create(ctx context.Context, item Category) (int64, error) {
	const query = `
		INSERT INTO categories (name, type)
		VALUES (?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, item.Name, item.Type)
	if err != nil {
		return 0, normalizeMySQLError(err)
	}

	return result.LastInsertId()
}

func (r *MySQLCategoryRepository) GetByID(ctx context.Context, id int64) (Category, error) {
	const query = `
		SELECT id, name, type, created_at, updated_at
		FROM categories
		WHERE id = ?
		LIMIT 1
	`

	var item Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Type,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Category{}, ErrNotFound
		}
		return Category{}, err
	}

	return item, nil
}

func (r *MySQLCategoryRepository) Update(ctx context.Context, item Category) error {
	const query = `
		UPDATE categories
		SET name = ?, type = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, item.Name, item.Type, item.ID)
	if err != nil {
		return normalizeMySQLError(err)
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

func (r *MySQLCategoryRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM categories WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
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

func (r *MySQLCategoryRepository) FindAll(ctx context.Context, filter ListFilter) ([]Category, error) {
	query := `
		SELECT id, name, type, created_at, updated_at
		FROM categories
	`

	args := make([]any, 0, 1)
	if filter.Type != nil && *filter.Type != "" {
		query += " WHERE type = ?"
		args = append(args, *filter.Type)
	}

	query += " ORDER BY type ASC, name ASC, id ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Category, 0)
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func normalizeMySQLError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrAlreadyExists
	}

	if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return ErrAlreadyExists
	}

	return err
}
