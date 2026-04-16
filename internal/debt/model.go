package debt

import "time"

type Status string

const (
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
	StatusOverdue Status = "overdue"
)

type Debt struct {
	ID                 int64     `json:"id"`
	UserID             int64     `json:"user_id"`
	Name               string    `json:"name"`
	TotalAmount        float64   `json:"total_amount"`
	MonthlyInstallment float64   `json:"monthly_installment"`
	DueDate            time.Time `json:"due_date"`
	PaidAmount         float64   `json:"paid_amount"`
	RemainingAmount    float64   `json:"remaining_amount"`
	Status             Status    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Installment struct {
	ID            int64      `json:"id"`
	DebtID        int64      `json:"debt_id"`
	InstallmentNo int        `json:"installment_no"`
	DueDate       time.Time  `json:"due_date"`
	Amount        float64    `json:"amount"`
	Status        Status     `json:"status"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Payment struct {
	ID            int64     `json:"id"`
	DebtID        int64     `json:"debt_id"`
	InstallmentID *int64    `json:"installment_id,omitempty"`
	Amount        float64   `json:"amount"`
	PaymentDate   time.Time `json:"payment_date"`
	ProofImage    string    `json:"proof_image"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DebtSummary struct {
	Debt
	PaidInstallments    int64 `json:"paid_installments"`
	UnpaidInstallments  int64 `json:"unpaid_installments"`
	OverdueInstallments int64 `json:"overdue_installments"`
}

type DebtDetail struct {
	DebtSummary
	Installments []Installment `json:"installments"`
	Payments     []Payment     `json:"payments"`
}
