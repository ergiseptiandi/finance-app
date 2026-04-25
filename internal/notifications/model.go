package notifications

import "time"

type ReminderKind string

const (
	ReminderKindDailyExpense   ReminderKind = "daily_expense_input"
	ReminderKindDebtPayment    ReminderKind = "debt_payment"
	ReminderKindSalary         ReminderKind = "salary_reminder"
	ReminderKindBudgetWarning  ReminderKind = "budget_warning"
	ReminderKindWeeklySummary  ReminderKind = "weekly_summary"
	ReminderKindLargeTransaction ReminderKind = "large_transaction"
	ReminderKindGoalReminder   ReminderKind = "goal_reminder"
)

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusSkipped   DeliveryStatus = "skipped"
)

type Settings struct {
	UserID                        int64     `json:"user_id"`
	Enabled                       bool      `json:"enabled"`
	DailyExpenseReminderEnabled   bool      `json:"daily_expense_reminder_enabled"`
	DailyExpenseReminderTime      string    `json:"daily_expense_reminder_time"`
	DebtPaymentReminderEnabled    bool      `json:"debt_payment_reminder_enabled"`
	DebtPaymentReminderTime       string    `json:"debt_payment_reminder_time"`
	DebtPaymentReminderDaysBefore int       `json:"debt_payment_reminder_days_before"`
	SalaryReminderEnabled         bool      `json:"salary_reminder_enabled"`
	SalaryReminderTime            string    `json:"salary_reminder_time"`
	SalaryReminderDaysBefore      int       `json:"salary_reminder_days_before"`
	SalaryDay                     int       `json:"salary_day"`
	BudgetAmount                  float64   `json:"budget_amount"`
	BudgetWarningEnabled          bool      `json:"budget_warning_enabled"`
	BudgetWarningThreshold        int       `json:"budget_warning_threshold"` // percentage (e.g., 80 = 80%)
	WeeklySummaryEnabled          bool      `json:"weekly_summary_enabled"`
	WeeklySummaryDay              int       `json:"weekly_summary_day"` // 0=Sunday, 1=Monday, etc.
	LargeTransactionEnabled       bool      `json:"large_transaction_enabled"`
	LargeTransactionThreshold     float64   `json:"large_transaction_threshold"` // amount threshold
	GoalReminderEnabled           bool      `json:"goal_reminder_enabled"`
	GoalReminderDaysBefore        int       `json:"goal_reminder_days_before"` // days before goal deadline
	PushToken                     string    `json:"push_token"`
	CreatedAt                     time.Time `json:"created_at"`
	UpdatedAt                     time.Time `json:"updated_at"`
}

type Notification struct {
	ID             int64             `json:"id"`
	UserID         int64             `json:"-"`
	Kind           ReminderKind      `json:"kind"`
	Type           string            `json:"type,omitempty"`
	Title          string            `json:"title"`
	Message        string            `json:"message"`
	Read           bool              `json:"read"`
	DeliveryStatus DeliveryStatus    `json:"delivery_status"`
	ScheduledFor   time.Time         `json:"scheduled_for"`
	SentAt         *time.Time        `json:"sent_at,omitempty"`
	ReadAt         *time.Time        `json:"read_at,omitempty"`
	DedupeKey      string            `json:"dedupe_key"`
	Data           map[string]string `json:"data,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type UpdateSettingsInput struct {
	Enabled                       *bool   `json:"enabled,omitempty"`
	DailyExpenseReminderEnabled   *bool   `json:"daily_expense_reminder_enabled,omitempty"`
	DailyExpenseReminderTime      *string `json:"daily_expense_reminder_time,omitempty"`
	DebtPaymentReminderEnabled    *bool   `json:"debt_payment_reminder_enabled,omitempty"`
	DebtPaymentReminderTime       *string `json:"debt_payment_reminder_time,omitempty"`
	DebtPaymentReminderDaysBefore *int    `json:"debt_payment_reminder_days_before,omitempty"`
	SalaryReminderEnabled         *bool   `json:"salary_reminder_enabled,omitempty"`
	SalaryReminderTime            *string `json:"salary_reminder_time,omitempty"`
	SalaryReminderDaysBefore      *int    `json:"salary_reminder_days_before,omitempty"`
	SalaryDay                     *int    `json:"salary_day,omitempty"`
	BudgetAmount                  *float64 `json:"budget_amount,omitempty"`
	BudgetWarningEnabled          *bool   `json:"budget_warning_enabled,omitempty"`
	BudgetWarningThreshold        *int    `json:"budget_warning_threshold,omitempty"`
	WeeklySummaryEnabled          *bool   `json:"weekly_summary_enabled,omitempty"`
	WeeklySummaryDay              *int    `json:"weekly_summary_day,omitempty"`
	LargeTransactionEnabled       *bool   `json:"large_transaction_enabled,omitempty"`
	LargeTransactionThreshold     *float64 `json:"large_transaction_threshold,omitempty"`
	GoalReminderEnabled           *bool   `json:"goal_reminder_enabled,omitempty"`
	GoalReminderDaysBefore        *int    `json:"goal_reminder_days_before,omitempty"`
	PushToken                     *string `json:"push_token,omitempty"`
}

type NotificationFilter struct {
	Kind *string
	Read *bool
}

type ReminderSummary struct {
	Count     int64      `json:"count"`
	Amount    float64    `json:"amount"`
	NextDueAt *time.Time `json:"next_due_at,omitempty"`
}
