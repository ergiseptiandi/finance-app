package reports

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseReportsFilterMonth(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/reports/expense-by-category?month=2026-04", nil)

	filter, err := parseReportsFilter(req)
	if err != nil {
		t.Fatalf("parseReportsFilter returned error: %v", err)
	}

	if filter.Mode != ReportFilterModeMonth {
		t.Fatalf("unexpected mode: %s", filter.Mode)
	}

	if filter.StartDate == nil || filter.StartDate.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("unexpected start date: %#v", filter.StartDate)
	}

	if filter.EndDate == nil || filter.EndDate.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("unexpected end date: %#v", filter.EndDate)
	}
}

func TestParseReportsFilterYear(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/reports/spending-trends?year=2026", nil)

	filter, err := parseReportsFilter(req)
	if err != nil {
		t.Fatalf("parseReportsFilter returned error: %v", err)
	}

	if filter.Mode != ReportFilterModeYear {
		t.Fatalf("unexpected mode: %s", filter.Mode)
	}

	if filter.StartDate == nil || filter.StartDate.Format("2006-01-02") != "2026-01-01" {
		t.Fatalf("unexpected start date: %#v", filter.StartDate)
	}

	if filter.EndDate == nil || filter.EndDate.Format("2006-01-02") != "2027-01-01" {
		t.Fatalf("unexpected end date: %#v", filter.EndDate)
	}
}

func TestParseReportsFilterRejectsMixedModes(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/reports/expense-by-category?month=2026-04&year=2026", nil)

	_, err := parseReportsFilter(req)
	if err == nil {
		t.Fatal("expected parseReportsFilter to reject mixed month and year filters")
	}

	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseReportsFilterRejectsRangeLongerThanOneYear(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/reports/expense-by-category?start_date=2026-01-01&end_date=2027-01-02", nil)

	_, err := parseReportsFilter(req)
	if err == nil {
		t.Fatal("expected parseReportsFilter to reject a range longer than one year")
	}

	if !strings.Contains(err.Error(), "date range cannot exceed 1 year") {
		t.Fatalf("unexpected error: %v", err)
	}
}
