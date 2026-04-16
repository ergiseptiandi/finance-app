package server

import "finance-backend/internal/server/routeinfo"

func coreOpenAPIComponents() map[string]any {
	return map[string]any{
		"securitySchemes": map[string]any{
			"BearerAuth": map[string]any{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
			},
		},
		"schemas": map[string]any{
			"SuccessResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"Status":  map[string]any{"type": "string"},
					"Message": map[string]any{"type": "string"},
					"Data":    map[string]any{"nullable": true},
				},
			},
			"ErrorResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"Status":  map[string]any{"type": "string"},
					"Message": map[string]any{"type": "string"},
				},
			},
			"RouteInfo": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"method":    map[string]any{"type": "string"},
					"path":      map[string]any{"type": "string"},
					"summary":   map[string]any{"type": "string"},
					"protected": map[string]any{"type": "boolean"},
				},
			},
		},
	}
}

func coreOperationID(route routeinfo.RouteInfo) (string, bool) {
	switch route.Method + " " + route.Path {
	case "GET /":
		return "getRootStatus", true
	case "GET /health":
		return "getHealthStatus", true
	case "GET /routes":
		return "listRoutes", true
	default:
		return "", false
	}
}

func coreRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /", "GET /health", "GET /routes":
		return nil
	default:
		return nil
	}
}

func coreResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /", "GET /health":
		return successResponse("#/components/schemas/SuccessResponse")
	case "GET /routes":
		return successResponse("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
