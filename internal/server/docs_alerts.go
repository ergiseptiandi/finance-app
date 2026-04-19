package server

import "finance-backend/internal/server/routeinfo"

func alertsOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"Alert": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":              map[string]any{"type": "integer"},
					"type":            map[string]any{"type": "string", "enum": []string{"daily_spending_spike"}},
					"title":           map[string]any{"type": "string"},
					"message":         map[string]any{"type": "string"},
					"severity":        map[string]any{"type": "string", "enum": []string{"info", "warning", "critical"}},
					"metric_value":    map[string]any{"type": "number"},
					"threshold_value": map[string]any{"type": "number"},
					"dedupe_key":      map[string]any{"type": "string"},
					"is_read":         map[string]any{"type": "boolean"},
					"created_at":      map[string]any{"type": "string", "format": "date-time"},
					"updated_at":      map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"EvaluateAlertRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"daily_spike_multiplier": map[string]any{"type": "number"},
				},
			},
		},
	}
}

func alertsRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/alerts/evaluate":
		return jsonRequestBody("#/components/schemas/EvaluateAlertRequest")
	default:
		return nil
	}
}

func alertsResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/alerts", "POST /v1/alerts/evaluate", "PATCH /v1/alerts/{id}/read":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
