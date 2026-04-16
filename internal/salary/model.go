package salary

import "time"

type SalaryRecord struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Amount    float64   `json:"amount"`
	PaidAt    time.Time `json:"paid_at"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CurrentSalary struct {
	SalaryRecord
	SalaryDay *int `json:"salary_day,omitempty"`
}

type SalarySchedule struct {
	UserID    int64     `json:"-"`
	SalaryDay int       `json:"salary_day"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
