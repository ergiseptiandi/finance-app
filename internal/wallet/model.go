package wallet

import "time"

type Wallet struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"-"`
	Name           string    `json:"name"`
	OpeningBalance float64   `json:"opening_balance"`
	IsLocked       bool      `json:"is_locked"`
	Balance        float64   `json:"balance"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Transfer struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"-"`
	FromWalletID int64     `json:"from_wallet_id"`
	ToWalletID   int64     `json:"to_wallet_id"`
	Amount       float64   `json:"amount"`
	Note         string    `json:"note"`
	TransferDate time.Time `json:"transfer_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Summary struct {
	TotalBalance float64  `json:"total_balance"`
	Wallets      []Wallet `json:"wallets"`
}
