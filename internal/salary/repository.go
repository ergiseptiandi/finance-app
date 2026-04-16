package salary

import "context"

type Repository interface {
	CreateRecord(ctx context.Context, item SalaryRecord) (int64, error)
	GetRecordByID(ctx context.Context, id int64, userID int64) (SalaryRecord, error)
	GetCurrentRecord(ctx context.Context, userID int64) (SalaryRecord, error)
	UpdateRecord(ctx context.Context, item SalaryRecord) error
	DeleteRecord(ctx context.Context, id int64, userID int64) error
	FindHistory(ctx context.Context, userID int64) ([]SalaryRecord, error)
	GetSchedule(ctx context.Context, userID int64) (*SalarySchedule, error)
	UpsertSchedule(ctx context.Context, userID int64, salaryDay int) (SalarySchedule, error)
}
