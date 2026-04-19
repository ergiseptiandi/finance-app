package wallet

import (
	"encoding/json"
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

func parseWalletID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func parseTransferID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "transferId"), 10, 64)
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

func decodeCreateTransferInput(r *http.Request) (CreateTransferInput, error) {
	var input CreateTransferInput
	if err := decodeJSON(r, &input); err != nil {
		return CreateTransferInput{}, err
	}
	input.Note = strings.TrimSpace(input.Note)
	if input.TransferDate.IsZero() {
		input.TransferDate = time.Now()
	}
	return input, nil
}
