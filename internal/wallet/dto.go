package wallet

import "time"

type CreateInput struct {
	Name           string  `json:"name"`
	OpeningBalance float64 `json:"opening_balance"`
}

type UpdateInput struct {
	Name           *string  `json:"name,omitempty"`
	OpeningBalance *float64 `json:"opening_balance,omitempty"`
}

type CreateTransferInput struct {
	FromWalletID int64     `json:"from_wallet_id"`
	ToWalletID   int64     `json:"to_wallet_id"`
	Amount       float64   `json:"amount"`
	Note         string    `json:"note"`
	TransferDate time.Time `json:"transfer_date"`
}
