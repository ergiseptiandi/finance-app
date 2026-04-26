package server

import "finance-backend/internal/server/routeinfo"

func dashboardOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"DashboardSummary": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"total_balance":   map[string]any{"type": "number"},
					"period_balance":  map[string]any{"type": "number"},
					"monthly_income":  map[string]any{"type": "number"},
					"monthly_expense": map[string]any{"type": "number"},
					"net_cashflow":    map[string]any{"type": "number"},
					"savings_rate":    map[string]any{"type": "number"},
					"expense_ratio":   map[string]any{"type": "number"},
					"budget_summary": map[string]any{
						"$ref": "#/components/schemas/BudgetSummary",
					},
					"goals_progress": map[string]any{
						"type": "array",
						"items": map[string]any{"$ref": "#/components/schemas/GoalProgress"},
					},
					"debt": map[string]any{
						"$ref": "#/components/schemas/DebtOverview",
					},
				},
			},
			"BudgetSummary": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"monthly_budget":    map[string]any{"type": "number"},
					"spent":             map[string]any{"type": "number"},
					"remaining":         map[string]any{"type": "number"},
					"usage_rate":        map[string]any{"type": "number"},
					"over_budget_amount": map[string]any{"type": "number"},
					"is_over_budget":    map[string]any{"type": "boolean"},
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
			"BudgetVsActual": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"budget_amount":      map[string]any{"type": "number"},
					"actual_spent":       map[string]any{"type": "number"},
					"remaining_budget":   map[string]any{"type": "number"},
					"usage_rate":         map[string]any{"type": "number"},
					"over_budget_amount": map[string]any{"type": "number"},
					"status":             map[string]any{"type": "string"},
				},
			},
			"CategoryBreakdownItem": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category":          map[string]any{"type": "string"},
					"amount":            map[string]any{"type": "number"},
					"percentage":        map[string]any{"type": "number"},
					"transaction_count": map[string]any{"type": "integer"},
				},
			},
			"UpcomingBill": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"bill_name":   map[string]any{"type": "string"},
					"amount":      map[string]any{"type": "number"},
					"due_date":    map[string]any{"type": "string"},
					"status":      map[string]any{"type": "string"},
					"source_type": map[string]any{"type": "string"},
				},
			},
			"TopMerchant": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"merchant_name":         map[string]any{"type": "string"},
					"amount":                map[string]any{"type": "number"},
					"transaction_count":     map[string]any{"type": "integer"},
					"last_transaction_date": map[string]any{"type": "string"},
				},
			},
			"Insight": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"type":         map[string]any{"type": "string"},
					"code":         map[string]any{"type": "string"},
					"title":        map[string]any{"type": "string"},
					"message":      map[string]any{"type": "string"},
					"severity":     map[string]any{"type": "string"},
					"change_value": map[string]any{"type": "number"},
				},
			},
			"GoalProgress": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":                map[string]any{"type": "string"},
					"target_amount":       map[string]any{"type": "number"},
					"current_amount":      map[string]any{"type": "number"},
					"progress_percentage": map[string]any{"type": "number"},
					"target_date":         map[string]any{"type": "string"},
					"status":              map[string]any{"type": "string"},
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
	case "GET /v1/dashboard/summary", "GET /v1/dashboard/daily-spending", "GET /v1/dashboard/monthly-spending", "GET /v1/dashboard/comparison", "GET /v1/dashboard/budget-vs-actual", "GET /v1/dashboard/category-breakdown", "GET /v1/dashboard/upcoming-bills", "GET /v1/dashboard/top-merchants", "GET /v1/dashboard/insights", "GET /v1/dashboard/goals-progress":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}

func dashboardParameterSchemas(route routeinfo.RouteInfo) []map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/dashboard/summary", "GET /v1/dashboard/daily-spending", "GET /v1/dashboard/monthly-spending", "GET /v1/dashboard/category-breakdown", "GET /v1/dashboard/top-merchants", "GET /v1/dashboard/insights", "GET /v1/dashboard/goals-progress":
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
	case "GET /v1/dashboard/budget-vs-actual":
		params := dashboardFilterParameterSchemas()
		return append(params, map[string]any{
			"name":        "budget_amount",
			"in":          "query",
			"required":    false,
			"description": "Optional explicit budget amount for the selected period. If omitted, the backend uses monthly income as the fallback budget.",
			"schema": map[string]any{
				"type":    "number",
				"example": 5000000,
			},
		})
	case "GET /v1/dashboard/upcoming-bills":
		return []map[string]any{
			{
				"name":        "days",
				"in":          "query",
				"required":    false,
				"description": "Lookahead window in days for upcoming bills. Defaults to 30.",
				"schema": map[string]any{
					"type":    "integer",
					"example": 30,
				},
			},
		}
	default:
		return nil
	}
}

func dashboardFilterParameterSchemas() []map[string]any {
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
}
