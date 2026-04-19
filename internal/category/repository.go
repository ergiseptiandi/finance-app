package category

import "context"

type Repository interface {
	Create(ctx context.Context, item Category) (int64, error)
	GetByID(ctx context.Context, userID, id int64) (Category, error)
	Update(ctx context.Context, item Category) error
	Delete(ctx context.Context, userID, id int64) error
	FindAll(ctx context.Context, userID int64, filter ListFilter) ([]Category, error)
}
