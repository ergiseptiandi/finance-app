package repository

import (
	"context"
	"database/sql"

	"finance-backend/internal/auth"
)

type MySQLAuthRepository struct {
	db *sql.DB
}

func NewMySQLAuthRepository(db *sql.DB) *MySQLAuthRepository {
	return &MySQLAuthRepository{db: db}
}

func (r *MySQLAuthRepository) FindByEmail(ctx context.Context, email string) (auth.User, error) {
	const query = `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = ?
		LIMIT 1
	`

	var user auth.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

func (r *MySQLAuthRepository) FindByID(ctx context.Context, id int64) (auth.User, error) {
	const query = `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
		LIMIT 1
	`

	var user auth.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

func (r *MySQLAuthRepository) Create(ctx context.Context, params auth.CreateRefreshTokenParams) error {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, device_name, expires_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		params.UserID,
		params.TokenHash,
		params.DeviceName,
		params.ExpiresAt.UTC(),
	)
	return err
}

func (r *MySQLAuthRepository) FindActiveByHash(ctx context.Context, tokenHash string) (auth.RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, device_name, expires_at, revoked_at, created_at, updated_at, last_used_at
		FROM refresh_tokens
		WHERE token_hash = ?
		LIMIT 1
	`

	var token auth.RefreshToken
	var revokedAt sql.NullTime
	var lastUsedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.DeviceName,
		&token.ExpiresAt,
		&revokedAt,
		&token.CreatedAt,
		&token.UpdatedAt,
		&lastUsedAt,
	)
	if err != nil {
		return auth.RefreshToken{}, err
	}

	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}

	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Time
	}

	return token, nil
}

func (r *MySQLAuthRepository) Rotate(ctx context.Context, currentTokenHash string, nextToken auth.CreateRefreshTokenParams) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const revokeQuery = `
		UPDATE refresh_tokens
		SET revoked_at = CURRENT_TIMESTAMP, last_used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE token_hash = ? AND revoked_at IS NULL
	`

	result, err := tx.ExecContext(ctx, revokeQuery, currentTokenHash)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return auth.ErrInvalidToken
	}

	const insertQuery = `
		INSERT INTO refresh_tokens (user_id, token_hash, device_name, expires_at)
		VALUES (?, ?, ?, ?)
	`

	if _, err := tx.ExecContext(
		ctx,
		insertQuery,
		nextToken.UserID,
		nextToken.TokenHash,
		nextToken.DeviceName,
		nextToken.ExpiresAt.UTC(),
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MySQLAuthRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE token_hash = ? AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, tokenHash)
	return err
}

func (r *MySQLAuthRepository) CreateUser(ctx context.Context, user auth.User) (int64, error) {
	const query = `
		INSERT INTO users (name, email, password_hash)
		VALUES (?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.PasswordHash)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *MySQLAuthRepository) UpdateUser(ctx context.Context, user auth.User) error {
	const query = `
		UPDATE users
		SET name = ?, email = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
	return err
}

func (r *MySQLAuthRepository) UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	const query = `
		UPDATE users
		SET password_hash = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, passwordHash, id)
	return err
}

type MySQLPasswordResetRepository struct {
	db *sql.DB
}

func NewMySQLPasswordResetRepository(db *sql.DB) *MySQLPasswordResetRepository {
	return &MySQLPasswordResetRepository{db: db}
}

func (r *MySQLPasswordResetRepository) Create(ctx context.Context, token auth.PasswordResetToken) error {
	const query = `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES (?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt.UTC())
	return err
}

func (r *MySQLPasswordResetRepository) FindByHash(ctx context.Context, tokenHash string) (auth.PasswordResetToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = ?
		LIMIT 1
	`
	var token auth.PasswordResetToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	return token, err
}

func (r *MySQLPasswordResetRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM password_reset_tokens WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MySQLPasswordResetRepository) DeleteAllForUser(ctx context.Context, userID int64) error {
	const query = `DELETE FROM password_reset_tokens WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
