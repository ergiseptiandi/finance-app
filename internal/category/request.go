package category

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func parseCategoryID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

func decodeCreateInput(r *http.Request) (CreateInput, error) {
	var input CreateInput
	if err := decodeJSON(r, &input); err != nil {
		return CreateInput{}, err
	}

	input.Name = strings.TrimSpace(input.Name)
	return input, nil
}

func decodeUpdateInput(r *http.Request) (UpdateInput, error) {
	var input UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		return UpdateInput{}, err
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}

	return input, nil
}

func parseListFilter(r *http.Request) ListFilter {
	var filter ListFilter
	if value := strings.TrimSpace(r.URL.Query().Get("type")); value != "" {
		categoryType := Type(value)
		filter.Type = &categoryType
	}
	return filter
}
