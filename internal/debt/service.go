package debt

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"finance-backend/internal/wallet"
)

var (
	ErrNotFound            = errors.New("debt not found")
	ErrInvalidInput        = errors.New("invalid debt input")
	ErrNoInstallment       = errors.New("no unpaid installment available")
	ErrInsufficientBalance = errors.New("saldo wallet tidak cukup")
)

type Service struct {
	repo    Repository
	wallets wallet.Resolver
}

func NewService(repo Repository, wallets wallet.Resolver) *Service {
	return &Service{repo: repo, wallets: wallets}
}

func (s *Service) Create(ctx context.Context, userID int64, input CreateInput) (DebtDetail, error) {
	debt, err := buildDebt(userID, input)
	if err != nil {
		return DebtDetail{}, err
	}

	installments, err := buildInstallments(debt)
	if err != nil {
		return DebtDetail{}, err
	}

	storedDebt, err := s.repo.CreateDebt(ctx, debt, installments)
	if err != nil {
		return DebtDetail{}, err
	}

	return s.Detail(ctx, userID, storedDebt.ID)
}

func (s *Service) Update(ctx context.Context, userID, debtID int64, input UpdateInput) (DebtDetail, error) {
	current, err := s.repo.GetDebtByID(ctx, userID, debtID)
	if err != nil {
		return DebtDetail{}, err
	}

	installments, err := s.repo.GetInstallments(ctx, userID, debtID)
	if err != nil {
		return DebtDetail{}, err
	}

	paidCount := countInstallmentsByStatus(installments, StatusPaid)
	changesSchedule := input.TotalAmount != nil || input.MonthlyInstallment != nil || input.DueDate != nil

	updated := current
	if input.Name != nil {
		updated.Name = strings.TrimSpace(*input.Name)
	}
	if input.TotalAmount != nil {
		updated.TotalAmount = *input.TotalAmount
	}
	if input.MonthlyInstallment != nil {
		updated.MonthlyInstallment = *input.MonthlyInstallment
	}
	if input.DueDate != nil {
		updated.DueDate = *input.DueDate
	}

	if err := validateDebt(updated.Name, updated.TotalAmount, updated.MonthlyInstallment, updated.DueDate); err != nil {
		return DebtDetail{}, err
	}

	if changesSchedule {
		if paidCount > 0 {
			return DebtDetail{}, errors.New("cannot change amount schedule after payments exist")
		}

		newInstallments, err := buildInstallments(updated)
		if err != nil {
			return DebtDetail{}, err
		}
		updated.PaidAmount = 0
		updated.RemainingAmount = updated.TotalAmount
		updated.Status = StatusPending

		if err := s.repo.UpdateDebt(ctx, updated); err != nil {
			return DebtDetail{}, err
		}
		if err := s.repo.ReplaceSchedule(ctx, debtID, newInstallments); err != nil {
			return DebtDetail{}, err
		}
	} else {
		if err := s.repo.UpdateDebt(ctx, updated); err != nil {
			return DebtDetail{}, err
		}
	}

	return s.Detail(ctx, userID, debtID)
}

func (s *Service) Delete(ctx context.Context, userID, debtID int64) error {
	return s.repo.DeleteDebt(ctx, userID, debtID)
}

func (s *Service) List(ctx context.Context, userID int64) ([]DebtSummary, error) {
	if err := s.repo.RefreshUserDebtStatuses(ctx, userID); err != nil {
		return nil, err
	}

	items, err := s.repo.ListDebts(ctx, userID)
	if err != nil {
		return nil, err
	}

	return ensureDebtSummarySlice(items), nil
}

func (s *Service) Detail(ctx context.Context, userID, debtID int64) (DebtDetail, error) {
	if err := s.repo.RefreshUserDebtStatuses(ctx, userID); err != nil {
		return DebtDetail{}, err
	}

	debt, err := s.repo.GetDebtByID(ctx, userID, debtID)
	if err != nil {
		return DebtDetail{}, err
	}

	installments, err := s.repo.GetInstallments(ctx, userID, debtID)
	if err != nil {
		return DebtDetail{}, err
	}

	payments, err := s.repo.GetPayments(ctx, userID, debtID)
	if err != nil {
		return DebtDetail{}, err
	}

	detail := DebtDetail{
		DebtSummary: DebtSummary{
			Debt:                debt,
			PaidInstallments:    countInstallmentsByStatus(installments, StatusPaid),
			UnpaidInstallments:  countInstallmentsByStatus(installments, StatusPending) + countInstallmentsByStatus(installments, StatusOverdue),
			OverdueInstallments: countInstallmentsByStatus(installments, StatusOverdue),
		},
		Installments: ensureInstallmentSlice(installments),
		Payments:     ensurePaymentSlice(payments),
	}

	return detail, nil
}

func (s *Service) CreatePayment(ctx context.Context, userID, debtID int64, input CreatePaymentInput) (Payment, error) {
	if input.Amount <= 0 {
		return Payment{}, errors.New("amount must be greater than zero")
	}
	if input.PaymentDate.IsZero() {
		return Payment{}, errors.New("payment_date is required")
	}
	if strings.TrimSpace(input.ProofImage) == "" {
		return Payment{}, errors.New("proof_image is required")
	}

	if _, err := s.repo.GetDebtByID(ctx, userID, debtID); err != nil {
		return Payment{}, err
	}

	if err := s.repo.RefreshUserDebtStatuses(ctx, userID); err != nil {
		return Payment{}, err
	}

	walletID, err := s.resolveWalletID(ctx, userID, input.WalletID)
	if err != nil {
		return Payment{}, err
	}

	nextInstallment, err := s.repo.GetNextUnpaidInstallment(ctx, debtID)
	if err != nil {
		return Payment{}, err
	}
	if input.Amount < nextInstallment.Amount {
		return Payment{}, errors.New("amount must be at least the installment amount")
	}
	if err := s.ensurePaymentBalance(ctx, userID, walletID, nil, input.Amount); err != nil {
		return Payment{}, err
	}

	payment := Payment{
		DebtID:      debtID,
		WalletID:    walletID,
		Amount:      input.Amount,
		PaymentDate: input.PaymentDate,
		ProofImage:  input.ProofImage,
	}

	return s.repo.CreatePaymentAndMarkInstallment(ctx, payment, nextInstallment.ID, input.PaymentDate)
}

func (s *Service) UpdatePayment(ctx context.Context, userID, debtID, paymentID int64, input UpdatePaymentInput) (Payment, error) {
	if _, err := s.repo.GetDebtByID(ctx, userID, debtID); err != nil {
		return Payment{}, err
	}

	if err := s.repo.RefreshUserDebtStatuses(ctx, userID); err != nil {
		return Payment{}, err
	}

	payment, err := s.repo.GetPaymentByID(ctx, userID, debtID, paymentID)
	if err != nil {
		return Payment{}, err
	}
	original := payment

	if input.Amount != nil {
		if *input.Amount <= 0 {
			return Payment{}, errors.New("amount must be greater than zero")
		}
		payment.Amount = *input.Amount
	}
	if input.PaymentDate != nil {
		if input.PaymentDate.IsZero() {
			return Payment{}, errors.New("payment_date is required")
		}
		payment.PaymentDate = *input.PaymentDate
	}
	if input.WalletID != nil {
		walletID, err := s.resolveWalletID(ctx, userID, input.WalletID)
		if err != nil {
			return Payment{}, err
		}
		payment.WalletID = walletID
	}
	if input.ProofImage != nil {
		payment.ProofImage = *input.ProofImage
	}
	if err := s.ensurePaymentBalance(ctx, userID, payment.WalletID, &original, payment.Amount); err != nil {
		return Payment{}, err
	}

	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		return Payment{}, err
	}

	return s.repo.GetPaymentByID(ctx, userID, debtID, paymentID)
}

func (s *Service) resolveWalletID(ctx context.Context, userID int64, walletID *int64) (int64, error) {
	if s.wallets == nil {
		return 0, errors.New("wallet service is required")
	}

	if walletID != nil {
		if *walletID <= 0 {
			return 0, errors.New("wallet_id must be a positive number")
		}
		item, err := s.wallets.GetByID(ctx, userID, *walletID)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	}

	item, err := s.wallets.DefaultWallet(ctx, userID)
	if err != nil {
		return 0, err
	}

	return item.ID, nil
}

func (s *Service) ensurePaymentBalance(ctx context.Context, userID, walletID int64, current *Payment, amount float64) error {
	if s.wallets == nil {
		return errors.New("wallet service is required")
	}

	item, err := s.wallets.GetByID(ctx, userID, walletID)
	if err != nil {
		return err
	}

	available := item.Balance
	if current != nil && current.WalletID == walletID {
		available += current.Amount
	}

	if amount > available {
		return ErrInsufficientBalance
	}

	return nil
}

func (s *Service) PaymentHistory(ctx context.Context, userID, debtID int64) ([]Payment, error) {
	items, err := s.repo.GetPayments(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}

	return ensurePaymentSlice(items), nil
}

func (s *Service) Installments(ctx context.Context, userID, debtID int64) ([]Installment, error) {
	if err := s.repo.RefreshUserDebtStatuses(ctx, userID); err != nil {
		return nil, err
	}

	items, err := s.repo.GetInstallments(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}

	return ensureInstallmentSlice(items), nil
}

func (s *Service) MarkInstallmentPaid(ctx context.Context, userID, debtID, installmentID int64, paidAt *time.Time) (Installment, error) {
	value := time.Now()
	if paidAt != nil && !paidAt.IsZero() {
		value = *paidAt
	}

	if _, err := s.repo.GetDebtByID(ctx, userID, debtID); err != nil {
		return Installment{}, err
	}

	if err := s.repo.RefreshUserDebtStatuses(ctx, userID); err != nil {
		return Installment{}, err
	}

	return s.repo.MarkInstallmentPaid(ctx, debtID, installmentID, value)
}

func buildDebt(userID int64, input CreateInput) (Debt, error) {
	debt := Debt{
		UserID:             userID,
		Name:               strings.TrimSpace(input.Name),
		TotalAmount:        input.TotalAmount,
		MonthlyInstallment: input.MonthlyInstallment,
		DueDate:            input.DueDate,
		Status:             StatusPending,
	}

	if err := validateDebt(debt.Name, debt.TotalAmount, debt.MonthlyInstallment, debt.DueDate); err != nil {
		return Debt{}, err
	}

	return debt, nil
}

func validateDebt(name string, totalAmount, monthlyInstallment float64, dueDate time.Time) error {
	switch {
	case name == "":
		return errors.New("name is required")
	case totalAmount <= 0:
		return errors.New("total_amount must be greater than zero")
	case monthlyInstallment <= 0:
		return errors.New("monthly_installment must be greater than zero")
	case dueDate.IsZero():
		return errors.New("due_date is required")
	}

	return nil
}

func buildInstallments(debt Debt) ([]Installment, error) {
	if debt.MonthlyInstallment <= 0 {
		return nil, errors.New("monthly_installment must be greater than zero")
	}

	remaining := debt.TotalAmount
	count := int(math.Ceil(debt.TotalAmount / debt.MonthlyInstallment))
	if count < 1 {
		count = 1
	}

	installments := make([]Installment, 0, count)
	for i := 0; i < count; i++ {
		amount := debt.MonthlyInstallment
		if remaining < amount {
			amount = remaining
		}
		dueDate := debt.DueDate.AddDate(0, i, 0)
		installments = append(installments, Installment{
			InstallmentNo: i + 1,
			DueDate:       dueDate,
			Amount:        amount,
			Status:        StatusPending,
		})
		remaining -= amount
		if remaining <= 0 {
			break
		}
	}

	return installments, nil
}

func countInstallmentsByStatus(items []Installment, status Status) int64 {
	var count int64
	for _, item := range items {
		if item.Status == status {
			count++
		}
	}
	return count
}

func ensureDebtSummarySlice(items []DebtSummary) []DebtSummary {
	if items == nil {
		return []DebtSummary{}
	}
	return items
}

func ensureInstallmentSlice(items []Installment) []Installment {
	if items == nil {
		return []Installment{}
	}
	return items
}

func ensurePaymentSlice(items []Payment) []Payment {
	if items == nil {
		return []Payment{}
	}
	return items
}
