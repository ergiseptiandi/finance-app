package category

import (
	"context"
	"testing"
	"time"
)

type stubRepository struct {
	items  map[int64]Category
	nextID int64
}

func newStubRepository() *stubRepository {
	return &stubRepository{
		items:  map[int64]Category{},
		nextID: 1,
	}
}

func (r *stubRepository) Create(_ context.Context, item Category) (int64, error) {
	for _, existing := range r.items {
		if existing.Name == item.Name && existing.Type == item.Type {
			return 0, ErrAlreadyExists
		}
	}

	item.ID = r.nextID
	item.CreatedAt = time.Now()
	item.UpdatedAt = item.CreatedAt
	r.items[item.ID] = item
	r.nextID++
	return item.ID, nil
}

func (r *stubRepository) GetByID(_ context.Context, id int64) (Category, error) {
	item, ok := r.items[id]
	if !ok {
		return Category{}, ErrNotFound
	}
	return item, nil
}

func (r *stubRepository) Update(_ context.Context, item Category) error {
	if _, ok := r.items[item.ID]; !ok {
		return ErrNotFound
	}

	for id, existing := range r.items {
		if id != item.ID && existing.Name == item.Name && existing.Type == item.Type {
			return ErrAlreadyExists
		}
	}

	createdAt := r.items[item.ID].CreatedAt
	item.CreatedAt = createdAt
	item.UpdatedAt = time.Now()
	r.items[item.ID] = item
	return nil
}

func (r *stubRepository) Delete(_ context.Context, id int64) error {
	if _, ok := r.items[id]; !ok {
		return ErrNotFound
	}

	delete(r.items, id)
	return nil
}

func (r *stubRepository) FindAll(_ context.Context, filter ListFilter) ([]Category, error) {
	items := make([]Category, 0, len(r.items))
	for _, item := range r.items {
		if filter.Type != nil && item.Type != *filter.Type {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func TestServiceCreateTrimsName(t *testing.T) {
	svc := NewService(newStubRepository())

	item, err := svc.Create(context.Background(), CreateInput{
		Name: "  Salary  ",
		Type: TypeIncome,
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got error: %v", err)
	}

	if item.Name != "Salary" {
		t.Fatalf("expected trimmed name Salary, got %q", item.Name)
	}
}

func TestServiceCreateRejectsInvalidType(t *testing.T) {
	svc := NewService(newStubRepository())

	_, err := svc.Create(context.Background(), CreateInput{
		Name: "Salary",
		Type: "other",
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestServiceListFiltersType(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)

	if _, err := svc.Create(context.Background(), CreateInput{Name: "Salary", Type: TypeIncome}); err != nil {
		t.Fatalf("failed to create income category: %v", err)
	}
	if _, err := svc.Create(context.Background(), CreateInput{Name: "Food", Type: TypeExpense}); err != nil {
		t.Fatalf("failed to create expense category: %v", err)
	}

	categoryType := TypeExpense
	items, err := svc.List(context.Background(), ListFilter{Type: &categoryType})
	if err != nil {
		t.Fatalf("expected list to succeed, got error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "Food" {
		t.Fatalf("expected Food category, got %q", items[0].Name)
	}
}
