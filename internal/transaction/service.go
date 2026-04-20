package transaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"finance-backend/internal/wallet"
)

var (
	ErrNotFound     = errors.New("transaction not found")
	ErrInvalidInput = errors.New("invalid transaction input")
	nowFunc         = time.Now
)

type Service struct {
	repo    Repository
	wallets wallet.Resolver
}

func NewService(repo Repository, wallets wallet.Resolver) *Service {
	return &Service{repo: repo, wallets: wallets}
}

func (s *Service) Create(ctx context.Context, userID int64, input CreateInput) (Transaction, error) {
	if input.Amount <= 0 {
		return Transaction{}, errors.New("amount must be greater than zero")
	}
	if input.Type != TypeIncome && input.Type != TypeExpense {
		return Transaction{}, errors.New("invalid transaction type")
	}

	walletID, err := s.resolveWalletID(ctx, userID, input.Type, input.WalletID)
	if err != nil {
		return Transaction{}, err
	}

	txn := Transaction{
		UserID:      userID,
		WalletID:    walletID,
		Type:        input.Type,
		Category:    input.Category,
		Amount:      input.Amount,
		Date:        input.Date,
		Description: input.Description,
	}

	id, err := s.repo.Create(ctx, txn)
	if err != nil {
		return Transaction{}, err
	}

	txn.ID = id
	return txn, nil
}

func (s *Service) Get(ctx context.Context, id int64, userID int64) (Transaction, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) Update(ctx context.Context, id int64, userID int64, input UpdateInput) (Transaction, error) {
	txn, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return Transaction{}, err
	}

	if input.Type != nil {
		if *input.Type != TypeIncome && *input.Type != TypeExpense {
			return Transaction{}, errors.New("invalid transaction type")
		}
		txn.Type = *input.Type
	}
	if txn.Type == TypeIncome {
		walletID, err := s.resolveWalletID(ctx, userID, TypeIncome, nil)
		if err != nil {
			return Transaction{}, err
		}
		txn.WalletID = walletID
	} else if input.WalletID != nil {
		walletID, err := s.resolveWalletID(ctx, userID, txn.Type, input.WalletID)
		if err != nil {
			return Transaction{}, err
		}
		txn.WalletID = walletID
	}

	if input.Category != nil {
		txn.Category = *input.Category
	}
	if input.Amount != nil {
		if *input.Amount <= 0 {
			return Transaction{}, errors.New("amount must be greater than zero")
		}
		txn.Amount = *input.Amount
	}
	if input.Date != nil {
		txn.Date = *input.Date
	}
	if input.Description != nil {
		txn.Description = *input.Description
	}

	err = s.repo.Update(ctx, txn)
	if err != nil {
		return Transaction{}, err
	}

	return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) Delete(ctx context.Context, id int64, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *Service) List(ctx context.Context, userID int64, filter ListFilter) (PaginatedList, error) {
	filter, err := s.normalizeListFilter(filter)
	if err != nil {
		return PaginatedList{}, err
	}
	return s.repo.FindAll(ctx, userID, filter)
}

func (s *Service) Summary(ctx context.Context, userID int64, filter ListFilter) (Summary, error) {
	filter, err := s.normalizeListFilter(filter)
	if err != nil {
		return Summary{}, err
	}

	return s.repo.GetSummary(ctx, userID, filter)
}

func (s *Service) normalizeListFilter(filter ListFilter) (ListFilter, error) {
	if (filter.StartDate == nil) != (filter.EndDate == nil) {
		return ListFilter{}, fmt.Errorf("%w: start_date and end_date must be provided together", ErrInvalidInput)
	}

	if filter.StartDate == nil && filter.EndDate == nil {
		now := nowFunc()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0).AddDate(0, 0, -1)

		filter.StartDate = &startOfMonth
		filter.EndDate = &endOfMonth
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.EndDate.Before(*filter.StartDate) {
			return ListFilter{}, fmt.Errorf("%w: end_date must be greater than or equal to start_date", ErrInvalidInput)
		}

		maxEndDate := filter.StartDate.AddDate(0, 2, 0)
		if filter.EndDate.After(maxEndDate) {
			return ListFilter{}, fmt.Errorf("%w: date range cannot exceed 2 months", ErrInvalidInput)
		}
	}

	if filter.Type != nil && *filter.Type != TypeIncome && *filter.Type != TypeExpense {
		return ListFilter{}, fmt.Errorf("%w: invalid transaction type", ErrInvalidInput)
	}

	if filter.WalletID != nil && *filter.WalletID <= 0 {
		return ListFilter{}, fmt.Errorf("%w: wallet_id must be a positive number", ErrInvalidInput)
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 10
	}

	return filter, nil
}

func (s *Service) resolveWalletID(ctx context.Context, userID int64, txnType Type, walletID *int64) (int64, error) {
	if txnType == TypeIncome {
		if s.wallets == nil {
			return 0, errors.New("wallet service is required")
		}

		item, err := s.wallets.DefaultWallet(ctx, userID)
		if err != nil {
			return 0, err
		}

		return item.ID, nil
	}

	if walletID != nil {
		if *walletID <= 0 {
			return 0, errors.New("wallet_id must be a positive number")
		}
		if s.wallets == nil {
			return 0, errors.New("wallet service is required")
		}
		item, err := s.wallets.GetByID(ctx, userID, *walletID)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	}

	if s.wallets == nil {
		return 0, errors.New("wallet service is required")
	}

	item, err := s.wallets.DefaultWallet(ctx, userID)
	if err != nil {
		return 0, err
	}
	return item.ID, nil
}
