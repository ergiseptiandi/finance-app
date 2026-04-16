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
	Type        Type      `json:"type"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateInput struct {
	Type        Type      `json:"type"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type UpdateInput struct {
	Type        *Type      `json:"type,omitempty"`
	Category    *string    `json:"category,omitempty"`
	Amount      *float64   `json:"amount,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	Description *string    `json:"description,omitempty"`
}

type ListFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
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
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"`
}
