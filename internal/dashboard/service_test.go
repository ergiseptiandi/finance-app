package dashboard

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"finance-backend/internal/alerts"
)

type dashboardRepoStub struct {
	refreshUserDebtStatusesFn func(context.Context, int64) error
	allTimeIncomeFn           func(context.Context, int64) (float64, error)
	allTimeExpenseFn          func(context.Context, int64) (float64, error)
	incomeBetweenFn           func(context.Context, int64, time.Time, time.Time) (float64, error)
	expenseBetweenFn          func(context.Context, int64, time.Time, time.Time) (float64, error)
	expenseByDayFn            func(context.Context, int64, time.Time, time.Time) (map[string]float64, error)
	expenseByMonthFn          func(context.Context, int64, time.Time, time.Time) (map[string]float64, error)
	debtOverviewFn            func(context.Context, int64, time.Time, time.Time) (DebtOverview, error)
	expenseByCategoryFn       func(context.Context, int64, time.Time, time.Time) ([]CategoryBreakdownItem, error)
	upcomingBillsFn           func(context.Context, int64, time.Time, time.Time) ([]UpcomingBill, error)
	topMerchantsFn            func(context.Context, int64, time.Time, time.Time, int) ([]TopMerchant, error)
}

type balanceProviderStub struct {
	totalBalanceFn func(context.Context, int64) (float64, error)
}

type alertSourceStub struct {
	listFn func(context.Context, int64, alerts.AlertListFilter) ([]alerts.Alert, error)
}

func (b balanceProviderStub) TotalBalance(ctx context.Context, userID int64) (float64, error) {
	if b.totalBalanceFn != nil {
		return b.totalBalanceFn(ctx, userID)
	}
	return 0, nil
}

func (a alertSourceStub) List(ctx context.Context, userID int64, filter alerts.AlertListFilter) ([]alerts.Alert, error) {
	if a.listFn != nil {
		return a.listFn(ctx, userID, filter)
	}
	return []alerts.Alert{}, nil
}

func (r dashboardRepoStub) RefreshUserDebtStatuses(ctx context.Context, userID int64) error {
	if r.refreshUserDebtStatusesFn != nil {
		return r.refreshUserDebtStatusesFn(ctx, userID)
	}
	return nil
}

func (r dashboardRepoStub) AllTimeIncome(ctx context.Context, userID int64) (float64, error) {
	if r.allTimeIncomeFn != nil {
		return r.allTimeIncomeFn(ctx, userID)
	}
	return 0, nil
}

func (r dashboardRepoStub) AllTimeExpense(ctx context.Context, userID int64) (float64, error) {
	if r.allTimeExpenseFn != nil {
		return r.allTimeExpenseFn(ctx, userID)
	}
	return 0, nil
}

func (r dashboardRepoStub) IncomeBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	if r.incomeBetweenFn != nil {
		return r.incomeBetweenFn(ctx, userID, start, end)
	}
	return 0, nil
}

func (r dashboardRepoStub) ExpenseBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	if r.expenseBetweenFn != nil {
		return r.expenseBetweenFn(ctx, userID, start, end)
	}
	return 0, nil
}

func (r dashboardRepoStub) ExpenseByDay(ctx context.Context, userID int64, start, end time.Time) (map[string]float64, error) {
	if r.expenseByDayFn != nil {
		return r.expenseByDayFn(ctx, userID, start, end)
	}
	return map[string]float64{}, nil
}

func (r dashboardRepoStub) ExpenseByMonth(ctx context.Context, userID int64, start, end time.Time) (map[string]float64, error) {
	if r.expenseByMonthFn != nil {
		return r.expenseByMonthFn(ctx, userID, start, end)
	}
	return map[string]float64{}, nil
}

func (r dashboardRepoStub) DebtOverview(ctx context.Context, userID int64, start, end time.Time) (DebtOverview, error) {
	if r.debtOverviewFn != nil {
		return r.debtOverviewFn(ctx, userID, start, end)
	}
	return DebtOverview{}, nil
}

func (r dashboardRepoStub) ExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]CategoryBreakdownItem, error) {
	if r.expenseByCategoryFn != nil {
		return r.expenseByCategoryFn(ctx, userID, start, end)
	}
	return []CategoryBreakdownItem{}, nil
}

func (r dashboardRepoStub) UpcomingBills(ctx context.Context, userID int64, start, end time.Time) ([]UpcomingBill, error) {
	if r.upcomingBillsFn != nil {
		return r.upcomingBillsFn(ctx, userID, start, end)
	}
	return []UpcomingBill{}, nil
}

func (r dashboardRepoStub) TopMerchants(ctx context.Context, userID int64, start, end time.Time, limit int) ([]TopMerchant, error) {
	if r.topMerchantsFn != nil {
		return r.topMerchantsFn(ctx, userID, start, end, limit)
	}
	return []TopMerchant{}, nil
}

func TestParseDashboardFilterDefaultsToCurrentMonth(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 19, 10, 30, 0, 0, time.FixedZone("WIB", 7*60*60))
	}
	defer func() { nowFunc = originalNowFunc }()

	req := httptest.NewRequest("GET", "/v1/dashboard/summary", nil)
	filter, err := parseDashboardFilter(req)
	if err != nil {
		t.Fatalf("parseDashboardFilter returned error: %v", err)
	}

	if filter.StartDate == nil || filter.StartDate.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("unexpected start date: %#v", filter.StartDate)
	}

	if filter.EndDate == nil || filter.EndDate.Format("2006-01-02") != "2026-04-30" {
		t.Fatalf("unexpected end date: %#v", filter.EndDate)
	}
}

func TestSummaryDefaultsToCurrentMonth(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 19, 10, 30, 0, 0, time.FixedZone("WIB", 7*60*60))
	}
	defer func() { nowFunc = originalNowFunc }()

	var receivedStart time.Time
	var receivedEnd time.Time

	svc := NewService(dashboardRepoStub{
		refreshUserDebtStatusesFn: func(context.Context, int64) error { return nil },
		allTimeIncomeFn:           func(context.Context, int64) (float64, error) { return 15000000, nil },
		allTimeExpenseFn:          func(context.Context, int64) (float64, error) { return 3000000, nil },
		incomeBetweenFn: func(_ context.Context, _ int64, start, end time.Time) (float64, error) {
			receivedStart = start
			receivedEnd = end
			return 12000000, nil
		},
		expenseBetweenFn: func(context.Context, int64, time.Time, time.Time) (float64, error) {
			return 3000000, nil
		},
		debtOverviewFn: func(_ context.Context, _ int64, start, end time.Time) (DebtOverview, error) {
			if start.Format("2006-01-02") != "2026-04-01" {
				t.Fatalf("unexpected debt overview start: %s", start.Format("2006-01-02"))
			}
			if end.Format("2006-01-02") != "2026-05-01" {
				t.Fatalf("unexpected debt overview end: %s", end.Format("2006-01-02"))
			}
			return DebtOverview{
				TotalDebt:               10000000,
				PaidDebt:                4000000,
				RemainingDebt:           6000000,
				TotalDebtCount:          2,
				ActiveDebtCount:         1,
				OverdueDebtCount:        1,
				PaidInstallments:        3,
				OverdueInstallments:     1,
				UpcomingDueAmount:       1500000,
				UpcomingDueInstallments: 2,
			}, nil
		},
	}, balanceProviderStub{
		totalBalanceFn: func(context.Context, int64) (float64, error) { return 20000000, nil },
	}, nil)

	summary, err := svc.Summary(context.Background(), 1, DashboardFilter{})
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}

	if receivedStart.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("unexpected range start: %s", receivedStart.Format("2006-01-02"))
	}

	if receivedEnd.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("unexpected range end: %s", receivedEnd.Format("2006-01-02"))
	}

	if summary.NetCashflow != 9000000 {
		t.Fatalf("unexpected net cashflow: %v", summary.NetCashflow)
	}

	if summary.PeriodBalance != 9000000 {
		t.Fatalf("unexpected period balance: %v", summary.PeriodBalance)
	}

	if summary.SavingsRate != 75 {
		t.Fatalf("unexpected savings rate: %v", summary.SavingsRate)
	}

	if summary.ExpenseRatio != 25 {
		t.Fatalf("unexpected expense ratio: %v", summary.ExpenseRatio)
	}

	if summary.Debt.DebtToIncomeRatio != 50 {
		t.Fatalf("unexpected debt to income ratio: %v", summary.Debt.DebtToIncomeRatio)
	}

	if summary.Debt.DebtToBalanceRatio != 30 {
		t.Fatalf("unexpected debt to balance ratio: %v", summary.Debt.DebtToBalanceRatio)
	}

	if summary.Debt.CompletionRate != 40 {
		t.Fatalf("unexpected completion rate: %v", summary.Debt.CompletionRate)
	}
}

func TestSummaryRejectsRangeLongerThanThreeMonths(t *testing.T) {
	svc := NewService(dashboardRepoStub{}, nil, nil)

	startDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC)

	_, err := svc.Summary(context.Background(), 1, DashboardFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err == nil {
		t.Fatal("expected Summary to reject a range longer than three months")
	}

	if !strings.Contains(err.Error(), "date range cannot exceed 3 months") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSummaryUsesPeriodBalanceWhileKeepingTotalBalanceRunning(t *testing.T) {
	startDate := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC)

	svc := NewService(dashboardRepoStub{
		refreshUserDebtStatusesFn: func(context.Context, int64) error { return nil },
		allTimeIncomeFn:           func(context.Context, int64) (float64, error) { return 4582000, nil },
		allTimeExpenseFn:          func(context.Context, int64) (float64, error) { return 2000000, nil },
		incomeBetweenFn: func(_ context.Context, _ int64, start, end time.Time) (float64, error) {
			if start.Format("2006-01-02") != "2026-03-01" || end.Format("2006-01-02") != "2026-04-01" {
				t.Fatalf("unexpected period range: %s - %s", start.Format("2006-01-02"), end.Format("2006-01-02"))
			}
			return 0, nil
		},
		expenseBetweenFn: func(context.Context, int64, time.Time, time.Time) (float64, error) { return 0, nil },
		debtOverviewFn: func(context.Context, int64, time.Time, time.Time) (DebtOverview, error) {
			return DebtOverview{RemainingDebt: 0, TotalDebt: 0, PaidDebt: 0}, nil
		},
	}, balanceProviderStub{
		totalBalanceFn: func(context.Context, int64) (float64, error) { return 2582000, nil },
	}, nil)

	summary, err := svc.Summary(context.Background(), 1, DashboardFilter{StartDate: &startDate, EndDate: &endDate})
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}

	if summary.TotalBalance != 2582000 {
		t.Fatalf("unexpected total balance: %v", summary.TotalBalance)
	}

	if summary.PeriodBalance != 0 {
		t.Fatalf("unexpected period balance: %v", summary.PeriodBalance)
	}

	if summary.NetCashflow != 0 {
		t.Fatalf("unexpected net cashflow: %v", summary.NetCashflow)
	}
}

func TestBudgetVsActualUsesExplicitBudgetAmount(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 19, 10, 30, 0, 0, time.FixedZone("WIB", 7*60*60))
	}
	defer func() { nowFunc = originalNowFunc }()

	svc := NewService(dashboardRepoStub{
		refreshUserDebtStatusesFn: func(context.Context, int64) error { return nil },
		allTimeIncomeFn:           func(context.Context, int64) (float64, error) { return 15000000, nil },
		allTimeExpenseFn:          func(context.Context, int64) (float64, error) { return 3000000, nil },
		incomeBetweenFn:           func(context.Context, int64, time.Time, time.Time) (float64, error) { return 12000000, nil },
		expenseBetweenFn:          func(context.Context, int64, time.Time, time.Time) (float64, error) { return 3000000, nil },
		debtOverviewFn: func(context.Context, int64, time.Time, time.Time) (DebtOverview, error) {
			return DebtOverview{}, nil
		},
	}, balanceProviderStub{
		totalBalanceFn: func(context.Context, int64) (float64, error) { return 20000000, nil },
	}, nil)

	budget := 5000000.0
	result, err := svc.BudgetVsActual(context.Background(), 1, DashboardFilter{}, &budget)
	if err != nil {
		t.Fatalf("BudgetVsActual returned error: %v", err)
	}

	if result.BudgetAmount != 5000000 {
		t.Fatalf("unexpected budget amount: %v", result.BudgetAmount)
	}

	if result.UsageRate != 60 {
		t.Fatalf("unexpected usage rate: %v", result.UsageRate)
	}

	if result.RemainingBudget != 2000000 {
		t.Fatalf("unexpected remaining budget: %v", result.RemainingBudget)
	}
}

func TestCategoryBreakdownCalculatesPercentage(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 19, 10, 30, 0, 0, time.FixedZone("WIB", 7*60*60))
	}
	defer func() { nowFunc = originalNowFunc }()

	svc := NewService(dashboardRepoStub{
		refreshUserDebtStatusesFn: func(context.Context, int64) error { return nil },
		allTimeIncomeFn:           func(context.Context, int64) (float64, error) { return 15000000, nil },
		allTimeExpenseFn:          func(context.Context, int64) (float64, error) { return 3000000, nil },
		incomeBetweenFn:           func(context.Context, int64, time.Time, time.Time) (float64, error) { return 12000000, nil },
		expenseBetweenFn:          func(context.Context, int64, time.Time, time.Time) (float64, error) { return 3000000, nil },
		debtOverviewFn: func(context.Context, int64, time.Time, time.Time) (DebtOverview, error) {
			return DebtOverview{}, nil
		},
		expenseByCategoryFn: func(context.Context, int64, time.Time, time.Time) ([]CategoryBreakdownItem, error) {
			return []CategoryBreakdownItem{
				{Category: "Food", Amount: 3000000, TransactionCount: 6},
				{Category: "Transport", Amount: 1000000, TransactionCount: 4},
			}, nil
		},
	}, balanceProviderStub{
		totalBalanceFn: func(context.Context, int64) (float64, error) { return 20000000, nil },
	}, nil)

	items, err := svc.CategoryBreakdown(context.Background(), 1, DashboardFilter{})
	if err != nil {
		t.Fatalf("CategoryBreakdown returned error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].Percentage != 75 {
		t.Fatalf("unexpected top category percentage: %v", items[0].Percentage)
	}
}

func TestUpcomingBillsUsesLookaheadDays(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 19, 10, 30, 0, 0, time.FixedZone("WIB", 7*60*60))
	}
	defer func() { nowFunc = originalNowFunc }()

	var receivedStart time.Time
	var receivedEnd time.Time

	svc := NewService(dashboardRepoStub{
		upcomingBillsFn: func(_ context.Context, _ int64, start, end time.Time) ([]UpcomingBill, error) {
			receivedStart = start
			receivedEnd = end
			return []UpcomingBill{
				{BillName: "Loan installment #1", Amount: 500000, DueDate: "2026-04-25", Status: "pending", SourceType: "debt"},
			}, nil
		},
	}, nil, nil)

	items, err := svc.UpcomingBills(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("UpcomingBills returned error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if receivedStart.Format("2006-01-02") != "2026-04-19" {
		t.Fatalf("unexpected start date: %s", receivedStart.Format("2006-01-02"))
	}

	if receivedEnd.Format("2006-01-02") != "2026-04-26" {
		t.Fatalf("unexpected end date: %s", receivedEnd.Format("2006-01-02"))
	}
}

func TestGoalsProgressReturnsEmptyUntilModuleExists(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 19, 10, 30, 0, 0, time.FixedZone("WIB", 7*60*60))
	}
	defer func() { nowFunc = originalNowFunc }()

	svc := NewService(dashboardRepoStub{
		refreshUserDebtStatusesFn: func(context.Context, int64) error { return nil },
		allTimeIncomeFn:           func(context.Context, int64) (float64, error) { return 15000000, nil },
		allTimeExpenseFn:          func(context.Context, int64) (float64, error) { return 3000000, nil },
		incomeBetweenFn:           func(context.Context, int64, time.Time, time.Time) (float64, error) { return 12000000, nil },
		expenseBetweenFn:          func(context.Context, int64, time.Time, time.Time) (float64, error) { return 3000000, nil },
		debtOverviewFn: func(context.Context, int64, time.Time, time.Time) (DebtOverview, error) {
			return DebtOverview{
				TotalDebt:     10000000,
				PaidDebt:      4000000,
				RemainingDebt: 6000000,
			}, nil
		},
	}, balanceProviderStub{
		totalBalanceFn: func(context.Context, int64) (float64, error) { return 20000000, nil },
	}, nil)

	items, err := svc.GoalsProgress(context.Background(), 1, DashboardFilter{})
	if err != nil {
		t.Fatalf("GoalsProgress returned error: %v", err)
	}

	if len(items) != 0 {
		t.Fatalf("expected no goals until a dedicated goals module exists, got %d", len(items))
	}
}

func TestInsightsReturnsAlerts(t *testing.T) {
	svc := NewService(dashboardRepoStub{}, nil, alertSourceStub{
		listFn: func(context.Context, int64, alerts.AlertListFilter) ([]alerts.Alert, error) {
			return []alerts.Alert{
				{
					Type:        alerts.AlertTypeDailySpendingSpike,
					Title:       "Daily spending spike detected",
					Message:     "Today's spending is above threshold.",
					Severity:    alerts.AlertSeverityWarning,
					MetricValue: 250000,
					DedupeKey:   "daily-spike:2026-04-19",
				},
			}, nil
		},
	})

	items, err := svc.Insights(context.Background(), 1, DashboardFilter{})
	if err != nil {
		t.Fatalf("Insights returned error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(items))
	}

	if items[0].Code != "daily-spike:2026-04-19" {
		t.Fatalf("unexpected insight code: %s", items[0].Code)
	}
}
