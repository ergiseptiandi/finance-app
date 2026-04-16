package dashboard

type Summary struct {
	TotalBalance   float64 `json:"total_balance"`
	MonthlyIncome  float64 `json:"monthly_income"`
	MonthlyExpense float64 `json:"monthly_expense"`
}

type SpendingPoint struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type MonthlySpendingPoint struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

type ComparisonMetric struct {
	Current          float64 `json:"current"`
	Previous         float64 `json:"previous"`
	Difference       float64 `json:"difference"`
	PercentageChange float64 `json:"percentage_change"`
}

type Comparison struct {
	TodayVsYesterday     ComparisonMetric `json:"today_vs_yesterday"`
	ThisMonthVsLastMonth ComparisonMetric `json:"this_month_vs_last_month"`
}

type ExpenseVsSalary struct {
	MonthlyExpense float64 `json:"monthly_expense"`
	CurrentSalary  float64 `json:"current_salary"`
	Percentage     float64 `json:"percentage"`
}
