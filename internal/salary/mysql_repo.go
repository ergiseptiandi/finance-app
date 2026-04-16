package salary

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	mysql "github.com/go-sql-driver/mysql"
)

type MySQLSalaryRepository struct {
	db *sql.DB
}

func NewMySQLSalaryRepository(db *sql.DB) *MySQLSalaryRepository {
	return &MySQLSalaryRepository{db: db}
}

func (r *MySQLSalaryRepository) CreateRecord(ctx context.Context, item SalaryRecord) (int64, error) {
	const query = `
		INSERT INTO salary_records (user_id, amount, paid_at, note)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, item.UserID, item.Amount, item.PaidAt, item.Note)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *MySQLSalaryRepository) GetRecordByID(ctx context.Context, id int64, userID int64) (SalaryRecord, error) {
	const query = `
		SELECT id, user_id, amount, paid_at, note, created_at, updated_at
		FROM salary_records
		WHERE id = ? AND user_id = ?
		LIMIT 1
	`

	var item SalaryRecord
	var note sql.NullString
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.Amount,
		&item.PaidAt,
		&note,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SalaryRecord{}, ErrNotFound
		}
		return SalaryRecord{}, err
	}

	if note.Valid {
		item.Note = note.String
	}

	return item, nil
}

func (r *MySQLSalaryRepository) GetCurrentRecord(ctx context.Context, userID int64) (SalaryRecord, error) {
	const query = `
		SELECT id, user_id, amount, paid_at, note, created_at, updated_at
		FROM salary_records
		WHERE user_id = ?
		ORDER BY paid_at DESC, id DESC
		LIMIT 1
	`

	var item SalaryRecord
	var note sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.Amount,
		&item.PaidAt,
		&note,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SalaryRecord{}, ErrNotFound
		}
		return SalaryRecord{}, err
	}

	if note.Valid {
		item.Note = note.String
	}

	return item, nil
}

func (r *MySQLSalaryRepository) UpdateRecord(ctx context.Context, item SalaryRecord) error {
	const query = `
		UPDATE salary_records
		SET amount = ?, paid_at = ?, note = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, item.Amount, item.PaidAt, item.Note, item.ID, item.UserID)
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

func (r *MySQLSalaryRepository) DeleteRecord(ctx context.Context, id int64, userID int64) error {
	const query = `DELETE FROM salary_records WHERE id = ? AND user_id = ?`

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

func (r *MySQLSalaryRepository) FindHistory(ctx context.Context, userID int64) ([]SalaryRecord, error) {
	const query = `
		SELECT id, user_id, amount, paid_at, note, created_at, updated_at
		FROM salary_records
		WHERE user_id = ?
		ORDER BY paid_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SalaryRecord, 0)
	for rows.Next() {
		var item SalaryRecord
		var note sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Amount,
			&item.PaidAt,
			&note,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if note.Valid {
			item.Note = note.String
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLSalaryRepository) GetSchedule(ctx context.Context, userID int64) (*SalarySchedule, error) {
	const query = `
		SELECT user_id, salary_day, created_at, updated_at
		FROM salary_settings
		WHERE user_id = ?
		LIMIT 1
	`

	var schedule SalarySchedule
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&schedule.UserID,
		&schedule.SalaryDay,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &schedule, nil
}

func (r *MySQLSalaryRepository) UpsertSchedule(ctx context.Context, userID int64, salaryDay int) (SalarySchedule, error) {
	const query = `
		INSERT INTO salary_settings (user_id, salary_day)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE salary_day = VALUES(salary_day)
	`

	_, err := r.db.ExecContext(ctx, query, userID, salaryDay)
	if err != nil {
		return SalarySchedule{}, normalizeMySQLError(err)
	}

	schedule, err := r.GetSchedule(ctx, userID)
	if err != nil {
		return SalarySchedule{}, err
	}
	if schedule == nil {
		return SalarySchedule{}, ErrNotFound
	}

	return *schedule, nil
}

func normalizeMySQLError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		if mysqlErr.Number == 1062 {
			return ErrInvalidInput
		}
	}

	if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return ErrInvalidInput
	}

	return err
}
