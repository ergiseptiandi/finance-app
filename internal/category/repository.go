package category

import "context"

type Repository interface {
	Create(ctx context.Context, item Category) (int64, error)
	GetByID(ctx context.Context, id int64) (Category, error)
	Update(ctx context.Context, item Category) error
	Delete(ctx context.Context, id int64) error
	FindAll(ctx context.Context, filter ListFilter) ([]Category, error)
}
