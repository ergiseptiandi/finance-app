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
			"ExpenseVsSalary": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"monthly_expense": map[string]any{"type": "number"},
					"current_salary":  map[string]any{"type": "number"},
					"percentage":      map[string]any{"type": "number"},
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
	case "GET /v1/dashboard/summary", "GET /v1/dashboard/daily-spending", "GET /v1/dashboard/monthly-spending", "GET /v1/dashboard/comparison", "GET /v1/dashboard/expense-vs-salary":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
