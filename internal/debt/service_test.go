package debt

import (
	"context"
	"testing"
	"time"
)

type repositoryStub struct {
	refreshUserDebtStatusesFn func(ctx context.Context, userID int64) error
	listDebtsFn               func(ctx context.Context, userID int64) ([]DebtSummary, error)
	getDebtByIDFn             func(ctx context.Context, userID, debtID int64) (Debt, error)
	getInstallmentsFn         func(ctx context.Context, userID, debtID int64) ([]Installment, error)
	getPaymentsFn             func(ctx context.Context, userID, debtID int64) ([]Payment, error)
}

func (r repositoryStub) CreateDebt(context.Context, Debt, []Installment) (Debt, error) {
	return Debt{}, nil
}

func (r repositoryStub) GetDebtByID(ctx context.Context, userID, debtID int64) (Debt, error) {
	if r.getDebtByIDFn != nil {
		return r.getDebtByIDFn(ctx, userID, debtID)
	}
	return Debt{}, nil
}

func (r repositoryStub) UpdateDebt(context.Context, Debt) error { return nil }

func (r repositoryStub) DeleteDebt(context.Context, int64, int64) error { return nil }

func (r repositoryStub) ListDebts(ctx context.Context, userID int64) ([]DebtSummary, error) {
	if r.listDebtsFn != nil {
		return r.listDebtsFn(ctx, userID)
	}
	return nil, nil
}

func (r repositoryStub) GetInstallments(ctx context.Context, userID, debtID int64) ([]Installment, error) {
	if r.getInstallmentsFn != nil {
		return r.getInstallmentsFn(ctx, userID, debtID)
	}
	return nil, nil
}

func (r repositoryStub) GetPayments(ctx context.Context, userID, debtID int64) ([]Payment, error) {
	if r.getPaymentsFn != nil {
		return r.getPaymentsFn(ctx, userID, debtID)
	}
	return nil, nil
}

func (r repositoryStub) GetPaymentByID(context.Context, int64, int64, int64) (Payment, error) {
	return Payment{}, nil
}

func (r repositoryStub) ReplaceSchedule(context.Context, int64, []Installment) error { return nil }

func (r repositoryStub) GetNextUnpaidInstallment(context.Context, int64) (Installment, error) {
	return Installment{}, nil
}

func (r repositoryStub) CreatePaymentAndMarkInstallment(context.Context, Payment, int64, time.Time) (Payment, error) {
	return Payment{}, nil
}

func (r repositoryStub) UpdatePayment(context.Context, Payment) error { return nil }

func (r repositoryStub) MarkInstallmentPaid(context.Context, int64, int64, time.Time) (Installment, error) {
	return Installment{}, nil
}

func (r repositoryStub) RefreshUserDebtStatuses(ctx context.Context, userID int64) error {
	if r.refreshUserDebtStatusesFn != nil {
		return r.refreshUserDebtStatusesFn(ctx, userID)
	}
	return nil
}

func TestListReturnsEmptySliceWhenRepositoryReturnsNil(t *testing.T) {
	t.Parallel()

	svc := NewService(repositoryStub{})

	items, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items == nil {
		t.Fatalf("expected non-nil slice, got nil")
	}
	if len(items) != 0 {
		t.Fatalf("expected empty slice, got %d item(s)", len(items))
	}
}

func TestPaymentHistoryReturnsEmptySliceWhenRepositoryReturnsNil(t *testing.T) {
	t.Parallel()

	svc := NewService(repositoryStub{})

	items, err := svc.PaymentHistory(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items == nil {
		t.Fatalf("expected non-nil slice, got nil")
	}
	if len(items) != 0 {
		t.Fatalf("expected empty slice, got %d item(s)", len(items))
	}
}

func TestInstallmentsReturnsEmptySliceWhenRepositoryReturnsNil(t *testing.T) {
	t.Parallel()

	svc := NewService(repositoryStub{})

	items, err := svc.Installments(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items == nil {
		t.Fatalf("expected non-nil slice, got nil")
	}
	if len(items) != 0 {
		t.Fatalf("expected empty slice, got %d item(s)", len(items))
	}
}

func TestDetailReturnsEmptyNestedSlicesWhenRepositoryReturnsNil(t *testing.T) {
	t.Parallel()

	svc := NewService(repositoryStub{
		getDebtByIDFn: func(context.Context, int64, int64) (Debt, error) {
			return Debt{
				ID:                 10,
				UserID:             1,
				Name:               "Debt A",
				TotalAmount:        1000,
				MonthlyInstallment: 100,
				DueDate:            time.Now(),
				Status:             StatusPending,
			}, nil
		},
	})

	detail, err := svc.Detail(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Installments == nil {
		t.Fatalf("expected non-nil installments slice, got nil")
	}
	if detail.Payments == nil {
		t.Fatalf("expected non-nil payments slice, got nil")
	}
	if len(detail.Installments) != 0 {
		t.Fatalf("expected empty installments, got %d item(s)", len(detail.Installments))
	}
	if len(detail.Payments) != 0 {
		t.Fatalf("expected empty payments, got %d item(s)", len(detail.Payments))
	}
}
