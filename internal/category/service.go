package category

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrNotFound      = errors.New("category not found")
	ErrAlreadyExists = errors.New("category already exists")
	ErrInvalidInput  = errors.New("invalid category input")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Category, error) {
	item := Category{
		Name: strings.TrimSpace(input.Name),
		Type: input.Type,
	}

	if err := validateCategory(item.Name, item.Type); err != nil {
		return Category{}, err
	}

	id, err := s.repo.Create(ctx, item)
	if err != nil {
		return Category{}, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (Category, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Category{}, err
	}

	if input.Name != nil {
		item.Name = strings.TrimSpace(*input.Name)
	}
	if input.Type != nil {
		item.Type = *input.Type
	}

	if err := validateCategory(item.Name, item.Type); err != nil {
		return Category{}, err
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return Category{}, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Category, error) {
	if filter.Type != nil && !isValidType(*filter.Type) {
		return nil, errors.New("invalid category type")
	}

	return s.repo.FindAll(ctx, filter)
}

func validateCategory(name string, categoryType Type) error {
	if name == "" {
		return errors.New("category name is required")
	}
	if !isValidType(categoryType) {
		return errors.New("invalid category type")
	}

	return nil
}

func isValidType(categoryType Type) bool {
	return categoryType == TypeIncome || categoryType == TypeExpense
}
