package debt

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

type MySQLDebtRepository struct {
	db *sql.DB
}

func NewMySQLDebtRepository(db *sql.DB) *MySQLDebtRepository {
	return &MySQLDebtRepository{db: db}
}

func (r *MySQLDebtRepository) CreateDebt(ctx context.Context, debt Debt, installments []Installment) (Debt, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Debt{}, err
	}
	defer rollback(tx)

	const insertDebt = `
		INSERT INTO debts (user_id, name, total_amount, monthly_installment, due_date, paid_amount, remaining_amount, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := tx.ExecContext(ctx, insertDebt, debt.UserID, debt.Name, debt.TotalAmount, debt.MonthlyInstallment, debt.DueDate, 0, debt.TotalAmount, StatusPending)
	if err != nil {
		return Debt{}, normalizeMySQLError(err)
	}

	debtID, err := res.LastInsertId()
	if err != nil {
		return Debt{}, err
	}
	debt.ID = debtID
	debt.PaidAmount = 0
	debt.RemainingAmount = debt.TotalAmount
	debt.Status = StatusPending

	for i := range installments {
		installments[i].DebtID = debtID
		if err := insertInstallment(ctx, tx, installments[i]); err != nil {
			return Debt{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Debt{}, err
	}

	return debt, nil
}

func (r *MySQLDebtRepository) GetDebtByID(ctx context.Context, userID, debtID int64) (Debt, error) {
	const query = `
		SELECT id, user_id, name, total_amount, monthly_installment, due_date, paid_amount, remaining_amount, status, created_at, updated_at
		FROM debts
		WHERE id = ? AND user_id = ?
		LIMIT 1
	`

	var debt Debt
	if err := r.db.QueryRowContext(ctx, query, debtID, userID).Scan(
		&debt.ID,
		&debt.UserID,
		&debt.Name,
		&debt.TotalAmount,
		&debt.MonthlyInstallment,
		&debt.DueDate,
		&debt.PaidAmount,
		&debt.RemainingAmount,
		&debt.Status,
		&debt.CreatedAt,
		&debt.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Debt{}, ErrNotFound
		}
		return Debt{}, err
	}

	return debt, nil
}

func (r *MySQLDebtRepository) UpdateDebt(ctx context.Context, debt Debt) error {
	const query = `
		UPDATE debts
		SET name = ?, total_amount = ?, monthly_installment = ?, due_date = ?, status = ?, paid_amount = ?, remaining_amount = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, debt.Name, debt.TotalAmount, debt.MonthlyInstallment, debt.DueDate, debt.Status, debt.PaidAmount, debt.RemainingAmount, debt.ID, debt.UserID)
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

func (r *MySQLDebtRepository) DeleteDebt(ctx context.Context, userID, debtID int64) error {
	const query = `DELETE FROM debts WHERE id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, debtID, userID)
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

func (r *MySQLDebtRepository) ListDebts(ctx context.Context, userID int64) ([]DebtSummary, error) {
	const query = `
		SELECT
			d.id, d.user_id, d.name, d.total_amount, d.monthly_installment, d.due_date, d.paid_amount, d.remaining_amount, d.status, d.created_at, d.updated_at,
			COALESCE((
				SELECT COUNT(*) FROM debt_installments di
				WHERE di.debt_id = d.id AND di.status = 'paid'
			), 0) AS paid_installments,
			COALESCE((
				SELECT COUNT(*) FROM debt_installments di
				WHERE di.debt_id = d.id AND di.status IN ('pending', 'overdue')
			), 0) AS unpaid_installments,
			COALESCE((
				SELECT COUNT(*) FROM debt_installments di
				WHERE di.debt_id = d.id AND di.status = 'overdue'
			), 0) AS overdue_installments
		FROM debts d
		WHERE d.user_id = ?
		ORDER BY d.due_date ASC, d.id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]DebtSummary, 0)
	for rows.Next() {
		var item DebtSummary
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Name,
			&item.TotalAmount,
			&item.MonthlyInstallment,
			&item.DueDate,
			&item.PaidAmount,
			&item.RemainingAmount,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.PaidInstallments,
			&item.UnpaidInstallments,
			&item.OverdueInstallments,
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

func (r *MySQLDebtRepository) GetInstallments(ctx context.Context, userID, debtID int64) ([]Installment, error) {
	const query = `
		SELECT di.id, di.debt_id, di.installment_no, di.due_date, di.amount, di.status, di.paid_at, di.created_at, di.updated_at
		FROM debt_installments di
		JOIN debts d ON d.id = di.debt_id
		WHERE d.id = ? AND d.user_id = ?
		ORDER BY di.installment_no ASC
	`

	rows, err := r.db.QueryContext(ctx, query, debtID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Installment, 0)
	for rows.Next() {
		var item Installment
		var paidAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.DebtID,
			&item.InstallmentNo,
			&item.DueDate,
			&item.Amount,
			&item.Status,
			&paidAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if paidAt.Valid {
			value := paidAt.Time
			item.PaidAt = &value
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLDebtRepository) GetPayments(ctx context.Context, userID, debtID int64) ([]Payment, error) {
	const query = `
		SELECT p.id, p.debt_id, p.installment_id, p.amount, p.payment_date, p.proof_image, p.created_at, p.updated_at
		FROM debt_payments p
		JOIN debts d ON d.id = p.debt_id
		WHERE d.id = ? AND d.user_id = ?
		ORDER BY p.payment_date DESC, p.id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, debtID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Payment, 0)
	for rows.Next() {
		var item Payment
		var installmentID sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.DebtID,
			&installmentID,
			&item.Amount,
			&item.PaymentDate,
			&item.ProofImage,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if installmentID.Valid {
			value := installmentID.Int64
			item.InstallmentID = &value
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MySQLDebtRepository) GetPaymentByID(ctx context.Context, userID, debtID, paymentID int64) (Payment, error) {
	const query = `
		SELECT p.id, p.debt_id, p.installment_id, p.amount, p.payment_date, p.proof_image, p.created_at, p.updated_at
		FROM debt_payments p
		JOIN debts d ON d.id = p.debt_id
		WHERE d.id = ? AND d.user_id = ? AND p.id = ?
		LIMIT 1
	`

	var payment Payment
	var installmentID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, debtID, userID, paymentID).Scan(
		&payment.ID,
		&payment.DebtID,
		&installmentID,
		&payment.Amount,
		&payment.PaymentDate,
		&payment.ProofImage,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Payment{}, ErrNotFound
		}
		return Payment{}, err
	}
	if installmentID.Valid {
		value := installmentID.Int64
		payment.InstallmentID = &value
	}

	return payment, nil
}

func (r *MySQLDebtRepository) ReplaceSchedule(ctx context.Context, debtID int64, installments []Installment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	if _, err := tx.ExecContext(ctx, `DELETE FROM debt_installments WHERE debt_id = ?`, debtID); err != nil {
		return err
	}

	for i := range installments {
		installments[i].DebtID = debtID
		if err := insertInstallment(ctx, tx, installments[i]); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *MySQLDebtRepository) GetNextUnpaidInstallment(ctx context.Context, debtID int64) (Installment, error) {
	const query = `
		SELECT id, debt_id, installment_no, due_date, amount, status, paid_at, created_at, updated_at
		FROM debt_installments
		WHERE debt_id = ? AND status IN ('pending', 'overdue')
		ORDER BY installment_no ASC
		LIMIT 1
	`

	var item Installment
	var paidAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, debtID).Scan(
		&item.ID,
		&item.DebtID,
		&item.InstallmentNo,
		&item.DueDate,
		&item.Amount,
		&item.Status,
		&paidAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Installment{}, ErrNoInstallment
		}
		return Installment{}, err
	}
	if paidAt.Valid {
		value := paidAt.Time
		item.PaidAt = &value
	}

	return item, nil
}

func (r *MySQLDebtRepository) CreatePaymentAndMarkInstallment(ctx context.Context, payment Payment, installmentID int64, paidAt time.Time) (Payment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Payment{}, err
	}
	defer rollback(tx)

	const insertPayment = `
		INSERT INTO debt_payments (debt_id, installment_id, amount, payment_date, proof_image)
		VALUES (?, ?, ?, ?, ?)
	`
	res, err := tx.ExecContext(ctx, insertPayment, payment.DebtID, installmentID, payment.Amount, payment.PaymentDate, payment.ProofImage)
	if err != nil {
		return Payment{}, err
	}
	paymentID, err := res.LastInsertId()
	if err != nil {
		return Payment{}, err
	}
	payment.ID = paymentID
	payment.InstallmentID = &installmentID

	if err := markInstallmentPaidTx(ctx, tx, installmentID, paidAt); err != nil {
		return Payment{}, err
	}

	if err := refreshDebtTotalsTx(ctx, tx, payment.DebtID); err != nil {
		return Payment{}, err
	}

	if err := tx.Commit(); err != nil {
		return Payment{}, err
	}

	return r.loadPaymentByDebtAndID(ctx, payment.DebtID, paymentID)
}

func (r *MySQLDebtRepository) UpdatePayment(ctx context.Context, payment Payment) error {
	const query = `
		UPDATE debt_payments
		SET amount = ?, payment_date = ?, proof_image = ?
		WHERE id = ? AND debt_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, payment.Amount, payment.PaymentDate, payment.ProofImage, payment.ID, payment.DebtID)
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

func (r *MySQLDebtRepository) MarkInstallmentPaid(ctx context.Context, debtID, installmentID int64, paidAt time.Time) (Installment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Installment{}, err
	}
	defer rollback(tx)

	_, err = markInstallmentPaidTxReturning(ctx, tx, debtID, installmentID, paidAt)
	if err != nil {
		return Installment{}, err
	}

	if err := refreshDebtTotalsTx(ctx, tx, debtID); err != nil {
		return Installment{}, err
	}

	if err := tx.Commit(); err != nil {
		return Installment{}, err
	}

	return r.loadInstallmentByDebtAndID(ctx, debtID, installmentID)
}

func (r *MySQLDebtRepository) RefreshUserDebtStatuses(ctx context.Context, userID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	const overdueQuery = `
		UPDATE debt_installments di
		JOIN debts d ON d.id = di.debt_id
		SET di.status = 'overdue'
		WHERE d.user_id = ? AND di.status = 'pending' AND di.due_date < CURDATE()
	`
	if _, err := tx.ExecContext(ctx, overdueQuery, userID); err != nil {
		return err
	}

	const debtRefreshQuery = `
		UPDATE debts d
		SET
			paid_amount = COALESCE((
				SELECT SUM(di.amount)
				FROM debt_installments di
				WHERE di.debt_id = d.id AND di.status = 'paid'
			), 0),
			remaining_amount = GREATEST(d.total_amount - COALESCE((
				SELECT SUM(di.amount)
				FROM debt_installments di
				WHERE di.debt_id = d.id AND di.status = 'paid'
			), 0), 0),
			status = CASE
				WHEN COALESCE((
					SELECT SUM(di.amount)
					FROM debt_installments di
					WHERE di.debt_id = d.id AND di.status = 'paid'
				), 0) >= d.total_amount THEN 'paid'
				WHEN EXISTS (
					SELECT 1
					FROM debt_installments di
					WHERE di.debt_id = d.id AND di.status = 'overdue'
				) THEN 'overdue'
				ELSE 'pending'
			END
		WHERE d.user_id = ?
	`
	if _, err := tx.ExecContext(ctx, debtRefreshQuery, userID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func insertInstallment(ctx context.Context, exec execer, installment Installment) error {
	const query = `
		INSERT INTO debt_installments (debt_id, installment_no, due_date, amount, status)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := exec.ExecContext(ctx, query, installment.DebtID, installment.InstallmentNo, installment.DueDate, installment.Amount, installment.Status)
	if err != nil {
		return err
	}

	return nil
}

type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func markInstallmentPaidTx(ctx context.Context, exec execer, installmentID int64, paidAt time.Time) error {
	const query = `
		UPDATE debt_installments
		SET status = 'paid', paid_at = ?
		WHERE id = ?
	`

	result, err := exec.ExecContext(ctx, query, paidAt, installmentID)
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

func markInstallmentPaidTxReturning(ctx context.Context, exec execer, debtID, installmentID int64, paidAt time.Time) (Installment, error) {
	const query = `
		UPDATE debt_installments
		SET status = 'paid', paid_at = ?
		WHERE id = ? AND debt_id = ?
	`

	result, err := exec.ExecContext(ctx, query, paidAt, installmentID, debtID)
	if err != nil {
		return Installment{}, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return Installment{}, err
	}
	if rows == 0 {
		return Installment{}, ErrNotFound
	}

	return Installment{
		ID:     installmentID,
		DebtID: debtID,
		Status: StatusPaid,
		PaidAt: &paidAt,
	}, nil
}

func refreshDebtTotalsTx(ctx context.Context, exec execer, debtID int64) error {
	const query = `
		UPDATE debts d
		SET
			paid_amount = COALESCE((
				SELECT SUM(di.amount)
				FROM debt_installments di
				WHERE di.debt_id = d.id AND di.status = 'paid'
			), 0),
			remaining_amount = GREATEST(d.total_amount - COALESCE((
				SELECT SUM(di.amount)
				FROM debt_installments di
				WHERE di.debt_id = d.id AND di.status = 'paid'
			), 0), 0),
			status = CASE
				WHEN COALESCE((
					SELECT SUM(di.amount)
					FROM debt_installments di
					WHERE di.debt_id = d.id AND di.status = 'paid'
				), 0) >= d.total_amount THEN 'paid'
				WHEN EXISTS (
					SELECT 1
					FROM debt_installments di
					WHERE di.debt_id = d.id AND di.status = 'overdue'
				) THEN 'overdue'
				ELSE 'pending'
			END
		WHERE d.id = ?
	`
	_, err := exec.ExecContext(ctx, query, debtID)
	return err
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

func normalizeMySQLError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrInvalidInput
	}

	if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return ErrInvalidInput
	}

	return err
}

func (r *MySQLDebtRepository) loadPaymentByDebtAndID(ctx context.Context, debtID, paymentID int64) (Payment, error) {
	const query = `
		SELECT id, debt_id, installment_id, amount, payment_date, proof_image, created_at, updated_at
		FROM debt_payments
		WHERE id = ? AND debt_id = ?
		LIMIT 1
	`

	var payment Payment
	var installmentID sql.NullInt64
	if err := r.db.QueryRowContext(ctx, query, paymentID, debtID).Scan(
		&payment.ID,
		&payment.DebtID,
		&installmentID,
		&payment.Amount,
		&payment.PaymentDate,
		&payment.ProofImage,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Payment{}, ErrNotFound
		}
		return Payment{}, err
	}
	if installmentID.Valid {
		value := installmentID.Int64
		payment.InstallmentID = &value
	}

	return payment, nil
}

func (r *MySQLDebtRepository) loadInstallmentByDebtAndID(ctx context.Context, debtID, installmentID int64) (Installment, error) {
	const query = `
		SELECT id, debt_id, installment_no, due_date, amount, status, paid_at, created_at, updated_at
		FROM debt_installments
		WHERE id = ? AND debt_id = ?
		LIMIT 1
	`

	var item Installment
	var paidAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, installmentID, debtID).Scan(
		&item.ID,
		&item.DebtID,
		&item.InstallmentNo,
		&item.DueDate,
		&item.Amount,
		&item.Status,
		&paidAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Installment{}, ErrNotFound
		}
		return Installment{}, err
	}
	if paidAt.Valid {
		value := paidAt.Time
		item.PaidAt = &value
	}

	return item, nil
}
