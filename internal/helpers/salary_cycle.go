package helpers

import "time"

// CurrentSalaryCycle menghitung rentang siklus gaji berdasarkan salaryDay.
// Siklus gaji: dari salaryDay bulan sebelumnya hingga salaryDay bulan berjalan.
// Contoh: salaryDay=25, sekarang 15 Juli 2025 → 25 Juni 2025 ~ 25 Juli 2025
// Contoh: salaryDay=25, sekarang 28 Juli 2025 → 25 Juli 2025 ~ 25 Agustus 2025
func CurrentSalaryCycle(salaryDay int, now ...time.Time) (start time.Time, end time.Time) {
	ref := time.Now()
	if len(now) > 0 && !now[0].IsZero() {
		ref = now[0]
	}

	if salaryDay < 1 {
		salaryDay = 25 // default
	}

	lastDayThisMonth := lastDayOfMonth(ref.Year(), ref.Month(), ref.Location())
	salaryDayThisMonth := salaryDay
	if salaryDayThisMonth > lastDayThisMonth {
		salaryDayThisMonth = lastDayThisMonth
	}

	thisMonthSalaryDate := time.Date(ref.Year(), ref.Month(), salaryDayThisMonth, 0, 0, 0, 0, ref.Location())

	var startMonth time.Time
	if ref.Day() >= salaryDayThisMonth || ref.After(thisMonthSalaryDate) || ref.Equal(thisMonthSalaryDate) {
		// Saat ini sudah melewati atau tepat di hari gajian bulan ini
		// Siklus: salaryDay bulan ini ~ salaryDay bulan depan
		startMonth = thisMonthSalaryDate

		nextMonth := ref.AddDate(0, 1, 0)
		lastDayNextMonth := lastDayOfMonth(nextMonth.Year(), nextMonth.Month(), ref.Location())
		salaryDayNextMonth := salaryDay
		if salaryDayNextMonth > lastDayNextMonth {
			salaryDayNextMonth = lastDayNextMonth
		}
		end = time.Date(nextMonth.Year(), nextMonth.Month(), salaryDayNextMonth, 0, 0, 0, 0, ref.Location())
	} else {
		// Saat ini sebelum hari gajian bulan ini
		// Siklus: salaryDay bulan lalu ~ salaryDay bulan ini
		prevMonth := ref.AddDate(0, -1, 0)
		lastDayPrevMonth := lastDayOfMonth(prevMonth.Year(), prevMonth.Month(), ref.Location())
		salaryDayPrevMonth := salaryDay
		if salaryDayPrevMonth > lastDayPrevMonth {
			salaryDayPrevMonth = lastDayPrevMonth
		}
		startMonth = time.Date(prevMonth.Year(), prevMonth.Month(), salaryDayPrevMonth, 0, 0, 0, 0, ref.Location())

		end = thisMonthSalaryDate
	}

	start = startMonth
	return
}

// SalaryCycleFor menghitung rentang siklus gaji untuk bulan tertentu.
// Parameter month adalah bulan referensi (biasanya bulan saat gaji diterima).
// Contoh: salaryDay=25, year=2025, month=July → 25 Juni 2025 ~ 25 Juli 2025
func SalaryCycleFor(salaryDay int, year int, month time.Month, loc ...*time.Location) (start time.Time, end time.Time) {
	location := time.Local
	if len(loc) > 0 && loc[0] != nil {
		location = loc[0]
	}

	if salaryDay < 1 {
		salaryDay = 25
	}

	lastDayThisMonth := lastDayOfMonth(year, month, location)
	salaryDayThisMonth := salaryDay
	if salaryDayThisMonth > lastDayThisMonth {
		salaryDayThisMonth = lastDayThisMonth
	}

	end = time.Date(year, month, salaryDayThisMonth, 0, 0, 0, 0, location)

	prevMonth := time.Date(year, month, 1, 0, 0, 0, 0, location).AddDate(0, -1, 0)
	lastDayPrevMonth := lastDayOfMonth(prevMonth.Year(), prevMonth.Month(), location)
	salaryDayPrevMonth := salaryDay
	if salaryDayPrevMonth > lastDayPrevMonth {
		salaryDayPrevMonth = lastDayPrevMonth
	}

	start = time.Date(prevMonth.Year(), prevMonth.Month(), salaryDayPrevMonth, 0, 0, 0, 0, location)

	return
}

func lastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
	return firstOfNextMonth.AddDate(0, 0, -1).Day()
}
