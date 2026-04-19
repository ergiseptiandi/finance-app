package dashboard

import "time"

type DashboardFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
}

type DebtOverview struct {
	TotalDebt               float64 `json:"total_debt"`
	PaidDebt                float64 `json:"paid_debt"`
	RemainingDebt           float64 `json:"remaining_debt"`
	TotalDebtCount          int64   `json:"total_debt_count"`
	ActiveDebtCount         int64   `json:"active_debt_count"`
	OverdueDebtCount        int64   `json:"overdue_debt_count"`
	PaidInstallments        int64   `json:"paid_installments"`
	OverdueInstallments     int64   `json:"overdue_installments"`
	UpcomingDueAmount       float64 `json:"upcoming_due_amount"`
	UpcomingDueInstallments int64   `json:"upcoming_due_installments"`
	DebtToIncomeRatio       float64 `json:"debt_to_income_ratio"`
	DebtToBalanceRatio      float64 `json:"debt_to_balance_ratio"`
	CompletionRate          float64 `json:"completion_rate"`
}

type Summary struct {
	TotalBalance   float64      `json:"total_balance"`
	MonthlyIncome  float64      `json:"monthly_income"`
	MonthlyExpense float64      `json:"monthly_expense"`
	NetCashflow    float64      `json:"net_cashflow"`
	SavingsRate    float64      `json:"savings_rate"`
	ExpenseRatio   float64      `json:"expense_ratio"`
	Debt           DebtOverview `json:"debt"`
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
