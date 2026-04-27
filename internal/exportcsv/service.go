package exportcsv

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"time"

	"finance-backend/internal/debt"
	"finance-backend/internal/reports"
	"finance-backend/internal/transaction"
)

type Service struct {
	transactions *transaction.Service
	debts        *debt.Service
	reports      *reports.Service
}

func NewService(transactions *transaction.Service, debts *debt.Service, reports *reports.Service) *Service {
	return &Service{
		transactions: transactions,
		debts:        debts,
		reports:      reports,
	}
}

func (s *Service) Export(ctx context.Context, userID int64, scope Scope, period Period) (Result, error) {
	switch scope {
	case ScopeTransactions:
		return s.exportTransactions(ctx, userID, period)
	case ScopeDebts:
		return s.exportDebts(ctx, userID, period)
	case ScopeReports:
		return s.exportReports(ctx, userID, period)
	default:
		return Result{}, errors.New("scope must be transactions, debts, or reports")
	}
}

type csvRow struct {
	RecordType        string
	ID                string
	ParentID          string
	Name              string
	Type              string
	Status            string
	Category          string
	Amount            string
	AmountSecondary   string
	CategoryID        string
	CategoryName      string
	WalletID          string
	WalletName        string
	Date              string
	DueDate           string
	InstallmentNo     string
	Note              string
	Description       string
	TotalAmount       string
	MonthlyInstallment string
	PaidAmount        string
	RemainingAmount   string
	Balance           string
	OpeningBalance    string
	CurrentAmount     string
	RemainingBudget   string
	MonthlyBudget     string
	Spent             string
	UsageRate         string
	OverBudgetAmount  string
	Percentage        string
	TransactionCount  string
	Progress          string
	ProofImage        string
	PaymentDate       string
	TransferDate      string
	CreatedAt         string
	UpdatedAt         string
	PaidAt            string
	Period            string
	Label             string
	Income            string
	Expense           string
	NetCashflow       string
	RemainingBalance  string
	SavingsRate       string
	ExpenseRatio      string
	DaysCount         string
	AverageDaily      string
	HighestDaily      string
	LowestDaily       string
	PayloadJSON       string
}

func buildCSV(headers []string, rows []csvRow) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("\ufeff")

	writer := csv.NewWriter(&buf)
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, row := range rows {
		if err := writer.Write(row.values(headers)); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r csvRow) values(headers []string) []string {
	values := make([]string, 0, len(headers))
	for _, header := range headers {
		values = append(values, r.value(header))
	}
	return values
}

func (r csvRow) value(header string) string {
	switch header {
	case "record_type":
		return r.RecordType
	case "id":
		return r.ID
	case "parent_id":
		return r.ParentID
	case "name":
		return r.Name
	case "type":
		return r.Type
	case "status":
		return r.Status
	case "category":
		return r.Category
	case "amount":
		return r.Amount
	case "amount_secondary":
		return r.AmountSecondary
	case "category_id":
		return r.CategoryID
	case "category_name":
		return r.CategoryName
	case "wallet_id":
		return r.WalletID
	case "wallet_name":
		return r.WalletName
	case "date":
		return r.Date
	case "due_date":
		return r.DueDate
	case "installment_no":
		return r.InstallmentNo
	case "note":
		return r.Note
	case "description":
		return r.Description
	case "total_amount":
		return r.TotalAmount
	case "monthly_installment":
		return r.MonthlyInstallment
	case "paid_amount":
		return r.PaidAmount
	case "remaining_amount":
		return r.RemainingAmount
	case "balance":
		return r.Balance
	case "opening_balance":
		return r.OpeningBalance
	case "current_amount":
		return r.CurrentAmount
	case "remaining_budget":
		return r.RemainingBudget
	case "monthly_budget":
		return r.MonthlyBudget
	case "spent":
		return r.Spent
	case "usage_rate":
		return r.UsageRate
	case "over_budget_amount":
		return r.OverBudgetAmount
	case "percentage":
		return r.Percentage
	case "transaction_count":
		return r.TransactionCount
	case "progress":
		return r.Progress
	case "proof_image":
		return r.ProofImage
	case "payment_date":
		return r.PaymentDate
	case "transfer_date":
		return r.TransferDate
	case "created_at":
		return r.CreatedAt
	case "updated_at":
		return r.UpdatedAt
	case "paid_at":
		return r.PaidAt
	case "period":
		return r.Period
	case "label":
		return r.Label
	case "income":
		return r.Income
	case "expense":
		return r.Expense
	case "net_cashflow":
		return r.NetCashflow
	case "remaining_balance":
		return r.RemainingBalance
	case "savings_rate":
		return r.SavingsRate
	case "expense_ratio":
		return r.ExpenseRatio
	case "days_count":
		return r.DaysCount
	case "average_daily":
		return r.AverageDaily
	case "highest_daily":
		return r.HighestDaily
	case "lowest_daily":
		return r.LowestDaily
	case "payload_json":
		return r.PayloadJSON
	default:
		return ""
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.In(time.Local).Format("2006-01-02 15:04:05")
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.In(time.Local).Format("2006-01-02")
}

func formatFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func (s *Service) exportTransactions(ctx context.Context, userID int64, period Period) (Result, error) {
	if s.transactions == nil {
		return Result{}, errors.New("transaction service is required")
	}

	filter := transaction.ListFilter{
		StartDate: period.StartDate,
		EndDate:   period.EndDate,
		Page:      1,
		PerPage:   100,
	}

	rows := make([]csvRow, 0)
	totalCount := 0

	for {
		page, err := s.transactions.List(ctx, userID, filter)
		if err != nil {
			return Result{}, err
		}

		totalCount += len(page.Data)
		for _, item := range page.Data {
			rows = append(rows, csvRow{
				RecordType:      "transaction",
				ID:              fmt.Sprintf("%d", item.ID),
				WalletID:        fmt.Sprintf("%d", item.WalletID),
				Type:            string(item.Type),
				Category:        item.Category,
				Amount:          formatFloat(item.Amount),
				Date:            formatDate(item.Date),
				Description:     item.Description,
				CreatedAt:       formatTime(item.CreatedAt),
				UpdatedAt:       formatTime(item.UpdatedAt),
				PayloadJSON:     "",
			})
		}

		if page.TotalPages <= filter.Page || len(page.Data) == 0 {
			break
		}

		filter.Page++
	}

	csvBytes, err := buildCSV([]string{
		"record_type",
		"id",
		"wallet_id",
		"type",
		"category",
		"amount",
		"date",
		"description",
		"created_at",
		"updated_at",
	}, rows)
	if err != nil {
		return Result{}, err
	}

	return Result{
		FileName:    buildFileName("transactions", period),
		CSV:         csvBytes,
		Partial:     false,
		RecordCount: totalCount,
	}, nil
}

func (s *Service) exportDebts(ctx context.Context, userID int64, period Period) (Result, error) {
	if s.debts == nil {
		return Result{}, errors.New("debt service is required")
	}

	items, err := s.debts.List(ctx, userID)
	if err != nil {
		return Result{}, err
	}

	filtered := make([]debt.DebtSummary, 0, len(items))
	for _, item := range items {
		if period.StartDate != nil && period.EndDate != nil {
			dueDate := item.DueDate.In(time.Local)
			start := period.StartDate.In(time.Local)
			end := period.EndDate.In(time.Local)
			if dueDate.Before(start) || dueDate.After(end) {
				continue
			}
		}
		filtered = append(filtered, item)
	}

	rows := make([]csvRow, 0)
	partial := false

	for _, item := range filtered {
		detail, err := s.debts.Detail(ctx, userID, item.ID)
		if err != nil {
			partial = true
			rows = append(rows, debtSummaryRow(item))
			continue
		}

		rows = append(rows, debtSummaryRow(detail.DebtSummary))

		for _, installment := range detail.Installments {
			rows = append(rows, csvRow{
				RecordType:    "installment",
				ParentID:      fmt.Sprintf("%d", item.ID),
				ID:            fmt.Sprintf("%d", installment.ID),
				Name:          item.Name,
				InstallmentNo: fmt.Sprintf("%d", installment.InstallmentNo),
				DueDate:       formatDate(installment.DueDate),
				Amount:        formatFloat(installment.Amount),
				Status:        string(installment.Status),
				PaidAt:        formatTimePtr(installment.PaidAt),
				CreatedAt:     formatTime(installment.CreatedAt),
				UpdatedAt:     formatTime(installment.UpdatedAt),
			})
		}

		for _, payment := range detail.Payments {
			rows = append(rows, csvRow{
				RecordType:   "payment",
				ParentID:     fmt.Sprintf("%d", item.ID),
				ID:           fmt.Sprintf("%d", payment.ID),
				Name:         item.Name,
				WalletID:     fmt.Sprintf("%d", payment.WalletID),
				Amount:       formatFloat(payment.Amount),
				PaymentDate:  formatDate(payment.PaymentDate),
				ProofImage:   payment.ProofImage,
				CreatedAt:    formatTime(payment.CreatedAt),
				UpdatedAt:    formatTime(payment.UpdatedAt),
				InstallmentNo: installmentRef(payment.InstallmentID),
			})
		}
	}

	csvBytes, err := buildCSV([]string{
		"record_type",
		"id",
		"parent_id",
		"name",
		"installment_no",
		"status",
		"due_date",
		"amount",
		"total_amount",
		"monthly_installment",
		"paid_amount",
		"remaining_amount",
		"wallet_id",
		"payment_date",
		"proof_image",
		"paid_at",
		"created_at",
		"updated_at",
	}, rows)
	if err != nil {
		return Result{}, err
	}

	return Result{
		FileName:    buildFileName("debts", period),
		CSV:         csvBytes,
		Partial:     partial,
		RecordCount: len(rows),
	}, nil
}

func (s *Service) exportReports(ctx context.Context, userID int64, period Period) (Result, error) {
	if s.reports == nil {
		return Result{}, errors.New("reports service is required")
	}

	filter := reports.ReportsFilter{}
	if period.StartDate != nil && period.EndDate != nil {
		start := *period.StartDate
		end := period.EndDate.AddDate(0, 0, 1)
		filter.Mode = reports.ReportFilterModeCustom
		filter.StartDate = &start
		filter.EndDate = &end
	}

	rows := make([]csvRow, 0)
	partial := false

	expenseReport, err := s.reports.ExpenseByCategory(ctx, userID, filter)
	if err != nil {
		partial = true
	} else {
		for _, item := range expenseReport.Items {
			rows = append(rows, csvRow{
				RecordType:       "expense_category",
				Period:           reportPeriodLabel(expenseReport.Period),
				Label:            item.Category,
				Amount:           formatFloat(item.Amount),
				Percentage:       formatFloat(item.Percentage),
				TransactionCount: fmt.Sprintf("%d", item.TransactionCount),
			})
		}
		rows = append(rows, csvRow{
			RecordType:       "expense_summary",
			Period:           reportPeriodLabel(expenseReport.Period),
			Amount:           formatFloat(expenseReport.Summary.TotalExpense),
			TransactionCount: fmt.Sprintf("%d", expenseReport.Summary.CategoryCount),
			Label:            expenseReport.Summary.TopCategory,
		})
	}

	trendReport, err := s.reports.SpendingTrends(ctx, userID, filter)
	if err != nil {
		partial = true
	} else {
		for _, item := range trendReport.Items {
			rows = append(rows, csvRow{
				RecordType:  "spending_trend",
				Period:      reportPeriodLabel(trendReport.Period),
				Label:       item.Period,
				Income:      formatFloat(item.Income),
				Expense:     formatFloat(item.Expense),
				NetCashflow: formatFloat(item.NetCashflow),
			})
		}
	}

	highestReport, err := s.reports.HighestSpendingCategory(ctx, userID, filter)
	if err != nil {
		partial = true
	} else {
		rows = append(rows, csvRow{
			RecordType:       "highest_category",
			Period:           reportPeriodLabel(highestReport.Period),
			Label:            highestReport.Category,
			Amount:           formatFloat(highestReport.Amount),
			Percentage:       formatFloat(highestReport.Percentage),
			TransactionCount: fmt.Sprintf("%d", highestReport.TransactionCount),
		})
	}

	averageReport, err := s.reports.AverageDailySpending(ctx, userID, filter)
	if err != nil {
		partial = true
	} else {
		rows = append(rows, csvRow{
			RecordType:    "average_daily",
			Period:        reportPeriodLabel(averageReport.Period),
			Amount:        formatFloat(averageReport.TotalExpense),
			DaysCount:     fmt.Sprintf("%d", averageReport.DaysCount),
			AverageDaily:  formatFloat(averageReport.AverageDailySpending),
			HighestDaily:  formatFloat(averageReport.HighestDailySpending),
			LowestDaily:   formatFloat(averageReport.LowestDailySpending),
		})
	}

	remainingReport, err := s.reports.RemainingBalance(ctx, userID, filter)
	if err != nil {
		partial = true
	} else {
		rows = append(rows, csvRow{
			RecordType:       "remaining_balance",
			Period:           reportPeriodLabel(remainingReport.Period),
			Income:           formatFloat(remainingReport.TotalIncome),
			Expense:          formatFloat(remainingReport.TotalExpense),
			RemainingBalance: formatFloat(remainingReport.RemainingBalance),
			SavingsRate:      formatFloat(remainingReport.SavingsRate),
			ExpenseRatio:     formatFloat(remainingReport.ExpenseRatio),
		})
	}

	csvBytes, err := buildCSV([]string{
		"record_type",
		"period",
		"label",
		"amount",
		"percentage",
		"transaction_count",
		"income",
		"expense",
		"net_cashflow",
		"remaining_balance",
		"savings_rate",
		"expense_ratio",
		"days_count",
		"average_daily",
		"highest_daily",
		"lowest_daily",
	}, rows)
	if err != nil {
		return Result{}, err
	}

	return Result{
		FileName:    buildFileName("reports", period),
		CSV:         csvBytes,
		Partial:     partial,
		RecordCount: len(rows),
	}, nil
}

func buildFileName(scope string, period Period) string {
	suffix := time.Now().Format("2006-01-02")
	if period.Label != "" {
		suffix = period.Label
	}

	return fmt.Sprintf("finance-go-%s-%s.csv", scope, suffix)
}

func debtSummaryRow(item debt.DebtSummary) csvRow {
	return csvRow{
		RecordType:         "debt",
		ID:                 fmt.Sprintf("%d", item.ID),
		Name:               item.Name,
		TotalAmount:        formatFloat(item.TotalAmount),
		MonthlyInstallment: formatFloat(item.MonthlyInstallment),
		DueDate:            formatDate(item.DueDate),
		PaidAmount:         formatFloat(item.PaidAmount),
		RemainingAmount:    formatFloat(item.RemainingAmount),
		Status:             string(item.Status),
		CreatedAt:          formatTime(item.CreatedAt),
		UpdatedAt:          formatTime(item.UpdatedAt),
	}
}

func formatTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return formatTime(*value)
}

func installmentRef(value *int64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}

func reportPeriodLabel(period reports.ReportPeriod) string {
	switch {
	case period.Month != "":
		return period.Month
	case period.Year != 0:
		return fmt.Sprintf("%d", period.Year)
	case period.StartDate != "" || period.EndDate != "":
		if period.StartDate != "" && period.EndDate != "" {
			return fmt.Sprintf("%s - %s", period.StartDate, period.EndDate)
		}
		if period.StartDate != "" {
			return period.StartDate
		}
		return period.EndDate
	default:
		return period.Mode
	}
}
