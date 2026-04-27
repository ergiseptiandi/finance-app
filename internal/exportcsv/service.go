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

func (s *Service) Export(ctx context.Context, userID int64, scope Scope, period Period, lang Language) (Result, error) {
	switch scope {
	case ScopeTransactions:
		return s.exportTransactions(ctx, userID, period, lang)
	case ScopeDebts:
		return s.exportDebts(ctx, userID, period, lang)
	case ScopeReports:
		return s.exportReports(ctx, userID, period, lang)
	default:
		return Result{}, errors.New("scope must be transactions, debts, or reports")
	}
}

func (s *Service) ExportXLSX(ctx context.Context, userID int64, scope Scope, period Period, lang Language) (Result, error) {
	result, err := s.Export(ctx, userID, scope, period, lang)
	if err != nil {
		return Result{}, err
	}

	xlsxBytes, err := buildXLSX(result.CSV, scope, period, lang, result.Partial)
	if err != nil {
		return Result{}, err
	}

	result.FileName = buildFileName(string(scope), period, "xlsx")
	result.CSV = nil
	result.XLSX = xlsxBytes
	return result, nil
}

type csvColumn struct {
	Key   string
	Label string
}

type csvRow struct {
	RecordType         string
	ID                 string
	ParentID           string
	Name               string
	Type               string
	Status             string
	Category           string
	Amount             string
	AmountSecondary    string
	CategoryID         string
	CategoryName       string
	WalletID           string
	WalletName         string
	Date               string
	DueDate            string
	InstallmentNo      string
	Note               string
	Description        string
	TotalAmount        string
	MonthlyInstallment string
	PaidAmount         string
	RemainingAmount    string
	Balance            string
	OpeningBalance     string
	CurrentAmount      string
	RemainingBudget    string
	MonthlyBudget      string
	Spent              string
	UsageRate          string
	OverBudgetAmount   string
	Percentage         string
	TransactionCount   string
	Progress           string
	ProofImage         string
	PaymentDate        string
	TransferDate       string
	CreatedAt          string
	UpdatedAt          string
	PaidAt             string
	Period             string
	Label              string
	Income             string
	Expense            string
	NetCashflow        string
	RemainingBalance   string
	SavingsRate        string
	ExpenseRatio       string
	DaysCount          string
	AverageDaily       string
	HighestDaily       string
	LowestDaily        string
	PayloadJSON        string
}

func delimiterForLanguage(lang Language) rune {
	if lang == LanguageEN {
		return ','
	}
	return ';'
}

func buildCSV(lang Language, columns []csvColumn, rows []csvRow) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("\ufeff")

	writer := csv.NewWriter(&buf)
	writer.Comma = delimiterForLanguage(lang)
	writer.UseCRLF = true
	headers := make([]string, 0, len(columns))
	for _, column := range columns {
		headers = append(headers, column.Label)
	}

	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, row := range rows {
		if err := writer.Write(row.values(columns)); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r csvRow) values(columns []csvColumn) []string {
	values := make([]string, 0, len(columns))
	for _, column := range columns {
		values = append(values, r.value(column.Key))
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

func text(lang Language, id string, en string) string {
	if lang == LanguageEN {
		return en
	}
	return id
}

func transactionTypeLabel(lang Language, value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "income":
		return text(lang, "Pemasukan", "Income")
	case "expense":
		return text(lang, "Pengeluaran", "Expense")
	default:
		return value
	}
}

func debtStatusLabel(lang Language, value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending":
		return text(lang, "Menunggu", "Pending")
	case "paid":
		return text(lang, "Lunas", "Paid")
	case "overdue":
		return text(lang, "Terlambat", "Overdue")
	default:
		return value
	}
}

func transactionColumns(lang Language) []csvColumn {
	return []csvColumn{
		{Key: "record_type", Label: text(lang, "Jenis Data", "Record Type")},
		{Key: "date", Label: text(lang, "Tanggal Transaksi", "Transaction Date")},
		{Key: "type", Label: text(lang, "Tipe", "Type")},
		{Key: "category", Label: text(lang, "Kategori", "Category")},
		{Key: "amount", Label: text(lang, "Nominal", "Amount")},
		{Key: "description", Label: text(lang, "Keterangan", "Description")},
		{Key: "created_at", Label: text(lang, "Dibuat Pada", "Created At")},
		{Key: "updated_at", Label: text(lang, "Diubah Pada", "Updated At")},
	}
}

func debtColumns(lang Language) []csvColumn {
	return []csvColumn{
		{Key: "record_type", Label: text(lang, "Jenis Data", "Record Type")},
		{Key: "name", Label: text(lang, "Nama Utang", "Debt Name")},
		{Key: "status", Label: text(lang, "Status", "Status")},
		{Key: "due_date", Label: text(lang, "Jatuh Tempo", "Due Date")},
		{Key: "total_amount", Label: text(lang, "Total Utang", "Total Amount")},
		{Key: "monthly_installment", Label: text(lang, "Cicilan Bulanan", "Monthly Installment")},
		{Key: "paid_amount", Label: text(lang, "Sudah Dibayar", "Paid Amount")},
		{Key: "remaining_amount", Label: text(lang, "Sisa Tagihan", "Remaining Amount")},
		{Key: "installment_no", Label: text(lang, "Nomor Cicilan", "Installment No.")},
		{Key: "amount", Label: text(lang, "Nominal", "Amount")},
		{Key: "payment_date", Label: text(lang, "Tanggal Pembayaran", "Payment Date")},
		{Key: "proof_image", Label: text(lang, "Bukti", "Proof")},
		{Key: "paid_at", Label: text(lang, "Waktu Pelunasan", "Paid At")},
		{Key: "created_at", Label: text(lang, "Dibuat Pada", "Created At")},
		{Key: "updated_at", Label: text(lang, "Diubah Pada", "Updated At")},
	}
}

func reportColumns(lang Language) []csvColumn {
	return []csvColumn{
		{Key: "record_type", Label: text(lang, "Jenis Data", "Record Type")},
		{Key: "period", Label: text(lang, "Periode", "Period")},
		{Key: "label", Label: text(lang, "Nama", "Name")},
		{Key: "amount", Label: text(lang, "Nominal", "Amount")},
		{Key: "percentage", Label: text(lang, "Persentase", "Percentage")},
		{Key: "transaction_count", Label: text(lang, "Jumlah Transaksi", "Transaction Count")},
		{Key: "income", Label: text(lang, "Pemasukan", "Income")},
		{Key: "expense", Label: text(lang, "Pengeluaran", "Expense")},
		{Key: "net_cashflow", Label: text(lang, "Arus Kas Bersih", "Net Cashflow")},
		{Key: "remaining_balance", Label: text(lang, "Sisa Saldo", "Remaining Balance")},
		{Key: "savings_rate", Label: text(lang, "Rasio Tabungan", "Savings Rate")},
		{Key: "expense_ratio", Label: text(lang, "Rasio Pengeluaran", "Expense Ratio")},
		{Key: "days_count", Label: text(lang, "Jumlah Hari", "Days Count")},
		{Key: "average_daily", Label: text(lang, "Rata-rata Harian", "Average Daily")},
		{Key: "highest_daily", Label: text(lang, "Tertinggi Harian", "Highest Daily")},
		{Key: "lowest_daily", Label: text(lang, "Terendah Harian", "Lowest Daily")},
	}
}

func (s *Service) exportTransactions(ctx context.Context, userID int64, period Period, lang Language) (Result, error) {
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
				RecordType:  text(lang, "Transaksi", "Transaction"),
				Type:        transactionTypeLabel(lang, string(item.Type)),
				Category:    item.Category,
				Amount:      formatFloat(item.Amount),
				Date:        formatDate(item.Date),
				Description: item.Description,
				CreatedAt:   formatTime(item.CreatedAt),
				UpdatedAt:   formatTime(item.UpdatedAt),
			})
		}

		if page.TotalPages <= filter.Page || len(page.Data) == 0 {
			break
		}

		filter.Page++
	}

	csvBytes, err := buildCSV(lang, transactionColumns(lang), rows)
	if err != nil {
		return Result{}, err
	}

	return Result{
		FileName:    buildFileName("transactions", period, "csv"),
		CSV:         csvBytes,
		Partial:     false,
		RecordCount: totalCount,
	}, nil
}

func (s *Service) exportDebts(ctx context.Context, userID int64, period Period, lang Language) (Result, error) {
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
			rows = append(rows, debtSummaryRow(item, lang))
			continue
		}

		rows = append(rows, debtSummaryRow(detail.DebtSummary, lang))

		for _, installment := range detail.Installments {
			rows = append(rows, csvRow{
				RecordType:    text(lang, "Cicilan", "Installment"),
				Name:          item.Name,
				InstallmentNo: fmt.Sprintf("%d", installment.InstallmentNo),
				DueDate:       formatDate(installment.DueDate),
				Amount:        formatFloat(installment.Amount),
				Status:        debtStatusLabel(lang, string(installment.Status)),
				PaidAt:        formatTimePtr(installment.PaidAt),
				CreatedAt:     formatTime(installment.CreatedAt),
				UpdatedAt:     formatTime(installment.UpdatedAt),
			})
		}

		for _, payment := range detail.Payments {
			rows = append(rows, csvRow{
				RecordType:    text(lang, "Pembayaran", "Payment"),
				Name:          item.Name,
				Amount:        formatFloat(payment.Amount),
				PaymentDate:   formatDate(payment.PaymentDate),
				ProofImage:    payment.ProofImage,
				CreatedAt:     formatTime(payment.CreatedAt),
				UpdatedAt:     formatTime(payment.UpdatedAt),
				InstallmentNo: installmentRef(payment.InstallmentID),
			})
		}
	}

	csvBytes, err := buildCSV(lang, debtColumns(lang), rows)
	if err != nil {
		return Result{}, err
	}

	return Result{
		FileName:    buildFileName("debts", period, "csv"),
		CSV:         csvBytes,
		Partial:     partial,
		RecordCount: len(rows),
	}, nil
}

func (s *Service) exportReports(ctx context.Context, userID int64, period Period, lang Language) (Result, error) {
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
				RecordType:       text(lang, "Kategori Pengeluaran", "Expense Category"),
				Period:           reportPeriodLabel(expenseReport.Period),
				Label:            item.Category,
				Amount:           formatFloat(item.Amount),
				Percentage:       formatFloat(item.Percentage),
				TransactionCount: fmt.Sprintf("%d", item.TransactionCount),
			})
		}
		rows = append(rows, csvRow{
			RecordType:       text(lang, "Ringkasan Pengeluaran", "Expense Summary"),
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
				RecordType:  text(lang, "Tren Pengeluaran", "Spending Trend"),
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
			RecordType:       text(lang, "Kategori Terbesar", "Highest Category"),
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
			RecordType:   text(lang, "Rata-rata Harian", "Average Daily"),
			Period:       reportPeriodLabel(averageReport.Period),
			Amount:       formatFloat(averageReport.TotalExpense),
			DaysCount:    fmt.Sprintf("%d", averageReport.DaysCount),
			AverageDaily: formatFloat(averageReport.AverageDailySpending),
			HighestDaily: formatFloat(averageReport.HighestDailySpending),
			LowestDaily:  formatFloat(averageReport.LowestDailySpending),
		})
	}

	remainingReport, err := s.reports.RemainingBalance(ctx, userID, filter)
	if err != nil {
		partial = true
	} else {
		rows = append(rows, csvRow{
			RecordType:       text(lang, "Sisa Saldo", "Remaining Balance"),
			Period:           reportPeriodLabel(remainingReport.Period),
			Income:           formatFloat(remainingReport.TotalIncome),
			Expense:          formatFloat(remainingReport.TotalExpense),
			RemainingBalance: formatFloat(remainingReport.RemainingBalance),
			SavingsRate:      formatFloat(remainingReport.SavingsRate),
			ExpenseRatio:     formatFloat(remainingReport.ExpenseRatio),
		})
	}

	csvBytes, err := buildCSV(lang, reportColumns(lang), rows)
	if err != nil {
		return Result{}, err
	}

	return Result{
		FileName:    buildFileName("reports", period, "csv"),
		CSV:         csvBytes,
		Partial:     partial,
		RecordCount: len(rows),
	}, nil
}

func buildFileName(scope string, period Period, ext string) string {
	suffix := time.Now().Format("2006-01-02")
	if period.Label != "" {
		suffix = period.Label
	}

	return fmt.Sprintf("finance-go-%s-%s.%s", scope, suffix, ext)
}

func debtSummaryRow(item debt.DebtSummary, lang Language) csvRow {
	return csvRow{
		RecordType:         text(lang, "Utang", "Debt"),
		Name:               item.Name,
		TotalAmount:        formatFloat(item.TotalAmount),
		MonthlyInstallment: formatFloat(item.MonthlyInstallment),
		DueDate:            formatDate(item.DueDate),
		PaidAmount:         formatFloat(item.PaidAmount),
		RemainingAmount:    formatFloat(item.RemainingAmount),
		Status:             debtStatusLabel(lang, string(item.Status)),
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
