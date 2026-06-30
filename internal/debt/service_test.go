package debt

import (
	"context"
	"errors"
	"testing"
	"time"

	"finance-backend/internal/wallet"
)

func ptrInt64(v int64) *int64 { return &v }

type repositoryStub struct {
	refreshUserDebtStatusesFn   func(ctx context.Context, userID int64) error
	listDebtsFn                 func(ctx context.Context, userID int64) ([]DebtSummary, error)
	getDebtByIDFn               func(ctx context.Context, userID, debtID int64) (Debt, error)
	getInstallmentsFn           func(ctx context.Context, userID, debtID int64) ([]Installment, error)
	getPaymentsFn               func(ctx context.Context, userID, debtID int64) ([]Payment, error)
	getPaymentByIDFn            func(ctx context.Context, userID, debtID, paymentID int64) (Payment, error)
	getNextUnpaidInstallmentFn  func(ctx context.Context, debtID int64) (Installment, error)
	getUnpaidInstallmentsFn     func(ctx context.Context, debtID int64) ([]Installment, error)
	updatePaymentFn             func(ctx context.Context, payment Payment) error
	replaceUnpaidScheduleFn     func(ctx context.Context, debtID int64, installments []Installment) error
	createPaymentAndMarkFns     func(ctx context.Context, payment Payment, installmentIDs []int64, paidAt time.Time) (Payment, error)
	refreshDebtTotalsFn         func(ctx context.Context, debtID int64) error
	updateDebtFn                func(ctx context.Context, debt Debt) error
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

func (r repositoryStub) UpdateDebt(ctx context.Context, debt Debt) error {
	if r.updateDebtFn != nil {
		return r.updateDebtFn(ctx, debt)
	}
	return nil
}

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

func (r repositoryStub) GetPaymentByID(ctx context.Context, userID, debtID, paymentID int64) (Payment, error) {
	if r.getPaymentByIDFn != nil {
		return r.getPaymentByIDFn(ctx, userID, debtID, paymentID)
	}
	return Payment{}, nil
}

func (r repositoryStub) ReplaceSchedule(context.Context, int64, []Installment) error { return nil }

func (r repositoryStub) ReplaceUnpaidSchedule(ctx context.Context, debtID int64, installments []Installment) error {
	if r.replaceUnpaidScheduleFn != nil {
		return r.replaceUnpaidScheduleFn(ctx, debtID, installments)
	}
	return nil
}

func (r repositoryStub) GetNextUnpaidInstallment(ctx context.Context, debtID int64) (Installment, error) {
	if r.getNextUnpaidInstallmentFn != nil {
		return r.getNextUnpaidInstallmentFn(ctx, debtID)
	}
	return Installment{}, nil
}

func (r repositoryStub) GetUnpaidInstallments(ctx context.Context, debtID int64) ([]Installment, error) {
	if r.getUnpaidInstallmentsFn != nil {
		return r.getUnpaidInstallmentsFn(ctx, debtID)
	}
	return nil, nil
}

func (r repositoryStub) CreatePaymentAndMarkInstallment(context.Context, Payment, int64, time.Time) (Payment, error) {
	return Payment{}, nil
}

func (r repositoryStub) CreatePaymentAndMarkInstallments(ctx context.Context, payment Payment, installmentIDs []int64, paidAt time.Time) (Payment, error) {
	if r.createPaymentAndMarkFns != nil {
		return r.createPaymentAndMarkFns(ctx, payment, installmentIDs, paidAt)
	}
	return Payment{}, nil
}

func (r repositoryStub) UpdatePayment(ctx context.Context, payment Payment) error {
	if r.updatePaymentFn != nil {
		return r.updatePaymentFn(ctx, payment)
	}
	return nil
}

func (r repositoryStub) MarkInstallmentPaid(context.Context, int64, int64, time.Time) (Installment, error) {
	return Installment{}, nil
}

func (r repositoryStub) RefreshUserDebtStatuses(ctx context.Context, userID int64) error {
	if r.refreshUserDebtStatusesFn != nil {
		return r.refreshUserDebtStatusesFn(ctx, userID)
	}
	return nil
}

func (r repositoryStub) RefreshDebtTotals(ctx context.Context, debtID int64) error {
	if r.refreshDebtTotalsFn != nil {
		return r.refreshDebtTotalsFn(ctx, debtID)
	}
	return nil
}

func TestListReturnsEmptySliceWhenRepositoryReturnsNil(t *testing.T) {
	t.Parallel()

	svc := NewService(repositoryStub{}, nil)

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

	svc := NewService(repositoryStub{}, nil)

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

	svc := NewService(repositoryStub{}, nil)

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
	}, nil)

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

type walletResolverStub struct {
	getByIDFn       func(context.Context, int64, int64) (wallet.Wallet, error)
	defaultWalletFn func(context.Context, int64) (wallet.Wallet, error)
}

func (r walletResolverStub) GetByID(ctx context.Context, userID, id int64) (wallet.Wallet, error) {
	if r.getByIDFn != nil {
		return r.getByIDFn(ctx, userID, id)
	}
	return wallet.Wallet{ID: id}, nil
}

func (r walletResolverStub) DefaultWallet(ctx context.Context, userID int64) (wallet.Wallet, error) {
	if r.defaultWalletFn != nil {
		return r.defaultWalletFn(ctx, userID)
	}
	return wallet.Wallet{ID: 1}, nil
}

func TestCreatePaymentRejectsWhenWalletBalanceTooLow(t *testing.T) {
	t.Parallel()

	repo := repositoryStub{
		getDebtByIDFn: func(context.Context, int64, int64) (Debt, error) {
			return Debt{ID: 10, UserID: 2}, nil
		},
		getUnpaidInstallmentsFn: func(context.Context, int64) ([]Installment, error) {
			return []Installment{{ID: 1, Amount: 100}}, nil
		},
	}

	svc := NewService(repo, walletResolverStub{
		getByIDFn: func(context.Context, int64, int64) (wallet.Wallet, error) {
			return wallet.Wallet{ID: 7, Balance: 90}, nil
		},
	})

	amount := 120.0
	payment, err := svc.CreatePayment(context.Background(), 2, 10, CreatePaymentInput{
		WalletID:    ptrInt64(7),
		Amount:      amount,
		PaymentDate: time.Date(2026, time.April, 21, 10, 0, 0, 0, time.UTC),
		ProofImage:  "proof.jpg",
	})
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v (payment=%+v)", err, payment)
	}
}

func TestUpdatePaymentUsesOriginalAmountWhenValidatingBalance(t *testing.T) {
	t.Parallel()

	storedPayment := Payment{
		ID:          1,
		DebtID:      10,
		WalletID:    7,
		Amount:      40,
		PaymentDate: time.Date(2026, time.April, 21, 10, 0, 0, 0, time.UTC),
		ProofImage:  "proof.jpg",
	}

	repo := repositoryStub{
		getDebtByIDFn: func(context.Context, int64, int64) (Debt, error) {
			return Debt{ID: 10, UserID: 2}, nil
		},
		getPaymentByIDFn: func(context.Context, int64, int64, int64) (Payment, error) {
			return storedPayment, nil
		},
		updatePaymentFn: func(_ context.Context, payment Payment) error {
			storedPayment = payment
			return nil
		},
	}

	svc := NewService(repo, walletResolverStub{
		getByIDFn: func(context.Context, int64, int64) (wallet.Wallet, error) {
			return wallet.Wallet{ID: 7, Balance: 100}, nil
		},
	})

	amount := 130.0
	payment, err := svc.UpdatePayment(context.Background(), 2, 10, 1, UpdatePaymentInput{Amount: &amount})
	if err != nil {
		t.Fatalf("UpdatePayment returned error: %v", err)
	}
	if payment.Amount != 130 {
		t.Fatalf("unexpected amount: %v", payment.Amount)
	}
}

func TestUpdateAllowsEditingTotalAfterPayment(t *testing.T) {
	t.Parallel()

	dueDate := time.Date(2026, time.January, 25, 0, 0, 0, 0, time.UTC)
	currentDebt := Debt{
		ID:                 10,
		UserID:             1,
		Name:               "Kredit Motor",
		TotalAmount:        3_000_000,
		MonthlyInstallment: 500_000,
		DueDate:            dueDate,
		PaidAmount:         500_000,
		RemainingAmount:    2_500_000,
		Status:             StatusPending,
	}

	installments := []Installment{
		{ID: 1, DebtID: 10, InstallmentNo: 1, Amount: 500_000, Status: StatusPaid},
		{ID: 2, DebtID: 10, InstallmentNo: 2, Amount: 500_000, Status: StatusPending},
		{ID: 3, DebtID: 10, InstallmentNo: 3, Amount: 500_000, Status: StatusPending},
		{ID: 4, DebtID: 10, InstallmentNo: 4, Amount: 500_000, Status: StatusPending},
		{ID: 5, DebtID: 10, InstallmentNo: 5, Amount: 500_000, Status: StatusPending},
		{ID: 6, DebtID: 10, InstallmentNo: 6, Amount: 500_000, Status: StatusPending},
	}

	var updatedDebt Debt
	var newInstallments []Installment

	repo := repositoryStub{
		getDebtByIDFn: func(context.Context, int64, int64) (Debt, error) {
			return currentDebt, nil
		},
		getInstallmentsFn: func(context.Context, int64, int64) ([]Installment, error) {
			return installments, nil
		},
		updateDebtFn: func(_ context.Context, debt Debt) error {
			updatedDebt = debt
			return nil
		},
		replaceUnpaidScheduleFn: func(_ context.Context, _ int64, inst []Installment) error {
			newInstallments = inst
			return nil
		},
	}

	svc := NewService(repo, nil)

	newTotal := 3_500_000.0
	_, err := svc.Update(context.Background(), 1, 10, UpdateInput{TotalAmount: &newTotal})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if updatedDebt.TotalAmount != 3_500_000 {
		t.Fatalf("expected total 3.5M, got %v", updatedDebt.TotalAmount)
	}
	if updatedDebt.PaidAmount != 500_000 {
		t.Fatalf("expected paid 500K, got %v", updatedDebt.PaidAmount)
	}
	if updatedDebt.RemainingAmount != 3_000_000 {
		t.Fatalf("expected remaining 3M, got %v", updatedDebt.RemainingAmount)
	}
	if len(newInstallments) != 6 {
		t.Fatalf("expected 6 new unpaid installments, got %d", len(newInstallments))
	}
	if newInstallments[0].InstallmentNo != 2 {
		t.Fatalf("expected first new installment_no = 2, got %d", newInstallments[0].InstallmentNo)
	}
}

func TestCreatePaymentMarksMultipleInstallmentsOnOverpayment(t *testing.T) {
	t.Parallel()

	unpaid := []Installment{
		{ID: 1, DebtID: 10, InstallmentNo: 1, Amount: 175_000, Status: StatusPending},
		{ID: 2, DebtID: 10, InstallmentNo: 2, Amount: 175_000, Status: StatusPending},
		{ID: 3, DebtID: 10, InstallmentNo: 3, Amount: 175_000, Status: StatusPending},
	}

	var capturedIDs []int64

	repo := repositoryStub{
		getDebtByIDFn: func(context.Context, int64, int64) (Debt, error) {
			return Debt{ID: 10, UserID: 1}, nil
		},
		getUnpaidInstallmentsFn: func(context.Context, int64) ([]Installment, error) {
			return unpaid, nil
		},
		createPaymentAndMarkFns: func(_ context.Context, payment Payment, ids []int64, _ time.Time) (Payment, error) {
			capturedIDs = ids
			payment.ID = 99
			return payment, nil
		},
	}

	svc := NewService(repo, walletResolverStub{
		getByIDFn: func(context.Context, int64, int64) (wallet.Wallet, error) {
			return wallet.Wallet{ID: 1, Balance: 1_000_000}, nil
		},
	})

	payment, err := svc.CreatePayment(context.Background(), 1, 10, CreatePaymentInput{
		Amount:      700_000,
		PaymentDate: time.Date(2026, time.January, 25, 0, 0, 0, 0, time.UTC),
		ProofImage:  "proof.jpg",
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}
	if payment.ID != 99 {
		t.Fatalf("expected payment ID 99, got %v", payment.ID)
	}
	if len(capturedIDs) != 3 {
		t.Fatalf("expected 3 installments covered, got %d", len(capturedIDs))
	}
}

func TestCreatePaymentMarksTwoInstallmentsOnPartialOverpayment(t *testing.T) {
	t.Parallel()

	unpaid := []Installment{
		{ID: 1, DebtID: 10, InstallmentNo: 1, Amount: 200_000, Status: StatusPending},
		{ID: 2, DebtID: 10, InstallmentNo: 2, Amount: 200_000, Status: StatusPending},
		{ID: 3, DebtID: 10, InstallmentNo: 3, Amount: 200_000, Status: StatusPending},
	}

	var capturedIDs []int64

	repo := repositoryStub{
		getDebtByIDFn: func(context.Context, int64, int64) (Debt, error) {
			return Debt{ID: 10, UserID: 1}, nil
		},
		getUnpaidInstallmentsFn: func(context.Context, int64) ([]Installment, error) {
			return unpaid, nil
		},
		createPaymentAndMarkFns: func(_ context.Context, payment Payment, ids []int64, _ time.Time) (Payment, error) {
			capturedIDs = ids
			return payment, nil
		},
	}

	svc := NewService(repo, walletResolverStub{
		getByIDFn: func(context.Context, int64, int64) (wallet.Wallet, error) {
			return wallet.Wallet{ID: 1, Balance: 1_000_000}, nil
		},
	})

	_, err := svc.CreatePayment(context.Background(), 1, 10, CreatePaymentInput{
		Amount:      500_000,
		PaymentDate: time.Date(2026, time.January, 25, 0, 0, 0, 0, time.UTC),
		ProofImage:  "proof.jpg",
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}
	if len(capturedIDs) != 2 {
		t.Fatalf("expected 2 installments covered (500K / 200K = 2), got %d", len(capturedIDs))
	}
}
