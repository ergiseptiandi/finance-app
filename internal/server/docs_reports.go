package server

import "finance-backend/internal/server/routeinfo"

func reportsOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"ReportCategoryExpense": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category":   map[string]any{"type": "string"},
					"amount":     map[string]any{"type": "number"},
					"percentage": map[string]any{"type": "number"},
				},
			},
			"ReportTrendPoint": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"period": map[string]any{"type": "string"},
					"amount": map[string]any{"type": "number"},
				},
			},
			"ReportAverageDailySpending": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"total_expense":          map[string]any{"type": "number"},
					"days_count":             map[string]any{"type": "integer"},
					"average_daily_spending": map[string]any{"type": "number"},
				},
			},
			"ReportRemainingBalance": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"total_income":      map[string]any{"type": "number"},
					"total_expense":     map[string]any{"type": "number"},
					"remaining_balance": map[string]any{"type": "number"},
				},
			},
		},
	}
}

func reportsRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	return nil
}

func reportsResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/reports/expense-by-category":
		return authResponses("#/components/schemas/SuccessResponse")
	case "GET /v1/reports/spending-trends":
		return authResponses("#/components/schemas/SuccessResponse")
	case "GET /v1/reports/highest-spending-category":
		return authResponses("#/components/schemas/SuccessResponse")
	case "GET /v1/reports/average-daily-spending":
		return authResponses("#/components/schemas/SuccessResponse")
	case "GET /v1/reports/remaining-balance":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
