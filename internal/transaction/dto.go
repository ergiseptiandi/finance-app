package transaction

import "time"

type CreateInput struct {
	WalletID    *int64    `json:"wallet_id,omitempty"`
	Type        Type      `json:"type"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type UpdateInput struct {
	WalletID    *int64     `json:"wallet_id,omitempty"`
	Type        *Type      `json:"type,omitempty"`
	Category    *string    `json:"category,omitempty"`
	Amount      *float64   `json:"amount,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	Description *string    `json:"description,omitempty"`
}

type ListFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	WalletID  *int64
	Category  *string
	Type      *Type
	Page      int
	PerPage   int
}

type PaginatedList struct {
	Data       []Transaction `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}

type Summary struct {
	TotalIncome       float64 `json:"total_income"`
	TotalExpense      float64 `json:"total_expense"`
	Balance           float64 `json:"balance"`
	ConsumptionExpense float64 `json:"consumption_expense"`
	DebtRepayment     float64 `json:"debt_repayment"`
	SavingsRate       float64 `json:"savings_rate"`
	ConsumptionRate   float64 `json:"consumption_rate"`
}
