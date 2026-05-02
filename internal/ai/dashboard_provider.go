package ai

import (
	"context"
	"time"

	"finance-backend/internal/dashboard"
)

type DashboardDataProvider struct {
	dashboardService *dashboard.Service
}

func NewDashboardDataProvider(dashboardService *dashboard.Service) *DashboardDataProvider {
	return &DashboardDataProvider{dashboardService: dashboardService}
}

func (p *DashboardDataProvider) GetFinancialSummary(ctx context.Context, userID int64) (FinancialSummary, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)

	filter := dashboard.DashboardFilter{
		StartDate: &start,
		EndDate:   &end,
	}

	summary, err := p.dashboardService.Summary(ctx, userID, filter)
	if err != nil {
		return FinancialSummary{}, err
	}

	categoryBreakdown, _ := p.dashboardService.CategoryBreakdown(ctx, userID, filter)
	categories := make([]CategorySummary, 0, len(categoryBreakdown))
	for _, cat := range categoryBreakdown {
		categories = append(categories, CategorySummary{
			Category:   cat.Category,
			Amount:     cat.Amount,
			Percentage: cat.Percentage,
		})
	}

	budgetUsage := 0.0
	budgetRemaining := 0.0
	if summary.BudgetSummary != nil {
		budgetUsage = summary.BudgetSummary.UsageRate
		budgetRemaining = summary.BudgetSummary.Remaining
	}

	return FinancialSummary{
		TotalBalance:       summary.TotalBalance,
		MonthlyIncome:      summary.MonthlyIncome,
		MonthlyExpense:     summary.MonthlyExpense,
		ConsumptionExpense: summary.ConsumptionExpense,
		DebtRepayment:      summary.DebtRepayment,
		NetCashflow:        summary.NetCashflow,
		SavingsRate:        summary.SavingsRate,
		ExpenseRatio:       summary.ExpenseRatio,
		DebtTotal:          summary.Debt.TotalDebt,
		DebtRemaining:      summary.Debt.RemainingDebt,
		DebtCompletionRate: summary.Debt.CompletionRate,
		BudgetUsage:        budgetUsage,
		BudgetRemaining:    budgetRemaining,
		CategoryBreakdown:  categories,
		RecentTransactions: int(summary.Debt.TotalDebtCount),
	}, nil
}
