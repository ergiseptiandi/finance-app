package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"finance-backend/internal/wallet"
)

type transactionRepoStub struct {
	current    Transaction
	hasUpdated bool
	updatedTxn Transaction
}

func (r *transactionRepoStub) Create(ctx context.Context, txn Transaction) (int64, error) {
	return 1, nil
}

func (r *transactionRepoStub) GetByID(ctx context.Context, id int64, userID int64) (Transaction, error) {
	if r.hasUpdated {
		return r.updatedTxn, nil
	}
	return r.current, nil
}

func (r *transactionRepoStub) Update(ctx context.Context, txn Transaction) error {
	r.updatedTxn = txn
	r.hasUpdated = true
	return nil
}

func (r *transactionRepoStub) Delete(ctx context.Context, id int64, userID int64) error {
	return nil
}

func (r *transactionRepoStub) FindAll(ctx context.Context, userID int64, filter ListFilter) (PaginatedList, error) {
	return PaginatedList{}, nil
}

func (r *transactionRepoStub) GetSummary(ctx context.Context, userID int64, filter ListFilter) (Summary, error) {
	return Summary{}, nil
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

func TestCreateRejectsExpenseWhenWalletBalanceTooLow(t *testing.T) {
	t.Parallel()

	svc := NewService(&transactionRepoStub{}, walletResolverStub{
		getByIDFn: func(context.Context, int64, int64) (wallet.Wallet, error) {
			return wallet.Wallet{ID: 7, Balance: 100}, nil
		},
	})

	amount := 150.0
	walletID := int64(7)
	_, err := svc.Create(context.Background(), 2, CreateInput{
		WalletID:    &walletID,
		Type:        TypeExpense,
		Category:    "Food",
		Amount:      amount,
		Date:        time.Date(2026, time.April, 21, 10, 0, 0, 0, time.UTC),
		Description: "Lunch",
	})
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestUpdateExpenseUsesOriginalAmountWhenValidatingBalance(t *testing.T) {
	t.Parallel()

	repo := &transactionRepoStub{
		current: Transaction{
			ID:          1,
			UserID:      2,
			WalletID:    7,
			Type:        TypeExpense,
			Category:    "Food",
			Amount:      40,
			Date:        time.Date(2026, time.April, 21, 10, 0, 0, 0, time.UTC),
			Description: "Lunch",
		},
	}

	svc := NewService(repo, walletResolverStub{
		getByIDFn: func(context.Context, int64, int64) (wallet.Wallet, error) {
			return wallet.Wallet{ID: 7, Balance: 100}, nil
		},
	})

	amount := 130.0
	item, err := svc.Update(context.Background(), 1, 2, UpdateInput{Amount: &amount})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if item.Amount != 130 {
		t.Fatalf("unexpected amount: %v", item.Amount)
	}
}
