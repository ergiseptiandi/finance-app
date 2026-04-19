package wallet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound      = errors.New("wallet not found")
	ErrAlreadyExists = errors.New("wallet already exists")
	ErrConflict      = errors.New("wallet cannot be deleted because it is referenced by history")
	ErrInvalidInput  = errors.New("invalid wallet input")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID int64, input CreateInput) (Wallet, error) {
	item := Wallet{
		UserID:         userID,
		Name:           strings.TrimSpace(input.Name),
		OpeningBalance: input.OpeningBalance,
	}

	if err := validateWallet(item.Name, item.OpeningBalance); err != nil {
		return Wallet{}, err
	}

	id, err := s.repo.Create(ctx, item)
	if err != nil {
		return Wallet{}, err
	}

	return s.repo.GetByID(ctx, userID, id)
}

func (s *Service) GetByID(ctx context.Context, userID, id int64) (Wallet, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *Service) Update(ctx context.Context, userID, id int64, input UpdateInput) (Wallet, error) {
	item, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return Wallet{}, err
	}

	if input.Name != nil {
		item.Name = strings.TrimSpace(*input.Name)
	}
	if input.OpeningBalance != nil {
		item.OpeningBalance = *input.OpeningBalance
	}

	if err := validateWallet(item.Name, item.OpeningBalance); err != nil {
		return Wallet{}, err
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return Wallet{}, err
	}

	return s.repo.GetByID(ctx, userID, id)
}

func (s *Service) Delete(ctx context.Context, userID, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *Service) List(ctx context.Context, userID int64) ([]Wallet, error) {
	items, err := s.repo.FindAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []Wallet{}, nil
	}
	return items, nil
}

func (s *Service) Summary(ctx context.Context, userID int64) (Summary, error) {
	items, err := s.List(ctx, userID)
	if err != nil {
		return Summary{}, err
	}

	total := 0.0
	for _, item := range items {
		total += item.Balance
	}

	return Summary{
		TotalBalance: total,
		Wallets:      items,
	}, nil
}

func (s *Service) TotalBalance(ctx context.Context, userID int64) (float64, error) {
	summary, err := s.Summary(ctx, userID)
	if err != nil {
		return 0, err
	}
	return summary.TotalBalance, nil
}

func (s *Service) DefaultWallet(ctx context.Context, userID int64) (Wallet, error) {
	item, err := s.repo.GetDefaultWallet(ctx, userID)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Wallet{}, err
	}

	createdID, err := s.repo.Create(ctx, Wallet{
		UserID:         userID,
		Name:           defaultWalletName,
		OpeningBalance: 0,
	})
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return s.repo.GetDefaultWallet(ctx, userID)
		}
		return Wallet{}, err
	}

	return s.repo.GetByID(ctx, userID, createdID)
}

func (s *Service) EnsureWallet(ctx context.Context, userID, walletID int64) (Wallet, error) {
	if walletID <= 0 {
		return s.DefaultWallet(ctx, userID)
	}

	return s.repo.GetByID(ctx, userID, walletID)
}

func (s *Service) CreateTransfer(ctx context.Context, userID int64, input CreateTransferInput) (Transfer, error) {
	if input.FromWalletID <= 0 || input.ToWalletID <= 0 {
		return Transfer{}, errors.New("wallet IDs must be positive numbers")
	}
	fromWallet, err := s.repo.GetByID(ctx, userID, input.FromWalletID)
	if err != nil {
		return Transfer{}, err
	}
	toWallet, err := s.repo.GetByID(ctx, userID, input.ToWalletID)
	if err != nil {
		return Transfer{}, err
	}

	if fromWallet.ID == toWallet.ID {
		return Transfer{}, errors.New("from_wallet_id and to_wallet_id must be different")
	}
	if input.Amount <= 0 {
		return Transfer{}, errors.New("amount must be greater than zero")
	}

	transfer := Transfer{
		UserID:       userID,
		FromWalletID: fromWallet.ID,
		ToWalletID:   toWallet.ID,
		Amount:       input.Amount,
		Note:         strings.TrimSpace(input.Note),
		TransferDate: input.TransferDate,
	}
	if transfer.TransferDate.IsZero() {
		transfer.TransferDate = time.Now()
	}

	id, err := s.repo.CreateTransfer(ctx, transfer)
	if err != nil {
		return Transfer{}, err
	}

	return s.repo.GetTransferByID(ctx, userID, id)
}

func (s *Service) ListTransfers(ctx context.Context, userID int64) ([]Transfer, error) {
	items, err := s.repo.FindTransfers(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []Transfer{}, nil
	}
	return items, nil
}

func validateWallet(name string, openingBalance float64) error {
	switch {
	case name == "":
		return errors.New("wallet name is required")
	case openingBalance < 0:
		return errors.New("opening_balance must be greater than or equal to zero")
	}

	return nil
}

func formatWalletError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrAlreadyExists), errors.Is(err, ErrConflict), errors.Is(err, ErrInvalidInput):
		return err
	default:
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
}
