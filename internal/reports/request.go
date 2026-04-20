package reports

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ReportFilterMode string

const (
	ReportFilterModeMonth  ReportFilterMode = "month"
	ReportFilterModeYear   ReportFilterMode = "year"
	ReportFilterModeCustom ReportFilterMode = "custom"
)

type ReportsFilter struct {
	Mode       ReportFilterMode
	StartDate  *time.Time
	EndDate    *time.Time
	MonthValue string
	YearValue  int
}

func (f ReportsFilter) IsZero() bool {
	return f.Mode == "" && f.StartDate == nil && f.EndDate == nil
}

func parseReportsUserID(r *http.Request, middleware Middleware) (int64, bool) {
	claims, ok := middleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	return parseReportsUserIDFromClaims(claims.Subject)
}

func parseReportsUserIDFromClaims(subject string) (int64, bool) {
	return parseInt64(subject)
}

func parseInt64(value string) (int64, bool) {
	var id int64
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		id = id*10 + int64(ch-'0')
	}
	if id == 0 {
		return 0, false
	}
	return id, true
}

func parseReportsFilter(r *http.Request) (ReportsFilter, error) {
	query := r.URL.Query()
	filter := ReportsFilter{}

	month := strings.TrimSpace(query.Get("month"))
	year := strings.TrimSpace(query.Get("year"))
	startDate := strings.TrimSpace(query.Get("start_date"))
	endDate := strings.TrimSpace(query.Get("end_date"))

	if month != "" {
		if year != "" || startDate != "" || endDate != "" {
			return ReportsFilter{}, errors.New("month cannot be combined with year or start_date/end_date")
		}

		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return ReportsFilter{}, errors.New("month must use format YYYY-MM")
		}

		startOfMonth := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		filter.Mode = ReportFilterModeMonth
		filter.StartDate = &startOfMonth
		filter.EndDate = &endOfMonth
		filter.MonthValue = startOfMonth.Format("2006-01")
		return filter, nil
	}

	if year != "" {
		if startDate != "" || endDate != "" {
			return ReportsFilter{}, errors.New("year cannot be combined with start_date/end_date")
		}

		parsedYear, err := strconv.Atoi(year)
		if err != nil || parsedYear <= 0 {
			return ReportsFilter{}, errors.New("year must use format YYYY")
		}

		startOfYear := time.Date(parsedYear, time.January, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := startOfYear.AddDate(1, 0, 0)
		filter.Mode = ReportFilterModeYear
		filter.StartDate = &startOfYear
		filter.EndDate = &endOfYear
		filter.YearValue = parsedYear
		return filter, nil
	}

	if startDate == "" && endDate == "" {
		return filter, nil
	}

	if startDate == "" || endDate == "" {
		return ReportsFilter{}, errors.New("start_date and end_date must be provided together")
	}

	parsedStartDate, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return ReportsFilter{}, errors.New("start_date must use format YYYY-MM-DD")
	}
	parsedStartDate = time.Date(parsedStartDate.Year(), parsedStartDate.Month(), parsedStartDate.Day(), 0, 0, 0, 0, time.UTC)

	parsedEndDate, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return ReportsFilter{}, errors.New("end_date must use format YYYY-MM-DD")
	}
	parsedEndDate = time.Date(parsedEndDate.Year(), parsedEndDate.Month(), parsedEndDate.Day(), 0, 0, 0, 0, time.UTC)

	if parsedEndDate.Before(parsedStartDate) {
		return ReportsFilter{}, errors.New("end_date must be greater than or equal to start_date")
	}

	if parsedEndDate.After(parsedStartDate.AddDate(1, 0, 0)) {
		return ReportsFilter{}, errors.New("date range cannot exceed 1 year")
	}

	filter.Mode = ReportFilterModeCustom
	filter.StartDate = &parsedStartDate
	filter.EndDate = &parsedEndDate
	return filter, nil
}
