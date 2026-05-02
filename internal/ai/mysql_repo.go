package ai

import (
	"context"
	"database/sql"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) GetChatCount(ctx context.Context, userID int64) (int, error) {
	const query = `SELECT COALESCE(ai_chat_count, 0) FROM users WHERE id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MySQLRepository) IncrementChatCount(ctx context.Context, userID int64) error {
	const query = `UPDATE users SET ai_chat_count = COALESCE(ai_chat_count, 0) + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
