package budget

type CreateInput struct {
	CategoryID    int64   `json:"category_id"`
	MonthlyAmount float64 `json:"monthly_amount"`
}

type UpdateInput struct {
	CategoryID    *int64   `json:"category_id,omitempty"`
	MonthlyAmount *float64 `json:"monthly_amount,omitempty"`
}
