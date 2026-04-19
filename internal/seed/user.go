package seed

import (
	"context"
	"database/sql"

	"finance-backend/internal/auth"
	"finance-backend/internal/config"
)

func UpsertBootstrapUser(ctx context.Context, db *sql.DB, cfg config.SeedConfig) (int64, error) {
	passwordHash, err := auth.HashPassword(cfg.UserPassword)
	if err != nil {
		return 0, err
	}

	const query = `
		INSERT INTO users (name, email, password_hash)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			password_hash = VALUES(password_hash),
			updated_at = CURRENT_TIMESTAMP
	`

	if _, err = db.ExecContext(ctx, query, cfg.UserName, cfg.UserEmail, passwordHash); err != nil {
		return 0, err
	}

	var userID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ? LIMIT 1`, cfg.UserEmail).Scan(&userID); err != nil {
		return 0, err
	}

	return userID, nil
}
