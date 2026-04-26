package budget

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("budget goal not found")
	ErrInvalidInput = errors.New("invalid budget goal input")
)

type Service struct {
	repo     Repository
	expenses ExpenseProvider
}

func NewService(repo Repository, expenses ExpenseProvider) *Service {
	return &Service{repo: repo, expenses: expenses}
}

func (s *Service) Create(ctx context.Context, userID int64, input CreateInput) (Goal, error) {
	if input.CategoryID <= 0 {
		return Goal{}, errors.New("category_id must be a positive number")
	}
	if input.MonthlyAmount <= 0 {
		return Goal{}, errors.New("monthly_amount must be greater than zero")
	}

	cat, err := s.repo.GetCategory(ctx, userID, input.CategoryID)
	if err != nil {
		return Goal{}, err
	}
	if !strings.EqualFold(cat.Type, "expense") {
		return Goal{}, errors.New("budget goals are only supported for expense categories")
	}

	item, err := s.repo.Create(ctx, userID, input.CategoryID, input.MonthlyAmount)
	if err != nil {
		return Goal{}, err
	}
	item.CategoryName = cat.Name
	item.CategoryType = cat.Type
	return item, nil
}

func (s *Service) Update(ctx context.Context, userID, id int64, input UpdateInput) (Goal, error) {
	current, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return Goal{}, err
	}

	if input.CategoryID != nil {
		if *input.CategoryID <= 0 {
			return Goal{}, errors.New("category_id must be a positive number")
		}
		cat, err := s.repo.GetCategory(ctx, userID, *input.CategoryID)
		if err != nil {
			return Goal{}, err
		}
		if !strings.EqualFold(cat.Type, "expense") {
			return Goal{}, errors.New("budget goals are only supported for expense categories")
		}
		current.CategoryID = cat.ID
		current.CategoryName = cat.Name
		current.CategoryType = cat.Type
	}
	if input.MonthlyAmount != nil {
		if *input.MonthlyAmount <= 0 {
			return Goal{}, errors.New("monthly_amount must be greater than zero")
		}
		current.MonthlyAmount = *input.MonthlyAmount
	}

	item, err := s.repo.Update(ctx, userID, current)
	if err != nil {
		return Goal{}, err
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, userID, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *Service) List(ctx context.Context, userID int64, start, end time.Time) ([]Progress, Summary, error) {
	goals, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, Summary{}, err
	}

	spendMap := map[string]float64{}
	if s.expenses != nil && !start.IsZero() && !end.IsZero() {
		spentItems, err := s.expenses.ExpenseByCategory(ctx, userID, start, end)
		if err != nil {
			return nil, Summary{}, err
		}
		for _, item := range spentItems {
			spendMap[strings.ToLower(strings.TrimSpace(item.Category))] += item.Amount
		}
	}

	progress := make([]Progress, 0, len(goals))
	var summary Summary

	for _, goal := range goals {
		current := spendMap[strings.ToLower(strings.TrimSpace(goal.CategoryName))]
		remaining := goal.MonthlyAmount - current
		ratio := 0.0
		if goal.MonthlyAmount > 0 {
			ratio = math.Round((current/goal.MonthlyAmount)*10000) / 100
		}

		status := StatusInactive
		switch {
		case goal.MonthlyAmount <= 0:
			status = StatusInactive
		case current > goal.MonthlyAmount:
			status = StatusOverBudget
		case current >= goal.MonthlyAmount*0.8:
			status = StatusOnTrack
		default:
			status = StatusUnderBudget
		}

		progress = append(progress, Progress{
			Goal:              goal,
			CurrentAmount:     current,
			RemainingAmount:   remaining,
			ProgressPercentage: ratio,
			Status:            status,
		})

		summary.MonthlyBudget += goal.MonthlyAmount
		summary.Spent += current
	}

	summary.MonthlyBudget = math.Round(summary.MonthlyBudget*100) / 100
	summary.Spent = math.Round(summary.Spent*100) / 100
	summary.Remaining = math.Round((summary.MonthlyBudget-summary.Spent)*100) / 100
	if summary.MonthlyBudget > 0 {
		summary.UsageRate = math.Round((summary.Spent/summary.MonthlyBudget)*10000) / 100
	}
	if summary.Spent > summary.MonthlyBudget {
		summary.OverBudgetAmount = math.Round((summary.Spent-summary.MonthlyBudget)*100) / 100
		summary.IsOverBudget = true
	}

	return progress, summary, nil
}
