package dashboard

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type dashboardRepoStub struct {
	allTimeIncomeFn  func(context.Context, int64) (float64, error)
	allTimeExpenseFn func(context.Context, int64) (float64, error)
	incomeBetweenFn  func(context.Context, int64, time.Time, time.Time) (float64, error)
	expenseBetweenFn func(context.Context, int64, time.Time, time.Time) (float64, error)
	expenseByDayFn   func(context.Context, int64, time.Time, time.Time) (map[string]float64, error)
	expenseByMonthFn func(context.Context, int64, time.Time, time.Time) (map[string]float64, error)
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
		allTimeIncomeFn:  func(context.Context, int64) (float64, error) { return 15000000, nil },
		allTimeExpenseFn: func(context.Context, int64) (float64, error) { return 3000000, nil },
		incomeBetweenFn: func(_ context.Context, _ int64, start, end time.Time) (float64, error) {
			receivedStart = start
			receivedEnd = end
			return 12000000, nil
		},
		expenseBetweenFn: func(context.Context, int64, time.Time, time.Time) (float64, error) {
			return 3000000, nil
		},
	}, nil)

	_, err := svc.Summary(context.Background(), 1, DashboardFilter{})
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}

	if receivedStart.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("unexpected range start: %s", receivedStart.Format("2006-01-02"))
	}

	if receivedEnd.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("unexpected range end: %s", receivedEnd.Format("2006-01-02"))
	}
}

func TestSummaryRejectsRangeLongerThanThreeMonths(t *testing.T) {
	svc := NewService(dashboardRepoStub{}, nil)

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
