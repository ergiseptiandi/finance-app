package transaction

import "time"

type Type string

const (
	TypeIncome  Type = "income"
	TypeExpense Type = "expense"
)

type Transaction struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	WalletID    int64     `json:"wallet_id"`
	Type        Type      `json:"type"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
