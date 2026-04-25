package wallet

import "context"

type Repository interface {
	Create(ctx context.Context, item Wallet) (int64, error)
	GetByID(ctx context.Context, userID, id int64) (Wallet, error)
	GetByName(ctx context.Context, userID int64, name string) (Wallet, error)
	GetDefaultWallet(ctx context.Context, userID int64) (Wallet, error)
	Update(ctx context.Context, item Wallet) error
	Archive(ctx context.Context, userID, id int64) error
	Delete(ctx context.Context, userID, id int64) error
	FindAll(ctx context.Context, userID int64) ([]Wallet, error)
	CreateTransfer(ctx context.Context, transfer Transfer) (int64, error)
	GetTransferByID(ctx context.Context, userID, id int64) (Transfer, error)
	FindTransfers(ctx context.Context, userID int64) ([]Transfer, error)
}

type Resolver interface {
	GetByID(ctx context.Context, userID, id int64) (Wallet, error)
	DefaultWallet(ctx context.Context, userID int64) (Wallet, error)
}

type BalanceProvider interface {
	TotalBalance(ctx context.Context, userID int64) (float64, error)
}
