package database

import (
	"context"
	"database/sql"

	"finance-backend/internal/auth"
	"finance-backend/internal/config"
)

func BootstrapAuth(ctx context.Context, db *sql.DB, cfg config.AuthConfig) error {
	statements := []string{
		`
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(191) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			token_hash CHAR(64) NOT NULL UNIQUE,
			device_name VARCHAR(120) NOT NULL DEFAULT '',
			expires_at TIMESTAMP NOT NULL,
			revoked_at TIMESTAMP NULL DEFAULT NULL,
			last_used_at TIMESTAMP NULL DEFAULT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_refresh_tokens_user_id (user_id),
			INDEX idx_refresh_tokens_expires_at (expires_at)
		)
		`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	passwordHash, err := auth.HashPassword(cfg.BootstrapPassword)
	if err != nil {
		return err
	}

	const upsertUser = `
		INSERT INTO users (name, email, password_hash)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			password_hash = VALUES(password_hash),
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = db.ExecContext(ctx, upsertUser, cfg.BootstrapName, cfg.BootstrapEmail, passwordHash)
	return err
}
