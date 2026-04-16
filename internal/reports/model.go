package reports

type CategoryExpense struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type TrendPoint struct {
	Period string  `json:"period"`
	Amount float64 `json:"amount"`
}

type AverageDailySpending struct {
	TotalExpense         float64 `json:"total_expense"`
	DaysCount            int     `json:"days_count"`
	AverageDailySpending float64 `json:"average_daily_spending"`
}

type RemainingBalance struct {
	TotalIncome      float64 `json:"total_income"`
	TotalExpense     float64 `json:"total_expense"`
	RemainingBalance float64 `json:"remaining_balance"`
}
