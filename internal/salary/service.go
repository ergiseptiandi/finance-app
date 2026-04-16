package salary

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrNotFound     = errors.New("salary record not found")
	ErrInvalidInput = errors.New("invalid salary input")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID int64, input CreateInput) (SalaryRecord, error) {
	if input.Amount <= 0 {
		return SalaryRecord{}, errors.New("amount must be greater than zero")
	}
	if input.PaidAt.IsZero() {
		return SalaryRecord{}, errors.New("paid_at is required")
	}

	record := SalaryRecord{
		UserID: userID,
		Amount: input.Amount,
		PaidAt: input.PaidAt,
		Note:   strings.TrimSpace(input.Note),
	}

	id, err := s.repo.CreateRecord(ctx, record)
	if err != nil {
		return SalaryRecord{}, err
	}

	return s.repo.GetRecordByID(ctx, id, userID)
}

func (s *Service) Update(ctx context.Context, id int64, userID int64, input UpdateInput) (SalaryRecord, error) {
	record, err := s.repo.GetRecordByID(ctx, id, userID)
	if err != nil {
		return SalaryRecord{}, err
	}

	if input.Amount != nil {
		if *input.Amount <= 0 {
			return SalaryRecord{}, errors.New("amount must be greater than zero")
		}
		record.Amount = *input.Amount
	}
	if input.PaidAt != nil {
		if input.PaidAt.IsZero() {
			return SalaryRecord{}, errors.New("paid_at is required")
		}
		record.PaidAt = *input.PaidAt
	}
	if input.Note != nil {
		record.Note = strings.TrimSpace(*input.Note)
	}

	if err := s.repo.UpdateRecord(ctx, record); err != nil {
		return SalaryRecord{}, err
	}

	return s.repo.GetRecordByID(ctx, id, userID)
}

func (s *Service) Delete(ctx context.Context, id int64, userID int64) error {
	return s.repo.DeleteRecord(ctx, id, userID)
}

func (s *Service) History(ctx context.Context, userID int64) ([]SalaryRecord, error) {
	return s.repo.FindHistory(ctx, userID)
}

func (s *Service) Current(ctx context.Context, userID int64) (CurrentSalary, error) {
	record, err := s.repo.GetCurrentRecord(ctx, userID)
	if err != nil {
		return CurrentSalary{}, err
	}

	current := CurrentSalary{
		SalaryRecord: record,
	}

	schedule, err := s.repo.GetSchedule(ctx, userID)
	if err != nil {
		return CurrentSalary{}, err
	}
	if schedule != nil {
		day := schedule.SalaryDay
		current.SalaryDay = &day
	}

	return current, nil
}

func (s *Service) SetSalaryDay(ctx context.Context, userID int64, salaryDay int) (SalarySchedule, error) {
	if salaryDay < 1 || salaryDay > 31 {
		return SalarySchedule{}, errors.New("salary_day must be between 1 and 31")
	}

	return s.repo.UpsertSchedule(ctx, userID, salaryDay)
}
