package dashboard

import (
	"context"
	"math"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Summary(ctx context.Context, userID int64) (Summary, error) {
	now := time.Now()
	monthStart := startOfMonth(now)
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	allIncome, err := s.repo.AllTimeIncome(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	allExpense, err := s.repo.AllTimeExpense(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	monthlyIncome, err := s.repo.IncomeBetween(ctx, userID, monthStart, nextMonthStart)
	if err != nil {
		return Summary{}, err
	}
	monthlyExpense, err := s.repo.ExpenseBetween(ctx, userID, monthStart, nextMonthStart)
	if err != nil {
		return Summary{}, err
	}

	return Summary{
		TotalBalance:   allIncome - allExpense,
		MonthlyIncome:  monthlyIncome,
		MonthlyExpense: monthlyExpense,
	}, nil
}

func (s *Service) DailySpending(ctx context.Context, userID int64) ([]SpendingPoint, error) {
	now := time.Now()
	monthStart := startOfMonth(now)
	todayStart := startOfDay(now)
	tomorrowStart := todayStart.AddDate(0, 0, 1)

	values, err := s.repo.ExpenseByDay(ctx, userID, monthStart, tomorrowStart)
	if err != nil {
		return nil, err
	}

	points := make([]SpendingPoint, 0)
	for day := monthStart; day.Before(tomorrowStart); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		if day.After(todayStart) {
			break
		}
		points = append(points, SpendingPoint{
			Date:   key,
			Amount: values[key],
		})
	}

	return points, nil
}

func (s *Service) MonthlySpending(ctx context.Context, userID int64) ([]MonthlySpendingPoint, error) {
	now := time.Now()
	end := startOfMonth(now).AddDate(0, 1, 0)
	start := startOfMonth(now).AddDate(0, -11, 0)

	values, err := s.repo.ExpenseByMonth(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	points := make([]MonthlySpendingPoint, 0, 12)
	for month := start; month.Before(end); month = month.AddDate(0, 1, 0) {
		key := month.Format("2006-01")
		points = append(points, MonthlySpendingPoint{
			Month:  key,
			Amount: values[key],
		})
	}

	return points, nil
}

func (s *Service) Comparison(ctx context.Context, userID int64) (Comparison, error) {
	now := time.Now()
	todayStart := startOfDay(now)
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	currentMonthStart := startOfMonth(now)
	nextMonthStart := currentMonthStart.AddDate(0, 1, 0)
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)

	todayExpense, err := s.repo.ExpenseBetween(ctx, userID, todayStart, tomorrowStart)
	if err != nil {
		return Comparison{}, err
	}
	yesterdayExpense, err := s.repo.ExpenseBetween(ctx, userID, yesterdayStart, todayStart)
	if err != nil {
		return Comparison{}, err
	}
	currentMonthExpense, err := s.repo.ExpenseBetween(ctx, userID, currentMonthStart, nextMonthStart)
	if err != nil {
		return Comparison{}, err
	}
	lastMonthExpense, err := s.repo.ExpenseBetween(ctx, userID, lastMonthStart, currentMonthStart)
	if err != nil {
		return Comparison{}, err
	}

	return Comparison{
		TodayVsYesterday: ComparisonMetric{
			Current:          todayExpense,
			Previous:         yesterdayExpense,
			Difference:       todayExpense - yesterdayExpense,
			PercentageChange: percentageChange(todayExpense, yesterdayExpense),
		},
		ThisMonthVsLastMonth: ComparisonMetric{
			Current:          currentMonthExpense,
			Previous:         lastMonthExpense,
			Difference:       currentMonthExpense - lastMonthExpense,
			PercentageChange: percentageChange(currentMonthExpense, lastMonthExpense),
		},
	}, nil
}

func (s *Service) ExpenseVsSalary(ctx context.Context, userID int64) (ExpenseVsSalary, error) {
	now := time.Now()
	monthStart := startOfMonth(now)
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	expense, err := s.repo.ExpenseBetween(ctx, userID, monthStart, nextMonthStart)
	if err != nil {
		return ExpenseVsSalary{}, err
	}
	salaryAmount, err := s.repo.LatestSalaryAmount(ctx, userID)
	if err != nil {
		return ExpenseVsSalary{}, err
	}

	percentage := 0.0
	if salaryAmount > 0 {
		percentage = math.Round((expense/salaryAmount)*10000) / 100
	}

	return ExpenseVsSalary{
		MonthlyExpense: expense,
		CurrentSalary:  salaryAmount,
		Percentage:     percentage,
	}, nil
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func percentageChange(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}

	return math.Round(((current-previous)/previous)*10000) / 100
}
