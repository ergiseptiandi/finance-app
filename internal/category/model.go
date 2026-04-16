package category

import "time"

type Type string

const (
	TypeIncome  Type = "income"
	TypeExpense Type = "expense"
)

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      Type      `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name string `json:"name"`
	Type Type   `json:"type"`
}

type UpdateInput struct {
	Name *string `json:"name,omitempty"`
	Type *Type   `json:"type,omitempty"`
}

type ListFilter struct {
	Type *Type
}
