package wallet

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	mysql "github.com/go-sql-driver/mysql"
)

const defaultWalletName = "Main"

type MySQLWalletRepository struct {
	db *sql.DB
}

func NewMySQLWalletRepository(db *sql.DB) *MySQLWalletRepository {
	return &MySQLWalletRepository{db: db}
}

func (r *MySQLWalletRepository) Create(ctx context.Context, item Wallet) (int64, error) {
	const query = `
		INSERT INTO wallets (user_id, name, opening_balance, is_locked, is_archived)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, item.UserID, item.Name, item.OpeningBalance, item.IsLocked, item.IsArchived)
	if err != nil {
		return 0, normalizeMySQLError(err)
	}

	return result.LastInsertId()
}

func (r *MySQLWalletRepository) GetByID(ctx context.Context, userID, id int64) (Wallet, error) {
	return r.loadWallet(ctx, "WHERE w.id = ? AND w.user_id = ?", id, userID)
}

func (r *MySQLWalletRepository) GetByName(ctx context.Context, userID int64, name string) (Wallet, error) {
	return r.loadWallet(ctx, "WHERE w.user_id = ? AND w.name = ?", userID, strings.TrimSpace(name))
}

func (r *MySQLWalletRepository) Update(ctx context.Context, item Wallet) error {
	const query = `
		UPDATE wallets
		SET name = ?, opening_balance = ?, is_archived = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, item.Name, item.OpeningBalance, item.IsArchived, item.ID, item.UserID)
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

func (r *MySQLWalletRepository) Delete(ctx context.Context, userID, id int64) error {
	const query = `DELETE FROM wallets WHERE id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, id, userID)
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

func (r *MySQLWalletRepository) Archive(ctx context.Context, userID, id int64) error {
	const query = `
		UPDATE wallets
		SET is_archived = 1
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id, userID)
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

func (r *MySQLWalletRepository) FindAll(ctx context.Context, userID int64) ([]Wallet, error) {
	const query = `
	SELECT
		w.id,
		w.user_id,
		w.name,
		w.opening_balance,
		w.is_locked,
		w.is_archived,
		w.created_at,
		w.updated_at,
			w.opening_balance +
			COALESCE((
				SELECT SUM(
					CASE
						WHEN t.type = 'income' THEN t.amount
						ELSE -t.amount
					END
				)
				FROM transactions t
				WHERE t.wallet_id = w.id
			), 0) +
			COALESCE((
				SELECT SUM(-p.amount)
				FROM debt_payments p
				WHERE p.wallet_id = w.id
			), 0) +
			COALESCE((
				SELECT SUM(
					CASE
						WHEN wt.to_wallet_id = w.id THEN wt.amount
						ELSE -wt.amount
					END
				)
				FROM wallet_transfers wt
				WHERE wt.from_wallet_id = w.id OR wt.to_wallet_id = w.id
			), 0) AS balance
		FROM wallets w
		WHERE w.user_id = ? AND w.is_archived = 0
		ORDER BY w.created_at ASC, w.id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Wallet, 0)
	for rows.Next() {
		var item Wallet
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Name,
			&item.OpeningBalance,
			&item.IsLocked,
			&item.IsArchived,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Balance,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLWalletRepository) CreateTransfer(ctx context.Context, transfer Transfer) (int64, error) {
	const query = `
		INSERT INTO wallet_transfers (user_id, from_wallet_id, to_wallet_id, amount, note, transfer_date)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, transfer.UserID, transfer.FromWalletID, transfer.ToWalletID, transfer.Amount, transfer.Note, transfer.TransferDate)
	if err != nil {
		return 0, normalizeMySQLError(err)
	}

	return result.LastInsertId()
}

func (r *MySQLWalletRepository) GetTransferByID(ctx context.Context, userID, id int64) (Transfer, error) {
	const query = `
		SELECT id, user_id, from_wallet_id, to_wallet_id, amount, note, transfer_date, created_at, updated_at
		FROM wallet_transfers
		WHERE id = ? AND user_id = ?
		LIMIT 1
	`

	var item Transfer
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.FromWalletID,
		&item.ToWalletID,
		&item.Amount,
		&item.Note,
		&item.TransferDate,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Transfer{}, ErrNotFound
		}
		return Transfer{}, err
	}

	return item, nil
}

func (r *MySQLWalletRepository) FindTransfers(ctx context.Context, userID int64) ([]Transfer, error) {
	const query = `
		SELECT id, user_id, from_wallet_id, to_wallet_id, amount, note, transfer_date, created_at, updated_at
		FROM wallet_transfers
		WHERE user_id = ?
		ORDER BY transfer_date DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Transfer, 0)
	for rows.Next() {
		var item Transfer
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.FromWalletID,
			&item.ToWalletID,
			&item.Amount,
			&item.Note,
			&item.TransferDate,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLWalletRepository) GetDefaultWallet(ctx context.Context, userID int64) (Wallet, error) {
	return r.GetByName(ctx, userID, defaultWalletName)
}

func (r *MySQLWalletRepository) loadWallet(ctx context.Context, where string, args ...any) (Wallet, error) {
	const queryTemplate = `
		SELECT
			w.id,
			w.user_id,
			w.name,
			w.opening_balance,
			w.is_locked,
			w.is_archived,
			w.created_at,
			w.updated_at,
			w.opening_balance +
			COALESCE((
				SELECT SUM(
					CASE
						WHEN t.type = 'income' THEN t.amount
						ELSE -t.amount
					END
				)
				FROM transactions t
				WHERE t.wallet_id = w.id
			), 0) +
			COALESCE((
				SELECT SUM(-p.amount)
				FROM debt_payments p
				WHERE p.wallet_id = w.id
			), 0) +
			COALESCE((
				SELECT SUM(
					CASE
						WHEN wt.to_wallet_id = w.id THEN wt.amount
						ELSE -wt.amount
					END
				)
				FROM wallet_transfers wt
				WHERE wt.from_wallet_id = w.id OR wt.to_wallet_id = w.id
			), 0) AS balance
		FROM wallets w
		%s
		LIMIT 1
	`

	query := strings.TrimSpace(queryTemplate)
	query = strings.Replace(query, "%s", where, 1)

	var item Wallet
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&item.ID,
		&item.UserID,
		&item.Name,
		&item.OpeningBalance,
		&item.IsLocked,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.Balance,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Wallet{}, ErrNotFound
		}
		return Wallet{}, err
	}
	if item.IsArchived {
		return Wallet{}, ErrNotFound
	}

	return item, nil
}

func normalizeMySQLError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return ErrAlreadyExists
		case 1451, 1452:
			return ErrConflict
		}
	}

	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "duplicate"):
		return ErrAlreadyExists
	case strings.Contains(lower, "foreign key"):
		return ErrConflict
	}

	return err
}

var _ Repository = (*MySQLWalletRepository)(nil)
