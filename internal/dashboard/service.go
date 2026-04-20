package dashboard

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
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

func (s *Service) BudgetVsActual(ctx context.Context, userID int64, filter DashboardFilter, budgetAmount *float64) (BudgetVsActual, error) {
	summary, err := s.Summary(ctx, userID, filter)
	if err != nil {
		return BudgetVsActual{}, err
	}

	budget := summary.MonthlyIncome
	if budgetAmount != nil {
		if *budgetAmount <= 0 {
			return BudgetVsActual{}, fmt.Errorf("%w: budget_amount must be greater than zero", ErrInvalidInput)
		}
		budget = *budgetAmount
	}
	if budget <= 0 {
		budget = summary.MonthlyExpense
	}
	if budget <= 0 {
		budget = summary.TotalBalance
	}
	if budget <= 0 {
		budget = summary.MonthlyExpense
	}
	if budget <= 0 {
		budget = 1
	}

	actual := summary.MonthlyExpense
	remaining := budget - actual
	overBudget := 0.0
	status := "on_track"
	switch {
	case remaining < 0:
		overBudget = math.Abs(remaining)
		status = "over_budget"
	case percentageOf(actual, budget) >= 90:
		status = "warning"
	case budgetAmount == nil && summary.MonthlyIncome <= 0:
		status = "unknown"
	}

	return BudgetVsActual{
		BudgetAmount:     budget,
		ActualSpent:      actual,
		RemainingBudget:  math.Round(remaining*100) / 100,
		UsageRate:        percentageOf(actual, budget),
		OverBudgetAmount: math.Round(overBudget*100) / 100,
		Status:           status,
	}, nil
}

func (s *Service) CategoryBreakdown(ctx context.Context, userID int64, filter DashboardFilter) ([]CategoryBreakdownItem, error) {
	start, end, err := s.resolveRange(filter)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ExpenseByCategory(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	total := 0.0
	for _, item := range items {
		total += item.Amount
	}

	for i := range items {
		items[i].Percentage = percentageOf(items[i].Amount, total)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Amount == items[j].Amount {
			return items[i].Category < items[j].Category
		}
		return items[i].Amount > items[j].Amount
	})

	if items == nil {
		return []CategoryBreakdownItem{}, nil
	}

	return items, nil
}

func (s *Service) UpcomingBills(ctx context.Context, userID int64, days int) ([]UpcomingBill, error) {
	if days <= 0 {
		return nil, fmt.Errorf("%w: days must be greater than zero", ErrInvalidInput)
	}
	if days > 365 {
		return nil, fmt.Errorf("%w: days cannot exceed 365", ErrInvalidInput)
	}

	now := nowFunc()
	start := startOfDay(now)
	end := start.AddDate(0, 0, days)

	items, err := s.repo.UpcomingBills(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	if items == nil {
		return []UpcomingBill{}, nil
	}

	return items, nil
}

func (s *Service) TopMerchants(ctx context.Context, userID int64, filter DashboardFilter) ([]TopMerchant, error) {
	start, end, err := s.resolveRange(filter)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.TopMerchants(ctx, userID, start, end, 5)
	if err != nil {
		return nil, err
	}

	if items == nil {
		return []TopMerchant{}, nil
	}

	return items, nil
}

func (s *Service) Insights(ctx context.Context, userID int64, filter DashboardFilter) ([]Insight, error) {
	summary, err := s.Summary(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	budget, err := s.BudgetVsActual(ctx, userID, filter, nil)
	if err != nil {
		return nil, err
	}

	categoryBreakdown, err := s.CategoryBreakdown(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	upcomingBills, err := s.UpcomingBills(ctx, userID, 30)
	if err != nil {
		return nil, err
	}

	insights := make([]Insight, 0, 6)

	switch {
	case budget.OverBudgetAmount > 0:
		insights = append(insights, Insight{
			Type:        "budget_warning",
			Code:        "OVER_BUDGET",
			Title:       "Budget terlampaui",
			Message:     fmt.Sprintf("Pengeluaran melebihi budget sebesar %.2f.", budget.OverBudgetAmount),
			Severity:    "danger",
			ChangeValue: budget.OverBudgetAmount,
		})
	case budget.UsageRate >= 90:
		insights = append(insights, Insight{
			Type:        "budget_warning",
			Code:        "BUDGET_NEAR_LIMIT",
			Title:       "Budget hampir habis",
			Message:     fmt.Sprintf("Pengeluaran sudah mencapai %.2f%% dari budget periode ini.", budget.UsageRate),
			Severity:    "warning",
			ChangeValue: budget.UsageRate,
		})
	}

	if summary.SavingsRate > 0 && summary.SavingsRate < 20 {
		insights = append(insights, Insight{
			Type:        "low_savings_rate",
			Code:        "LOW_SAVINGS_RATE",
			Title:       "Savings rate rendah",
			Message:     fmt.Sprintf("Savings rate saat ini hanya %.2f%%.", summary.SavingsRate),
			Severity:    "warning",
			ChangeValue: summary.SavingsRate,
		})
	}

	if summary.Debt.OverdueDebtCount > 0 || summary.Debt.OverdueInstallments > 0 {
		insights = append(insights, Insight{
			Type:        "overdue_bill",
			Code:        "OVERDUE_DEBT",
			Title:       "Ada tagihan terlambat",
			Message:     fmt.Sprintf("Terdapat %d utang dan %d cicilan yang sudah overdue.", summary.Debt.OverdueDebtCount, summary.Debt.OverdueInstallments),
			Severity:    "danger",
			ChangeValue: float64(summary.Debt.OverdueInstallments),
		})
	}

	if len(upcomingBills) > 0 {
		insights = append(insights, Insight{
			Type:        "upcoming_bill",
			Code:        "UPCOMING_BILL",
			Title:       "Tagihan akan segera jatuh tempo",
			Message:     fmt.Sprintf("Ada %d tagihan yang jatuh tempo dalam 30 hari ke depan.", len(upcomingBills)),
			Severity:    "info",
			ChangeValue: float64(len(upcomingBills)),
		})
	}

	if len(categoryBreakdown) > 0 && categoryBreakdown[0].Percentage >= 50 {
		insights = append(insights, Insight{
			Type:        "category_concentration",
			Code:        "CATEGORY_CONCENTRATION",
			Title:       "Belanja terkonsentrasi",
			Message:     fmt.Sprintf("%s menyumbang %.2f%% dari total pengeluaran periode ini.", categoryBreakdown[0].Category, categoryBreakdown[0].Percentage),
			Severity:    "info",
			ChangeValue: categoryBreakdown[0].Percentage,
		})
	}

	if len(insights) == 0 {
		insights = append(insights, Insight{
			Type:        "healthy_status",
			Code:        "NO_ALERTS",
			Title:       "Tidak ada insight kritis",
			Message:     "Dashboard tidak menemukan kondisi yang perlu mendapat perhatian khusus pada periode ini.",
			Severity:    "info",
			ChangeValue: 0,
		})
	}

	return insights, nil
}

func (s *Service) GoalsProgress(ctx context.Context, userID int64, filter DashboardFilter) ([]GoalProgress, error) {
	summary, err := s.Summary(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	goals := make([]GoalProgress, 0, 2)

	emergencyTarget := summary.MonthlyExpense * 3
	if emergencyTarget <= 0 {
		emergencyTarget = summary.MonthlyIncome * 3
	}
	if emergencyTarget <= 0 {
		emergencyTarget = 1
	}
	emergencyProgress := percentageOf(summary.TotalBalance, emergencyTarget)
	emergencyStatus := "building"
	if emergencyProgress >= 100 {
		emergencyStatus = "funded"
	}
	goals = append(goals, GoalProgress{
		Name:               "Emergency Fund",
		TargetAmount:       emergencyTarget,
		CurrentAmount:      summary.TotalBalance,
		ProgressPercentage: emergencyProgress,
		Status:             emergencyStatus,
	})

	debtTarget := summary.Debt.TotalDebt
	if debtTarget > 0 {
		debtProgress := percentageOf(summary.Debt.PaidDebt, debtTarget)
		debtStatus := "active"
		if summary.Debt.RemainingDebt <= 0 {
			debtStatus = "completed"
		}
		goals = append(goals, GoalProgress{
			Name:               "Debt Freedom",
			TargetAmount:       debtTarget,
			CurrentAmount:      summary.Debt.PaidDebt,
			ProgressPercentage: debtProgress,
			Status:             debtStatus,
		})
	}

	if goals == nil {
		return []GoalProgress{}, nil
	}

	return goals, nil
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
