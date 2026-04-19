package server

import "finance-backend/internal/server/routeinfo"

func notificationsOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"NotificationKind": map[string]any{
				"type": "string",
				"enum": []string{"daily_expense_input", "debt_payment"},
			},
			"DeliveryStatus": map[string]any{
				"type": "string",
				"enum": []string{"pending", "sent", "failed", "skipped"},
			},
			"Notification": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":              map[string]any{"type": "integer"},
					"kind":            map[string]any{"$ref": "#/components/schemas/NotificationKind"},
					"title":           map[string]any{"type": "string"},
					"message":         map[string]any{"type": "string"},
					"delivery_status": map[string]any{"$ref": "#/components/schemas/DeliveryStatus"},
					"scheduled_for":   map[string]any{"type": "string", "format": "date-time"},
					"sent_at":         map[string]any{"type": "string", "format": "date-time", "nullable": true},
					"read_at":         map[string]any{"type": "string", "format": "date-time", "nullable": true},
					"dedupe_key":      map[string]any{"type": "string"},
					"created_at":      map[string]any{"type": "string", "format": "date-time"},
					"updated_at":      map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"NotificationSettings": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"user_id":                           map[string]any{"type": "integer"},
					"enabled":                           map[string]any{"type": "boolean"},
					"daily_expense_reminder_enabled":    map[string]any{"type": "boolean"},
					"daily_expense_reminder_time":       map[string]any{"type": "string"},
					"debt_payment_reminder_enabled":     map[string]any{"type": "boolean"},
					"debt_payment_reminder_time":        map[string]any{"type": "string"},
					"debt_payment_reminder_days_before": map[string]any{"type": "integer"},
					"push_token":                        map[string]any{"type": "string"},
				},
			},
			"UpdateNotificationSettingsRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"enabled":                           map[string]any{"type": "boolean"},
					"daily_expense_reminder_enabled":    map[string]any{"type": "boolean"},
					"daily_expense_reminder_time":       map[string]any{"type": "string"},
					"debt_payment_reminder_enabled":     map[string]any{"type": "boolean"},
					"debt_payment_reminder_time":        map[string]any{"type": "string"},
					"debt_payment_reminder_days_before": map[string]any{"type": "integer"},
					"push_token":                        map[string]any{"type": "string"},
				},
			},
		},
	}
}

func notificationsRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "PATCH /v1/notifications/settings":
		return jsonRequestBody("#/components/schemas/UpdateNotificationSettingsRequest")
	default:
		return nil
	}
}

func notificationsResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/notifications", "GET /v1/notifications/settings", "PATCH /v1/notifications/settings", "POST /v1/notifications/generate", "PATCH /v1/notifications/{id}/read":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
