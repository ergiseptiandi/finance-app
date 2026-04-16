package debt

import "time"

type CreateInput struct {
	Name               string    `json:"name"`
	TotalAmount        float64   `json:"total_amount"`
	MonthlyInstallment float64   `json:"monthly_installment"`
	DueDate            time.Time `json:"due_date"`
}

type UpdateInput struct {
	Name               *string    `json:"name,omitempty"`
	TotalAmount        *float64   `json:"total_amount,omitempty"`
	MonthlyInstallment *float64   `json:"monthly_installment,omitempty"`
	DueDate            *time.Time `json:"due_date,omitempty"`
}

type CreatePaymentInput struct {
	Amount      float64   `json:"amount"`
	PaymentDate time.Time `json:"payment_date"`
	ProofImage  string    `json:"-"`
}

type UpdatePaymentInput struct {
	Amount      *float64   `json:"amount,omitempty"`
	PaymentDate *time.Time `json:"payment_date,omitempty"`
	ProofImage  *string    `json:"-"`
}

type MarkInstallmentPaidInput struct {
	PaidAt *time.Time `json:"paid_at,omitempty"`
}
