package debt

import "context"
import "time"

type Repository interface {
	CreateDebt(ctx context.Context, debt Debt, installments []Installment) (Debt, error)
	GetDebtByID(ctx context.Context, userID, debtID int64) (Debt, error)
	UpdateDebt(ctx context.Context, debt Debt) error
	DeleteDebt(ctx context.Context, userID, debtID int64) error
	ListDebts(ctx context.Context, userID int64) ([]DebtSummary, error)
	GetInstallments(ctx context.Context, userID, debtID int64) ([]Installment, error)
	GetPayments(ctx context.Context, userID, debtID int64) ([]Payment, error)
	GetPaymentByID(ctx context.Context, userID, debtID, paymentID int64) (Payment, error)
	ReplaceSchedule(ctx context.Context, debtID int64, installments []Installment) error
	GetNextUnpaidInstallment(ctx context.Context, debtID int64) (Installment, error)
	CreatePaymentAndMarkInstallment(ctx context.Context, payment Payment, installmentID int64, paidAt time.Time) (Payment, error)
	UpdatePayment(ctx context.Context, payment Payment) error
	MarkInstallmentPaid(ctx context.Context, debtID, installmentID int64, paidAt time.Time) (Installment, error)
	RefreshUserDebtStatuses(ctx context.Context, userID int64) error
}
