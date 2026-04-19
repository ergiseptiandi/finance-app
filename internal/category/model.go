package category

import "time"

type Type string

const (
	TypeIncome  Type = "income"
	TypeExpense Type = "expense"
)

type Category struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"-"`
	Name      string    `json:"name"`
	Type      Type      `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
