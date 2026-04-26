package budget

import "time"

type Status string

const (
	StatusUnderBudget Status = "under_budget"
	StatusOnTrack     Status = "on_track"
	StatusOverBudget  Status = "over_budget"
	StatusInactive    Status = "inactive"
)

type Goal struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	CategoryID    int64     `json:"category_id"`
	CategoryName  string    `json:"category_name"`
	CategoryType  string    `json:"category_type"`
	MonthlyAmount float64   `json:"monthly_amount"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Progress struct {
	Goal
	CurrentAmount      float64 `json:"current_amount"`
	RemainingAmount    float64 `json:"remaining_amount"`
	ProgressPercentage  float64 `json:"progress_percentage"`
	Status             Status  `json:"status"`
}

type Summary struct {
	MonthlyBudget    float64 `json:"monthly_budget"`
	Spent            float64 `json:"spent"`
	Remaining        float64 `json:"remaining"`
	UsageRate        float64 `json:"usage_rate"`
	OverBudgetAmount float64 `json:"over_budget_amount"`
	IsOverBudget     bool    `json:"is_over_budget"`
}
