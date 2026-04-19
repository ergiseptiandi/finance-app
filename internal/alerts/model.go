package alerts

import "time"

type AlertType string

const (
	AlertTypeDailySpendingSpike AlertType = "daily_spending_spike"
)

type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

type Alert struct {
	ID             int64         `json:"id"`
	UserID         int64         `json:"-"`
	Type           AlertType     `json:"type"`
	Title          string        `json:"title"`
	Message        string        `json:"message"`
	Severity       AlertSeverity `json:"severity"`
	MetricValue    float64       `json:"metric_value"`
	ThresholdValue float64       `json:"threshold_value"`
	DedupeKey      string        `json:"dedupe_key"`
	IsRead         bool          `json:"is_read"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type AlertListFilter struct {
	Type *string
	Read *bool
}

type EvaluateInput struct {
	DailySpikeMultiplier *float64 `json:"daily_spike_multiplier,omitempty"`
}
