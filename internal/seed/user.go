package seed

import (
	"context"
	"database/sql"

	"finance-backend/internal/auth"
	"finance-backend/internal/config"
)

func UpsertBootstrapUser(ctx context.Context, db *sql.DB, cfg config.SeedConfig) error {
	passwordHash, err := auth.HashPassword(cfg.UserPassword)
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO users (name, email, password_hash)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			password_hash = VALUES(password_hash),
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = db.ExecContext(ctx, query, cfg.UserName, cfg.UserEmail, passwordHash)
	return err
}
