package dashboard

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func parseDashboardUserID(r *http.Request, middleware Middleware) (int64, bool) {
	claims, ok := middleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	id, err := strconv.ParseInt(strings.TrimSpace(claims.Subject), 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}

func parseDashboardFilter(r *http.Request) (DashboardFilter, error) {
	query := r.URL.Query()
	filter := DashboardFilter{}

	month := strings.TrimSpace(query.Get("month"))
	startDate := strings.TrimSpace(query.Get("start_date"))
	endDate := strings.TrimSpace(query.Get("end_date"))

	if month != "" {
		if startDate != "" || endDate != "" {
			return DashboardFilter{}, errors.New("month cannot be combined with start_date or end_date")
		}

		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return DashboardFilter{}, errors.New("month must use format YYYY-MM")
		}

		startOfMonth := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, parsedMonth.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0).AddDate(0, 0, -1)
		filter.StartDate = &startOfMonth
		filter.EndDate = &endOfMonth
		return filter, nil
	}

	if startDate == "" && endDate == "" {
		now := nowFunc()
		monthStart := startOfMonth(now)
		endOfMonth := monthStart.AddDate(0, 1, 0).AddDate(0, 0, -1)
		filter.StartDate = &monthStart
		filter.EndDate = &endOfMonth
		return filter, nil
	}

	if startDate == "" || endDate == "" {
		return DashboardFilter{}, errors.New("start_date and end_date must be provided together")
	}

	parsedStartDate, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return DashboardFilter{}, errors.New("start_date must use format YYYY-MM-DD")
	}

	parsedEndDate, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return DashboardFilter{}, errors.New("end_date must use format YYYY-MM-DD")
	}

	filter.StartDate = &parsedStartDate
	filter.EndDate = &parsedEndDate
	return filter, nil
}

func parseOptionalFloatQuery(r *http.Request, key string) (*float64, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, errors.New(key + " must be a number")
	}

	return &parsed, nil
}

func parsePositiveIntQuery(r *http.Request, key string, defaultValue int) (int, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New(key + " must be an integer")
	}
	if parsed <= 0 {
		return 0, errors.New(key + " must be greater than zero")
	}

	return parsed, nil
}
