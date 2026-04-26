package server

import "finance-backend/internal/server/routeinfo"

func budgetOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"BudgetGoalStatus": map[string]any{
				"type": "string",
				"enum": []string{"under_budget", "on_track", "over_budget", "inactive"},
			},
			"BudgetGoal": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":             map[string]any{"type": "integer"},
					"user_id":        map[string]any{"type": "integer"},
					"category_id":    map[string]any{"type": "integer"},
					"category_name":  map[string]any{"type": "string"},
					"category_type":  map[string]any{"type": "string"},
					"monthly_amount": map[string]any{"type": "number"},
					"created_at":     map[string]any{"type": "string", "format": "date-time"},
					"updated_at":     map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"BudgetGoalProgress": map[string]any{
				"allOf": []map[string]any{
					{"$ref": "#/components/schemas/BudgetGoal"},
				},
			},
			"BudgetGoalProgressItem": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":                  map[string]any{"type": "integer"},
					"user_id":             map[string]any{"type": "integer"},
					"category_id":         map[string]any{"type": "integer"},
					"category_name":       map[string]any{"type": "string"},
					"category_type":       map[string]any{"type": "string"},
					"monthly_amount":      map[string]any{"type": "number"},
					"created_at":          map[string]any{"type": "string", "format": "date-time"},
					"updated_at":          map[string]any{"type": "string", "format": "date-time"},
					"current_amount":      map[string]any{"type": "number"},
					"remaining_amount":    map[string]any{"type": "number"},
					"progress_percentage": map[string]any{"type": "number"},
					"status":              map[string]any{"$ref": "#/components/schemas/BudgetGoalStatus"},
				},
			},
			"BudgetSummary": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"monthly_budget":     map[string]any{"type": "number"},
					"spent":              map[string]any{"type": "number"},
					"remaining":          map[string]any{"type": "number"},
					"usage_rate":         map[string]any{"type": "number"},
					"over_budget_amount":  map[string]any{"type": "number"},
					"is_over_budget":     map[string]any{"type": "boolean"},
				},
			},
			"BudgetGoalsList": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"summary": map[string]any{"$ref": "#/components/schemas/BudgetSummary"},
					"items": map[string]any{
						"type":  "array",
						"items": map[string]any{"$ref": "#/components/schemas/BudgetGoalProgressItem"},
					},
				},
			},
			"CreateBudgetGoalRequest": map[string]any{
				"type":     "object",
				"required": []string{"category_id", "monthly_amount"},
				"properties": map[string]any{
					"category_id":    map[string]any{"type": "integer"},
					"monthly_amount": map[string]any{"type": "number"},
				},
			},
			"UpdateBudgetGoalRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category_id":    map[string]any{"type": "integer"},
					"monthly_amount": map[string]any{"type": "number"},
				},
			},
		},
	}
}

func budgetRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/budgets/category-goals":
		return jsonRequestBody("#/components/schemas/CreateBudgetGoalRequest")
	case "PATCH /v1/budgets/category-goals/{id}":
		return jsonRequestBody("#/components/schemas/UpdateBudgetGoalRequest")
	default:
		return nil
	}
}

func budgetResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/budgets/category-goals":
		return authResponses("#/components/schemas/SuccessResponse")
	case "POST /v1/budgets/category-goals", "PATCH /v1/budgets/category-goals/{id}":
		return authResponses("#/components/schemas/SuccessResponse")
	case "DELETE /v1/budgets/category-goals/{id}":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}

func budgetParameterSchemas(route routeinfo.RouteInfo) []map[string]any {
	if route.Method+" "+route.Path == "GET /v1/budgets/category-goals" {
		return []map[string]any{
			{
				"name":        "month",
				"in":          "query",
				"required":    false,
				"description": "Filter budget progress by month using YYYY-MM. Defaults to the current month.",
				"schema": map[string]any{
					"type":    "string",
					"pattern": "^\\d{4}-\\d{2}$",
					"example": "2026-04",
				},
			},
		}
	}

	return nil
}
