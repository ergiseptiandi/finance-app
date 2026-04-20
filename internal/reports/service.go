package reports

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"finance-backend/internal/wallet"
)

var (
	ErrNotFound     = errors.New("reports data not found")
	ErrInvalidInput = errors.New("invalid reports input")
	nowFunc         = time.Now
)

type Service struct {
	repo     Repository
	balances wallet.BalanceProvider
}

func NewService(repo Repository, balances wallet.BalanceProvider) *Service {
	return &Service{repo: repo, balances: balances}
}

func (s *Service) ExpenseByCategory(ctx context.Context, userID int64, filter ReportsFilter) (ExpenseByCategoryReport, error) {
	start, end, period, err := s.resolveCategoryRange(filter)
	if err != nil {
		return ExpenseByCategoryReport{}, err
	}

	items, err := s.repo.ExpenseByCategory(ctx, userID, start, end)
	if err != nil {
		return ExpenseByCategoryReport{}, err
	}
	if items == nil {
		items = []CategoryExpense{}
	}

	total := 0.0
	for _, item := range items {
		total += item.Amount
	}

	for i := range items {
		if total > 0 {
			items[i].Percentage = round2((items[i].Amount / total) * 100)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Amount == items[j].Amount {
			return items[i].Category < items[j].Category
		}
		return items[i].Amount > items[j].Amount
	})

	summary := ExpenseByCategorySummary{
		TotalExpense:  round2(total),
		CategoryCount: len(items),
	}
	if len(items) > 0 {
		summary.TopCategory = items[0].Category
	}

	return ExpenseByCategoryReport{
		Period:  period,
		Summary: summary,
		Items:   items,
	}, nil
}

func (s *Service) SpendingTrends(ctx context.Context, userID int64, filter ReportsFilter) (SpendingTrendsReport, error) {
	start, end, period, groupBy, err := s.resolveTrendRange(filter)
	if err != nil {
		return SpendingTrendsReport{}, err
	}

	values, err := s.repo.TrendByPeriod(ctx, userID, start, end, groupBy)
	if err != nil {
		return SpendingTrendsReport{}, err
	}

	items := buildTrendPoints(start, end, groupBy, values)
	return SpendingTrendsReport{
		Period:  period,
		GroupBy: string(groupBy),
		Items:   items,
	}, nil
}

func (s *Service) HighestSpendingCategory(ctx context.Context, userID int64, filter ReportsFilter) (HighestSpendingCategoryReport, error) {
	report, err := s.ExpenseByCategory(ctx, userID, filter)
	if err != nil {
		return HighestSpendingCategoryReport{}, err
	}

	if len(report.Items) == 0 {
		return HighestSpendingCategoryReport{Period: report.Period}, nil
	}

	top := report.Items[0]
	return HighestSpendingCategoryReport{
		Period:           report.Period,
		Category:         top.Category,
		Amount:           top.Amount,
		Percentage:       top.Percentage,
		TransactionCount: top.TransactionCount,
	}, nil
}

func (s *Service) AverageDailySpending(ctx context.Context, userID int64, filter ReportsFilter) (AverageDailySpendingReport, error) {
	start, end, period, err := s.resolveCategoryRange(filter)
	if err != nil {
		return AverageDailySpendingReport{}, err
	}

	values, err := s.repo.TrendByPeriod(ctx, userID, start, end, TrendGroupByDay)
	if err != nil {
		return AverageDailySpendingReport{}, err
	}

	points := buildTrendPoints(start, end, TrendGroupByDay, values)
	totalExpense := 0.0
	highest := 0.0
	lowest := 0.0
	for i, point := range points {
		totalExpense += point.Expense
		if i == 0 || point.Expense > highest {
			highest = point.Expense
		}
		if i == 0 || point.Expense < lowest {
			lowest = point.Expense
		}
	}

	daysCount := inclusiveDayCount(start, end)
	average := 0.0
	if daysCount > 0 {
		average = round2(totalExpense / float64(daysCount))
	}

	return AverageDailySpendingReport{
		Period:               period,
		TotalExpense:         round2(totalExpense),
		DaysCount:            daysCount,
		AverageDailySpending: average,
		HighestDailySpending: round2(highest),
		LowestDailySpending:  round2(lowest),
	}, nil
}

func (s *Service) RemainingBalance(ctx context.Context, userID int64, filter ReportsFilter) (RemainingBalanceReport, error) {
	if filter.IsZero() {
		income, err := s.repo.AllTimeIncome(ctx, userID)
		if err != nil {
			return RemainingBalanceReport{}, err
		}

		expense, err := s.repo.AllTimeExpense(ctx, userID)
		if err != nil {
			return RemainingBalanceReport{}, err
		}

		remainingBalance := income - expense
		if s.balances != nil {
			if balance, err := s.balances.TotalBalance(ctx, userID); err == nil {
				remainingBalance = balance
			} else {
				return RemainingBalanceReport{}, err
			}
		}

		return RemainingBalanceReport{
			Period:           ReportPeriod{Mode: "all-time"},
			TotalIncome:      income,
			TotalExpense:     expense,
			RemainingBalance: remainingBalance,
			SavingsRate:      percentageOf(income-expense, income),
			ExpenseRatio:     percentageOf(expense, income),
		}, nil
	}

	start, end, period, err := s.resolveBalanceRange(filter)
	if err != nil {
		return RemainingBalanceReport{}, err
	}

	income, err := s.repo.IncomeBetween(ctx, userID, start, end)
	if err != nil {
		return RemainingBalanceReport{}, err
	}

	expense, err := s.repo.ExpenseBetween(ctx, userID, start, end)
	if err != nil {
		return RemainingBalanceReport{}, err
	}

	remainingBalance := income - expense
	return RemainingBalanceReport{
		Period:           period,
		TotalIncome:      income,
		TotalExpense:     expense,
		RemainingBalance: remainingBalance,
		SavingsRate:      percentageOf(remainingBalance, income),
		ExpenseRatio:     percentageOf(expense, income),
	}, nil
}

func (s *Service) resolveCategoryRange(filter ReportsFilter) (time.Time, time.Time, ReportPeriod, error) {
	if filter.IsZero() {
		start, end := currentMonthRange()
		return start, end, buildMonthPeriod(start, end), nil
	}

	switch filter.Mode {
	case ReportFilterModeMonth:
		start, end := rangeFromFilter(filter)
		return start, end, buildPeriodFromFilter(filter), nil
	case ReportFilterModeYear:
		start, end := rangeFromFilter(filter)
		return start, end, buildPeriodFromFilter(filter), nil
	case ReportFilterModeCustom:
		start, end := rangeFromFilter(filter)
		return start, end, buildPeriodFromFilter(filter), nil
	default:
		start, end := currentMonthRange()
		return start, end, buildMonthPeriod(start, end), nil
	}
}

func (s *Service) resolveTrendRange(filter ReportsFilter) (time.Time, time.Time, ReportPeriod, TrendGroupBy, error) {
	if filter.IsZero() {
		start, end := rollingTwelveMonthsRange()
		return start, end, buildRollingPeriod(start, end), TrendGroupByMonth, nil
	}

	switch filter.Mode {
	case ReportFilterModeMonth:
		start, end := rangeFromFilter(filter)
		return start, end, buildPeriodFromFilter(filter), TrendGroupByDay, nil
	case ReportFilterModeYear:
		start, end := rangeFromFilter(filter)
		return start, end, buildPeriodFromFilter(filter), TrendGroupByMonth, nil
	case ReportFilterModeCustom:
		start, end := rangeFromFilter(filter)
		groupBy := TrendGroupByDay
		if inclusiveDayCount(start, end) > 93 {
			groupBy = TrendGroupByMonth
		}
		return start, end, buildPeriodFromFilter(filter), groupBy, nil
	default:
		start, end := rollingTwelveMonthsRange()
		return start, end, buildRollingPeriod(start, end), TrendGroupByMonth, nil
	}
}

func (s *Service) resolveBalanceRange(filter ReportsFilter) (time.Time, time.Time, ReportPeriod, error) {
	switch filter.Mode {
	case ReportFilterModeMonth, ReportFilterModeYear, ReportFilterModeCustom:
		start, end := rangeFromFilter(filter)
		return start, end, buildPeriodFromFilter(filter), nil
	default:
		return time.Time{}, time.Time{}, ReportPeriod{}, fmt.Errorf("%w: unsupported report filter", ErrInvalidInput)
	}
}

func rangeFromFilter(filter ReportsFilter) (time.Time, time.Time) {
	if filter.StartDate == nil || filter.EndDate == nil {
		return time.Time{}, time.Time{}
	}
	return *filter.StartDate, *filter.EndDate
}

func buildPeriodFromFilter(filter ReportsFilter) ReportPeriod {
	start, end := rangeFromFilter(filter)
	if start.IsZero() || end.IsZero() {
		return ReportPeriod{}
	}

	period := ReportPeriod{
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.AddDate(0, 0, -1).Format("2006-01-02"),
	}

	switch filter.Mode {
	case ReportFilterModeMonth:
		period.Mode = string(ReportFilterModeMonth)
		period.Month = start.Format("2006-01")
	case ReportFilterModeYear:
		period.Mode = string(ReportFilterModeYear)
		period.Year = start.Year()
	case ReportFilterModeCustom:
		period.Mode = string(ReportFilterModeCustom)
	default:
		period.Mode = string(ReportFilterModeCustom)
	}

	return period
}

func buildMonthPeriod(start, end time.Time) ReportPeriod {
	return ReportPeriod{
		Mode:      string(ReportFilterModeMonth),
		Month:     start.Format("2006-01"),
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.AddDate(0, 0, -1).Format("2006-01-02"),
	}
}

func buildRollingPeriod(start, end time.Time) ReportPeriod {
	return ReportPeriod{
		Mode:      "rolling_12_months",
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.AddDate(0, 0, -1).Format("2006-01-02"),
	}
}

func currentMonthRange() (time.Time, time.Time) {
	now := nowFunc()
	start := startOfMonth(now)
	return start, start.AddDate(0, 1, 0)
}

func rollingTwelveMonthsRange() (time.Time, time.Time) {
	now := nowFunc()
	end := startOfMonth(now).AddDate(0, 1, 0)
	start := startOfMonth(now).AddDate(0, -11, 0)
	return start, end
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func buildTrendPoints(start, end time.Time, groupBy TrendGroupBy, values map[string]TrendTotals) []TrendPoint {
	points := make([]TrendPoint, 0)

	switch groupBy {
	case TrendGroupByDay:
		for day := startOfDay(start); day.Before(end); day = day.AddDate(0, 0, 1) {
			key := day.Format("2006-01-02")
			total := values[key]
			points = append(points, TrendPoint{
				Period:      key,
				Income:      round2(total.Income),
				Expense:     round2(total.Expense),
				NetCashflow: round2(total.Income - total.Expense),
			})
		}
	case TrendGroupByMonth:
		monthStart := startOfMonth(start)
		monthEnd := startOfMonth(end.AddDate(0, 0, -1)).AddDate(0, 1, 0)
		for month := monthStart; month.Before(monthEnd); month = month.AddDate(0, 1, 0) {
			key := month.Format("2006-01")
			total := values[key]
			points = append(points, TrendPoint{
				Period:      key,
				Income:      round2(total.Income),
				Expense:     round2(total.Expense),
				NetCashflow: round2(total.Income - total.Expense),
			})
		}
	}

	return points
}

func inclusiveDayCount(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}

	return int(end.Sub(start).Hours() / 24)
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func percentageOf(part, whole float64) float64 {
	if whole <= 0 {
		return 0
	}

	return round2((part / whole) * 100)
}
