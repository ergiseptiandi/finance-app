package transaction

import "context"

type Repository interface {
	Create(ctx context.Context, txn Transaction) (int64, error)
	GetByID(ctx context.Context, id int64, userID int64) (Transaction, error)
	Update(ctx context.Context, txn Transaction) error
	Delete(ctx context.Context, id int64, userID int64) error
	FindAll(ctx context.Context, userID int64, filter ListFilter) (PaginatedList, error)
	GetSummary(ctx context.Context, userID int64) (Summary, error)
}
