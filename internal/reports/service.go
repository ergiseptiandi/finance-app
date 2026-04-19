package reports

import (
	"context"
	"math"
	"sort"
	"time"

	"finance-backend/internal/wallet"
)

type Service struct {
	repo     Repository
	balances wallet.BalanceProvider
}

func NewService(repo Repository, balances wallet.BalanceProvider) *Service {
	return &Service{repo: repo, balances: balances}
}

func (s *Service) ExpenseByCategory(ctx context.Context, userID int64) ([]CategoryExpense, error) {
	start, end := currentMonthRange()

	items, err := s.repo.ExpenseByCategory(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	total := 0.0
	for _, item := range items {
		total += item.Amount
	}

	for i := range items {
		if total > 0 {
			items[i].Percentage = round2((items[i].Amount / total) * 100)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Amount == items[j].Amount {
			return items[i].Category < items[j].Category
		}
		return items[i].Amount > items[j].Amount
	})

	return items, nil
}

func (s *Service) SpendingTrends(ctx context.Context, userID int64) ([]TrendPoint, error) {
	end := startOfMonth(time.Now()).AddDate(0, 1, 0)
	start := startOfMonth(time.Now()).AddDate(0, -11, 0)

	values, err := s.repo.ExpenseTrendByMonth(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	points := make([]TrendPoint, 0, 12)
	for month := start; month.Before(end); month = month.AddDate(0, 1, 0) {
		key := month.Format("2006-01")
		points = append(points, TrendPoint{
			Period: key,
			Amount: values[key],
		})
	}

	return points, nil
}

func (s *Service) HighestSpendingCategory(ctx context.Context, userID int64) (CategoryExpense, error) {
	items, err := s.ExpenseByCategory(ctx, userID)
	if err != nil {
		return CategoryExpense{}, err
	}
	if len(items) == 0 {
		return CategoryExpense{}, nil
	}

	return items[0], nil
}

func (s *Service) AverageDailySpending(ctx context.Context, userID int64) (AverageDailySpending, error) {
	now := time.Now()
	start, end := currentMonthRange()
	monthExpense, err := s.repo.ExpenseBetween(ctx, userID, start, end)
	if err != nil {
		return AverageDailySpending{}, err
	}

	daysCount := daysElapsedInMonth(now)
	average := 0.0
	if daysCount > 0 {
		average = round2(monthExpense / float64(daysCount))
	}

	return AverageDailySpending{
		TotalExpense:         monthExpense,
		DaysCount:            daysCount,
		AverageDailySpending: average,
	}, nil
}

func (s *Service) RemainingBalance(ctx context.Context, userID int64) (RemainingBalance, error) {
	income, err := s.repo.AllTimeIncome(ctx, userID)
	if err != nil {
		return RemainingBalance{}, err
	}

	expense, err := s.repo.AllTimeExpense(ctx, userID)
	if err != nil {
		return RemainingBalance{}, err
	}

	remainingBalance := income - expense
	if s.balances != nil {
		if balance, err := s.balances.TotalBalance(ctx, userID); err == nil {
			remainingBalance = balance
		} else {
			return RemainingBalance{}, err
		}
	}

	return RemainingBalance{
		TotalIncome:      income,
		TotalExpense:     expense,
		RemainingBalance: remainingBalance,
	}, nil
}

func currentMonthRange() (time.Time, time.Time) {
	now := time.Now()
	start := startOfMonth(now)
	return start, start.AddDate(0, 1, 0)
}

func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func daysElapsedInMonth(t time.Time) int {
	start := startOfMonth(t)
	end := start.AddDate(0, 0, 1)
	count := 0
	for day := start; !day.After(t); day = day.AddDate(0, 0, 1) {
		count++
		if day.Equal(end) {
			break
		}
	}
	if count < 1 {
		return 1
	}
	return count
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
