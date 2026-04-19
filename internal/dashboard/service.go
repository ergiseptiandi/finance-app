package dashboard

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"finance-backend/internal/wallet"
)

var (
	ErrNotFound     = errors.New("dashboard data not found")
	ErrInvalidInput = errors.New("invalid dashboard input")
)

var nowFunc = time.Now

type Service struct {
	repo     Repository
	balances wallet.BalanceProvider
}

func NewService(repo Repository, balances wallet.BalanceProvider) *Service {
	return &Service{repo: repo, balances: balances}
}

func (s *Service) Summary(ctx context.Context, userID int64, filter DashboardFilter) (Summary, error) {
	start, end, err := s.resolveRange(filter)
	if err != nil {
		return Summary{}, err
	}

	if err := s.repo.RefreshUserDebtStatuses(ctx, userID); err != nil {
		return Summary{}, err
	}

	allIncome, err := s.repo.AllTimeIncome(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	allExpense, err := s.repo.AllTimeExpense(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	monthlyIncome, err := s.repo.IncomeBetween(ctx, userID, start, end)
	if err != nil {
		return Summary{}, err
	}
	monthlyExpense, err := s.repo.ExpenseBetween(ctx, userID, start, end)
	if err != nil {
		return Summary{}, err
	}

	debtOverview, err := s.repo.DebtOverview(ctx, userID, start, end)
	if err != nil {
		return Summary{}, err
	}

	totalBalance := allIncome - allExpense
	if s.balances != nil {
		if balance, err := s.balances.TotalBalance(ctx, userID); err == nil {
			totalBalance = balance
		} else {
			return Summary{}, err
		}
	}

	netCashflow := monthlyIncome - monthlyExpense

	debtOverview.DebtToIncomeRatio = percentageOf(debtOverview.RemainingDebt, monthlyIncome)
	debtOverview.DebtToBalanceRatio = percentageOf(debtOverview.RemainingDebt, totalBalance)
	debtOverview.CompletionRate = percentageOf(debtOverview.PaidDebt, debtOverview.TotalDebt)

	return Summary{
		TotalBalance:   totalBalance,
		MonthlyIncome:  monthlyIncome,
		MonthlyExpense: monthlyExpense,
		NetCashflow:    netCashflow,
		SavingsRate:    percentageOf(netCashflow, monthlyIncome),
		ExpenseRatio:   percentageOf(monthlyExpense, monthlyIncome),
		Debt:           debtOverview,
	}, nil
}

func (s *Service) DailySpending(ctx context.Context, userID int64, filter DashboardFilter) ([]SpendingPoint, error) {
	start, end, err := s.resolveRange(filter)
	if err != nil {
		return nil, err
	}

	now := nowFunc()
	todayStart := startOfDay(now)
	visibleEnd := end
	todayEnd := todayStart.AddDate(0, 0, 1)
	if visibleEnd.After(todayEnd) {
		visibleEnd = todayEnd
	}

	values, err := s.repo.ExpenseByDay(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	points := make([]SpendingPoint, 0)
	for day := start; day.Before(visibleEnd); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		points = append(points, SpendingPoint{
			Date:   key,
			Amount: values[key],
		})
	}

	return points, nil
}

func (s *Service) MonthlySpending(ctx context.Context, userID int64, filter DashboardFilter) ([]MonthlySpendingPoint, error) {
	start, end, err := s.resolveRange(filter)
	if err != nil {
		return nil, err
	}

	values, err := s.repo.ExpenseByMonth(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	monthLimit := startOfMonth(end.AddDate(0, 0, -1)).AddDate(0, 1, 0)
	points := make([]MonthlySpendingPoint, 0)
	for month := startOfMonth(start); month.Before(monthLimit); month = month.AddDate(0, 1, 0) {
		key := month.Format("2006-01")
		points = append(points, MonthlySpendingPoint{
			Month:  key,
			Amount: values[key],
		})
	}

	return points, nil
}

func (s *Service) Comparison(ctx context.Context, userID int64) (Comparison, error) {
	now := nowFunc()
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

func (s *Service) resolveRange(filter DashboardFilter) (time.Time, time.Time, error) {
	if (filter.StartDate == nil) != (filter.EndDate == nil) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: start_date and end_date must be provided together", ErrInvalidInput)
	}

	if filter.StartDate == nil && filter.EndDate == nil {
		now := nowFunc()
		start := startOfMonth(now)
		end := start.AddDate(0, 1, 0)
		return start, end, nil
	}

	if filter.EndDate.Before(*filter.StartDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: end_date must be greater than or equal to start_date", ErrInvalidInput)
	}

	maxEndDate := filter.StartDate.AddDate(0, 3, 0)
	if filter.EndDate.After(maxEndDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: date range cannot exceed 3 months", ErrInvalidInput)
	}

	start := startOfDay(*filter.StartDate)
	end := startOfDay(filter.EndDate.AddDate(0, 0, 1))
	return start, end, nil
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

func percentageOf(part, whole float64) float64 {
	if whole <= 0 {
		return 0
	}

	return math.Round((part/whole)*10000) / 100
}
