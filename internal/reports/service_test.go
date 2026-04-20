package reports

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type reportsRepoStub struct {
	expenseByCategoryFn func(context.Context, int64, time.Time, time.Time) ([]CategoryExpense, error)
	trendByPeriodFn     func(context.Context, int64, time.Time, time.Time, TrendGroupBy) (map[string]TrendTotals, error)
	allTimeIncomeFn     func(context.Context, int64) (float64, error)
	allTimeExpenseFn    func(context.Context, int64) (float64, error)
	incomeBetweenFn     func(context.Context, int64, time.Time, time.Time) (float64, error)
	expenseBetweenFn    func(context.Context, int64, time.Time, time.Time) (float64, error)
}

func (r reportsRepoStub) ExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]CategoryExpense, error) {
	if r.expenseByCategoryFn != nil {
		return r.expenseByCategoryFn(ctx, userID, start, end)
	}
	return []CategoryExpense{}, nil
}

func (r reportsRepoStub) TrendByPeriod(ctx context.Context, userID int64, start, end time.Time, groupBy TrendGroupBy) (map[string]TrendTotals, error) {
	if r.trendByPeriodFn != nil {
		return r.trendByPeriodFn(ctx, userID, start, end, groupBy)
	}
	return map[string]TrendTotals{}, nil
}

func (r reportsRepoStub) AllTimeIncome(ctx context.Context, userID int64) (float64, error) {
	if r.allTimeIncomeFn != nil {
		return r.allTimeIncomeFn(ctx, userID)
	}
	return 0, nil
}

func (r reportsRepoStub) AllTimeExpense(ctx context.Context, userID int64) (float64, error) {
	if r.allTimeExpenseFn != nil {
		return r.allTimeExpenseFn(ctx, userID)
	}
	return 0, nil
}

func (r reportsRepoStub) IncomeBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	if r.incomeBetweenFn != nil {
		return r.incomeBetweenFn(ctx, userID, start, end)
	}
	return 0, nil
}

func (r reportsRepoStub) ExpenseBetween(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	if r.expenseBetweenFn != nil {
		return r.expenseBetweenFn(ctx, userID, start, end)
	}
	return 0, nil
}

type reportsBalanceStub struct {
	totalBalanceFn func(context.Context, int64) (float64, error)
}

func (b reportsBalanceStub) TotalBalance(ctx context.Context, userID int64) (float64, error) {
	if b.totalBalanceFn != nil {
		return b.totalBalanceFn(ctx, userID)
	}
	return 0, nil
}

func TestExpenseByCategoryBuildsSummary(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 19, 10, 30, 0, 0, time.UTC)
	}
	defer func() { nowFunc = originalNowFunc }()

	var receivedStart time.Time
	var receivedEnd time.Time

	svc := NewService(reportsRepoStub{
		expenseByCategoryFn: func(_ context.Context, _ int64, start, end time.Time) ([]CategoryExpense, error) {
			receivedStart = start
			receivedEnd = end
			return []CategoryExpense{
				{Category: "Transport", Amount: 1000000, TransactionCount: 4},
				{Category: "Food", Amount: 3000000, TransactionCount: 6},
			}, nil
		},
	}, nil)

	report, err := svc.ExpenseByCategory(context.Background(), 1, ReportsFilter{})
	if err != nil {
		t.Fatalf("ExpenseByCategory returned error: %v", err)
	}

	if receivedStart.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("unexpected start date: %s", receivedStart.Format("2006-01-02"))
	}

	if receivedEnd.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("unexpected end date: %s", receivedEnd.Format("2006-01-02"))
	}

	if report.Summary.TotalExpense != 4000000 {
		t.Fatalf("unexpected total expense: %v", report.Summary.TotalExpense)
	}

	if report.Summary.TopCategory != "Food" {
		t.Fatalf("unexpected top category: %s", report.Summary.TopCategory)
	}

	if len(report.Items) != 2 {
		t.Fatalf("unexpected items length: %d", len(report.Items))
	}

	if report.Items[0].Category != "Food" || report.Items[0].Percentage != 75 {
		t.Fatalf("unexpected first item: %+v", report.Items[0])
	}
}

func TestSpendingTrendsUsesYearGrouping(t *testing.T) {
	var receivedGroupBy TrendGroupBy
	var receivedStart time.Time
	var receivedEnd time.Time

	svc := NewService(reportsRepoStub{
		trendByPeriodFn: func(_ context.Context, _ int64, start, end time.Time, groupBy TrendGroupBy) (map[string]TrendTotals, error) {
			receivedStart = start
			receivedEnd = end
			receivedGroupBy = groupBy
			return map[string]TrendTotals{
				"2026-01": {Income: 5000000, Expense: 1000000},
				"2026-03": {Income: 4000000, Expense: 1500000},
			}, nil
		},
	}, nil)

	filter, err := parseReportsFilter(httptestRequest("GET", "/v1/reports/spending-trends?year=2026"))
	if err != nil {
		t.Fatalf("parseReportsFilter returned error: %v", err)
	}

	report, err := svc.SpendingTrends(context.Background(), 1, filter)
	if err != nil {
		t.Fatalf("SpendingTrends returned error: %v", err)
	}

	if receivedGroupBy != TrendGroupByMonth {
		t.Fatalf("unexpected group by: %s", receivedGroupBy)
	}

	if receivedStart.Format("2006-01-02") != "2026-01-01" {
		t.Fatalf("unexpected start date: %s", receivedStart.Format("2006-01-02"))
	}

	if receivedEnd.Format("2006-01-02") != "2027-01-01" {
		t.Fatalf("unexpected end date: %s", receivedEnd.Format("2006-01-02"))
	}

	if report.GroupBy != string(TrendGroupByMonth) {
		t.Fatalf("unexpected response group by: %s", report.GroupBy)
	}

	if len(report.Items) != 12 {
		t.Fatalf("expected 12 monthly points, got %d", len(report.Items))
	}

	if report.Items[0].Period != "2026-01" || report.Items[0].NetCashflow != 4000000 {
		t.Fatalf("unexpected first point: %+v", report.Items[0])
	}
}

func TestAverageDailySpendingCalculatesDailyExtremes(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 19, 10, 30, 0, 0, time.UTC)
	}
	defer func() { nowFunc = originalNowFunc }()

	svc := NewService(reportsRepoStub{
		trendByPeriodFn: func(_ context.Context, _ int64, start, end time.Time, groupBy TrendGroupBy) (map[string]TrendTotals, error) {
			if groupBy != TrendGroupByDay {
				t.Fatalf("expected daily grouping, got %s", groupBy)
			}
			return map[string]TrendTotals{
				"2026-04-01": {Expense: 50000},
				"2026-04-02": {Expense: 150000},
				"2026-04-03": {Expense: 0},
			}, nil
		},
	}, nil)

	filter, err := parseReportsFilter(httptestRequest("GET", "/v1/reports/average-daily-spending?month=2026-04"))
	if err != nil {
		t.Fatalf("parseReportsFilter returned error: %v", err)
	}

	report, err := svc.AverageDailySpending(context.Background(), 1, filter)
	if err != nil {
		t.Fatalf("AverageDailySpending returned error: %v", err)
	}

	if report.DaysCount != 30 {
		t.Fatalf("unexpected days count: %d", report.DaysCount)
	}

	if report.TotalExpense != 200000 {
		t.Fatalf("unexpected total expense: %v", report.TotalExpense)
	}

	if report.AverageDailySpending != 6666.67 {
		t.Fatalf("unexpected average daily spending: %v", report.AverageDailySpending)
	}

	if report.HighestDailySpending != 150000 || report.LowestDailySpending != 0 {
		t.Fatalf("unexpected daily extremes: %+v", report)
	}
}

func TestRemainingBalanceUsesAllTimeBalanceProviderWhenUnfiltered(t *testing.T) {
	svc := NewService(reportsRepoStub{
		allTimeIncomeFn:  func(context.Context, int64) (float64, error) { return 12000000, nil },
		allTimeExpenseFn: func(context.Context, int64) (float64, error) { return 3000000, nil },
	}, reportsBalanceStub{
		totalBalanceFn: func(context.Context, int64) (float64, error) { return 20000000, nil },
	})

	report, err := svc.RemainingBalance(context.Background(), 1, ReportsFilter{})
	if err != nil {
		t.Fatalf("RemainingBalance returned error: %v", err)
	}

	if report.Period.Mode != "all-time" {
		t.Fatalf("unexpected period mode: %s", report.Period.Mode)
	}

	if report.RemainingBalance != 20000000 {
		t.Fatalf("unexpected remaining balance: %v", report.RemainingBalance)
	}

	if report.TotalIncome != 12000000 || report.TotalExpense != 3000000 {
		t.Fatalf("unexpected income/expense: %+v", report)
	}
}

func httptestRequest(method, target string) *http.Request {
	return httptest.NewRequest(method, target, nil)
}
