package salary

import "time"

type CreateInput struct {
	Amount float64   `json:"amount"`
	PaidAt time.Time `json:"paid_at"`
	Note   string    `json:"note"`
}

type UpdateInput struct {
	Amount *float64   `json:"amount,omitempty"`
	PaidAt *time.Time `json:"paid_at,omitempty"`
	Note   *string    `json:"note,omitempty"`
}

type SetSalaryDayInput struct {
	SalaryDay int `json:"salary_day"`
}
