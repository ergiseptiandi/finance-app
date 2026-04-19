package server

import "finance-backend/internal/server/routeinfo"

func dashboardOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"DashboardSummary": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"total_balance":   map[string]any{"type": "number"},
					"monthly_income":  map[string]any{"type": "number"},
					"monthly_expense": map[string]any{"type": "number"},
					"net_cashflow":    map[string]any{"type": "number"},
					"savings_rate":    map[string]any{"type": "number"},
					"expense_ratio":   map[string]any{"type": "number"},
					"debt": map[string]any{
						"$ref": "#/components/schemas/DebtOverview",
					},
				},
			},
			"DebtOverview": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"total_debt":                map[string]any{"type": "number"},
					"paid_debt":                 map[string]any{"type": "number"},
					"remaining_debt":            map[string]any{"type": "number"},
					"total_debt_count":          map[string]any{"type": "integer"},
					"active_debt_count":         map[string]any{"type": "integer"},
					"overdue_debt_count":        map[string]any{"type": "integer"},
					"paid_installments":         map[string]any{"type": "integer"},
					"overdue_installments":      map[string]any{"type": "integer"},
					"upcoming_due_amount":       map[string]any{"type": "number"},
					"upcoming_due_installments": map[string]any{"type": "integer"},
					"debt_to_income_ratio":      map[string]any{"type": "number"},
					"debt_to_balance_ratio":     map[string]any{"type": "number"},
					"completion_rate":           map[string]any{"type": "number"},
				},
			},
			"SpendingPoint": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"date":   map[string]any{"type": "string"},
					"amount": map[string]any{"type": "number"},
				},
			},
			"MonthlySpendingPoint": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"month":  map[string]any{"type": "string"},
					"amount": map[string]any{"type": "number"},
				},
			},
			"ComparisonMetric": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"current":           map[string]any{"type": "number"},
					"previous":          map[string]any{"type": "number"},
					"difference":        map[string]any{"type": "number"},
					"percentage_change": map[string]any{"type": "number"},
				},
			},
			"DashboardComparison": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"today_vs_yesterday":       map[string]any{"$ref": "#/components/schemas/ComparisonMetric"},
					"this_month_vs_last_month": map[string]any{"$ref": "#/components/schemas/ComparisonMetric"},
				},
			},
		},
	}
}

func dashboardRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	return nil
}

func dashboardResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/dashboard/summary", "GET /v1/dashboard/daily-spending", "GET /v1/dashboard/monthly-spending", "GET /v1/dashboard/comparison":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}

func dashboardParameterSchemas(route routeinfo.RouteInfo) []map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/dashboard/summary", "GET /v1/dashboard/daily-spending", "GET /v1/dashboard/monthly-spending":
		return []map[string]any{
			{
				"name":        "month",
				"in":          "query",
				"required":    false,
				"description": "Filter one full month using YYYY-MM. Cannot be combined with start_date or end_date. If omitted together with start_date and end_date, the API uses the current month.",
				"schema": map[string]any{
					"type":    "string",
					"pattern": "^\\d{4}-\\d{2}$",
					"example": "2026-04",
				},
			},
			{
				"name":        "start_date",
				"in":          "query",
				"required":    false,
				"description": "Inclusive custom range start date using YYYY-MM-DD. Must be sent together with end_date. Maximum range is 3 months.",
				"schema": map[string]any{
					"type":    "string",
					"format":  "date",
					"example": "2026-04-01",
				},
			},
			{
				"name":        "end_date",
				"in":          "query",
				"required":    false,
				"description": "Inclusive custom range end date using YYYY-MM-DD. Must be sent together with start_date. Maximum range is 3 months.",
				"schema": map[string]any{
					"type":    "string",
					"format":  "date",
					"example": "2026-06-30",
				},
			},
		}
	default:
		return nil
	}
}
