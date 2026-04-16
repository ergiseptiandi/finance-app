package transaction

import (
	"encoding/json"
	"net/http"
	"strconv"
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

func parseListFilter(r *http.Request) ListFilter {
	q := r.URL.Query()
	var filter ListFilter

	if s := q.Get("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			filter.StartDate = &t
		}
	}
	if s := q.Get("end_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			filter.EndDate = &t
		}
	}
	if s := q.Get("category"); s != "" {
		filter.Category = &s
	}
	if s := q.Get("type"); s != "" {
		t := Type(s)
		filter.Type = &t
	}
	if pageStr := q.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = p
		}
	}
	if perPageStr := q.Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil {
			filter.PerPage = pp
		}
	}

	return filter
}
