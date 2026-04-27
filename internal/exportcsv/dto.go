package exportcsv

import "time"

type Scope string
type Language string

const (
	ScopeTransactions Scope = "transactions"
	ScopeDebts        Scope = "debts"
	ScopeReports      Scope = "reports"

	LanguageID Language = "id"
	LanguageEN Language = "en"
)

type Period struct {
	Month      string
	StartDate  *time.Time
	EndDate    *time.Time
	Label      string
	HasFilters bool
}

type Result struct {
	FileName    string
	CSV         []byte
	Partial     bool
	RecordCount int
}
