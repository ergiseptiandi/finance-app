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
	TotalBalance       float64        `json:"total_balance"`
	NetWorth           float64        `json:"net_worth"`
	PeriodBalance      float64        `json:"period_balance"`
	MonthlyIncome      float64        `json:"monthly_income"`
	MonthlyExpense     float64        `json:"monthly_expense"`
	ConsumptionExpense float64        `json:"consumption_expense"`
	DebtRepayment      float64        `json:"debt_repayment"`
	NetCashflow        float64        `json:"net_cashflow"`
	SavingsRate        float64        `json:"savings_rate"`
	ExpenseRatio       float64        `json:"expense_ratio"`
	ConsumptionRate    float64        `json:"consumption_rate"`
	BudgetSummary      *BudgetSummary `json:"budget_summary,omitempty"`
	GoalsProgress      []GoalProgress `json:"goals_progress,omitempty"`
	Debt               DebtOverview   `json:"debt"`
}

type BudgetSummary struct {
	MonthlyBudget    float64 `json:"monthly_budget"`
	Spent            float64 `json:"spent"`
	Remaining        float64 `json:"remaining"`
	UsageRate        float64 `json:"usage_rate"`
	OverBudgetAmount float64 `json:"over_budget_amount"`
	IsOverBudget     bool    `json:"is_over_budget"`
}

type BudgetVsActual struct {
	BudgetAmount     float64 `json:"budget_amount"`
	ActualSpent      float64 `json:"actual_spent"`
	RemainingBudget  float64 `json:"remaining_budget"`
	UsageRate        float64 `json:"usage_rate"`
	OverBudgetAmount float64 `json:"over_budget_amount"`
	Status           string  `json:"status"`
}

type CategoryBreakdownItem struct {
	Category         string  `json:"category"`
	Amount           float64 `json:"amount"`
	Percentage       float64 `json:"percentage"`
	TransactionCount int64   `json:"transaction_count"`
}

type UpcomingBill struct {
	BillName   string  `json:"bill_name"`
	Amount     float64 `json:"amount"`
	DueDate    string  `json:"due_date"`
	Status     string  `json:"status"`
	SourceType string  `json:"source_type"`
}

type TopMerchant struct {
	MerchantName        string  `json:"merchant_name"`
	Amount              float64 `json:"amount"`
	TransactionCount    int64   `json:"transaction_count"`
	LastTransactionDate string  `json:"last_transaction_date"`
}

type Insight struct {
	Type        string  `json:"type"`
	Code        string  `json:"code"`
	Title       string  `json:"title"`
	Message     string  `json:"message"`
	Severity    string  `json:"severity"`
	ChangeValue float64 `json:"change_value,omitempty"`
}

type GoalProgress struct {
	Name               string  `json:"name"`
	TargetAmount       float64 `json:"target_amount"`
	CurrentAmount      float64 `json:"current_amount"`
	ProgressPercentage float64 `json:"progress_percentage"`
	TargetDate         string  `json:"target_date,omitempty"`
	Status             string  `json:"status"`
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
