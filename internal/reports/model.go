package reports

type ReportPeriod struct {
	Mode      string `json:"mode"`
	Month     string `json:"month,omitempty"`
	Year      int    `json:"year,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

type CategoryExpense struct {
	Category         string  `json:"category"`
	Amount           float64 `json:"amount"`
	Percentage       float64 `json:"percentage"`
	TransactionCount int     `json:"transaction_count"`
}

type ExpenseByCategorySummary struct {
	TotalExpense  float64 `json:"total_expense"`
	CategoryCount int     `json:"category_count"`
	TopCategory   string  `json:"top_category"`
}

type ExpenseByCategoryReport struct {
	Period  ReportPeriod             `json:"period"`
	Summary ExpenseByCategorySummary `json:"summary"`
	Items   []CategoryExpense        `json:"items"`
}

type TrendTotals struct {
	Income  float64
	Expense float64
}

type TrendPoint struct {
	Period      string  `json:"period"`
	Income      float64 `json:"income"`
	Expense     float64 `json:"expense"`
	NetCashflow float64 `json:"net_cashflow"`
}

type SpendingTrendsReport struct {
	Period  ReportPeriod `json:"period"`
	GroupBy string       `json:"group_by"`
	Items   []TrendPoint `json:"items"`
}

type HighestSpendingCategoryReport struct {
	Period           ReportPeriod `json:"period"`
	Category         string       `json:"category"`
	Amount           float64      `json:"amount"`
	Percentage       float64      `json:"percentage"`
	TransactionCount int          `json:"transaction_count"`
}

type AverageDailySpendingReport struct {
	Period               ReportPeriod `json:"period"`
	TotalExpense         float64      `json:"total_expense"`
	DaysCount            int          `json:"days_count"`
	AverageDailySpending float64      `json:"average_daily_spending"`
	HighestDailySpending float64      `json:"highest_daily_spending"`
	LowestDailySpending  float64      `json:"lowest_daily_spending"`
}

type RemainingBalanceReport struct {
	Period           ReportPeriod `json:"period"`
	TotalIncome      float64      `json:"total_income"`
	TotalExpense     float64      `json:"total_expense"`
	RemainingBalance float64      `json:"remaining_balance"`
	SavingsRate      float64      `json:"savings_rate"`
	ExpenseRatio     float64      `json:"expense_ratio"`
}
