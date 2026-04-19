package transaction

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type listRepositoryStub struct {
	findAllFn func(ctx context.Context, userID int64, filter ListFilter) (PaginatedList, error)
}

func (r listRepositoryStub) Create(ctx context.Context, txn Transaction) (int64, error) {
	return 0, nil
}

func (r listRepositoryStub) GetByID(ctx context.Context, id int64, userID int64) (Transaction, error) {
	return Transaction{}, nil
}

func (r listRepositoryStub) Update(ctx context.Context, txn Transaction) error {
	return nil
}

func (r listRepositoryStub) Delete(ctx context.Context, id int64, userID int64) error {
	return nil
}

func (r listRepositoryStub) FindAll(ctx context.Context, userID int64, filter ListFilter) (PaginatedList, error) {
	if r.findAllFn != nil {
		return r.findAllFn(ctx, userID, filter)
	}
	return PaginatedList{}, nil
}

func (r listRepositoryStub) GetSummary(ctx context.Context, userID int64) (Summary, error) {
	return Summary{}, nil
}

func TestParseListFilterMonth(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/transactions?month=2026-04&type=income&page=2&per_page=5", nil)

	filter, err := parseListFilter(req)
	if err != nil {
		t.Fatalf("parseListFilter returned error: %v", err)
	}

	if filter.StartDate == nil || filter.StartDate.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("unexpected start date: %#v", filter.StartDate)
	}

	if filter.EndDate == nil || filter.EndDate.Format("2006-01-02") != "2026-04-30" {
		t.Fatalf("unexpected end date: %#v", filter.EndDate)
	}

	if filter.Type == nil || *filter.Type != TypeIncome {
		t.Fatalf("unexpected type: %#v", filter.Type)
	}

	if filter.Page != 2 {
		t.Fatalf("unexpected page: %d", filter.Page)
	}

	if filter.PerPage != 5 {
		t.Fatalf("unexpected per_page: %d", filter.PerPage)
	}
}

func TestParseListFilterRejectsMixedMonthAndCustomRange(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/transactions?month=2026-04&start_date=2026-04-01&end_date=2026-04-30", nil)

	_, err := parseListFilter(req)
	if err == nil {
		t.Fatal("expected parseListFilter to reject mixed month and custom range")
	}

	if !strings.Contains(err.Error(), "month cannot be combined") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceListRejectsRangeOverTwoMonths(t *testing.T) {
	svc := NewService(listRepositoryStub{})

	startDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC)

	_, err := svc.List(context.Background(), 1, ListFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err == nil {
		t.Fatal("expected List to reject a range longer than two months")
	}

	if !strings.Contains(err.Error(), "date range cannot exceed 2 months") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceListRejectsIncompleteCustomRange(t *testing.T) {
	svc := NewService(listRepositoryStub{})

	startDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	_, err := svc.List(context.Background(), 1, ListFilter{
		StartDate: &startDate,
	})
	if err == nil {
		t.Fatal("expected List to reject incomplete custom range")
	}

	if !strings.Contains(err.Error(), "start_date and end_date must be provided together") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceListAppliesDefaultPagination(t *testing.T) {
	var received ListFilter

	svc := NewService(listRepositoryStub{
		findAllFn: func(ctx context.Context, userID int64, filter ListFilter) (PaginatedList, error) {
			received = filter
			return PaginatedList{}, nil
		},
	})

	startDate := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)

	_, err := svc.List(context.Background(), 1, ListFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if received.Page != 1 {
		t.Fatalf("unexpected default page: %d", received.Page)
	}

	if received.PerPage != 10 {
		t.Fatalf("unexpected default per_page: %d", received.PerPage)
	}
}

func TestServiceListDefaultsDateRangeToCurrentMonth(t *testing.T) {
	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 19, 10, 30, 0, 0, time.FixedZone("WIB", 7*60*60))
	}
	defer func() {
		nowFunc = originalNowFunc
	}()

	var received ListFilter

	svc := NewService(listRepositoryStub{
		findAllFn: func(ctx context.Context, userID int64, filter ListFilter) (PaginatedList, error) {
			received = filter
			return PaginatedList{}, nil
		},
	})

	_, err := svc.List(context.Background(), 1, ListFilter{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if received.StartDate == nil || received.StartDate.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("unexpected default start date: %#v", received.StartDate)
	}

	if received.EndDate == nil || received.EndDate.Format("2006-01-02") != "2026-04-30" {
		t.Fatalf("unexpected default end date: %#v", received.EndDate)
	}
}
