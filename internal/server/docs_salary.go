package server

import "finance-backend/internal/server/routeinfo"

func salaryOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"SalaryRecord": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":         map[string]any{"type": "integer"},
					"user_id":    map[string]any{"type": "integer"},
					"amount":     map[string]any{"type": "number"},
					"paid_at":    map[string]any{"type": "string", "format": "date-time"},
					"note":       map[string]any{"type": "string"},
					"created_at": map[string]any{"type": "string", "format": "date-time"},
					"updated_at": map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"CurrentSalary": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":         map[string]any{"type": "integer"},
					"user_id":    map[string]any{"type": "integer"},
					"amount":     map[string]any{"type": "number"},
					"paid_at":    map[string]any{"type": "string", "format": "date-time"},
					"note":       map[string]any{"type": "string"},
					"salary_day": map[string]any{"type": "integer", "minimum": 1, "maximum": 31},
					"created_at": map[string]any{"type": "string", "format": "date-time"},
					"updated_at": map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"SalarySchedule": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"salary_day": map[string]any{"type": "integer", "minimum": 1, "maximum": 31},
					"created_at": map[string]any{"type": "string", "format": "date-time"},
					"updated_at": map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"CreateSalaryRequest": map[string]any{
				"type": "object",
				"required": []string{
					"amount",
					"paid_at",
				},
				"properties": map[string]any{
					"amount":  map[string]any{"type": "number"},
					"paid_at": map[string]any{"type": "string", "format": "date-time"},
					"note":    map[string]any{"type": "string"},
				},
			},
			"UpdateSalaryRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"amount":  map[string]any{"type": "number"},
					"paid_at": map[string]any{"type": "string", "format": "date-time"},
					"note":    map[string]any{"type": "string"},
				},
			},
			"SetSalaryDayRequest": map[string]any{
				"type": "object",
				"required": []string{
					"salary_day",
				},
				"properties": map[string]any{
					"salary_day": map[string]any{"type": "integer", "minimum": 1, "maximum": 31},
				},
			},
		},
	}
}

func salaryRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/salaries":
		return jsonRequestBody("#/components/schemas/CreateSalaryRequest")
	case "PATCH /v1/salaries/{id}":
		return jsonRequestBody("#/components/schemas/UpdateSalaryRequest")
	case "PATCH /v1/salaries/schedule":
		return jsonRequestBody("#/components/schemas/SetSalaryDayRequest")
	default:
		return nil
	}
}

func salaryResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/salaries", "GET /v1/salaries", "GET /v1/salaries/current", "PATCH /v1/salaries/{id}", "DELETE /v1/salaries/{id}", "PATCH /v1/salaries/schedule":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
