package alerts

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound     = errors.New("alert not found")
	ErrInvalidInput = errors.New("invalid alert input")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Evaluate(ctx context.Context, userID int64, input EvaluateInput) ([]Alert, error) {
	spikeMultiplier := 1.5
	if input.DailySpikeMultiplier != nil {
		if *input.DailySpikeMultiplier <= 0 {
			return nil, ErrInvalidInput
		}
		spikeMultiplier = *input.DailySpikeMultiplier
	}

	now := time.Now()

	monthExpense, err := s.repo.CurrentMonthExpense(ctx, userID)
	if err != nil {
		return nil, err
	}
	todayExpense, err := s.repo.TodayExpense(ctx, userID)
	if err != nil {
		return nil, err
	}
	alerts := make([]Alert, 0, 1)

	daysElapsed := float64(now.Day())
	if daysElapsed < 1 {
		daysElapsed = 1
	}
	averageDaily := monthExpense / daysElapsed
	spikeThreshold := averageDaily * spikeMultiplier
	if todayExpense > spikeThreshold && todayExpense > 0 {
		alert := Alert{
			UserID:         userID,
			Type:           AlertTypeDailySpendingSpike,
			Title:          "Daily spending spike detected",
			Message:        fmt.Sprintf("Today's spending %.2f is above the %.2f daily threshold based on the current month average.", todayExpense, spikeThreshold),
			Severity:       AlertSeverityWarning,
			MetricValue:    todayExpense,
			ThresholdValue: spikeThreshold,
			DedupeKey:      fmt.Sprintf("daily-spike:%s", now.Format("2006-01-02")),
		}

		stored, err := s.repo.UpsertAlert(ctx, alert)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, stored)
	}

	return alerts, nil
}

func (s *Service) List(ctx context.Context, userID int64, filter AlertListFilter) ([]Alert, error) {
	return s.repo.ListAlerts(ctx, userID, filter)
}

func (s *Service) MarkRead(ctx context.Context, userID, alertID int64) (Alert, error) {
	return s.repo.MarkAlertRead(ctx, userID, alertID, true)
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}
