package server

import "finance-backend/internal/server/routeinfo"

func reportsOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"ReportPeriod": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"mode":       map[string]any{"type": "string"},
					"month":      map[string]any{"type": "string"},
					"year":       map[string]any{"type": "integer"},
					"start_date": map[string]any{"type": "string", "format": "date"},
					"end_date":   map[string]any{"type": "string", "format": "date"},
				},
			},
			"ReportCategoryExpense": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category":          map[string]any{"type": "string"},
					"amount":            map[string]any{"type": "number"},
					"percentage":        map[string]any{"type": "number"},
					"transaction_count": map[string]any{"type": "integer"},
				},
			},
			"ReportExpenseByCategorySummary": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"total_expense":  map[string]any{"type": "number"},
					"category_count": map[string]any{"type": "integer"},
					"top_category":   map[string]any{"type": "string"},
				},
			},
			"ReportExpenseByCategoryResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"period":  map[string]any{"$ref": "#/components/schemas/ReportPeriod"},
					"summary": map[string]any{"$ref": "#/components/schemas/ReportExpenseByCategorySummary"},
					"items": map[string]any{
						"type":  "array",
						"items": map[string]any{"$ref": "#/components/schemas/ReportCategoryExpense"},
					},
				},
			},
			"ReportTrendPoint": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"period":       map[string]any{"type": "string"},
					"income":       map[string]any{"type": "number"},
					"expense":      map[string]any{"type": "number"},
					"net_cashflow": map[string]any{"type": "number"},
				},
			},
			"ReportSpendingTrendsResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"period":   map[string]any{"$ref": "#/components/schemas/ReportPeriod"},
					"group_by": map[string]any{"type": "string"},
					"items": map[string]any{
						"type":  "array",
						"items": map[string]any{"$ref": "#/components/schemas/ReportTrendPoint"},
					},
				},
			},
			"ReportHighestSpendingCategoryResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"period":            map[string]any{"$ref": "#/components/schemas/ReportPeriod"},
					"category":          map[string]any{"type": "string"},
					"amount":            map[string]any{"type": "number"},
					"percentage":        map[string]any{"type": "number"},
					"transaction_count": map[string]any{"type": "integer"},
				},
			},
			"ReportAverageDailySpendingResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"period":                 map[string]any{"$ref": "#/components/schemas/ReportPeriod"},
					"total_expense":          map[string]any{"type": "number"},
					"days_count":             map[string]any{"type": "integer"},
					"average_daily_spending": map[string]any{"type": "number"},
					"highest_daily_spending": map[string]any{"type": "number"},
					"lowest_daily_spending":  map[string]any{"type": "number"},
				},
			},
			"ReportRemainingBalanceResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"period":            map[string]any{"$ref": "#/components/schemas/ReportPeriod"},
					"total_income":      map[string]any{"type": "number"},
					"total_expense":     map[string]any{"type": "number"},
					"remaining_balance": map[string]any{"type": "number"},
					"savings_rate":      map[string]any{"type": "number"},
					"expense_ratio":     map[string]any{"type": "number"},
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
		return authResponses("#/components/schemas/ReportExpenseByCategoryResponse")
	case "GET /v1/reports/spending-trends":
		return authResponses("#/components/schemas/ReportSpendingTrendsResponse")
	case "GET /v1/reports/highest-spending-category":
		return authResponses("#/components/schemas/ReportHighestSpendingCategoryResponse")
	case "GET /v1/reports/average-daily-spending":
		return authResponses("#/components/schemas/ReportAverageDailySpendingResponse")
	case "GET /v1/reports/remaining-balance":
		return authResponses("#/components/schemas/ReportRemainingBalanceResponse")
	default:
		return nil
	}
}

func reportsParameterSchemas(route routeinfo.RouteInfo) []map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/reports/expense-by-category", "GET /v1/reports/spending-trends", "GET /v1/reports/highest-spending-category", "GET /v1/reports/average-daily-spending", "GET /v1/reports/remaining-balance":
		return reportFilterParameterSchemas()
	default:
		return nil
	}
}

func reportFilterParameterSchemas() []map[string]any {
	return []map[string]any{
		{
			"name":        "month",
			"in":          "query",
			"required":    false,
			"description": "Filter one full month using YYYY-MM. Cannot be combined with year or start_date/end_date.",
			"schema": map[string]any{
				"type":    "string",
				"pattern": "^\\d{4}-\\d{2}$",
				"example": "2026-04",
			},
		},
		{
			"name":        "year",
			"in":          "query",
			"required":    false,
			"description": "Filter one full year using YYYY. Cannot be combined with month or start_date/end_date.",
			"schema": map[string]any{
				"type":    "string",
				"pattern": "^\\d{4}$",
				"example": "2026",
			},
		},
		{
			"name":        "start_date",
			"in":          "query",
			"required":    false,
			"description": "Inclusive custom range start date using YYYY-MM-DD. Must be sent together with end_date. Maximum range is 1 year.",
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
			"description": "Inclusive custom range end date using YYYY-MM-DD. Must be sent together with start_date. Maximum range is 1 year.",
			"schema": map[string]any{
				"type":    "string",
				"format":  "date",
				"example": "2026-12-31",
			},
		},
	}
}
