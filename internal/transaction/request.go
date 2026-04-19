package transaction

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func parseTransactionUserID(r *http.Request, middleware Middleware) (int64, bool) {
	claims, ok := middleware.GetAccessClaims(r)
	if !ok {
		return 0, false
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}

func parseTransactionID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

func decodeCreateInput(r *http.Request) (CreateInput, error) {
	var input CreateInput
	if err := decodeJSON(r, &input); err != nil {
		return CreateInput{}, err
	}
	return input, nil
}

func decodeUpdateInput(r *http.Request) (UpdateInput, error) {
	var input UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		return UpdateInput{}, err
	}
	return input, nil
}

func parseListFilter(r *http.Request) (ListFilter, error) {
	q := r.URL.Query()
	var filter ListFilter

	month := strings.TrimSpace(q.Get("month"))
	startDate := strings.TrimSpace(q.Get("start_date"))
	endDate := strings.TrimSpace(q.Get("end_date"))

	if month != "" {
		if startDate != "" || endDate != "" {
			return ListFilter{}, errors.New("month cannot be combined with start_date or end_date")
		}

		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return ListFilter{}, errors.New("month must use format YYYY-MM")
		}

		startOfMonth := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0).AddDate(0, 0, -1)

		filter.StartDate = &startOfMonth
		filter.EndDate = &endOfMonth
	}

	if month == "" && (startDate != "" || endDate != "") {
		if startDate == "" || endDate == "" {
			return ListFilter{}, errors.New("start_date and end_date must be provided together")
		}

		parsedStartDate, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return ListFilter{}, errors.New("start_date must use format YYYY-MM-DD")
		}

		parsedEndDate, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return ListFilter{}, errors.New("end_date must use format YYYY-MM-DD")
		}

		filter.StartDate = &parsedStartDate
		filter.EndDate = &parsedEndDate
	}
	if walletIDStr := strings.TrimSpace(q.Get("wallet_id")); walletIDStr != "" {
		walletID, err := strconv.ParseInt(walletIDStr, 10, 64)
		if err != nil || walletID <= 0 {
			return ListFilter{}, errors.New("wallet_id must be a positive number")
		}
		filter.WalletID = &walletID
	}
	if s := q.Get("category"); s != "" {
		filter.Category = &s
	}
	if s := q.Get("type"); s != "" {
		t := Type(s)
		if t != TypeIncome && t != TypeExpense {
			return ListFilter{}, errors.New("type must be either income or expense")
		}
		filter.Type = &t
	}
	if pageStr := q.Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			return ListFilter{}, errors.New("page must be a number")
		}
		filter.Page = p
	}
	if perPageStr := q.Get("per_page"); perPageStr != "" {
		pp, err := strconv.Atoi(perPageStr)
		if err != nil {
			return ListFilter{}, errors.New("per_page must be a number")
		}
		filter.PerPage = pp
	}

	return filter, nil
}
